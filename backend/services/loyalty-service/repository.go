package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/loyalty"
)

type Repository interface {
	GetAccountByUserID(ctx context.Context, userID string) (*LoyaltyAccount, error)
	EnsureAccountExists(ctx context.Context, userID string) (*LoyaltyAccount, error)
	ReservePointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, int64, error)
	CommitPointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, error)
	ReleasePointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, int64, error)
	EarnPointsTx(ctx context.Context, orderID, userID string, totalPaise int64) (earned int64, oldTier, newTier string, upgraded bool, newBalance int64, err error)
	ReversePointsTx(ctx context.Context, orderID, userID, returnID string, returnedAmountPaise, originalTotalPaise int64) (reversed int64, newBalance int64, err error)
	GetLedgerHistory(ctx context.Context, userID string, page, pageSize int) ([]LedgerItemResponse, error)
	GetTierConfigs(ctx context.Context) ([]LoyaltyTierConfig, error)
	DeleteUserLoyaltyData(ctx context.Context, userID string) (int, error)
	GetAccountByReferralCode(ctx context.Context, code string) (*LoyaltyAccount, error)
	ApplyReferral(ctx context.Context, referredUserID, referralCode string) error
	ProcessFirstOrderReferralReward(ctx context.Context, userID, orderID string) (referrerID string, referrerPoints int64, referredPoints int64, rewarded bool, err error)
	ExpirePendingReferrals(ctx context.Context) (int64, error)
}

type PostgresRepository struct {
	db       *sql.DB
	isSQLite bool
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	isSQLite := false
	if db != nil && fmt.Sprintf("%T", db.Driver()) == "*sqlite3.SQLiteDriver" {
		isSQLite = true
	}
	return &PostgresRepository{db: db, isSQLite: isSQLite}
}

func (r *PostgresRepository) lockClause() string {
	if r.isSQLite {
		return ""
	}
	return " FOR UPDATE"
}

var codeChars = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

func generateReferralCode() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = codeChars[time.Now().UnixNano()%int64(len(codeChars))]
	}
	return string(b)
}

func (r *PostgresRepository) GetAccountByUserID(ctx context.Context, userID string) (*LoyaltyAccount, error) {
	query := `
		SELECT user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, tier_updated_at, created_at, updated_at
		FROM loyalty_accounts
		WHERE user_id = $1
	`
	if r.isSQLite {
		query = `
			SELECT user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, tier_updated_at, created_at, updated_at
			FROM loyalty_accounts
			WHERE user_id = ?
		`
	}
	row := r.db.QueryRowContext(ctx, query, userID)

	var acc LoyaltyAccount
	var refCode sql.NullString
	err := row.Scan(
		&acc.UserID, &acc.PointsBalance, &acc.PointsReserved, &acc.LifetimePointsEarned,
		&acc.Tier, &refCode, &acc.TierUpdatedAt, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan loyalty account: %w", err)
	}

	if refCode.Valid && refCode.String != "" {
		acc.ReferralCode = refCode.String
	} else {
		code := generateReferralCode()
		updateQ := `UPDATE loyalty_accounts SET referral_code = $1 WHERE user_id = $2 AND (referral_code IS NULL OR referral_code = '')`
		if r.isSQLite {
			updateQ = `UPDATE loyalty_accounts SET referral_code = ? WHERE user_id = ? AND (referral_code IS NULL OR referral_code = '')`
		}
		_, _ = r.db.ExecContext(ctx, updateQ, code, userID)
		acc.ReferralCode = code
	}

	return &acc, nil
}

