package repository_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tmdgusya/localingo/internal/repository"
)

var (
	testPool *pgxpool.Pool
	testRepo repository.ChatHistoryRepository
)

func TestMain(m *testing.M) {
	// Setup Docker PostgreSQL container
	pool, resource, dbPool := setupTestDatabase()
	testPool = dbPool
	testRepo = repository.NewPostgresRepository(dbPool)

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestDatabase(pool, resource, dbPool)

	os.Exit(code)
}

func setupTestDatabase() (*dockertest.Pool, *dockertest.Resource, *pgxpool.Pool) {
	// Create Docker pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// Start PostgreSQL container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "18-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=test_user",
			"POSTGRES_DB=test_db",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://test_user:secret@%s/test_db?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url:", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	var dbPool *pgxpool.Pool

	// Exponential backoff-retry to wait for PostgreSQL to be ready
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		dbPool, err = pgxpool.New(context.Background(), databaseUrl)
		if err != nil {
			return err
		}
		return dbPool.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Run migrations
	ctx := context.Background()
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE conversations (
			id UUID PRIMARY KEY DEFAULT uuidv7(),
			title TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			metadata JSONB DEFAULT '{}'::jsonb
		);

		CREATE INDEX idx_conversations_created_at ON conversations(created_at DESC);
		CREATE INDEX idx_conversations_updated_at ON conversations(updated_at DESC);

		CREATE TABLE messages (
			id UUID PRIMARY KEY DEFAULT uuidv7(),
			conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
			content TEXT NOT NULL,
			model TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			metadata JSONB DEFAULT '{}'::jsonb,
			token_count INTEGER,
			completion_time_ms INTEGER
		);

		CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
		CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
		CREATE INDEX idx_messages_conversation_created ON messages(conversation_id, created_at);
	`)
	if err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	return pool, resource, dbPool
}

func cleanupTestDatabase(pool *dockertest.Pool, resource *dockertest.Resource, dbPool *pgxpool.Pool) {
	if dbPool != nil {
		dbPool.Close()
	}
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func TestCreateConversation(t *testing.T) {
	ctx := context.Background()

	title := "Test Conversation"
	params := repository.CreateConversationParams{
		Title: &title,
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	conv, err := testRepo.CreateConversation(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	if conv.ID == uuid.Nil {
		t.Error("Expected non-nil UUID")
	}

	if conv.Title == nil || *conv.Title != title {
		t.Errorf("Expected title %q, got %v", title, conv.Title)
	}

	if conv.CreatedAt.IsZero() {
		t.Error("Expected non-zero created_at")
	}

	if conv.UpdatedAt.IsZero() {
		t.Error("Expected non-zero updated_at")
	}

	if conv.Metadata == nil {
		t.Error("Expected non-nil metadata")
	}
}

func TestGetConversation(t *testing.T) {
	ctx := context.Background()

	// Create a conversation first
	title := "Get Test Conversation"
	params := repository.CreateConversationParams{
		Title: &title,
	}

	created, err := testRepo.CreateConversation(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Get the conversation
	fetched, err := testRepo.GetConversation(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get conversation: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("Expected ID %v, got %v", created.ID, fetched.ID)
	}

	if fetched.Title == nil || *fetched.Title != title {
		t.Errorf("Expected title %q, got %v", title, fetched.Title)
	}
}

func TestGetConversation_NotFound(t *testing.T) {
	ctx := context.Background()

	nonExistentID := uuid.New()

	_, err := testRepo.GetConversation(ctx, nonExistentID)
	if err == nil {
		t.Fatal("Expected error for non-existent conversation")
	}

	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCreateMessage(t *testing.T) {
	ctx := context.Background()

	// Create a conversation first
	title := "Message Test Conversation"
	conv, err := testRepo.CreateConversation(ctx, repository.CreateConversationParams{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Create a message
	model := "test-model"
	tokenCount := 100
	completionTime := 1500
	params := repository.CreateMessageParams{
		ConversationID:   conv.ID,
		Role:             repository.MessageRoleUser,
		Content:          "Hello, world!",
		Model:            &model,
		TokenCount:       &tokenCount,
		CompletionTimeMs: &completionTime,
		Metadata: map[string]interface{}{
			"test": "data",
		},
	}

	msg, err := testRepo.CreateMessage(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.ID == uuid.Nil {
		t.Error("Expected non-nil UUID")
	}

	if msg.ConversationID != conv.ID {
		t.Errorf("Expected conversation_id %v, got %v", conv.ID, msg.ConversationID)
	}

	if msg.Role != repository.MessageRoleUser {
		t.Errorf("Expected role %q, got %q", repository.MessageRoleUser, msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("Expected content %q, got %q", "Hello, world!", msg.Content)
	}

	if msg.Model == nil || *msg.Model != model {
		t.Errorf("Expected model %q, got %v", model, msg.Model)
	}

	if msg.TokenCount == nil || *msg.TokenCount != tokenCount {
		t.Errorf("Expected token count %d, got %v", tokenCount, msg.TokenCount)
	}

	if msg.CompletionTimeMs == nil || *msg.CompletionTimeMs != completionTime {
		t.Errorf("Expected completion time %d, got %v", completionTime, msg.CompletionTimeMs)
	}

	if msg.CreatedAt.IsZero() {
		t.Error("Expected non-zero created_at")
	}
}

func TestListMessagesByConversation(t *testing.T) {
	ctx := context.Background()

	// Create a conversation
	title := "List Messages Test"
	conv, err := testRepo.CreateConversation(ctx, repository.CreateConversationParams{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// Create multiple messages
	messages := []string{"First message", "Second message", "Third message"}
	for _, content := range messages {
		_, err := testRepo.CreateMessage(ctx, repository.CreateMessageParams{
			ConversationID: conv.ID,
			Role:           repository.MessageRoleUser,
			Content:        content,
		})
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List messages
	fetched, err := testRepo.ListMessagesByConversation(ctx, conv.ID)
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}

	if len(fetched) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(fetched))
	}

	// Verify chronological order
	for i, msg := range fetched {
		if msg.Content != messages[i] {
			t.Errorf("Expected message %d to be %q, got %q", i, messages[i], msg.Content)
		}
	}
}

func TestListMessagesByConversation_Empty(t *testing.T) {
	ctx := context.Background()

	// Create a conversation without messages
	title := "Empty Conversation"
	conv, err := testRepo.CreateConversation(ctx, repository.CreateConversationParams{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("Failed to create conversation: %v", err)
	}

	// List messages
	messages, err := testRepo.ListMessagesByConversation(ctx, conv.ID)
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}

	if messages == nil {
		t.Error("Expected non-nil slice")
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

func TestWithTx_Commit(t *testing.T) {
	ctx := context.Background()

	var createdConvID uuid.UUID

	err := testRepo.WithTx(ctx, func(repo repository.ChatHistoryRepository) error {
		title := "Transaction Test"
		conv, err := repo.CreateConversation(ctx, repository.CreateConversationParams{
			Title: &title,
		})
		if err != nil {
			return err
		}
		createdConvID = conv.ID

		// Create a message in the same transaction
		_, err = repo.CreateMessage(ctx, repository.CreateMessageParams{
			ConversationID: conv.ID,
			Role:           repository.MessageRoleUser,
			Content:        "Test message",
		})
		return err
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// Verify the conversation exists
	conv, err := testRepo.GetConversation(ctx, createdConvID)
	if err != nil {
		t.Fatalf("Failed to get conversation after transaction: %v", err)
	}

	// Verify the message exists
	messages, err := testRepo.ListMessagesByConversation(ctx, conv.ID)
	if err != nil {
		t.Fatalf("Failed to list messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestWithTx_Rollback(t *testing.T) {
	ctx := context.Background()

	var createdConvID uuid.UUID

	err := testRepo.WithTx(ctx, func(repo repository.ChatHistoryRepository) error {
		title := "Rollback Test"
		conv, err := repo.CreateConversation(ctx, repository.CreateConversationParams{
			Title: &title,
		})
		if err != nil {
			return err
		}
		createdConvID = conv.ID

		// Intentionally return an error to trigger rollback
		return fmt.Errorf("intentional error")
	})

	if err == nil {
		t.Fatal("Expected error from transaction")
	}

	// Verify the conversation was not created
	_, err = testRepo.GetConversation(ctx, createdConvID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestPing(t *testing.T) {
	ctx := context.Background()

	err := testRepo.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}
