-- Migration 000020: Cart Offer Authoring & Audit Schema

CREATE TABLE IF NOT EXISTS offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    store_id UUID NULL,
    type VARCHAR(30) NOT NULL,
    applies_to VARCHAR(20) NOT NULL,
    target_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    rule_config JSONB NOT NULL,
    min_cart_value_paise BIGINT NOT NULL DEFAULT 0,
    max_discount_paise BIGINT NULL,
    priority INT NOT NULL DEFAULT 100,
    active_from TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active_until TIMESTAMPTZ NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_offers_store ON offers(store_id) WHERE store_id IS NOT NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_offers_chain_wide ON offers(chain_id) WHERE store_id IS NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_offers_schedule ON offers(active_from, active_until) WHERE is_active = true;

-- Ensure offer_rules_audit table exists (if not created by earlier migrations)
CREATE TABLE IF NOT EXISTS offer_rules_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    ruleset JSONB NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_offer_rules_audit_store ON offer_rules_audit(store_id, activated_at DESC);
