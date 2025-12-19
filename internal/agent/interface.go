package agent

import (
	"context"
)

// LLMClient is the interface for LLM operations
type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
