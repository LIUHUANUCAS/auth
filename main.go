package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	ctx := context.Background()

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
		protected.GET("/v1/daily_new_house", proxyHandler)
		protected.GET("/v1/daily_unfinished_house", proxyHandler)
		protected.GET("/v1/month_house", proxyHandler)
		protected.GET("/v2/sh/new_daily_house", proxyHandler)
		protected.GET("/v2/sh/old_daily_house", proxyHandler)
		protected.GET("/v3/fortune/daily", proxyHandler)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	listener, err := newNgrokListener(ctx, cfg)
	if err != nil {
		log.Fatalf("new listner err:%s\n", err)
	}
	defer listener.Close()

	go func() {
		log.Println("Starting server on", listener.Addr(), listener.Addr().String(), "port:", cfg.Server.Port)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	// Accept graceful shutdown signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %s\n", err)
	}
	log.Println("Server exiting")

}
