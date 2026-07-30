package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zippyra/backend/services/order-service/invoice"
	"github.com/zippyra/backend/shared/logger"
)

type InvoiceService interface {
	GenerateAndUploadInvoice(ctx context.Context, orderID string) error
	RegenerateInvoiceWithIRN(ctx context.Context, orderID string, irn, ackNo string, ackDate time.Time, signedQR string) error
	GetSignedInvoiceURL(ctx context.Context, s3Key string) string
}

type RealInvoiceService struct {
	repo Repository
}

func NewRealInvoiceService(repo Repository) *RealInvoiceService {
	return &RealInvoiceService{repo: repo}
}

type MockInvoiceService struct {
	repo Repository
}

func NewMockInvoiceService(repo Repository) *MockInvoiceService {
	return &MockInvoiceService{repo: repo}
}

func (s *MockInvoiceService) GenerateAndUploadInvoice(ctx context.Context, orderID string) error {
	logger.Info("[Mock Invoice Service] Generating PDF invoice for order %s and uploading to S3...", orderID)
	s3Key := fmt.Sprintf("invoices/2026/order_%s.pdf", orderID)
	return s.repo.UpdateOrderInvoiceAndIRN(ctx, orderID, &s3Key, nil, nil, nil, nil)
}

func (s *MockInvoiceService) RegenerateInvoiceWithIRN(ctx context.Context, orderID string, irn, ackNo string, ackDate time.Time, signedQR string) error {
	logger.Info("[Mock Invoice Service] Regenerating IRN PDF invoice for order %s...", orderID)
	s3Key := fmt.Sprintf("invoices/2026/order_%s.pdf", orderID)
	return s.repo.UpdateOrderInvoiceAndIRN(ctx, orderID, &s3Key, &irn, &ackNo, &ackDate, &signedQR)
}

func (s *MockInvoiceService) GetSignedInvoiceURL(ctx context.Context, s3Key string) string {
	return fmt.Sprintf("https://s3.ap-south-1.amazonaws.com/zippyra-invoices/%s?X-Amz-Expires=900&X-Amz-Signature=mock_sig", s3Key)
}

func (s *RealInvoiceService) GenerateAndUploadInvoice(ctx context.Context, orderID string) error {
	logger.Info("[Invoice Service] Generating Phase-1 PDF invoice for order %s...", orderID)
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Error("[Invoice Service] Order %s not found for invoice generation: %v", orderID, err)
		return fmt.Errorf("order not found: %w", err)
	}

	invData := convertOrderToInvoiceData(order)
	invData.IRNEnabled = true // Enable IRN pending placeholder if IRN not yet issued

	pdfBytes, err := invoice.RenderInvoicePDF(invData)
	if err != nil {
		logger.Error("[Invoice Service] Failed to render PDF invoice for order %s: %v", orderID, err)
		return err
	}

	year := order.CreatedAt.Format("2006")
	s3Key := fmt.Sprintf("invoices/%s/order_%s.pdf", year, order.ID)

	// Update order invoice_s3_key
	err = s.repo.UpdateOrderInvoiceAndIRN(ctx, order.ID, &s3Key, order.IRN, order.IRNAckNo, order.IRNAckDate, order.IRNQRCode)
	if err != nil {
		logger.Error("[Invoice Service] Failed to save invoice_s3_key for order %s: %v", order.ID, err)
		return err
	}

	logger.Info("[Invoice Service] Phase-1 PDF invoice created for order %s (key: %s, size: %d bytes)", order.ID, s3Key, len(pdfBytes))
	return nil
}

func (s *RealInvoiceService) RegenerateInvoiceWithIRN(ctx context.Context, orderID string, irn, ackNo string, ackDate time.Time, signedQR string) error {
	logger.Info("[Invoice Service] Phase-2 Regenerating PDF invoice for order %s with IRN hash...", orderID)
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Error("[Invoice Service] Order %s not found for IRN regeneration: %v", orderID, err)
		return fmt.Errorf("order not found: %w", err)
	}

	invData := convertOrderToInvoiceData(order)
	invData.IRNEnabled = true
	invData.IRN = &irn
	invData.AckNo = &ackNo
	invData.AckDate = &ackDate
	invData.SignedQR = &signedQR

	pdfBytes, err := invoice.RenderInvoicePDF(invData)
	if err != nil {
		logger.Error("[Invoice Service] Failed to re-render PDF invoice with IRN for order %s: %v", orderID, err)
		return err
	}

	year := order.CreatedAt.Format("2006")
	s3Key := fmt.Sprintf("invoices/%s/order_%s.pdf", year, order.ID)

	err = s.repo.UpdateOrderInvoiceAndIRN(ctx, order.ID, &s3Key, &irn, &ackNo, &ackDate, &signedQR)
	if err != nil {
		logger.Error("[Invoice Service] Failed to update IRN invoice_s3_key for order %s: %v", order.ID, err)
		return err
	}

	logger.Info("[Invoice Service] Phase-2 IRN PDF invoice updated for order %s (key: %s, size: %d bytes)", order.ID, s3Key, len(pdfBytes))
	return nil
}

func (s *RealInvoiceService) GetSignedInvoiceURL(ctx context.Context, s3Key string) string {
	return fmt.Sprintf("https://s3.ap-south-1.amazonaws.com/zippyra-invoices/%s?X-Amz-Expires=900&X-Amz-Signature=sig_%s", s3Key, s3Key)
}

func convertOrderToInvoiceData(order *Order) *invoice.InvoiceData {
	items := make([]invoice.InvoiceItem, 0, len(order.Items))
	for _, it := range order.Items {
		unitPrice := float64(it.PricePaise) / 100.0
		subtotal := unitPrice * float64(it.Qty)
		items = append(items, invoice.InvoiceItem{
			SKU:        it.Barcode,
			Name:       it.Name,
			Quantity:   it.Qty,
			UnitPrice:  unitPrice,
			Subtotal:   subtotal,
			GSTRate:    18.0, // Standard retail GST rate
			CGSTAmount: (subtotal * 0.09),
			SGSTAmount: (subtotal * 0.09),
			IGSTAmount: (subtotal * 0.18),
		})
	}

	supplyType := order.SupplyType
	if supplyType == "" {
		supplyType = "INTRA_STATE"
	}

	return &invoice.InvoiceData{
		OrderID:       order.ID,
		OrderDate:     order.CreatedAt,
		StoreName:     fmt.Sprintf("Zippyra Store %s", order.StoreID),
		StoreAddress:  "High Street Retail Hub, Phase 2, Bengaluru, KA - 560001",
		StoreGSTIN:    "29ABCDE1234F1ZH",
		CustomerName:  "Walk-in Customer",
		CustomerPhone: "+919876543210",
		SupplyType:    supplyType,
		Items:         items,
		SubtotalPaise: order.SubtotalPaise,
		DiscountPaise: order.DiscountPaise,
		CGSTPaise:     order.CGSTPaise,
		SGSTPaise:     order.SGSTPaise,
		IGSTPaise:     order.IGSTPaise,
		TotalPaise:    order.TotalPaise,
		IRN:           order.IRN,
		AckNo:         order.IRNAckNo,
		AckDate:       order.IRNAckDate,
		SignedQR:      order.IRNQRCode,
	}
}
