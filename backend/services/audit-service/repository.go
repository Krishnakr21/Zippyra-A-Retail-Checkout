package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/featureflags"
)

type Repository interface {
	CreateAction(ctx context.Context, action *AdminAction) error
	ListActions(ctx context.Context, filter AuditFilter) ([]*AdminAction, int64, error)

	// DLQ Discarded Offsets
	DiscardOffsets(ctx context.Context, topic string, offsets []int64, userID uuid.UUID, reason string) error
	GetDiscardedOffsets(ctx context.Context, topic string) (map[int64]bool, error)

	// Feature Flags
	SaveFeatureFlag(ctx context.Context, flag *featureflags.FeatureFlag) error
	GetFeatureFlag(ctx context.Context, key string) (*featureflags.FeatureFlag, error)
	ListFeatureFlags(ctx context.Context) ([]*featureflags.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, key string) error
}

type PostgresRepository struct {
	database *db.DB
}

func NewPostgresRepository(database *db.DB) Repository {
	return &PostgresRepository{database: database}
}

func (p *PostgresRepository) CreateAction(ctx context.Context, action *AdminAction) error {
	if action.ID == "" {
		action.ID = uuid.New().String()
	}
	if action.CreatedAt.IsZero() {
		action.CreatedAt = time.Now()
	}

	payloadJSON, _ := json.Marshal(action.Payload)

	query := `
		INSERT INTO admin_actions (id, actor_id, actor_name, action_type, target_type, target_id, payload, source_service, request_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (source_service, request_id, action_type) DO NOTHING
	`
	_, err := p.database.ExecContext(ctx, query,
		action.ID, action.ActorID, action.ActorName, action.ActionType, action.TargetType, action.TargetID,
		payloadJSON, action.SourceService, action.RequestID, action.CreatedAt,
	)
	return err
}

