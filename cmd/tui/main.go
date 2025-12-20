package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ollama/ollama/api"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/config"
	"github.com/tmdgusya/localingo/internal/prompt"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui"
)

func main() {
	// Load configuration from environment
	cfg := config.LoadFromEnv()
	
	// Initialize database if chat history is enabled
	var dbPool *pgxpool.Pool
	var chatRepo repository.ChatHistoryRepository

	if cfg.ChatHistory.Enabled {
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
	}

	// Initialize Ollama client
	ollamaURLStr := cfg.Ollama.Host
	if ollamaURLStr == "" {
		ollamaURLStr = "http://localhost:11434"
	}
	ollamaURL, err := url.Parse(ollamaURLStr)
	if err != nil {
		log.Fatalf("Invalid OLLAMA_HOST: %v", err)
	}

	ollamaAPIClient := api.NewClient(ollamaURL, http.DefaultClient)
	ollamaClient := agent.NewOllamaClient(ollamaAPIClient)

	// Initialize Prompt Manager
	promptManager, err := prompt.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize prompt manager: %v", err)
	}

	// Initialize agent
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(ollamaClient, promptManager)

	// Enable logging to file
	if f, err := tea.LogToFile("debug.log", "debug"); err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	} else {
		defer f.Close()
	}

	// Initialize TUI
	p := tea.NewProgram(
		tui.NewModel(phraseSenseiAgent, chatRepo, cfg.Ollama.DefaultModel),
		tea.WithAltScreen(),       // Use full screen
		tea.WithMouseCellMotion(), // Turn on mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
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
