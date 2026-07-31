CREATE TABLE IF NOT EXISTS device_heartbeats (
    device_id UUID NOT NULL,
    store_id UUID NOT NULL,
    ts TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL
);

SELECT create_hypertable('device_heartbeats', 'ts', if_not_exists => TRUE);
SELECT add_retention_policy('device_heartbeats', INTERVAL '30 days', if_not_exists => TRUE);