func (r *PostgresRepository) EnsureAccountExists(ctx context.Context, userID string) (*LoyaltyAccount, error) {
	acc, err := r.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if acc != nil {
		return acc, nil
	}

	// Create zero-balance account
	now := time.Now()
	code := generateReferralCode()
	insertQuery := `
		INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, created_at, updated_at)
		VALUES ($1, 0, 0, 0, 'BRONZE', $2, $3, $4)
		ON CONFLICT (user_id) DO NOTHING
	`
	if r.isSQLite {
		insertQuery = `
			INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, created_at, updated_at)
			VALUES (?, 0, 0, 0, 'BRONZE', ?, ?, ?)
			ON CONFLICT (user_id) DO NOTHING
		`
	}
	_, err = r.db.ExecContext(ctx, insertQuery, userID, code, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create loyalty account: %w", err)
	}

	return r.GetAccountByUserID(ctx, userID)
}

func (r *PostgresRepository) ReservePointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check idempotency first: query existing ledger entry
	var existingBalanceAfter int64
	err = tx.QueryRowContext(ctx, `SELECT balance_after FROM loyalty_ledger WHERE idempotency_key = $1`, idempotencyKey).Scan(&existingBalanceAfter)
	if err == nil {
		// Idempotent duplicate call: return existing balance after
		return true, existingBalanceAfter, nil
	}

	// Ensure account exists & lock row
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier, created_at, updated_at)
		VALUES ($1, 0, 0, 0, 'BRONZE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to ensure account: %w", err)
	}

	var balance, reserved int64
	queryReserve := `SELECT points_balance, points_reserved FROM loyalty_accounts WHERE user_id = $1` + r.lockClause()
	err = tx.QueryRowContext(ctx, queryReserve, userID).Scan(&balance, &reserved)
	if err != nil {
		return false, 0, fmt.Errorf("failed to lock loyalty account: %w", err)
	}

	if balance < points {
		return false, balance, fmt.Errorf("INSUFFICIENT_LOYALTY_POINTS")
	}

	newBalance := balance - points
	newReserved := reserved + points

	// Update account
	_, err = tx.ExecContext(ctx, `
		UPDATE loyalty_accounts
		SET points_balance = $1, points_reserved = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $3
	`, newBalance, newReserved, userID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to update points balance: %w", err)
	}

	// Insert ledger row
	refType := "PAYMENT"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`, uuid.New().String(), userID, EntryRedeemReserve, -points, refType, idempotencyKey, newBalance)
	if err != nil {
		return false, 0, fmt.Errorf("failed to insert reserve ledger row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, 0, fmt.Errorf("failed to commit reserve tx: %w", err)
	}

	return true, newBalance, nil
}

func (r *PostgresRepository) CommitPointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check idempotency
	var count int
	_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM loyalty_ledger WHERE idempotency_key = $1`, idempotencyKey).Scan(&count)
	if count > 0 {
		return true, nil
	}

	var reserved int64
	queryCommit := `SELECT points_reserved FROM loyalty_accounts WHERE user_id = $1` + r.lockClause()
	err = tx.QueryRowContext(ctx, queryCommit, userID).Scan(&reserved)
	if err != nil {
		return false, fmt.Errorf("failed to lock loyalty account: %w", err)
	}

	newReserved := reserved - points
	if newReserved < 0 {
		newReserved = 0
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE loyalty_accounts
		SET points_reserved = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`, newReserved, userID)
	if err != nil {
		return false, fmt.Errorf("failed to update points_reserved: %w", err)
	}

	refType := "PAYMENT"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, $3, 0, $4, $5, (SELECT points_balance FROM loyalty_accounts WHERE user_id = $2), CURRENT_TIMESTAMP)
	`, uuid.New().String(), userID, EntryRedeemCommit, refType, idempotencyKey)
	if err != nil {
		return false, fmt.Errorf("failed to insert commit ledger row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit commit tx: %w", err)
	}

	return true, nil
}

