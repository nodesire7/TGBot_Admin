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

// GetBots returns all bots
func GetBots(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	query := `
		SELECT id, bot_id, username, name, is_active, is_primary, config, status,
			   created_at, updated_at, last_active_at,
			   group_count, today_verified, today_failed
		FROM bot_stats
		ORDER BY is_primary DESC, created_at ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var bots []models.Bot
	for rows.Next() {
		var bot models.Bot
		var configBytes, statusBytes []byte
		var lastActive sql.NullTime
		var username, name sql.NullString

		err := rows.Scan(
			&bot.ID, &bot.BotID, &username, &name, &bot.IsActive, &bot.IsPrimary,
			&configBytes, &statusBytes,
			&bot.CreatedAt, &bot.UpdatedAt, &lastActive,
			&bot.GroupCount, &bot.TodayVerified, &bot.TodayFailed,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if username.Valid {
			bot.Username = username.String
		}
		if name.Valid {
			bot.Name = name.String
		}
		if lastActive.Valid {
			bot.LastActiveAt = &lastActive.Time
		}

		json.Unmarshal(configBytes, &bot.Config)
		json.Unmarshal(statusBytes, &bot.Status)

		bots = append(bots, bot)
	}

	if bots == nil {
		bots = []models.Bot{}
	}

	c.JSON(http.StatusOK, gin.H{"bots": bots})
}

// GetBot returns a single bot by ID
func GetBot(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	query := `
		SELECT id, bot_id, username, name, token, is_active, is_primary, config, status,
			   created_at, updated_at, last_active_at
		FROM bots WHERE id = $1
	`

	var bot models.Bot
	var configBytes, statusBytes []byte
	var lastActive sql.NullTime
	var username, name, token sql.NullString

	err = db.QueryRow(context.Background(), query, botID).Scan(
		&bot.ID, &bot.BotID, &username, &name, &token, &bot.IsActive, &bot.IsPrimary,
		&configBytes, &statusBytes,
		&bot.CreatedAt, &bot.UpdatedAt, &lastActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if username.Valid {
		bot.Username = username.String
	}
	if name.Valid {
		bot.Name = name.String
	}
	if token.Valid {
		bot.Token = maskToken(token.String)
	}
	if lastActive.Valid {
		bot.LastActiveAt = &lastActive.Time
	}

	json.Unmarshal(configBytes, &bot.Config)
	json.Unmarshal(statusBytes, &bot.Status)

	c.JSON(http.StatusOK, bot)
}

// CreateBotRequest represents the request body for creating a bot
type CreateBotRequest struct {
	Token     string         `json:"token" binding:"required"`
	Name      string         `json:"name"`
	IsPrimary bool           `json:"is_primary"`
	Config    models.BotConfig `json:"config"`
}

// CreateBot adds a new bot
func CreateBot(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	var req CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate token by making a test call to Telegram API
	botInfo, err := validateBotToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot token: " + err.Error()})
		return
	}

	// Set default config if not provided
	if req.Config.VerificationTimeout == 0 {
		req.Config.VerificationTimeout = 300
	}
	if req.Config.Difficulty == "" {
		req.Config.Difficulty = "easy"
	}
	if req.Config.MaxFailCount == 0 {
		req.Config.MaxFailCount = 3
	}

	configBytes, _ := json.Marshal(req.Config)
	statusBytes, _ := json.Marshal(models.BotStatus{Online: false})

	// If this is primary, unset other primary bots
	if req.IsPrimary {
		db.Exec(context.Background(), "UPDATE bots SET is_primary = false")
	}

	query := `
		INSERT INTO bots (bot_id, username, name, token, is_active, is_primary, config, status)
		VALUES ($1, $2, $3, $4, true, $5, $6, $7)
		RETURNING id, created_at
	`

	var id int64
	var createdAt time.Time
	err = db.QueryRow(context.Background(), query,
		botInfo.ID, botInfo.Username, req.Name, req.Token,
		req.IsPrimary, configBytes, statusBytes,
	).Scan(&id, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"bot_id":   botInfo.ID,
		"username": botInfo.Username,
		"name":     req.Name,
		"message":  "Bot created successfully",
	})
}

// UpdateBotRequest represents the request body for updating a bot
type UpdateBotRequest struct {
	Name      *string         `json:"name"`
	IsActive  *bool           `json:"is_active"`
	IsPrimary *bool           `json:"is_primary"`
	Config    *models.BotConfig `json:"config"`
	Token     *string         `json:"token"`
}

// UpdateBot updates a bot's settings
func UpdateBot(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	var req UpdateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If setting as primary, unset other primary bots
	if req.IsPrimary != nil && *req.IsPrimary {
		db.Exec(context.Background(), "UPDATE bots SET is_primary = false")
	}

	// Build dynamic update query
	query := "UPDATE bots SET updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		query += ", name = $" + strconv.Itoa(argIdx)
		args = append(args, *req.Name)
		argIdx++
	}
	if req.IsActive != nil {
		query += ", is_active = $" + strconv.Itoa(argIdx)
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.IsPrimary != nil {
		query += ", is_primary = $" + strconv.Itoa(argIdx)
		args = append(args, *req.IsPrimary)
		argIdx++
	}
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		query += ", config = $" + strconv.Itoa(argIdx)
		args = append(args, configBytes)
		argIdx++
	}
	if req.Token != nil {
		// Validate new token
		botInfo, err := validateBotToken(*req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot token: " + err.Error()})
			return
		}
		query += ", token = $" + strconv.Itoa(argIdx) + ", bot_id = $" + strconv.Itoa(argIdx+1) + ", username = $" + strconv.Itoa(argIdx+2)
		args = append(args, *req.Token, botInfo.ID, botInfo.Username)
		argIdx += 3
	}

	query += " WHERE id = $" + strconv.Itoa(argIdx)
	args = append(args, botID)

	result, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bot updated successfully"})
}

// DeleteBot removes a bot
func DeleteBot(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
		return
	}

	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	result, err := db.Exec(context.Background(), "DELETE FROM bots WHERE id = $1", botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bot deleted successfully"})
}

// StartBot starts a bot instance
func StartBot(c *gin.Context) {
	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	// TODO: Implement actual bot start logic
	// This would involve signaling the bot manager to start a new bot process

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot start signal sent",
		"bot_id":  botID,
	})
}

// StopBot stops a bot instance
func StopBot(c *gin.Context) {
	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	// TODO: Implement actual bot stop logic

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot stop signal sent",
		"bot_id":  botID,
	})
}

// RestartBot restarts a bot instance
func RestartBot(c *gin.Context) {
	id := c.Param("id")
	botID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot ID"})
		return
	}

	// TODO: Implement actual bot restart logic

	c.JSON(http.StatusOK, gin.H{
		"message": "Bot restart signal sent",
		"bot_id":  botID,
	})
}

// TestBotToken validates a bot token
func TestBotToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token required"})
		return
	}

	botInfo, err := validateBotToken(req.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Token 验证成功",
		"bot_id":   botInfo.ID,
		"username": botInfo.Username,
		"name":     botInfo.FirstName,
	})
}

// BotInfo contains basic bot information from Telegram API
type BotInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// validateBotToken makes a test call to Telegram API to validate the token
func validateBotToken(token string) (*BotInfo, error) {
	// Use Telegram's getMe API to validate token
	resp, err := http.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool     `json:"ok"`
		Result *BotInfo `json:"result"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Ok {
		if result.Description != "" {
			return nil, &BotTokenError{Message: result.Description}
		}
		return nil, &BotTokenError{Message: "Invalid bot token"}
	}

	return result.Result, nil
}

// BotTokenError represents a bot token validation error
type BotTokenError struct {
	Message string
}

func (e *BotTokenError) Error() string {
	return e.Message
}

// maskToken masks a bot token for display
func maskToken(token string) string {
	if len(token) < 10 {
		return "***"
	}
	return token[:10] + "***" + token[len(token)-3:]
}
