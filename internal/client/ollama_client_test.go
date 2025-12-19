package client

import (
	"os"
	"testing"
)

func TestCreateOllamaClient(t *testing.T) {
	client, err := CreateOllamaClient()

	if client == nil || err != nil {
		t.Fatal("Expected client to be non-nil")
	}

	err = VerifyConnection(client)
	if err != nil {
		t.Fatalf("Failed to verify connection: %v", err)
	}
}

func TestCreateOllamaClientWithCustomHost(t *testing.T) {
	originalHost := os.Getenv("OLLAMA_HOST")
	defer func() {
		if originalHost != "" {
			os.Setenv("OLLAMA_HOST", originalHost)
		} else {
			os.Unsetenv("OLLAMA_HOST")
		}
	}()

	os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	client, err := CreateOllamaClient()

	if client == nil || err != nil {
		t.Fatal("Expected client to be non-nil")
	}

	err = VerifyConnection(client)
	if err != nil {
		t.Fatalf("Failed to verify connection: %v", err)
	}
}

func TestCreateOllamaClientWithWrongHost(t *testing.T) {
	os.Setenv("OLLAMA_HOST", "http://localhost:999999")
	defer os.Unsetenv("OLLAMA_HOST")

	client, err := CreateOllamaClient()

	if client == nil || err != nil {
		t.Fatal("Expected client to be non-nil")
	}

	err = VerifyConnection(client)
	if err == nil {
		t.Fatal("Expected error when connecting to invalid host, but got none")
	}

	t.Logf("Got expected error: %v", err)
}
