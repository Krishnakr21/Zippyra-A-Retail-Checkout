DROP TABLE IF EXISTS account_recovery_requests;
ALTER TABLE users DROP COLUMN recovery_email;
ALTER TABLE users DROP COLUMN recovery_email_verified_at;
