package routes

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
)

// GetSetupStatus returns whether the system has been configured
func GetSetupStatus(c *gin.Context) {
	isConfigured := config.IsConfigured()

	c.JSON(http.StatusOK, gin.H{
		"is_configured": isConfigured,
		"version":       "1.1.1",
	})
}

// GetSetupConfig returns current configuration (partial, for setup wizard)
func GetSetupConfig(c *gin.Context) {
	cfg := config.GetSetupConfig()
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{
			"db_host":       "postgres",
			"db_port":       "5432",
			"db_user":       "tgbot",
			"db_name":       "tgbot",
			"redis_host":    "redis",
			"redis_port":    "6379",
			"is_configured": false,
		})
		return
	}

	// Return config without sensitive data
	c.JSON(http.StatusOK, gin.H{
		"db_host":        cfg.DBHost,
		"db_port":        cfg.DBPort,
		"db_user":        cfg.DBUser,
		"db_name":        cfg.DBName,
		"redis_host":     cfg.RedisHost,
		"redis_port":     cfg.RedisPort,
		"admin_username": cfg.AdminUsername,
		"is_configured":  cfg.IsConfigured,
		"has_bot_token":  cfg.BotToken != "",
	})
}

// TestDatabase tests database connection
func TestDatabase(c *gin.Context) {
	var req config.SetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := config.TestDatabaseConnection(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据库连接成功",
	})
}

// TestRedis tests Redis connection
func TestRedis(c *gin.Context) {
	var req config.SetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := config.TestRedisConnection(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Redis 连接成功",
	})
}

// SaveSetup saves the configuration and initializes connections
func SaveSetup(c *gin.Context) {
	var req config.SetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate required fields
	if req.DBHost == "" || req.DBPort == "" || req.DBUser == "" || req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库配置不完整"})
		return
	}
	if req.RedisHost == "" || req.RedisPort == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Redis 配置不完整"})
		return
	}
	if req.AdminUsername == "" || req.AdminPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员账号不能为空"})
		return
	}
	if req.BotToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bot Token 不能为空"})
		return
	}

	// Generate JWT secret if not provided
	if req.JWTSecret == "" {
		req.JWTSecret = generateRandomSecret()
	}

	// Test connections before saving
	if err := config.TestDatabaseConnection(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "数据库连接失败: " + err.Error(),
		})
		return
	}

	if err := config.TestRedisConnection(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Redis 连接失败: " + err.Error(),
		})
		return
	}

	// Save configuration
	if err := config.SaveSetupConfig(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "保存配置失败: " + err.Error(),
		})
		return
	}

	// Initialize connections
	if err := config.InitDBFromConfig(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "初始化数据库失败: " + err.Error(),
		})
		return
	}

	if err := config.InitRedisFromConfig(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "初始化 Redis 失败: " + err.Error(),
		})
		return
	}

	// Start bot process via supervisorctl
	go startBotProcess()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置保存成功",
	})
}

// startBotProcess starts the bot process using supervisorctl
func startBotProcess() {
	// Try to start bot via supervisorctl
	cmd := exec.Command("supervisorctl", "start", "bot")
	cmd.Env = append(os.Environ(), "SUPERVISOR_SOCKET=/var/run/supervisor.sock")
	cmd.Run()
}

func generateRandomSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
