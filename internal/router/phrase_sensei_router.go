package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tmdgusya/localingo/internal/agent"
)

type RouterConfig struct {
	// CORS settings
	AllowOrigin      string
	AllowCredentials bool

	// Timeout settings
	StreamTimeout time.Duration

	// Default model
	DefaultModel string

	// Enable debug logging
	Debug bool
}

func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		AllowOrigin:      "*",
		AllowCredentials: false,
		StreamTimeout:    5 * time.Minute,
		DefaultModel:     "llama3.2:latest",
		Debug:            false,
	}
}

type PhraseSenseiRouter struct {
	agent  *agent.PhraseSenseiAgent
	config *RouterConfig
}

func NewPhraseSenseiRouter(agent *agent.PhraseSenseiAgent) *PhraseSenseiRouter {
	return &PhraseSenseiRouter{
		agent:  agent,
		config: DefaultRouterConfig(),
	}
}

func NewPhraseSenseiRouterWithConfig(agent *agent.PhraseSenseiAgent, config *RouterConfig) *PhraseSenseiRouter {
	if config == nil {
		config = DefaultRouterConfig()
	}
	return &PhraseSenseiRouter{
		agent:  agent,
		config: config,
	}
}

type RephraseRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HandleRephraseStream handles streaming rephrase requests with production-grade error handling
func (h *PhraseSenseiRouter) HandleRephraseStream(w http.ResponseWriter, r *http.Request) {
	// Parse and validate request
	var req RephraseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logError("Failed to decode request body", err)
		h.sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		h.sendJSONError(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = h.config.DefaultModel
	}

	// Set SSE headers
	h.setSSEHeaders(w)

	// Check if streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logError("Streaming not supported", nil)
		h.sendJSONError(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.StreamTimeout)
	defer cancel()

	agentReq := &agent.GenerateRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		System: req.System,
	}

	// Channel to communicate errors from goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)

		err := h.agent.RephraseStream(ctx, agentReq, func(resp *agent.GenerateResponse) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				h.logInfo("Stream cancelled by context (client disconnected or timeout)")
				return ctx.Err()
			default:
			}

			// Send data as SSE
			data, marshalErr := json.Marshal(resp)
			if marshalErr != nil {
				h.logError("Failed to marshal response", marshalErr)
				return marshalErr
			}

			if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", data); writeErr != nil {
				h.logError("Failed to write response", writeErr)
				return writeErr
			}

			flusher.Flush()
			return nil
		})

		if err != nil {
			errChan <- err
		}
	}()

	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			h.logInfo("Request timeout")
			h.sendSSEError(w, "Request timeout", flusher)
		case context.Canceled:
			h.logInfo("Request cancelled by client")
			h.sendSSEError(w, "Request cancelled", flusher)
		}
		return

	case err := <-errChan:
		h.logError("Stream error", err)
		h.sendSSEError(w, err.Error(), flusher)
		return

	case <-doneChan:
		h.logDebug("Stream completed successfully")
		return
	}
}

// setSSEHeaders sets all necessary headers for Server-Sent Events
func (h *PhraseSenseiRouter) setSSEHeaders(w http.ResponseWriter) {
	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Disable buffering in nginx proxy
	w.Header().Set("X-Accel-Buffering", "no")

	// CORS headers
	if h.config.AllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.config.AllowOrigin)
	}
	if h.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// sendJSONError sends a JSON error response
func (h *PhraseSenseiRouter) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// sendSSEError sends an error event in SSE format
func (h *PhraseSenseiRouter) sendSSEError(w http.ResponseWriter, message string, flusher http.Flusher) {
	errorData, _ := json.Marshal(ErrorResponse{
		Error:   "StreamError",
		Message: message,
	})
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
	flusher.Flush()
}

// logError logs an error message
func (h *PhraseSenseiRouter) logError(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	} else {
		log.Printf("[ERROR] %s", message)
	}
}

// logInfo logs an info message
func (h *PhraseSenseiRouter) logInfo(message string) {
	log.Printf("[INFO] %s", message)
}

// logDebug logs a debug message
func (h *PhraseSenseiRouter) logDebug(message string) {
	if h.config.Debug {
		log.Printf("[DEBUG] %s", message)
	}
}
