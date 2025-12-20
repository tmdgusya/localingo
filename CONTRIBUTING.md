# Contributing to Localingo

Welcome to **Localingo**! We're building a personalized language observatory that merges linguistic pedagogy with data science.

Whether you're an English professor, a Go developer, a Data Analyst, or just a passionate learner, your contributions are welcome.

## 🧠 Our Philosophy
Before contributing, please read [PERSONA.md](PERSONA.md). This document defines the "Heart" of our application.
- We don't just fix errors; we analyze habits.
- We treat mistakes as "fossils" that need to be excavated and understood.
- We value **structured data** over unstructured text.

## 🛠️ Development Setup

1.  **Prerequisites:**
    -   Go 1.21+
    -   Docker & Docker Compose
    -   Ollama (Local LLM)

2.  **Environment:**
    -   Copy `.env.example` to `.env`.
    -   Run `make docker-up` to start PostgreSQL.

3.  **Running Tests:**
    -   We follow a strict **Test-Driven Development (TDD)** approach.
    -   Run all tests: `make test`

## 🏗️ Project Structure

-   `cmd/tui`: The Bubble Tea frontend application.
-   `internal/agent`: Logic for interacting with LLMs (Ollama).
-   `internal/repository`: Database access (using `pgx` and `sqlc`).
-   `internal/srs`: Spaced Repetition System logic (SM-2 Algorithm).
-   `internal/prompt`: Externalized prompt templates.

## 🤝 Contribution Guidelines

### 1. Code Style
-   Follow standard Go conventions (`go fmt`).
-   Use `make build` to ensure everything compiles.

### 2. Adding Features
-   **TUI Components:** Follow the "Atomic Design" in `internal/tui`. Use `helper`, `input`, etc., as examples.
-   **Database:** If you change the schema, create a new migration in `db/migrations` and update `sqlc` queries.
-   **Prompts:** Do NOT hardcode prompts in Go. Add them to `internal/prompt/templates` so they can be managed easily.

### 3. Commit Messages
-   Be descriptive.
-   Example: `feat(srs): implement SM-2 algorithm for review scheduling`

## 🧪 Reporting Bugs
-   Please check existing issues first.
-   Describe the bug and steps to reproduce.
-   Attach `debug.log` if relevant (but remove sensitive info).

Thank you for helping us build the best language learning tool!
