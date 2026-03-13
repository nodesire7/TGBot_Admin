package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tgbot/admin/config"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WSEvents handles WebSocket connections for real-time events
func WSEvents(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	redis := config.GetRedis()
	ctx := context.Background()

	// Subscribe to event stream
	subscriber := redis.Subscribe(ctx, "channel:events")
	defer subscriber.Close()

	// Get last 10 events on connect
	streams, _ := redis.XRevRange(ctx, "stream:events", "+", "-").Result()
	for i, stream := range streams {
		if i >= 10 {
			break
		}
		data, _ := json.Marshal(stream.Values)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Listen for new events
	ch := subscriber.Channel()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		case <-ticker.C:
			// Send ping to keep connection alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WSMetrics handles WebSocket connections for real-time metrics
func WSMetrics(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	redis := config.GetRedis()
	ctx := context.Background()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get bot metrics
			metrics := make(map[string]interface{})

			online, _ := redis.HGet(ctx, "bot:status", "online").Bool()
			pid, _ := redis.HGet(ctx, "bot:metrics", "pid").Int()
			memoryMB, _ := redis.HGet(ctx, "bot:metrics", "memory_mb").Int()
			cpuPercent, _ := redis.HGet(ctx, "bot:metrics", "cpu_percent").Int()

			metrics["bot_status"] = map[string]interface{}{
				"online":      online,
				"pid":         pid,
				"memory_mb":   memoryMB,
				"cpu_percent": cpuPercent,
			}

			// Get today's stats
			today := time.Now().Format("2006-01-02")
			verified, _ := redis.Get(ctx, "stats:"+today+":verified").Int64()
			failed, _ := redis.Get(ctx, "stats:"+today+":failed").Int64()
			kicked, _ := redis.Get(ctx, "stats:"+today+":kicked").Int64()

			metrics["today_stats"] = map[string]interface{}{
				"verified": verified,
				"failed":   failed,
				"kicked":   kicked,
			}

			data, _ := json.Marshal(metrics)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}
