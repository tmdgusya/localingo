package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ollama/ollama/api"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/router"
)

func main() {
	// Initialize Ollama client
	ollamaURL, _ := url.Parse("http://localhost:11434")
	ollamaAPIClient := api.NewClient(ollamaURL, http.DefaultClient)
	ollamaClient := agent.NewOllamaClient(ollamaAPIClient)

	// Initialize agent
	phraseSenseiAgent := agent.NewPhraseSenseiAgent(ollamaClient)

	// Initialize router
	phraseSenseiRouter := router.NewPhraseSenseiRouter(phraseSenseiAgent)

	// Setup HTTP router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Localingo API Server"))
	})

	// Register endpoints
	r.Post("/api/rephrase/stream", phraseSenseiRouter.HandleRephraseStream)

	log.Println("Server starting on :3000")
	http.ListenAndServe(":3000", r)
}
