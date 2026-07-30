CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id VARCHAR(100) NOT NULL,
    requester_type VARCHAR(10) NOT NULL CHECK (requester_type IN ('CUSTOMER', 'STAFF')),
    store_id VARCHAR(100),
    chain_id VARCHAR(100),
    category VARCHAR(30) NOT NULL,
    related_order_id VARCHAR(100),
    subject VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(10) NOT NULL DEFAULT 'NORMAL' CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ASSIGNED', 'WAITING_ON_CUSTOMER', 'RESOLVED', 'CLOSED')),
    assigned_agent_id VARCHAR(100),
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    sla_due_at TIMESTAMPTZ NOT NULL,
    is_sla_warned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_id VARCHAR(100) NOT NULL,
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('CUSTOMER', 'STAFF', 'ADMIN', 'SYSTEM')),
    body TEXT NOT NULL,
    is_internal_note BOOLEAN NOT NULL DEFAULT false,
    attachments JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_auto_priority_rules (
    category VARCHAR(30) PRIMARY KEY,
    default_priority VARCHAR(10) NOT NULL CHECK (default_priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT'))
);

-- Seed auto-priority defaults
INSERT INTO ticket_auto_priority_rules (category, default_priority) VALUES
    ('EXIT_GATE_ISSUE', 'URGENT'),
    ('PAYMENT_ISSUE', 'HIGH'),
    ('ORDER_ISSUE', 'NORMAL'),
    ('ACCOUNT_ISSUE', 'NORMAL'),
    ('APP_BUG', 'LOW'),
    ('DEVICE_ISSUE', 'NORMAL'),
    ('OTHER', 'LOW')
ON CONFLICT (category) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_support_tickets_requester ON support_tickets(requester_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_open ON support_tickets(status, sla_due_at) WHERE status NOT IN ('RESOLVED', 'CLOSED');
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned ON support_tickets(assigned_agent_id) WHERE status NOT IN ('RESOLVED', 'CLOSED');
