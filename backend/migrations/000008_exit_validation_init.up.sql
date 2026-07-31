CREATE TABLE IF NOT EXISTS exit_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    gate_id VARCHAR(50) NOT NULL,
    result VARCHAR(30) NOT NULL,
    is_alarm BOOLEAN NOT NULL DEFAULT false,
    rfid_tag_ids JSONB NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS staff_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NULL,
    store_id UUID NOT NULL,
    gate_id VARCHAR(50) NOT NULL,
    staff_user_id UUID NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_exit_attempts_store_created ON exit_attempts(store_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exit_attempts_alarm ON exit_attempts(store_id) WHERE is_alarm = true;
