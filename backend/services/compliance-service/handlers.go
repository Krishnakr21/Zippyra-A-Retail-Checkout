package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type ComplianceHandler struct {
	repo         Repository
	consentSvc   *DPDPConsentService
	requestSvc   *DPDPRequestService
	deletionProc *DPDPDeletionProcessor
	accessHandler *AccessFulfillmentHandler
	kycSvc       *KYCService
	reconJob     *ReconciliationJob
	irnRetryJob  *IRNRetryJob
	jwtSecret    string
}

func NewComplianceHandler(
	repo Repository,
	consentSvc *DPDPConsentService,
	requestSvc *DPDPRequestService,
	deletionProc *DPDPDeletionProcessor,
	accessHandler *AccessFulfillmentHandler,
	kycSvc *KYCService,
	reconJob *ReconciliationJob,
	irnRetryJob *IRNRetryJob,
	jwtSecret string,
) *ComplianceHandler {
	return &ComplianceHandler{
		repo:          repo,
		consentSvc:    consentSvc,
		requestSvc:    requestSvc,
		deletionProc:  deletionProc,
		accessHandler: accessHandler,
		kycSvc:        kycSvc,
		reconJob:      reconJob,
		irnRetryJob:   irnRetryJob,
		jwtSecret:     jwtSecret,
	}
}

func (h *ComplianceHandler) extractClaims(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.NewAPIError(errors.CodeUnauthorized, "Missing authorization header", nil)
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.NewAPIError(errors.CodeUnauthorized, "Invalid token format", nil)
	}
	claims, err := jwt.ParseAndVerifyToken(parts[1], h.jwtSecret)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeUnauthorized, "Invalid or expired token", nil)
	}
	return claims, nil
}

func (h *ComplianceHandler) extractUserType(claims *jwt.Claims) string {
	if claims == nil {
		return "CUSTOMER"
	}
	if claims.UserType != "" {
		return claims.UserType
	}
	if claims.AdminID != "" || claims.Role == "ADMIN" {
		return "ADMIN"
	}
	if claims.Role == "MANAGER" || claims.Role == "CASHIER" || claims.Role == "STAFF" || claims.Role == "SECURITY" {
		return "STAFF"
	}
	if claims.Role == "OWNER" || claims.ChainID != "" {
		return "CHAIN_HQ"
	}
	return "CUSTOMER"
}

// 1. Get IRN Record
func (h *ComplianceHandler) GetIRNRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	rec, err := h.repo.GetIRNRecordByOrderID(r.Context(), orderID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "IRN record not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rec)
}

// 2. List IRN Records
func (h *ComplianceHandler) ListIRNRecordsHandler(w http.ResponseWriter, r *http.Request) {
	storeID := r.URL.Query().Get("store_id")
	status := r.URL.Query().Get("status")

	records, err := h.repo.ListIRNRecords(r.Context(), storeID, status)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list IRN records", nil)
		return
	}

	if records == nil {
		records = []*IRNRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"records": records})
}

// 3. Retry IRN Manual
func (h *ComplianceHandler) RetryIRNRecordsHandler(w http.ResponseWriter, r *http.Request) {
	h.irnRetryJob.RunOnce(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Manual IRN retry triggered successfully"})
}

// 4. Submit Consent
func (h *ComplianceHandler) SubmitConsentHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req struct {
		ConsentType string `json:"consent_type"`
		Granted     bool   `json:"granted"`
		Version     string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	userType := h.extractUserType(claims)
	consent, err := h.consentSvc.RecordConsent(r.Context(), claims.UserID, userType, req.ConsentType, req.Granted, req.Version)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(consent)
}

// 4b. Update Single Consent Type (PUT /v1/compliance/consents/{type})
func (h *ComplianceHandler) UpdateSingleConsentHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	consentType := vars["type"]

	var body struct {
		Granted bool   `json:"granted"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	userType := h.extractUserType(claims)
	consent, err := h.consentSvc.RecordConsent(r.Context(), claims.UserID, userType, consentType, body.Granted, body.Version)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(consent)
}

// 5. List Consents
func (h *ComplianceHandler) GetConsentsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	consents, err := h.consentSvc.GetUserConsents(r.Context(), claims.UserID)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to get consents", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"consents": consents})
}

// 6. Create DPDP Request
func (h *ComplianceHandler) CreateDPDPRequestHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	var req struct {
		RequestType string `json:"request_type"`
		Detail      string `json:"detail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	userID := claims.UserID
	if userID == "" && claims.AdminID != "" {
		userID = claims.AdminID
	}

	userType := h.extractUserType(claims)
	dpdpReq, err := h.requestSvc.CreateRequest(r.Context(), userID, userType, req.RequestType, req.Detail)
	if err != nil {
		if strings.Contains(err.Error(), "ADMIN_DELETION_EXCLUDED") {
			errors.WriteError(w, http.StatusForbidden, "ADMIN_DELETION_EXCLUDED", err.Error(), nil)
		} else {
			errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dpdpReq)
}

// 7. List DPDP Requests
func (h *ComplianceHandler) ListDPDPRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	status := r.URL.Query().Get("status")
	reqType := r.URL.Query().Get("request_type")

	requests, err := h.repo.ListDPDPRequests(r.Context(), userID, status, reqType)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list DPDP requests", nil)
		return
	}

	if requests == nil {
		requests = []*DPDPRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"requests": requests})
}

