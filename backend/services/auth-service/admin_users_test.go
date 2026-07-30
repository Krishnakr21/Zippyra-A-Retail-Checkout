package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminListUsers_MasksContactInfo(t *testing.T) {
	memRepo := NewMemoryRepository()
	handler := NewAuthHandler(memRepo, nil, nil)

	phone := "+919876543210"
	email := "user@example.com"
	u1, _ := memRepo.CreateUserWithPhone(context.Background(), phone)
	memRepo.users[u1.ID].Email = &email

	req := httptest.NewRequest("GET", "/v1/auth/admin/users", nil)
	w := httptest.NewRecorder()
	handler.HandleAdminListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	usersList := resp["users"].([]interface{})
	if len(usersList) != 1 {
		t.Fatalf("expected 1 user in list, got %d", len(usersList))
	}

	uObj := usersList[0].(map[string]interface{})
	if uObj["phone_masked"] != "+91XXXXXX3210" {
		t.Fatalf("expected masked phone +91XXXXXX3210, got %v", uObj["phone_masked"])
	}
	if uObj["email_masked"] != "u***@example.com" {
		t.Fatalf("expected masked email u***@example.com, got %v", uObj["email_masked"])
	}
	if uObj["phone"] != nil && uObj["phone"] != "" {
		t.Fatalf("raw phone must not be present in list response")
	}
	if uObj["email"] != nil && uObj["email"] != "" {
		t.Fatalf("raw email must not be present in list response")
	}
}

func TestAdminGetUserDetail_ReturnsUnmaskedContactInfo(t *testing.T) {
	memRepo := NewMemoryRepository()
	handler := NewAuthHandler(memRepo, nil, nil)

	phone := "+919876543210"
	email := "user@example.com"
	u1, _ := memRepo.CreateUserWithPhone(context.Background(), phone)
	memRepo.users[u1.ID].Email = &email

	req := httptest.NewRequest("GET", "/v1/auth/admin/users/"+u1.ID, nil)
	w := httptest.NewRecorder()
	handler.HandleAdminGetUserDetail(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var uObj map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &uObj)

	if uObj["phone"] != "+919876543210" {
		t.Fatalf("expected raw phone +919876543210 on detail reveal, got %v", uObj["phone"])
	}
	if uObj["email"] != "user@example.com" {
		t.Fatalf("expected raw email user@example.com on detail reveal, got %v", uObj["email"])
	}
}
