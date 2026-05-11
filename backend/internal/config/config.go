package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	AppEnv             string
	ServerPort         string
}

func Load() (*Config, error) {
	if os.Getenv("APP_ENV") == "development" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Warning: .env file not found, using environment variables")
		}
	}

	jwtExpiryStr := getEnv("JWT_EXPIRY", "15m")
	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	refreshExpiryStr := getEnv("REFRESH_TOKEN_EXPIRY", "7d")
	refreshExpiry, err := time.ParseDuration(refreshExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY: %w", err)
	}

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://kms_admin:secure_password_change_me@localhost:5432/kms_db?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "super_secret_key_change_in_prod"),
		JWTExpiry:          jwtExpiry,
		RefreshTokenExpiry: refreshExpiry,
		AppEnv:             getEnv("APP_ENV", "development"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
