-- TGBot_Admin Plugin System
-- PostgreSQL 15+

-- ==================== Plugins ====================
-- 插件主表
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) DEFAULT '1.0.0',
    author VARCHAR(255),
    description TEXT,
    main_file TEXT,                    -- 主代码文件内容
    manifest JSONB DEFAULT '{}',       -- 完整 manifest (hooks, config_schema, permissions)
    source VARCHAR(50) DEFAULT 'local', -- official/community/local/github
    github_url VARCHAR(500),           -- GitHub 仓库地址
    is_system BOOLEAN DEFAULT FALSE,   -- 系统插件不可删除
    is_enabled BOOLEAN DEFAULT TRUE,   -- 全局启用状态
    priority INTEGER DEFAULT 0,        -- 执行优先级
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plugins_plugin_id ON plugins(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugins_source ON plugins(source);
CREATE INDEX IF NOT EXISTS idx_plugins_is_enabled ON plugins(is_enabled);

-- ==================== Bot Plugins ====================
-- 每个 Bot 独立的插件配置
CREATE TABLE IF NOT EXISTS bot_plugins (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    plugin_id VARCHAR(100) NOT NULL REFERENCES plugins(plugin_id) ON DELETE CASCADE,
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{}',          -- Bot 级别的插件配置覆盖
    priority INTEGER DEFAULT 0,         -- Bot 内的执行优先级
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(bot_id, plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_bot_plugins_bot_id ON bot_plugins(bot_id);
CREATE INDEX IF NOT EXISTS idx_bot_plugins_enabled ON bot_plugins(is_enabled);

-- ==================== Plugin Versions ====================
-- 插件版本历史
CREATE TABLE IF NOT EXISTS plugin_versions (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) NOT NULL REFERENCES plugins(plugin_id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    code TEXT,                         -- 该版本的代码
    manifest JSONB DEFAULT '{}',
    changelog TEXT,
    published_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(plugin_id, version)
);

CREATE INDEX IF NOT EXISTS idx_plugin_versions_plugin_id ON plugin_versions(plugin_id);

-- ==================== User Plugins ====================
-- 用户在线开发的插件
CREATE TABLE IF NOT EXISTS user_plugins (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    code TEXT,                         -- 用户代码
    manifest JSONB DEFAULT '{}',
    is_testing BOOLEAN DEFAULT FALSE,  -- 是否正在测试
    test_bot_id INTEGER REFERENCES bots(id) ON DELETE SET NULL,
    test_results JSONB,                -- 测试结果
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_plugins_user_id ON user_plugins(user_id);

-- ==================== Plugin Execution Logs ====================
-- 插件执行日志
CREATE TABLE IF NOT EXISTS plugin_logs (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) NOT NULL,
    bot_id INTEGER REFERENCES bots(id) ON DELETE SET NULL,
    chat_id BIGINT,
    event_type VARCHAR(50),            -- on_join, on_message, etc.
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    input_data JSONB,                  -- 输入数据快照
    output_data JSONB,                 -- 输出数据快照
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plugin_logs_plugin_id ON plugin_logs(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_logs_bot_id ON plugin_logs(bot_id);
CREATE INDEX IF NOT EXISTS idx_plugin_logs_created ON plugin_logs(created_at DESC);

-- ==================== Hooks Registry ====================
-- 钩子注册表 (记录哪些插件注册了哪些钩子)
CREATE TABLE IF NOT EXISTS hook_registry (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) NOT NULL REFERENCES plugins(plugin_id) ON DELETE CASCADE,
    hook_name VARCHAR(50) NOT NULL,    -- on_join, on_message, on_command, etc.
    handler_name VARCHAR(100),         -- 处理函数名
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(plugin_id, hook_name, handler_name)
);

CREATE INDEX IF NOT EXISTS idx_hook_registry_hook ON hook_registry(hook_name);
CREATE INDEX IF NOT EXISTS idx_hook_registry_plugin ON hook_registry(plugin_id);

-- ==================== Market Plugins Cache ====================
-- 市场插件缓存 (从远程市场获取的插件信息)
CREATE TABLE IF NOT EXISTS market_plugins (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50),
    author VARCHAR(255),
    description TEXT,
    manifest JSONB DEFAULT '{}',
    github_url VARCHAR(500),
    stars INTEGER DEFAULT 0,
    downloads INTEGER DEFAULT 0,
    category VARCHAR(50),              -- verification, moderation, utility, fun, etc.
    tags JSONB DEFAULT '[]',
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_plugins_category ON market_plugins(category);

-- ==================== Functions ====================

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_plugin_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_plugin_updated_at();

CREATE TRIGGER update_bot_plugins_updated_at BEFORE UPDATE ON bot_plugins
    FOR EACH ROW EXECUTE FUNCTION update_plugin_updated_at();

CREATE TRIGGER update_user_plugins_updated_at BEFORE UPDATE ON user_plugins
    FOR EACH ROW EXECUTE FUNCTION update_plugin_updated_at();

-- ==================== Views ====================

-- Bot 插件统计视图
CREATE OR REPLACE VIEW bot_plugin_stats AS
SELECT
    bp.bot_id,
    b.username as bot_username,
    COUNT(*) FILTER (WHERE bp.is_enabled) as enabled_count,
    COUNT(*) as total_count
FROM bot_plugins bp
JOIN bots b ON bp.bot_id = b.id
GROUP BY bp.bot_id, b.username;

-- 插件使用统计视图
CREATE OR REPLACE VIEW plugin_usage_stats AS
SELECT
    p.plugin_id,
    p.name,
    p.source,
    COUNT(DISTINCT bp.bot_id) as bot_count,
    COUNT(DISTINCT pl.id) FILTER (WHERE pl.created_at > NOW() - INTERVAL '24 hours') as executions_24h,
    AVG(pl.execution_time_ms) FILTER (WHERE pl.created_at > NOW() - INTERVAL '24 hours') as avg_time_24h
FROM plugins p
LEFT JOIN bot_plugins bp ON p.plugin_id = bp.plugin_id
LEFT JOIN plugin_logs pl ON p.plugin_id = pl.plugin_id
GROUP BY p.plugin_id, p.name, p.source;

-- ==================== Default System Plugins ====================

INSERT INTO plugins (plugin_id, name, version, author, description, source, is_system, priority, manifest) VALUES
('arithmetic_verification', '算术验证', '1.0.0', 'TGBot Admin', '新用户入群算术验证，支持多难度级别', 'official', TRUE, 1,
 '{"hooks": ["on_join", "on_callback"], "config_schema": {"type": "object", "properties": {"difficulty": {"type": "string", "enum": ["easy", "medium", "hard"], "default": "easy"}, "timeout": {"type": "integer", "default": 300, "minimum": 60, "maximum": 600}, "max_attempts": {"type": "integer", "default": 3}}, "permissions": ["read_messages", "send_messages", "kick_members", "restrict_members"]}}'::jsonb),

('welcome_message', '入群欢迎', '1.0.0', 'TGBot Admin', '新用户入群发送欢迎消息', 'official', TRUE, 2,
 '{"hooks": ["on_join"], "config_schema": {"type": "object", "properties": {"message": {"type": "string", "default": "欢迎加入本群！"}, "delete_after": {"type": "integer", "default": 0, "description": "多少秒后删除消息，0为不删除"}, "mention_user": {"type": "boolean", "default": true}}, "permissions": ["send_messages", "delete_messages"]}}'::jsonb),

('keyword_filter', '关键词过滤', '1.0.0', 'TGBot Admin', '过滤消息中的敏感关键词', 'official', TRUE, 3,
 '{"hooks": ["on_message"], "config_schema": {"type": "object", "properties": {"keywords": {"type": "array", "items": {"type": "string"}, "default": []}, "action": {"type": "string", "enum": ["delete", "warn", "kick"], "default": "delete"}, "warn_message": {"type": "string", "default": "请勿发送敏感内容"}}, "permissions": ["read_messages", "delete_messages", "kick_members"]}}'::jsonb),

('flood_protection', '防洪水攻击', '1.0.0', 'TGBot Admin', '限制用户发送消息频率，防止刷屏', 'official', TRUE, 4,
 '{"hooks": ["on_message"], "config_schema": {"type": "object", "properties": {"max_messages": {"type": "integer", "default": 5}, "window_seconds": {"type": "integer", "default": 10}, "action": {"type": "string", "enum": ["mute", "kick", "warn"], "default": "mute"}, "mute_duration": {"type": "integer", "default": 300}}, "permissions": ["read_messages", "restrict_members", "kick_members"]}}'::jsonb),

('link_filter', '链接过滤', '1.0.0', 'TGBot Admin', '过滤群组中的链接，支持白名单', 'official', TRUE, 5,
 '{"hooks": ["on_message"], "config_schema": {"type": "object", "properties": {"allow_admins": {"type": "boolean", "default": true}, "whitelist": {"type": "array", "items": {"type": "string"}, "default": []}, "action": {"type": "string", "enum": ["delete", "warn"], "default": "delete"}}, "permissions": ["read_messages", "delete_messages"]}}'::jsonb),

('anti_spam', '反垃圾消息', '1.0.0', 'TGBot Admin', '检测并处理垃圾消息、广告', 'official', TRUE, 6,
 '{"hooks": ["on_message"], "config_schema": {"type": "object", "properties": {"sensitivity": {"type": "string", "enum": ["low", "medium", "high"], "default": "medium"}, "action": {"type": "string", "enum": ["delete", "kick", "report"], "default": "delete"}}, "permissions": ["read_messages", "delete_messages", "kick_members"]}}'::jsonb),

('auto_reply', '自动回复', '1.0.0', 'TGBot Admin', '根据关键词自动回复消息', 'official', FALSE, 10,
 '{"hooks": ["on_message"], "config_schema": {"type": "object", "properties": {"rules": {"type": "array", "items": {"type": "object", "properties": {"trigger": {"type": "string"}, "reply": {"type": "string"}, "is_regex": {"type": "boolean", "default": false}}}}}, "permissions": ["read_messages", "send_messages"]}}'::jsonb),

('stats_reporter', '统计报告', '1.0.0', 'TGBot Admin', '定期生成群组统计报告', 'official', FALSE, 20,
 '{"hooks": ["on_command"], "config_schema": {"type": "object", "properties": {"schedule": {"type": "string", "default": "0 9 * * *"}, "report_channel": {"type": "integer"}}}, "permissions": ["read_messages", "send_messages"]}}'::jsonb)

ON CONFLICT (plugin_id) DO NOTHING;

-- ==================== Grant Permissions ====================
-- 确保 API 用户有权限访问所有新表
