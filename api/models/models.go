package models

import (
	"time"
)

// Bot represents a Telegram bot instance
type Bot struct {
	ID           int64              `json:"id"`
	BotID        int64              `json:"bot_id"`        // Telegram bot user ID
	Username     string             `json:"username"`      // @bot_username
	Name         string             `json:"name"`          // Bot display name
	Token        string             `json:"token"`         // Bot API token (masked in responses)
	IsActive     bool               `json:"is_active"`     // Bot enabled/disabled
	IsPrimary    bool               `json:"is_primary"`    // Primary bot flag
	Config       BotConfig          `json:"config"`        // Bot-level configuration
	Status       BotStatus          `json:"status"`        // Runtime status
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	LastActiveAt *time.Time         `json:"last_active_at"`
	GroupCount   int                `json:"group_count,omitempty"`   // From bot_stats view
	TodayVerified int               `json:"today_verified,omitempty"` // From bot_stats view
	TodayFailed   int               `json:"today_failed,omitempty"`  // From bot_stats view
}

// BotConfig stores bot-level settings
type BotConfig struct {
	VerificationTimeout int    `json:"verification_timeout"`
	Difficulty          string `json:"difficulty"`
	AutoApprove         bool   `json:"auto_approve"`
	KickOnFail          bool   `json:"kick_on_fail"`
	MaxFailCount        int    `json:"max_fail_count"`
}

// BotStatus represents current bot runtime status
type BotStatus struct {
	Online     bool   `json:"online"`
	PID        int    `json:"pid"`
	MemoryMB   int    `json:"memory_mb"`
	CPUPercent int    `json:"cpu_percent"`
	StartedAt  int64  `json:"started_at"`
	Error      string `json:"error,omitempty"`
}