func (r *PostgresRepository) ReleasePointsTx(ctx context.Context, userID string, points int64, idempotencyKey string) (bool, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check idempotency
	var existingBalance int64
	err = tx.QueryRowContext(ctx, `SELECT balance_after FROM loyalty_ledger WHERE idempotency_key = $1`, idempotencyKey).Scan(&existingBalance)
	if err == nil {
		return true, existingBalance, nil
	}

	var balance, reserved int64
	queryRelease := `SELECT points_balance, points_reserved FROM loyalty_accounts WHERE user_id = $1` + r.lockClause()
	err = tx.QueryRowContext(ctx, queryRelease, userID).Scan(&balance, &reserved)
	if err != nil {
		return false, 0, fmt.Errorf("failed to lock loyalty account: %w", err)
	}

	newBalance := balance + points
	newReserved := reserved - points
	if newReserved < 0 {
		newReserved = 0
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE loyalty_accounts
		SET points_balance = $1, points_reserved = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $3
	`, newBalance, newReserved, userID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to release points: %w", err)
	}

	refType := "PAYMENT"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`, uuid.New().String(), userID, EntryRedeemRelease, points, refType, idempotencyKey, newBalance)
	if err != nil {
		return false, 0, fmt.Errorf("failed to insert release ledger row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, 0, fmt.Errorf("failed to commit release tx: %w", err)
	}

	return true, newBalance, nil
}

func (r *PostgresRepository) EarnPointsTx(ctx context.Context, orderID, userID string, totalPaise int64) (int64, string, string, bool, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	idempotencyKey := fmt.Sprintf("earn:%s", orderID)

	// Check idempotency
	var count int
	_ = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM loyalty_ledger WHERE idempotency_key = $1`, idempotencyKey).Scan(&count)
	if count > 0 {
		// Idempotent duplicate: fetch current balance
		var b int64
		var t string
		_ = tx.QueryRowContext(ctx, `SELECT points_balance, tier FROM loyalty_accounts WHERE user_id = $1`, userID).Scan(&b, &t)
		return 0, t, t, false, b, nil
	}

	// Ensure account exists
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_accounts (user_id, points_balance, points_reserved, lifetime_points_earned, tier, created_at, updated_at)
		VALUES ($1, 0, 0, 0, 'BRONZE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	if err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to ensure account: %w", err)
	}

	var balance, lifetime int64
	var oldTier string
	queryEarn := `
		SELECT points_balance, lifetime_points_earned, tier
		FROM loyalty_accounts
		WHERE user_id = $1` + r.lockClause()
	err = tx.QueryRowContext(ctx, queryEarn, userID).Scan(&balance, &lifetime, &oldTier)
	if err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to lock account: %w", err)
	}

	tierConfigs, err := r.GetTierConfigs(ctx)
	if err != nil {
		tierConfigs = nil
	}

	// Map shared TierConfig
	sharedTiers := make([]loyalty.TierConfig, len(tierConfigs))
	for i, tc := range tierConfigs {
		sharedTiers[i] = loyalty.TierConfig{
			Tier:              tc.Tier,
			MinLifetimePoints: tc.MinLifetimePoints,
			EarnMultiplier:    tc.EarnMultiplier,
			DisplayName:       tc.DisplayName,
			DisplayOrder:      tc.DisplayOrder,
		}
	}

	// Get current tier multiplier at order time
	currentTierConfig := loyalty.CalculateTier(lifetime, sharedTiers)
	multiplier := currentTierConfig.EarnMultiplier + r.getSubscriptionBonus(ctx, userID)
	earnedPoints := loyalty.CalculateEarnPoints(totalPaise, multiplier)

	newBalance := balance + earnedPoints
	newLifetime := lifetime + earnedPoints

	// Recompute new tier
	newTierConfig := loyalty.CalculateTier(newLifetime, sharedTiers)
	newTier := newTierConfig.Tier
	isUpgraded := newTier != oldTier

	now := time.Now()
	if isUpgraded {
		_, err = tx.ExecContext(ctx, `
			UPDATE loyalty_accounts
			SET points_balance = $1, lifetime_points_earned = $2, tier = $3, tier_updated_at = $4, updated_at = $5
			WHERE user_id = $6
		`, newBalance, newLifetime, newTier, now, now, userID)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE loyalty_accounts
			SET points_balance = $1, lifetime_points_earned = $2, updated_at = $3
			WHERE user_id = $4
		`, newBalance, newLifetime, now, userID)
	}
	if err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to update loyalty account points: %w", err)
	}

	// Insert EARN ledger row
	refType := "ORDER"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, uuid.New().String(), userID, EntryEarn, earnedPoints, refType, orderID, idempotencyKey, newBalance, now)
	if err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to insert earn ledger row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", "", false, 0, fmt.Errorf("failed to commit earn tx: %w", err)
	}

	return earnedPoints, oldTier, newTier, isUpgraded, newBalance, nil
}

