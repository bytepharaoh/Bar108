package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort string
	AppEnv  string
	DB      DBConfig
	JWT     JWTConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds everything needed to sign and verify tokens.
type JWTConfig struct {
	Secret      string // the secret key — NEVER expose this
	ExpiryHours int    // how many hours until token expires
}

func Load() (*Config, error) {
	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		expiryHours = 24
	}

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
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpiryHours: expiryHours,
		},
	}

	if cfg.DB.User == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.DB.Name == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
