CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    user_type VARCHAR(10) NOT NULL CHECK (user_type IN ('CUSTOMER', 'STAFF')),
    fcm_token TEXT NOT NULL,
    platform VARCHAR(10) NOT NULL CHECK (platform IN ('ANDROID', 'IOS')),
    device_id VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id VARCHAR(100) NOT NULL,
    user_type VARCHAR(10) NOT NULL CHECK (user_type IN ('CUSTOMER', 'STAFF')),
    notification_type VARCHAR(30) NOT NULL,
    channel VARCHAR(10) NOT NULL CHECK (channel IN ('PUSH', 'WHATSAPP', 'BOTH', 'NONE')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, notification_type)
);

CREATE TABLE IF NOT EXISTS notification_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    user_type VARCHAR(10) NOT NULL CHECK (user_type IN ('CUSTOMER', 'STAFF')),
    notification_type VARCHAR(30) NOT NULL,
    channel_sent VARCHAR(10) NOT NULL,
    title VARCHAR(200),
    body TEXT,
    deep_link VARCHAR(200),
    source_event_type VARCHAR(50) NOT NULL,
    source_event_id VARCHAR(150) NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_event_type, source_event_id, user_id, notification_type)
);

CREATE TABLE IF NOT EXISTS whatsapp_template_config (
    template_key VARCHAR(50) PRIMARY KEY,
    whatsapp_template_name VARCHAR(100) NOT NULL,
    is_approved BOOLEAN NOT NULL DEFAULT false,
    language VARCHAR(5) NOT NULL DEFAULT 'en',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ops_alert_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_type VARCHAR(10) NOT NULL CHECK (channel_type IN ('SLACK', 'EMAIL')),
    target VARCHAR(200) NOT NULL,
    alert_types JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_log_user ON notification_log(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_device_tokens_user_active ON device_tokens(user_id) WHERE is_active = true;
