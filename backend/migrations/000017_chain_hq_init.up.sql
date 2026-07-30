CREATE TABLE IF NOT EXISTS chain_hq_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    phone VARCHAR(15) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL, -- OWNER | FINANCE | OPERATIONS
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chain_hq_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_hq_user_id UUID NOT NULL REFERENCES chain_hq_users(id) ON DELETE CASCADE,
    device_id VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS chain_bulk_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    per_store_job_ids JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PROCESSING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chain_hq_users_chain ON chain_hq_users(chain_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_chain_hq_sessions_active ON chain_hq_sessions(chain_hq_user_id) WHERE revoked_at IS NULL;
