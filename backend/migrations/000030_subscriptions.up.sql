CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY,
    chain_id VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    price_paise BIGINT NOT NULL,
    billing_interval VARCHAR(20) NOT NULL,
    benefits JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS member_subscriptions (
    id UUID PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    razorpay_subscription_id VARCHAR(100) UNIQUE NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_member_subscriptions_user ON member_subscriptions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_member_subscriptions_rzp ON member_subscriptions(razorpay_subscription_id);

CREATE TABLE IF NOT EXISTS subscription_webhook_events (
    event_id VARCHAR(100) PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
