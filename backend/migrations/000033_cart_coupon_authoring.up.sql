CREATE TABLE IF NOT EXISTS coupons (
    id VARCHAR(36) PRIMARY KEY,
    chain_id VARCHAR(36) NOT NULL,
    store_id VARCHAR(36) NULL,
    code VARCHAR(30) NOT NULL,
    discount_type VARCHAR(20) NOT NULL,
    discount_value NUMERIC(10,2) NOT NULL,
    min_cart_value_paise BIGINT NOT NULL DEFAULT 0,
    max_uses INT NULL,
    max_uses_per_customer INT NOT NULL DEFAULT 1,
    current_use_count INT NOT NULL DEFAULT 0,
    active_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active_until TIMESTAMP WITH TIME ZONE NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (chain_id, code)
);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id VARCHAR(36) PRIMARY KEY,
    coupon_id VARCHAR(36) NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    checkout_session_id VARCHAR(36) NOT NULL,
    redeemed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (coupon_id, checkout_session_id)
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_user ON coupon_redemptions(coupon_id, user_id);
