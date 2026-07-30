package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var (
	ErrChainNotFound  = errors.New("chain not found")
	ErrChainSuspended = errors.New("chain suspended")
)

type Repository interface {
	GetNearbyStores(ctx context.Context, lat, lng, radiusKM float64) ([]*Store, error)
	GetStoreByID(ctx context.Context, id string) (*Store, error)
	// CreateStore validates chain status locally — only used by legacy code paths still in the DB.
	// New store creation goes through CreateStoreInternal (called by InternalAdminWriteHandler).
	CreateStore(ctx context.Context, s *Store) error
	// CreateStoreInternal inserts a store row WITHOUT chain validation — admin-store-service
	// has already validated chain status before calling this endpoint.
	CreateStoreInternal(ctx context.Context, s *Store) error
	ListStores(ctx context.Context, chainIDFilter, statusFilter string, page, pageSize int) ([]*Store, int64, error)
	UpdateGeofence(ctx context.Context, storeID string, polygon []Point, radiusMeters int) error
	UpdateHours(ctx context.Context, storeID, openingTime, closingTime, timezone string) error
	UpdateCapacity(ctx context.Context, storeID string, capacityMax int) error
	UpdateStatus(ctx context.Context, storeID, status string) error
	UpdatePaymentSetup(ctx context.Context, storeID, razorpayAccountID, razorpayKYCStatus, note string) error
	GetActiveQRToken(ctx context.Context, token string) (*StoreQRToken, error)
	GetActiveQRTokens(ctx context.Context, storeID string) ([]*StoreQRToken, error)
	CreateQRToken(ctx context.Context, token *StoreQRToken) error
	DeactivateQRTokens(ctx context.Context, storeID, gateID string) error
	GetActiveSessionByUser(ctx context.Context, userID string) (*StoreSession, error)
	CreateSession(ctx context.Context, sess *StoreSession) error
	UnbindSession(ctx context.Context, sessionID string) (*StoreSession, error)
	UnbindUserActiveSession(ctx context.Context, userID string) (*StoreSession, error)
	AutoExpireStaleSessions(ctx context.Context, staleDuration time.Duration) ([]*StoreSession, error)

	// Chain methods — retained for backward compatibility during migration bake period.
	// After chains table is confirmed dropped, these stubs return ErrChainNotFound.
	CreateChain(ctx context.Context, c *Chain) error
	ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error)
	GetChainByID(ctx context.Context, id string) (*Chain, error)
	UpdateChain(ctx context.Context, c *Chain) error
	UpdateChainStatus(ctx context.Context, id, status string) error
}

type PostgresRepository struct {
	database *db.DB
}

func NewPostgresRepository(database *db.DB) Repository {
	return &PostgresRepository{database: database}
}

