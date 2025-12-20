package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	cfg := LoadFromEnv()

	if cfg.Database.URL != "postgres://localingo_user:localingo_pass@localhost:5432/localingo?sslmode=disable" {
		t.Errorf("Unexpected database URL: %s", cfg.Database.URL)
	}

	if cfg.Database.MaxConnections != 25 {
		t.Errorf("Expected max connections 25, got %d", cfg.Database.MaxConnections)
	}

	if cfg.Database.MaxIdleConnections != 5 {
		t.Errorf("Expected max idle connections 5, got %d", cfg.Database.MaxIdleConnections)
	}

	if cfg.Database.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("Expected conn max lifetime 5m, got %v", cfg.Database.ConnMaxLifetime)
	}

	if !cfg.ChatHistory.Enabled {
		t.Error("Expected chat history enabled by default")
	}

	if !cfg.ChatHistory.AutoSave {
		t.Error("Expected chat history auto-save enabled by default")
	}
}

func TestLoadFromEnv_CustomValues(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://custom_user:custom_pass@localhost:5433/custom_db?sslmode=require")
	os.Setenv("DATABASE_MAX_CONNECTIONS", "50")
	os.Setenv("DATABASE_MAX_IDLE_CONNECTIONS", "10")
	os.Setenv("CHAT_HISTORY_ENABLED", "false")
	os.Setenv("CHAT_HISTORY_AUTO_SAVE", "false")

	defer os.Clearenv()

	cfg := LoadFromEnv()

	if cfg.Database.URL != "postgres://custom_user:custom_pass@localhost:5433/custom_db?sslmode=require" {
		t.Errorf("Unexpected database URL: %s", cfg.Database.URL)
	}

	if cfg.Database.MaxConnections != 50 {
		t.Errorf("Expected max connections 50, got %d", cfg.Database.MaxConnections)
	}

	if cfg.Database.MaxIdleConnections != 10 {
		t.Errorf("Expected max idle connections 10, got %d", cfg.Database.MaxIdleConnections)
	}

	if cfg.ChatHistory.Enabled {
		t.Error("Expected chat history disabled")
	}

	if cfg.ChatHistory.AutoSave {
		t.Error("Expected chat history auto-save disabled")
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	os.Setenv("TEST_INT", "not_a_number")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 42)
	if result != 42 {
		t.Errorf("Expected default value 42, got %d", result)
	}
}

func TestGetEnvBool_InvalidValue(t *testing.T) {
	os.Setenv("TEST_BOOL", "not_a_bool")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvBool("TEST_BOOL", true)
	if result != true {
		t.Error("Expected default value true")
	}
}
