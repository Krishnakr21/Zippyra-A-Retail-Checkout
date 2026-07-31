CREATE TABLE IF NOT EXISTS stock_levels (
    store_id UUID NOT NULL,
    barcode VARCHAR(20) NOT NULL,
    on_hand_qty BIGINT NOT NULL DEFAULT 0 CHECK (on_hand_qty >= 0),
    reorder_point INT NOT NULL DEFAULT 10,
    reorder_qty INT NOT NULL DEFAULT 50,
    low_stock_alerted BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (store_id, barcode)
);

CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    barcode VARCHAR(20) NOT NULL,
    movement_type VARCHAR(20) NOT NULL,
    qty_delta BIGINT NOT NULL,
    reference_type VARCHAR(20),
    reference_id UUID,
    note TEXT,
    created_by UUID NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (store_id, barcode, reference_type, reference_id, movement_type)
);

CREATE TABLE IF NOT EXISTS stock_counts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    barcode VARCHAR(20) NOT NULL,
    expected_qty BIGINT NOT NULL,
    counted_qty BIGINT NOT NULL,
    variance_qty BIGINT NOT NULL,
    counted_by UUID NOT NULL,
    counted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shrinkage_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    date DATE NOT NULL,
    total_variance_qty BIGINT NOT NULL,
    total_expected_qty BIGINT NOT NULL,
    shrinkage_percent NUMERIC(5,2) NOT NULL,
    item_count INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (store_id, date)
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_store_barcode ON stock_movements(store_id, barcode, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_levels_low_stock ON stock_levels(store_id) WHERE on_hand_qty <= reorder_point;
