package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kms-messenger/backend/internal/auth"
	"github.com/kms-messenger/backend/internal/config"
	"github.com/kms-messenger/backend/internal/database"
	"github.com/kms-messenger/backend/internal/handlers"
	"github.com/kms-messenger/backend/internal/repository"
	"github.com/kms-messenger/backend/internal/websocket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	userRepo := repository.NewUserRepository()
	refreshTokenRepo := repository.NewRefreshTokenRepository()
	authService := auth.NewAuthService(userRepo, refreshTokenRepo, cfg)
	
	hub := websocket.NewHub()
	go hub.Run()

	authHandler := handlers.NewAuthHandler(authService)
	wsHandler := handlers.NewWSHandler(hub)

	router := gin.Default()

	router.POST("/api/v1/auth/register", authHandler.Register)
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)
	router.POST("/api/v1/auth/logout", handlers.AuthMiddleware(authService), authHandler.Logout)
	router.GET("/api/v1/auth/me", handlers.AuthMiddleware(authService), authHandler.Me)

	router.GET("/api/v1/ws", handlers.AuthMiddleware(authService), wsHandler.HandleWebSocket)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
