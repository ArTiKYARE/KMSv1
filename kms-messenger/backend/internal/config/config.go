package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL            string
	RedisURL         string
	JWTSecret        string
	JWTExpiry        time.Duration
	RefreshTokenExpiry time.Duration
	AppEnv           string
	ServerPort       string
}

func Load() *Config {
	// Load .env file if exists
	godotenv.Load()

	jwtExpiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "15m"))
	refreshExpiry, _ := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "7d"))

	return &Config{
		DBURL:            getEnv("DATABASE_URL", "postgres://kms_admin:secure_password_change_me@localhost:5432/kms_db?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", "super_secret_key_change_in_prod"),
		JWTExpiry:        jwtExpiry,
		RefreshTokenExpiry: refreshExpiry,
		AppEnv:           getEnv("APP_ENV", "development"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
