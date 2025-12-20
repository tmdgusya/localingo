package agent

import (
	"context"
	"strings"

	"github.com/ollama/ollama/api"
)

// OllamaClient is an adapter that implements LLMClient using Ollama API
type OllamaClient struct {
	client *api.Client
}

// NewOllamaClient creates a new Ollama client adapter
func NewOllamaClient(client *api.Client) *OllamaClient {
	return &OllamaClient{
		client: client,
	}
}

// Generate sends a request to Ollama and returns the complete response
func (o *OllamaClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	ollamaReq := &api.GenerateRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		System: req.System,
	}

	var result strings.Builder
	var done bool

	err := o.client.Generate(ctx, ollamaReq, func(resp api.GenerateResponse) error {
		result.WriteString(resp.Response)
		done = resp.Done
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &GenerateResponse{
		Text: result.String(),
		Done: done,
	}, nil
}

// GenerateStream sends a request to Ollama and streams responses through the callback
func (o *OllamaClient) GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error {
	ollamaReq := &api.GenerateRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		System: req.System,
	}

	return o.client.Generate(ctx, ollamaReq, func(resp api.GenerateResponse) error {
		// Only send non-empty responses
		if resp.Response != "" || resp.Done {
			return fn(&GenerateResponse{
				Text: resp.Response,
				Done: resp.Done,
			})
		}
		return nil
	})
}
