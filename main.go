package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ollama/ollama/api"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/config"
	"github.com/tmdgusya/localingo/internal/prompt"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/router"
)

func main() {
	// Load configuration from environment
	cfg := config.LoadFromEnv()
	log.Println("Configuration loaded")

	// Initialize database if chat history is enabled
	var dbPool *pgxpool.Pool
	var chatRepo repository.ChatHistoryRepository

	if cfg.ChatHistory.Enabled {
		log.Println("Chat history enabled, initializing database...")

		var err error
		dbPool, err = initDatabasePool(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer dbPool.Close()

		chatRepo = repository.NewPostgresRepository(dbPool)

		// Verify database connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := chatRepo.Ping(ctx); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		log.Println("Database connection established successfully")
	} else {
		log.Println("Chat history disabled")
	}

	// Initialize Ollama client
	ollamaURL, _ := url.Parse("http://localhost:11434")
	ollamaAPIClient := api.NewClient(ollamaURL, http.DefaultClient)
	ollamaClient := agent.NewOllamaClient(ollamaAPIClient)

	// Initialize Prompt Manager
	promptManager, err := prompt.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize prompt manager: %v", err)
	}

	// Initialize agent
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(ollamaClient, promptManager)

	// Initialize router with chat history support
	routerConfig := &router.RouterConfig{
		AllowOrigin:        "*",
		AllowCredentials:   false,
		StreamTimeout:      5 * time.Minute,
		DefaultModel:       cfg.Ollama.DefaultModel,
		Debug:              false,
		ChatHistoryEnabled: cfg.ChatHistory.Enabled,
		ChatHistoryRepo:    chatRepo,
	}

	phraseSenseiRouter := router.NewPhraseSenseiRouterWithConfig(phraseSenseiAgent, routerConfig)

	// Setup HTTP router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Localingo API Server"))
	})

	// Register endpoints
	r.Post("/api/rephrase/stream", phraseSenseiRouter.HandleRephraseStream)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Println("Server starting on :3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}

// initDatabasePool creates and configures a PostgreSQL connection pool
func initDatabasePool(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
