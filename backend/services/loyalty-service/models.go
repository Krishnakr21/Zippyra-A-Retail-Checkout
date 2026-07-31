package main

import (
	"time"
)

const (
	EntryEarn          = "EARN"
	EntryRedeemReserve = "REDEEM_RESERVE"
	EntryRedeemCommit  = "REDEEM_COMMIT"
	EntryRedeemRelease = "REDEEM_RELEASE"
	EntryReversal      = "REVERSAL"
	EntryAdjustment    = "ADJUSTMENT"

	TopicTierUpgraded = "loyalty.tier_upgraded"
	TopicPointsEarned = "loyalty.points_earned"
)

type LoyaltyAccount struct {
	UserID               string     `json:"user_id"`
	PointsBalance        int64      `json:"points_balance"`
	PointsReserved       int64      `json:"points_reserved"`
	LifetimePointsEarned int64      `json:"lifetime_points_earned"`
	Tier                 string     `json:"tier"`
	ReferralCode         string     `json:"referral_code"`
	TierUpdatedAt        *time.Time `json:"tier_updated_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ReferralEvent struct {
	ID             string     `json:"id"`
	ReferrerUserID string     `json:"referrer_user_id"`
	ReferredUserID string     `json:"referred_user_id"`
	ReferralCode   string     `json:"referral_code"`
	Status         string     `json:"status"` // PENDING, REWARDED, EXPIRED
	FirstOrderID   *string    `json:"first_order_id,omitempty"`
	RewardedAt     *time.Time `json:"rewarded_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
}

type ReferralCodeResponse struct {
	ReferralCode         string `json:"referral_code"`
	ShareText            string `json:"share_text"`
	ReferrerRewardPoints int64  `json:"referrer_reward_points"`
	ReferredRewardPoints int64  `json:"referred_reward_points"`
}

type ApplyReferralRequest struct {
	ReferralCode string `json:"referral_code"`
}

type LoyaltyLedger struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	EntryType      string    `json:"entry_type"`
	PointsDelta    int64     `json:"points_delta"`
	ReferenceType  *string   `json:"reference_type,omitempty"`
	ReferenceID    *string   `json:"reference_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	BalanceAfter   int64     `json:"balance_after"`
	CreatedAt      time.Time `json:"created_at"`
}

type LoyaltyTierConfig struct {
	Tier              string  `json:"tier"`
	MinLifetimePoints int64   `json:"min_lifetime_points"`
	EarnMultiplier    float64 `json:"earn_multiplier"`
	DisplayName       string  `json:"display_name"`
	DisplayOrder      int     `json:"display_order"`
}

// Request & Response DTOs
type ReservePointsRequest struct {
	UserID         string `json:"user_id"`
	Points         int64  `json:"points"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ReservePointsResponse struct {
	Reserved           bool  `json:"reserved"`
	PointsBalanceAfter int64 `json:"points_balance_after"`
}

type CommitPointsRequest struct {
	UserID         string `json:"user_id"`
	Points         int64  `json:"points"`
	IdempotencyKey string `json:"idempotency_key"`
}

type CommitPointsResponse struct {
	Committed bool `json:"committed"`
}

type ReleasePointsRequest struct {
	UserID         string `json:"user_id"`
	Points         int64  `json:"points"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ReleasePointsResponse struct {
	Released           bool  `json:"released"`
	PointsBalanceAfter int64 `json:"points_balance_after"`
}

type LoyaltyBalanceResponse struct {
	PointsBalance        int64   `json:"points_balance"`
	PointsReserved       int64   `json:"points_reserved"`
	Tier                 string  `json:"tier"`
	TierDisplayName      string  `json:"tier_display_name"`
	LifetimePointsEarned int64   `json:"lifetime_points_earned"`
	PointsToNextTier     *int64  `json:"points_to_next_tier"`
	NextTierName         *string `json:"next_tier_name"`
}

type LedgerItemResponse struct {
	EntryType     string    `json:"entry_type"`
	PointsDelta   int64     `json:"points_delta"`
	ReferenceType *string   `json:"reference_type,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	BalanceAfter  int64     `json:"balance_after"`
}

type LedgerHistoryResponse struct {
	Items    []LedgerItemResponse `json:"items"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// Kafka Payloads
type OrderCompletedPayload struct {
	OrderID           string    `json:"order_id"`
	UserID            string    `json:"user_id"`
	StoreID           string    `json:"store_id"`
	TotalPaise        int64     `json:"total_paise"`
	LoyaltyPointsUsed int64     `json:"loyalty_points_used"`
	PaymentMethod     string    `json:"payment_method"`
	Timestamp         time.Time `json:"ts"`
}

type OrderReturnedPayload struct {
	OrderID             string    `json:"order_id"`
	UserID              string    `json:"user_id"`
	StoreID             string    `json:"store_id"`
	ReturnedAmountPaise int64     `json:"returned_amount_paise"`
	OriginalTotalPaise  int64     `json:"original_total_paise"`
	ReturnID            string    `json:"return_id,omitempty"`
	Timestamp           time.Time `json:"ts"`
}

type TierUpgradedPayload struct {
	UserID    string    `json:"user_id"`
	OldTier   string    `json:"old_tier"`
	NewTier   string    `json:"new_tier"`
	Timestamp time.Time `json:"ts"`
}

type PointsEarnedPayload struct {
	UserID       string    `json:"user_id"`
	OrderID      string    `json:"order_id"`
	PointsEarned int64     `json:"points_earned"`
	NewBalance   int64     `json:"new_balance"`
	Timestamp    time.Time `json:"ts"`
}
