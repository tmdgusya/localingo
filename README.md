# Localingo 🌍

> **"I don't just correct you; I help you fossilize the *right* patterns."** - Phrase Sensei

Localingo is not just a translation tool. It is a personalized **Language Observatory** and **Habit-Breaking Laboratory**. It combines the linguistic intuition of an English Professor with the rigorous pattern recognition of a Data Analyst.

## 🚀 Features

### 1. 🤖 Intelligent Chat & Diagnosis
Chat naturally with **Phrase Sensei**. Unlike standard AI, Sensei doesn't just reply; it analyzes your input in real-time.
-   **Structured Analysis:** Every sentence you type is tagged (Grammar, Nuance, Vocabulary, etc.).
-   **Context-Aware Rephrasing:** Receive corrections that sound native, not robotic.

### 2. 📊 Pattern Analyst (`/report`)
Stop guessing what you're doing wrong.
-   **Data-Driven Insights:** View a statistical breakdown of your most frequent error types.
-   **Visualization:** See bar charts of your linguistic weak points directly in the terminal.

### 3. 👩‍🏫 Private Tutor (`/lesson`)
Receive a personalized "One-Point Lesson" based on your recent history.
-   **Meta-Analysis:** Sensei reads your last 20 corrections to find recurring habits.
-   **Native Mindset:** Learn *why* you make mistakes and how a native speaker thinks.

### 4. 🏋️ Fossil Breaker (`/review`)
Break bad habits with science.
-   **SRS (Spaced Repetition System):** Uses the **SM-2 Algorithm** to schedule reviews based on your memory curve.
-   **Active Recall:** Quiz yourself on your own past mistakes.
-   **Adaptive:** If you get it right, you'll see it less often. If you fail, you'll see it tomorrow.

---

## 🛠️ Installation & Setup

### Prerequisites
-   **Go 1.21+**
-   **Docker & Docker Compose** (for PostgreSQL)
-   **Ollama** (running locally with `qwen3:4b` or similar)

### 1. Clone & Setup
```bash
git clone https://github.com/tmdgusya/localingo.git
cd localingo
```

### 2. Database (PostgreSQL)
Start the database container:
```bash
make docker-up
```
*Note: The app will automatically handle migrations.*

### 3. AI Engine (Ollama)
Ensure Ollama is running and the model is pulled:
```bash
ollama pull qwen3:4b
```
*(You can configure the model in `.env` or environment variables)*

### 4. Build
Compile both the API Server and the Terminal UI:
```bash
make build
```

---

## 🖥️ Usage

### Running the TUI (Recommended)
The Terminal User Interface is the main way to interact with Localingo.

```bash
make tui
# or directly:
./bin/tui
```

### Commands
Type `/` in the chat input to see the helper menu.

| Command | Description |
| :--- | :--- |
| **`/new`** | Start a fresh conversation context. |
| **`/history`** | Browse and load past conversations. |
| **`/report`** | View your linguistic error statistics. |
| **`/lesson`** | Generate a personalized lesson based on recent chats. |
| **`/review`** | Start an SRS quiz session to fix old mistakes. |
| **`/quit`** | Exit the application. |

---

## 🏗️ Architecture

-   **Backend:** Go (Chi Router), PostgreSQL (pgx), Ollama API.
-   **Frontend:** Bubble Tea (TUI Framework).
-   **Core Philosophy:** See [PERSONA.md](PERSONA.md) for our detailed design philosophy.

---
*Localingo helps you sound like a native, one habit at a time.*
