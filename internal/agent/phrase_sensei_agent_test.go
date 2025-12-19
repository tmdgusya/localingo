package agent

import (
	"context"
	"testing"
)

func TestNewPhraseSenseiAgent(t *testing.T) {
	mockClient := &MockLLMClient{}
	agent := NewPhraseSenseiAgent(mockClient)

	if agent == nil {
		t.Fatal("Expected agent to be non-nil")
	}

	if agent.client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestPhraseSenseiAgentWithMockResponse(t *testing.T) {
	mockClient := &MockLLMClient{
		GenerateFunc: func(ctx context.Context, prompt string) (string, error) {
			return "This is a test phrase", nil
		},
	}

	agent := NewPhraseSenseiAgent(mockClient)

	if agent == nil {
		t.Fatal("Expected agent to be non-nil")
	}

	// client가 제대로 동작하는지 테스트
	response, err := agent.client.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response != "This is a test phrase" {
		t.Fatalf("Expected 'This is a test phrase', got '%s'", response)
	}
}
