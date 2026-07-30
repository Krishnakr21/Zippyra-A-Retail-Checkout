package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/health"
)

func RegisterRoutes(r *mux.Router, h *ComplianceHandler) {
	// Health check endpoints
	r.HandleFunc("/healthz/live", health.LiveHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthz/ready", health.ReadyHandler).Methods(http.MethodGet)

	// GST E-Invoicing (IRN)
	r.HandleFunc("/v1/compliance/irn", h.ListIRNRecordsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/irn/{order_id}", h.GetIRNRecordHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/irn/records", h.ListIRNRecordsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/irn/{order_id}/retry", h.RetryIRNRecordsHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/irn/retry", h.RetryIRNRecordsHandler).Methods(http.MethodPost)

	// DPDP Act Consents & Rights
	r.HandleFunc("/v1/compliance/dpdp/consents", h.SubmitConsentHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/dpdp/consents", h.GetConsentsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/consents/{type}", h.UpdateSingleConsentHandler).Methods(http.MethodPut, http.MethodPost)
	r.HandleFunc("/v1/compliance/dpdp/requests", h.CreateDPDPRequestHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/dpdp/requests", h.ListDPDPRequestsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/requests", h.ListDPDPRequestsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/requests/mine", h.GetMyDPDPRequestsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/dpdp/requests/{id}/review", h.ReviewDPDPRequestHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/dpdp/requests/{id}/process-deletion", h.ProcessDeletionHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/requests/{id}/process-deletion", h.ProcessDeletionHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/dpdp/requests/{id}/process-access", h.accessHandler.HandleProcessAccessRequest).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/requests/{id}/process-access", h.accessHandler.HandleProcessAccessRequest).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/access-exports/{dpdp_request_id}/download", h.accessHandler.HandleDownloadAccessExport).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/grievance-officer", h.GetGrievanceOfficerHandler).Methods(http.MethodGet)

	// Merchant KYC
	r.HandleFunc("/v1/compliance/kyc", h.GetKYCHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/kyc/{store_id}", h.UpsertKYCHandler).Methods(http.MethodPut, http.MethodPost)
	r.HandleFunc("/v1/compliance/kyc/stores/{store_id}", h.UpsertKYCHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/kyc/stores/{store_id}", h.GetKYCHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/kyc/incomplete", h.ListIncompleteKYCHandler).Methods(http.MethodGet)

	// Velocity Monitoring & Fraud Prevention
	r.HandleFunc("/v1/compliance/velocity-alerts", h.ListVelocityAlertsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/velocity-alerts/{id}/resolve", h.ResolveVelocityAlertHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/velocity/alerts", h.ListVelocityAlertsHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/velocity/alerts/{id}/resolve", h.ResolveVelocityAlertHandler).Methods(http.MethodPost)

	// Settlement Reconciliation
	r.HandleFunc("/v1/compliance/reconciliation/run", h.RunReconciliationHandler).Methods(http.MethodPost)
	r.HandleFunc("/v1/compliance/reconciliation/reports", h.GetReconciliationReportHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/compliance/reconciliation-reports", h.GetReconciliationReportHandler).Methods(http.MethodGet)
}
