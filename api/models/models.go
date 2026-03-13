package models

import (
	"time"
)

// Group represents a Telegram group configuration
type Group struct {
	ID              int64              `json:"id"`
	ChatID          int64              `json:"chat_id"`
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
	ID             int64      `json:"id"`
	ChatID         int64      `json:"chat_id"`
	UserID         int64      `json:"user_id"`
	Username       *string    `json:"username"`
	FirstName      *string    `json:"first_name"`
	Status         string     `json:"status"` // success, failed, timeout
	Question       *string    `json:"question"`
	Answer         *string    `json:"answer"`
	UserAnswer     *string    `json:"user_answer"`
	AttemptCount   int        `json:"attempt_count"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAt      time.Time  `json:"created_at"`
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
	BotStatus      BotStatus `json:"bot_status"`
	TotalGroups    int64     `json:"total_groups"`
	ActiveGroups   int64     `json:"active_groups"`
	TotalBlocked   int64     `json:"total_blocked"`
	TodayVerified  int64     `json:"today_verified"`
	TodayFailed    int64     `json:"today_failed"`
	TodayKicked    int64     `json:"today_kicked"`
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
