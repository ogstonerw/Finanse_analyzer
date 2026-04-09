package api

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host                   string
	Port                   string
	Environment            string
	SessionTTLHours        int
	AIMode                 string
	AIProvider             string
	AIModel                string
	AIAPIEndpoint          string
	AIAPIKey               string
	DatabaseHost           string
	DatabasePort           string
	DatabaseUser           string
	DatabasePassword       string
	DatabaseName           string
	DatabaseSSLMode        string
	DatabaseMaxOpenConns   int
	DatabaseMaxIdleConns   int
	DatabaseConnMaxLifeMin int
}

func LoadConfig() Config {
	return Config{
		Host:                   getEnv("APP_HOST", "0.0.0.0"),
		Port:                   getEnv("APP_PORT", "8080"),
		Environment:            getEnv("APP_ENV", "development"),
		SessionTTLHours:        getEnvInt("APP_SESSION_TTL_HOURS", 24),
		AIMode:                 getEnv("AI_MODE", "fallback"),
		AIProvider:             getEnv("AI_PROVIDER", "openai"),
		AIModel:                getEnv("AI_MODEL", ""),
		AIAPIEndpoint:          getEnv("AI_API_ENDPOINT", ""),
		AIAPIKey:               getEnv("AI_API_KEY", ""),
		DatabaseHost:           getEnv("DB_HOST", "localhost"),
		DatabasePort:           getEnv("DB_PORT", "5432"),
		DatabaseUser:           getEnv("DB_USER", "postgres"),
		DatabasePassword:       getEnv("DB_PASSWORD", "postgres"),
		DatabaseName:           getEnv("DB_NAME", "market_ai"),
		DatabaseSSLMode:        getEnv("DB_SSLMODE", "disable"),
		DatabaseMaxOpenConns:   getEnvInt("DB_MAX_OPEN_CONNS", 10),
		DatabaseMaxIdleConns:   getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DatabaseConnMaxLifeMin: getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
