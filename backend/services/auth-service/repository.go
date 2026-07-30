package main

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByGoogleSub(ctx context.Context, googleSub string) (*User, error)
	CreateUserWithPhone(ctx context.Context, phone string) (*User, error)
	CreateUserWithEmail(ctx context.Context, email string) (*User, error)
	CreateUserWithGoogle(ctx context.Context, googleSub, email string, emailVerified bool) (*User, error)
	LinkGoogleSubToUser(ctx context.Context, userID, googleSub string) error
	UpdateUserVerifiedAt(ctx context.Context, userID, channel string) error
	UpdateAuthProviderLast(ctx context.Context, userID, provider string) error
	CreateSession(ctx context.Context, session *AuthSession) error
	GetSessionByRefreshHash(ctx context.Context, refreshHash string) (*AuthSession, error)
	RevokeSession(ctx context.Context, refreshHash string) error
	ListUsersAdmin(ctx context.Context, phoneFilter, emailFilter string, page, pageSize int) ([]*User, int64, error)
	DeleteUserPII(ctx context.Context, userID string) (int, error)
	GetAppVersion(ctx context.Context, platform string) (*AppVersion, error)
	UpsertAppVersion(ctx context.Context, av *AppVersion) error
	UpdateUserName(ctx context.Context, userID, name string) (*User, error)
	GetUserSessions(ctx context.Context, userID string) ([]*AuthSession, error)
	RevokeSessionByID(ctx context.Context, sessionID, userID string) error
	RevokeAllOtherSessions(ctx context.Context, currentSessionID, userID string) error
	UpdateSessionLastUsed(ctx context.Context, sessionID string) error

	SetRecoveryEmail(ctx context.Context, userID, email string) error
	ConfirmRecoveryEmail(ctx context.Context, userID string) error
	CreateRecoveryRequest(ctx context.Context, req *AccountRecoveryRequest) error
	GetRecoveryRequestByID(ctx context.Context, requestID string) (*AccountRecoveryRequest, error)
	UpdateRecoveryRequestStatus(ctx context.Context, requestID, status string) error
	UpdateUserPhoneAndRevokeSessions(ctx context.Context, userID, newPhone string) error
}

type PostgresRepository struct {
	db *db.DB
}

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{db: database}
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last FROM users WHERE id = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last FROM users WHERE phone = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last FROM users WHERE email = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserByGoogleSub(ctx context.Context, googleSub string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last FROM users WHERE google_sub = $1`
	var u User
	err := r.db.QueryRowContext(ctx, query, googleSub).Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) UpdateUserName(ctx context.Context, userID, name string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		UPDATE users
		SET name = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, name, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last
	`
	var u User
	err := r.db.QueryRowContext(ctx, query, name, userID).Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return &u, err
}

func (r *PostgresRepository) CreateUserWithPhone(ctx context.Context, phone string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id := uuid.New().String()
	now := time.Now()
	provider := "phone"
	query := `
		INSERT INTO users (id, phone, phone_verified_at, auth_provider_last)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last
	`
	var u User
	err := r.db.QueryRowContext(ctx, query, id, phone, now, provider).Scan(
		&u.ID, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) CreateUserWithEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id := uuid.New().String()
	now := time.Now()
	provider := "email"
	query := `
		INSERT INTO users (id, email, email_verified_at, auth_provider_last)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last
	`
	var u User
	err := r.db.QueryRowContext(ctx, query, id, email, now, provider).Scan(
		&u.ID, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) CreateUserWithGoogle(ctx context.Context, googleSub, email string, emailVerified bool) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id := uuid.New().String()
	provider := "google"
	var emailVerifiedAt *time.Time
	if emailVerified {
		now := time.Now()
		emailVerifiedAt = &now
	}
	var emailVal *string
	if email != "" {
		emailVal = &email
	}

	query := `
		INSERT INTO users (id, google_sub, email, email_verified_at, auth_provider_last)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last
	`
	var u User
	err := r.db.QueryRowContext(ctx, query, id, googleSub, emailVal, emailVerifiedAt, provider).Scan(
		&u.ID, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) LinkGoogleSubToUser(ctx context.Context, userID, googleSub string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE users SET google_sub = $1, auth_provider_last = 'google' WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, googleSub, userID)
	return err
}

