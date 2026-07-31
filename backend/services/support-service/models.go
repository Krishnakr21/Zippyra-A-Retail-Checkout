package main

import (
	"time"
)

type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityNormal Priority = "NORMAL"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

type Status string

const (
	StatusOpen              Status = "OPEN"
	StatusAssigned          Status = "ASSIGNED"
	StatusWaitingOnCustomer Status = "WAITING_ON_CUSTOMER"
	StatusResolved          Status = "RESOLVED"
	StatusClosed            Status = "CLOSED"
)

type Category string

const (
	CategoryPaymentIssue  Category = "PAYMENT_ISSUE"
	CategoryExitGateIssue Category = "EXIT_GATE_ISSUE"
	CategoryOrderIssue    Category = "ORDER_ISSUE"
	CategoryAccountIssue  Category = "ACCOUNT_ISSUE"
	CategoryAppBug        Category = "APP_BUG"
	CategoryDeviceIssue   Category = "DEVICE_ISSUE"
	CategoryOther         Category = "OTHER"
)

var SLADurations = map[Priority]time.Duration{
	PriorityUrgent: 4 * time.Hour,
	PriorityHigh:   24 * time.Hour,
	PriorityNormal: 72 * time.Hour,
	PriorityLow:    168 * time.Hour, // 7 days
}

type SupportTicket struct {
	ID              string     `json:"id"`
	RequesterID     string     `json:"requester_id"`
	RequesterType   string     `json:"requester_type"` // CUSTOMER | STAFF
	StoreID         *string    `json:"store_id,omitempty"`
	ChainID         *string    `json:"chain_id,omitempty"`
	Category        Category   `json:"category"`
	RelatedOrderID  *string    `json:"related_order_id,omitempty"`
	Subject         string     `json:"subject"`
	Description     string     `json:"description"`
	Priority        Priority   `json:"priority"`
	Status          Status     `json:"status"`
	AssignedAgentID *string    `json:"assigned_agent_id,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	SLADueAt        time.Time  `json:"sla_due_at"`
	IsSLAWarned     bool       `json:"is_sla_warned"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Messages        []*TicketMessage `json:"messages,omitempty"`
}

type TicketMessage struct {
	ID             string    `json:"id"`
	TicketID       string    `json:"ticket_id"`
	SenderID       string    `json:"sender_id"`
	SenderType     string    `json:"sender_type"` // CUSTOMER | STAFF | ADMIN | SYSTEM
	Body           string    `json:"body"`
	IsInternalNote bool      `json:"is_internal_note"`
	Attachments    []string  `json:"attachments"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateTicketRequest struct {
	Category       Category `json:"category"`
	RelatedOrderID *string  `json:"related_order_id,omitempty"`
	Subject        string   `json:"subject"`
	Description    string   `json:"description"`
	StoreID        *string  `json:"store_id,omitempty"`
}

type AddMessageRequest struct {
	Body           string   `json:"body"`
	IsInternalNote bool     `json:"is_internal_note"`
	Attachments    []string `json:"attachments"`
}

type AssignTicketRequest struct {
	AgentID string `json:"agent_id"`
}

type UpdateStatusRequest struct {
	Status         Status  `json:"status"`
	ResolutionNote *string `json:"resolution_note,omitempty"`
}

type FeedbackSubmission struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"`
	SourceApp string    `json:"source_app"`
	NPSScore  *int      `json:"nps_score,omitempty"`
	Comment   *string   `json:"comment,omitempty"`
	Context   string    `json:"context"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateFeedbackRequest struct {
	NPSScore  *int    `json:"nps_score,omitempty"`
	Comment   *string `json:"comment,omitempty"`
	SourceApp string  `json:"source_app"`
	Context   string  `json:"context,omitempty"`
}
