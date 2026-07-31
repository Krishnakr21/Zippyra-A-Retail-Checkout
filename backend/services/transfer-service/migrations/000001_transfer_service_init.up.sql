CREATE TABLE IF NOT EXISTS transfer_orders (
    id UUID PRIMARY KEY,
    source_store_id UUID NOT NULL,
    dest_store_id UUID NOT NULL,
    chain_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'REQUESTED',
    requested_by UUID NULL,
    rejection_reason TEXT NULL,
    shipped_at TIMESTAMPTZ NULL,
    received_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transfer_line_items (
    id UUID PRIMARY KEY,
    transfer_id UUID NOT NULL REFERENCES transfer_orders(id) ON DELETE CASCADE,
    barcode VARCHAR(64) NOT NULL,
    qty_requested INT NOT NULL,
    qty_shipped INT NULL,
    qty_received INT NULL
);

CREATE INDEX IF NOT EXISTS idx_transfer_dest_status ON transfer_orders(dest_store_id, status);
CREATE INDEX IF NOT EXISTS idx_transfer_source_status ON transfer_orders(source_store_id, status);
