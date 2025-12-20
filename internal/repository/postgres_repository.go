package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmdgusya/localingo/internal/repository/db"
)

// PostgresRepository implements ChatHistoryRepository using PostgreSQL
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

// CreateConversation creates a new conversation
func (r *PostgresRepository) CreateConversation(ctx context.Context, params CreateConversationParams) (*Conversation, error) {
	metadata, err := json.Marshal(params.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	dbConv, err := r.queries.CreateConversation(ctx, db.CreateConversationParams{
		Title:    params.Title,
		Metadata: metadata,
	})
	if err != nil {
		return nil, WrapDBError(err)
	}

	return r.convertDBConversationToDomain(dbConv)
}

// GetConversation retrieves a conversation by ID
func (r *PostgresRepository) GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(id.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	dbConv, err := r.queries.GetConversation(ctx, pgUUID)
	if err != nil {
		return nil, WrapDBError(err)
	}

	return r.convertDBConversationToDomain(dbConv)
}

// ListConversations retrieves conversations with pagination
func (r *PostgresRepository) ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	dbConvs, err := r.queries.ListConversations(ctx, db.ListConversationsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, WrapDBError(err)
	}

	conversations := make([]*Conversation, len(dbConvs))
	for i, dbConv := range dbConvs {
		conv, err := r.convertDBConversationToDomain(dbConv)
		if err != nil {
			return nil, err
		}
		conversations[i] = conv
	}

	return conversations, nil
}

// UpdateConversation updates a conversation
func (r *PostgresRepository) UpdateConversation(ctx context.Context, params UpdateConversationParams) (*Conversation, error) {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(params.ID.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	metadata, err := json.Marshal(params.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	dbConv, err := r.queries.UpdateConversation(ctx, db.UpdateConversationParams{
		ID:       pgUUID,
		Title:    params.Title,
		Metadata: metadata,
	})
	if err != nil {
		return nil, WrapDBError(err)
	}

	return r.convertDBConversationToDomain(dbConv)
}

// DeleteConversation deletes a conversation
func (r *PostgresRepository) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(id.String()); err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	err := r.queries.DeleteConversation(ctx, pgUUID)
	return WrapDBError(err)
}

// CreateMessage creates a new message
func (r *PostgresRepository) CreateMessage(ctx context.Context, params CreateMessageParams) (*Message, error) {
	pgConvUUID := pgtype.UUID{}
	if err := pgConvUUID.Scan(params.ConversationID.String()); err != nil {
		return nil, fmt.Errorf("invalid conversation UUID: %w", err)
	}

	metadata, err := json.Marshal(params.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var tokenCount *int32
	if params.TokenCount != nil {
		val := int32(*params.TokenCount)
		tokenCount = &val
	}

	var completionTime *int32
	if params.CompletionTimeMs != nil {
		val := int32(*params.CompletionTimeMs)
		completionTime = &val
	}

	dbMsg, err := r.queries.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID:   pgConvUUID,
		Role:             string(params.Role),
		Content:          params.Content,
		Model:            params.Model,
		Metadata:         metadata,
		TokenCount:       tokenCount,
		CompletionTimeMs: completionTime,
	})
	if err != nil {
		return nil, WrapDBError(err)
	}

	return r.convertDBMessageToDomain(dbMsg)
}

// GetMessage retrieves a message by ID
func (r *PostgresRepository) GetMessage(ctx context.Context, id uuid.UUID) (*Message, error) {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(id.String()); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	dbMsg, err := r.queries.GetMessage(ctx, pgUUID)
	if err != nil {
		return nil, WrapDBError(err)
	}

	return r.convertDBMessageToDomain(dbMsg)
}

// ListMessagesByConversation retrieves all messages for a conversation
func (r *PostgresRepository) ListMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*Message, error) {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(conversationID.String()); err != nil {
		return nil, fmt.Errorf("invalid conversation UUID: %w", err)
	}

	dbMsgs, err := r.queries.ListMessagesByConversation(ctx, pgUUID)
	if err != nil {
		return nil, WrapDBError(err)
	}

	messages := make([]*Message, len(dbMsgs))
	for i, dbMsg := range dbMsgs {
		msg, err := r.convertDBMessageToDomain(dbMsg)
		if err != nil {
			return nil, err
		}
		messages[i] = msg
	}

	return messages, nil
}

