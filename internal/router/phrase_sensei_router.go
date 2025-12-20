package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/repository"
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

	// Chat history settings
	ChatHistoryEnabled bool
	ChatHistoryRepo    repository.ChatHistoryRepository
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
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	System         string  `json:"system,omitempty"`
	ConversationID *string `json:"conversation_id,omitempty"`
	SaveHistory    *bool   `json:"save_history,omitempty"`
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

	// Determine if history should be saved
	shouldSaveHistory := h.shouldSaveHistory(&req)
	var conversationID uuid.UUID

	// Save user message if history is enabled
	if shouldSaveHistory {
		var err error
		conversationID, err = h.getOrCreateConversation(ctx, &req)
		if err != nil {
			h.logError("Failed to get/create conversation", err)
			// Continue without saving - don't fail the request
			shouldSaveHistory = false
		} else {
			_, err = h.saveUserMessage(ctx, conversationID, &req)
			if err != nil {
				h.logError("Failed to save user message", err)
				// Continue without saving - don't fail the request
				shouldSaveHistory = false
			}
		}
	}

	agentReq := &agent.GenerateRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		System: req.System,
	}

	// Channel to communicate errors from goroutine
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Accumulate response for saving
	var responseBuilder strings.Builder
	startTime := time.Now()

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

			// Accumulate response text
			if shouldSaveHistory && resp.Text != "" {
				responseBuilder.WriteString(resp.Text)
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

		// Save assistant message if history is enabled
		if shouldSaveHistory && responseBuilder.Len() > 0 {
			completionTime := int(time.Since(startTime).Milliseconds())
			if err := h.saveAssistantMessage(ctx, conversationID, responseBuilder.String(), req.Model, completionTime); err != nil {
				h.logError("Failed to save assistant message", err)
				// Don't fail the request
			}
		}

		return
	}
}

// shouldSaveHistory determines if chat history should be saved
func (h *PhraseSenseiRouter) shouldSaveHistory(req *RephraseRequest) bool {
	if !h.config.ChatHistoryEnabled || h.config.ChatHistoryRepo == nil {
		return false
	}

	// Check per-request override
	if req.SaveHistory != nil {
		return *req.SaveHistory
	}

	// Default to enabled if configured
	return true
}

// getOrCreateConversation gets an existing conversation or creates a new one
func (h *PhraseSenseiRouter) getOrCreateConversation(ctx context.Context, req *RephraseRequest) (uuid.UUID, error) {
	if req.ConversationID != nil && *req.ConversationID != "" {
		// Try to parse existing conversation ID
		convID, err := uuid.Parse(*req.ConversationID)
		if err != nil {
			h.logError("Invalid conversation ID", err)
			// Create new conversation if ID is invalid
		} else {
			// Verify conversation exists
			_, err := h.config.ChatHistoryRepo.GetConversation(ctx, convID)
			if err == nil {
				return convID, nil
			}
			h.logError("Conversation not found, creating new one", err)
		}
	}

	// Create new conversation
	conv, err := h.config.ChatHistoryRepo.CreateConversation(ctx, repository.CreateConversationParams{
		Title:    nil, // Could be auto-generated from first prompt
		Metadata: make(map[string]interface{}),
	})
	if err != nil {
		return uuid.Nil, err
	}

	return conv.ID, nil
}

// saveUserMessage saves the user's message to the database
func (h *PhraseSenseiRouter) saveUserMessage(ctx context.Context, conversationID uuid.UUID, req *RephraseRequest) (uuid.UUID, error) {
	msg, err := h.config.ChatHistoryRepo.CreateMessage(ctx, repository.CreateMessageParams{
		ConversationID: conversationID,
		Role:           repository.MessageRoleUser,
		Content:        req.Prompt,
		Model:          &req.Model,
		Metadata:       make(map[string]interface{}),
	})
	if err != nil {
		return uuid.Nil, err
	}

	return msg.ID, nil
}

// saveAssistantMessage saves the assistant's response to the database
func (h *PhraseSenseiRouter) saveAssistantMessage(ctx context.Context, conversationID uuid.UUID, content string, model string, completionTimeMs int) error {
	_, err := h.config.ChatHistoryRepo.CreateMessage(ctx, repository.CreateMessageParams{
		ConversationID:   conversationID,
		Role:             repository.MessageRoleAssistant,
		Content:          content,
		Model:            &model,
		Metadata:         make(map[string]interface{}),
		CompletionTimeMs: &completionTimeMs,
	})
	return err
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
