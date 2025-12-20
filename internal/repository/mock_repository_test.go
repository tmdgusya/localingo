package repository

import (
	"context"

	"github.com/google/uuid"
)

// MockChatHistoryRepository is a mock implementation for testing
type MockChatHistoryRepository struct {
	CreateConversationFunc        func(ctx context.Context, params CreateConversationParams) (*Conversation, error)
	GetConversationFunc           func(ctx context.Context, id uuid.UUID) (*Conversation, error)
	ListConversationsFunc         func(ctx context.Context, limit, offset int) ([]*Conversation, error)
	UpdateConversationFunc        func(ctx context.Context, params UpdateConversationParams) (*Conversation, error)
	DeleteConversationFunc        func(ctx context.Context, id uuid.UUID) error
	CreateMessageFunc             func(ctx context.Context, params CreateMessageParams) (*Message, error)
	GetMessageFunc                func(ctx context.Context, id uuid.UUID) (*Message, error)
	ListMessagesByConversationFunc func(ctx context.Context, conversationID uuid.UUID) ([]*Message, error)
	DeleteMessageFunc             func(ctx context.Context, id uuid.UUID) error
	WithTxFunc                    func(ctx context.Context, fn func(repo ChatHistoryRepository) error) error
	PingFunc                      func(ctx context.Context) error
	GetErrorStatsFunc             func(ctx context.Context) (*ErrorStats, error)
}

// CreateConversation mocks conversation creation
func (m *MockChatHistoryRepository) CreateConversation(ctx context.Context, params CreateConversationParams) (*Conversation, error) {
	if m.CreateConversationFunc != nil {
		return m.CreateConversationFunc(ctx, params)
	}
	return nil, nil
}

// GetConversation mocks getting a conversation
func (m *MockChatHistoryRepository) GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	if m.GetConversationFunc != nil {
		return m.GetConversationFunc(ctx, id)
	}
	return nil, nil
}

// ListConversations mocks listing conversations
func (m *MockChatHistoryRepository) ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	if m.ListConversationsFunc != nil {
		return m.ListConversationsFunc(ctx, limit, offset)
	}
	return nil, nil
}

// UpdateConversation mocks updating a conversation
func (m *MockChatHistoryRepository) UpdateConversation(ctx context.Context, params UpdateConversationParams) (*Conversation, error) {
	if m.UpdateConversationFunc != nil {
		return m.UpdateConversationFunc(ctx, params)
	}
	return nil, nil
}

// DeleteConversation mocks deleting a conversation
func (m *MockChatHistoryRepository) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	if m.DeleteConversationFunc != nil {
		return m.DeleteConversationFunc(ctx, id)
	}
	return nil
}

// CreateMessage mocks creating a message
func (m *MockChatHistoryRepository) CreateMessage(ctx context.Context, params CreateMessageParams) (*Message, error) {
	if m.CreateMessageFunc != nil {
		return m.CreateMessageFunc(ctx, params)
	}
	return nil, nil
}

// GetMessage mocks getting a message
func (m *MockChatHistoryRepository) GetMessage(ctx context.Context, id uuid.UUID) (*Message, error) {
	if m.GetMessageFunc != nil {
		return m.GetMessageFunc(ctx, id)
	}
	return nil, nil
}

// ListMessagesByConversation mocks listing messages by conversation
func (m *MockChatHistoryRepository) ListMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*Message, error) {
	if m.ListMessagesByConversationFunc != nil {
		return m.ListMessagesByConversationFunc(ctx, conversationID)
	}
	return nil, nil
}

// DeleteMessage mocks deleting a message
func (m *MockChatHistoryRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, id)
	}
	return nil
}

// WithTx mocks transaction execution
func (m *MockChatHistoryRepository) WithTx(ctx context.Context, fn func(repo ChatHistoryRepository) error) error {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(ctx, fn)
	}
	return fn(m)
}

// Ping mocks health check
func (m *MockChatHistoryRepository) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

// GetErrorStats mocks getting error statistics
func (m *MockChatHistoryRepository) GetErrorStats(ctx context.Context) (*ErrorStats, error) {
	if m.GetErrorStatsFunc != nil {
		return m.GetErrorStatsFunc(ctx)
	}
	return &ErrorStats{CategoryCount: make(map[string]int)}, nil
}
