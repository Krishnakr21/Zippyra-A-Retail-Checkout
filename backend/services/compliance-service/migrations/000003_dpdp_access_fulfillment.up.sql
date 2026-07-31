CREATE TABLE IF NOT EXISTS dpdp_access_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dpdp_request_id UUID UNIQUE NOT NULL REFERENCES dpdp_requests(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'ASSEMBLING',
    s3_key TEXT NULL,
    expected_services JSONB NOT NULL,
    services_reported JSONB NOT NULL DEFAULT '[]',
    sections JSONB NOT NULL DEFAULT '{}',
    ready_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dpdp_access_exports_status ON dpdp_access_exports(status);
CREATE INDEX IF NOT EXISTS idx_dpdp_access_exports_expires ON dpdp_access_exports(expires_at) WHERE status != 'EXPIRED';
