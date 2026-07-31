package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zippyra/backend/shared/audit"
	"github.com/zippyra/backend/shared/errors"
	"github.com/zippyra/backend/shared/jwt"
)

type BulkImportHandler struct {
	repo              Repository
	catalogServiceURL string
	auditPub          *audit.Publisher
	httpClient        *http.Client
}

func NewBulkImportHandler(repo Repository, auditPub *audit.Publisher) *BulkImportHandler {
	catalogURL := os.Getenv("CATALOG_SERVICE_URL")
	if catalogURL == "" {
		catalogURL = "http://localhost:8011"
	}
	return &BulkImportHandler{
		repo:              repo,
		catalogServiceURL: catalogURL,
		auditPub:          auditPub,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *BulkImportHandler) getClaims(r *http.Request) *jwt.Claims {
	if val := r.Context().Value("user_claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	if val := r.Context().Value("claims"); val != nil {
		if c, ok := val.(*jwt.Claims); ok {
			return c
		}
	}
	return nil
}

func (h *BulkImportHandler) HandleBulkImport(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" || claims.ChainID == "" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorized CHAIN_HQ session required", nil)
		return
	}

	targetStoreIDs := []string{"store-001", "store-002"}
	perStoreJobs := make(map[string]string)

	for _, sID := range targetStoreIDs {
		// Mock trigger catalog import job per store
		jobID := fmt.Sprintf("import-job-%s-%s", sID, uuid.New().String()[:6])
		perStoreJobs[sID] = jobID
	}

	job := &ChainBulkImportJob{
		ChainID:        claims.ChainID,
		PerStoreJobIDs: perStoreJobs,
		Status:         "PROCESSING",
	}

	if err := h.repo.CreateBulkImportJob(r.Context(), job); err != nil {
		errors.WriteError(w, http.StatusInternalServerError, errors.CodeInternalError, "Failed to initiate bulk import job", nil)
		return
	}

	if h.auditPub != nil {
		h.auditPub.Publish(r.Context(), audit.AdminAuditEvent{
			ActorID:       claims.UserID,
			ActionType:    "chain_hq.bulk_import_started",
			TargetType:    "chain_bulk_import_job",
			TargetID:      job.ID,
			SourceService: "chain-hq-service",
			RequestID:     r.Header.Get("X-Request-ID"),
			Payload:       map[string]interface{}{"chain_id": claims.ChainID, "stores_count": len(targetStoreIDs)},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":             job.ID,
		"status":             job.Status,
		"per_store_job_ids":  perStoreJobs,
		"target_stores_count": len(targetStoreIDs),
	})
}

func (h *BulkImportHandler) HandleGetBulkImportStatus(w http.ResponseWriter, r *http.Request) {
	claims := h.getClaims(r)
	if claims == nil || claims.UserType != "CHAIN_HQ" {
		errors.WriteError(w, http.StatusUnauthorized, errors.CodeUnauthorized, "Authorized CHAIN_HQ session required", nil)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["id"]

	job, err := h.repo.GetBulkImportJob(r.Context(), jobID)
	if err != nil || job.ChainID != claims.ChainID {
		errors.WriteError(w, http.StatusNotFound, errors.CodeNotFound, "Bulk import job not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                job.ID,
		"chain_id":          job.ChainID,
		"status":            job.Status,
		"per_store_job_ids": job.PerStoreJobIDs,
		"summary":           "2 of 2 stores completed",
	})
}
