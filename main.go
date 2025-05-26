package main

import (
	"context"
	"log"
	"net/http"

	"github.com/LIUHUANUCAS/auth/config"
	"github.com/LIUHUANUCAS/auth/handlers"
	"github.com/LIUHUANUCAS/auth/middleware"
	"github.com/LIUHUANUCAS/auth/models"
	"github.com/LIUHUANUCAS/auth/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration
	cfg := config.GetConfig()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Ping Redis to check connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize user store
	userStore := models.NewUserStore(redisClient)

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(&cfg.JWT)

	// Initialize WeChat manager
	wechatManager := utils.NewWeChatManager(&cfg.WeChat)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(userStore, jwtManager, wechatManager, redisClient)

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Public routes
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/refresh", authHandler.RefreshToken)
	router.POST("/logout", authHandler.Logout)
	router.POST("/wechat/login", authHandler.WeChatLogin)

	// Protected routes
	protected := router.Group("/")
	protected.Use(authMiddleware.AuthRequired())
	{
		protected.GET("/me", authHandler.Me)

		// Example protected API endpoint
		protected.GET("/api/protected", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(http.StatusOK, gin.H{
				"message": "This is a protected endpoint",
				"user_id": userID,
			})
		})
	}

	// Start the server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
