package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
)

func GetVerificationLogs(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Filters
	chatID := c.Query("chat_id")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Build query
	query := "SELECT id, chat_id, user_id, username, first_name, status, question, answer, user_answer, attempt_count, duration_seconds, created_at FROM verification_logs WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM verification_logs WHERE 1=1"
	args := make([]interface{}, 0)
	argPos := 1

	if chatID != "" {
		query += " AND chat_id = $" + strconv.Itoa(argPos)
		countQuery += " AND chat_id = $" + strconv.Itoa(argPos)
		args = append(args, chatID)
		argPos++
	}

	if status != "" {
		query += " AND status = $" + strconv.Itoa(argPos)
		countQuery += " AND status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}

	if startDate != "" {
		query += " AND created_at >= $" + strconv.Itoa(argPos)
		countQuery += " AND created_at >= $" + strconv.Itoa(argPos)
		args = append(args, startDate+" 00:00:00")
		argPos++
	}

	if endDate != "" {
		query += " AND created_at <= $" + strconv.Itoa(argPos)
		countQuery += " AND created_at <= $" + strconv.Itoa(argPos)
		args = append(args, endDate+" 23:59:59")
		argPos++
	}

	// Get total count
	var total int64
	db.QueryRow(ctx, countQuery, args...).Scan(&total)

	// Add pagination
	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var log struct {
			ID              int64
			ChatID          int64
			UserID          int64
			Username        *string
			FirstName       *string
			Status          string
			Question        *string
			Answer          *string
			UserAnswer      *string
			AttemptCount    int
			DurationSeconds int
			CreatedAt       time.Time
		}
		err := rows.Scan(&log.ID, &log.ChatID, &log.UserID, &log.Username, &log.FirstName, &log.Status, &log.Question, &log.Answer, &log.UserAnswer, &log.AttemptCount, &log.DurationSeconds, &log.CreatedAt)
		if err != nil {
			continue
		}
		logs = append(logs, gin.H{
			"id":               log.ID,
			"chat_id":          log.ChatID,
			"user_id":          log.UserID,
			"username":         log.Username,
			"first_name":       log.FirstName,
			"status":           log.Status,
			"question":         log.Question,
			"answer":           log.Answer,
			"user_answer":      log.UserAnswer,
			"attempt_count":    log.AttemptCount,
			"duration_seconds": log.DurationSeconds,
			"created_at":       log.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func GetActionLogs(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Filters
	actionType := c.Query("action_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Build query
	query := "SELECT id, chat_id, user_id, action_type, action_data, operator_id, created_at FROM action_logs WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM action_logs WHERE 1=1"
	args := make([]interface{}, 0)
	argPos := 1

	if actionType != "" {
		query += " AND action_type = $" + strconv.Itoa(argPos)
		countQuery += " AND action_type = $" + strconv.Itoa(argPos)
		args = append(args, actionType)
		argPos++
	}

	if startDate != "" {
		query += " AND created_at >= $" + strconv.Itoa(argPos)
		countQuery += " AND created_at >= $" + strconv.Itoa(argPos)
		args = append(args, startDate+" 00:00:00")
		argPos++
	}

	if endDate != "" {
		query += " AND created_at <= $" + strconv.Itoa(argPos)
		countQuery += " AND created_at <= $" + strconv.Itoa(argPos)
		args = append(args, endDate+" 23:59:59")
		argPos++
	}

	// Get total count
	var total int64
	db.QueryRow(ctx, countQuery, args...).Scan(&total)

	// Add pagination
	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var log struct {
			ID         int64
			ChatID     *int64
			UserID     *int64
			ActionType string
			ActionData map[string]interface{}
			OperatorID *int64
			CreatedAt  time.Time
		}
		err := rows.Scan(&log.ID, &log.ChatID, &log.UserID, &log.ActionType, &log.ActionData, &log.OperatorID, &log.CreatedAt)
		if err != nil {
			continue
		}
		logs = append(logs, gin.H{
			"id":           log.ID,
			"chat_id":      log.ChatID,
			"user_id":      log.UserID,
			"action_type":  log.ActionType,
			"action_data":  log.ActionData,
			"operator_id":  log.OperatorID,
			"created_at":   log.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
