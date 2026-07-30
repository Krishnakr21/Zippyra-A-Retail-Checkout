-- Create app_versions table for Customer App force-update & version management
CREATE TABLE IF NOT EXISTS app_versions (
    platform VARCHAR(10) PRIMARY KEY,
    min_supported_version VARCHAR(20) NOT NULL,
    latest_version VARCHAR(20) NOT NULL,
    hard_update_below VARCHAR(20) NOT NULL,
    soft_update_message TEXT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default version rows for ANDROID and IOS platforms
INSERT INTO app_versions (platform, min_supported_version, latest_version, hard_update_below, soft_update_message, updated_at)
VALUES 
    ('ANDROID', '1.0.0', '1.0.0', '1.0.0', NULL, NOW()),
    ('IOS', '1.0.0', '1.0.0', '1.0.0', NULL, NOW())
ON CONFLICT (platform) DO NOTHING;