// Plugin represents a bot plugin
type Plugin struct {
	ID          int64                  `json:"id"`
	PluginID    string                 `json:"plugin_id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	MainFile    string                 `json:"main_file,omitempty"`
	Manifest    map[string]interface{} `json:"manifest"`
	Source      string                 `json:"source"`      // official/community/local/github
	GitHubURL   string                 `json:"github_url,omitempty"`
	IsSystem    bool                   `json:"is_system"`
	IsEnabled   bool                   `json:"is_enabled"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BotPlugin represents a plugin configuration for a specific bot
type BotPlugin struct {
	ID        int64                  `json:"id"`
	BotID     int64                  `json:"bot_id"`
	PluginID  string                 `json:"plugin_id"`
	IsEnabled bool                   `json:"is_enabled"`
	Config    map[string]interface{} `json:"config"`
	Priority  int                    `json:"priority"`
}

// PluginLog represents a plugin execution log
type PluginLog struct {
	ID              int64                  `json:"id"`
	PluginID        string                 `json:"plugin_id"`
	BotID           *int64                 `json:"bot_id"`
	ChatID          *int64                 `json:"chat_id"`
	EventType       string                 `json:"event_type"`
	ExecutionTimeMs int                    `json:"execution_time_ms"`
	Success         bool                   `json:"success"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	InputData       map[string]interface{} `json:"input_data,omitempty"`
	OutputData      map[string]interface{} `json:"output_data,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// HookRegistry represents a registered hook
type HookRegistry struct {
	ID          int64  `json:"id"`
	PluginID    string `json:"plugin_id"`
	HookName    string `json:"hook_name"`
	HandlerName string `json:"handler_name"`
	Priority    int    `json:"priority"`
	IsActive    bool   `json:"is_active"`
}

// UserPlugin represents a user-developed plugin
type UserPlugin struct {
	ID          int64                  `json:"id"`
	UserID      int64                  `json:"user_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	Manifest    map[string]interface{} `json:"manifest"`
	IsTesting   bool                   `json:"is_testing"`
	TestBotID   *int64                 `json:"test_bot_id"`
	TestResults map[string]interface{} `json:"test_results,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MarketPlugin represents a plugin from the marketplace
type MarketPlugin struct {
	PluginID    string                 `json:"plugin_id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Description string                 `json:"description"`
	Manifest    map[string]interface{} `json:"manifest"`
	GitHubURL   string                 `json:"github_url,omitempty"`
	Stars       int                    `json:"stars"`
	Downloads   int                    `json:"downloads"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
}

// Group represents a Telegram group configuration
type Group struct {
	ID              int64              `json:"id"`
	ChatID          int64              `json:"chat_id"`
	BotID           *int64             `json:"bot_id"`            // Associated bot ID
	BotName         string             `json:"bot_name,omitempty"` // Bot name (from join)
	BotUsername     string             `json:"bot_username,omitempty"` // Bot username (from join)
	Title           string             `json:"title"`
	Username        *string            `json:"username"`
	Description     *string            `json:"description"`
	MemberCount     int                `json:"member_count"`
	IsActive        bool               `json:"is_active"`
	Config          GroupConfig        `json:"config"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	LastActiveAt    *time.Time         `json:"last_active_at"`
	TodayVerified   int                `json:"today_verified"`
	TodayBlocked    int                `json:"today_blocked"`
}

// GroupConfig stores group-specific settings
type GroupConfig struct {
	VerificationTimeout int    `json:"verification_timeout"`
	Difficulty          string `json:"difficulty"`
	AutoApprove         bool   `json:"auto_approve"`
	KickOnFail          bool   `json:"kick_on_fail"`
	MaxFailCount        int    `json:"max_fail_count"`
	AdminWhitelist      []int64 `json:"admin_whitelist"`
}

// BlacklistEntry represents a banned user in a group
type BlacklistEntry struct {
	ID        int64      `json:"id"`
	ChatID    int64      `json:"chat_id"`
	UserID    int64      `json:"user_id"`
	Username  *string    `json:"username"`
	FirstName *string    `json:"first_name"`
	Reason    *string    `json:"reason"`
	BannedBy  *int64     `json:"banned_by"`
	CreatedAt time.Time  `json:"created_at"`
}

// VerificationLog represents a verification attempt
type VerificationLog struct {
	ID              int64      `json:"id"`
	ChatID          int64      `json:"chat_id"`
	BotID           *int64     `json:"bot_id"`        // Associated bot ID
	UserID          int64      `json:"user_id"`
	Username        *string    `json:"username"`
	FirstName       *string    `json:"first_name"`
	Status          string     `json:"status"` // success, failed, timeout
	Question        *string    `json:"question"`
	Answer          *string    `json:"answer"`
	UserAnswer      *string    `json:"user_answer"`
	AttemptCount    int        `json:"attempt_count"`
	DurationSeconds int        `json:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Plugin represents a bot plugin
type Plugin struct {
	ID            int64              `json:"id"`
	PluginID      string             `json:"plugin_id"`
	Name          string             `json:"name"`
	Description   *string            `json:"description"`
	IsEnabled     bool               `json:"is_enabled"`
	Priority      int                `json:"priority"`
	Config        map[string]interface{} `json:"config"`
	LastRestartAt *time.Time         `json:"last_restart_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// Admin represents a dashboard administrator
type Admin struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Username    *string   `json:"username"`
	Role        string    `json:"role"` // super_admin, admin, viewer
	Permissions []string  `json:"permissions"`
	IsActive    bool      `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActionLog represents an action log entry
type ActionLog struct {
	ID          int64                  `json:"id"`
	ChatID      *int64                 `json:"chat_id"`
	UserID      *int64                 `json:"user_id"`
	ActionType  string                 `json:"action_type"`
	ActionData  map[string]interface{} `json:"action_data"`
	OperatorID  *int64                 `json:"operator_id"`
	CreatedAt   time.Time              `json:"created_at"`
}

// DashboardStats aggregates dashboard statistics
type DashboardStats struct {
	TotalBots      int64     `json:"total_bots"`
	ActiveBots     int64     `json:"active_bots"`
	TotalGroups    int64     `json:"total_groups"`
	ActiveGroups   int64     `json:"active_groups"`
	TotalBlocked   int64     `json:"total_blocked"`
	TodayVerified  int64     `json:"today_verified"`
	TodayFailed    int64     `json:"today_failed"`
	TodayKicked    int64     `json:"today_kicked"`
	BotStats       []BotStat `json:"bot_stats"`
}

// BotStat represents individual bot statistics
type BotStat struct {
	BotID        int64     `json:"bot_id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	IsOnline     bool      `json:"is_online"`
	GroupCount   int       `json:"group_count"`
	TodayVerified int      `json:"today_verified"`
	TodayFailed   int      `json:"today_failed"`
}

// BotStatus represents current bot runtime status
type BotStatus struct {
	Online    bool   `json:"online"`
	PID       int    `json:"pid"`
	MemoryMB  int    `json:"memory_mb"`
	CPUPercent int   `json:"cpu_percent"`
	StartedAt int64  `json:"started_at"`
}

// TimelineEvent represents a real-time event
type TimelineEvent struct {
	ID        int64                  `json:"id"`
	Type      string                 `json:"type"`
	ChatID    int64                  `json:"chat_id"`
	ChatTitle string                 `json:"chat_title"`
	UserID    int64                  `json:"user_id"`
	Username  string                 `json:"username"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
