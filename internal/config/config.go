package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	Environment    string
	JWTSecret      string
	PaddleVendorID string
	PaddleAPIKey   string
	AIAPIKey       string
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", ":8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost/should_i_get_it?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		PaddleVendorID: getEnv("PADDLE_VENDOR_ID", ""),
		PaddleAPIKey:   getEnv("PADDLE_API_KEY", ""),
		AIAPIKey:       getEnv("AI_API_KEY", ""),
		AllowedOrigins: []string{
			getEnv("ALLOWED_ORIGIN", "http://localhost:8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
