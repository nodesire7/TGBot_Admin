package main

import (
	"log"
	"os"

	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/routers"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Printf("Warning: .env file not found, using system env")
	}

	// Initialize database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer config.CloseDB()

	// Initialize Redis
	if err := config.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer config.CloseRedis()

	// Setup router
	r := routers.SetupRouter()

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
