package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeatureFlagAdmin_CRUDAndStepUp(t *testing.T) {
	repo := NewMemoryRepository()
	kafkaAdmin := NewKafkaAdminClient("localhost:9092")
	handler := NewAuditHandler(repo, kafkaAdmin, "zippyra-dev-jwt-secret-key-32bytes", nil)
	routes := SetupRoutes(handler)

	token := generateAdminTestToken()

	// 1. Create Feature Flag
	createBody, _ := json.Marshal(map[string]interface{}{
		"flag_key":    "cart.dynamic_discounts",
		"description": "Enable dynamic cart rules engine",
		"scope_type":  "STORE",
	})
	reqCreate, _ := http.NewRequest("POST", "/v1/audit/feature-flags", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	routes.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for flag creation, got %d", rrCreate.Code)
	}

	// 2. Update High-Risk Flag without Step-Up Header -> 403 Forbidden
	createHighRiskBody, _ := json.Marshal(map[string]interface{}{
		"flag_key":    "payment.cashfree_fallback",
		"description": "Fallback payment gateway flag",
		"scope_type":  "GLOBAL",
	})
	reqHighRisk, _ := http.NewRequest("POST", "/v1/audit/feature-flags", bytes.NewBuffer(createHighRiskBody))
	reqHighRisk.Header.Set("Authorization", "Bearer "+token)
	reqHighRisk.Header.Set("Content-Type", "application/json")
	rrHighRisk := httptest.NewRecorder()
	routes.ServeHTTP(rrHighRisk, reqHighRisk)

	updateHighRiskBody, _ := json.Marshal(map[string]interface{}{
		"enabled_globally": true,
	})
	reqUpdateNoStepUp, _ := http.NewRequest("PUT", "/v1/audit/feature-flags/payment.cashfree_fallback", bytes.NewBuffer(updateHighRiskBody))
	reqUpdateNoStepUp.Header.Set("Authorization", "Bearer "+token)
	reqUpdateNoStepUp.Header.Set("Content-Type", "application/json")
	rrUpdateNoStepUp := httptest.NewRecorder()
	routes.ServeHTTP(rrUpdateNoStepUp, reqUpdateNoStepUp)

	if rrUpdateNoStepUp.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 STEP_UP_REQUIRED for high-risk flag without header, got %d", rrUpdateNoStepUp.Code)
	}

	// 3. Update High-Risk Flag WITH Step-Up Header -> 200 OK
	reqUpdateStepUp, _ := http.NewRequest("PUT", "/v1/audit/feature-flags/payment.cashfree_fallback", bytes.NewBuffer(updateHighRiskBody))
	reqUpdateStepUp.Header.Set("Authorization", "Bearer "+token)
	reqUpdateStepUp.Header.Set("X-StepUp-Token", "valid-stepup-session-token")
	reqUpdateStepUp.Header.Set("Content-Type", "application/json")
	rrUpdateStepUp := httptest.NewRecorder()
	routes.ServeHTTP(rrUpdateStepUp, reqUpdateStepUp)

	if rrUpdateStepUp.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for high-risk flag with step-up header, got %d", rrUpdateStepUp.Code)
	}
}
