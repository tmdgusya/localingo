package agent

import (
	"context"
)

// MockLLMClient is a mock implementation of LLMClient for testing
type MockLLMClient struct {
	GenerateFunc       func(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStreamFunc func(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error
}

func (m *MockLLMClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, req)
	}
	// Default mock behavior
	return &GenerateResponse{
		Text: "mock response",
		Done: true,
	}, nil
}

func (m *MockLLMClient) GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(ctx, req, fn)
	}
	// Default mock behavior
	return fn(&GenerateResponse{
		Text: "mock response",
		Done: true,
	})
}
