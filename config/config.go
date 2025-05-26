package config

import (
	"os"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Redis  RedisConfig
	JWT    JWTConfig
	Server ServerConfig
	WeChat WeChatConfig
}

// WeChatConfig holds WeChat Mini Program configuration
type WeChatConfig struct {
	AppID     string
	AppSecret string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	return &Config{
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			SecretKey:       os.Getenv("JWT_SECRET_KEY"),
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Server: ServerConfig{
			Port: "8080",
		},
		WeChat: WeChatConfig{
			AppID:     os.Getenv("WECHAT_APPID"),
			AppSecret: os.Getenv("WECHAT_APPSECRET"),
		},
	}
}
