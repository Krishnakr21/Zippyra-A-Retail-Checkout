package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetActivePlansByChainID(ctx context.Context, chainID string) ([]*SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id string) (*SubscriptionPlan, error)
	CreateSubscription(ctx context.Context, sub *MemberSubscription) error
	GetActiveUserSubscription(ctx context.Context, userID string) (*MemberSubscription, error)
	UpdateSubscriptionStatusByRazorpayID(ctx context.Context, rzpSubID, status string, periodEnd *time.Time) error
	CancelSubscriptionByUserID(ctx context.Context, userID string) error
	ProcessWebhookEventIdempotent(ctx context.Context, eventID, eventType, rzpSubID, status string, periodEnd *time.Time) (bool, error)
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
	repo := &PostgresRepository{db: db, isSQLite: isSQLite}
	_ = repo.seedDefaultPlans(context.Background())
	return repo
}

func (r *PostgresRepository) seedDefaultPlans(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	var count int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscription_plans").Scan(&count)
	if count > 0 {
		return nil
	}

	monthlyBenefits, _ := json.Marshal(BenefitsDTO{LoyaltyMultiplierBonus: 0.5, FreeDelivery: true})
	annualBenefits, _ := json.Marshal(BenefitsDTO{LoyaltyMultiplierBonus: 0.5, FreeDelivery: true})

	now := time.Now()
	p1 := &SubscriptionPlan{
		ID:              "plan-smart-saver-monthly",
		ChainID:         "chain-hq-001",
		Name:            "Smart Saver Monthly",
		PricePaise:      19900, // ₹199/month
		BillingInterval: "MONTHLY",
		Benefits:        monthlyBenefits,
		IsActive:        true,
		CreatedAt:       now,
	}
	p2 := &SubscriptionPlan{
		ID:              "plan-smart-saver-annual",
		ChainID:         "chain-hq-001",
		Name:            "Smart Saver Annual",
		PricePaise:      149900, // ₹1,499/year
		BillingInterval: "ANNUAL",
		Benefits:        annualBenefits,
		IsActive:        true,
		CreatedAt:       now,
	}

	for _, p := range []*SubscriptionPlan{p1, p2} {
		q := `INSERT INTO subscription_plans (id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING`
		if r.isSQLite {
			q = `INSERT INTO subscription_plans (id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`
		}
		_, _ = r.db.ExecContext(ctx, q, p.ID, p.ChainID, p.Name, p.PricePaise, p.BillingInterval, string(p.Benefits), p.IsActive, p.CreatedAt)
	}
	return nil
}

func (r *PostgresRepository) GetActivePlansByChainID(ctx context.Context, chainID string) ([]*SubscriptionPlan, error) {
	if chainID == "" {
		chainID = "chain-hq-001"
	}
	query := `SELECT id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at
			  FROM subscription_plans WHERE chain_id = $1 AND is_active = true ORDER BY price_paise ASC`
	if r.isSQLite {
		query = `SELECT id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at
				 FROM subscription_plans WHERE chain_id = ? AND is_active = 1 ORDER BY price_paise ASC`
	}
	rows, err := r.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		var bStr string
		if err := rows.Scan(&p.ID, &p.ChainID, &p.Name, &p.PricePaise, &p.BillingInterval, &bStr, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Benefits = json.RawMessage(bStr)
		plans = append(plans, &p)
	}
	return plans, nil
}

func (r *PostgresRepository) GetPlanByID(ctx context.Context, id string) (*SubscriptionPlan, error) {
	query := `SELECT id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at
			  FROM subscription_plans WHERE id = $1`
	if r.isSQLite {
		query = `SELECT id, chain_id, name, price_paise, billing_interval, benefits, is_active, created_at
				 FROM subscription_plans WHERE id = ?`
	}
	row := r.db.QueryRowContext(ctx, query, id)

	var p SubscriptionPlan
	var bStr string
	err := row.Scan(&p.ID, &p.ChainID, &p.Name, &p.PricePaise, &p.BillingInterval, &bStr, &p.IsActive, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Benefits = json.RawMessage(bStr)
	return &p, nil
}

func (r *PostgresRepository) CreateSubscription(ctx context.Context, sub *MemberSubscription) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	query := `INSERT INTO member_subscriptions (id, user_id, plan_id, status, razorpay_subscription_id, current_period_end, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if r.isSQLite {
		query = `INSERT INTO member_subscriptions (id, user_id, plan_id, status, razorpay_subscription_id, current_period_end, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`
	}
	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.RazorpaySubscriptionID, sub.CurrentPeriodEnd, sub.CreatedAt)
	return err
}

func (r *PostgresRepository) GetActiveUserSubscription(ctx context.Context, userID string) (*MemberSubscription, error) {
	query := `SELECT id, user_id, plan_id, status, razorpay_subscription_id, current_period_end, created_at
			  FROM member_subscriptions WHERE user_id = $1 AND status = 'ACTIVE' ORDER BY created_at DESC LIMIT 1`
	if r.isSQLite {
		query = `SELECT id, user_id, plan_id, status, razorpay_subscription_id, current_period_end, created_at
				 FROM member_subscriptions WHERE user_id = ? AND status = 'ACTIVE' ORDER BY created_at DESC LIMIT 1`
	}
	row := r.db.QueryRowContext(ctx, query, userID)

	var s MemberSubscription
	err := row.Scan(&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.RazorpaySubscriptionID, &s.CurrentPeriodEnd, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	plan, _ := r.GetPlanByID(ctx, s.PlanID)
	s.Plan = plan
	return &s, nil
}

func (r *PostgresRepository) UpdateSubscriptionStatusByRazorpayID(ctx context.Context, rzpSubID, status string, periodEnd *time.Time) error {
	query := `UPDATE member_subscriptions SET status = $1, current_period_end = $2 WHERE razorpay_subscription_id = $3`
	if r.isSQLite {
		query = `UPDATE member_subscriptions SET status = ?, current_period_end = ? WHERE razorpay_subscription_id = ?`
	}
	_, err := r.db.ExecContext(ctx, query, status, periodEnd, rzpSubID)
	return err
}

func (r *PostgresRepository) CancelSubscriptionByUserID(ctx context.Context, userID string) error {
	query := `UPDATE member_subscriptions SET status = 'CANCELLED' WHERE user_id = $1 AND status = 'ACTIVE'`
	if r.isSQLite {
		query = `UPDATE member_subscriptions SET status = 'CANCELLED' WHERE user_id = ? AND status = 'ACTIVE'`
	}
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *PostgresRepository) ProcessWebhookEventIdempotent(ctx context.Context, eventID, eventType, rzpSubID, status string, periodEnd *time.Time) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// 1. Check idempotency
	insQ := `INSERT INTO subscription_webhook_events (event_id, event_type, processed_at) VALUES ($1, $2, $3)`
	if r.isSQLite {
		insQ = `INSERT INTO subscription_webhook_events (event_id, event_type, processed_at) VALUES (?, ?, ?)`
	}
	_, err = tx.ExecContext(ctx, insQ, eventID, eventType, time.Now())
	if err != nil {
		// Already processed event
		return false, nil
	}

	// 2. Update subscription status
	updQ := `UPDATE member_subscriptions SET status = $1, current_period_end = $2 WHERE razorpay_subscription_id = $3`
	if r.isSQLite {
		updQ = `UPDATE member_subscriptions SET status = ?, current_period_end = ? WHERE razorpay_subscription_id = ?`
	}
	_, err = tx.ExecContext(ctx, updQ, status, periodEnd, rzpSubID)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
