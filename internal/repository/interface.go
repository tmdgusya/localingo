package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChatHistoryRepository defines the interface for chat history persistence
type ChatHistoryRepository interface {
	// Conversation operations
	CreateConversation(ctx context.Context, params CreateConversationParams) (*Conversation, error)
	GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error)
	ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error)
	UpdateConversation(ctx context.Context, params UpdateConversationParams) (*Conversation, error)
	DeleteConversation(ctx context.Context, id uuid.UUID) error

	// Message operations
	CreateMessage(ctx context.Context, params CreateMessageParams) (*Message, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*Message, error)
	ListMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*Message, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error

	// Transaction support
	WithTx(ctx context.Context, fn func(repo ChatHistoryRepository) error) error

	// Healthcheck
	Ping(ctx context.Context) error

	// Analysis operations
	GetErrorStats(ctx context.Context) (*ErrorStats, error)
	GetCorrectionPairs(ctx context.Context, limit int) ([]CorrectionPair, error)
	
	// SRS operations
	GetDueReviewPairs(ctx context.Context, limit int) ([]CorrectionPair, error)
	UpdateSRSStatus(ctx context.Context, id uuid.UUID, grade int) error
}

// CorrectionPair represents a user input and its correction
type CorrectionPair struct {
	ID        uuid.UUID // Message ID of the assistant's correction
	Original  string
	Corrected string
	Analysis  []byte
}

// ErrorStats holds aggregated statistics about user errors
type ErrorStats struct {
	CategoryCount map[string]int
	TotalAnalyzed int
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID
	Title     *string
	CreatedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]interface{}
}

// Message represents a single message in a conversation
type Message struct {
	ID               uuid.UUID
	ConversationID   uuid.UUID
	Role             MessageRole
	Content          string
	Model            *string
	CreatedAt        time.Time
	Metadata         map[string]interface{}
	TokenCount       *int
	CompletionTimeMs *int
}

// MessageRole defines the role of a message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// CreateConversationParams contains parameters for creating a conversation
type CreateConversationParams struct {
	Title    *string
	Metadata map[string]interface{}
}

// UpdateConversationParams contains parameters for updating a conversation
type UpdateConversationParams struct {
	ID       uuid.UUID
	Title    *string
	Metadata map[string]interface{}
}

// CreateMessageParams contains parameters for creating a message
type CreateMessageParams struct {
	ConversationID   uuid.UUID
	Role             MessageRole
	Content          string
	Model            *string
	Metadata         map[string]interface{}
	TokenCount       *int
	CompletionTimeMs *int
}