func (r *PostgresRepository) ReversePointsTx(ctx context.Context, orderID, userID, returnID string, returnedAmountPaise, originalTotalPaise int64) (int64, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	keyRef := returnID
	if keyRef == "" {
		keyRef = fmt.Sprintf("%s:%d", orderID, returnedAmountPaise)
	}
	idempotencyKey := fmt.Sprintf("reversal:%s", keyRef)

	// Check idempotency
	var existingBalance int64
	err = tx.QueryRowContext(ctx, `SELECT balance_after FROM loyalty_ledger WHERE idempotency_key = $1`, idempotencyKey).Scan(&existingBalance)
	if err == nil {
		return 0, existingBalance, nil
	}

	// Query original EARN ledger entry for originalEarnedPoints
	var originalEarnedPoints int64
	err = tx.QueryRowContext(ctx, `
		SELECT points_delta FROM loyalty_ledger
		WHERE reference_id = $1 AND entry_type = 'EARN'
		ORDER BY created_at ASC LIMIT 1
	`, orderID).Scan(&originalEarnedPoints)
	if err != nil || originalEarnedPoints <= 0 {
		originalEarnedPoints = (originalTotalPaise / 1000)
	}

	var reversalPoints int64
	if originalTotalPaise > 0 {
		reversalPoints = (originalEarnedPoints * returnedAmountPaise) / originalTotalPaise
	}

	var balance int64
	queryReverse := `SELECT points_balance FROM loyalty_accounts WHERE user_id = $1` + r.lockClause()
	err = tx.QueryRowContext(ctx, queryReverse, userID).Scan(&balance)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to lock account: %w", err)
	}

	// Floor at 0 (never let reversal push balance negative)
	actualDeducted := reversalPoints
	if actualDeducted > balance {
		actualDeducted = balance
	}

	newBalance := balance - actualDeducted

	// Update balance ONLY (lifetime_points_earned and tier remain unchanged!)
	_, err = tx.ExecContext(ctx, `
		UPDATE loyalty_accounts
		SET points_balance = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`, newBalance, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update points balance on reversal: %w", err)
	}

	// Insert REVERSAL ledger row
	refType := "RETURN"
	refID := orderID
	if returnID != "" {
		refID = returnID
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	`, uuid.New().String(), userID, EntryReversal, -actualDeducted, refType, refID, idempotencyKey, newBalance)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert reversal ledger row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("failed to commit reversal tx: %w", err)
	}

	return actualDeducted, newBalance, nil
}

func (r *PostgresRepository) GetLedgerHistory(ctx context.Context, userID string, page, pageSize int) ([]LedgerItemResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT entry_type, points_delta, reference_type, created_at, balance_after
		FROM loyalty_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query ledger history: %w", err)
	}
	defer rows.Close()

	var items []LedgerItemResponse
	for rows.Next() {
		var item LedgerItemResponse
		if err := rows.Scan(&item.EntryType, &item.PointsDelta, &item.ReferenceType, &item.CreatedAt, &item.BalanceAfter); err != nil {
			return nil, fmt.Errorf("failed to scan ledger item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *PostgresRepository) GetTierConfigs(ctx context.Context) ([]LoyaltyTierConfig, error) {
	query := `
		SELECT tier, min_lifetime_points, earn_multiplier, display_name, display_order
		FROM loyalty_tier_config
		ORDER BY display_order ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier configs: %w", err)
	}
	defer rows.Close()

	var tiers []LoyaltyTierConfig
	for rows.Next() {
		var t LoyaltyTierConfig
		if err := rows.Scan(&t.Tier, &t.MinLifetimePoints, &t.EarnMultiplier, &t.DisplayName, &t.DisplayOrder); err != nil {
			return nil, fmt.Errorf("failed to scan tier config: %w", err)
		}
		tiers = append(tiers, t)
	}

	if len(tiers) == 0 {
		// Fallback defaults
		for _, dt := range loyalty.DefaultTiers {
			tiers = append(tiers, LoyaltyTierConfig{
				Tier:              dt.Tier,
				MinLifetimePoints: dt.MinLifetimePoints,
				EarnMultiplier:    dt.EarnMultiplier,
				DisplayName:       dt.DisplayName,
				DisplayOrder:      dt.DisplayOrder,
			})
		}
	}

	return tiers, nil
}

