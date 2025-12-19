package client

import (
	"os"
	"testing"
)

const testModel = "llama3.2"

func TestNewOllamaClient(t *testing.T) {
	client, err := NewOllamaClient(testModel)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	err = client.VerifyConnection()
	if err != nil {
		t.Fatalf("Failed to verify connection: %v", err)
	}
}

func TestNewOllamaClientWithCustomHost(t *testing.T) {
	originalHost := os.Getenv("OLLAMA_HOST")
	defer func() {
		if originalHost != "" {
			os.Setenv("OLLAMA_HOST", originalHost)
		} else {
			os.Unsetenv("OLLAMA_HOST")
		}
	}()

	os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	client, err := NewOllamaClient(testModel)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	err = client.VerifyConnection()
	if err != nil {
		t.Fatalf("Failed to verify connection: %v", err)
	}
}

func TestNewOllamaClientWithWrongHost(t *testing.T) {
	os.Setenv("OLLAMA_HOST", "http://localhost:999999")
	defer os.Unsetenv("OLLAMA_HOST")

	client, err := NewOllamaClient(testModel)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	// 실제 연결 테스트 - 잘못된 호스트이므로 에러가 예상됨
	err = client.VerifyConnection()
	if err == nil {
		t.Fatal("Expected error when connecting to invalid host, but got none")
	}

	t.Logf("Got expected error: %v", err)
}
