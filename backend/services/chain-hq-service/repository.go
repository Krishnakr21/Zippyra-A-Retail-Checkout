package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var (
	ErrUserNotFound  = errors.New("chain hq user not found")
	ErrUserExists    = errors.New("user with this phone already exists")
	ErrOwnerExists   = errors.New("chain already has an active owner")
	ErrJobNotFound   = errors.New("bulk import job not found")
)

type Repository interface {
	GetUserByPhone(ctx context.Context, phone string) (*ChainHQUser, error)
	GetUserByID(ctx context.Context, id string) (*ChainHQUser, error)
	GetActiveOwnerByChainID(ctx context.Context, chainID string) (*ChainHQUser, error)
	CreateUser(ctx context.Context, user *ChainHQUser) error
	UpdateUser(ctx context.Context, user *ChainHQUser) error
	AnonymizeUser(ctx context.Context, userID string) error
	ListUsersByChainID(ctx context.Context, chainID string, roleFilter string) ([]*ChainHQUser, error)

	CreateSession(ctx context.Context, session *ChainHQSession) error
	RevokeUserSessions(ctx context.Context, userID string) error

	CreateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error
	GetBulkImportJob(ctx context.Context, id string) (*ChainBulkImportJob, error)
	UpdateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error
}

type PostgresRepository struct {
	db *db.DB
}

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{db: database}
}

