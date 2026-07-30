-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_session_id UUID UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    amount_paise BIGINT NOT NULL,
    loyalty_points_used BIGINT NOT NULL DEFAULT 0,
    loyalty_discount_paise BIGINT NOT NULL DEFAULT 0,
    payable_amount_paise BIGINT NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    gateway VARCHAR(20) NOT NULL,
    gateway_order_id VARCHAR(100) NULL,
    gateway_payment_id VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'INITIATED',
    failure_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_store ON payments(store_id);

-- Create payment_outbox table (Outbox Pattern)
CREATE TABLE IF NOT EXISTS payment_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ NULL,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partial index for high-performance outbox relay polling
CREATE INDEX IF NOT EXISTS idx_payment_outbox_unpublished ON payment_outbox(created_at) WHERE published_at IS NULL;

-- Create payment_webhook_events table (Idempotency log)
CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway VARCHAR(20) NOT NULL,
    gateway_event_id VARCHAR(150) UNIQUE NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    raw_payload JSONB NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create refunds table
CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id),
    amount_paise BIGINT NOT NULL,
    reason VARCHAR(50) NOT NULL,
    gateway_refund_id VARCHAR(100) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'INITIATED',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ NULL
);
