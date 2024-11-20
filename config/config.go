package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type RetryConfig struct {
	BaseDelay  time.Duration `validate:"required,gt=0"`
	MaxDelay   time.Duration `validate:"required,gtefield=BaseDelay"`
	MaxRetries int           `validate:"gte=0"`
}

type Config struct {
	DatabaseURL string `validate:"required,url"`
	ServerPort  string `validate:"required"`
	LogLevel    slog.Level
	GinMode     string      `validate:"required,oneof=debug release test"`
	RetryConfig RetryConfig `validate:"required"`
}

func (c *Config) Validate() error {
	validate := validator.New()

	// Register custom validation for time.Duration comparison
	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		rc := sl.Current().Interface().(RetryConfig)
		if rc.MaxDelay < rc.BaseDelay {
			sl.ReportError(rc.MaxDelay, "MaxDelay", "MaxDelay", "gtefield", "BaseDelay")
		}
	}, RetryConfig{})

	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
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

	// Handle BindEnv errors
	if err := v.BindEnv("database_url", "DATABASE_URL"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("server_port", "SERVER_PORT"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("log_level", "LOG_LEVEL"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("gin_mode", "GIN_MODE"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("retry.base_delay", "RETRY_BASE_DELAY"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("retry.max_delay", "RETRY_MAX_DELAY"); err != nil {
		return nil, err
	}
	if err := v.BindEnv("retry.max_retries", "RETRY_MAX_RETRIES"); err != nil {
		return nil, err
	}

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

	if err := config.Validate(); err != nil {
		return nil, err
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