func (p *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*ChainHQUser, error) {
	query := `SELECT id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at FROM chain_hq_users WHERE phone = $1`
	u := &ChainHQUser{}
	err := p.db.QueryRowContext(ctx, query, phone).Scan(&u.ID, &u.ChainID, &u.Phone, &u.Name, &u.Role, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (p *PostgresRepository) GetUserByID(ctx context.Context, id string) (*ChainHQUser, error) {
	query := `SELECT id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at FROM chain_hq_users WHERE id = $1`
	u := &ChainHQUser{}
	err := p.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.ChainID, &u.Phone, &u.Name, &u.Role, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (p *PostgresRepository) GetActiveOwnerByChainID(ctx context.Context, chainID string) (*ChainHQUser, error) {
	query := `SELECT id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at FROM chain_hq_users WHERE chain_id = $1 AND role = 'OWNER' AND is_active = true LIMIT 1`
	u := &ChainHQUser{}
	err := p.db.QueryRowContext(ctx, query, chainID).Scan(&u.ID, &u.ChainID, &u.Phone, &u.Name, &u.Role, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (p *PostgresRepository) CreateUser(ctx context.Context, user *ChainHQUser) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `INSERT INTO chain_hq_users (id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := p.db.ExecContext(ctx, query, user.ID, user.ChainID, user.Phone, user.Name, user.Role, user.IsActive, user.CreatedBy, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return ErrUserExists
	}
	return nil
}

func (p *PostgresRepository) UpdateUser(ctx context.Context, user *ChainHQUser) error {
	user.UpdatedAt = time.Now().UTC()
	query := `UPDATE chain_hq_users SET name = $1, role = $2, is_active = $3, updated_at = $4 WHERE id = $5`
	_, err := p.db.ExecContext(ctx, query, user.Name, user.Role, user.IsActive, user.UpdatedAt, user.ID)
	return err
}

func (p *PostgresRepository) ListUsersByChainID(ctx context.Context, chainID string, roleFilter string) ([]*ChainHQUser, error) {
	query := `SELECT id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at FROM chain_hq_users WHERE chain_id = $1`
	args := []interface{}{chainID}

	if roleFilter != "" {
		query += ` AND role = $2`
		args = append(args, roleFilter)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*ChainHQUser
	for rows.Next() {
		u := &ChainHQUser{}
		if err := rows.Scan(&u.ID, &u.ChainID, &u.Phone, &u.Name, &u.Role, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}

func (p *PostgresRepository) CreateSession(ctx context.Context, session *ChainHQSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	query := `INSERT INTO chain_hq_sessions (id, chain_hq_user_id, device_id, created_at) VALUES ($1, $2, $3, $4)`
	_, err := p.db.ExecContext(ctx, query, session.ID, session.ChainHQUserID, session.DeviceID, session.CreatedAt)
	return err
}

func (p *PostgresRepository) RevokeUserSessions(ctx context.Context, userID string) error {
	query := `UPDATE chain_hq_sessions SET revoked_at = $1 WHERE chain_hq_user_id = $2 AND revoked_at IS NULL`
	_, err := p.db.ExecContext(ctx, query, time.Now().UTC(), userID)
	return err
}

func (p *PostgresRepository) CreateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now

	jsonBytes, err := json.Marshal(job.PerStoreJobIDs)
	if err != nil {
		return err
	}

	query := `INSERT INTO chain_bulk_import_jobs (id, chain_id, per_store_job_ids, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = p.db.ExecContext(ctx, query, job.ID, job.ChainID, string(jsonBytes), job.Status, job.CreatedAt, job.UpdatedAt)
	return err
}

func (p *PostgresRepository) GetBulkImportJob(ctx context.Context, id string) (*ChainBulkImportJob, error) {
	query := `SELECT id, chain_id, per_store_job_ids, status, created_at, updated_at FROM chain_bulk_import_jobs WHERE id = $1`
	j := &ChainBulkImportJob{}
	var jsonStr string
	err := p.db.QueryRowContext(ctx, query, id).Scan(&j.ID, &j.ChainID, &jsonStr, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, ErrJobNotFound
	}
	_ = json.Unmarshal([]byte(jsonStr), &j.PerStoreJobIDs)
	return j, nil
}

func (p *PostgresRepository) UpdateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error {
	job.UpdatedAt = time.Now().UTC()
	jsonBytes, err := json.Marshal(job.PerStoreJobIDs)
	if err != nil {
		return err
	}
	query := `UPDATE chain_bulk_import_jobs SET per_store_job_ids = $1, status = $2, updated_at = $3 WHERE id = $4`
	_, err = p.db.ExecContext(ctx, query, string(jsonBytes), job.Status, job.UpdatedAt, job.ID)
	return err
}

// MemoryRepository for testing
type MemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]*ChainHQUser
	sessions map[string]*ChainHQSession
	jobs     map[string]*ChainBulkImportJob
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:    make(map[string]*ChainHQUser),
		sessions: make(map[string]*ChainHQSession),
		jobs:     make(map[string]*ChainBulkImportJob),
	}
}

func (m *MemoryRepository) GetUserByPhone(ctx context.Context, phone string) (*ChainHQUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Phone == phone {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MemoryRepository) GetUserByID(ctx context.Context, id string) (*ChainHQUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *MemoryRepository) GetActiveOwnerByChainID(ctx context.Context, chainID string) (*ChainHQUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.ChainID == chainID && u.Role == RoleOwner && u.IsActive {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MemoryRepository) CreateUser(ctx context.Context, user *ChainHQUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Phone == user.Phone {
			return ErrUserExists
		}
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *MemoryRepository) UpdateUser(ctx context.Context, user *ChainHQUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[user.ID]; !ok {
		return ErrUserNotFound
	}
	user.UpdatedAt = time.Now().UTC()
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *MemoryRepository) ListUsersByChainID(ctx context.Context, chainID string, roleFilter string) ([]*ChainHQUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*ChainHQUser
	for _, u := range m.users {
		if u.ChainID == chainID {
			if roleFilter == "" || u.Role == roleFilter {
				cp := *u
				result = append(result, &cp)
			}
		}
	}
	return result, nil
}

func (m *MemoryRepository) CreateSession(ctx context.Context, session *ChainHQSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	cp := *session
	m.sessions[session.ID] = &cp
	return nil
}

func (m *MemoryRepository) RevokeUserSessions(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, sess := range m.sessions {
		if sess.ChainHQUserID == userID && sess.RevokedAt == nil {
			sess.RevokedAt = &now
		}
	}
	return nil
}

func (m *MemoryRepository) CreateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now

	cp := *job
	m.jobs[job.ID] = &cp
	return nil
}

func (m *MemoryRepository) GetBulkImportJob(ctx context.Context, id string) (*ChainBulkImportJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	cp := *job
	return &cp, nil
}

func (m *MemoryRepository) UpdateBulkImportJob(ctx context.Context, job *ChainBulkImportJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.jobs[job.ID]; !ok {
		return ErrJobNotFound
	}
	job.UpdatedAt = time.Now().UTC()
	cp := *job
	m.jobs[job.ID] = &cp
	return nil
}

func (p *PostgresRepository) AnonymizeUser(ctx context.Context, userID string) error {
	tombstonePhone := fmt.Sprintf("deleted_%s", userID)
	query := `UPDATE chain_hq_users SET name = 'Anonymized HQ User', phone = $1, is_active = false, updated_at = NOW() WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, tombstonePhone, userID)
	return err
}

func (m *MemoryRepository) AnonymizeUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.Name = "Anonymized HQ User"
	u.Phone = fmt.Sprintf("deleted_%s", userID)
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
	return nil
}
