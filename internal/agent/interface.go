package agent

import (
	"context"
)

// LLMClient is the interface for LLM operations
type LLMClient interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error
}
