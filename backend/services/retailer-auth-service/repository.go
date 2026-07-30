package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var (
	ErrShiftActive    = errors.New("shift already active")
	ErrPhoneConflict  = errors.New("phone already staff")
	ErrStaffNotFound  = errors.New("staff member not found")
	ErrNoActiveShift  = errors.New("no active shift found")
)

type Repository interface {
	CreateStaffMember(ctx context.Context, s *StaffMember) error
	GetStaffByPhone(ctx context.Context, phone string) (*StaffMember, error)
	GetStaffByID(ctx context.Context, id string) (*StaffMember, error)
	ListStaffByStore(ctx context.Context, storeID string, roleFilter string, includeInactive bool, page, pageSize int) ([]*StaffMember, int64, error)
	UpdateStaffMember(ctx context.Context, s *StaffMember) error
	DeactivateStaffMemberTx(ctx context.Context, staffID string) error
	AnonymizeStaffMember(ctx context.Context, staffID string) error
	CreateSession(ctx context.Context, sess *StaffSession) error
	GetSessionByID(ctx context.Context, id string) (*StaffSession, error)
	RevokeSession(ctx context.Context, id string) error
	UpdateStaffPin(ctx context.Context, staffID string, pinHash string) error
	UpdatePinLockout(ctx context.Context, staffID string, failedAttempts int, lockedUntil *time.Time) error
	StartShiftTx(ctx context.Context, shift *StaffShift) error
	EndShift(ctx context.Context, staffID string) error
	GetActiveShift(ctx context.Context, staffID string) (*StaffShift, error)
	GetShiftHistory(ctx context.Context, storeID string, dateFrom, dateTo *time.Time, page, pageSize int) ([]*StaffShift, int64, error)
}

type PostgresRepository struct {
	database *db.DB
}

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}

func (p *PostgresRepository) CreateStaffMember(ctx context.Context, s *StaffMember) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.IsActive = true

	query := `
		INSERT INTO staff_members (id, store_id, chain_id, phone, name, role, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (phone) DO NOTHING
	`
	res, err := p.database.ExecContext(ctx, query, s.ID, s.StoreID, s.ChainID, s.Phone, s.Name, s.Role, s.IsActive, s.CreatedBy, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert staff member: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPhoneConflict
	}
	return nil
}

