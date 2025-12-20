package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/prompt"
)

// Updated Mock to support mocking Generate
type mockLLMClient struct {
	generateFunc       func(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error)
	generateStreamFunc func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error
}

func (m *mockLLMClient) Generate(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &agent.GenerateResponse{Text: "default mock response", Done: true}, nil
}

func (m *mockLLMClient) GenerateStream(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
	if m.generateStreamFunc != nil {
		return m.generateStreamFunc(ctx, req, fn)
	}
	return nil
}

func TestHandleRephraseStream_Success(t *testing.T) {
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
			// Return valid JSON so the agent parses it correctly
			return &agent.GenerateResponse{
				Text: `{"rephrased": "Hello World!", "analysis": {"categories": ["Grammar"], "explanation": "Fixed typo"}}`,
				Done: true,
			}, nil
		},
	}

	pm, _ := prompt.NewManager()
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient, pm)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	reqBody := RephraseRequest{
		Model:  "test-model",
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "Hello World!") {
		t.Error("Expected rephrased text in response")
	}
}

func TestHandleRephraseStream_ContextTimeout(t *testing.T) {
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
			// Simulate slow response
			select {
			case <-time.After(200 * time.Millisecond):
				return &agent.GenerateResponse{Text: "slow", Done: true}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	pm, _ := prompt.NewManager()
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient, pm)
	config := &RouterConfig{
		AllowOrigin:   "*",
		StreamTimeout: 50 * time.Millisecond,
		DefaultModel:  "test-model",
	}
	router := NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, config)

	reqBody := RephraseRequest{Prompt: "test"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if !strings.Contains(w.Body.String(), "error") {
		t.Error("Expected timeout error")
	}
}

func TestHandleRephraseStream_StreamError(t *testing.T) {
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
			return nil, errors.New("generation failed")
		},
	}

	pm, _ := prompt.NewManager()
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient, pm)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	reqBody := RephraseRequest{Prompt: "test"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if !strings.Contains(w.Body.String(), "generation failed") {
		t.Error("Expected generation failed error")
	}
}

func TestHandleRephraseStream_DefaultModel(t *testing.T) {
	var capturedModel string
	mockClient := &mockLLMClient{
		generateFunc: func(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
			capturedModel = req.Model
			return &agent.GenerateResponse{Text: "{}", Done: true}, nil
		},
	}

	pm, _ := prompt.NewManager()
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient, pm)
	config := &RouterConfig{DefaultModel: "custom-model", StreamTimeout: 5 * time.Minute}
	router := NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, config)

	reqBody := RephraseRequest{Prompt: "test"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if capturedModel != "custom-model" {
		t.Errorf("Expected model 'custom-model', got '%s'", capturedModel)
	}
}

func TestDefaultRouterConfig(t *testing.T) {
	config := DefaultRouterConfig()
	if config.DefaultModel != "llama3.2:latest" {
		t.Error("Default model mismatch")
	}
}
