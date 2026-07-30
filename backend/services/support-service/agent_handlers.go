package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/audit"
	sharedErrors "github.com/zippyra/backend/shared/errors"
)

// 6. GET /v1/support/tickets (Agent view)
func (h *TicketHandler) ListTicketsAgent(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	isManager := claims.Role == "MANAGER"
	isAdmin := claims.Role == "ADMIN" || claims.Role == "SUPER_ADMIN" || claims.Role == "OPERATIONS"

	if !isManager && !isAdmin {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Agent access required", nil)
		return
	}

	requestedStoreID := r.URL.Query().Get("store_id")

	// MANAGER store scope enforcement
	if isManager {
		if claims.StoreID == "" {
			sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Manager has no assigned store_id", nil)
			return
		}
		if requestedStoreID != "" && requestedStoreID != claims.StoreID {
			sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Manager cannot query tickets outside assigned store", nil)
			return
		}
		requestedStoreID = claims.StoreID
	}

	statusFilter := r.URL.Query().Get("status")
	priorityFilter := r.URL.Query().Get("priority")
	agentIDFilter := r.URL.Query().Get("assigned_agent_id")

	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if parsed, err := strconv.Atoi(pStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	tickets, err := h.repo.ListTicketsForAgent(r.Context(), statusFilter, priorityFilter, agentIDFilter, requestedStoreID, page, 20)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list tickets", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tickets": tickets, "page": page})
}

// 7. GET /v1/support/tickets/{id} (Agent View - INCLUDES internal notes)
func (h *TicketHandler) GetTicketDetailsAgent(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	isManager := claims.Role == "MANAGER"
	isAdmin := claims.Role == "ADMIN" || claims.Role == "SUPER_ADMIN" || claims.Role == "OPERATIONS"

	if !isManager && !isAdmin {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Agent access required", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	// MANAGER store scope check
	if isManager && ticket.StoreID != nil && *ticket.StoreID != claims.StoreID {
		sharedErrors.WriteError(w, http.StatusForbidden, sharedErrors.CodeForbidden, "Manager cannot view tickets outside assigned store", nil)
		return
	}

	messages, _ := h.repo.ListMessages(r.Context(), ticketID, true) // true = include internal notes
	ticket.Messages = messages

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticket)
}

// 8. PUT /v1/support/tickets/{id}/assign
func (h *TicketHandler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	var req AssignTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AgentID == "" {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Agent ID required", nil)
		return
	}

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	ticket.AssignedAgentID = &req.AgentID
	ticket.Status = StatusAssigned

	if err := h.repo.UpdateTicket(r.Context(), ticket); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to assign ticket", nil)
		return
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "support.ticket_assigned",
			TargetType: "support_ticket",
			TargetID:   ticketID,
			Payload:    map[string]interface{}{"agent_id": req.AgentID},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticket)
}

// 10. PUT /v1/support/tickets/{id}/status
func (h *TicketHandler) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	ticketID := vars["id"]

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedErrors.WriteError(w, http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "Invalid payload", nil)
		return
	}

	// Resolution note required for RESOLVED transition
	if req.Status == StatusResolved && (req.ResolutionNote == nil || *req.ResolutionNote == "") {
		sharedErrors.WriteError(w, http.StatusBadRequest, "RESOLUTION_NOTE_REQUIRED", "Resolution note is required when resolving a ticket", nil)
		return
	}

	ticket, err := h.repo.GetTicketByID(r.Context(), ticketID)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "Ticket not found", nil)
		return
	}

	now := time.Now()
	ticket.Status = req.Status

	if req.Status == StatusResolved {
		ticket.ResolvedAt = &now
	} else if req.Status == StatusClosed {
		ticket.ClosedAt = &now
	}

	if err := h.repo.UpdateTicket(r.Context(), ticket); err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to update ticket status", nil)
		return
	}

	// Add resolution note as internal/system message if provided
	if req.ResolutionNote != nil && *req.ResolutionNote != "" {
		_ = h.repo.AddMessage(r.Context(), &TicketMessage{
			TicketID:       ticketID,
			SenderID:       claims.UserID,
			SenderType:     "ADMIN",
			Body:           "Resolution Note: " + *req.ResolutionNote,
			IsInternalNote: false,
		})
	}

	// Publish support.ticket_resolved event
	if (req.Status == StatusResolved || req.Status == StatusClosed) && h.kafkaProducer != nil {
		_ = h.kafkaProducer.PublishEvent(r.Context(), "support.ticket_resolved", ticket.ID, map[string]interface{}{
			"ticket_id":    ticket.ID,
			"requester_id": ticket.RequesterID,
			"status":       req.Status,
		})
	}

	if h.auditPublisher != nil {
		_ = h.auditPublisher.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:    claims.UserID,
			ActionType: "support.ticket_status_updated",
			TargetType: "support_ticket",
			TargetID:   ticketID,
			Payload:    map[string]interface{}{"status": req.Status},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticket)
}

// 11. GET /v1/support/tickets/overdue
func (h *TicketHandler) ListOverdueTickets(w http.ResponseWriter, r *http.Request) {
	_, err := h.extractClaims(r)
	if err != nil {
		sharedErrors.WriteError(w, http.StatusUnauthorized, sharedErrors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	tickets, err := h.repo.ListOverdueTickets(r.Context())
	if err != nil {
		sharedErrors.WriteError(w, http.StatusInternalServerError, sharedErrors.CodeInternalError, "Failed to list overdue tickets", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"overdue_tickets": tickets})
}
