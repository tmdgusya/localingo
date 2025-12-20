# TUI Architecture & Component Design

This document outlines the architectural guidelines for building composable Terminal UI (TUI) components using Bubble Tea for Localingo.

## 1. Architectural Philosophy

We follow the **Atomic Design** principle adapted for the Elm Architecture (Model-View-Update):

1.  **Atoms (Base Styles & Primitives):** Colors, Borders, basic Lipgloss styles.
2.  **Molecules (Components):** specific functional units (e.g., a Text Input with a border, a Chat Bubble).
3.  **Organisms (Widgets):** Complex compositions (e.g., A Chat Viewport containing many Chat Bubbles, A Form containing multiple Inputs).
4.  **Templates (Screens):** Full-page views (e.g., The Main Chat Screen).

## 2. Directory Structure

```
internal/tui/
├── style/          # Centralized Lipgloss styles (Theming)
├── components/     # Reusable UI components
│   ├── input/      # User input handling
│   ├── chatview/   # Message history display
│   └── history/    # (New) Conversation history list
├── screens/        # Full-page compositions (optional for now, can stay in root)
├── model.go        # Main entry point orchestration
└── msg.go          # Global message types
```

## 3. Component Contract

Every component should adhere to the standard Bubble Tea interface but should be strictly isolated.

### Pattern

```go
// Model holds the state for this specific component
type Model struct {
    State    State
    Styles   Styles
    // ... embedded bubble models (e.g. textinput.Model)
}

// New creates the component with default configuration
func New() Model { ... }

// Init initializes the component (cursor blink, etc)
func (m Model) Init() tea.Cmd { ... }

// Update handles messages specific to this component.
// It should IGNORE messages meant for other components.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { ... }

// View renders the component as a string
func (m Model) View() string { ... }
```

## 4. Specific Component Designs

### A. Style Package (`internal/tui/style`)
*   **Responsibility:** Export `lipgloss.Style` definitions.

### B. ChatInput Component (`internal/tui/components/input`)
*   **Responsibility:** Manage focus, Handle "Enter".
*   **Events:** Returns `SendRequestedMsg`.

### C. ChatViewport Component (`internal/tui/components/chatview`)
*   **Responsibility:** Render messages, Auto-scroll.
*   **Inputs:** `AddMessage(role, text)` method.

### D. History Component (`internal/tui/components/history`) **(New)**
*   **Wraps:** `github.com/charmbracelet/bubbles/list`
*   **Responsibility:**
    *   Fetch and display a list of past conversations.
    *   Allow navigation (Up/Down) and Selection (Enter).
    *   Filtering (built-in by bubbles/list).
*   **Events:** Returns `ConversationSelectedMsg{ID: uuid.UUID}`.

## 5. Main Model Composition & Command System

The `MainModel` acts as the Router/Controller.

### State Management
*   `viewState`: Enum (`ViewChat`, `ViewHistory`)
*   `currentConversationID`: `uuid.UUID` (Tracks active session)

### Command Handling
The `MainModel` intercepts `SendRequestedMsg` from `ChatInput`.

1.  **Command Detection:** If text starts with `/`, parse it.
    *   `/history`: Switch `viewState` to `ViewHistory`, trigger `repo.ListConversations`.
    *   `/new`: Clear `ChatView`, reset `currentConversationID`.
    *   `/quit`: Exit.
2.  **Regular Chat:**
    *   If `currentConversationID` is nil, create a new conversation first (or lazy create).
    *   Append to `ChatView`.
    *   Call Agent.

### Flow
1.  **User types `/history`** -> `MainModel` hides `ChatView`/`Input`, shows `HistoryView`, calls `Cmd` to fetch data.
2.  **Data Loaded** -> `HistoryView` updates items.
3.  **User selects item** -> `MainModel` receives `ConversationSelectedMsg`.
4.  **Transition** -> `MainModel` switches to `ViewChat`, calls `Cmd` to fetch messages for that ID.
5.  **Messages Loaded** -> `ChatView` is populated.