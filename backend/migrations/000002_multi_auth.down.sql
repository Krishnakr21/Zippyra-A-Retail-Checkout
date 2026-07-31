-- Rollback multi-auth schema changes

DROP TABLE IF EXISTS auth_sessions;

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_auth_identity;

ALTER TABLE users 
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS google_sub,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS phone_verified_at,
    DROP COLUMN IF EXISTS auth_provider_last;
