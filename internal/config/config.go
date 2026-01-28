package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all application configuration
type Config struct {
	// Database configuration
	DatabaseURL         string
	DatabaseMaxConns    int32
	DatabaseMinConns    int32
	DatabaseMaxLifetime time.Duration

	// Redis configuration
	RedisURL string

	// HuggingFace API configuration
	HuggingFaceAPIKey  string
	HuggingFaceModelURL string
	HuggingFaceTimeout time.Duration

	// Service ports
	GatewayPort       string
	ModerationPort    string
	PolicyEnginePort  string
	ReviewPort        string

	// Service hosts (for Docker networking)
	ModerationHost   string
	PolicyEngineHost string
	ReviewHost       string

	// Logging
	LogLevel string
	LogJSON  bool

	// Application
	Environment string
	Version     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Database defaults
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/text_moderator?sslmode=disable"),
		DatabaseMaxConns:    getEnvAsInt32("DATABASE_MAX_CONNS", 25),
		DatabaseMinConns:    getEnvAsInt32("DATABASE_MIN_CONNS", 5),
		DatabaseMaxLifetime: getEnvAsDuration("DATABASE_MAX_LIFETIME", time.Hour),

		// Redis defaults
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),

		// HuggingFace defaults
		HuggingFaceAPIKey:   getEnv("HUGGINGFACE_API_KEY", ""),
		HuggingFaceModelURL: getEnv("HUGGINGFACE_MODEL_URL", "https://api-inference.huggingface.co/models/unitary/toxic-bert"),
		HuggingFaceTimeout:  getEnvAsDuration("HUGGINGFACE_TIMEOUT", 30*time.Second),

		// Service ports
		GatewayPort:      getEnv("GATEWAY_PORT", "8080"),
		ModerationPort:   getEnv("MODERATION_PORT", "8081"),
		PolicyEnginePort: getEnv("POLICY_ENGINE_PORT", "8082"),
		ReviewPort:       getEnv("REVIEW_PORT", "8083"),

		// Service hosts (default to Docker Compose service names, fallback to localhost)
		ModerationHost:   getEnv("MODERATION_HOST", "moderation"),
		PolicyEngineHost: getEnv("POLICY_ENGINE_HOST", "policy-engine"),
		ReviewHost:       getEnv("REVIEW_HOST", "review"),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogJSON:  getEnvAsBool("LOG_JSON", true),

		// Application
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "0.1.0"),
	}

	return cfg, nil
}

// NewLogger creates a new zap logger based on configuration
func (c *Config) NewLogger() (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", c.LogLevel, err)
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableStacktrace = true
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if !c.LogJSON {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.InitialFields = map[string]interface{}{
		"environment": c.Environment,
		"version":     c.Version,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt32 retrieves an environment variable as int32 or returns a default value
func getEnvAsInt32(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(value)
}

// getEnvAsBool retrieves an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration retrieves an environment variable as duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
