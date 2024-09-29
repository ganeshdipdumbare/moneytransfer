package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type RetryConfig struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	MaxRetries int
}

type Config struct {
	DatabaseURL string
	ServerPort  string
	LogLevel    slog.Level
	GinMode     string
	RetryConfig RetryConfig
}

func LoadConfig() (*Config, error) {
	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://user:password@localhost:5432/qonto_accounts?sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		LogLevel:    getLogLevel(getEnv("LOG_LEVEL", "info")),
		GinMode:     getEnv("GIN_MODE", "debug"),
		RetryConfig: RetryConfig{
			BaseDelay:  time.Duration(getIntEnv("RETRY_BASE_DELAY", 100)) * time.Millisecond,
			MaxDelay:   time.Duration(getIntEnv("RETRY_MAX_DELAY", 5000)) * time.Millisecond,
			MaxRetries: getIntEnv("RETRY_MAX_RETRIES", 5),
		},
	}
	return config, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}

func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