func (r *PostgresRepository) UpdateUserVerifiedAt(ctx context.Context, userID, channel string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	var query string
	if channel == "phone" {
		query = `UPDATE users SET phone_verified_at = $1, auth_provider_last = 'phone' WHERE id = $2`
	} else {
		query = `UPDATE users SET email_verified_at = $1, auth_provider_last = 'email' WHERE id = $2`
	}
	_, err := r.db.ExecContext(ctx, query, now, userID)
	return err
}

func (r *PostgresRepository) UpdateAuthProviderLast(ctx context.Context, userID, provider string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE users SET auth_provider_last = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, provider, userID)
	return err
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session *AuthSession) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	query := `
		INSERT INTO auth_sessions (id, device_id, device_label, refresh_token_hash, user_id, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, session.ID, session.DeviceID, session.DeviceLabel, session.RefreshTokenHash, session.UserID, session.CreatedAt, session.LastUsedAt)
	return err
}

func (r *PostgresRepository) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (*AuthSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, device_id, device_label, refresh_token_hash, user_id, created_at, last_used_at, revoked_at FROM auth_sessions WHERE refresh_token_hash = $1`
	var s AuthSession
	err := r.db.QueryRowContext(ctx, query, refreshHash).Scan(
		&s.ID, &s.DeviceID, &s.DeviceLabel, &s.RefreshTokenHash, &s.UserID, &s.CreatedAt, &s.LastUsedAt, &s.RevokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) GetUserSessions(ctx context.Context, userID string) ([]*AuthSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, device_id, device_label, refresh_token_hash, user_id, created_at, last_used_at, revoked_at
		FROM auth_sessions
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*AuthSession
	for rows.Next() {
		var s AuthSession
		if err := rows.Scan(&s.ID, &s.DeviceID, &s.DeviceLabel, &s.RefreshTokenHash, &s.UserID, &s.CreatedAt, &s.LastUsedAt, &s.RevokedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	if sessions == nil {
		sessions = []*AuthSession{}
	}
	return sessions, nil
}

func (r *PostgresRepository) RevokeSessionByID(ctx context.Context, sessionID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE auth_sessions SET revoked_at = $1 WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), sessionID, userID)
	return err
}

func (r *PostgresRepository) RevokeAllOtherSessions(ctx context.Context, currentSessionID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE auth_sessions SET revoked_at = $1 WHERE user_id = $2 AND id != $3 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID, currentSessionID)
	return err
}

func (r *PostgresRepository) UpdateSessionLastUsed(ctx context.Context, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE auth_sessions SET last_used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), sessionID)
	return err
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, refreshHash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	query := `UPDATE auth_sessions SET revoked_at = $1 WHERE refresh_token_hash = $2`
	_, err := r.db.ExecContext(ctx, query, now, refreshHash)
	return err
}

