package main

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
)

func SetupRoutes(r *mux.Router, handler *TicketHandler, db *sql.DB) {
	// Customer / Staff Ticket Endpoints
	r.HandleFunc("/v1/support/tickets", handler.CreateTicket).Methods("POST")
	r.HandleFunc("/v1/support/tickets/mine", handler.ListMyTickets).Methods("GET")
	r.HandleFunc("/v1/support/tickets/overdue", handler.ListOverdueTickets).Methods("GET") // Special overdue queue
	r.HandleFunc("/v1/support/tickets/{id}", handler.GetTicketDetailsRequester).Methods("GET")
	r.HandleFunc("/v1/support/tickets/{id}/messages", handler.AddMessage).Methods("POST")
	r.HandleFunc("/v1/support/tickets/{id}/reopen", handler.ReopenTicket).Methods("POST")

	// Agent Endpoints
	r.HandleFunc("/v1/support/tickets", handler.ListTicketsAgent).Methods("GET")
	r.HandleFunc("/v1/support/tickets/{id}/agent-view", handler.GetTicketDetailsAgent).Methods("GET")
	r.HandleFunc("/v1/support/tickets/{id}/assign", handler.AssignTicket).Methods("PUT")
	r.HandleFunc("/v1/support/tickets/{id}/status", handler.UpdateTicketStatus).Methods("PUT")

	// Feedback Endpoints
	r.HandleFunc("/v1/support/feedback", handler.SubmitFeedback).Methods("POST")
	r.HandleFunc("/v1/support/feedback", handler.ListFeedback).Methods("GET")

	// Health Check Probes
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods("GET")
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods("GET")
	r.HandleFunc("/healthz/startup", health.StartupHandler).Methods("GET")
}
