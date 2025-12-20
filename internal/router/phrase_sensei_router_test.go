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
)

type mockLLMClient struct {
	generateStreamFunc func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error
}

func (m *mockLLMClient) Generate(ctx context.Context, req *agent.GenerateRequest) (*agent.GenerateResponse, error) {
	return &agent.GenerateResponse{Text: "test", Done: true}, nil
}

func (m *mockLLMClient) GenerateStream(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
	if m.generateStreamFunc != nil {
		return m.generateStreamFunc(ctx, req, fn)
	}
	return nil
}

func TestHandleRephraseStream_Success(t *testing.T) {
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			// Simulate streaming responses
			responses := []string{"Hello", " ", "World", "!"}
			for i, text := range responses {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				err := fn(&agent.GenerateResponse{
					Text: text,
					Done: i == len(responses)-1,
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
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

	// Check SSE headers
	headers := w.Header()
	if ct := headers.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", ct)
	}
	if cc := headers.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cc)
	}
	if conn := headers.Get("Connection"); conn != "keep-alive" {
		t.Errorf("Expected Connection 'keep-alive', got '%s'", conn)
	}
	if xab := headers.Get("X-Accel-Buffering"); xab != "no" {
		t.Errorf("Expected X-Accel-Buffering 'no', got '%s'", xab)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "data:") {
		t.Error("Expected SSE format with 'data:' prefix")
	}
}

func TestHandleRephraseStream_CORS(t *testing.T) {
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			return fn(&agent.GenerateResponse{Text: "test", Done: true})
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	config := &RouterConfig{
		AllowOrigin:      "https://example.com",
		AllowCredentials: true,
		StreamTimeout:    5 * time.Minute,
		DefaultModel:     "test-model",
		Debug:            false,
	}
	router := NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, config)

	reqBody := RephraseRequest{
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	// Check CORS headers
	headers := w.Header()
	if origin := headers.Get("Access-Control-Allow-Origin"); origin != "https://example.com" {
		t.Errorf("Expected CORS origin 'https://example.com', got '%s'", origin)
	}
	if creds := headers.Get("Access-Control-Allow-Credentials"); creds != "true" {
		t.Errorf("Expected Allow-Credentials 'true', got '%s'", creds)
	}
}

func TestHandleRephraseStream_InvalidJSON(t *testing.T) {
	mockClient := &mockLLMClient{}
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	// Check error response format
	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	if errResp.Message != "Invalid request body" {
		t.Errorf("Expected error message 'Invalid request body', got '%s'", errResp.Message)
	}
}

func TestHandleRephraseStream_MissingPrompt(t *testing.T) {
	mockClient := &mockLLMClient{}
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	reqBody := RephraseRequest{
		Model: "test-model",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	if errResp.Message != "Prompt is required" {
		t.Errorf("Expected error message 'Prompt is required', got '%s'", errResp.Message)
	}
}

func TestHandleRephraseStream_ContextTimeout(t *testing.T) {
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			// Simulate slow response
			select {
			case <-time.After(200 * time.Millisecond):
				return fn(&agent.GenerateResponse{Text: "slow", Done: true})
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	config := &RouterConfig{
		AllowOrigin:   "*",
		StreamTimeout: 50 * time.Millisecond, // Very short timeout
		DefaultModel:  "test-model",
		Debug:         false,
	}
	router := NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, config)

	reqBody := RephraseRequest{
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "error") {
		t.Error("Expected timeout error in response")
	}
	if !strings.Contains(bodyStr, "timeout") {
		t.Errorf("Expected 'timeout' in error message, got: %s", bodyStr)
	}
}

func TestHandleRephraseStream_ContextCancellation(t *testing.T) {
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			// Check if context is cancelled immediately
			<-ctx.Done()
			return ctx.Err()
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	reqBody := RephraseRequest{
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "error") && !strings.Contains(bodyStr, "cancel") {
		t.Errorf("Expected cancellation error in response, got: %s", bodyStr)
	}
}

func TestHandleRephraseStream_StreamError(t *testing.T) {
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			// Send one response, then error
			fn(&agent.GenerateResponse{Text: "partial", Done: false})
			return errors.New("stream failed")
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	router := NewPhraseSenseiRouter(phraseSenseiAgent)

	reqBody := RephraseRequest{
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "error") {
		t.Error("Expected error event in response")
	}
	if !strings.Contains(bodyStr, "stream failed") {
		t.Errorf("Expected error message 'stream failed', got: %s", bodyStr)
	}
}

func TestHandleRephraseStream_DefaultModel(t *testing.T) {
	var capturedModel string
	mockClient := &mockLLMClient{
		generateStreamFunc: func(ctx context.Context, req *agent.GenerateRequest, fn func(*agent.GenerateResponse) error) error {
			capturedModel = req.Model
			return fn(&agent.GenerateResponse{Text: "test", Done: true})
		},
	}

	phraseSenseiAgent := agent.NewPhraseSenseiAgent(mockClient)
	config := &RouterConfig{
		DefaultModel:  "custom-default-model",
		StreamTimeout: 5 * time.Minute,
		AllowOrigin:   "*",
	}
	router := NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, config)

	// Request without model
	reqBody := RephraseRequest{
		Prompt: "test prompt",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/rephrase/stream", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.HandleRephraseStream(w, req)

	if capturedModel != "custom-default-model" {
		t.Errorf("Expected default model 'custom-default-model', got '%s'", capturedModel)
	}
}

func TestDefaultRouterConfig(t *testing.T) {
	config := DefaultRouterConfig()

	if config.AllowOrigin != "*" {
		t.Errorf("Expected default AllowOrigin '*', got '%s'", config.AllowOrigin)
	}
	if config.StreamTimeout != 5*time.Minute {
		t.Errorf("Expected default timeout 5m, got %v", config.StreamTimeout)
	}
	if config.DefaultModel != "llama3.2:latest" {
		t.Errorf("Expected default model 'llama3.2:latest', got '%s'", config.DefaultModel)
	}
	if config.Debug {
		t.Error("Expected debug to be false by default")
	}
}