// DeleteMessage deletes a message
func (r *PostgresRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	pgUUID := pgtype.UUID{}
	if err := pgUUID.Scan(id.String()); err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	err := r.queries.DeleteMessage(ctx, pgUUID)
	return WrapDBError(err)
}

// WithTx executes a function within a transaction
func (r *PostgresRepository) WithTx(ctx context.Context, fn func(repo ChatHistoryRepository) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return WrapDBError(err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			// Log the error but don't override the return error
			fmt.Printf("error rolling back transaction: %v\n", err)
		}
	}()

	txRepo := &PostgresRepository{
		pool:    r.pool,
		queries: r.queries.WithTx(tx),
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return WrapDBError(err)
	}

	return nil
}

// Ping checks the database connection
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return WrapDBError(r.pool.Ping(ctx))
}

// GetErrorStats aggregates error statistics from all messages
func (r *PostgresRepository) GetErrorStats(ctx context.Context) (*ErrorStats, error) {
	stats := &ErrorStats{
		CategoryCount: make(map[string]int),
	}

	// 1. Get category counts
	categoryQuery := `
		SELECT 
			category, 
			count(*)::int as count
		FROM 
			messages, 
			jsonb_array_elements_text(metadata->'analysis'->'categories') as category
		WHERE 
			role = 'assistant' 
			AND metadata->'analysis' IS NOT NULL
		GROUP BY 
			category
		ORDER BY 
			count DESC`

	rows, err := r.pool.Query(ctx, categoryQuery)
	if err != nil {
		return nil, WrapDBError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, WrapDBError(err)
		}
		stats.CategoryCount[category] = count
	}

	// 2. Get total analyzed messages
	totalQuery := `
		SELECT 
			count(*)::int
		FROM 
			messages
		WHERE 
			role = 'assistant' 
			AND metadata->'analysis' IS NOT NULL`

	err = r.pool.QueryRow(ctx, totalQuery).Scan(&stats.TotalAnalyzed)
	if err != nil {
		return nil, WrapDBError(err)
	}

	return stats, nil
}

// Helper functions to convert between DB and domain types

func (r *PostgresRepository) convertDBConversationToDomain(dbConv db.Conversation) (*Conversation, error) {
	id, err := uuid.FromBytes(dbConv.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	createdAt := dbConv.CreatedAt.Time
	updatedAt := dbConv.UpdatedAt.Time

	var metadata map[string]interface{}
	if len(dbConv.Metadata) > 0 {
		if err := json.Unmarshal(dbConv.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Conversation{
		ID:        id,
		Title:     dbConv.Title,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Metadata:  metadata,
	}, nil
}

func (r *PostgresRepository) convertDBMessageToDomain(dbMsg db.Message) (*Message, error) {
	id, err := uuid.FromBytes(dbMsg.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to convert message UUID: %w", err)
	}

	conversationID, err := uuid.FromBytes(dbMsg.ConversationID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to convert conversation UUID: %w", err)
	}

	createdAt := dbMsg.CreatedAt.Time

	var metadata map[string]interface{}
	if len(dbMsg.Metadata) > 0 {
		if err := json.Unmarshal(dbMsg.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	var tokenCount *int
	if dbMsg.TokenCount != nil {
		val := int(*dbMsg.TokenCount)
		tokenCount = &val
	}

	var completionTime *int
	if dbMsg.CompletionTimeMs != nil {
		val := int(*dbMsg.CompletionTimeMs)
		completionTime = &val
	}

	return &Message{
		ID:               id,
		ConversationID:   conversationID,
		Role:             MessageRole(dbMsg.Role),
		Content:          dbMsg.Content,
		Model:            dbMsg.Model,
		CreatedAt:        createdAt,
		Metadata:         metadata,
		TokenCount:       tokenCount,
		CompletionTimeMs: completionTime,
	}, nil
}
