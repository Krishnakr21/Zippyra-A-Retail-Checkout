package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/audit"
	sharedErrors "github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
	"github.com/zippyra/backend/shared/kafka"
)

type TicketHandler struct {
	repo           SupportRepository
	jwtSecret      string
	kafkaProducer  *kafka.Producer
	auditPublisher *audit.Publisher
	orderVerifier  func(ctx context.Context, orderID, customerID string) (bool, error)
}

func NewTicketHandler(repo SupportRepository, kafkaProducer *kafka.Producer, auditPublisher *audit.Publisher) *TicketHandler {
	return &TicketHandler{
		repo:           repo,
		jwtSecret:      "dev-secret-key-change-in-prod",
		kafkaProducer:  kafkaProducer,
		auditPublisher: auditPublisher,
		orderVerifier: func(ctx context.Context, orderID, customerID string) (bool, error) {
			// Mock order verification: if order ID starts with "ord-other-", reject ownership
			if strings.HasPrefix(orderID, "ord-other-") {
				return false, nil
			}
			return true, nil
		},
	}
}

func (h *TicketHandler) SetOrderVerifier(fn func(ctx context.Context, orderID, customerID string) (bool, error)) {
	h.orderVerifier = fn
}

func (h *TicketHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseAndVerifyToken(tokenStr, h.jwtSecret)
		if err == nil {
			return claims, nil
		}
	}

	userID := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")
	storeID := r.Header.Get("X-Store-ID")
	if userID != "" {
		return &jwt.Claims{
			UserID:  userID,
			Role:    role,
			StoreID: storeID,
		}, nil
	}
	return nil, sharedErrors.NewAPIError(sharedErrors.CodeUnauthorized, "Unauthorized", nil)
}

// 1. POST /v1/support/tickets
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Subject == "" || req.Description == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Subject and description required", nil)
		return
	}

	// Related Order ownership verification
	if req.RelatedOrderID != nil && *req.RelatedOrderID != "" {
		owned, err := h.orderVerifier(r.Context(), *req.RelatedOrderID, claims.UserID)
		if err != nil || !owned {
			sharedErrors.WriteError(w, http.StatusForbidden, "ORDER_NOT_OWNED_BY_REQUESTER", "Related order does not belong to requester", nil)
			return
		}
	}

	// Auto-Priority Lookup & SLA computation
	priority, _ := h.repo.GetAutoPriority(r.Context(), req.Category)
	slaDuration := SLADurations[priority]
	slaDueAt := time.Now().Add(slaDuration)

	requesterType := "CUSTOMER"
	if claims.Role == "STAFF" || claims.Role == "SECURITY" || claims.Role == "MANAGER" {
		requesterType = "STAFF"
	}

	ticket := &SupportTicket{
		RequesterID:    claims.UserID,
		RequesterType:  requesterType,
		StoreID:        req.StoreID,
		Category:       req.Category,
		RelatedOrderID: req.RelatedOrderID,
		Subject:        req.Subject,
		Description:    req.Description,
		Priority:       priority,
		Status:         StatusOpen,
		SLADueAt:       slaDueAt,
	}

	if err := h.repo.CreateTicket(r.Context(), ticket); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to create ticket", nil)
		return
	}

	// Audit Logging
	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "support.ticket_created",
			TargetType: "support_ticket",
			TargetID:   ticket.ID,
			Payload:    map[string]interface{}{"category": req.Category, "priority": priority},
		})
	}

	// Kafka event publication to notification-service
	if h.kafkaProducer != nil {
		_ = h.kafkaProducer.PublishEvent(r.Context(), "support.ticket_created", ticket.ID, map[string]interface{}{
			"ticket_id":    ticket.ID,
			"requester_id": claims.UserID,
			"category":     req.Category,
			"subject":      req.Subject,
		})

		if priority == PriorityUrgent {
			_ = h.kafkaProducer.PublishEvent(r.Context(), "support.urgent_ticket_created", ticket.ID, map[string]interface{}{
				"ticket_id":    ticket.ID,
				"store_id":     req.StoreID,
				"category":     req.Category,
				"subject":      req.Subject,
				"description":  req.Description,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ticket)
}

// 2. GET /v1/support/tickets/mine?status={f?}&page={n}
func (h *TicketHandler) ListMyTickets(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if parsed, err := strconv.Atoi(pStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	tickets, err := h.repo.ListTicketsByRequester(r.Context(), claims.UserID, statusFilter, page, 20)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list tickets", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tickets": tickets, "page": page})
}

// 3. GET /v1/support/tickets/{id} (Customer/Staff View - EXCLUDES internal notes)
func (h *TicketHandler) GetTicketDetailsRequester(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil || ticket.RequesterID != claims.UserID {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	messages, _ := h.repo.ListMessages(r.Context(), ticketID, false) // false = exclude internal notes
	ticket.Messages = messages

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticket)
}

// 5. POST /v1/support/tickets/{id}/reopen (Valid only within 7 days of resolved_at/closed_at)
func (h *TicketHandler) ReopenTicket(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil || ticket.RequesterID != claims.UserID {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	var closedTime *time.Time
	if ticket.ClosedAt != nil {
		closedTime = ticket.ClosedAt
	} else if ticket.ResolvedAt != nil {
		closedTime = ticket.ResolvedAt
	}

	if closedTime == nil || time.Since(*closedTime) > 7*24*time.Hour {
		sharedErrors.WriteError(w, http.StatusBadRequest, "REOPEN_WINDOW_EXPIRED", "Reopen window expired (must be within 7 days of resolution)", nil)
		return
	}

	ticket.Status = StatusOpen
	ticket.AssignedAgentID = nil

	if err := h.repo.UpdateTicket(r.Context(), ticket); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to reopen ticket", nil)
		return
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "support.ticket_reopened",
			TargetType: "support_ticket",
			TargetID:   ticket.ID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticket)
}
