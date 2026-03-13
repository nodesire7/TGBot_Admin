package routes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/models"
)

func GetBlacklist(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	rows, err := db.Query(ctx,
		"SELECT id, chat_id, user_id, username, first_name, reason, banned_by, created_at FROM blacklist WHERE chat_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		chatID, limit, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	entries := make([]models.BlacklistEntry, 0)
	for rows.Next() {
		var e models.BlacklistEntry
		err := rows.Scan(&e.ID, &e.ChatID, &e.UserID, &e.Username, &e.FirstName, &e.Reason, &e.BannedBy, &e.CreatedAt)
		if err != nil {
			continue
		}
		entries = append(entries, e)
	}

	// Get total count
	var total int64
	db.QueryRow(ctx, "SELECT COUNT(*) FROM blacklist WHERE chat_id = $1", chatID).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"blacklist": entries,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

type AddBlacklistRequest struct {
	UserID    int64   `json:"user_id" binding:"required"`
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	Reason    *string `json:"reason"`
}

func AddToBlacklist(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)

	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert into blacklist
	_, err := db.Exec(ctx,
		"INSERT INTO blacklist (chat_id, user_id, username, first_name, reason, created_at) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (chat_id, user_id) DO UPDATE SET reason = $5",
		chatID, req.UserID, req.Username, req.FirstName, req.Reason, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	rdb := config.GetRedis()
	rdb.Del(ctx, "cache:blacklist:"+strconv.FormatInt(chatID, 10))

	c.JSON(http.StatusOK, gin.H{"message": "User added to blacklist"})
}

func RemoveFromBlacklist(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)

	_, err := db.Exec(ctx, "DELETE FROM blacklist WHERE chat_id = $1 AND user_id = $2", chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	rdb := config.GetRedis()
	rdb.Del(ctx, "cache:blacklist:"+strconv.FormatInt(chatID, 10))

	c.JSON(http.StatusOK, gin.H{"message": "User removed from blacklist"})
}
