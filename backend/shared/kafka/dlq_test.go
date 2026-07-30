package kafka

import (
	"context"
	"errors"
	"testing"
)

func TestConsumerDLQ_FailureRoutesToDLQTopic(t *testing.T) {
	consumer := NewConsumer("localhost:9092", "test-group", []string{"order.completed"})

	handlerCalled := false
	consumer.RegisterHandler("order.completed", func(ctx context.Context, value []byte) error {
		handlerCalled = true
		return errors.New("simulated database transaction failure")
	})

	ctx := context.Background()
	err := consumer.HandleMessageWithDLQ(ctx, "order.completed", "ord-123", []byte(`{"order_id":"ord-123"}`))

	if !handlerCalled {
		t.Fatalf("Expected handler to be invoked")
	}
	if err == nil {
		t.Fatalf("Expected error from HandleMessageWithDLQ due to handler failure")
	}
	if consumer.dlqHandled != 1 {
		t.Fatalf("Expected dlqHandled count to be 1, got %d", consumer.dlqHandled)
	}
}
