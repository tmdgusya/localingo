package client

import (
	"context"
	"time"

	"github.com/ollama/ollama/api"
)

// OllamaClient wraps the Ollama API client and implements TextGenerator
type OllamaClient struct {
	client *api.Client
	model  string
}

// NewOllamaClient creates a new OllamaClient from environment
func NewOllamaClient(model string) (*OllamaClient, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

// Generate implements TextGenerator interface
func (o *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	var response string

	req := &api.GenerateRequest{
		Model:  o.model,
		Prompt: prompt,
	}

	err := o.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response += resp.Response
		return nil
	})

	if err != nil {
		return "", err
	}

	return response, nil
}

// VerifyConnection checks if the client can connect to the Ollama server
func (o *OllamaClient) VerifyConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := o.client.List(ctx)
	return err
}
