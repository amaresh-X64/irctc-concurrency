package config

import (
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	SpringBootURL string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8002"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://irctc_user:irctc_pass@localhost:5432/irctc_db"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		SpringBootURL: getEnv("SPRINGBOOT_URL", "http://localhost:8003"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
