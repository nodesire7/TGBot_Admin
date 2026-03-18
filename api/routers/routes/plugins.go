package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
)

var defaultPlugins = []map[string]interface{}{
	{
		"plugin_id":  "arithmetic_verification",
		"name":       "算术验证",
		"description": "入群算术题验证，防止机器人加入",
		"is_enabled": true,
		"priority":    1,
		"config": map[string]interface{}{
			"default_timeout":    300,
			"default_difficulty": "easy",
		},
	},
	{
		"plugin_id":  "keyword_filter",
		"name":       "关键词过滤",
		"description": "自动检测并删除包含敏感关键词的消息",
		"is_enabled": false,
		"priority":    2,
		"config": map[string]interface{}{
			"keywords": []string{},
			"action":   "delete",
		},
	},
	{
		"plugin_id":  "welcome_message",
		"name":       "入群欢迎",
		"description": "新成员入群时发送欢迎消息",
		"is_enabled": false,
		"priority":    3,
		"config": map[string]interface{}{
			"message":       "欢迎加入本群！",
			"delete_after":  0,
		},
	},
	{
		"plugin_id":  "flood_protection",
		"name":       "防洪水",
		"description": "限制用户发送消息频率",
		"is_enabled": false,
		"priority":    4,
		"config": map[string]interface{}{
			"max_messages":   5,
			"window_seconds": 10,
			"action":         "mute",
		},
	},
	{
		"plugin_id":  "link_filter",
		"name":       "链接过滤",
		"description": "自动删除包含链接的消息",
		"is_enabled": false,
		"priority":    5,
		"config": map[string]interface{}{
			"allow_admins": true,
			"whitelist":    []string{},
		},
	},
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
				p["plugin_id"], p["name"], p["description"], p["is_enabled"], p["priority"], p["config"], time.Now(), time.Now(),
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
		var id int64
		var pluginID, name, description string
		var isEnabled bool
		var priority int
		var config map[string]interface{}
		var lastRestartAt *time.Time
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &pluginID, &name, &description, &isEnabled, &priority, &config, &lastRestartAt, &createdAt, &updatedAt)
		if err != nil {
			continue
		}
		plugins = append(plugins, gin.H{
			"id":              id,
			"plugin_id":       pluginID,
			"name":            name,
			"description":     description,
			"is_enabled":      isEnabled,
			"priority":        priority,
			"config":          config,
			"last_restart_at": lastRestartAt,
			"created_at":      createdAt,
			"updated_at":      updatedAt,
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
	rdb := config.GetRedis()
	rdb.Del(ctx, "cache:plugins")

	c.JSON(http.StatusOK, gin.H{"message": "Plugin updated successfully"})
}

func ReloadPlugin(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	rdb := config.GetRedis()
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
	rdb.Publish(ctx, "bot:command", "reload_plugin:"+pluginID)

	c.JSON(http.StatusOK, gin.H{"message": "Reload signal sent"})
}

func GetPlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func InstallPlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func UninstallPlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func EnablePlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func DisablePlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func TestPlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func GetPluginLogs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func GetBotPlugins(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func UpdateBotPlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
