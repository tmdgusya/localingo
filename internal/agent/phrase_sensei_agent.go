package agent

import (
	"context"
)

type PhraseSenseiAgent struct {
	client LLMClient
}

func NewPhraseSenseiAgent(client LLMClient) *PhraseSenseiAgent {
	return &PhraseSenseiAgent{
		client: client,
	}
}

// Rephrase a given text using the PhraseSenseiAgent.
func (a *PhraseSenseiAgent) Rephrase(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	return a.client.Generate(ctx, req)
}

// RephraseStream rephrases text and streams the response through a callback.
func (a *PhraseSenseiAgent) RephraseStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error {
	return a.client.GenerateStream(ctx, req, fn)
}