func (p *PostgresRepository) ListActions(ctx context.Context, filter AuditFilter) ([]*AdminAction, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	where := "1=1"
	args := []interface{}{}

	if filter.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", len(args)+1)
		args = append(args, filter.ActorID)
	}
	if filter.ActionType != "" {
		where += fmt.Sprintf(" AND action_type = $%d", len(args)+1)
		args = append(args, filter.ActionType)
	}
	if filter.TargetType != "" {
		where += fmt.Sprintf(" AND target_type = $%d", len(args)+1)
		args = append(args, filter.TargetType)
	}
	if filter.DateFrom != "" {
		where += fmt.Sprintf(" AND created_at >= $%d::timestamptz", len(args)+1)
		args = append(args, filter.DateFrom)
	}
	if filter.DateTo != "" {
		where += fmt.Sprintf(" AND created_at <= $%d::timestamptz", len(args)+1)
		args = append(args, filter.DateTo)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_actions WHERE %s", where)
	if err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, filter.PageSize, offset)

	query := fmt.Sprintf(`
		SELECT id, actor_id, COALESCE(actor_name, ''), action_type, COALESCE(target_type, ''), COALESCE(target_id, ''),
		       COALESCE(payload, '{}'::jsonb), source_service, COALESCE(request_id, ''), created_at
		FROM admin_actions
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArg, offsetArg)

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var actions []*AdminAction
	for rows.Next() {
		var a AdminAction
		var payloadJSON []byte
		if err := rows.Scan(
			&a.ID, &a.ActorID, &a.ActorName, &a.ActionType, &a.TargetType, &a.TargetID,
			&payloadJSON, &a.SourceService, &a.RequestID, &a.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if len(payloadJSON) > 0 {
			_ = json.Unmarshal(payloadJSON, &a.Payload)
		}
		actions = append(actions, &a)
	}

	if actions == nil {
		actions = []*AdminAction{}
	}

	return actions, total, nil
}

// MemoryRepository for unit testing
type MemoryRepository struct {
	mu               sync.RWMutex
	actions          map[string]*AdminAction // key: source_service + ":" + request_id + ":" + action_type
	discardedOffsets map[string]map[int64]bool
	flags            map[string]*featureflags.FeatureFlag
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		actions: make(map[string]*AdminAction),
	}
}

func (m *MemoryRepository) CreateAction(ctx context.Context, action *AdminAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", action.SourceService, action.RequestID, action.ActionType)
	if _, exists := m.actions[key]; exists {
		return nil // Idempotent skip
	}

	if action.ID == "" {
		action.ID = uuid.New().String()
	}
	if action.CreatedAt.IsZero() {
		action.CreatedAt = time.Now()
	}

	m.actions[key] = action
	return nil
}

func (m *MemoryRepository) ListActions(ctx context.Context, filter AuditFilter) ([]*AdminAction, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*AdminAction
	for _, a := range m.actions {
		if filter.ActorID != "" && a.ActorID != filter.ActorID {
			continue
		}
		if filter.ActionType != "" && a.ActionType != filter.ActionType {
			continue
		}
		if filter.TargetType != "" && a.TargetType != filter.TargetType {
			continue
		}
		matched = append(matched, a)
	}

	total := int64(len(matched))
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize
	if offset >= len(matched) {
		return []*AdminAction{}, total, nil
	}

	end := offset + filter.PageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

// PostgresRepository DLQ & Feature Flag methods
func (p *PostgresRepository) DiscardOffsets(ctx context.Context, topic string, offsets []int64, userID uuid.UUID, reason string) error {
	query := `
		INSERT INTO dlq_discarded_offsets (topic, offset, discarded_by, discarded_at, reason)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (topic, offset) DO NOTHING
	`
	for _, offset := range offsets {
		_, err := p.database.ExecContext(ctx, query, topic, offset, userID, reason)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresRepository) GetDiscardedOffsets(ctx context.Context, topic string) (map[int64]bool, error) {
	query := `SELECT offset FROM dlq_discarded_offsets WHERE topic = $1`
	rows, err := p.database.QueryContext(ctx, query, topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]bool)
	for rows.Next() {
		var off int64
		if err := rows.Scan(&off); err == nil {
			result[off] = true
		}
	}
	return result, nil
}

func (p *PostgresRepository) SaveFeatureFlag(ctx context.Context, flag *featureflags.FeatureFlag) error {
	return featureflags.SetFlag(ctx, p.database.DB, nil, flag)
}

func (p *PostgresRepository) GetFeatureFlag(ctx context.Context, key string) (*featureflags.FeatureFlag, error) {
	flags, err := p.ListFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}
	for _, f := range flags {
		if f.FlagKey == key {
			return f, nil
		}
	}
	return nil, nil
}

func (p *PostgresRepository) ListFeatureFlags(ctx context.Context) ([]*featureflags.FeatureFlag, error) {
	query := `SELECT flag_key, description, scope_type, enabled_globally, enabled_scope_ids, user_percentage, updated_by, updated_at, created_at FROM feature_flags ORDER BY flag_key`
	rows, err := p.database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*featureflags.FeatureFlag
	for rows.Next() {
		var f featureflags.FeatureFlag
		var scopeIDsJSON []byte
		var userPct *int
		if err := rows.Scan(&f.FlagKey, &f.Description, &f.ScopeType, &f.EnabledGlobally, &scopeIDsJSON, &userPct, &f.UpdatedBy, &f.UpdatedAt, &f.CreatedAt); err != nil {
			return nil, err
		}
		if len(scopeIDsJSON) > 0 {
			_ = json.Unmarshal(scopeIDsJSON, &f.EnabledScopeIDs)
		}
		f.UserPercentage = userPct
		flags = append(flags, &f)
	}
	return flags, nil
}

func (p *PostgresRepository) DeleteFeatureFlag(ctx context.Context, key string) error {
	return featureflags.DeleteFlag(ctx, p.database.DB, nil, key)
}

// MemoryRepository extensions
func (m *MemoryRepository) DiscardOffsets(ctx context.Context, topic string, offsets []int64, userID uuid.UUID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.discardedOffsets == nil {
		m.discardedOffsets = make(map[string]map[int64]bool)
	}
	if m.discardedOffsets[topic] == nil {
		m.discardedOffsets[topic] = make(map[int64]bool)
	}
	for _, off := range offsets {
		m.discardedOffsets[topic][off] = true
	}
	return nil
}

func (m *MemoryRepository) GetDiscardedOffsets(ctx context.Context, topic string) (map[int64]bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.discardedOffsets == nil || m.discardedOffsets[topic] == nil {
		return make(map[int64]bool), nil
	}
	res := make(map[int64]bool)
	for k, v := range m.discardedOffsets[topic] {
		res[k] = v
	}
	return res, nil
}

func (m *MemoryRepository) SaveFeatureFlag(ctx context.Context, flag *featureflags.FeatureFlag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.flags == nil {
		m.flags = make(map[string]*featureflags.FeatureFlag)
	}
	m.flags[flag.FlagKey] = flag
	return featureflags.SetFlag(ctx, nil, nil, flag)
}

func (m *MemoryRepository) GetFeatureFlag(ctx context.Context, key string) (*featureflags.FeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.flags == nil {
		return nil, nil
	}
	return m.flags[key], nil
}

func (m *MemoryRepository) ListFeatureFlags(ctx context.Context) ([]*featureflags.FeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*featureflags.FeatureFlag
	for _, f := range m.flags {
		res = append(res, f)
	}
	return res, nil
}

func (m *MemoryRepository) DeleteFeatureFlag(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.flags, key)
	return featureflags.DeleteFlag(ctx, nil, nil, key)
}
