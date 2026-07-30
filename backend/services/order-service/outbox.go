package main

import (
	"time"
)

type OrderOutboxEvent struct {
	ID          string     `json:"id"`
	Topic       string     `json:"topic"`
	Payload     []byte     `json:"payload"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
}
