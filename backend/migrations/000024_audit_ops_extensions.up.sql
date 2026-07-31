-- Feature Flags Table
CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key VARCHAR(100) PRIMARY KEY,
    description TEXT,
    scope_type VARCHAR(20) NOT NULL DEFAULT 'GLOBAL', -- GLOBAL | CHAIN | STORE | USER_PERCENTAGE
    enabled_globally BOOLEAN NOT NULL DEFAULT false,
    enabled_scope_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    user_percentage INT NULL CHECK (user_percentage IS NULL OR (user_percentage >= 0 AND user_percentage <= 100)),
    updated_by UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_scope ON feature_flags(scope_type);

-- Kafka DLQ Soft Discard Offsets Table
CREATE TABLE IF NOT EXISTS dlq_discarded_offsets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    "offset" BIGINT NOT NULL,
    discarded_by UUID NOT NULL,
    discarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason TEXT,
    CONSTRAINT uq_dlq_discarded_topic_offset UNIQUE (topic, "offset")
);

CREATE INDEX IF NOT EXISTS idx_dlq_discarded_topic ON dlq_discarded_offsets(topic);
