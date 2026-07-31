ALTER TABLE loyalty_accounts ADD COLUMN IF NOT EXISTS referral_code VARCHAR(8) UNIQUE;

CREATE TABLE IF NOT EXISTS referral_events (
    id UUID PRIMARY KEY,
    referrer_user_id VARCHAR(64) NOT NULL,
    referred_user_id VARCHAR(64) UNIQUE NOT NULL,
    referral_code VARCHAR(8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    first_order_id VARCHAR(64) NULL,
    rewarded_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_referral_events_code ON referral_events(referral_code);
CREATE INDEX IF NOT EXISTS idx_referral_events_status_expires ON referral_events(status, expires_at);
