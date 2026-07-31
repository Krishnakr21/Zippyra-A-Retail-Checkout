package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var (
	ErrAdminNotFound      = errors.New("admin user not found")
	ErrAdminAlreadyExists = errors.New("admin user already exists")
)

type Repository interface {
	GetAdminByEmail(ctx context.Context, email string) (*AdminUser, error)
	GetAdminByID(ctx context.Context, id string) (*AdminUser, error)
	CreateAdmin(ctx context.Context, admin *AdminUser) error
	UpdateAdmin(ctx context.Context, admin *AdminUser) error
	ListAdmins(ctx context.Context, roleFilter string, page, pageSize int) ([]*AdminUser, int, error)

	CreateSession(ctx context.Context, session *AdminSession) error
	RevokeAdminSessions(ctx context.Context, adminID string) error
}

type PostgresRepository struct {
	db *db.DB
}

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{db: database}
}

func (p *PostgresRepository) GetAdminByEmail(ctx context.Context, email string) (*AdminUser, error) {
	query := `
		SELECT id, email, name, role, google_sub, totp_secret_encrypted, totp_enabled_at,
		       totp_failed_attempts, totp_locked_until, is_active, created_by, created_at, updated_at
		FROM admin_users
		WHERE LOWER(email) = LOWER($1)
	`
	u := &AdminUser{}
	err := p.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.GoogleSub, &u.TOTPSecretEncrypted, &u.TOTPEnabledAt,
		&u.TOTPFailedAttempts, &u.TOTPLockedUntil, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, ErrAdminNotFound
	}
	return u, nil
}

func (p *PostgresRepository) GetAdminByID(ctx context.Context, id string) (*AdminUser, error) {
	query := `
		SELECT id, email, name, role, google_sub, totp_secret_encrypted, totp_enabled_at,
		       totp_failed_attempts, totp_locked_until, is_active, created_by, created_at, updated_at
		FROM admin_users
		WHERE id = $1
	`
	u := &AdminUser{}
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.GoogleSub, &u.TOTPSecretEncrypted, &u.TOTPEnabledAt,
		&u.TOTPFailedAttempts, &u.TOTPLockedUntil, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, ErrAdminNotFound
	}
	return u, nil
}

