-- Migration: 000004_catalog_service_init.up.sql
-- Catalog Service schema for categories, products, hsn_gst_rates, catalog_import_jobs, and catalog_sync_seq

CREATE SEQUENCE IF NOT EXISTS catalog_sync_seq START WITH 1 INCREMENT BY 1;

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    parent_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hsn_gst_rates (
    hsn_code VARCHAR(8) PRIMARY KEY,
    gst_rate_percent NUMERIC(4,2) NOT NULL,
    cess_percent NUMERIC(4,2) DEFAULT 0,
    description VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    barcode VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    price_paise BIGINT NOT NULL,
    mrp_paise BIGINT NOT NULL,
    hsn_code VARCHAR(8) NOT NULL REFERENCES hsn_gst_rates(hsn_code),
    is_active BOOLEAN DEFAULT true,
    is_returnable BOOLEAN DEFAULT true,
    image_url TEXT,
    thumbnail_url TEXT,
    sync_seq BIGINT NOT NULL DEFAULT nextval('catalog_sync_seq'),
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Unique index ensuring active barcode uniqueness per store
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_store_barcode ON products(store_id, barcode) WHERE deleted_at IS NULL;

-- Delta sync and category lookup performance indexes
CREATE INDEX IF NOT EXISTS idx_products_store_sync ON products(store_id, sync_seq);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

CREATE TABLE IF NOT EXISTS catalog_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL,
    chain_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    total_rows INT DEFAULT 0,
    processed_rows INT DEFAULT 0,
    error_rows JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ NULL
);
