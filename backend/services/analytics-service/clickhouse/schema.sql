-- ClickHouse Analytics Schema
-- Applied via backend/services/analytics-service/clickhouse/migrate.go

CREATE TABLE IF NOT EXISTS sales_events (
    event_date Date,
    event_time DateTime,
    store_id UUID,
    chain_id UUID,
    order_id UUID,
    total_paise Int64,
    discount_paise Int64,
    cgst_paise Int64,
    sgst_paise Int64,
    igst_paise Int64,
    payment_method LowCardinality(String),
    item_count UInt16
) ENGINE = ReplacingMergeTree(event_time)
PARTITION BY toYYYYMM(event_date)
ORDER BY (store_id, event_date, order_id);

CREATE TABLE IF NOT EXISTS order_items_events (
    event_date Date,
    order_id UUID,
    store_id UUID,
    chain_id UUID,
    barcode String,
    product_name String,
    qty UInt16,
    line_total_paise Int64
) ENGINE = ReplacingMergeTree(event_date)
PARTITION BY toYYYYMM(event_date)
ORDER BY (store_id, event_date, order_id, barcode);

CREATE TABLE IF NOT EXISTS funnel_events (
    event_date Date,
    event_time DateTime,
    store_id UUID,
    session_id UUID,
    stage LowCardinality(String)
    -- SESSION_STARTED | CHECKOUT_INITIATED | PAYMENT_CONFIRMED | ORDER_COMPLETED | EXIT_VALIDATED
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (store_id, event_date, session_id, stage);

CREATE TABLE IF NOT EXISTS transaction_hourly (
    event_date Date,
    hour UInt8,
    day_of_week UInt8, -- 0=Sunday, 1=Monday, ..., 6=Saturday
    store_id UUID,
    transaction_count UInt32
) ENGINE = SummingMergeTree(transaction_count)
PARTITION BY toYYYYMM(event_date)
ORDER BY (store_id, day_of_week, hour, event_date);
