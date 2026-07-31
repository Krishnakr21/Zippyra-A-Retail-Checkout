DROP TABLE IF EXISTS account_recovery_requests;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_email;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_email_verified_at;
