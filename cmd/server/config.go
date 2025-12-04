package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// AppConfig holds all application configuration.
type AppConfig struct {
	// Server
	ServerHost string
	ServerPort int
	ServerMode string // debug, release, test

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecretKey     string
	JWTTokenDuration time.Duration

	// CORS
	CORSAllowOrigins []string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *AppConfig {
	return &AppConfig{
		// Server
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),
		ServerMode: getEnv("SERVER_MODE", "debug"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "go_backend_template"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// JWT
		JWTSecretKey:     getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		JWTTokenDuration: getEnvAsDuration("JWT_TOKEN_DURATION", 24*time.Hour),

		// CORS
		CORSAllowOrigins: getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
	}
}

// Validate checks if the configuration is valid.
func (c *AppConfig) Validate() error {
	if c.JWTSecretKey == "your-secret-key-change-in-production" && c.ServerMode == "release" {
		log.Println("WARNING: Using default JWT secret key in production mode!")
	}
	return nil
}

// Helper functions for environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

