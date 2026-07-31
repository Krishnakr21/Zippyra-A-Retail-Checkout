ALTER TABLE auth_sessions DROP COLUMN IF EXISTS device_label;
ALTER TABLE auth_sessions DROP COLUMN IF EXISTS last_used_at;
