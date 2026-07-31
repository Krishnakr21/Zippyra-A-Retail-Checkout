package main

import (
	"encoding/json"
	"time"
)

type AdminAction struct {
	ID            string                 `json:"id"`
	ActorID       string                 `json:"actor_id"`
	ActorName     string                 `json:"actor_name"`
	ActionType    string                 `json:"action_type"`
	TargetType    string                 `json:"target_type"`
	TargetID      string                 `json:"target_id"`
	Payload       map[string]interface{} `json:"payload"`
	SourceService string                 `json:"source_service"`
	RequestID     string                 `json:"request_id"`
	CreatedAt     time.Time              `json:"created_at"`
}

type AdminActionDB struct {
	ID            string          `json:"id"`
	ActorID       string          `json:"actor_id"`
	ActorName     string          `json:"actor_name"`
	ActionType    string          `json:"action_type"`
	TargetType    string          `json:"target_type"`
	TargetID      string          `json:"target_id"`
	PayloadRaw    json.RawMessage `json:"payload"`
	SourceService string          `json:"source_service"`
	RequestID     string          `json:"request_id"`
	CreatedAt     time.Time       `json:"created_at"`
}

type AuditFilter struct {
	ActorID    string
	ActionType string
	TargetType string
	DateFrom   string
	DateTo     string
	Page       int
	PageSize   int
}
