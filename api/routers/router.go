package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/tgbot/admin/config"
	"github.com/tgbot/admin/middleware"
	"github.com/tgbot/admin/routers/routes"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORS())

	// Static files
	r.Static("/assets", "../web/assets")
	r.StaticFile("/", "../web/index.html")
	r.StaticFile("/dashboard", "../web/index.html")
	r.StaticFile("/groups", "../web/index.html")
	r.StaticFile("/plugins", "../web/index.html")
	r.StaticFile("/logs", "../web/index.html")
	r.StaticFile("/setup", "../web/index.html")
	r.StaticFile("/bots", "../web/index.html")

	// API routes
	api := r.Group("/api")
	{
		// Setup routes (public, only available before configuration)
		setup := api.Group("/setup")
		{
			setup.GET("/status", routes.GetSetupStatus)
			setup.GET("/config", routes.GetSetupConfig)
			setup.POST("/test-db", routes.TestDatabase)
			setup.POST("/test-redis", routes.TestRedis)
			setup.POST("/save", routes.SaveSetup)
		}

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", routes.Login)
			auth.POST("/logout", routes.Logout)
			auth.POST("/refresh", routes.RefreshToken)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// Auth
			protected.GET("/auth/me", routes.GetCurrentUser)

			// Dashboard
			protected.GET("/dashboard/stats", routes.GetDashboardStats)
			protected.GET("/dashboard/timeline", routes.GetTimeline)

			// Bots
			protected.GET("/bots", routes.GetBots)
			protected.GET("/bots/:id", routes.GetBot)
			protected.POST("/bots", routes.CreateBot)
			protected.PUT("/bots/:id", routes.UpdateBot)
			protected.DELETE("/bots/:id", routes.DeleteBot)
			protected.POST("/bots/:id/start", routes.StartBot)
			protected.POST("/bots/:id/stop", routes.StopBot)
			protected.POST("/bots/:id/restart", routes.RestartBot)
			protected.POST("/bots/test-token", routes.TestBotToken)

			// Groups
			protected.GET("/groups", routes.GetGroups)
			protected.GET("/groups/:chat_id", routes.GetGroup)
			protected.PUT("/groups/:chat_id", routes.UpdateGroup)
			protected.DELETE("/groups/:chat_id", routes.DeleteGroup)
			protected.POST("/groups/:chat_id/sync", routes.SyncGroup)

			// Blacklist
			protected.GET("/groups/:chat_id/blacklist", routes.GetBlacklist)
			protected.POST("/groups/:chat_id/blacklist", routes.AddToBlacklist)
			protected.DELETE("/groups/:chat_id/blacklist/:user_id", routes.RemoveFromBlacklist)

			// Plugins
			protected.GET("/plugins", routes.GetPlugins)
			protected.GET("/plugins/:id", routes.GetPlugin)
			protected.POST("/plugins/install", routes.InstallPlugin)
			protected.PUT("/plugins/:id", routes.UpdatePlugin)
			protected.DELETE("/plugins/:id", routes.UninstallPlugin)
			protected.POST("/plugins/:id/enable", routes.EnablePlugin)
			protected.POST("/plugins/:id/disable", routes.DisablePlugin)
			protected.POST("/plugins/:id/test", routes.TestPlugin)
			protected.POST("/plugins/:id/reload", routes.ReloadPlugin)
			protected.GET("/plugins/:id/logs", routes.GetPluginLogs)

			// Bot Plugins
			protected.GET("/bots/:id/plugins", routes.GetBotPlugins)
			protected.PUT("/bots/:id/plugins/:plugin_id", routes.UpdateBotPlugin)

			// Marketplace
			protected.GET("/market/plugins", routes.GetMarketPlugins)
			protected.GET("/market/plugins/:id", routes.GetMarketPlugin)
			protected.GET("/market/plugins/:id/code", routes.GetMarketPluginCode)
			protected.POST("/market/install/:id", routes.InstallFromMarket)
			protected.GET("/market/categories", routes.GetMarketCategories)
			protected.GET("/market/repositories", routes.GetMarketRepositories)

			// Logs
			protected.GET("/logs/verification", routes.GetVerificationLogs)
			protected.GET("/logs/action", routes.GetActionLogs)
		}
	}

	// WebSocket
	r.GET("/ws/events", func(c *gin.Context) {
		if !config.IsConfigured() {
			c.JSON(503, gin.H{"error": "System not configured"})
			return
		}
		routes.WSEvents(c)
	})
	r.GET("/ws/metrics", func(c *gin.Context) {
		if !config.IsConfigured() {
			c.JSON(503, gin.H{"error": "System not configured"})
			return
		}
		routes.WSMetrics(c)
	})

	return r
}
