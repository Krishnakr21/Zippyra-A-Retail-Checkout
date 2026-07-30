CREATE TABLE IF NOT EXISTS loyalty_accounts (
    user_id UUID PRIMARY KEY,
    points_balance BIGINT NOT NULL DEFAULT 0,
    points_reserved BIGINT NOT NULL DEFAULT 0,
    lifetime_points_earned BIGINT NOT NULL DEFAULT 0,
    tier VARCHAR(20) NOT NULL DEFAULT 'BRONZE',
    tier_updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (points_balance >= 0 AND points_reserved >= 0)
);

CREATE TABLE IF NOT EXISTS loyalty_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    entry_type VARCHAR(20) NOT NULL,
    points_delta BIGINT NOT NULL,
    reference_type VARCHAR(20),
    reference_id UUID,
    idempotency_key VARCHAR(150) UNIQUE NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS loyalty_tier_config (
    tier VARCHAR(20) PRIMARY KEY,
    min_lifetime_points BIGINT NOT NULL,
    earn_multiplier NUMERIC(3,2) NOT NULL DEFAULT 1.0,
    display_name VARCHAR(50) NOT NULL,
    display_order INT NOT NULL
);

-- Seed Loyalty Tiers
INSERT INTO loyalty_tier_config (tier, min_lifetime_points, earn_multiplier, display_name, display_order)
VALUES 
    ('BRONZE', 0, 1.00, 'Bronze Tier', 1),
    ('SILVER', 5000, 1.20, 'Silver Tier', 2),
    ('GOLD', 20000, 1.50, 'Gold Tier', 3),
    ('PLATINUM', 50000, 2.00, 'Platinum Tier', 4)
ON CONFLICT (tier) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_loyalty_ledger_user_created ON loyalty_ledger(user_id, created_at DESC);
