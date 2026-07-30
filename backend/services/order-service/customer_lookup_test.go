package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/zippyra/backend/shared/jwt"
)

func generateTestJWT(userType, role, storeID, userID string) string {
	secret := "sec"
	claims := jwt.Claims{
		UserID:   userID,
		UserType: userType,
		Role:     role,
		StoreID:  storeID,
	}
	token, _ := jwt.GenerateToken(&claims, secret, 1*time.Hour)
	return token
}

func TestCustomerLookup_SingleMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewOrderHandler(repo, NewMockRedisExitTokenService("sec"), NewMockInvoiceService(repo), "sec")

	// Seed single recent order
	order := &Order{
		ID:        "ord-lk-1",
		PaymentID: "pay-lk-1",
		UserID:    "user-9876541234",
		StoreID:   "store-lookup-1",
		Items:     []OrderItem{{Barcode: "111", Name: "Item 1", Qty: 1, PricePaise: 100}},
		SubtotalPaise: 100, TotalPaise: 100, PaymentMethod: "UPI", SupplyType: "INTRASTATE", Status: StatusCompleted,
		CreatedAt: time.Now(),
	}
	flags := []OrderItemReturnableFlag{{OrderID: order.ID, Barcode: "111", IsReturnable: true, ReturnedQty: 0}}
	_, _ = repo.CreateOrderAndOutboxTx(context.Background(), order, flags, NewMockRedisExitTokenService("sec"), TopicOrderCompleted, []byte("{}"))

	r := mux.NewRouter()
	relay := &OutboxRelay{db: db}
	RegisterRoutes(r, handler, relay)

	staffJWT := generateTestJWT("STAFF", "CASHIER", "store-lookup-1", "staff-1")

	req, _ := http.NewRequest(http.MethodGet, "/v1/order/internal/lookup-by-phone-last4?store_id=store-lookup-1&phone_last4=1234", nil)
	req.Header.Set("X-User-ID", "staff-1")
	req.Header.Set("Authorization", "Bearer "+staffJWT)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp CustomerLookupResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "SINGLE", resp.MatchType)
	assert.NotNil(t, resp.Customer)
	assert.Equal(t, "user-9876541234", resp.Customer.CustomerID)
	assert.Equal(t, "+91XXXXXX1234", resp.Customer.PhoneMasked)
}

func TestCustomerLookup_MultipleMatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewOrderHandler(repo, NewMockRedisExitTokenService("sec"), NewMockInvoiceService(repo), "sec")

	// Seed 2 orders from 2 different users matching last 4 digits "5555"
	order1 := &Order{
		ID:        "ord-lk-2",
		PaymentID: "pay-lk-2",
		UserID:    "user-1111115555",
		StoreID:   "store-lookup-2",
		Items:     []OrderItem{{Barcode: "111", Name: "Item 1", Qty: 1, PricePaise: 100}},
		SubtotalPaise: 100, TotalPaise: 100, PaymentMethod: "UPI", SupplyType: "INTRASTATE", Status: StatusCompleted,
		CreatedAt: time.Now(),
	}
	order2 := &Order{
		ID:        "ord-lk-3",
		PaymentID: "pay-lk-3",
		UserID:    "user-9999995555",
		StoreID:   "store-lookup-2",
		Items:     []OrderItem{{Barcode: "222", Name: "Item 2", Qty: 1, PricePaise: 200}},
		SubtotalPaise: 200, TotalPaise: 200, PaymentMethod: "UPI", SupplyType: "INTRASTATE", Status: StatusCompleted,
		CreatedAt: time.Now(),
	}

	flags := []OrderItemReturnableFlag{{OrderID: "ord-lk-2", Barcode: "111", IsReturnable: true, ReturnedQty: 0}}
	_, _ = repo.CreateOrderAndOutboxTx(context.Background(), order1, flags, NewMockRedisExitTokenService("sec"), TopicOrderCompleted, []byte("{}"))
	flags2 := []OrderItemReturnableFlag{{OrderID: "ord-lk-3", Barcode: "222", IsReturnable: true, ReturnedQty: 0}}
	_, _ = repo.CreateOrderAndOutboxTx(context.Background(), order2, flags2, NewMockRedisExitTokenService("sec"), TopicOrderCompleted, []byte("{}"))

	r := mux.NewRouter()
	relay := &OutboxRelay{db: db}
	RegisterRoutes(r, handler, relay)

	staffJWT := generateTestJWT("STAFF", "MANAGER", "store-lookup-2", "staff-mgr-1")

	req, _ := http.NewRequest(http.MethodGet, "/v1/order/internal/lookup-by-phone-last4?store_id=store-lookup-2&phone_last4=5555", nil)
	req.Header.Set("Authorization", "Bearer "+staffJWT)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp CustomerLookupResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "MULTIPLE", resp.MatchType)
	assert.Len(t, resp.Candidates, 2)
}

func TestCustomerLookup_NonStaffForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	handler := NewOrderHandler(repo, NewMockRedisExitTokenService("sec"), NewMockInvoiceService(repo), "sec")

	r := mux.NewRouter()
	relay := &OutboxRelay{db: db}
	RegisterRoutes(r, handler, relay)

	customerJWT := generateTestJWT("CUSTOMER", "CUSTOMER", "store-lookup-1", "cust-777")

	req, _ := http.NewRequest(http.MethodGet, "/v1/order/internal/lookup-by-phone-last4?store_id=store-lookup-1&phone_last4=1234", nil)
	req.Header.Set("Authorization", "Bearer "+customerJWT)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