func (r *PostgresRepository) GetAppVersion(ctx context.Context, platform string) (*AppVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT platform, min_supported_version, latest_version, hard_update_below, soft_update_message, updated_at FROM app_versions WHERE platform = $1`
	var av AppVersion
	err := r.db.QueryRowContext(ctx, query, platform).Scan(
		&av.Platform, &av.MinSupportedVersion, &av.LatestVersion, &av.HardUpdateBelow, &av.SoftUpdateMessage, &av.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &av, nil
}

func (r *PostgresRepository) UpsertAppVersion(ctx context.Context, av *AppVersion) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO app_versions (platform, min_supported_version, latest_version, hard_update_below, soft_update_message, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (platform) DO UPDATE SET
			min_supported_version = EXCLUDED.min_supported_version,
			latest_version = EXCLUDED.latest_version,
			hard_update_below = EXCLUDED.hard_update_below,
			soft_update_message = EXCLUDED.soft_update_message,
			updated_at = NOW()
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		av.Platform, av.MinSupportedVersion, av.LatestVersion, av.HardUpdateBelow, av.SoftUpdateMessage,
	).Scan(&av.UpdatedAt)
}

// In-Memory Repository for unit tests and fallback
type MemoryRepository struct {
	mu           sync.RWMutex
	users        map[string]*User
	sessions     map[string]*AuthSession
	appVersions  map[string]*AppVersion
	recoveryReqs map[string]*AccountRecoveryRequest
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:        make(map[string]*User),
		sessions:     make(map[string]*AuthSession),
		appVersions:  make(map[string]*AppVersion),
		recoveryReqs: make(map[string]*AccountRecoveryRequest),
	}
}

func (m *MemoryRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return copyUser(u), nil
}

func (m *MemoryRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Phone != nil && *u.Phone == phone {
			return copyUser(u), nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email != nil && *u.Email == email {
			return copyUser(u), nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetUserByGoogleSub(ctx context.Context, googleSub string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.GoogleSub != nil && *u.GoogleSub == googleSub {
			return copyUser(u), nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) CreateUserWithPhone(ctx context.Context, phone string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	prov := "phone"
	u := &User{
		ID:               uuid.New().String(),
		Phone:            &phone,
		PhoneVerifiedAt:  &now,
		AuthProviderLast: &prov,
	}
	m.users[u.ID] = u
	return copyUser(u), nil
}

func (m *MemoryRepository) CreateUserWithEmail(ctx context.Context, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	prov := "email"
	u := &User{
		ID:               uuid.New().String(),
		Email:            &email,
		EmailVerifiedAt:  &now,
		AuthProviderLast: &prov,
	}
	m.users[u.ID] = u
	return copyUser(u), nil
}

func (m *MemoryRepository) CreateUserWithGoogle(ctx context.Context, googleSub, email string, emailVerified bool) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prov := "google"
	var emailVal *string
	if email != "" {
		emailVal = &email
	}
	var emailVerifiedAt *time.Time
	if emailVerified {
		now := time.Now()
		emailVerifiedAt = &now
	}
	u := &User{
		ID:               uuid.New().String(),
		GoogleSub:        &googleSub,
		Email:            emailVal,
		EmailVerifiedAt:  emailVerifiedAt,
		AuthProviderLast: &prov,
	}
	m.users[u.ID] = u
	return copyUser(u), nil
}

func (m *MemoryRepository) LinkGoogleSubToUser(ctx context.Context, userID, googleSub string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.GoogleSub = &googleSub
	prov := "google"
	u.AuthProviderLast = &prov
	return nil
}

func (m *MemoryRepository) UpdateUserVerifiedAt(ctx context.Context, userID, channel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	now := time.Now()
	if channel == "phone" {
		u.PhoneVerifiedAt = &now
		prov := "phone"
		u.AuthProviderLast = &prov
	} else {
		u.EmailVerifiedAt = &now
		prov := "email"
		u.AuthProviderLast = &prov
	}
	return nil
}

func (m *MemoryRepository) UpdateAuthProviderLast(ctx context.Context, userID, provider string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.AuthProviderLast = &provider
	return nil
}

func (m *MemoryRepository) CreateSession(ctx context.Context, session *AuthSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	m.sessions[session.RefreshTokenHash] = session
	return nil
}

func (m *MemoryRepository) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (*AuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[refreshHash]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *MemoryRepository) RevokeSession(ctx context.Context, refreshHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[refreshHash]
	if ok {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func copyUser(u *User) *User {
	if u == nil {
		return nil
	}
	res := *u
	return &res
}

func (r *PostgresRepository) ListUsersAdmin(ctx context.Context, phoneFilter, emailFilter string, page, pageSize int) ([]*User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "1=1"
	args := []interface{}{}
	if phoneFilter != "" {
		where += " AND phone LIKE $" + string(rune(len(args)+1+'0'))
		args = append(args, "%"+phoneFilter+"%")
	}
	if emailFilter != "" {
		where += " AND email ILIKE $" + string(rune(len(args)+1+'0'))
		args = append(args, "%"+emailFilter+"%")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM users WHERE " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, pageSize, offset)
	query := "SELECT id, phone, email, google_sub, email_verified_at, phone_verified_at, auth_provider_last FROM users WHERE " + where + " ORDER BY id DESC LIMIT $" + string(rune(len(args)-1+'0')) + " OFFSET $" + string(rune(len(args)+'0'))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Phone, &u.Email, &u.GoogleSub, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.AuthProviderLast); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}

	if users == nil {
		users = []*User{}
	}
	return users, total, nil
}

func (m *MemoryRepository) ListUsersAdmin(ctx context.Context, phoneFilter, emailFilter string, page, pageSize int) ([]*User, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*User
	for _, u := range m.users {
		if phoneFilter != "" && (u.Phone == nil || !stringsContains(*u.Phone, phoneFilter)) {
			continue
		}
		if emailFilter != "" && (u.Email == nil || !stringsContains(stringsToLower(*u.Email), stringsToLower(emailFilter))) {
			continue
		}
		matched = append(matched, copyUser(u))
	}

	total := int64(len(matched))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= len(matched) {
		return []*User{}, total, nil
	}

	end := offset + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (r *PostgresRepository) DeleteUserPII(ctx context.Context, userID string) (int, error) {
	res, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return int(rows), nil
}

func (m *MemoryRepository) DeleteUserPII(ctx context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[userID]; ok {
		delete(m.users, userID)
		return 1, nil
	}
	return 0, nil
}

func (m *MemoryRepository) GetAppVersion(ctx context.Context, platform string) (*AppVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.appVersions == nil {
		return nil, nil
	}
	av, exists := m.appVersions[platform]
	if !exists {
		return nil, nil
	}
	cp := *av
	return &cp, nil
}

func (m *MemoryRepository) UpsertAppVersion(ctx context.Context, av *AppVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.appVersions == nil {
		m.appVersions = make(map[string]*AppVersion)
	}
	av.UpdatedAt = time.Now()
	cp := *av
	m.appVersions[av.Platform] = &cp
	return nil
}

func (m *MemoryRepository) UpdateUserName(ctx context.Context, userID, name string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	u.Name = &name
	return u, nil
}

func (m *MemoryRepository) GetUserSessions(ctx context.Context, userID string) ([]*AuthSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*AuthSession
	for _, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			cp := *s
			result = append(result, &cp)
		}
	}
	if result == nil {
		result = []*AuthSession{}
	}
	return result, nil
}

func (m *MemoryRepository) RevokeSessionByID(ctx context.Context, sessionID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, s := range m.sessions {
		if s.ID == sessionID && s.UserID == userID && s.RevokedAt == nil {
			s.RevokedAt = &now
			break
		}
	}
	return nil
}

func (m *MemoryRepository) RevokeAllOtherSessions(ctx context.Context, currentSessionID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, s := range m.sessions {
		if s.UserID == userID && s.ID != currentSessionID && s.RevokedAt == nil {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *MemoryRepository) UpdateSessionLastUsed(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, s := range m.sessions {
		if s.ID == sessionID {
			s.LastUsedAt = &now
			break
		}
	}
	return nil
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && stringsIndex(s, substr) >= 0))
}

func stringsIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func stringsToLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func (r *PostgresRepository) SetRecoveryEmail(ctx context.Context, userID, email string) error {
	query := `UPDATE users SET recovery_email = $1, recovery_email_verified_at = NULL WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, email, userID)
	return err
}

