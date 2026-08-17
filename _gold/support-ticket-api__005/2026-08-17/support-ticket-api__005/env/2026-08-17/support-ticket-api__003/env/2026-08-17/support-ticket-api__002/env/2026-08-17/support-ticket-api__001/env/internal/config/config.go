package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
	DBMaxConns  int32
	SLAOverride string
}

func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "18106"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:55432/support_ticket?sslmode=disable"),
		DBMaxConns:  getEnvInt32("DB_MAX_CONNS", 10),
		SLAOverride: getEnv("SLA_NOW", ""),
	}
}

func (c Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT cannot be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL cannot be empty")
	}
	if c.DBMaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS must be positive")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(parsed)
}
