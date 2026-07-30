package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type AccessFulfillmentHandler struct {
	manager *AccessExportManager
}

func NewAccessFulfillmentHandler(manager *AccessExportManager) *AccessFulfillmentHandler {
	return &AccessFulfillmentHandler{
		manager: manager,
	}
}

// POST /v1/compliance/requests/{id}/process-access (ADMIN JWT)
func (h *AccessFulfillmentHandler) HandleProcessAccessRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	userType := extractUserTypeFromReq(r)
	if userType != "ADMIN" && userType != "SUPER_ADMIN" {
		writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "Only ADMIN role can initiate DPDP access export fulfillment")
		return
	}

	export, err := h.manager.ProcessAccessRequest(r.Context(), requestID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "PROCESS_FAILED", err.Error())
		return
	}

	writeJSONResponse(w, http.StatusAccepted, map[string]interface{}{
		"message":           "DPDP access export assembly initiated",
		"export_id":         export.ID,
		"dpdp_request_id":   export.DPDPRequestID,
		"status":            export.Status,
		"expected_services": export.ExpectedServices,
	})
}

// GET /v1/compliance/access-exports/{dpdp_request_id}/download (CUSTOMER / STAFF / CHAIN_HQ JWT)
func (h *AccessFulfillmentHandler) HandleDownloadAccessExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["dpdp_request_id"]

	requestingUserID := extractUserID(r.Context())
	if requestingUserID == "" {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing user authentication claims")
		return
	}

	export, err := h.manager.GetAccessExportForDownload(r.Context(), requestID, requestingUserID, r)
	if err != nil {
		switch err {
		case ErrExportNotYours:
			writeJSONError(w, http.StatusForbidden, "EXPORT_NOT_YOURS", "You are not authorized to download this data export")
		case ErrExportNotReady:
			writeJSONError(w, http.StatusBadRequest, "EXPORT_NOT_READY", "Data export is still assembling. Please try again shortly.")
		case ErrExportExpired:
			writeJSONError(w, http.StatusNotFound, "EXPORT_EXPIRED", "Data export has expired after 7 days. Please submit a new access request.")
		default:
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	if export.DownloadURL == nil {
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Download URL generation failed")
		return
	}

	// 302 Redirect to presigned S3 download URL
	http.Redirect(w, r, *export.DownloadURL, http.StatusFound)
}

func extractUserTypeFromReq(r *http.Request) string {
	role := r.Header.Get("X-User-Role")
	if role != "" {
		return strings.ToUpper(role)
	}
	return "CUSTOMER"
}

func extractUserID(ctx interface{}) string {
	if val := ctx; val != nil {
		if claims, ok := val.(map[string]interface{}); ok {
			if uid, ok := claims["user_id"].(string); ok {
				return uid
			}
			if sub, ok := claims["sub"].(string); ok {
				return sub
			}
		}
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
