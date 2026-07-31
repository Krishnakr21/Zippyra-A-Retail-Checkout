-- Multi-auth additions for users table and auth_sessions table

ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE,
    ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255) UNIQUE,
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS auth_provider_last VARCHAR(20) NULL;

-- Ensure phone is nullable (if not already) and set length
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- Add check constraint ensuring at least one identifier is present
ALTER TABLE users 
    DROP CONSTRAINT IF EXISTS chk_user_auth_identity;

ALTER TABLE users 
    ADD CONSTRAINT chk_user_auth_identity 
    CHECK (phone IS NOT NULL OR email IS NOT NULL OR google_sub IS NOT NULL);

-- Create auth_sessions table
CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_refresh_hash ON auth_sessions(refresh_token_hash);
