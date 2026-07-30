package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/audit"
	sharedErrors "github.com/zippyra/backend/shared/errors"
)

// 4 & 9. POST /v1/support/tickets/{id}/messages
func (h *TicketHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	// Reject if ticket is CLOSED or RESOLVED
	if ticket.Status == StatusClosed || ticket.Status == StatusResolved {
		sharedErrors.WriteError(w, http.StatusConflict, "TICKET_ALREADY_CLOSED", "Cannot post message to a resolved or closed ticket", nil)
		return
	}

	var req AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Message body required", nil)
		return
	}

	senderType := "CUSTOMER"
	isAgent := claims.Role == "ADMIN" || claims.Role == "SUPER_ADMIN" || claims.Role == "OPERATIONS"
	if isAgent {
		senderType = "ADMIN"
	} else if claims.Role == "STAFF" || claims.Role == "SECURITY" || claims.Role == "MANAGER" {
		if ticket.RequesterID == claims.UserID {
			senderType = "STAFF"
		} else {
			senderType = "ADMIN" // Acting as agent
			isAgent = true
		}
	}

	// Internal notes allowed for agents only
	isInternalNote := req.IsInternalNote && isAgent

	msg := &TicketMessage{
		TicketID:       ticketID,
		SenderID:       claims.UserID,
		SenderType:     senderType,
		Body:           req.Body,
		IsInternalNote: isInternalNote,
		Attachments:    req.Attachments,
	}

	if err := h.repo.AddMessage(r.Context(), msg); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to add message", nil)
		return
	}

	// Auto-transition WAITING_ON_CUSTOMER -> ASSIGNED if customer replies
	if !isAgent && ticket.Status == StatusWaitingOnCustomer {
		ticket.Status = StatusAssigned
		_ = h.repo.UpdateTicket(r.Context(), ticket)
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "support.message_added",
			TargetType: "support_ticket",
			TargetID:   ticketID,
			Payload:    map[string]interface{}{"message_id": msg.ID},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}