func (p *PostgresRepository) CreateAdmin(ctx context.Context, admin *AdminUser) error {
	if admin.ID == "" {
		admin.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	query := `
		INSERT INTO admin_users (id, email, name, role, google_sub, totp_secret_encrypted, totp_enabled_at,
		                         totp_failed_attempts, totp_locked_until, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := p.db.ExecContext(ctx, query,
		admin.ID, strings.ToLower(admin.Email), admin.Name, admin.Role, admin.GoogleSub,
		admin.TOTPSecretEncrypted, admin.TOTPEnabledAt, admin.TOTPFailedAttempts,
		admin.TOTPLockedUntil, admin.IsActive, admin.CreatedBy, admin.CreatedAt, admin.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return ErrAdminAlreadyExists
		}
		return fmt.Errorf("failed to insert admin_user: %w", err)
	}
	return nil
}

func (p *PostgresRepository) UpdateAdmin(ctx context.Context, admin *AdminUser) error {
	admin.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE admin_users
		SET name = $1, role = $2, google_sub = $3, totp_secret_encrypted = $4, totp_enabled_at = $5,
		    totp_failed_attempts = $6, totp_locked_until = $7, is_active = $8, updated_at = $9
		WHERE id = $10
	`
	_, err := p.db.ExecContext(ctx, query,
		admin.Name, admin.Role, admin.GoogleSub, admin.TOTPSecretEncrypted, admin.TOTPEnabledAt,
		admin.TOTPFailedAttempts, admin.TOTPLockedUntil, admin.IsActive, admin.UpdatedAt, admin.ID,
	)
	return err
}

func (p *PostgresRepository) ListAdmins(ctx context.Context, roleFilter string, page, pageSize int) ([]*AdminUser, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var countQuery string
	var args []interface{}

	if roleFilter != "" {
		countQuery = `SELECT COUNT(*) FROM admin_users WHERE role = $1`
		args = append(args, roleFilter)
	} else {
		countQuery = `SELECT COUNT(*) FROM admin_users`
	}

	var total int
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	var dataQuery string
	if roleFilter != "" {
		dataQuery = `
			SELECT id, email, name, role, google_sub, totp_secret_encrypted, totp_enabled_at,
			       totp_failed_attempts, totp_locked_until, is_active, created_by, created_at, updated_at
			FROM admin_users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		args = append(args, pageSize, offset)
	} else {
		dataQuery = `
			SELECT id, email, name, role, google_sub, totp_secret_encrypted, totp_enabled_at,
			       totp_failed_attempts, totp_locked_until, is_active, created_by, created_at, updated_at
			FROM admin_users ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`
		args = []interface{}{pageSize, offset}
	}

	rows, err := p.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*AdminUser
	for rows.Next() {
		u := &AdminUser{}
		err := rows.Scan(
			&u.ID, &u.Email, &u.Name, &u.Role, &u.GoogleSub, &u.TOTPSecretEncrypted, &u.TOTPEnabledAt,
			&u.TOTPFailedAttempts, &u.TOTPLockedUntil, &u.IsActive, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, u)
	}

	return result, total, nil
}

func (p *PostgresRepository) CreateSession(ctx context.Context, session *AdminSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	query := `INSERT INTO admin_sessions (id, admin_id, device_id, created_at) VALUES ($1, $2, $3, $4)`
	_, err := p.db.ExecContext(ctx, query, session.ID, session.AdminID, session.DeviceID, session.CreatedAt)
	return err
}

func (p *PostgresRepository) RevokeAdminSessions(ctx context.Context, adminID string) error {
	query := `UPDATE admin_sessions SET revoked_at = $1 WHERE admin_id = $2 AND revoked_at IS NULL`
	_, err := p.db.ExecContext(ctx, query, time.Now().UTC(), adminID)
	return err
}

// MemoryRepository for standalone unit testing
type MemoryRepository struct {
	mu       sync.RWMutex
	admins   map[string]*AdminUser
	sessions map[string]*AdminSession
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		admins:   make(map[string]*AdminUser),
		sessions: make(map[string]*AdminSession),
	}
}

func (m *MemoryRepository) GetAdminByEmail(ctx context.Context, email string) (*AdminUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, a := range m.admins {
		if strings.EqualFold(a.Email, email) {
			copy := *a
			return &copy, nil
		}
	}
	return nil, ErrAdminNotFound
}

func (m *MemoryRepository) GetAdminByID(ctx context.Context, id string) (*AdminUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	a, ok := m.admins[id]
	if !ok {
		return nil, ErrAdminNotFound
	}
	copy := *a
	return &copy, nil
}

func (m *MemoryRepository) CreateAdmin(ctx context.Context, admin *AdminUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.admins {
		if strings.EqualFold(existing.Email, admin.Email) {
			return ErrAdminAlreadyExists
		}
	}

	if admin.ID == "" {
		admin.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	copy := *admin
	m.admins[admin.ID] = &copy
	return nil
}

func (m *MemoryRepository) UpdateAdmin(ctx context.Context, admin *AdminUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.admins[admin.ID]; !ok {
		return ErrAdminNotFound
	}
	admin.UpdatedAt = time.Now().UTC()
	copy := *admin
	m.admins[admin.ID] = &copy
	return nil
}

func (m *MemoryRepository) ListAdmins(ctx context.Context, roleFilter string, page, pageSize int) ([]*AdminUser, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*AdminUser
	for _, a := range m.admins {
		if roleFilter == "" || a.Role == roleFilter {
			copy := *a
			matched = append(matched, &copy)
		}
	}

	total := len(matched)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*AdminUser{}, total, nil
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	return matched[offset:end], total, nil
}

func (m *MemoryRepository) CreateSession(ctx context.Context, session *AdminSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	copy := *session
	m.sessions[session.ID] = &copy
	return nil
}

func (m *MemoryRepository) RevokeAdminSessions(ctx context.Context, adminID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, sess := range m.sessions {
		if sess.AdminID == adminID && sess.RevokedAt == nil {
			sess.RevokedAt = &now
		}
	}
	return nil
}
