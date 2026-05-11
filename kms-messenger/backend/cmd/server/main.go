package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"kms-backend/internal/config"
	"kms-backend/internal/database"
	"kms-backend/internal/handlers"
	"kms-backend/internal/websocket"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewDatabase(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub(db.DB)
	go hub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db.DB, cfg)
	chatHandler := handlers.NewChatHandler(db.DB)

	// Setup Gin
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", handlers.AuthMiddleware(cfg), authHandler.Logout)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(handlers.AuthMiddleware(cfg))
	{
		// Chat routes
		chats := api.Group("/chats")
		{
			chats.GET("", chatHandler.GetChats)
			chats.POST("", chatHandler.CreateChat)
			chats.GET("/:id/messages", chatHandler.GetMessages)
			chats.POST("/:id/messages", chatHandler.SendMessage)
		}

		// WebSocket endpoint
		chats.GET("/ws", websocket.WSHandler(hub))
	}

	// Get port from config
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
