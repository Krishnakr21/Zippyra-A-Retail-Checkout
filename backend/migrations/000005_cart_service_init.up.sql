-- Create checkout_sessions table
CREATE TABLE IF NOT EXISTS checkout_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    items JSONB NOT NULL,
    subtotal_paise BIGINT NOT NULL,
    discount_paise BIGINT NOT NULL DEFAULT 0,
    cgst_paise BIGINT NOT NULL DEFAULT 0,
    sgst_paise BIGINT NOT NULL DEFAULT 0,
    igst_paise BIGINT NOT NULL DEFAULT 0,
    total_paise BIGINT NOT NULL,
    coupon_code VARCHAR(30) NULL,
    supply_type VARCHAR(15) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Partial index for active PENDING checkout sessions
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_user ON checkout_sessions(user_id) WHERE status='PENDING';
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_store ON checkout_sessions(store_id);

-- Create offer_rules_audit table for dispute resolution logging
CREATE TABLE IF NOT EXISTS offer_rules_audit (
    id UUID PRIMARY KEY,
    store_id UUID NOT NULL,
    ruleset JSONB NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
