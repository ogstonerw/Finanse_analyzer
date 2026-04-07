package api

import (
	"fmt"
	"os"
)

type Config struct {
	Host             string
	Port             string
	Environment      string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string
}

func LoadConfig() Config {
	return Config{
		Host:             getEnv("APP_HOST", "0.0.0.0"),
		Port:             getEnv("APP_PORT", "8080"),
		Environment:      getEnv("APP_ENV", "development"),
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     getEnv("DB_USER", "postgres"),
		DatabasePassword: getEnv("DB_PASSWORD", "postgres"),
		DatabaseName:     getEnv("DB_NAME", "market_ai"),
		DatabaseSSLMode:  getEnv("DB_SSLMODE", "disable"),
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