func (p *PostgresRepository) GetNearbyStores(ctx context.Context, lat, lng, radiusKM float64) ([]*Store, error) {
	if radiusKM > 25 {
		radiusKM = 25
	}

	query := `
		SELECT id, chain_id, name, address, city, state, pincode, gstin, lat, lng,
		       geofence_polygon, geofence_radius_meters, capacity_max,
		       COALESCE(opening_time::text, ''), COALESCE(closing_time::text, ''), timezone,
		       status, rfid_enabled, catalog_version, created_at, updated_at
		FROM stores
		WHERE status = 'ACTIVE'
		  AND (6371 * acos(cos(radians($1)) * cos(radians(lat)) * cos(radians(lng) - radians($2)) + sin(radians($1)) * sin(radians(lat)))) <= $3
		ORDER BY (6371 * acos(cos(radians($1)) * cos(radians(lat)) * cos(radians(lng) - radians($2)) + sin(radians($1)) * sin(radians(lat)))) ASC
	`

	rows, err := p.database.QueryContext(ctx, query, lat, lng, radiusKM)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*Store
	for rows.Next() {
		var s Store
		var polygonJSON []byte
		if err := rows.Scan(
			&s.ID, &s.ChainID, &s.Name, &s.Address, &s.City, &s.State, &s.Pincode, &s.GSTIN, &s.Lat, &s.Lng,
			&polygonJSON, &s.GeofenceRadiusMeters, &s.CapacityMax, &s.OpeningTime, &s.ClosingTime, &s.Timezone,
			&s.Status, &s.RFIDEnabled, &s.CatalogVersion, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(polygonJSON) > 0 {
			_ = json.Unmarshal(polygonJSON, &s.GeofencePolygon)
		}
		stores = append(stores, &s)
	}

	if stores == nil {
		stores = []*Store{}
	}
	return stores, nil
}

func (p *PostgresRepository) GetStoreByID(ctx context.Context, id string) (*Store, error) {
	query := `
		SELECT id, chain_id, name, address, city, state, pincode, gstin, lat, lng,
		       geofence_polygon, geofence_radius_meters, capacity_max,
		       COALESCE(opening_time::text, ''), COALESCE(closing_time::text, ''), timezone,
		       status, rfid_enabled, catalog_version,
		       COALESCE(razorpay_account_id, ''), COALESCE(razorpay_kyc_status, ''), COALESCE(payment_setup_note, ''),
		       created_at, updated_at
		FROM stores
		WHERE id = $1
	`
	var s Store
	var polygonJSON []byte
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.ChainID, &s.Name, &s.Address, &s.City, &s.State, &s.Pincode, &s.GSTIN, &s.Lat, &s.Lng,
		&polygonJSON, &s.GeofenceRadiusMeters, &s.CapacityMax, &s.OpeningTime, &s.ClosingTime, &s.Timezone,
		&s.Status, &s.RFIDEnabled, &s.CatalogVersion,
		&s.RazorpayAccountID, &s.RazorpayKYCStatus, &s.PaymentSetupNote,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(polygonJSON) > 0 {
		_ = json.Unmarshal(polygonJSON, &s.GeofencePolygon)
	}
	return &s, nil
}

func (p *PostgresRepository) createStoreRow(ctx context.Context, s *Store) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = "ACTIVE"
	}
	if s.CatalogVersion == 0 {
		s.CatalogVersion = 1
	}

	var polygonJSON []byte
	if len(s.GeofencePolygon) > 0 {
		polygonJSON, _ = json.Marshal(s.GeofencePolygon)
	}

	query := `
		INSERT INTO stores (id, chain_id, name, address, city, state, pincode, gstin, lat, lng,
		                    geofence_polygon, geofence_radius_meters, capacity_max, opening_time, closing_time,
		                    timezone, status, rfid_enabled, catalog_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::time, $15::time, $16, $17, $18, $19, $20, $21)
	`
	_, err := p.database.ExecContext(ctx, query,
		s.ID, s.ChainID, s.Name, s.Address, s.City, s.State, s.Pincode, s.GSTIN, s.Lat, s.Lng,
		polygonJSON, s.GeofenceRadiusMeters, s.CapacityMax, s.OpeningTime, s.ClosingTime, s.Timezone,
		s.Status, s.RFIDEnabled, s.CatalogVersion, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

// CreateStore validates chain status before inserting (legacy path, used by existing tests).
func (p *PostgresRepository) CreateStore(ctx context.Context, s *Store) error {
	chain, err := p.GetChainByID(ctx, s.ChainID)
	if err != nil || chain == nil {
		return ErrChainNotFound
	}
	if chain.Status != ChainStatusActive {
		return ErrChainSuspended
	}
	return p.createStoreRow(ctx, s)
}

// CreateStoreInternal inserts a store row without chain validation.
// Chain status was already validated by admin-store-service before calling this.
func (p *PostgresRepository) CreateStoreInternal(ctx context.Context, s *Store) error {
	return p.createStoreRow(ctx, s)
}

func (p *PostgresRepository) UpdateGeofence(ctx context.Context, storeID string, polygon []Point, radiusMeters int) error {
	now := time.Now()
	var polygonJSON []byte
	if len(polygon) > 0 {
		polygonJSON, _ = json.Marshal(polygon)
	}

	query := `
		UPDATE stores
		SET geofence_polygon = $1, geofence_radius_meters = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := p.database.ExecContext(ctx, query, polygonJSON, radiusMeters, now, storeID)
	return err
}

func (p *PostgresRepository) UpdateHours(ctx context.Context, storeID, openingTime, closingTime, timezone string) error {
	now := time.Now()
	query := `
		UPDATE stores
		SET opening_time = $1::time, closing_time = $2::time, timezone = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := p.database.ExecContext(ctx, query, openingTime, closingTime, timezone, now, storeID)
	return err
}

func (p *PostgresRepository) UpdateCapacity(ctx context.Context, storeID string, capacityMax int) error {
	now := time.Now()
	query := `
		UPDATE stores
		SET capacity_max = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := p.database.ExecContext(ctx, query, capacityMax, now, storeID)
	return err
}

func (p *PostgresRepository) UpdateStatus(ctx context.Context, storeID, status string) error {
	now := time.Now()
	query := `
		UPDATE stores
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := p.database.ExecContext(ctx, query, status, now, storeID)
	return err
}

func (p *PostgresRepository) UpdatePaymentSetup(ctx context.Context, storeID, razorpayAccountID, razorpayKYCStatus, note string) error {
	now := time.Now()
	query := `
		UPDATE stores
		SET razorpay_account_id = $1, razorpay_kyc_status = $2, payment_setup_note = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := p.database.ExecContext(ctx, query, razorpayAccountID, razorpayKYCStatus, note, now, storeID)
	return err
}

func (p *PostgresRepository) GetActiveQRToken(ctx context.Context, token string) (*StoreQRToken, error) {
	query := `
		SELECT id, store_id, gate_id, token, is_active, expires_at, created_at
		FROM store_qr_tokens
		WHERE token = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP
	`
	var t StoreQRToken
	err := p.database.QueryRowContext(ctx, query, token).Scan(
		&t.ID, &t.StoreID, &t.GateID, &t.Token, &t.IsActive, &t.ExpiresAt, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *PostgresRepository) GetActiveQRTokens(ctx context.Context, storeID string) ([]*StoreQRToken, error) {
	query := `
		SELECT id, store_id, gate_id, token, is_active, expires_at, created_at
		FROM store_qr_tokens
		WHERE store_id = $1 AND is_active = true AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
	`
	rows, err := p.database.QueryContext(ctx, query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*StoreQRToken
	for rows.Next() {
		var t StoreQRToken
		if err := rows.Scan(&t.ID, &t.StoreID, &t.GateID, &t.Token, &t.IsActive, &t.ExpiresAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, &t)
	}

	if tokens == nil {
		tokens = []*StoreQRToken{}
	}
	return tokens, nil
}

func (p *PostgresRepository) CreateQRToken(ctx context.Context, token *StoreQRToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	token.CreatedAt = time.Now()

	query := `
		INSERT INTO store_qr_tokens (id, store_id, gate_id, token, is_active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.database.ExecContext(ctx, query, token.ID, token.StoreID, token.GateID, token.Token, token.IsActive, token.ExpiresAt, token.CreatedAt)
	return err
}

func (p *PostgresRepository) DeactivateQRTokens(ctx context.Context, storeID, gateID string) error {
	query := `
		UPDATE store_qr_tokens
		SET is_active = false
		WHERE store_id = $1 AND gate_id = $2 AND is_active = true
	`
	_, err := p.database.ExecContext(ctx, query, storeID, gateID)
	return err
}

func (p *PostgresRepository) GetActiveSessionByUser(ctx context.Context, userID string) (*StoreSession, error) {
	query := `
		SELECT id, user_id, store_id, COALESCE(device_id, ''), bound_at, unbound_at, catalog_version_at_bind
		FROM store_sessions
		WHERE user_id = $1 AND unbound_at IS NULL
	`
	var sess StoreSession
	err := p.database.QueryRowContext(ctx, query, userID).Scan(
		&sess.ID, &sess.UserID, &sess.StoreID, &sess.DeviceID, &sess.BoundAt, &sess.UnboundAt, &sess.CatalogVersionAtBind,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (p *PostgresRepository) CreateSession(ctx context.Context, sess *StoreSession) error {
	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.BoundAt.IsZero() {
		sess.BoundAt = time.Now()
	}

	query := `
		INSERT INTO store_sessions (id, user_id, store_id, device_id, bound_at, catalog_version_at_bind)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := p.database.ExecContext(ctx, query, sess.ID, sess.UserID, sess.StoreID, sess.DeviceID, sess.BoundAt, sess.CatalogVersionAtBind)
	return err
}

func (p *PostgresRepository) UnbindSession(ctx context.Context, sessionID string) (*StoreSession, error) {
	now := time.Now()
	query := `
		UPDATE store_sessions
		SET unbound_at = $1
		WHERE id = $2 AND unbound_at IS NULL
		RETURNING id, user_id, store_id, COALESCE(device_id, ''), bound_at, unbound_at, catalog_version_at_bind
	`
	var sess StoreSession
	err := p.database.QueryRowContext(ctx, query, now, sessionID).Scan(
		&sess.ID, &sess.UserID, &sess.StoreID, &sess.DeviceID, &sess.BoundAt, &sess.UnboundAt, &sess.CatalogVersionAtBind,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (p *PostgresRepository) UnbindUserActiveSession(ctx context.Context, userID string) (*StoreSession, error) {
	now := time.Now()
	query := `
		UPDATE store_sessions
		SET unbound_at = $1
		WHERE user_id = $2 AND unbound_at IS NULL
		RETURNING id, user_id, store_id, COALESCE(device_id, ''), bound_at, unbound_at, catalog_version_at_bind
	`
	var sess StoreSession
	err := p.database.QueryRowContext(ctx, query, now, userID).Scan(
		&sess.ID, &sess.UserID, &sess.StoreID, &sess.DeviceID, &sess.BoundAt, &sess.UnboundAt, &sess.CatalogVersionAtBind,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (p *PostgresRepository) AutoExpireStaleSessions(ctx context.Context, staleDuration time.Duration) ([]*StoreSession, error) {
	cutoff := time.Now().Add(-staleDuration)
	query := `
		UPDATE store_sessions
		SET unbound_at = CURRENT_TIMESTAMP
		WHERE unbound_at IS NULL AND bound_at < $1
		RETURNING id, user_id, store_id, COALESCE(device_id, ''), bound_at, unbound_at, catalog_version_at_bind
	`
	rows, err := p.database.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []*StoreSession
	for rows.Next() {
		var s StoreSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.StoreID, &s.DeviceID, &s.BoundAt, &s.UnboundAt, &s.CatalogVersionAtBind); err != nil {
			return nil, err
		}
		expired = append(expired, &s)
	}

	if expired == nil {
		expired = []*StoreSession{}
	}
	return expired, nil
}

// Chain Methods Implementation
func (p *PostgresRepository) CreateChain(ctx context.Context, c *Chain) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = ChainStatusActive
	}

	query := `
		INSERT INTO chains (id, name, legal_entity_name, default_gstin_prefix, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.database.ExecContext(ctx, query, c.ID, c.Name, c.LegalEntityName, c.DefaultGstInPrefix, c.Status, c.CreatedAt, c.UpdatedAt)
	return err
}

func (p *PostgresRepository) ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where := "1=1"
	args := []interface{}{}
	if statusFilter != "" {
		where += " AND c.status = $1"
		args = append(args, statusFilter)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM chains c WHERE %s", where)
	if err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT c.id, c.name, COALESCE(c.legal_entity_name, ''), COALESCE(c.default_gstin_prefix, ''), c.status, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM stores s WHERE s.chain_id = c.id) as store_count
		FROM chains c
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArg, offsetArg)

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var chains []*Chain
	for rows.Next() {
		var c Chain
		if err := rows.Scan(&c.ID, &c.Name, &c.LegalEntityName, &c.DefaultGstInPrefix, &c.Status, &c.CreatedAt, &c.UpdatedAt, &c.StoreCount); err != nil {
			return nil, 0, err
		}
		chains = append(chains, &c)
	}

	if chains == nil {
		chains = []*Chain{}
	}

	return chains, total, nil
}

func (p *PostgresRepository) GetChainByID(ctx context.Context, id string) (*Chain, error) {
	query := `
		SELECT c.id, c.name, COALESCE(c.legal_entity_name, ''), COALESCE(c.default_gstin_prefix, ''), c.status, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM stores s WHERE s.chain_id = c.id) as store_count
		FROM chains c
		WHERE c.id = $1
	`
	var c Chain
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.LegalEntityName, &c.DefaultGstInPrefix, &c.Status, &c.CreatedAt, &c.UpdatedAt, &c.StoreCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *PostgresRepository) UpdateChain(ctx context.Context, c *Chain) error {
	now := time.Now()
	c.UpdatedAt = now
	query := `
		UPDATE chains
		SET name = COALESCE(NULLIF($1, ''), name),
		    legal_entity_name = COALESCE(NULLIF($2, ''), legal_entity_name),
		    default_gstin_prefix = COALESCE(NULLIF($3, ''), default_gstin_prefix),
		    status = COALESCE(NULLIF($4, ''), status),
		    updated_at = $5
		WHERE id = $6
	`
	_, err := p.database.ExecContext(ctx, query, c.Name, c.LegalEntityName, c.DefaultGstInPrefix, c.Status, now, c.ID)
	return err
}

func (p *PostgresRepository) UpdateChainStatus(ctx context.Context, id, status string) error {
	now := time.Now()
	query := `UPDATE chains SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := p.database.ExecContext(ctx, query, status, now, id)
	return err
}

// MemoryRepository Implementation for Tests
type MemoryRepository struct {
	mu       sync.RWMutex
	chains   map[string]*Chain
	stores   map[string]*Store
	tokens   map[string]*StoreQRToken
	sessions map[string]*StoreSession
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		chains:   make(map[string]*Chain),
		stores:   make(map[string]*Store),
		tokens:   make(map[string]*StoreQRToken),
		sessions: make(map[string]*StoreSession),
	}
	// Seed default chains for backward compatible test runs
	for _, id := range []string{"chain-001", "chain-hq-001", "chain-A", "chain-B", "chain-indiranagar"} {
		repo.chains[id] = &Chain{
			ID:     id,
			Name:   "Default Chain " + id,
			Status: ChainStatusActive,
		}
	}
	return repo
}

func (m *MemoryRepository) GetNearbyStores(ctx context.Context, lat, lng, radiusKM float64) ([]*Store, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Store
	for _, s := range m.stores {
		if s.Status == "ACTIVE" {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MemoryRepository) GetStoreByID(ctx context.Context, id string) (*Store, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.stores[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *MemoryRepository) CreateStore(ctx context.Context, s *Store) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	chain, ok := m.chains[s.ChainID]
	if !ok && s.ChainID != "" {
		chain = &Chain{ID: s.ChainID, Name: "Chain " + s.ChainID, Status: ChainStatusActive}
		m.chains[s.ChainID] = chain
	}
	if chain != nil && chain.Status != ChainStatusActive {
		return ErrChainSuspended
	}

	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = "ACTIVE"
	}
	m.stores[s.ID] = s
	return nil
}

// CreateStoreInternal inserts a store without chain validation (admin-store-service already validated).
func (m *MemoryRepository) CreateStoreInternal(ctx context.Context, s *Store) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = "ACTIVE"
	}
	m.stores[s.ID] = s
	return nil
}

func (m *MemoryRepository) UpdateGeofence(ctx context.Context, storeID string, polygon []Point, radiusMeters int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stores[storeID]
	if !ok {
		return errors.New("store not found")
	}
	s.GeofencePolygon = polygon
	s.GeofenceRadiusMeters = radiusMeters
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) UpdateHours(ctx context.Context, storeID, openingTime, closingTime, timezone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stores[storeID]
	if !ok {
		return errors.New("store not found")
	}
	s.OpeningTime = openingTime
	s.ClosingTime = closingTime
	s.Timezone = timezone
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) UpdateCapacity(ctx context.Context, storeID string, capacityMax int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stores[storeID]
	if !ok {
		return errors.New("store not found")
	}
	s.CapacityMax = capacityMax
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) UpdateStatus(ctx context.Context, storeID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stores[storeID]
	if !ok {
		return errors.New("store not found")
	}
	s.Status = status
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) UpdatePaymentSetup(ctx context.Context, storeID, razorpayAccountID, razorpayKYCStatus, note string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stores[storeID]
	if !ok {
		return errors.New("store not found")
	}
	s.RazorpayAccountID = razorpayAccountID
	s.RazorpayKYCStatus = razorpayKYCStatus
	s.PaymentSetupNote = note
	s.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) GetActiveQRToken(ctx context.Context, token string) (*StoreQRToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.tokens {
		if t.Token == token && t.IsActive && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) GetActiveQRTokens(ctx context.Context, storeID string) ([]*StoreQRToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*StoreQRToken
	for _, t := range m.tokens {
		if t.StoreID == storeID && t.IsActive && t.ExpiresAt.After(time.Now()) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MemoryRepository) CreateQRToken(ctx context.Context, token *StoreQRToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	token.CreatedAt = time.Now()
	m.tokens[token.ID] = token
	return nil
}

func (m *MemoryRepository) DeactivateQRTokens(ctx context.Context, storeID, gateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.tokens {
		if t.StoreID == storeID && t.GateID == gateID && t.IsActive {
			t.IsActive = false
		}
	}
	return nil
}

func (m *MemoryRepository) GetActiveSessionByUser(ctx context.Context, userID string) (*StoreSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.sessions {
		if s.UserID == userID && s.UnboundAt == nil {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) CreateSession(ctx context.Context, sess *StoreSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.BoundAt.IsZero() {
		sess.BoundAt = time.Now()
	}
	m.sessions[sess.ID] = sess
	return nil
}

func (m *MemoryRepository) UnbindSession(ctx context.Context, sessionID string) (*StoreSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[sessionID]
	if !ok || sess.UnboundAt != nil {
		return nil, nil
	}
	now := time.Now()
	sess.UnboundAt = &now
	return sess, nil
}

func (m *MemoryRepository) UnbindUserActiveSession(ctx context.Context, userID string) (*StoreSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		if s.UserID == userID && s.UnboundAt == nil {
			now := time.Now()
			s.UnboundAt = &now
			return s, nil
		}
	}
	return nil, nil
}

func (m *MemoryRepository) AutoExpireStaleSessions(ctx context.Context, staleDuration time.Duration) ([]*StoreSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expired []*StoreSession
	cutoff := time.Now().Add(-staleDuration)
	for _, s := range m.sessions {
		if s.UnboundAt == nil && s.BoundAt.Before(cutoff) {
			now := time.Now()
			s.UnboundAt = &now
			expired = append(expired, s)
		}
	}
	return expired, nil
}

func (m *MemoryRepository) CreateChain(ctx context.Context, c *Chain) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = ChainStatusActive
	}
	m.chains[c.ID] = c
	return nil
}

func (m *MemoryRepository) ListChains(ctx context.Context, statusFilter string, page, pageSize int) ([]*Chain, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Chain
	for _, c := range m.chains {
		if statusFilter != "" && c.Status != statusFilter {
			continue
		}
		// Count stores under this chain
		count := 0
		for _, s := range m.stores {
			if s.ChainID == c.ID {
				count++
			}
		}
		c.StoreCount = count
		result = append(result, c)
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
		return []*Chain{}, total, nil
	}

	end := offset + pageSize
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], total, nil
}

func (m *MemoryRepository) GetChainByID(ctx context.Context, id string) (*Chain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.chains[id]
	if !ok {
		return nil, nil
	}
	count := 0
	for _, s := range m.stores {
		if s.ChainID == c.ID {
			count++
		}
	}
	c.StoreCount = count
	return c, nil
}

func (m *MemoryRepository) UpdateChain(ctx context.Context, c *Chain) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.chains[c.ID]
	if !ok {
		return ErrChainNotFound
	}
	if c.Name != "" {
		existing.Name = c.Name
	}
	if c.LegalEntityName != "" {
		existing.LegalEntityName = c.LegalEntityName
	}
	if c.DefaultGstInPrefix != "" {
		existing.DefaultGstInPrefix = c.DefaultGstInPrefix
	}
	if c.Status != "" {
		existing.Status = c.Status
	}
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MemoryRepository) UpdateChainStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.chains[id]
	if !ok {
		return ErrChainNotFound
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}

func (p *PostgresRepository) ListStores(ctx context.Context, chainIDFilter, statusFilter string, page, pageSize int) ([]*Store, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}
	idx := 1

	if chainIDFilter != "" {
		conditions = append(conditions, fmt.Sprintf("chain_id = $%d", idx))
		args = append(args, chainIDFilter)
		idx++
	}
	if statusFilter != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
		args = append(args, statusFilter)
		idx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stores %s", whereClause)
	var total int64
	err := p.database.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf("SELECT id, chain_id, name, address, city, state, pincode, gstin, status, created_at, updated_at FROM stores %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", whereClause, idx, idx+1)
	args = append(args, pageSize, offset)

	rows, err := p.database.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*Store
	for rows.Next() {
		s := &Store{}
		if err := rows.Scan(&s.ID, &s.ChainID, &s.Name, &s.Address, &s.City, &s.State, &s.Pincode, &s.GSTIN, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, s)
	}

	return result, total, nil
}

func (m *MemoryRepository) ListStores(ctx context.Context, chainIDFilter, statusFilter string, page, pageSize int) ([]*Store, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*Store
	for _, s := range m.stores {
		if chainIDFilter != "" && s.ChainID != chainIDFilter {
			continue
		}
		if statusFilter != "" && s.Status != statusFilter {
			continue
		}
		copy := *s
		matched = append(matched, &copy)
	}

	total := int64(len(matched))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= len(matched) {
		return []*Store{}, total, nil
	}
	end := offset + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}
