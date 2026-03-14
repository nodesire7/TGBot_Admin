-- TGBot_Admin Multi-Bot Support
-- PostgreSQL 15+

-- ==================== Bots ====================
CREATE TABLE IF NOT EXISTS bots (
    id SERIAL PRIMARY KEY,
    bot_id BIGINT UNIQUE NOT NULL,           -- Telegram bot user ID
    username VARCHAR(255),                    -- @bot_username
    name VARCHAR(255),                        -- Bot display name
    token VARCHAR(255) NOT NULL,              -- Bot API token
    is_active BOOLEAN DEFAULT TRUE,           -- Bot enabled/disabled
    is_primary BOOLEAN DEFAULT FALSE,         -- Primary bot flag
    config JSONB DEFAULT '{
        "verification_timeout": 300,
        "difficulty": "easy",
        "auto_approve": false,
        "kick_on_fail": true,
        "max_fail_count": 3
    }'::jsonb,
    status JSONB DEFAULT '{
        "online": false,
        "pid": null,
        "memory_mb": 0,
        "cpu_percent": 0,
        "started_at": null,
        "error": null
    }'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_bots_bot_id ON bots(bot_id);
CREATE INDEX IF NOT EXISTS idx_bots_is_active ON bots(is_active);
CREATE INDEX IF NOT EXISTS idx_bots_is_primary ON bots(is_primary);

-- Update groups table to reference bots
ALTER TABLE groups ADD COLUMN IF NOT EXISTS bot_id INTEGER REFERENCES bots(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_groups_bot_id ON groups(bot_id);

-- Update verification_logs to reference bots
ALTER TABLE verification_logs ADD COLUMN IF NOT EXISTS bot_id INTEGER REFERENCES bots(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_verification_logs_bot_id ON verification_logs(bot_id);

-- Update action_logs to reference bots
ALTER TABLE action_logs ADD COLUMN IF NOT EXISTS bot_id INTEGER REFERENCES bots(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_action_logs_bot_id ON action_logs(bot_id);

-- Update triggers
CREATE TRIGGER update_bots_updated_at BEFORE UPDATE ON bots
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Update group_summary view to include bot info
DROP VIEW IF EXISTS group_summary;
CREATE OR REPLACE VIEW group_summary AS
SELECT
    g.id,
    g.chat_id,
    g.title,
    g.member_count,
    g.is_active,
    g.last_active_at,
    g.bot_id,
    b.name as bot_name,
    b.username as bot_username,
    COUNT(DISTINCT bl.user_id) as blacklist_count,
    COALESCE(
        (SELECT success_count FROM verification_stats vs
         WHERE vs.chat_id = g.chat_id AND vs.date = CURRENT_DATE), 0
    ) as today_success,
    COALESCE(
        (SELECT failed_count FROM verification_stats vs
         WHERE vs.chat_id = g.chat_id AND vs.date = CURRENT_DATE), 0
    ) as today_failed
FROM groups g
LEFT JOIN blacklist bl ON g.chat_id = bl.chat_id
LEFT JOIN bots b ON g.bot_id = b.id
GROUP BY g.id, g.chat_id, g.title, g.member_count, g.is_active, g.last_active_at, g.bot_id, b.name, b.username;

-- Bot stats view
CREATE OR REPLACE VIEW bot_stats AS
SELECT
    b.id,
    b.bot_id,
    b.username,
    b.name,
    b.is_active,
    b.is_primary,
    b.status,
    COUNT(DISTINCT g.id) as group_count,
    COALESCE(
        (SELECT COUNT(*) FROM verification_logs vl
         WHERE vl.bot_id = b.id AND DATE(vl.created_at) = CURRENT_DATE
         AND vl.status = 'success'), 0
    ) as today_verified,
    COALESCE(
        (SELECT COUNT(*) FROM verification_logs vl
         WHERE vl.bot_id = b.id AND DATE(vl.created_at) = CURRENT_DATE
         AND vl.status IN ('failed', 'timeout')), 0
    ) as today_failed
FROM bots b
LEFT JOIN groups g ON b.id = g.bot_id
GROUP BY b.id, b.bot_id, b.username, b.name, b.is_active, b.is_primary, b.status;
