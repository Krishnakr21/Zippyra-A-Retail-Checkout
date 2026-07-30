CREATE TABLE IF NOT EXISTS feedback_submissions (
    id UUID PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    source_app VARCHAR(50) NOT NULL,
    nps_score INT NULL,
    comment TEXT NULL,
    context VARCHAR(50) NOT NULL DEFAULT 'general',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedback_source_created ON feedback_submissions(source_app, created_at DESC);
