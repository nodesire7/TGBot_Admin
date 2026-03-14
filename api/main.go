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

	// Load setup configuration
	cfg, err := config.LoadSetupConfig()
	if err != nil {
		log.Printf("Warning: Failed to load setup config: %v", err)
	}

	// Check if system is configured
	if cfg != nil && cfg.IsConfigured {
		// Initialize database from saved config
		if err := config.InitDBFromConfig(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer config.CloseDB()

		// Initialize Redis from saved config
		if err := config.InitRedisFromConfig(); err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		defer config.CloseRedis()

		log.Println("System configured, starting in normal mode")
	} else {
		log.Println("System not configured, starting in setup mode")
		log.Println("Please visit http://localhost:8000 to complete setup")
	}

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
