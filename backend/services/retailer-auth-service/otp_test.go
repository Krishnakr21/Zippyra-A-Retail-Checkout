package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type SpySmsSender struct {
	calls int32
}

func (s *SpySmsSender) SendSMS(ctx context.Context, phone, code string) error {
	atomic.AddInt32(&s.calls, 1)
	return nil
}

func TestOTPSend_UnregisteredPhone_Returns403_NoSMS(t *testing.T) {
	repo := NewMemoryRepository()
	spySMS := &SpySmsSender{}
	otpSvc := NewOTPService(repo, nil, spySMS)
	authH := NewAuthHandler(repo, otpSvc, nil, "dev-secret")

	body, _ := json.Marshal(map[string]string{"phone": "+919876543210"})
	req := httptest.NewRequest("POST", "/v1/retailer-auth/otp/send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	authH.HandleSendOTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for unregistered phone, got %d", w.Code)
	}

	if atomic.LoadInt32(&spySMS.calls) != 0 {
		t.Fatalf("expected 0 SMS calls for unregistered phone, got %d", spySMS.calls)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != CodeStaffNotRegistered {
		t.Fatalf("expected error code %s, got %v", CodeStaffNotRegistered, errObj["code"])
	}
}

func TestOTPSend_DeactivatedStaff_ByteIdenticalResponseToUnregistered(t *testing.T) {
	repo := NewMemoryRepository()
	spySMS := &SpySmsSender{}
	otpSvc := NewOTPService(repo, nil, spySMS)
	authH := NewAuthHandler(repo, otpSvc, nil, "dev-secret")

	// 1. Unregistered phone response
	body1, _ := json.Marshal(map[string]string{"phone": "+919876543210"})
	req1 := httptest.NewRequest("POST", "/v1/retailer-auth/otp/send", bytes.NewReader(body1))
	w1 := httptest.NewRecorder()
	authH.HandleSendOTP(w1, req1)

	// 2. Register & Deactivate staff
	staff := &StaffMember{
		ID:       "staff-deactivated-1",
		StoreID:  "store-1",
		ChainID:  "chain-1",
		Phone:    "+919999988888",
		Name:     "Deactivated User",
		Role:     "CASHIER",
		IsActive: false,
	}
	_ = repo.CreateStaffMember(context.Background(), staff)
	_ = repo.DeactivateStaffMemberTx(context.Background(), staff.ID)

	body2, _ := json.Marshal(map[string]string{"phone": "+919999988888"})
	req2 := httptest.NewRequest("POST", "/v1/retailer-auth/otp/send", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	authH.HandleSendOTP(w2, req2)

	if w1.Code != w2.Code {
		t.Fatalf("expected status code match (%d vs %d)", w1.Code, w2.Code)
	}

	if atomic.LoadInt32(&spySMS.calls) != 0 {
		t.Fatalf("expected 0 SMS calls for deactivated phone, got %d", spySMS.calls)
	}

	// Verify error code equivalence
	var resp1, resp2 map[string]interface{}
	_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)

	err1 := resp1["error"].(map[string]interface{})
	err2 := resp2["error"].(map[string]interface{})

	if err1["code"] != err2["code"] || err1["message"] != err2["message"] {
		t.Fatalf("expected identical error response, got %v vs %v", err1, err2)
	}
}
