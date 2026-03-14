package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/models"
)

// GetPlugins returns all installed plugins
func GetPlugins(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	query := `
		SELECT p.id, p.plugin_id, p.name, p.version, p.author, p.description,
			   p.source, p.is_system, p.is_enabled, p.priority, p.manifest,
			   COALESCE(stats.bot_count, 0) as bot_count,
			   COALESCE(stats.executions_24h, 0) as executions_24h
		FROM plugins p
		LEFT JOIN plugin_usage_stats stats ON p.plugin_id = stats.plugin_id
		ORDER BY p.priority ASC, p.created_at ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var plugins []map[string]interface{}
	for rows.Next() {
		var manifestBytes []byte
		plugin := make(map[string]interface{})
		err := rows.Scan(
			&plugin["id"], &plugin["plugin_id"], &plugin["name"], &plugin["version"],
			&plugin["author"], &plugin["description"], &plugin["source"],
			&plugin["is_system"], &plugin["is_enabled"], &plugin["priority"],
			&manifestBytes,
			&plugin["bot_count"], &plugin["executions_24h"],
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var manifest map[string]interface{}
		json.Unmarshal(manifestBytes, &manifest)
		plugin["manifest"] = manifest
		plugin["config_schema"] = manifest["config_schema"]
		plugin["hooks"] = manifest["hooks"]
		plugin["permissions"] = manifest["permissions"]

		plugins = append(plugins, plugin)
	}

	if plugins == nil {
		plugins = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}

// GetPlugin returns a single plugin by ID
func GetPlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")

	query := `
		SELECT id, plugin_id, name, version, author, description,
			   main_file, source, github_url, is_system, is_enabled, priority, manifest
		FROM plugins WHERE plugin_id = $1
	`

	var plugin models.Plugin
	var manifestBytes []byte
	var mainFile, githubURL, author sql.NullString

	err := db.QueryRow(context.Background(), query, pluginID).Scan(
		&plugin.ID, &plugin.PluginID, &plugin.Name, &plugin.Version,
		&author, &plugin.Description, &mainFile,
		&plugin.Source, &githubURL, &plugin.IsSystem, &plugin.IsEnabled,
		&plugin.Priority, &manifestBytes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal(manifestBytes, &plugin.Manifest)

	if author.Valid {
		plugin.Author = author.String
	}
	if mainFile.Valid {
		plugin.MainFile = mainFile.String
	}
	if githubURL.Valid {
		plugin.GitHubURL = githubURL.String
	}

	c.JSON(http.StatusOK, plugin)
}

// InstallPluginRequest represents the request body for installing a plugin
type InstallPluginRequest struct {
	Source     string `json:"source"`     // official, github, local
	PluginID   string `json:"plugin_id"`  // For github/official
	GitHubURL  string `json:"github_url"` // For github source
	Code       string `json:"code"`       // For local source
	Manifest   string `json:"manifest"`   // For local source
}

// InstallPlugin installs a new plugin
func InstallPlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var req InstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle different installation sources
	switch req.Source {
	case "github":
		// TODO: Fetch plugin from GitHub
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub installation not yet implemented"})
		return

	case "local":
		// Validate manifest
		if req.Manifest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Manifest is required for local plugins"})
			return
		}

		var manifest map[string]interface{}
		if err := json.Unmarshal([]byte(req.Manifest), &manifest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manifest JSON: " + err.Error()})
			return
		}

		pluginID, ok := manifest["id"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID is required in manifest"})
			return
		}

		name, _ := manifest["name"].(string)
		version, _ := manifest["version"].(string)
		author, _ := manifest["author"].(string)

		// Check if plugin already exists
		var exists bool
		err := db.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM plugins WHERE plugin_id = $1)", pluginID).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if exists {
			// Update existing plugin
			_, err = db.Exec(context.Background(), `
				UPDATE plugins SET
					name = $2, version = $3, author = $4,
					main_file = $5, manifest = $6, source = 'local',
					updated_at = NOW()
				WHERE plugin_id = $1
			`, pluginID, name, version, author, req.Code, req.Manifest)
		} else {
			// Insert new plugin
			_, err = db.Exec(context.Background(), `
				INSERT INTO plugins (plugin_id, name, version, author, main_file, manifest, source)
				VALUES ($1, $2, $3, $4, $5, $6, 'local')
			`, pluginID, name, version, author, req.Code, req.Manifest)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Plugin installed successfully",
			"plugin_id": pluginID,
		})
		return

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source type"})
		return
	}
}

// UpdatePluginConfigRequest represents the request body for updating plugin config
type UpdatePluginConfigRequest struct {
	IsEnabled *bool                  `json:"is_enabled"`
	Priority  *int                   `json:"priority"`
	Config    *map[string]interface{} `json:"config"`
}

// UpdatePlugin updates plugin settings
func UpdatePlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")

	var req UpdatePluginConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE plugins SET updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.IsEnabled != nil {
		query += ", is_enabled = $" + strconv.Itoa(argIdx)
		args = append(args, *req.IsEnabled)
		argIdx++
	}
	if req.Priority != nil {
		query += ", priority = $" + strconv.Itoa(argIdx)
		args = append(args, *req.Priority)
		argIdx++
	}

	query += " WHERE plugin_id = $" + strconv.Itoa(argIdx)
	args = append(args, pluginID)

	result, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin updated successfully"})
}

// UninstallPlugin removes a plugin
func UninstallPlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")

	// Check if it's a system plugin
	var isSystem bool
	err := db.QueryRow(context.Background(),
		"SELECT is_system FROM plugins WHERE plugin_id = $1", pluginID).Scan(&isSystem)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isSystem {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot uninstall system plugin"})
		return
	}

	result, err := db.Exec(context.Background(), "DELETE FROM plugins WHERE plugin_id = $1", pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin uninstalled successfully"})
}

// EnablePlugin enables a plugin
func EnablePlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")

	result, err := db.Exec(context.Background(),
		"UPDATE plugins SET is_enabled = true, updated_at = NOW() WHERE plugin_id = $1",
		pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Notify bot via Redis
	notifyPluginChange(pluginID, "enable")

	c.JSON(http.StatusOK, gin.H{"message": "Plugin enabled successfully"})
}

// DisablePlugin disables a plugin
func DisablePlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")

	result, err := db.Exec(context.Background(),
		"UPDATE plugins SET is_enabled = false, updated_at = NOW() WHERE plugin_id = $1",
		pluginID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	notifyPluginChange(pluginID, "disable")

	c.JSON(http.StatusOK, gin.H{"message": "Plugin disabled successfully"})
}

// TestPlugin runs a plugin in sandbox mode
func TestPlugin(c *gin.Context) {
	pluginID := c.Param("id")

	// TODO: Implement sandbox testing
	// This would involve:
	// 1. Loading plugin code
	// 2. Creating sandbox environment
	// 3. Simulating events
	// 4. Collecting output

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin test initiated",
		"plugin_id": pluginID,
		"status": "running",
	})
}

// GetBotPlugins returns plugin configuration for a specific bot
func GetBotPlugins(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	botID := c.Param("bot_id")

	query := `
		SELECT bp.id, bp.plugin_id, p.name, bp.is_enabled, bp.config, bp.priority
		FROM bot_plugins bp
		JOIN plugins p ON bp.plugin_id = p.plugin_id
		WHERE bp.bot_id = $1
		ORDER BY bp.priority ASC
	`

	rows, err := db.Query(context.Background(), query, botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var botPlugins []map[string]interface{}
	for rows.Next() {
		var id int
		var pluginID, name string
		var isEnabled bool
		var configBytes []byte
		var priority int

		err := rows.Scan(&id, &pluginID, &name, &isEnabled, &configBytes, &priority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var cfg map[string]interface{}
		json.Unmarshal(configBytes, &cfg)

		botPlugins = append(botPlugins, map[string]interface{}{
			"id":        id,
			"plugin_id": pluginID,
			"name":      name,
			"is_enabled": isEnabled,
			"config":     cfg,
			"priority":   priority,
		})
	}

	if botPlugins == nil {
		botPlugins = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"bot_plugins": botPlugins})
}

// UpdateBotPluginRequest represents the request body for updating bot plugin config
type UpdateBotPluginRequest struct {
	IsEnabled *bool                  `json:"is_enabled"`
	Config    *map[string]interface{} `json:"config"`
	Priority  *int                   `json:"priority"`
}

// UpdateBotPlugin updates plugin configuration for a specific bot
func UpdateBotPlugin(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	botID := c.Param("bot_id")
	pluginID := c.Param("plugin_id")

	var req UpdateBotPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if bot_plugin exists, if not create it
	var exists bool
	err := db.QueryRow(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM bot_plugins WHERE bot_id = $1 AND plugin_id = $2)
	`, botID, pluginID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !exists {
		// Create new bot_plugin entry
		_, err = db.Exec(context.Background(), `
			INSERT INTO bot_plugins (bot_id, plugin_id, is_enabled, config)
			VALUES ($1, $2, true, '{}')
		`, botID, pluginID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Build update query
	query := "UPDATE bot_plugins SET updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.IsEnabled != nil {
		query += ", is_enabled = $" + strconv.Itoa(argIdx)
		args = append(args, *req.IsEnabled)
		argIdx++
	}
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		query += ", config = $" + strconv.Itoa(argIdx)
		args = append(args, configBytes)
		argIdx++
	}
	if req.Priority != nil {
		query += ", priority = $" + strconv.Itoa(argIdx)
		args = append(args, *req.Priority)
		argIdx++
	}

	query += " WHERE bot_id = $" + strconv.Itoa(argIdx) + " AND plugin_id = $" + strconv.Itoa(argIdx+1)
	args = append(args, botID, pluginID)

	_, err = db.Exec(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bot plugin updated successfully"})
}

// GetPluginLogs returns execution logs for a plugin
func GetPluginLogs(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	pluginID := c.Param("id")
	limit := c.DefaultQuery("limit", "100")

	query := `
		SELECT id, bot_id, chat_id, event_type, execution_time_ms,
			   success, error_message, created_at
		FROM plugin_logs
		WHERE plugin_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := db.Query(context.Background(), query, pluginID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var log map[string]interface{} = make(map[string]interface{})
		var errorMsg sql.NullString

		err := rows.Scan(
			&log["id"], &log["bot_id"], &log["chat_id"], &log["event_type"],
			&log["execution_time_ms"], &log["success"], &errorMsg, &log["created_at"],
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if errorMsg.Valid {
			log["error_message"] = errorMsg.String
		}
		logs = append(logs, log)
	}

	if logs == nil {
		logs = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// notifyPluginChange sends a notification to the bot via Redis
func notifyPluginChange(pluginID, action string) {
	redisClient := config.GetRedis()
	if redisClient == nil {
		return
	}

	ctx := context.Background()
	message := map[string]interface{}{
		"type":      "plugin_change",
		"plugin_id": pluginID,
		"action":    action,
		"timestamp": time.Now().Unix(),
	}

	msgBytes, _ := json.Marshal(message)
	redisClient.Publish(ctx, "bot:command", string(msgBytes))
}
