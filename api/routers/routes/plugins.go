package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/models"
)

var defaultPlugins = []models.Plugin{
	{
		PluginID:    "arithmetic_verification",
		Name:        "算术验证",
		Description: strPtr("入群算术题验证，防止机器人加入"),
		IsEnabled:   true,
		Priority:    1,
		Config: map[string]interface{}{
			"default_timeout":    300,
			"default_difficulty": "easy",
		},
	},
	{
		PluginID:    "keyword_filter",
		Name:        "关键词过滤",
		Description: strPtr("自动检测并删除包含敏感关键词的消息"),
		IsEnabled:   false,
		Priority:    2,
		Config: map[string]interface{}{
			"keywords": []string{},
			"action":   "delete",
		},
	},
	{
		PluginID:    "welcome_message",
		Name:        "入群欢迎",
		Description: strPtr("新成员入群时发送欢迎消息"),
		IsEnabled:   false,
		Priority:    3,
		Config: map[string]interface{}{
			"message":       "欢迎加入本群！",
			"delete_after":  0,
		},
	},
	{
		PluginID:    "flood_protection",
		Name:        "防洪水",
		Description: strPtr("限制用户发送消息频率"),
		IsEnabled:   false,
		Priority:    4,
		Config: map[string]interface{}{
			"max_messages": 5,
			"window_seconds": 10,
			"action": "mute",
		},
	},
	{
		PluginID:    "link_filter",
		Name:        "链接过滤",
		Description: strPtr("自动删除包含链接的消息"),
		IsEnabled:   false,
		Priority:    5,
		Config: map[string]interface{}{
			"allow_admins": true,
			"whitelist":    []string{},
		},
	},
}

func strPtr(s string) *string {
	return &s
}

func GetPlugins(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()

	// Check if plugins exist in database
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM plugins").Scan(&count)

	// Initialize default plugins if empty
	if count == 0 {
		for _, p := range defaultPlugins {
			_, err := db.Exec(ctx,
				"INSERT INTO plugins (plugin_id, name, description, is_enabled, priority, config, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				p.PluginID, p.Name, p.Description, p.IsEnabled, p.Priority, p.Config, time.Now(), time.Now(),
			)
			if err != nil {
				continue
			}
		}
	}

	// Fetch plugins
	rows, err := db.Query(ctx,
		"SELECT id, plugin_id, name, description, is_enabled, priority, config, last_restart_at, created_at, updated_at FROM plugins ORDER BY priority",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	plugins := make([]map[string]interface{}, 0)
	for rows.Next() {
		var p models.Plugin
		err := rows.Scan(&p.ID, &p.PluginID, &p.Name, &p.Description, &p.IsEnabled, &p.Priority, &p.Config, &p.LastRestartAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		plugins = append(plugins, gin.H{
			"id":             p.ID,
			"plugin_id":      p.PluginID,
			"name":           p.Name,
			"description":    p.Description,
			"is_enabled":     p.IsEnabled,
			"priority":       p.Priority,
			"config":         p.Config,
			"last_restart_at": p.LastRestartAt,
			"created_at":     p.CreatedAt,
			"updated_at":     p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

type UpdatePluginRequest struct {
	IsEnabled *bool                   `json:"is_enabled"`
	Config    map[string]interface{} `json:"config"`
}

func UpdatePlugin(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	pluginID := c.Param("plugin_id")

	var req UpdatePluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update query
	if req.IsEnabled != nil {
		_, err := db.Exec(ctx,
			"UPDATE plugins SET is_enabled = $1, updated_at = $2 WHERE plugin_id = $3",
			*req.IsEnabled, time.Now(), pluginID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Config != nil {
		_, err := db.Exec(ctx,
			"UPDATE plugins SET config = $1, updated_at = $2 WHERE plugin_id = $3",
			req.Config, time.Now(), pluginID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Invalidate cache
	redis := config.GetRedis()
	redis.Del(ctx, "cache:plugins")

	c.JSON(http.StatusOK, gin.H{"message": "Plugin updated successfully"})
}

func ReloadPlugin(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	redis := config.GetRedis()
	pluginID := c.Param("plugin_id")

	// Update last_restart_at
	_, err := db.Exec(ctx,
		"UPDATE plugins SET last_restart_at = $1 WHERE plugin_id = $2",
		time.Now(), pluginID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send reload signal to bot via Redis
	redis.Publish(ctx, "bot:command", "reload_plugin:"+pluginID)

	c.JSON(http.StatusOK, gin.H{"message": "Reload signal sent"})
}
