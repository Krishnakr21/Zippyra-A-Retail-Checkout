-- Reversal: recreate chains table structure in store-service DB.
-- Note: Data would need to be restored from backup or re-copied from
-- admin-store-service's DB; this .down.sql only recreates the schema.
CREATE TABLE IF NOT EXISTS chains (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    legal_entity_name VARCHAR(255),
    default_gstin_prefix VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