func (r *PostgresRepository) DeleteUserLoyaltyData(ctx context.Context, userID string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	q1 := "DELETE FROM loyalty_ledger WHERE user_id = $1"
	q2 := "DELETE FROM loyalty_accounts WHERE user_id = $1"
	if r.isSQLite {
		q1 = "DELETE FROM loyalty_ledger WHERE user_id = ?"
		q2 = "DELETE FROM loyalty_accounts WHERE user_id = ?"
	}

	var rows1, rows2 int64
	if res1, err1 := tx.ExecContext(ctx, q1, userID); err1 == nil && res1 != nil {
		rows1, _ = res1.RowsAffected()
	}
	if res2, err2 := tx.ExecContext(ctx, q2, userID); err2 == nil && res2 != nil {
		rows2, _ = res2.RowsAffected()
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(rows1 + rows2), nil
}

func (r *PostgresRepository) GetAccountByReferralCode(ctx context.Context, code string) (*LoyaltyAccount, error) {
	query := `
		SELECT user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, tier_updated_at, created_at, updated_at
		FROM loyalty_accounts
		WHERE referral_code = $1
	`
	if r.isSQLite {
		query = `
			SELECT user_id, points_balance, points_reserved, lifetime_points_earned, tier, referral_code, tier_updated_at, created_at, updated_at
			FROM loyalty_accounts
			WHERE referral_code = ?
		`
	}
	row := r.db.QueryRowContext(ctx, query, code)

	var acc LoyaltyAccount
	var refCode sql.NullString
	err := row.Scan(
		&acc.UserID, &acc.PointsBalance, &acc.PointsReserved, &acc.LifetimePointsEarned,
		&acc.Tier, &refCode, &acc.TierUpdatedAt, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan account by referral code: %w", err)
	}
	if refCode.Valid {
		acc.ReferralCode = refCode.String
	}
	return &acc, nil
}

func (r *PostgresRepository) ApplyReferral(ctx context.Context, referredUserID, referralCode string) error {
	referrerAcc, err := r.GetAccountByReferralCode(ctx, referralCode)
	if err != nil {
		return err
	}
	if referrerAcc == nil {
		return fmt.Errorf("INVALID_REFERRAL_CODE")
	}
	if referrerAcc.UserID == referredUserID {
		return fmt.Errorf("CANNOT_REFER_SELF")
	}

	// Check if referred_user_id already has a referral_events row
	var existingCount int
	checkQuery := `SELECT COUNT(*) FROM referral_events WHERE referred_user_id = $1`
	if r.isSQLite {
		checkQuery = `SELECT COUNT(*) FROM referral_events WHERE referred_user_id = ?`
	}
	_ = r.db.QueryRowContext(ctx, checkQuery, referredUserID).Scan(&existingCount)
	if existingCount > 0 {
		return fmt.Errorf("REFERRAL_ALREADY_APPLIED")
	}

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)
	id := uuid.New().String()

	insertQuery := `
		INSERT INTO referral_events (id, referrer_user_id, referred_user_id, referral_code, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, 'PENDING', $5, $6)
	`
	if r.isSQLite {
		insertQuery = `
			INSERT INTO referral_events (id, referrer_user_id, referred_user_id, referral_code, status, created_at, expires_at)
			VALUES (?, ?, ?, ?, 'PENDING', ?, ?)
		`
	}

	_, err = r.db.ExecContext(ctx, insertQuery, id, referrerAcc.UserID, referredUserID, referralCode, now, expiresAt)
	if err != nil {
		return fmt.Errorf("REFERRAL_ALREADY_APPLIED")
	}

	return nil
}

