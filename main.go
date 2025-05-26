package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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

	// Create reverse proxy for localhost:8081
	targetURL, err := url.Parse(cfg.Server.ProxyURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Handler function for reverse proxy
	proxyHandler := func(c *gin.Context) {
		// Update the request URL
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		c.Request.Host = targetURL.Host

		// Serve the request using the reverse proxy
		proxy.ServeHTTP(c.Writer, c.Request)
	}

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

		// Proxy routes that require authentication
		protected.GET("/v1/daily_house", proxyHandler)
		protected.GET("/v1/month_house", proxyHandler)
		protected.GET("/v2/sh/new_daily_house", proxyHandler)
		protected.GET("/v2/sh/old_daily_house", proxyHandler)
		protected.GET("/v3/fortune/daily", proxyHandler)
	}

	// Start the server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
