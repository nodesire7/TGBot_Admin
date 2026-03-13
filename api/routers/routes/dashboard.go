package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
)

func GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()
	db := config.GetDB()
	redis := config.GetRedis()

	// Get bot status from Redis
	botStatus := getBotStatus(redis)

	// Get group counts
	var totalGroups, activeGroups int64
	db.QueryRow(ctx, "SELECT COUNT(*) FROM groups").Scan(&totalGroups)
	db.QueryRow(ctx, "SELECT COUNT(*) FROM groups WHERE is_active = true").Scan(&activeGroups)

	// Get total blocked
	var totalBlocked int64
	db.QueryRow(ctx, "SELECT COUNT(*) FROM blacklist").Scan(&totalBlocked)

	// Get today's stats from Redis
	today := time.Now().Format("2006-01-02")
	todayVerified, _ := redis.Get(ctx, "stats:"+today+":verified").Int64()
	todayFailed, _ := redis.Get(ctx, "stats:"+today+":failed").Int64()
	todayKicked, _ := redis.Get(ctx, "stats:"+today+":kicked").Int64()

	c.JSON(http.StatusOK, gin.H{
		"bot_status": gin.H{
			"online":     botStatus.Online,
			"pid":        botStatus.PID,
			"memory_mb":  botStatus.MemoryMB,
			"cpu_percent": botStatus.CPUPercent,
			"started_at": botStatus.StartedAt,
		},
		"total_groups":   totalGroups,
		"active_groups":  activeGroups,
		"total_blocked":  totalBlocked,
		"today_verified": todayVerified,
		"today_failed":   todayFailed,
		"today_kicked":   todayKicked,
	})
}

func GetTimeline(c *gin.Context) {
	ctx := c.Request.Context()
	redis := config.GetRedis()

	// Get recent events from Redis Stream
	streams, err := redis.XRevRange(ctx, "stream:events", "+", "-").Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"events": []interface{}{}})
		return
	}

	events := make([]map[string]interface{}, 0, len(streams))
	for i, stream := range streams {
		if i >= 50 { // Limit to 50 events
			break
		}
		event := map[string]interface{}{
			"id":        stream.ID,
			"type":      stream.Values["type"],
			"chat_id":   stream.Values["chat_id"],
			"chat_title": stream.Values["chat_title"],
			"user_id":   stream.Values["user_id"],
			"username":  stream.Values["username"],
			"timestamp": stream.Values["timestamp"],
		}
		events = append(events, event)
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

type BotStatusData struct {
	Online    bool   `json:"online"`
	PID       int    `json:"pid"`
	MemoryMB  int    `json:"memory_mb"`
	CPUPercent int   `json:"cpu_percent"`
	StartedAt int64  `json:"started_at"`
}

func getBotStatus(redis *redis.Client) BotStatusData {
	ctx := context.Background()
	status := BotStatusData{}

	online, _ := redis.HGet(ctx, "bot:status", "online").Bool()
	pid, _ := redis.HGet(ctx, "bot:metrics", "pid").Int()
	memoryMB, _ := redis.HGet(ctx, "bot:metrics", "memory_mb").Int()
	cpuPercent, _ := redis.HGet(ctx, "bot:metrics", "cpu_percent").Int()
	startedAt, _ := redis.HGet(ctx, "bot:status", "started_at").Int64()

	status.Online = online
	status.PID = pid
	status.MemoryMB = memoryMB
	status.CPUPercent = cpuPercent
	status.StartedAt = startedAt

	return status
}
