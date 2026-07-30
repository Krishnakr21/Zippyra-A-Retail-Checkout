package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zippyra/backend/shared/db"
	"github.com/zippyra/backend/shared/logger"
)

type CheckoutSessionRepository interface {
	CreateCheckoutSession(ctx context.Context, session *CheckoutSession) error
	GetCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error)
	GetPendingSessionByUserID(ctx context.Context, userID string) (*CheckoutSession, error)
	ConsumeCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error)
	ExpireStaleSessions(ctx context.Context) ([]*CheckoutSession, error)
}

type PostgresCheckoutRepository struct {
	database *db.DB
}

func NewPostgresCheckoutRepository(database *db.DB) CheckoutSessionRepository {
	return &PostgresCheckoutRepository{database: database}
}

func (p *PostgresCheckoutRepository) CreateCheckoutSession(ctx context.Context, s *CheckoutSession) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = s.CreatedAt.Add(10 * time.Minute)
	}
	if s.Status == "" {
		s.Status = "PENDING"
	}

	itemsJSON, err := json.Marshal(s.Items)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO checkout_sessions (id, user_id, store_id, items, subtotal_paise, discount_paise,
		                             cgst_paise, sgst_paise, igst_paise, total_paise, coupon_code,
		                             supply_type, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = p.database.ExecContext(ctx, query,
		s.ID, s.UserID, s.StoreID, itemsJSON, s.SubtotalPaise, s.DiscountPaise,
		s.CGSTPaise, s.SGSTPaise, s.IGSTPaise, s.TotalPaise, s.CouponCode,
		s.SupplyType, s.Status, s.CreatedAt, s.ExpiresAt,
	)
	return err
}

func (p *PostgresCheckoutRepository) GetCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error) {
	query := `
		SELECT id, user_id, store_id, items, subtotal_paise, discount_paise, cgst_paise, sgst_paise,
		       igst_paise, total_paise, COALESCE(coupon_code, ''), supply_type, status, created_at, expires_at
		FROM checkout_sessions
		WHERE id = $1
	`
	var s CheckoutSession
	var itemsJSON []byte
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.StoreID, &itemsJSON, &s.SubtotalPaise, &s.DiscountPaise,
		&s.CGSTPaise, &s.SGSTPaise, &s.IGSTPaise, &s.TotalPaise, &s.CouponCode,
		&s.SupplyType, &s.Status, &s.CreatedAt, &s.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(itemsJSON) > 0 {
		_ = json.Unmarshal(itemsJSON, &s.Items)
	}
	return &s, nil
}

