package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zippyra/backend/shared/kafka"
)

type EventConsumer struct {
	engine     *NotificationEngine
	opsAlerts  OpsAlertDispatcher
	staffRoster map[string]map[string][]string // store_id -> role -> []staff_user_id
}

func NewEventConsumer(engine *NotificationEngine, opsAlerts OpsAlertDispatcher) *EventConsumer {
	return &EventConsumer{
		engine:     engine,
		opsAlerts:  opsAlerts,
		staffRoster: make(map[string]map[string][]string),
	}
}

func (c *EventConsumer) SetStaffRoster(storeID, role string, staffIDs []string) {
	if _, ok := c.staffRoster[storeID]; !ok {
		c.staffRoster[storeID] = make(map[string][]string)
	}
	c.staffRoster[storeID][role] = staffIDs
}

func (c *EventConsumer) ProcessOrderCompleted(ctx context.Context, value []byte) error {
	var ev struct {
		EventID    string `json:"event_id"`
		OrderID    string `json:"order_id"`
		CustomerID string `json:"customer_id"`
		TotalPaise int64  `json:"total_paise"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	totalRs := float64(ev.TotalPaise) / 100.0
	title := "Order Confirmed"
	body := fmt.Sprintf("Your order for ₹%.2f is confirmed!", totalRs)
	deepLink := fmt.Sprintf("/orders/%s", ev.OrderID)

	return c.engine.Notify(
		ctx,
		ev.CustomerID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		title,
		body,
		deepLink,
		"order.completed",
		ev.EventID,
		"order_confirmation",
		[]map[string]interface{}{{"order_id": ev.OrderID, "total": totalRs}},
	)
}

func (c *EventConsumer) ProcessLoyaltyPointsEarned(ctx context.Context, value []byte) error {
	var ev struct {
		EventID    string `json:"event_id"`
		CustomerID string `json:"customer_id"`
		Points     int    `json:"points"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	title := "Points Earned"
	body := fmt.Sprintf("You earned %d loyalty points on your recent purchase!", ev.Points)

	return c.engine.Notify(
		ctx,
		ev.CustomerID,
		UserTypeCustomer,
		NotificationTypeLoyaltyUpdates,
		title,
		body,
		"/loyalty",
		"loyalty.points_earned",
		ev.EventID,
		"loyalty_points_earned",
		nil,
	)
}

func (c *EventConsumer) ProcessLoyaltyTierUpgraded(ctx context.Context, value []byte) error {
	var ev struct {
		EventID    string `json:"event_id"`
		CustomerID string `json:"customer_id"`
		NewTier    string `json:"new_tier"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	title := "Tier Upgraded!"
	body := fmt.Sprintf("Congratulations! You've been upgraded to %s tier.", ev.NewTier)

	return c.engine.Notify(
		ctx,
		ev.CustomerID,
		UserTypeCustomer,
		NotificationTypeLoyaltyUpdates,
		title,
		body,
		"/loyalty",
		"loyalty.tier_upgraded",
		ev.EventID,
		"tier_upgraded",
		nil,
	)
}

func (c *EventConsumer) ProcessPaymentRefundInitiated(ctx context.Context, value []byte) error {
	var ev struct {
		EventID     string `json:"event_id"`
		OrderID     string `json:"order_id"`
		CustomerID  string `json:"customer_id"`
		AmountPaise int64  `json:"amount_paise"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	amountRs := float64(ev.AmountPaise) / 100.0
	title := "Refund Initiated"
	body := fmt.Sprintf("Your refund of ₹%.2f for order %s is in progress.", amountRs, ev.OrderID)

	return c.engine.Notify(
		ctx,
		ev.CustomerID,
		UserTypeCustomer,
		NotificationTypePaymentRefund,
		title,
		body,
		fmt.Sprintf("/orders/%s", ev.OrderID),
		"payment.refund_initiated",
		ev.EventID,
		"",
		nil,
	)
}

func (c *EventConsumer) ProcessOrderReturnRejected(ctx context.Context, value []byte) error {
	var ev struct {
		EventID    string `json:"event_id"`
		OrderID    string `json:"order_id"`
		CustomerID string `json:"customer_id"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	title := "Return Request Update"
	body := fmt.Sprintf("Your return request for order %s was declined: %s", ev.OrderID, ev.Reason)

	return c.engine.Notify(
		ctx,
		ev.CustomerID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		title,
		body,
		fmt.Sprintf("/orders/%s", ev.OrderID),
		"order.return_rejected",
		ev.EventID,
		"return_rejected",
		nil,
	)
}

func (c *EventConsumer) ProcessDPDPRequestReceived(ctx context.Context, value []byte) error {
	var ev struct {
		EventID string `json:"event_id"`
		UserID  string `json:"user_id"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	title := "Privacy Request Received"
	body := "We received your DPDP data request. Processing under 30-day SLA."

	return c.engine.Notify(
		ctx,
		ev.UserID,
		UserTypeCustomer,
		NotificationTypeMarketing,
		title,
		body,
		"/settings/privacy",
		"dpdp.request_received",
		ev.EventID,
		"",
		nil,
	)
}

func (c *EventConsumer) ProcessExitRFIDFailure(ctx context.Context, value []byte) error {
	var ev struct {
		EventID string `json:"event_id"`
		StoreID string `json:"store_id"`
		TagID   string `json:"tag_id"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	// SECURITY staff roster lookup
	securityStaff := c.staffRoster[ev.StoreID]["SECURITY"]
	if len(securityStaff) == 0 {
		log.Printf("[EventConsumer] No SECURITY staff rostered for store %s", ev.StoreID)
		return nil
	}

	title := "SECURITY ALARM: Gate Barrier Locked"
	body := fmt.Sprintf("RFID validation failure on tag %s at store %s. Customer physically at gate.", ev.TagID, ev.StoreID)

	for _, staffID := range securityStaff {
		_ = c.engine.Notify(
			ctx,
			staffID,
			UserTypeStaff,
			NotificationTypeSecurityAlerts,
			title,
			body,
			"/staff/security/alarms",
			"exit.rfid_failure",
			ev.EventID,
			"",
			nil,
		)
	}

	return nil
}

func (c *EventConsumer) ProcessInternalStaffAlert(ctx context.Context, topic string, value []byte) error {
	var ev struct {
		EventID string `json:"event_id"`
		StoreID string `json:"store_id"`
		Details string `json:"details"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	managers := c.staffRoster[ev.StoreID]["MANAGER"]
	if len(managers) == 0 {
		return nil
	}

	title := fmt.Sprintf("Store Alert: %s", topic)
	body := fmt.Sprintf("Alert for store %s: %s", ev.StoreID, ev.Details)

	for _, mgrID := range managers {
		_ = c.engine.Notify(
			ctx,
			mgrID,
			UserTypeStaff,
			NotificationTypeSecurityAlerts,
			title,
			body,
			"/staff/alerts",
			topic,
			ev.EventID,
			"",
			nil,
		)
	}

	return nil
}

func (c *EventConsumer) ProcessSupportTicketCreated(ctx context.Context, value []byte) error {
	var ev struct {
		TicketID    string `json:"ticket_id"`
		RequesterID string `json:"requester_id"`
		Subject     string `json:"subject"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	return c.engine.Notify(
		ctx,
		ev.RequesterID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		"Support Ticket Created",
		fmt.Sprintf("Your ticket '%s' has been received.", ev.Subject),
		fmt.Sprintf("/support/tickets/%s", ev.TicketID),
		"support.ticket_created",
		ev.TicketID,
		"",
		nil,
	)
}

func (c *EventConsumer) ProcessSupportUrgentTicketCreated(ctx context.Context, value []byte) error {
	var ev struct {
		TicketID string `json:"ticket_id"`
		StoreID  string `json:"store_id"`
		Subject  string `json:"subject"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	storeID := ev.StoreID
	if storeID == "" {
		return nil
	}

	staffList := append(c.staffRoster[storeID]["SECURITY"], c.staffRoster[storeID]["MANAGER"]...)
	for _, staffID := range staffList {
		_ = c.engine.Notify(
			ctx,
			staffID,
			UserTypeStaff,
			NotificationTypeSecurityAlerts,
			"URGENT STORE TICKET",
			fmt.Sprintf("Urgent issue at store %s: %s", storeID, ev.Subject),
			fmt.Sprintf("/staff/tickets/%s", ev.TicketID),
			"support.urgent_ticket_created",
			ev.TicketID,
			"",
			nil,
		)
	}
	return nil
}

func (c *EventConsumer) ProcessSupportTicketResolved(ctx context.Context, value []byte) error {
	var ev struct {
		TicketID    string `json:"ticket_id"`
		RequesterID string `json:"requester_id"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	return c.engine.Notify(
		ctx,
		ev.RequesterID,
		UserTypeCustomer,
		NotificationTypeOrderUpdates,
		"Support Ticket Resolved",
		fmt.Sprintf("Your support ticket #%s has been resolved.", ev.TicketID),
		fmt.Sprintf("/support/tickets/%s", ev.TicketID),
		"support.ticket_resolved",
		ev.TicketID,
		"",
		nil,
	)
}

func (c *EventConsumer) ProcessComplianceReconciliationDiscrepancy(ctx context.Context, value []byte) error {
	var ev map[string]interface{}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	return c.opsAlerts.DispatchOpsAlert(ctx, "compliance.reconciliation_discrepancy", ev)
}

func (c *EventConsumer) ProcessSupportTicketSLAWarning(ctx context.Context, value []byte) error {
	var ev map[string]interface{}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	return c.opsAlerts.DispatchOpsAlert(ctx, "support.ticket_sla_warning", ev)
}

func (c *EventConsumer) ProcessAccountRecovered(ctx context.Context, value []byte) error {
	var ev struct {
		UserID   string `json:"user_id"`
		OldPhone string `json:"old_phone"`
		NewPhone string `json:"new_phone"`
	}
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}

	title := "Security Alert: Account Phone Updated"
	body := fmt.Sprintf("Your account phone number was updated to %s. If this wasn't you, contact support immediately.", ev.NewPhone)
	return c.engine.Notify(ctx, ev.UserID, UserTypeCustomer, NotificationTypeSecurityAlerts, title, body, "/settings/security", "account.recovered", ev.UserID, "account_recovered", nil)
}

func (c *EventConsumer) RegisterKafkaHandlers(consumer *kafka.Consumer) {
	consumer.RegisterHandler("order.completed", c.ProcessOrderCompleted)
	consumer.RegisterHandler("account.recovered", c.ProcessAccountRecovered)
	consumer.RegisterHandler("loyalty.points_earned", c.ProcessLoyaltyPointsEarned)
	consumer.RegisterHandler("loyalty.tier_upgraded", c.ProcessLoyaltyTierUpgraded)
	consumer.RegisterHandler("payment.refund_initiated", c.ProcessPaymentRefundInitiated)
	consumer.RegisterHandler("order.return_rejected", c.ProcessOrderReturnRejected)
	consumer.RegisterHandler("dpdp.request_received", c.ProcessDPDPRequestReceived)
	consumer.RegisterHandler("exit.rfid_failure", c.ProcessExitRFIDFailure)
	consumer.RegisterHandler("compliance.reconciliation_discrepancy", c.ProcessComplianceReconciliationDiscrepancy)

	consumer.RegisterHandler("support.ticket_created", c.ProcessSupportTicketCreated)
	consumer.RegisterHandler("support.urgent_ticket_created", c.ProcessSupportUrgentTicketCreated)
	consumer.RegisterHandler("support.ticket_resolved", c.ProcessSupportTicketResolved)
	consumer.RegisterHandler("support.ticket_sla_warning", c.ProcessSupportTicketSLAWarning)

	staffTopics := []string{"inventory.low_stock", "inventory.shrinkage_alert", "warehouse.po_auto_created", "warehouse.transfer_discrepancy", "device.offline", "device.back_online"}
	for _, top := range staffTopics {
		t := top
		consumer.RegisterHandler(t, func(ctx context.Context, val []byte) error {
			return c.ProcessInternalStaffAlert(ctx, t, val)
		})
	}
}
