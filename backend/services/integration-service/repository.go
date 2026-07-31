package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrConnectionNotFound = errors.New("connection not found")
var ErrSyncJobNotFound = errors.New("sync job not found")

type IntegrationRepository interface {
	CreateConnection(ctx context.Context, conn *ERPConnection) error
	GetConnectionByID(ctx context.Context, id string) (*ERPConnection, error)
	ListConnectionsByChain(ctx context.Context, chainID string) ([]*ERPConnection, error)
	UpdateConnection(ctx context.Context, conn *ERPConnection) error
	DeleteConnection(ctx context.Context, id string) error
	UpdateConnectionTimestamps(ctx context.Context, id string, lastInbound, lastOutbound, lastAgentPoll *time.Time, newStatus *ConnectionStatus) error

	CreateWebhookEvent(ctx context.Context, event *ERPWebhookEvent) (bool, error)
	UpdateWebhookEventResult(ctx context.Context, id string, result ProcessingResult, failureReason *string) error
	ListWebhookEvents(ctx context.Context, connectionID string, resultFilter *string) ([]*ERPWebhookEvent, error)

	CreateSyncJob(ctx context.Context, job *ERPSyncJob) (bool, error)
	GetSyncJobByID(ctx context.Context, id string) (*ERPSyncJob, error)
	ListPendingSyncJobs(ctx context.Context, connectionID string, limit int) ([]*ERPSyncJob, error)
	MarkSyncJobsDelivered(ctx context.Context, jobIDs []string) error
	MarkSyncJobsAcknowledged(ctx context.Context, connectionID string, jobIDs []string) error
	UpdateSyncJobStatus(ctx context.Context, id string, status SyncJobStatus, attemptCount int, failureReason *string) error
	ListSyncJobs(ctx context.Context, connectionID string, statusFilter *string, direction string) ([]*ERPSyncJob, error)
	ListFailedDirectSyncJobs(ctx context.Context, maxAttempts int) ([]*ERPSyncJob, error)
}

// In-Memory Repository Implementation for Unit Tests & Fallback
type MemoryIntegrationRepository struct {
	mu            sync.RWMutex
	connections   map[string]*ERPConnection
	webhookEvents map[string]*ERPWebhookEvent // key: connID + ":" + eventID
	syncJobs      map[string]*ERPSyncJob     // key: connID + ":" + sourceEventType + ":" + sourceEventID
}

func NewMemoryIntegrationRepository() *MemoryIntegrationRepository {
	return &MemoryIntegrationRepository{
		connections:   make(map[string]*ERPConnection),
		webhookEvents: make(map[string]*ERPWebhookEvent),
		syncJobs:      make(map[string]*ERPSyncJob),
	}
}

func (r *MemoryIntegrationRepository) CreateConnection(ctx context.Context, conn *ERPConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now
	r.connections[conn.ID] = conn
	return nil
}

func (r *MemoryIntegrationRepository) GetConnectionByID(ctx context.Context, id string) (*ERPConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.connections[id]
	if !exists || conn.Status == "DELETED" {
		return nil, ErrConnectionNotFound
	}
	return conn, nil
}

func (r *MemoryIntegrationRepository) ListConnectionsByChain(ctx context.Context, chainID string) ([]*ERPConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ERPConnection
	for _, conn := range r.connections {
		if conn.ChainID == chainID && conn.Status != "DELETED" {
			result = append(result, conn)
		}
	}
	return result, nil
}

func (r *MemoryIntegrationRepository) UpdateConnection(ctx context.Context, conn *ERPConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.connections[conn.ID]
	if !exists || existing.Status == "DELETED" {
		return ErrConnectionNotFound
	}
	conn.UpdatedAt = time.Now()
	r.connections[conn.ID] = conn
	return nil
}

