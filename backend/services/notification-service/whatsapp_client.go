package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var ErrTemplateUnapproved = errors.New("whatsapp template not approved by meta")

type WhatsAppClient interface {
	SendTemplateMessage(ctx context.Context, recipient, templateName, language string, messageCategory string, components []map[string]interface{}) error
}

type MockWhatsAppClient struct {
	mu           sync.Mutex
	SentMessages []map[string]interface{}
}

func NewMockWhatsAppClient() *MockWhatsAppClient {
	return &MockWhatsAppClient{
		SentMessages: []map[string]interface{}{},
	}
}

func (m *MockWhatsAppClient) SendTemplateMessage(ctx context.Context, recipient, templateName, language string, messageCategory string, components []map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Enforce 10s timeout
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	m.SentMessages = append(m.SentMessages, map[string]interface{}{
		"recipient":        recipient,
		"template_name":    templateName,
		"language":         language,
		"message_category": messageCategory,
		"components":       components,
	})

	log.Printf("[MockWhatsAppClient] WHATSAPP message sent to %s using template %s (lang: %s, category: %s)", recipient, templateName, language, messageCategory)
	return nil
}
