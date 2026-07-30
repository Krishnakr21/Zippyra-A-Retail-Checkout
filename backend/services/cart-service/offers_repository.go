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
)

var ErrOfferNotFound = errors.New("offer not found")

// MemoryRedis provides an in-memory fallback for redis.Cmdable in unit tests
type MemoryRedis struct {
	redis.Cmdable
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryRedis() *MemoryRedis {
	return &MemoryRedis{data: make(map[string]string)}
}

func (m *MemoryRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = fmt.Sprintf("%v", value)
	cmd.SetVal("OK")
	return cmd
}

func (m *MemoryRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(val)
	return cmd
}

func (m *MemoryRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, k := range keys {
		if _, ok := m.data[k]; ok {
			delete(m.data, k)
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

type OfferRepository interface {
	CreateOffer(ctx context.Context, offer *Offer) error
	GetOffer(ctx context.Context, id string) (*Offer, error)
	UpdateOffer(ctx context.Context, offer *Offer) error
	DeleteOffer(ctx context.Context, id string) error
	ToggleOffer(ctx context.Context, id string, isActive bool) error
	ListOffers(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Offer, error)
	ListActiveOffersForStore(ctx context.Context, chainID string, storeID string) ([]*Offer, error)
	ListStoresForChain(ctx context.Context, chainID string) ([]string, error)
	GetStoreChainID(ctx context.Context, storeID string) (string, error)
	ListStoresWithNoAudit(ctx context.Context) ([]string, error)
	LogOfferRulesAudit(ctx context.Context, storeID string, rulesetJSON []byte) error
	HasAuditRow(ctx context.Context, storeID string) (bool, error)

	CreateCoupon(ctx context.Context, coupon *Coupon) error
	GetCouponByID(ctx context.Context, id string) (*Coupon, error)
	GetCouponByCode(ctx context.Context, chainID, code string) (*Coupon, error)
	ListCoupons(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Coupon, error)
	ListActiveCoupons(ctx context.Context) ([]*Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *Coupon) error
	SoftDeleteCoupon(ctx context.Context, id string) error
	ToggleCoupon(ctx context.Context, id string, isActive bool) error
	GetUserCouponRedemptions(ctx context.Context, couponID, userID string) (int, error)
	RecordCouponRedemption(ctx context.Context, couponID, userID, checkoutSessionID string) error
	GetStoreIDsForChain(ctx context.Context, chainID string) ([]string, error)
}

// MemoryOfferRepository for unit tests
type MemoryOfferRepository struct {
	mu           sync.RWMutex
	offers       map[string]*Offer
	coupons      map[string]*Coupon
	redemptions  map[string][]*CouponRedemption
	storeChains  map[string]string
	auditHistory map[string][]*OfferRulesAuditRow
}

func NewMemoryOfferRepository() *MemoryOfferRepository {
	return &MemoryOfferRepository{
		offers:       make(map[string]*Offer),
		coupons:      make(map[string]*Coupon),
		redemptions:  make(map[string][]*CouponRedemption),
		storeChains:  make(map[string]string),
		auditHistory: make(map[string][]*OfferRulesAuditRow),
	}
}

func (m *MemoryOfferRepository) SetStoreChain(storeID, chainID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storeChains[storeID] = chainID
}

func (m *MemoryOfferRepository) CreateOffer(ctx context.Context, offer *Offer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offer.ID == "" {
		offer.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	offer.CreatedAt = now
	offer.UpdatedAt = now

	m.offers[offer.ID] = offer
	return nil
}

func (m *MemoryOfferRepository) GetOffer(ctx context.Context, id string) (*Offer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	off, ok := m.offers[id]
	if !ok {
		return nil, ErrOfferNotFound
	}
	return off, nil
}

func (m *MemoryOfferRepository) UpdateOffer(ctx context.Context, offer *Offer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.offers[offer.ID]
	if !ok {
		return ErrOfferNotFound
	}

	offer.CreatedAt = existing.CreatedAt
	offer.UpdatedAt = time.Now().UTC()
	m.offers[offer.ID] = offer
	return nil
}

func (m *MemoryOfferRepository) DeleteOffer(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	off, ok := m.offers[id]
	if !ok {
		return ErrOfferNotFound
	}
	off.IsActive = false
	off.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryOfferRepository) ToggleOffer(ctx context.Context, id string, isActive bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	off, ok := m.offers[id]
	if !ok {
		return ErrOfferNotFound
	}
	off.IsActive = isActive
	off.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryOfferRepository) ListOffers(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Offer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Offer
	for _, o := range m.offers {
		if o.ChainID != chainID {
			continue
		}
		if !includeInactive && !o.IsActive {
			continue
		}
		if storeID != nil && *storeID != "" {
			if o.StoreID != nil && *o.StoreID != *storeID {
				continue
			}
		}
		result = append(result, o)
	}
	return result, nil
}

func (m *MemoryOfferRepository) ListActiveOffersForStore(ctx context.Context, chainID string, storeID string) ([]*Offer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UTC()
	var result []*Offer

	for _, o := range m.offers {
		if !o.IsActive {
			continue
		}
		if o.ActiveFrom.After(now) {
			continue
		}
		if o.ActiveUntil != nil && o.ActiveUntil.Before(now) {
			continue
		}
		if o.ChainID != chainID {
			continue
		}

		// Store-specific OR Chain-wide
		if o.StoreID == nil || *o.StoreID == storeID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (m *MemoryOfferRepository) ListStoresForChain(ctx context.Context, chainID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stores []string
	for s, c := range m.storeChains {
		if c == chainID {
			stores = append(stores, s)
		}
	}
	return stores, nil
}

func (m *MemoryOfferRepository) GetStoreChainID(ctx context.Context, storeID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.storeChains[storeID]
	if !ok {
		return "chain-default-001", nil
	}
	return c, nil
}

func (m *MemoryOfferRepository) ListStoresWithNoAudit(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	for s := range m.storeChains {
		if len(m.auditHistory[s]) == 0 {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MemoryOfferRepository) LogOfferRulesAudit(ctx context.Context, storeID string, rulesetJSON []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	row := &OfferRulesAuditRow{
		ID:          uuid.New().String(),
		StoreID:     storeID,
		Ruleset:     json.RawMessage(rulesetJSON),
		ActivatedAt: time.Now().UTC(),
	}
	m.auditHistory[storeID] = append(m.auditHistory[storeID], row)
	return nil
}

func (m *MemoryOfferRepository) HasAuditRow(ctx context.Context, storeID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.auditHistory[storeID]) > 0, nil
}

// PostgresOfferRepository implementation for DB
type PostgresOfferRepository struct {
	db *sql.DB
}

func NewPostgresOfferRepository(db *sql.DB) *PostgresOfferRepository {
	return &PostgresOfferRepository{db: db}
}

func (p *PostgresOfferRepository) CreateOffer(ctx context.Context, offer *Offer) error {
	query := `
		INSERT INTO offers (
			id, chain_id, store_id, type, applies_to, target_ids, rule_config,
			min_cart_value_paise, max_discount_paise, priority, active_from, active_until,
			is_active, created_by, created_at, updated_at
		) VALUES (
			COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`
	targetJSON, _ := json.Marshal(offer.TargetIDs)
	configJSON, _ := json.Marshal(offer.RuleConfig)

	var storeIDVal *string
	if offer.StoreID != nil && *offer.StoreID != "" {
		storeIDVal = offer.StoreID
	}

	return p.db.QueryRowContext(
		ctx, query,
		offer.ID, offer.ChainID, storeIDVal, offer.Type, offer.AppliesTo,
		targetJSON, configJSON, offer.MinCartValuePaise, offer.MaxDiscountPaise,
		offer.Priority, offer.ActiveFrom, offer.ActiveUntil, offer.IsActive, offer.CreatedBy,
	).Scan(&offer.ID, &offer.CreatedAt, &offer.UpdatedAt)
}

func (p *PostgresOfferRepository) GetOffer(ctx context.Context, id string) (*Offer, error) {
	query := `
		SELECT id, chain_id, store_id, type, applies_to, target_ids, rule_config,
		       min_cart_value_paise, max_discount_paise, priority, active_from, active_until,
		       is_active, created_by, created_at, updated_at
		FROM offers WHERE id = $1
	`
	var o Offer
	var targetJSON, configJSON []byte
	var storeIDVal sql.NullString
	var maxDisc sql.NullInt64
	var activeUntil sql.NullTime

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.ChainID, &storeIDVal, &o.Type, &o.AppliesTo, &targetJSON, &configJSON,
		&o.MinCartValuePaise, &maxDisc, &o.Priority, &o.ActiveFrom, &activeUntil,
		&o.IsActive, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrOfferNotFound
	} else if err != nil {
		return nil, err
	}

	if storeIDVal.Valid {
		o.StoreID = &storeIDVal.String
	}
	if maxDisc.Valid {
		o.MaxDiscountPaise = &maxDisc.Int64
	}
	if activeUntil.Valid {
		o.ActiveUntil = &activeUntil.Time
	}
	_ = json.Unmarshal(targetJSON, &o.TargetIDs)
	_ = json.Unmarshal(configJSON, &o.RuleConfig)

	return &o, nil
}

func (p *PostgresOfferRepository) UpdateOffer(ctx context.Context, offer *Offer) error {
	query := `
		UPDATE offers SET
			type = $2, applies_to = $3, target_ids = $4, rule_config = $5,
			min_cart_value_paise = $6, max_discount_paise = $7, priority = $8,
			active_from = $9, active_until = $10, is_active = $11, updated_at = NOW()
		WHERE id = $1
	`
	targetJSON, _ := json.Marshal(offer.TargetIDs)
	configJSON, _ := json.Marshal(offer.RuleConfig)

	res, err := p.db.ExecContext(
		ctx, query,
		offer.ID, offer.Type, offer.AppliesTo, targetJSON, configJSON,
		offer.MinCartValuePaise, offer.MaxDiscountPaise, offer.Priority,
		offer.ActiveFrom, offer.ActiveUntil, offer.IsActive,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrOfferNotFound
	}
	return nil
}

func (p *PostgresOfferRepository) DeleteOffer(ctx context.Context, id string) error {
	return p.ToggleOffer(ctx, id, false)
}

func (p *PostgresOfferRepository) ToggleOffer(ctx context.Context, id string, isActive bool) error {
	query := `UPDATE offers SET is_active = $2, updated_at = NOW() WHERE id = $1`
	res, err := p.db.ExecContext(ctx, query, id, isActive)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrOfferNotFound
	}
	return nil
}

func (p *PostgresOfferRepository) ListOffers(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Offer, error) {
	query := `
		SELECT id, chain_id, store_id, type, applies_to, target_ids, rule_config,
		       min_cart_value_paise, max_discount_paise, priority, active_from, active_until,
		       is_active, created_by, created_at, updated_at
		FROM offers
		WHERE chain_id = $1
		  AND ($2::boolean = true OR is_active = true)
		  AND ($3::uuid IS NULL OR store_id = $3::uuid OR store_id IS NULL)
		ORDER BY priority ASC, created_at ASC
	`
	var storeIDParam *string
	if storeID != nil && *storeID != "" {
		storeIDParam = storeID
	}

	rows, err := p.db.QueryContext(ctx, query, chainID, includeInactive, storeIDParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Offer
	for rows.Next() {
		var o Offer
		var targetJSON, configJSON []byte
		var storeIDVal sql.NullString
		var maxDisc sql.NullInt64
		var activeUntil sql.NullTime

		if err := rows.Scan(
			&o.ID, &o.ChainID, &storeIDVal, &o.Type, &o.AppliesTo, &targetJSON, &configJSON,
			&o.MinCartValuePaise, &maxDisc, &o.Priority, &o.ActiveFrom, &activeUntil,
			&o.IsActive, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if storeIDVal.Valid {
			o.StoreID = &storeIDVal.String
		}
		if maxDisc.Valid {
			o.MaxDiscountPaise = &maxDisc.Int64
		}
		if activeUntil.Valid {
			o.ActiveUntil = &activeUntil.Time
		}
		_ = json.Unmarshal(targetJSON, &o.TargetIDs)
		_ = json.Unmarshal(configJSON, &o.RuleConfig)

		result = append(result, &o)
	}
	return result, nil
}

func (p *PostgresOfferRepository) ListActiveOffersForStore(ctx context.Context, chainID string, storeID string) ([]*Offer, error) {
	query := `
		SELECT id, chain_id, store_id, type, applies_to, target_ids, rule_config,
		       min_cart_value_paise, max_discount_paise, priority, active_from, active_until,
		       is_active, created_by, created_at, updated_at
		FROM offers
		WHERE is_active = true
		  AND active_from <= NOW()
		  AND (active_until IS NULL OR active_until > NOW())
		  AND ((store_id = $1::uuid) OR (store_id IS NULL AND chain_id = $2::uuid))
		ORDER BY priority ASC, created_at ASC
	`
	rows, err := p.db.QueryContext(ctx, query, storeID, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Offer
	for rows.Next() {
		var o Offer
		var targetJSON, configJSON []byte
		var storeIDVal sql.NullString
		var maxDisc sql.NullInt64
		var activeUntil sql.NullTime

		if err := rows.Scan(
			&o.ID, &o.ChainID, &storeIDVal, &o.Type, &o.AppliesTo, &targetJSON, &configJSON,
			&o.MinCartValuePaise, &maxDisc, &o.Priority, &o.ActiveFrom, &activeUntil,
			&o.IsActive, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if storeIDVal.Valid {
			o.StoreID = &storeIDVal.String
		}
		if maxDisc.Valid {
			o.MaxDiscountPaise = &maxDisc.Int64
		}
		if activeUntil.Valid {
			o.ActiveUntil = &activeUntil.Time
		}
		_ = json.Unmarshal(targetJSON, &o.TargetIDs)
		_ = json.Unmarshal(configJSON, &o.RuleConfig)

		result = append(result, &o)
	}
	return result, nil
}

func (p *PostgresOfferRepository) ListStoresForChain(ctx context.Context, chainID string) ([]string, error) {
	query := `SELECT id FROM stores WHERE chain_id = $1`
	rows, err := p.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			stores = append(stores, id)
		}
	}
	return stores, nil
}

func (p *PostgresOfferRepository) GetStoreChainID(ctx context.Context, storeID string) (string, error) {
	query := `SELECT chain_id FROM stores WHERE id = $1`
	var chainID string
	err := p.db.QueryRowContext(ctx, query, storeID).Scan(&chainID)
	if err != nil {
		return "chain-default-001", nil
	}
	return chainID, nil
}

func (p *PostgresOfferRepository) ListStoresWithNoAudit(ctx context.Context) ([]string, error) {
	query := `
		SELECT s.id FROM stores s
		LEFT JOIN offer_rules_audit a ON s.id = a.store_id
		WHERE a.id IS NULL
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			stores = append(stores, id)
		}
	}
	return stores, nil
}

func (p *PostgresOfferRepository) LogOfferRulesAudit(ctx context.Context, storeID string, rulesetJSON []byte) error {
	query := `INSERT INTO offer_rules_audit (store_id, ruleset, activated_at) VALUES ($1, $2, NOW())`
	_, err := p.db.ExecContext(ctx, query, storeID, rulesetJSON)
	return err
}

func (p *PostgresOfferRepository) HasAuditRow(ctx context.Context, storeID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM offer_rules_audit WHERE store_id = $1)`
	var exists bool
	err := p.db.QueryRowContext(ctx, query, storeID).Scan(&exists)
	return exists, err
}

func (m *MemoryOfferRepository) CreateCoupon(ctx context.Context, c *Coupon) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	m.coupons[c.ID] = c
	return nil
}

func (m *MemoryOfferRepository) GetCouponByID(ctx context.Context, id string) (*Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.coupons[id]
	if !ok {
		return nil, errors.New("coupon not found")
	}
	return c, nil
}

func (m *MemoryOfferRepository) GetCouponByCode(ctx context.Context, chainID, code string) (*Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.coupons {
		if c.ChainID == chainID && c.Code == code {
			return c, nil
		}
	}
	return nil, errors.New("coupon not found")
}

func (m *MemoryOfferRepository) ListCoupons(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*Coupon
	for _, c := range m.coupons {
		if c.ChainID != chainID {
			continue
		}
		if !includeInactive && !c.IsActive {
			continue
		}
		if storeID != nil && *storeID != "" && c.StoreID != nil && *c.StoreID != *storeID {
			continue
		}
		res = append(res, c)
	}
	return res, nil
}

func (m *MemoryOfferRepository) ListActiveCoupons(ctx context.Context) ([]*Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*Coupon
	for _, c := range m.coupons {
		if c.IsActive {
			res = append(res, c)
		}
	}
	return res, nil
}

func (m *MemoryOfferRepository) UpdateCoupon(ctx context.Context, coupon *Coupon) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	coupon.UpdatedAt = time.Now().UTC()
	m.coupons[coupon.ID] = coupon
	return nil
}

func (m *MemoryOfferRepository) SoftDeleteCoupon(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.coupons[id]; ok {
		c.IsActive = false
		c.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (m *MemoryOfferRepository) ToggleCoupon(ctx context.Context, id string, isActive bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.coupons[id]; ok {
		c.IsActive = isActive
		c.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (m *MemoryOfferRepository) GetUserCouponRedemptions(ctx context.Context, couponID, userID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reds, ok := m.redemptions[couponID]
	if !ok {
		return 0, nil
	}

	count := 0
	for _, r := range reds {
		if r.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MemoryOfferRepository) RecordCouponRedemption(ctx context.Context, couponID, userID, checkoutSessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.coupons[couponID]
	if !ok {
		return errors.New("coupon not found")
	}

	// Idempotency check: UNIQUE(coupon_id, checkout_session_id)
	for _, r := range m.redemptions[couponID] {
		if r.CheckoutSessionID == checkoutSessionID {
			return nil // Already recorded, idempotent
		}
	}

	r := &CouponRedemption{
		ID:                uuid.New().String(),
		CouponID:          couponID,
		UserID:            userID,
		CheckoutSessionID: checkoutSessionID,
		RedeemedAt:        time.Now().UTC(),
	}
	m.redemptions[couponID] = append(m.redemptions[couponID], r)
	c.CurrentUseCount++
	return nil
}


func (m *MemoryOfferRepository) GetStoreIDsForChain(ctx context.Context, chainID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stores []string
	for sID, cID := range m.storeChains {
		if cID == chainID {
			stores = append(stores, sID)
		}
	}
	if len(stores) == 0 {
		return []string{"store-001", "store-002"}, nil
	}
	return stores, nil
}

func (p *PostgresOfferRepository) CreateCoupon(ctx context.Context, c *Coupon) error {
	query := `INSERT INTO coupons (id, chain_id, store_id, code, discount_type, discount_value, min_cart_value_paise, max_uses, max_uses_per_customer, current_use_count, active_from, active_until, is_active, created_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())`
	_, err := p.db.ExecContext(ctx, query, c.ID, c.ChainID, c.StoreID, c.Code, c.DiscountType, c.DiscountValue, c.MinCartValuePaise, c.MaxUses, c.MaxUsesPerCustomer, c.CurrentUseCount, c.ActiveFrom, c.ActiveUntil, c.IsActive, c.CreatedBy)
	return err
}

func (p *PostgresOfferRepository) GetCouponByID(ctx context.Context, id string) (*Coupon, error) {
	query := `SELECT id, chain_id, store_id, code, discount_type, discount_value, min_cart_value_paise, max_uses, max_uses_per_customer, current_use_count, active_from, active_until, is_active, created_by, created_at, updated_at
			  FROM coupons WHERE id = $1`
	var c Coupon
	err := p.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.ChainID, &c.StoreID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinCartValuePaise, &c.MaxUses, &c.MaxUsesPerCustomer, &c.CurrentUseCount, &c.ActiveFrom, &c.ActiveUntil, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *PostgresOfferRepository) GetCouponByCode(ctx context.Context, chainID, code string) (*Coupon, error) {
	query := `SELECT id, chain_id, store_id, code, discount_type, discount_value, min_cart_value_paise, max_uses, max_uses_per_customer, current_use_count, active_from, active_until, is_active, created_by, created_at, updated_at
			  FROM coupons WHERE chain_id = $1 AND code = $2`
	var c Coupon
	err := p.db.QueryRowContext(ctx, query, chainID, code).Scan(&c.ID, &c.ChainID, &c.StoreID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinCartValuePaise, &c.MaxUses, &c.MaxUsesPerCustomer, &c.CurrentUseCount, &c.ActiveFrom, &c.ActiveUntil, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *PostgresOfferRepository) ListCoupons(ctx context.Context, chainID string, storeID *string, includeInactive bool) ([]*Coupon, error) {
	query := `SELECT id, chain_id, store_id, code, discount_type, discount_value, min_cart_value_paise, max_uses, max_uses_per_customer, current_use_count, active_from, active_until, is_active, created_by, created_at, updated_at
			  FROM coupons WHERE chain_id = $1`
	if !includeInactive {
		query += ` AND is_active = true`
	}
	if storeID != nil && *storeID != "" {
		query += fmt.Sprintf(" AND (store_id IS NULL OR store_id = '%s')", *storeID)
	}

	rows, err := p.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*Coupon
	for rows.Next() {
		var c Coupon
		if err := rows.Scan(&c.ID, &c.ChainID, &c.StoreID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinCartValuePaise, &c.MaxUses, &c.MaxUsesPerCustomer, &c.CurrentUseCount, &c.ActiveFrom, &c.ActiveUntil, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err == nil {
			res = append(res, &c)
		}
	}
	return res, nil
}

func (p *PostgresOfferRepository) ListActiveCoupons(ctx context.Context) ([]*Coupon, error) {
	query := `SELECT id, chain_id, store_id, code, discount_type, discount_value, min_cart_value_paise, max_uses, max_uses_per_customer, current_use_count, active_from, active_until, is_active, created_by, created_at, updated_at
			  FROM coupons WHERE is_active = true`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*Coupon
	for rows.Next() {
		var c Coupon
		if err := rows.Scan(&c.ID, &c.ChainID, &c.StoreID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinCartValuePaise, &c.MaxUses, &c.MaxUsesPerCustomer, &c.CurrentUseCount, &c.ActiveFrom, &c.ActiveUntil, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err == nil {
			res = append(res, &c)
		}
	}
	return res, nil
}

func (p *PostgresOfferRepository) UpdateCoupon(ctx context.Context, coupon *Coupon) error {
	query := `UPDATE coupons SET discount_type = $1, discount_value = $2, min_cart_value_paise = $3, max_uses = $4, max_uses_per_customer = $5, active_from = $6, active_until = $7, updated_at = NOW() WHERE id = $8`
	_, err := p.db.ExecContext(ctx, query, coupon.DiscountType, coupon.DiscountValue, coupon.MinCartValuePaise, coupon.MaxUses, coupon.MaxUsesPerCustomer, coupon.ActiveFrom, coupon.ActiveUntil, coupon.ID)
	return err
}

func (p *PostgresOfferRepository) SoftDeleteCoupon(ctx context.Context, id string) error {
	query := `UPDATE coupons SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

func (p *PostgresOfferRepository) ToggleCoupon(ctx context.Context, id string, isActive bool) error {
	query := `UPDATE coupons SET is_active = $1, updated_at = NOW() WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, isActive, id)
	return err
}

func (p *PostgresOfferRepository) GetUserCouponRedemptions(ctx context.Context, couponID, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`
	var count int
	err := p.db.QueryRowContext(ctx, query, couponID, userID).Scan(&count)
	return count, err
}

func (p *PostgresOfferRepository) RecordCouponRedemption(ctx context.Context, couponID, userID, checkoutSessionID string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// ON CONFLICT (coupon_id, checkout_session_id) DO NOTHING
	redemptionID := uuid.New().String()
	res, err := tx.ExecContext(ctx, `INSERT INTO coupon_redemptions (id, coupon_id, user_id, checkout_session_id, redeemed_at)
									 VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (coupon_id, checkout_session_id) DO NOTHING`, redemptionID, couponID, userID, checkoutSessionID)
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected > 0 {
		// Increment current_use_count on newly recorded redemption
		_, err = tx.ExecContext(ctx, `UPDATE coupons SET current_use_count = current_use_count + 1 WHERE id = $1`, couponID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *PostgresOfferRepository) GetStoreIDsForChain(ctx context.Context, chainID string) ([]string, error) {
	query := `SELECT id FROM stores WHERE chain_id = $1`
	rows, err := p.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			stores = append(stores, id)
		}
	}
	if len(stores) == 0 {
		return []string{"store-001", "store-002"}, nil
	}
	return stores, nil
}
