package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
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
	v := viper.New()

	// Set defaults
	v.SetDefault("database_url", "postgresql://user:password@localhost:5432/qonto_accounts?sslmode=disable")
	v.SetDefault("server_port", "8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("gin_mode", "debug")
	v.SetDefault("retry.base_delay", 100)
	v.SetDefault("retry.max_delay", 5000)
	v.SetDefault("retry.max_retries", 5)

	// Tell Viper to automatically override values from environment variables
	v.AutomaticEnv()

	// Optional: If you want to keep your existing env var names
	v.BindEnv("database_url", "DATABASE_URL")
	v.BindEnv("server_port", "SERVER_PORT")
	v.BindEnv("log_level", "LOG_LEVEL")
	v.BindEnv("gin_mode", "GIN_MODE")
	v.BindEnv("retry.base_delay", "RETRY_BASE_DELAY")
	v.BindEnv("retry.max_delay", "RETRY_MAX_DELAY")
	v.BindEnv("retry.max_retries", "RETRY_MAX_RETRIES")

	config := &Config{
		DatabaseURL: v.GetString("database_url"),
		ServerPort:  v.GetString("server_port"),
		LogLevel:    getLogLevel(v.GetString("log_level")),
		GinMode:     v.GetString("gin_mode"),
		RetryConfig: RetryConfig{
			BaseDelay:  time.Duration(v.GetInt("retry.base_delay")) * time.Millisecond,
			MaxDelay:   time.Duration(v.GetInt("retry.max_delay")) * time.Millisecond,
			MaxRetries: v.GetInt("retry.max_retries"),
		},
	}

	return config, nil
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
