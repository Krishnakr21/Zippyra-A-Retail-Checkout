package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/jwt"
)

func generateAdminTestToken() string {
	claims := &jwt.Claims{
		UserID:   uuid.New().String(),
		Email:    "admin@zippyra.com",
		Role:     "ADMIN",
		UserType: "STAFF",
	}
	token, _ := jwt.GenerateToken(claims, "zippyra-dev-jwt-secret-key-32bytes", 30*time.Minute)
	return token
}

func TestDLQ_ListAndPeekMessages(t *testing.T) {
	repo := NewMemoryRepository()
	kafkaAdmin := NewKafkaAdminClient("localhost:9092")
	kafkaAdmin.SeedMockDLQMessage("payment.confirmed.dlq", 101, "pay-100", `{"payment_id":"pay-100","error":"timeout"}`)
	kafkaAdmin.SeedMockDLQMessage("payment.confirmed.dlq", 102, "pay-101", `{"payment_id":"pay-101","error":"gateway error"}`)

	handler := NewAuditHandler(repo, kafkaAdmin, "zippyra-dev-jwt-secret-key-32bytes", nil)
	routes := SetupRoutes(handler)

	token := generateAdminTestToken()

	// 1. List DLQ topics
	req, _ := http.NewRequest("GET", "/v1/audit/kafka/dlq-topics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	routes.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var listRes map[string][]DLQTopicSummary
	_ = json.NewDecoder(rr.Body).Decode(&listRes)
	if len(listRes["dlq_topics"]) != 1 {
		t.Fatalf("Expected 1 DLQ topic, got %d", len(listRes["dlq_topics"]))
	}

	// 2. Peek DLQ messages
	reqPeek, _ := http.NewRequest("GET", "/v1/audit/kafka/dlq-topics/payment.confirmed.dlq/messages?limit=10", nil)
	reqPeek.Header.Set("Authorization", "Bearer "+token)
	rrPeek := httptest.NewRecorder()
	routes.ServeHTTP(rrPeek, reqPeek)

	if rrPeek.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for peek, got %d", rrPeek.Code)
	}

	var peekRes map[string]interface{}
	_ = json.NewDecoder(rrPeek.Body).Decode(&peekRes)
	msgs := peekRes["messages"].([]interface{})
	if len(msgs) != 2 {
		t.Fatalf("Expected 2 DLQ messages, got %d", len(msgs))
	}
}

func TestDLQ_ReplayAndSoftDiscard(t *testing.T) {
	repo := NewMemoryRepository()
	kafkaAdmin := NewKafkaAdminClient("localhost:9092")
	kafkaAdmin.SeedMockDLQMessage("order.completed.dlq", 201, "ord-1", `{"order_id":"ord-1"}`)
	kafkaAdmin.SeedMockDLQMessage("order.completed.dlq", 202, "ord-2", `{"order_id":"ord-2"}`)

	handler := NewAuditHandler(repo, kafkaAdmin, "zippyra-dev-jwt-secret-key-32bytes", nil)
	routes := SetupRoutes(handler)

	token := generateAdminTestToken()

	// 1. Replay offset 201
	replayBody, _ := json.Marshal(map[string]interface{}{"offsets": []int64{201}})
	reqReplay, _ := http.NewRequest("POST", "/v1/audit/kafka/dlq-topics/order.completed.dlq/replay", bytes.NewBuffer(replayBody))
	reqReplay.Header.Set("Authorization", "Bearer "+token)
	reqReplay.Header.Set("Content-Type", "application/json")
	rrReplay := httptest.NewRecorder()
	routes.ServeHTTP(rrReplay, reqReplay)

	if rrReplay.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for replay, got %d", rrReplay.Code)
	}

	var replayRes map[string]interface{}
	_ = json.NewDecoder(rrReplay.Body).Decode(&replayRes)
	if int(replayRes["replayed_count"].(float64)) != 1 {
		t.Fatalf("Expected 1 replayed message, got %v", replayRes["replayed_count"])
	}

	// 2. Soft-discard offset 202
	discardBody, _ := json.Marshal(map[string]interface{}{"offsets": []int64{202}, "reason": "unrecoverable corruption"})
	reqDiscard, _ := http.NewRequest("DELETE", "/v1/audit/kafka/dlq-topics/order.completed.dlq/messages", bytes.NewBuffer(discardBody))
	reqDiscard.Header.Set("Authorization", "Bearer "+token)
	reqDiscard.Header.Set("Content-Type", "application/json")
	rrDiscard := httptest.NewRecorder()
	routes.ServeHTTP(rrDiscard, reqDiscard)

	if rrDiscard.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for discard, got %d", rrDiscard.Code)
	}

	// 3. Peek again -> offset 202 should be filtered out
	reqPeek, _ := http.NewRequest("GET", "/v1/audit/kafka/dlq-topics/order.completed.dlq/messages?limit=10", nil)
	reqPeek.Header.Set("Authorization", "Bearer "+token)
	rrPeek := httptest.NewRecorder()
	routes.ServeHTTP(rrPeek, reqPeek)

	var peekRes map[string]interface{}
	_ = json.NewDecoder(rrPeek.Body).Decode(&peekRes)
	msgs := peekRes["messages"].([]interface{})
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message remaining after soft discard, got %d", len(msgs))
	}
}
