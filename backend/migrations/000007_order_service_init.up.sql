-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    items JSONB NOT NULL,
    subtotal_paise BIGINT NOT NULL,
    discount_paise BIGINT NOT NULL DEFAULT 0,
    cgst_paise BIGINT NOT NULL DEFAULT 0,
    sgst_paise BIGINT NOT NULL DEFAULT 0,
    igst_paise BIGINT NOT NULL DEFAULT 0,
    total_paise BIGINT NOT NULL,
    loyalty_points_used BIGINT NOT NULL DEFAULT 0,
    payment_method VARCHAR(20) NOT NULL,
    supply_type VARCHAR(15) NOT NULL DEFAULT 'INTRASTATE',
    status VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    invoice_s3_key TEXT NULL,
    irn VARCHAR(64) NULL,
    irn_ack_no VARCHAR(20) NULL,
    irn_ack_date TIMESTAMPTZ NULL,
    irn_qr_code TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ NULL
);

-- Create order_creation_outbox table
CREATE TABLE IF NOT EXISTS order_creation_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ NULL,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create order_items_returnable_flags table
CREATE TABLE IF NOT EXISTS order_items_returnable_flags (
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    barcode VARCHAR(20) NOT NULL,
    is_returnable BOOLEAN NOT NULL DEFAULT true,
    returned_qty INT NOT NULL DEFAULT 0,
    PRIMARY KEY (order_id, barcode)
);

-- Create return_requests table
CREATE TABLE IF NOT EXISTS return_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    items JSONB NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING_STAFF_REVIEW',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create Indexes
CREATE INDEX IF NOT EXISTS idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_store_created ON orders(store_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_creation_outbox_unpublished ON order_creation_outbox(created_at) WHERE published_at IS NULL;
