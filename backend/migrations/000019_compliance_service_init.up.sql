CREATE TABLE IF NOT EXISTS irn_records (
    id UUID PRIMARY KEY,
    order_id UUID UNIQUE NOT NULL,
    store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    irn VARCHAR(64) NULL,
    ack_no VARCHAR(20) NULL,
    ack_date TIMESTAMPTZ NULL,
    signed_qr_code TEXT NULL,
    irp_payload JSONB NOT NULL,
    irp_response JSONB NULL,
    failure_reason TEXT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    submitted_at TIMESTAMPTZ NULL,
    issued_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS irn_outbox (
    id UUID PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ NULL,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dpdp_consents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    consent_type VARCHAR(30) NOT NULL,
    granted BOOLEAN NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    consent_version VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dpdp_requests (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    request_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'RECEIVED',
    detail TEXT NULL,
    handled_by UUID NULL,
    rejection_reason TEXT NULL,
    sla_due_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS dpdp_deletion_audit (
    id UUID PRIMARY KEY,
    dpdp_request_id UUID NOT NULL REFERENCES dpdp_requests(id) ON DELETE CASCADE,
    service_name VARCHAR(50) NOT NULL,
    tables_affected JSONB NULL,
    rows_deleted_count INT NOT NULL DEFAULT 0,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS merchant_kyc (
    store_id UUID PRIMARY KEY,
    gstin_verified BOOLEAN NOT NULL DEFAULT false,
    pan_number VARCHAR(10) NULL,
    pan_verified BOOLEAN NOT NULL DEFAULT false,
    bank_account_last4 VARCHAR(4) NULL,
    razorpay_account_id VARCHAR(100) NULL,
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    kyc_completed_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS velocity_alerts (
    id UUID PRIMARY KEY,
    store_id UUID NOT NULL,
    alert_type VARCHAR(30) NOT NULL,
    detail JSONB NULL,
    resolved_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS settlement_reconciliation_reports (
    id UUID PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,
    total_transactions INT NOT NULL DEFAULT 0,
    total_amount_paise BIGINT NOT NULL DEFAULT 0,
    total_settled_amount_paise BIGINT NOT NULL DEFAULT 0,
    discrepancy_count INT NOT NULL DEFAULT 0,
    discrepancies JSONB NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dpdp_requests_status ON dpdp_requests(status) WHERE status != 'COMPLETED';
CREATE INDEX IF NOT EXISTS idx_velocity_alerts_unresolved ON velocity_alerts(store_id) WHERE resolved_at IS NULL;
