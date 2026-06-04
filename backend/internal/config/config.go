package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DBHost                  string
	DBPort                  string
	DBName                  string
	DBUser                  string
	DBPassword              string
	DBMaxOpenConns          int
	DBMaxIdleConns          int
	DBConnMaxLifetimeMinutes int

	JWTSecret      string
	JWTExpiryHours int

	AllowedOrigins string
}

func Load() (*Config, error) {
	// .env is optional in production (env vars already set)
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:                  getEnv("DB_HOST", "127.0.0.1"),
		DBPort:                  getEnv("DB_PORT", "3306"),
		DBName:                  getEnv("DB_NAME", "onecare"),
		DBUser:                  getEnv("DB_USER", "root"),
		DBPassword:              getEnv("DB_PASSWORD", ""),
		DBMaxOpenConns:          getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:          getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetimeMinutes: getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5),

		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),

		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