func (r *PostgresRepository) ConfirmRecoveryEmail(ctx context.Context, userID string) error {
	query := `UPDATE users SET recovery_email_verified_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *PostgresRepository) CreateRecoveryRequest(ctx context.Context, req *AccountRecoveryRequest) error {
	query := `INSERT INTO account_recovery_requests (id, user_id, original_identifier, new_identifier, verification_method, status, support_ticket_id, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, req.ID, req.UserID, req.OriginalIdentifier, req.NewIdentifier, req.VerificationMethod, req.Status, req.SupportTicketID, req.CreatedAt)
	return err
}

func (r *PostgresRepository) GetRecoveryRequestByID(ctx context.Context, requestID string) (*AccountRecoveryRequest, error) {
	query := `SELECT id, user_id, original_identifier, new_identifier, verification_method, status, support_ticket_id, created_at, resolved_at
			  FROM account_recovery_requests WHERE id = $1`
	var req AccountRecoveryRequest
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(&req.ID, &req.UserID, &req.OriginalIdentifier, &req.NewIdentifier, &req.VerificationMethod, &req.Status, &req.SupportTicketID, &req.CreatedAt, &req.ResolvedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &req, nil
}

func (r *PostgresRepository) UpdateRecoveryRequestStatus(ctx context.Context, requestID, status string) error {
	now := time.Now()
	query := `UPDATE account_recovery_requests SET status = $1, resolved_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, now, requestID)
	return err
}

func (r *PostgresRepository) UpdateUserPhoneAndRevokeSessions(ctx context.Context, userID, newPhone string) error {
	// Check if newPhone is already in use
	var count int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE phone = $1 AND id != $2`, newPhone, userID).Scan(&count)
	if count > 0 {
		return errors.New("PHONE_ALREADY_IN_USE")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	_, err = tx.ExecContext(ctx, `UPDATE users SET phone = $1, phone_verified_at = $2 WHERE id = $3`, newPhone, now, userID)
	if err != nil {
		return err
	}

	// Revoke ALL active sessions for user to force re-authentication
	_, err = tx.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`, now, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MemoryRepository) SetRecoveryEmail(ctx context.Context, userID, email string) error {
	u, ok := r.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.RecoveryEmail = &email
	u.RecoveryEmailVerifiedAt = nil
	return nil
}

func (r *MemoryRepository) ConfirmRecoveryEmail(ctx context.Context, userID string) error {
	u, ok := r.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	now := time.Now()
	u.RecoveryEmailVerifiedAt = &now
	return nil
}

func (r *MemoryRepository) CreateRecoveryRequest(ctx context.Context, req *AccountRecoveryRequest) error {
	if r.recoveryReqs == nil {
		r.recoveryReqs = make(map[string]*AccountRecoveryRequest)
	}
	r.recoveryReqs[req.ID] = req
	return nil
}

func (r *MemoryRepository) GetRecoveryRequestByID(ctx context.Context, requestID string) (*AccountRecoveryRequest, error) {
	if r.recoveryReqs == nil {
		return nil, ErrUserNotFound
	}
	req, ok := r.recoveryReqs[requestID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return req, nil
}

func (r *MemoryRepository) UpdateRecoveryRequestStatus(ctx context.Context, requestID, status string) error {
	if r.recoveryReqs != nil {
		if req, ok := r.recoveryReqs[requestID]; ok {
			now := time.Now()
			req.Status = status
			req.ResolvedAt = &now
		}
	}
	return nil
}

func (r *MemoryRepository) UpdateUserPhoneAndRevokeSessions(ctx context.Context, userID, newPhone string) error {
	// Check phone duplicate
	for _, u := range r.users {
		if u.Phone != nil && *u.Phone == newPhone && u.ID != userID {
			return errors.New("PHONE_ALREADY_IN_USE")
		}
	}

	u, ok := r.users[userID]
	if !ok {
		return ErrUserNotFound
	}

	now := time.Now()
	u.Phone = &newPhone
	u.PhoneVerifiedAt = &now

	// Revoke sessions
	for _, s := range r.sessions {
		if s.UserID == userID {
			s.RevokedAt = &now
		}
	}
	return nil
}
