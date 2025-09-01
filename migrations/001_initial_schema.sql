-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_pro BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Profiles table
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    income DECIMAL(10,2) NOT NULL DEFAULT 0,
    fixed_expenses DECIMAL(10,2) NOT NULL DEFAULT 0,
    debts_min DECIMAL(10,2) NOT NULL DEFAULT 0,
    savings_liquid DECIMAL(10,2) NOT NULL DEFAULT 0,
    guardrails_json JSONB,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    anon BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '30 days'
);

-- Decisions table
CREATE TABLE decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    input_json JSONB NOT NULL,
    verdict VARCHAR(10) NOT NULL CHECK (verdict IN ('YES', 'NO')),
    score DECIMAL(5,2) NOT NULL,
    reasons_json JSONB NOT NULL,
    cost DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    shared BOOLEAN NOT NULL DEFAULT FALSE,
    share_url VARCHAR(255) UNIQUE
);

-- Credit ledger table
CREATE TABLE credit_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delta INTEGER NOT NULL,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('purchase', 'referral', 'grant')),
    ref_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    processor VARCHAR(50) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    raw_webhook_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Referrals table
CREATE TABLE referrals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inviter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'expired')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Feature flags table
CREATE TABLE feature_flags (
    key VARCHAR(100) PRIMARY KEY,
    value_json JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- FX rates table
CREATE TABLE fx_rates (
    code VARCHAR(3) PRIMARY KEY,
    rate_to_usd DECIMAL(10,6) NOT NULL,
    fetched_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_decisions_user_id ON decisions(user_id);
CREATE INDEX idx_decisions_created_at ON decisions(created_at);
CREATE INDEX idx_decisions_session_id ON decisions(session_id);
CREATE INDEX idx_credit_ledger_user_id ON credit_ledger(user_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_referrals_inviter ON referrals(inviter_user_id);
CREATE INDEX idx_referrals_email ON referrals(invitee_email);

-- Insert default FX rates
INSERT INTO fx_rates (code, rate_to_usd, fetched_at) VALUES
    ('USD', 1.000000, NOW()),
    ('EUR', 0.850000, NOW()),
    ('MAD', 10.500000, NOW()),
    ('GBP', 0.750000, NOW());

-- Insert default feature flags
INSERT INTO feature_flags (key, value_json) VALUES
    ('decision_rate_limit', '{"free": 3, "starter": 100, "pro": 1000}'),
    ('ai_explanations_enabled', 'true'),
    ('sharing_enabled', 'true'),
    ('referrals_enabled', 'true');