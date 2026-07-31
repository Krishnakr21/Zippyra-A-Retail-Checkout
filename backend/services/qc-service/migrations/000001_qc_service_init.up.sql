CREATE TABLE IF NOT EXISTS qc_reviews (
    id UUID PRIMARY KEY,
    grn_id UUID NOT NULL UNIQUE,
    store_id UUID NOT NULL,
    line_items JSONB NOT NULL,
    overall_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    reviewed_by UUID NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qc_reviews_store_status ON qc_reviews(store_id, overall_status);