func (r *MemoryIntegrationRepository) DeleteConnection(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[id]
	if !exists {
		return ErrConnectionNotFound
	}
	conn.Status = "DELETED"
	conn.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryIntegrationRepository) UpdateConnectionTimestamps(ctx context.Context, id string, lastInbound, lastOutbound, lastAgentPoll *time.Time, newStatus *ConnectionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, exists := r.connections[id]
	if !exists {
		return ErrConnectionNotFound
	}
	if lastInbound != nil {
		conn.LastInboundAt = lastInbound
	}
	if lastOutbound != nil {
		conn.LastOutboundAt = lastOutbound
	}
	if lastAgentPoll != nil {
		conn.LastAgentPollAt = lastAgentPoll
	}
	if newStatus != nil {
		conn.Status = *newStatus
	}
	conn.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryIntegrationRepository) CreateWebhookEvent(ctx context.Context, event *ERPWebhookEvent) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := event.ConnectionID + ":" + event.EventID
	if _, exists := r.webhookEvents[key]; exists {
		return false, nil // Already exists (DO NOTHING)
	}

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	event.CreatedAt = time.Now()
	r.webhookEvents[key] = event
	return true, nil
}

func (r *MemoryIntegrationRepository) UpdateWebhookEventResult(ctx context.Context, id string, result ProcessingResult, failureReason *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, ev := range r.webhookEvents {
		if ev.ID == id {
			ev.ProcessingResult = result
			ev.FailureReason = failureReason
			now := time.Now()
			ev.ProcessedAt = &now
			return nil
		}
	}
	return nil
}

func (r *MemoryIntegrationRepository) ListWebhookEvents(ctx context.Context, connectionID string, resultFilter *string) ([]*ERPWebhookEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ERPWebhookEvent
	for _, ev := range r.webhookEvents {
		if ev.ConnectionID == connectionID {
			if resultFilter != nil && string(ev.ProcessingResult) != *resultFilter {
				continue
			}
			result = append(result, ev)
		}
	}
	return result, nil
}

func (r *MemoryIntegrationRepository) CreateSyncJob(ctx context.Context, job *ERPSyncJob) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := job.ConnectionID + ":" + job.SourceEventType + ":" + job.SourceEventID
	if _, exists := r.syncJobs[key]; exists {
		return false, nil // Already exists (DO NOTHING)
	}

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	job.CreatedAt = time.Now()
	r.syncJobs[key] = job
	return true, nil
}

func (r *MemoryIntegrationRepository) GetSyncJobByID(ctx context.Context, id string) (*ERPSyncJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, job := range r.syncJobs {
		if job.ID == id {
			return job, nil
		}
	}
	return nil, ErrSyncJobNotFound
}

