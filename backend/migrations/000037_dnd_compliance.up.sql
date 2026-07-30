CREATE TABLE IF NOT EXISTS dnd_registry_cache (
    phone VARCHAR(15) PRIMARY KEY,
    is_dnd_registered BOOLEAN NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE whatsapp_template_config ADD COLUMN IF NOT EXISTS meta_category VARCHAR(20) NOT NULL DEFAULT 'UTILITY';
