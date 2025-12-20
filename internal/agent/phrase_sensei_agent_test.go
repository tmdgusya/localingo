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

func TestPhraseSenseiAgentRephrase(t *testing.T) {
	expectedText := "This is a rephrased test phrase"
	mockClient := &MockLLMClient{
		GenerateFunc: func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
			return &GenerateResponse{
				Text: expectedText,
				Done: true,
			}, nil
		},
	}

	agent := NewPhraseSenseiAgent(mockClient)

	if agent == nil {
		t.Fatal("Expected agent to be non-nil")
	}

	// Test the Rephrase method
	req := &GenerateRequest{
		Model:  "test-model",
		Prompt: "test prompt",
	}

	resp, err := agent.Rephrase(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Text != expectedText {
		t.Fatalf("Expected '%s', got '%s'", expectedText, resp.Text)
	}

	if !resp.Done {
		t.Fatal("Expected Done to be true")
	}
}

func TestPhraseSenseiAgentRephraseStream(t *testing.T) {
	expectedText := "This is a streamed rephrased test phrase"
	var capturedResponse string

	mockClient := &MockLLMClient{
		GenerateStreamFunc: func(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error {
			return fn(&GenerateResponse{
				Text: expectedText,
				Done: true,
			})
		},
	}

	agent := NewPhraseSenseiAgent(mockClient)

	// Test the RephraseStream method
	req := &GenerateRequest{
		Model:  "test-model",
		Prompt: "test prompt",
	}

	err := agent.RephraseStream(context.Background(), req, func(resp *GenerateResponse) error {
		capturedResponse = resp.Text
		return nil
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedResponse != expectedText {
		t.Fatalf("Expected '%s', got '%s'", expectedText, capturedResponse)
	}
}
