package config

import (
	"fmt"
	"os"
)

// Config holds all configuration values for the app.
// We read these from environment variables (loaded from .env).

type Config struct {
	AppPort string
	AppEnv  string
	DB      DBConfig
}

// DBConfig holds everything needed to connect to PostgreSQL.

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load reads environment variables and returns a Config struct.
// If any required variable is missing, it returns an error.

func Load() (*Config, error) {
	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
	//validate all required fields
	if cfg.DB.User == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.DB.Name == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}
	return cfg, nil
}

// DSN builds the PostgreSQL connection string from the config.
// Example: "host=localhost port=5432 user=bar108_user ..."
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// getEnv reads an environment variable and returns a fallback
// value if it's not set. This prevents panics on missing vars.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