// 7b. List Authenticated User's Own DPDP Requests (GET /v1/compliance/requests/mine)
func (h *ComplianceHandler) GetMyDPDPRequestsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Unauthorized", nil)
		return
	}

	requests, err := h.repo.ListDPDPRequests(r.Context(), claims.UserID, "", "")
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list user requests", nil)
		return
	}

	if requests == nil {
		requests = []*DPDPRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"requests": requests})
}

// 7c. Public Grievance Officer Contact Details (GET /v1/compliance/grievance-officer)
func (h *ComplianceHandler) GetGrievanceOfficerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"name":               "Nisha Sharma",
		"title":              "Data Protection & Grievance Officer",
		"email":              "grievance@zippyra.com",
		"address":            "Zippyra India Tech Pvt Ltd, 4th Floor, HSR Layout, Bengaluru 560102",
		"acknowledgment_sla": "72 hours",
	})
}

// 8. Review DPDP Request (Admin / Grievance Officer)
func (h *ComplianceHandler) ReviewDPDPRequestHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil || claims.Role != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Admin role required", nil)
		return
	}

	vars := mux.Vars(r)
	reqID := vars["id"]

	var body struct {
		Action string `json:"action"` // APPROVE | REJECT | IN_PROGRESS
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	updated, err := h.requestSvc.ReviewRequest(r.Context(), reqID, body.Action, claims.UserID, body.Reason)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

// 9. Process Deletion Request with Step-Up Requirement
func (h *ComplianceHandler) ProcessDeletionHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.extractClaims(r)
	if err != nil || claims.Role != "ADMIN" {
		errors.WriteError(w, http.StatusForbidden, errors.CodeForbidden, "Admin role required", nil)
		return
	}

	vars := mux.Vars(r)
	reqID := vars["id"]

	var body struct {
		StepUpVerified bool `json:"step_up_verified"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	updated, err := h.deletionProc.ProcessDeletionRequest(r.Context(), reqID, claims.UserID, body.StepUpVerified)
	if err != nil {
		if strings.Contains(err.Error(), "STEP_UP_REQUIRED") {
			errors.WriteError(w, http.StatusBadRequest, "STEP_UP_REQUIRED", err.Error(), nil)
			return
		}
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

// 10. Upsert Merchant KYC
func (h *ComplianceHandler) UpsertKYCHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storeID := vars["store_id"]
	if storeID == "" {
		storeID = r.URL.Query().Get("store_id")
	}

	var body struct {
		GSTIN             string `json:"gstin"`
		PANNumber         string `json:"pan_number"`
		BankAccountLast4  string `json:"bank_account_last4"`
		RazorpayAccountID string `json:"razorpay_account_id"`
		KYCStatus         string `json:"kyc_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "Invalid JSON payload", nil)
		return
	}

	kyc, err := h.kycSvc.SubmitKYC(r.Context(), storeID, body.GSTIN, body.PANNumber, body.BankAccountLast4, body.RazorpayAccountID)
	if err != nil {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, err.Error(), nil)
		return
	}

	if body.KYCStatus != "" {
		kyc.KYCStatus = body.KYCStatus
		if body.KYCStatus == "VERIFIED" && kyc.KYCCompletedAt == nil {
			now := time.Now().UTC()
			kyc.KYCCompletedAt = &now
		}
		_ = h.repo.UpsertMerchantKYC(r.Context(), kyc)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(kyc)
}

// 11. Get Merchant KYC
func (h *ComplianceHandler) GetKYCHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storeID := vars["store_id"]
	if storeID == "" {
		storeID = r.URL.Query().Get("store_id")
	}

	kyc, err := h.repo.GetMerchantKYC(r.Context(), storeID)
	if err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "KYC record not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(kyc)
}

// 12. List Incomplete KYC
func (h *ComplianceHandler) ListIncompleteKYCHandler(w http.ResponseWriter, r *http.Request) {
	kycs, err := h.repo.ListIncompleteKYC(r.Context())
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list incomplete KYCs", nil)
		return
	}

	if kycs == nil {
		kycs = []*MerchantKYC{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"kycs": kycs})
}

// 13. List Velocity Alerts
func (h *ComplianceHandler) ListVelocityAlertsHandler(w http.ResponseWriter, r *http.Request) {
	storeID := r.URL.Query().Get("store_id")
	unresolvedOnly := r.URL.Query().Get("unresolved_only") == "true" || r.URL.Query().Get("resolved") == "false"

	alerts, err := h.repo.ListVelocityAlerts(r.Context(), storeID, unresolvedOnly)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list velocity alerts", nil)
		return
	}

	if alerts == nil {
		alerts = []*VelocityAlert{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"alerts": alerts})
}

// 14. Resolve Velocity Alert
func (h *ComplianceHandler) ResolveVelocityAlertHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["id"]

	if err := h.repo.ResolveVelocityAlert(r.Context(), alertID); err != nil {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Velocity alert not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Alert resolved successfully"})
}

// 15. Run Reconciliation Job
func (h *ComplianceHandler) RunReconciliationHandler(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	report, err := h.reconJob.RunReconciliationForDate(r.Context(), dateStr)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

// 16. Get Reconciliation Report
func (h *ComplianceHandler) GetReconciliationReportHandler(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	if dateStr != "" {
		report, err := h.repo.GetSettlementReportByDate(r.Context(), dateStr)
		if err != nil {
			errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Report not found for date", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
		return
	}

	reports, err := h.repo.ListSettlementReports(r.Context(), dateFrom, dateTo)
	if err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to list reconciliation reports", nil)
		return
	}

	if reports == nil {
		reports = []*SettlementReport{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"reports": reports})
}
