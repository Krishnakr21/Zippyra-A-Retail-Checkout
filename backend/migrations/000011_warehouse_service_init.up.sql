CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    vendor_name VARCHAR(150) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    source VARCHAR(20) NOT NULL DEFAULT 'MANUAL',
    created_by UUID NULL,
    auto_reorder_item_barcode VARCHAR(20) NULL,
    auto_reorder_date DATE NULL,
    expected_delivery_date DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMP WITH TIME ZONE NULL,
    completed_at TIMESTAMP WITH TIME ZONE NULL,
    CONSTRAINT uq_po_auto_reorder UNIQUE (store_id, auto_reorder_item_barcode, auto_reorder_date)
);

CREATE TABLE IF NOT EXISTS po_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    barcode VARCHAR(20) NOT NULL,
    qty_ordered INT NOT NULL,
    unit_cost_paise BIGINT NOT NULL,
    qty_received INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS goods_received_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id UUID NULL REFERENCES purchase_orders(id) ON DELETE SET NULL,
    store_id UUID NOT NULL,
    received_by UUID NOT NULL,
    vendor_invoice_ref VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS grn_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grn_id UUID NOT NULL REFERENCES goods_received_notes(id) ON DELETE CASCADE,
    barcode VARCHAR(20) NOT NULL,
    qty_expected INT NULL,
    qty_received INT NOT NULL,
    unit_cost_paise BIGINT NOT NULL,
    qc_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    qc_note TEXT
);

CREATE TABLE IF NOT EXISTS transfer_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_store_id UUID NOT NULL,
    dest_store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'REQUESTED',
    requested_by UUID NOT NULL,
    rejection_reason TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP WITH TIME ZONE NULL,
    shipped_at TIMESTAMP WITH TIME ZONE NULL,
    received_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS transfer_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES transfer_orders(id) ON DELETE CASCADE,
    barcode VARCHAR(20) NOT NULL,
    qty_requested INT NOT NULL,
    qty_shipped INT NULL,
    qty_received INT NULL
);

CREATE INDEX IF NOT EXISTS idx_po_store_status ON purchase_orders(store_id, status);
CREATE INDEX IF NOT EXISTS idx_grn_store_status ON goods_received_notes(store_id, status);
CREATE INDEX IF NOT EXISTS idx_transfer_dest_status ON transfer_orders(dest_store_id, status);
CREATE INDEX IF NOT EXISTS idx_transfer_source_status ON transfer_orders(source_store_id, status);