func (r *MemoryIntegrationRepository) ListPendingSyncJobs(ctx context.Context, connectionID string, limit int) ([]*ERPSyncJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ERPSyncJob
	for _, job := range r.syncJobs {
		if job.ConnectionID == connectionID && job.Status == SyncJobStatusPending {
			result = append(result, job)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *MemoryIntegrationRepository) MarkSyncJobsDelivered(ctx context.Context, jobIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	idMap := make(map[string]bool)
	for _, id := range jobIDs {
		idMap[id] = true
	}

	for _, job := range r.syncJobs {
		if idMap[job.ID] {
			job.Status = SyncJobStatusDelivered
			job.DeliveredAt = &now
		}
	}
	return nil
}

func (r *MemoryIntegrationRepository) MarkSyncJobsAcknowledged(ctx context.Context, connectionID string, jobIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	idMap := make(map[string]bool)
	for _, id := range jobIDs {
		idMap[id] = true
	}

	for _, job := range r.syncJobs {
		if job.ConnectionID == connectionID && idMap[job.ID] {
			job.Status = SyncJobStatusAcknowledged
			job.AcknowledgedAt = &now
		}
	}
	return nil
}

func (r *MemoryIntegrationRepository) UpdateSyncJobStatus(ctx context.Context, id string, status SyncJobStatus, attemptCount int, failureReason *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, job := range r.syncJobs {
		if job.ID == id {
			job.Status = status
			job.AttemptCount = attemptCount
			job.FailureReason = failureReason
			if status == SyncJobStatusDelivered {
				now := time.Now()
				job.DeliveredAt = &now
			}
			return nil
		}
	}
	return ErrSyncJobNotFound
}

func (r *MemoryIntegrationRepository) ListSyncJobs(ctx context.Context, connectionID string, statusFilter *string, direction string) ([]*ERPSyncJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ERPSyncJob
	for _, job := range r.syncJobs {
		if job.ConnectionID == connectionID {
			if statusFilter != nil && string(job.Status) != *statusFilter {
				continue
			}
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *MemoryIntegrationRepository) ListFailedDirectSyncJobs(ctx context.Context, maxAttempts int) ([]*ERPSyncJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ERPSyncJob
	for _, job := range r.syncJobs {
		conn, exists := r.connections[job.ConnectionID]
		if exists && conn.IntegrationMode == IntegrationModeDirect && job.Status == SyncJobStatusFailed && job.AttemptCount < maxAttempts {
			result = append(result, job)
		}
	}
	return result, nil
}

// Postgres Repository Implementation
type PostgresIntegrationRepository struct {
	db *sql.DB
}

func NewPostgresIntegrationRepository(db *sql.DB) *PostgresIntegrationRepository {
	return &PostgresIntegrationRepository{db: db}
}

func (r *PostgresIntegrationRepository) CreateConnection(ctx context.Context, conn *ERPConnection) error {
	eventsJSON, _ := json.Marshal(conn.EnabledOutboundEvents)
	query := `
		INSERT INTO erp_connections (
			id, chain_id, erp_type, integration_mode, display_name,
			inbound_webhook_secret_encrypted, agent_api_key_hash, outbound_config_encrypted,
			enabled_outbound_events, status, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
	`
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.ChainID, conn.ERPType, conn.IntegrationMode, conn.DisplayName,
		conn.InboundWebhookSecretEncrypted, conn.AgentAPIKeyHash, conn.OutboundConfigEncrypted,
		eventsJSON, conn.Status, conn.CreatedBy,
	)
	return err
}

func (r *PostgresIntegrationRepository) GetConnectionByID(ctx context.Context, id string) (*ERPConnection, error) {
	query := `
		SELECT id, chain_id, erp_type, integration_mode, display_name,
		       inbound_webhook_secret_encrypted, agent_api_key_hash, outbound_config_encrypted,
		       enabled_outbound_events, status, last_inbound_at, last_outbound_at, last_agent_poll_at,
		       created_by, created_at, updated_at
		FROM erp_connections WHERE id = $1 AND status != 'DELETED'
	`
	row := r.db.QueryRowContext(ctx, query, id)
	var conn ERPConnection
	var eventsRaw []byte
	err := row.Scan(
		&conn.ID, &conn.ChainID, &conn.ERPType, &conn.IntegrationMode, &conn.DisplayName,
		&conn.InboundWebhookSecretEncrypted, &conn.AgentAPIKeyHash, &conn.OutboundConfigEncrypted,
		&eventsRaw, &conn.Status, &conn.LastInboundAt, &conn.LastOutboundAt, &conn.LastAgentPollAt,
		&conn.CreatedBy, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(eventsRaw, &conn.EnabledOutboundEvents)
	return &conn, nil
}

func (r *PostgresIntegrationRepository) ListConnectionsByChain(ctx context.Context, chainID string) ([]*ERPConnection, error) {
	query := `
		SELECT id, chain_id, erp_type, integration_mode, display_name,
		       inbound_webhook_secret_encrypted, agent_api_key_hash, outbound_config_encrypted,
		       enabled_outbound_events, status, last_inbound_at, last_outbound_at, last_agent_poll_at,
		       created_by, created_at, updated_at
		FROM erp_connections WHERE chain_id = $1 AND status != 'DELETED'
	`
	rows, err := r.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ERPConnection
	for rows.Next() {
		var conn ERPConnection
		var eventsRaw []byte
		if err := rows.Scan(
			&conn.ID, &conn.ChainID, &conn.ERPType, &conn.IntegrationMode, &conn.DisplayName,
			&conn.InboundWebhookSecretEncrypted, &conn.AgentAPIKeyHash, &conn.OutboundConfigEncrypted,
			&eventsRaw, &conn.Status, &conn.LastInboundAt, &conn.LastOutboundAt, &conn.LastAgentPollAt,
			&conn.CreatedBy, &conn.CreatedAt, &conn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(eventsRaw, &conn.EnabledOutboundEvents)
		list = append(list, &conn)
	}
	return list, nil
}

func (r *PostgresIntegrationRepository) UpdateConnection(ctx context.Context, conn *ERPConnection) error {
	eventsJSON, _ := json.Marshal(conn.EnabledOutboundEvents)
	query := `
		UPDATE erp_connections
		SET display_name = $1, enabled_outbound_events = $2, status = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND status != 'DELETED'
	`
	res, err := r.db.ExecContext(ctx, query, conn.DisplayName, eventsJSON, conn.Status, conn.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func (r *PostgresIntegrationRepository) DeleteConnection(ctx context.Context, id string) error {
	query := `UPDATE erp_connections SET status = 'DELETED', updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func (r *PostgresIntegrationRepository) UpdateConnectionTimestamps(ctx context.Context, id string, lastInbound, lastOutbound, lastAgentPoll *time.Time, newStatus *ConnectionStatus) error {
	query := `
		UPDATE erp_connections
		SET last_inbound_at = COALESCE($1, last_inbound_at),
		    last_outbound_at = COALESCE($2, last_outbound_at),
		    last_agent_poll_at = COALESCE($3, last_agent_poll_at),
		    status = COALESCE($4, status),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	var s *string
	if newStatus != nil {
		str := string(*newStatus)
		s = &str
	}
	_, err := r.db.ExecContext(ctx, query, lastInbound, lastOutbound, lastAgentPoll, s, id)
	return err
}

func (r *PostgresIntegrationRepository) CreateWebhookEvent(ctx context.Context, event *ERPWebhookEvent) (bool, error) {
	query := `
		INSERT INTO erp_webhook_events (id, connection_id, event_id, event_type, raw_payload, processing_result, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (connection_id, event_id) DO NOTHING
	`
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	res, err := r.db.ExecContext(ctx, query, event.ID, event.ConnectionID, event.EventID, event.EventType, event.RawPayload, event.ProcessingResult, event.FailureReason)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *PostgresIntegrationRepository) UpdateWebhookEventResult(ctx context.Context, id string, result ProcessingResult, failureReason *string) error {
	query := `UPDATE erp_webhook_events SET processing_result = $1, failure_reason = $2, processed_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, result, failureReason, id)
	return err
}

func (r *PostgresIntegrationRepository) ListWebhookEvents(ctx context.Context, connectionID string, resultFilter *string) ([]*ERPWebhookEvent, error) {
	query := `SELECT id, connection_id, event_id, event_type, raw_payload, processing_result, failure_reason, processed_at, created_at FROM erp_webhook_events WHERE connection_id = $1`
	var args []interface{}
	args = append(args, connectionID)
	if resultFilter != nil {
		query += ` AND processing_result = $2`
		args = append(args, *resultFilter)
	}
	query += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ERPWebhookEvent
	for rows.Next() {
		var ev ERPWebhookEvent
		if err := rows.Scan(&ev.ID, &ev.ConnectionID, &ev.EventID, &ev.EventType, &ev.RawPayload, &ev.ProcessingResult, &ev.FailureReason, &ev.ProcessedAt, &ev.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &ev)
	}
	return list, nil
}

func (r *PostgresIntegrationRepository) CreateSyncJob(ctx context.Context, job *ERPSyncJob) (bool, error) {
	query := `
		INSERT INTO erp_sync_jobs (id, connection_id, direction, source_event_type, source_event_id, payload, status, attempt_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		ON CONFLICT (connection_id, source_event_type, source_event_id) DO NOTHING
	`
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	res, err := r.db.ExecContext(ctx, query, job.ID, job.ConnectionID, job.Direction, job.SourceEventType, job.SourceEventID, job.Payload, job.Status, job.AttemptCount)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *PostgresIntegrationRepository) GetSyncJobByID(ctx context.Context, id string) (*ERPSyncJob, error) {
	query := `SELECT id, connection_id, direction, source_event_type, source_event_id, payload, status, attempt_count, failure_reason, created_at, delivered_at, acknowledged_at FROM erp_sync_jobs WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var job ERPSyncJob
	err := row.Scan(&job.ID, &job.ConnectionID, &job.Direction, &job.SourceEventType, &job.SourceEventID, &job.Payload, &job.Status, &job.AttemptCount, &job.FailureReason, &job.CreatedAt, &job.DeliveredAt, &job.AcknowledgedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSyncJobNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (r *PostgresIntegrationRepository) ListPendingSyncJobs(ctx context.Context, connectionID string, limit int) ([]*ERPSyncJob, error) {
	query := `SELECT id, connection_id, direction, source_event_type, source_event_id, payload, status, attempt_count, failure_reason, created_at, delivered_at, acknowledged_at FROM erp_sync_jobs WHERE connection_id = $1 AND status = 'PENDING' ORDER BY created_at ASC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, connectionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ERPSyncJob
	for rows.Next() {
		var job ERPSyncJob
		if err := rows.Scan(&job.ID, &job.ConnectionID, &job.Direction, &job.SourceEventType, &job.SourceEventID, &job.Payload, &job.Status, &job.AttemptCount, &job.FailureReason, &job.CreatedAt, &job.DeliveredAt, &job.AcknowledgedAt); err != nil {
			return nil, err
		}
		list = append(list, &job)
	}
	return list, nil
}

func (r *PostgresIntegrationRepository) MarkSyncJobsDelivered(ctx context.Context, jobIDs []string) error {
	query := `UPDATE erp_sync_jobs SET status = 'DELIVERED', delivered_at = CURRENT_TIMESTAMP WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, jobIDs)
	return err
}

func (r *PostgresIntegrationRepository) MarkSyncJobsAcknowledged(ctx context.Context, connectionID string, jobIDs []string) error {
	query := `UPDATE erp_sync_jobs SET status = 'ACKNOWLEDGED', acknowledged_at = CURRENT_TIMESTAMP WHERE connection_id = $1 AND id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, connectionID, jobIDs)
	return err
}

func (r *PostgresIntegrationRepository) UpdateSyncJobStatus(ctx context.Context, id string, status SyncJobStatus, attemptCount int, failureReason *string) error {
	query := `UPDATE erp_sync_jobs SET status = $1, attempt_count = $2, failure_reason = $3, delivered_at = CASE WHEN $1 = 'DELIVERED' THEN CURRENT_TIMESTAMP ELSE delivered_at END WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, status, attemptCount, failureReason, id)
	return err
}

func (r *PostgresIntegrationRepository) ListSyncJobs(ctx context.Context, connectionID string, statusFilter *string, direction string) ([]*ERPSyncJob, error) {
	query := `SELECT id, connection_id, direction, source_event_type, source_event_id, payload, status, attempt_count, failure_reason, created_at, delivered_at, acknowledged_at FROM erp_sync_jobs WHERE connection_id = $1`
	var args []interface{}
	args = append(args, connectionID)
	if statusFilter != nil {
		query += ` AND status = $2`
		args = append(args, *statusFilter)
	}
	query += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ERPSyncJob
	for rows.Next() {
		var job ERPSyncJob
		if err := rows.Scan(&job.ID, &job.ConnectionID, &job.Direction, &job.SourceEventType, &job.SourceEventID, &job.Payload, &job.Status, &job.AttemptCount, &job.FailureReason, &job.CreatedAt, &job.DeliveredAt, &job.AcknowledgedAt); err != nil {
			return nil, err
		}
		list = append(list, &job)
	}
	return list, nil
}

func (r *PostgresIntegrationRepository) ListFailedDirectSyncJobs(ctx context.Context, maxAttempts int) ([]*ERPSyncJob, error) {
	query := `
		SELECT j.id, j.connection_id, j.direction, j.source_event_type, j.source_event_id, j.payload, j.status, j.attempt_count, j.failure_reason, j.created_at, j.delivered_at, j.acknowledged_at
		FROM erp_sync_jobs j
		JOIN erp_connections c ON j.connection_id = c.id
		WHERE c.integration_mode = 'DIRECT' AND j.status = 'FAILED' AND j.attempt_count < $1
	`
	rows, err := r.db.QueryContext(ctx, query, maxAttempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ERPSyncJob
	for rows.Next() {
		var job ERPSyncJob
		if err := rows.Scan(&job.ID, &job.ConnectionID, &job.Direction, &job.SourceEventType, &job.SourceEventID, &job.Payload, &job.Status, &job.AttemptCount, &job.FailureReason, &job.CreatedAt, &job.DeliveredAt, &job.AcknowledgedAt); err != nil {
			return nil, err
		}
		list = append(list, &job)
	}
	return list, nil
}
