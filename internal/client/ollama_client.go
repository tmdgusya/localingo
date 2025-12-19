package client

import (
	"context"
	"time"

	"github.com/ollama/ollama/api"
)

func CreateOllamaClient() (*api.Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// VerifyConnection checks if the client can connect to the Ollama server
func VerifyConnection(client *api.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to list models as a simple health check
	_, err := client.List(ctx)

	return err
}
