package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zippyra/backend/shared/logger"
)

type PaymentConfirmedConsumer struct {
	repo         Repository
	exitTokenSvc ExitTokenService
	invoiceSvc   InvoiceService
	maxRetries   int
	backoffBase  time.Duration
}

func NewPaymentConfirmedConsumer(
	repo Repository,
	exitTokenSvc ExitTokenService,
	invoiceSvc InvoiceService,
) *PaymentConfirmedConsumer {
	return &PaymentConfirmedConsumer{
		repo:         repo,
		exitTokenSvc: exitTokenSvc,
		invoiceSvc:   invoiceSvc,
		maxRetries:   3,
		backoffBase:  100 * time.Millisecond,
	}
}

func (c *PaymentConfirmedConsumer) ProcessPaymentConfirmed(ctx context.Context, msgValue []byte) error {
	var payload PaymentConfirmedPayload
	if err := json.Unmarshal(msgValue, &payload); err != nil {
		logger.Error("Failed to unmarshal payment.confirmed payload: %v", err)
		return fmt.Errorf("invalid json payload: %w", err)
	}

	if payload.PaymentID == "" {
		return fmt.Errorf("payment_id is required in payment.confirmed event")
	}

	// Retry loop for core order & exit-token creation (3 retries with backoff)
	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		err := c.processCoreOrder(ctx, &payload)
		if err == nil {
			// Core order created successfully!
			return nil
		}

		lastErr = err
		logger.Warn("Core order creation attempt %d/%d failed for payment %s: %v", attempt, c.maxRetries, payload.PaymentID, err)
		if attempt < c.maxRetries {
			time.Sleep(time.Duration(attempt) * c.backoffBase)
		}
	}

	// Compensating Action (Gap #43 Saga Pattern):
	// Core order creation failed after 3 retries!
	// Insert order.creation_failed outbox event to trigger payment-service refund compensation
	logger.Error("CRITICAL: Core order creation failed after %d retries for payment %s. Triggering compensating payment refund!", c.maxRetries, payload.PaymentID)
	_ = c.repo.CreateOrderFailureOutboxTx(
		ctx,
		payload.PaymentID,
		payload.UserID,
		payload.StoreID,
		payload.AmountPaise,
		fmt.Sprintf("Order creation failed after retries: %v", lastErr),
	)

	return nil
}

func (c *PaymentConfirmedConsumer) processCoreOrder(ctx context.Context, payload *PaymentConfirmedPayload) error {
	// Construct Items array (using payload items if provided, or default fallback item)
	items := payload.Items
	if len(items) == 0 {
		// Default item representation carrying payment amount
		items = []OrderItem{
			{
				Barcode:      "8901234567890",
				Name:         "Store Checkout Item",
				Qty:          1,
				PricePaise:   payload.AmountPaise,
				HSNCode:      "84713010",
				IsReturnable: true,
			},
		}
	}

	// Prepare returnable flags
	flags := make([]OrderItemReturnableFlag, len(items))
	for i, item := range items {
		flags[i] = OrderItemReturnableFlag{
			Barcode:      item.Barcode,
			IsReturnable: item.IsReturnable,
			ReturnedQty:  0,
		}
	}

	subtotal := payload.AmountPaise
	payable := payload.PayableAmountPaise
	if payable == 0 {
		payable = subtotal
	}

	order := &Order{
		PaymentID:         payload.PaymentID,
		UserID:            payload.UserID,
		StoreID:           payload.StoreID,
		Items:             items,
		SubtotalPaise:     subtotal,
		DiscountPaise:     subtotal - payable,
		CGSTPaise:         (payable * 9) / 100,
		SGSTPaise:         (payable * 9) / 100,
		IGSTPaise:         0,
		TotalPaise:        payable,
		LoyaltyPointsUsed: payload.LoyaltyPointsUsed,
		PaymentMethod:     payload.PaymentMethod,
		SupplyType:        "INTRASTATE",
		Status:            StatusCompleted,
	}

	// Outbox event payload for order.completed
	completedPayload := OrderCompletedPayload{
		OrderID:           order.ID,
		SessionID:         payload.SessionID,
		ChainID:           payload.ChainID,
		UserID:            order.UserID,
		StoreID:           order.StoreID,
		TotalPaise:        order.TotalPaise,
		Items:             order.Items,
		LoyaltyPointsUsed: order.LoyaltyPointsUsed,
		PaymentMethod:     order.PaymentMethod,
		Timestamp:         time.Now(),
	}

	// Transactional step 1, 4 & 7: INSERT orders ON CONFLICT DO NOTHING + Issue Exit Token + INSERT order_creation_outbox
	completedBytes, _ := json.Marshal(completedPayload)
	inserted, err := c.repo.CreateOrderAndOutboxTx(ctx, order, flags, c.exitTokenSvc, TopicOrderCompleted, completedBytes)
	if err != nil {
		return fmt.Errorf("database transaction error: %w", err)
	}

	if !inserted {
		// Idempotency: order already exists for this payment_id
		logger.Info("Payment %s already processed (order exists). Idempotent ack.", payload.PaymentID)
		return nil
	}

	// Step 5: Async non-blocking invoice generation (invoice failure NEVER fails the order)
	if c.invoiceSvc != nil {
		go func(orderID string) {
			asyncCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = c.invoiceSvc.GenerateAndUploadInvoice(asyncCtx, orderID)
		}(order.ID)
	}

	return nil
}
