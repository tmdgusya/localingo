package agent

import (
	"context"
)

// MockLLMClient is a mock implementation of LLMClient for testing
type MockLLMClient struct {
	GenerateFunc func(ctx context.Context, prompt string) (string, error)
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, prompt)
	}
	return "mock response", nil
}
