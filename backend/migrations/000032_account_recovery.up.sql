ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_email VARCHAR(255) NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_email_verified_at TIMESTAMP WITH TIME ZONE NULL;

CREATE TABLE IF NOT EXISTS account_recovery_requests (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    original_identifier VARCHAR(255) NOT NULL,
    new_identifier VARCHAR(255) NOT NULL,
    verification_method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    support_ticket_id VARCHAR(36) NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE NULL
);
