CREATE TABLE IF NOT EXISTS chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    legal_entity_name VARCHAR(200),
    default_gstin_prefix VARCHAR(2),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE stores ADD CONSTRAINT fk_stores_chain FOREIGN KEY (chain_id) REFERENCES chains(id);
