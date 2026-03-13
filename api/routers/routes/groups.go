package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/models"
)

func GetGroups(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	rdb := config.GetRedis()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filter
	isActive := c.Query("is_active")
	search := c.Query("search")

	// Build query
	query := "SELECT id, chat_id, title, username, member_count, is_active, config, created_at, updated_at, last_active_at FROM groups WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM groups WHERE 1=1"
	args := make([]interface{}, 0)
	argPos := 1

	if isActive != "" {
		query += " AND is_active = $" + strconv.Itoa(argPos)
		countQuery += " AND is_active = $" + strconv.Itoa(argPos)
		args = append(args, isActive == "true")
		argPos++
	}

	if search != "" {
		query += " AND (title ILIKE $" + strconv.Itoa(argPos) + " OR username ILIKE $" + strconv.Itoa(argPos) + ")"
		countQuery += " AND (title ILIKE $" + strconv.Itoa(argPos) + " OR username ILIKE $" + strconv.Itoa(argPos) + ")"
		args = append(args, "%"+search+"%")
		argPos++
	}

	// Get total count
	var total int64
	db.QueryRow(ctx, countQuery, args...).Scan(&total)

	// Add pagination
	query += " ORDER BY last_active_at DESC NULLS LAST LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	groups := make([]map[string]interface{}, 0)
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.ID, &g.ChatID, &g.Title, &g.Username, &g.MemberCount, &g.IsActive, &g.Config, &g.CreatedAt, &g.UpdatedAt, &g.LastActiveAt)
		if err != nil {
			continue
		}

		// Get today's stats from Redis
		today := time.Now().Format("2006-01-02")
		todayVerified, _ := rdb.Get(ctx, "stats:group:"+strconv.FormatInt(g.ChatID, 10)+":"+today+":verified").Int64()
		todayBlocked, _ := rdb.Get(ctx, "stats:group:"+strconv.FormatInt(g.ChatID, 10)+":"+today+":blocked").Int64()

		groups = append(groups, gin.H{
			"id":             g.ID,
			"chat_id":        g.ChatID,
			"title":          g.Title,
			"username":       g.Username,
			"member_count":   g.MemberCount,
			"is_active":      g.IsActive,
			"config":         g.Config,
			"created_at":     g.CreatedAt,
			"updated_at":     g.UpdatedAt,
			"last_active_at": g.LastActiveAt,
			"today_verified": todayVerified,
			"today_blocked":  todayBlocked,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func GetGroup(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)

	var g models.Group
	err := db.QueryRow(ctx,
		"SELECT id, chat_id, title, username, description, member_count, is_active, config, created_at, updated_at, last_active_at FROM groups WHERE chat_id = $1",
		chatID,
	).Scan(&g.ID, &g.ChatID, &g.Title, &g.Username, &g.Description, &g.MemberCount, &g.IsActive, &g.Config, &g.CreatedAt, &g.UpdatedAt, &g.LastActiveAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, g)
}

type UpdateGroupRequest struct {
	Title       *string         `json:"title"`
	IsActive    *bool           `json:"is_active"`
	Config      *models.GroupConfig `json:"config"`
}

func UpdateGroup(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	rdb := config.GetRedis()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update query dynamically
	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argPos := 1

	if req.Title != nil {
		updates = append(updates, "title = $"+strconv.Itoa(argPos))
		args = append(args, *req.Title)
		argPos++
	}

	if req.IsActive != nil {
		updates = append(updates, "is_active = $"+strconv.Itoa(argPos))
		args = append(args, *req.IsActive)
		argPos++
	}

	if req.Config != nil {
		updates = append(updates, "config = $"+strconv.Itoa(argPos))
		args = append(args, req.Config)
		argPos++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = $"+strconv.Itoa(argPos))
	args = append(args, time.Now())
	argPos++

	args = append(args, chatID)
	query := "UPDATE groups SET " + joinUpdates(updates) + " WHERE chat_id = $" + strconv.Itoa(argPos)

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	rdb.Del(ctx, "cache:group:"+strconv.FormatInt(chatID, 10))

	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

func DeleteGroup(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	rdb := config.GetRedis()
	chatID, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)

	_, err := db.Exec(ctx, "DELETE FROM groups WHERE chat_id = $1", chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	rdb.Del(ctx, "cache:group:"+strconv.FormatInt(chatID, 10))

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func SyncGroup(c *gin.Context) {
	// TODO: Send sync signal to bot via Redis
	chatID := c.Param("chat_id")
	rdb := config.GetRedis()
	ctx := c.Request.Context()

	rdb.Publish(ctx, "bot:command", "sync_group:"+chatID)

	c.JSON(http.StatusOK, gin.H{"message": "Sync signal sent"})
}

func joinUpdates(updates []string) string {
	result := ""
	for i, u := range updates {
		if i > 0 {
			result += ", "
		}
		result += u
	}
	return result
}
