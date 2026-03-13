-- TGBot_Admin Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================== Groups ====================
CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    description TEXT,
    member_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{
        "verification_timeout": 300,
        "difficulty": "easy",
        "auto_approve": false,
        "kick_on_fail": true,
        "max_fail_count": 3,
        "admin_whitelist": []
    }'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_groups_chat_id ON groups(chat_id);
CREATE INDEX IF NOT EXISTS idx_groups_is_active ON groups(is_active);
CREATE INDEX IF NOT EXISTS idx_groups_last_active ON groups(last_active_at DESC NULLS LAST);

-- ==================== Blacklist ====================
CREATE TABLE IF NOT EXISTS blacklist (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES groups(chat_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    reason TEXT,
    banned_by BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chat_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_blacklist_chat_id ON blacklist(chat_id);
CREATE INDEX IF NOT EXISTS idx_blacklist_user_id ON blacklist(user_id);
CREATE INDEX IF NOT EXISTS idx_blacklist_created ON blacklist(created_at DESC);

-- ==================== Verification Logs ====================
CREATE TABLE IF NOT EXISTS verification_logs (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'timeout', 'auto_approved')),
    question TEXT,
    answer TEXT,
    user_answer TEXT,
    attempt_count INTEGER DEFAULT 1,
    duration_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_chat_id ON verification_logs(chat_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_user_id ON verification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_created ON verification_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_verification_logs_status ON verification_logs(status);

-- ==================== Plugins ====================
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    config JSONB DEFAULT '{}'::jsonb,
    last_restart_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plugins_enabled ON plugins(is_enabled);
CREATE INDEX IF NOT EXISTS idx_plugins_priority ON plugins(priority);

-- ==================== Action Logs ====================
CREATE TABLE IF NOT EXISTS action_logs (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT,
    user_id BIGINT,
    action_type VARCHAR(50) NOT NULL,
    action_data JSONB DEFAULT '{}'::jsonb,
    operator_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_action_logs_chat_id ON action_logs(chat_id);
CREATE INDEX IF NOT EXISTS idx_action_logs_created ON action_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_action_logs_type ON action_logs(action_type);

-- ==================== Admins ====================
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    role VARCHAR(20) DEFAULT 'admin' CHECK (role IN ('super_admin', 'admin', 'viewer')),
    permissions JSONB DEFAULT '[]'::jsonb,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admins_user_id ON admins(user_id);
CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role);

-- ==================== Functions ====================

-- Update timestamp function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_groups_updated_at BEFORE UPDATE ON groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================== Views ====================

-- Verification stats view
CREATE OR REPLACE VIEW verification_stats AS
SELECT
    chat_id,
    DATE(created_at) as date,
    COUNT(*) FILTER (WHERE status = 'success') as success_count,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_count,
    COUNT(*) FILTER (WHERE status = 'timeout') as timeout_count,
    COUNT(*) as total_count
FROM verification_logs
GROUP BY chat_id, DATE(created_at);

-- Group summary view
CREATE OR REPLACE VIEW group_summary AS
SELECT
    g.id,
    g.chat_id,
    g.title,
    g.member_count,
    g.is_active,
    g.last_active_at,
    COUNT(DISTINCT b.user_id) as blacklist_count,
    COALESCE(
        (SELECT success_count FROM verification_stats vs
         WHERE vs.chat_id = g.chat_id AND vs.date = CURRENT_DATE), 0
    ) as today_success,
    COALESCE(
        (SELECT failed_count FROM verification_stats vs
         WHERE vs.chat_id = g.chat_id AND vs.date = CURRENT_DATE), 0
    ) as today_failed
FROM groups g
LEFT JOIN blacklist b ON g.chat_id = b.chat_id
GROUP BY g.id, g.chat_id, g.title, g.member_count, g.is_active, g.last_active_at;
