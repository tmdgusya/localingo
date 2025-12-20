package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Database    DatabaseConfig
	ChatHistory ChatHistoryConfig
	Ollama      OllamaConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL                string
	MaxConnections     int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
}

// ChatHistoryConfig holds chat history feature configuration
type ChatHistoryConfig struct {
	Enabled  bool
	AutoSave bool
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	Host         string
	DefaultModel string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	return &Config{
		Database: DatabaseConfig{
			URL:                getEnv("DATABASE_URL", "postgres://localingo_user:localingo_pass@localhost:5432/localingo?sslmode=disable"),
			MaxConnections:     getEnvInt("DATABASE_MAX_CONNECTIONS", 25),
			MaxIdleConnections: getEnvInt("DATABASE_MAX_IDLE_CONNECTIONS", 5),
			ConnMaxLifetime:    5 * time.Minute,
		},
		ChatHistory: ChatHistoryConfig{
			Enabled:  getEnvBool("CHAT_HISTORY_ENABLED", true),
			AutoSave: getEnvBool("CHAT_HISTORY_AUTO_SAVE", true),
		},
		Ollama: OllamaConfig{
			Host:         getEnv("OLLAMA_HOST", "http://localhost:11434"),
			DefaultModel: getEnv("OLLAMA_MODEL", "qwen3:4b"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
