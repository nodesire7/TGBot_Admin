package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/middleware"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin credentials from setup config
	setupCfg := config.GetSetupConfig()
	adminUsername := "admin"
	adminPassword := "admin123"

	if setupCfg != nil && setupCfg.IsConfigured {
		adminUsername = setupCfg.AdminUsername
		adminPassword = setupCfg.AdminPassword
	}

	if req.Username != adminUsername || req.Password != adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// Generate JWT token
	userID := int64(1)
	role := "super_admin"
	token, err := middleware.GenerateToken(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// Store session in Redis (if available)
	redisClient := config.GetRedis()
	if redisClient != nil {
		ctx := c.Request.Context()
		sessionKey := "session:" + token
		redisClient.HSet(ctx, sessionKey, "user_id", userID, "role", role)
		redisClient.Expire(ctx, sessionKey, 24*time.Hour)
	}

	resp := LoginResponse{
		Token: token,
	}
	resp.User.ID = userID
	resp.User.Username = req.Username
	resp.User.Role = role

	c.JSON(http.StatusOK, resp)
}

func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	token := authHeader[7:] // Remove "Bearer " prefix
	redisClient := config.GetRedis()
	redisClient.Del(c.Request.Context(), "session:"+token)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func RefreshToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	token, err := middleware.GenerateToken(userID.(int64), role.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	// Invalidate old token
	oldToken := c.GetHeader("Authorization")[7:]
	redisClient := config.GetRedis()
	redisClient.Del(c.Request.Context(), "session:"+oldToken)

	// Store new session
	sessionKey := "session:" + token
	redisClient.HSet(c.Request.Context(), sessionKey, "user_id", userID, "role", role)
	redisClient.Expire(c.Request.Context(), sessionKey, 24*time.Hour)

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       userID,
			"username": "admin",
			"role":     role,
		},
	})
}