func (p *PostgresCheckoutRepository) GetPendingSessionByUserID(ctx context.Context, userID string) (*CheckoutSession, error) {
	query := `
		SELECT id, user_id, store_id, items, subtotal_paise, discount_paise, cgst_paise, sgst_paise,
		       igst_paise, total_paise, COALESCE(coupon_code, ''), supply_type, status, created_at, expires_at
		FROM checkout_sessions
		WHERE user_id = $1 AND status = 'PENDING' AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s CheckoutSession
	var itemsJSON []byte
	err := p.database.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.StoreID, &itemsJSON, &s.SubtotalPaise, &s.DiscountPaise,
		&s.CGSTPaise, &s.SGSTPaise, &s.IGSTPaise, &s.TotalPaise, &s.CouponCode,
		&s.SupplyType, &s.Status, &s.CreatedAt, &s.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(itemsJSON) > 0 {
		_ = json.Unmarshal(itemsJSON, &s.Items)
	}
	return &s, nil
}

func (p *PostgresCheckoutRepository) ConsumeCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error) {
	query := `
		UPDATE checkout_sessions
		SET status = 'CONSUMED'
		WHERE id = $1 AND status = 'PENDING' AND expires_at > CURRENT_TIMESTAMP
		RETURNING id, user_id, store_id, items, subtotal_paise, discount_paise, cgst_paise, sgst_paise,
		          igst_paise, total_paise, COALESCE(coupon_code, ''), supply_type, status, created_at, expires_at
	`
	var s CheckoutSession
	var itemsJSON []byte
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.StoreID, &itemsJSON, &s.SubtotalPaise, &s.DiscountPaise,
		&s.CGSTPaise, &s.SGSTPaise, &s.IGSTPaise, &s.TotalPaise, &s.CouponCode,
		&s.SupplyType, &s.Status, &s.CreatedAt, &s.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(itemsJSON) > 0 {
		_ = json.Unmarshal(itemsJSON, &s.Items)
	}
	return &s, nil
}

func (p *PostgresCheckoutRepository) ExpireStaleSessions(ctx context.Context) ([]*CheckoutSession, error) {
	query := `
		UPDATE checkout_sessions
		SET status = 'EXPIRED'
		WHERE status = 'PENDING' AND expires_at <= CURRENT_TIMESTAMP
		RETURNING id, user_id, store_id, items, subtotal_paise, discount_paise, cgst_paise, sgst_paise,
		          igst_paise, total_paise, COALESCE(coupon_code, ''), supply_type, status, created_at, expires_at
	`
	rows, err := p.database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []*CheckoutSession
	for rows.Next() {
		var s CheckoutSession
		var itemsJSON []byte
		err := rows.Scan(
			&s.ID, &s.UserID, &s.StoreID, &itemsJSON, &s.SubtotalPaise, &s.DiscountPaise,
			&s.CGSTPaise, &s.SGSTPaise, &s.IGSTPaise, &s.TotalPaise, &s.CouponCode,
			&s.SupplyType, &s.Status, &s.CreatedAt, &s.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		if len(itemsJSON) > 0 {
			_ = json.Unmarshal(itemsJSON, &s.Items)
		}
		expired = append(expired, &s)
	}

	return expired, nil
}

type LockManager interface {
	AcquireCheckoutLock(ctx context.Context, userID string) (bool, error)
	ReleaseCheckoutLock(ctx context.Context, userID string) error
}

type RedisLockManager struct {
	rdb redis.Cmdable
}

func NewRedisLockManager(rdb redis.Cmdable) LockManager {
	return &RedisLockManager{rdb: rdb}
}

func (l *RedisLockManager) AcquireCheckoutLock(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("checkout_lock:%s", userID)
	res, err := l.rdb.SetNX(ctx, key, "LOCKED", 5*time.Minute).Result()
	if err != nil {
		return true, nil
	}
	return res, nil
}

func (l *RedisLockManager) ReleaseCheckoutLock(ctx context.Context, userID string) error {
	key := fmt.Sprintf("checkout_lock:%s", userID)
	return l.rdb.Del(ctx, key).Err()
}

// Background Stale Session Cleaner
func StartStaleSessionCleaner(ctx context.Context, repo CheckoutSessionRepository, lockMgr LockManager) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expiredSessions, err := repo.ExpireStaleSessions(context.Background())
			if err != nil {
				logger.Error("Error cleaning expired checkout sessions: %v", err)
				continue
			}

			for _, s := range expiredSessions {
				_ = lockMgr.ReleaseCheckoutLock(context.Background(), s.UserID)
				logger.Info("Expired stale checkout session %s for user %s and released checkout lock", s.ID, s.UserID)
			}
		}
	}
}

// MemoryCheckoutRepository for testing
type MemoryCheckoutRepository struct {
	mu       sync.Mutex
	sessions map[string]*CheckoutSession
}

func NewMemoryCheckoutRepository() CheckoutSessionRepository {
	return &MemoryCheckoutRepository{
		sessions: make(map[string]*CheckoutSession),
	}
}

func (m *MemoryCheckoutRepository) CreateCheckoutSession(ctx context.Context, s *CheckoutSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = s.CreatedAt.Add(10 * time.Minute)
	}
	if s.Status == "" {
		s.Status = "PENDING"
	}

	m.sessions[s.ID] = s
	return nil
}

func (m *MemoryCheckoutRepository) GetCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *MemoryCheckoutRepository) GetPendingSessionByUserID(ctx context.Context, userID string) (*CheckoutSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, s := range m.sessions {
		if s.UserID == userID && s.Status == "PENDING" && s.ExpiresAt.After(now) {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MemoryCheckoutRepository) ConsumeCheckoutSession(ctx context.Context, id string) (*CheckoutSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok || s.Status != "PENDING" || time.Now().After(s.ExpiresAt) {
		return nil, errors.New("checkout session not active")
	}

	s.Status = "CONSUMED"
	return s, nil
}

func (m *MemoryCheckoutRepository) ExpireStaleSessions(ctx context.Context) ([]*CheckoutSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expired []*CheckoutSession
	for _, s := range m.sessions {
		if s.Status == "PENDING" && !s.ExpiresAt.After(now) {
			s.Status = "EXPIRED"
			expired = append(expired, s)
		}
	}
	return expired, nil
}

type MemoryLockManager struct {
	mu    sync.Mutex
	locks map[string]bool
}

func NewMemoryLockManager() LockManager {
	return &MemoryLockManager{locks: make(map[string]bool)}
}

func (m *MemoryLockManager) AcquireCheckoutLock(ctx context.Context, userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.locks[userID] {
		return false, nil
	}
	m.locks[userID] = true
	return true, nil
}

func (m *MemoryLockManager) ReleaseCheckoutLock(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, userID)
	return nil
}