func (r *PostgresRepository) ProcessFirstOrderReferralReward(ctx context.Context, userID, orderID string) (referrerID string, referrerPoints int64, referredPoints int64, rewarded bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", 0, 0, false, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, referrer_user_id, referral_code, expires_at
		FROM referral_events
		WHERE referred_user_id = $1 AND status = 'PENDING'
	` + r.lockClause()
	if r.isSQLite {
		query = `
			SELECT id, referrer_user_id, referral_code, expires_at
			FROM referral_events
			WHERE referred_user_id = ? AND status = 'PENDING'
		`
	}

	var eventID, refUserID, refCode string
	var expiresAtVal interface{}
	err = tx.QueryRowContext(ctx, query, userID).Scan(&eventID, &refUserID, &refCode, &expiresAtVal)
	if err == sql.ErrNoRows {
		return "", 0, 0, false, nil
	}
	if err != nil {
		return "", 0, 0, false, err
	}

	var expiresAt time.Time
	switch v := expiresAtVal.(type) {
	case time.Time:
		expiresAt = v
	case string:
		expiresAt, _ = time.Parse("2006-01-02 15:04:05", v)
		if expiresAt.IsZero() {
			expiresAt, _ = time.Parse(time.RFC3339, v)
		}
	}

	now := time.Now()
	if now.After(expiresAt) {
		expQuery := `UPDATE referral_events SET status = 'EXPIRED' WHERE id = $1`
		if r.isSQLite {
			expQuery = `UPDATE referral_events SET status = 'EXPIRED' WHERE id = ?`
		}
		_, _ = tx.ExecContext(ctx, expQuery, eventID)
		_ = tx.Commit()
		return "", 0, 0, false, nil
	}

	// 1. Award 100 points to referrer
	referrerPoints = 100
	referredPoints = 50

	refAccountQ := `UPDATE loyalty_accounts SET points_balance = points_balance + 100, lifetime_points_earned = lifetime_points_earned + 100, updated_at = $1 WHERE user_id = $2 RETURNING points_balance`
	if r.isSQLite {
		refAccountQ = `UPDATE loyalty_accounts SET points_balance = points_balance + 100, lifetime_points_earned = lifetime_points_earned + 100, updated_at = ? WHERE user_id = ?`
	}
	var refNewBal int64
	if r.isSQLite {
		_, _ = tx.ExecContext(ctx, refAccountQ, now, refUserID)
		_ = tx.QueryRowContext(ctx, "SELECT points_balance FROM loyalty_accounts WHERE user_id = ?", refUserID).Scan(&refNewBal)
	} else {
		_ = tx.QueryRowContext(ctx, refAccountQ, now, refUserID).Scan(&refNewBal)
	}

	refLedgerQ := `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, 'EARN', 100, 'REFERRAL_BONUS_REFERRER', $3, $4, $5, $6)
	`
	if r.isSQLite {
		refLedgerQ = `
			INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
			VALUES (?, ?, 'EARN', 100, 'REFERRAL_BONUS_REFERRER', ?, ?, ?, ?)
		`
	}
	refIdempotency := "referral:referrer:" + eventID
	_, _ = tx.ExecContext(ctx, refLedgerQ, uuid.New().String(), refUserID, orderID, refIdempotency, refNewBal, now)

	// 2. Award 50 points to referred user
	refedAccountQ := `UPDATE loyalty_accounts SET points_balance = points_balance + 50, lifetime_points_earned = lifetime_points_earned + 50, updated_at = $1 WHERE user_id = $2 RETURNING points_balance`
	if r.isSQLite {
		refedAccountQ = `UPDATE loyalty_accounts SET points_balance = points_balance + 50, lifetime_points_earned = lifetime_points_earned + 50, updated_at = ? WHERE user_id = ?`
	}
	var refedNewBal int64
	if r.isSQLite {
		_, _ = tx.ExecContext(ctx, refedAccountQ, now, userID)
		_ = tx.QueryRowContext(ctx, "SELECT points_balance FROM loyalty_accounts WHERE user_id = ?", userID).Scan(&refedNewBal)
	} else {
		_ = tx.QueryRowContext(ctx, refedAccountQ, now, userID).Scan(&refedNewBal)
	}

	refedLedgerQ := `
		INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
		VALUES ($1, $2, 'EARN', 50, 'REFERRAL_BONUS_REFERRED', $3, $4, $5, $6)
	`
	if r.isSQLite {
		refedLedgerQ = `
			INSERT INTO loyalty_ledger (id, user_id, entry_type, points_delta, reference_type, reference_id, idempotency_key, balance_after, created_at)
			VALUES (?, ?, 'EARN', 50, 'REFERRAL_BONUS_REFERRED', ?, ?, ?, ?)
		`
	}
	refedIdempotency := "referral:referred:" + eventID
	_, _ = tx.ExecContext(ctx, refedLedgerQ, uuid.New().String(), userID, orderID, refedIdempotency, refedNewBal, now)

	// 3. Mark referral event REWARDED
	updateEventQ := `
		UPDATE referral_events
		SET status = 'REWARDED', first_order_id = $1, rewarded_at = $2
		WHERE id = $3
	`
	if r.isSQLite {
		updateEventQ = `
			UPDATE referral_events
			SET status = 'REWARDED', first_order_id = ?, rewarded_at = ?
			WHERE id = ?
		`
	}
	_, err = tx.ExecContext(ctx, updateEventQ, orderID, now, eventID)
	if err != nil {
		return "", 0, 0, false, err
	}

	if err := tx.Commit(); err != nil {
		return "", 0, 0, false, err
	}

	return refUserID, referrerPoints, referredPoints, true, nil
}

func (r *PostgresRepository) ExpirePendingReferrals(ctx context.Context) (int64, error) {
	query := `UPDATE referral_events SET status = 'EXPIRED' WHERE status = 'PENDING' AND expires_at < $1`
	if r.isSQLite {
		query = `UPDATE referral_events SET status = 'EXPIRED' WHERE status = 'PENDING' AND expires_at < ?`
	}
	res, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *PostgresRepository) getSubscriptionBonus(ctx context.Context, userID string) float64 {
	if r.db == nil {
		return 0.0
	}
	var bonus float64
	query := `
		SELECT 0.5
		FROM member_subscriptions s
		JOIN subscription_plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'ACTIVE'
		LIMIT 1
	`
	if r.isSQLite {
		query = `
			SELECT 0.5
			FROM member_subscriptions s
			JOIN subscription_plans p ON s.plan_id = p.id
			WHERE s.user_id = ? AND s.status = 'ACTIVE'
			LIMIT 1
		`
	}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&bonus)
	if err == nil && bonus > 0 {
		return bonus
	}
	return 0.0
}