func (p *PostgresRepository) GetStaffByPhone(ctx context.Context, phone string) (*StaffMember, error) {
	query := `
		SELECT id, store_id, chain_id, phone, name, role, pin_hash, pin_set_at, pin_failed_attempts, pin_locked_until, is_active, created_by, created_at, updated_at
		FROM staff_members
		WHERE phone = $1 AND is_active = true
	`
	var s StaffMember
	err := p.database.QueryRowContext(ctx, query, phone).Scan(
		&s.ID, &s.StoreID, &s.ChainID, &s.Phone, &s.Name, &s.Role, &s.PinHash, &s.PinSetAt, &s.PinFailedAttempts, &s.PinLockedUntil, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *PostgresRepository) GetStaffByID(ctx context.Context, id string) (*StaffMember, error) {
	query := `
		SELECT id, store_id, chain_id, phone, name, role, pin_hash, pin_set_at, pin_failed_attempts, pin_locked_until, is_active, created_by, created_at, updated_at
		FROM staff_members
		WHERE id = $1
	`
	var s StaffMember
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.StoreID, &s.ChainID, &s.Phone, &s.Name, &s.Role, &s.PinHash, &s.PinSetAt, &s.PinFailedAttempts, &s.PinLockedUntil, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *PostgresRepository) ListStaffByStore(ctx context.Context, storeID string, roleFilter string, includeInactive bool, page, pageSize int) ([]*StaffMember, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "store_id = $1"
	args := []interface{}{storeID}
	argCount := 1

	if !includeInactive {
		where += " AND is_active = true"
	}
	if roleFilter != "" {
		argCount++
		where += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, roleFilter)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM staff_members WHERE %s", where)
	if err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argCount++
	limitArg := argCount
	args = append(args, pageSize)

	argCount++
	offsetArg := argCount
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT id, store_id, chain_id, phone, name, role, pin_hash, pin_set_at, pin_failed_attempts, pin_locked_until, is_active, created_by, created_at, updated_at
		FROM staff_members
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArg, offsetArg)

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var staff []*StaffMember
	for rows.Next() {
		var s StaffMember
		if err := rows.Scan(
			&s.ID, &s.StoreID, &s.ChainID, &s.Phone, &s.Name, &s.Role, &s.PinHash, &s.PinSetAt, &s.PinFailedAttempts, &s.PinLockedUntil, &s.IsActive, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		staff = append(staff, &s)
	}

	if staff == nil {
		staff = []*StaffMember{}
	}

	return staff, total, nil
}

func (p *PostgresRepository) UpdateStaffMember(ctx context.Context, s *StaffMember) error {
	s.UpdatedAt = time.Now()
	query := `
		UPDATE staff_members
		SET name = COALESCE(NULLIF($1, ''), name),
		    role = COALESCE(NULLIF($2, ''), role),
		    updated_at = $3
		WHERE id = $4 AND is_active = true
	`
	res, err := p.database.ExecContext(ctx, query, s.Name, s.Role, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrStaffNotFound
	}
	return nil
}

func (p *PostgresRepository) DeactivateStaffMemberTx(ctx context.Context, staffID string) error {
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	// 1. Soft delete staff member
	_, err = tx.ExecContext(ctx, `UPDATE staff_members SET is_active = false, updated_at = $1 WHERE id = $2`, now, staffID)
	if err != nil {
		return err
	}

	// 2. Revoke all active sessions
	_, err = tx.ExecContext(ctx, `UPDATE staff_sessions SET revoked_at = $1 WHERE staff_id = $2 AND revoked_at IS NULL`, now, staffID)
	if err != nil {
		return err
	}

	// 3. End any open shift
	_, err = tx.ExecContext(ctx, `UPDATE staff_shifts SET ended_at = $1 WHERE staff_id = $2 AND ended_at IS NULL`, now, staffID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *PostgresRepository) CreateSession(ctx context.Context, sess *StaffSession) error {
	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO staff_sessions (id, staff_id, device_id, created_at, revoked_at, auth_method)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := p.database.ExecContext(ctx, query, sess.ID, sess.StaffID, sess.DeviceID, sess.CreatedAt, sess.RevokedAt, sess.AuthMethod)
	return err
}

func (p *PostgresRepository) GetSessionByID(ctx context.Context, id string) (*StaffSession, error) {
	query := `
		SELECT id, staff_id, device_id, created_at, revoked_at, auth_method
		FROM staff_sessions
		WHERE id = $1
	`
	var s StaffSession
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.StaffID, &s.DeviceID, &s.CreatedAt, &s.RevokedAt, &s.AuthMethod,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *PostgresRepository) RevokeSession(ctx context.Context, id string) error {
	query := `UPDATE staff_sessions SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL`
	_, err := p.database.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (p *PostgresRepository) UpdateStaffPin(ctx context.Context, staffID string, pinHash string) error {
	now := time.Now()
	query := `
		UPDATE staff_members
		SET pin_hash = $1, pin_set_at = $2, pin_failed_attempts = 0, pin_locked_until = NULL, updated_at = $2
		WHERE id = $3 AND is_active = true
	`
	res, err := p.database.ExecContext(ctx, query, pinHash, now, staffID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrStaffNotFound
	}
	return nil
}

func (p *PostgresRepository) UpdatePinLockout(ctx context.Context, staffID string, failedAttempts int, lockedUntil *time.Time) error {
	query := `
		UPDATE staff_members
		SET pin_failed_attempts = $1, pin_locked_until = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := p.database.ExecContext(ctx, query, failedAttempts, lockedUntil, time.Now(), staffID)
	return err
}

func (p *PostgresRepository) StartShiftTx(ctx context.Context, shift *StaffShift) error {
	if shift.ID == "" {
		shift.ID = uuid.New().String()
	}
	shift.StartedAt = time.Now()

	query := `
		INSERT INTO staff_shifts (id, staff_id, store_id, started_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := p.database.ExecContext(ctx, query, shift.ID, shift.StaffID, shift.StoreID, shift.StartedAt)
	if err != nil {
		// Unique partial index violation check
		return ErrShiftActive
	}
	return nil
}

func (p *PostgresRepository) EndShift(ctx context.Context, staffID string) error {
	now := time.Now()
	query := `
		UPDATE staff_shifts
		SET ended_at = $1
		WHERE staff_id = $2 AND ended_at IS NULL
	`
	res, err := p.database.ExecContext(ctx, query, now, staffID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNoActiveShift
	}
	return nil
}

func (p *PostgresRepository) GetActiveShift(ctx context.Context, staffID string) (*StaffShift, error) {
	query := `
		SELECT s.id, s.staff_id, s.store_id, s.started_at, s.ended_at, m.name
		FROM staff_shifts s
		JOIN staff_members m ON s.staff_id = m.id
		WHERE s.staff_id = $1 AND s.ended_at IS NULL
	`
	var s StaffShift
	err := p.database.QueryRowContext(ctx, query, staffID).Scan(
		&s.ID, &s.StaffID, &s.StoreID, &s.StartedAt, &s.EndedAt, &s.StaffName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *PostgresRepository) GetShiftHistory(ctx context.Context, storeID string, dateFrom, dateTo *time.Time, page, pageSize int) ([]*StaffShift, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "s.store_id = $1"
	args := []interface{}{storeID}
	argCount := 1

	if dateFrom != nil {
		argCount++
		where += fmt.Sprintf(" AND s.started_at >= $%d", argCount)
		args = append(args, *dateFrom)
	}
	if dateTo != nil {
		argCount++
		where += fmt.Sprintf(" AND s.started_at <= $%d", argCount)
		args = append(args, *dateTo)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM staff_shifts s WHERE %s", where)
	if err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	argCount++
	limitArg := argCount
	args = append(args, pageSize)

	argCount++
	offsetArg := argCount
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT s.id, s.staff_id, s.store_id, s.started_at, s.ended_at, m.name
		FROM staff_shifts s
		JOIN staff_members m ON s.staff_id = m.id
		WHERE %s
		ORDER BY s.started_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArg, offsetArg)

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []*StaffShift
	for rows.Next() {
		var sh StaffShift
		if err := rows.Scan(&sh.ID, &sh.StaffID, &sh.StoreID, &sh.StartedAt, &sh.EndedAt, &sh.StaffName); err != nil {
			return nil, 0, err
		}
		shifts = append(shifts, &sh)
	}

	if shifts == nil {
		shifts = []*StaffShift{}
	}

	return shifts, total, nil
}

// MemoryRepository for testing
type MemoryRepository struct {
	mu       sync.RWMutex
	staff    map[string]*StaffMember
	sessions map[string]*StaffSession
	shifts   map[string]*StaffShift
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		staff:    make(map[string]*StaffMember),
		sessions: make(map[string]*StaffSession),
		shifts:   make(map[string]*StaffShift),
	}
}

func (m *MemoryRepository) CreateStaffMember(ctx context.Context, s *StaffMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.staff {
		if existing.Phone == s.Phone {
			return ErrPhoneConflict
		}
	}

	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.IsActive = true
	m.staff[s.ID] = s
	return nil
}

func (m *MemoryRepository) GetStaffByPhone(ctx context.Context, phone string) (*StaffMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.staff {
		if s.Phone == phone && s.IsActive {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetStaffByID(ctx context.Context, id string) (*StaffMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.staff[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *MemoryRepository) ListStaffByStore(ctx context.Context, storeID string, roleFilter string, includeInactive bool, page, pageSize int) ([]*StaffMember, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*StaffMember
	for _, s := range m.staff {
		if s.StoreID != storeID {
			continue
		}
		if !includeInactive && !s.IsActive {
			continue
		}
		if roleFilter != "" && s.Role != roleFilter {
			continue
		}
		result = append(result, s)
	}

	total := int64(len(result))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= len(result) {
		return []*StaffMember{}, total, nil
	}

	end := offset + pageSize
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], total, nil
}

func (m *MemoryRepository) UpdateStaffMember(ctx context.Context, s *StaffMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.staff[s.ID]
	if !ok || !existing.IsActive {
		return ErrStaffNotFound
	}

	if s.Name != "" {
		existing.Name = s.Name
	}
	if s.Role != "" {
		existing.Role = s.Role
	}
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) DeactivateStaffMemberTx(ctx context.Context, staffID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.staff[staffID]
	if !ok {
		return ErrStaffNotFound
	}

	now := time.Now()
	s.IsActive = false
	s.UpdatedAt = now

	for _, sess := range m.sessions {
		if sess.StaffID == staffID && sess.RevokedAt == nil {
			sess.RevokedAt = &now
		}
	}

	for _, sh := range m.shifts {
		if sh.StaffID == staffID && sh.EndedAt == nil {
			sh.EndedAt = &now
		}
	}

	return nil
}

func (m *MemoryRepository) CreateSession(ctx context.Context, sess *StaffSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = time.Now()
	}
	m.sessions[sess.ID] = sess
	return nil
}

func (m *MemoryRepository) GetSessionByID(ctx context.Context, id string) (*StaffSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return sess, nil
}

func (m *MemoryRepository) RevokeSession(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[id]
	if !ok {
		return nil
	}
	now := time.Now()
	sess.RevokedAt = &now
	return nil
}

func (m *MemoryRepository) UpdateStaffPin(ctx context.Context, staffID string, pinHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.staff[staffID]
	if !ok || !s.IsActive {
		return ErrStaffNotFound
	}
	now := time.Now()
	s.PinHash = &pinHash
	s.PinSetAt = &now
	s.PinFailedAttempts = 0
	s.PinLockedUntil = nil
	s.UpdatedAt = now
	return nil
}

func (m *MemoryRepository) UpdatePinLockout(ctx context.Context, staffID string, failedAttempts int, lockedUntil *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.staff[staffID]
	if !ok {
		return ErrStaffNotFound
	}
	s.PinFailedAttempts = failedAttempts
	s.PinLockedUntil = lockedUntil
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) StartShiftTx(ctx context.Context, shift *StaffShift) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sh := range m.shifts {
		if sh.StaffID == shift.StaffID && sh.EndedAt == nil {
			return ErrShiftActive
		}
	}

	if shift.ID == "" {
		shift.ID = uuid.New().String()
	}
	shift.StartedAt = time.Now()
	m.shifts[shift.ID] = shift
	return nil
}

func (m *MemoryRepository) EndShift(ctx context.Context, staffID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var found *StaffShift
	for _, sh := range m.shifts {
		if sh.StaffID == staffID && sh.EndedAt == nil {
			found = sh
			break
		}
	}
	if found == nil {
		return ErrNoActiveShift
	}
	now := time.Now()
	found.EndedAt = &now
	return nil
}

func (m *MemoryRepository) GetActiveShift(ctx context.Context, staffID string) (*StaffShift, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sh := range m.shifts {
		if sh.StaffID == staffID && sh.EndedAt == nil {
			if s, ok := m.staff[staffID]; ok {
				sh.StaffName = s.Name
			}
			return sh, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetShiftHistory(ctx context.Context, storeID string, dateFrom, dateTo *time.Time, page, pageSize int) ([]*StaffShift, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*StaffShift
	for _, sh := range m.shifts {
		if sh.StoreID != storeID {
			continue
		}
		if dateFrom != nil && sh.StartedAt.Before(*dateFrom) {
			continue
		}
		if dateTo != nil && sh.StartedAt.After(*dateTo) {
			continue
		}
		if s, ok := m.staff[sh.StaffID]; ok {
			sh.StaffName = s.Name
		}
		matched = append(matched, sh)
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
		return []*StaffShift{}, total, nil
	}

	end := offset + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (p *PostgresRepository) AnonymizeStaffMember(ctx context.Context, staffID string) error {
	tombstonePhone := fmt.Sprintf("deleted_%s", staffID)
	query := `UPDATE staff_members SET name = 'Anonymized Staff Member', phone = $1, is_active = false, updated_at = NOW() WHERE id = $2`
	_, err := p.database.ExecContext(ctx, query, tombstonePhone, staffID)
	return err
}

func (m *MemoryRepository) AnonymizeStaffMember(ctx context.Context, staffID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.staff[staffID]
	if !ok {
		return ErrStaffNotFound
	}
	s.Name = "Anonymized Staff Member"
	s.Phone = fmt.Sprintf("deleted_%s", staffID)
	s.IsActive = false
	s.UpdatedAt = time.Now().UTC()
	return nil
}
