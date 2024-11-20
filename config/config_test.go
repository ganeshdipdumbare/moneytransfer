package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "8080",
				LogLevel:    slog.LevelInfo,
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  100 * time.Millisecond,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid database URL",
			config: Config{
				DatabaseURL: "not-a-url",
				ServerPort:  "8080",
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  100 * time.Millisecond,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "empty server port",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "",
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  100 * time.Millisecond,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid gin mode",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "8080",
				GinMode:     "invalid",
				RetryConfig: RetryConfig{
					BaseDelay:  100 * time.Millisecond,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "zero base delay",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "8080",
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  0,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "max delay less than base delay",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "8080",
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  200 * time.Millisecond,
					MaxDelay:   100 * time.Millisecond,
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: Config{
				DatabaseURL: "postgresql://user:pass@localhost:5432/db",
				ServerPort:  "8080",
				GinMode:     "release",
				RetryConfig: RetryConfig{
					BaseDelay:  100 * time.Millisecond,
					MaxDelay:   1000 * time.Millisecond,
					MaxRetries: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
