package agent

import (
	"context"
)

// LLMClient defines the interface for interacting with LLM providers
type LLMClient interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error
}