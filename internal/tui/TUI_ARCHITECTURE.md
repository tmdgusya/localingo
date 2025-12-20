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
│   └── chatview/   # Message history display
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
*   **Responsibility:** Export `lipgloss.Style` definitions for borders, active colors, text colors, and layout utilities.
*   **Goal:** Changing a color here updates the entire app.

### B. ChatInput Component (`internal/tui/components/input`)
*   **Wraps:** `github.com/charmbracelet/bubbles/textinput`
*   **Responsibility:**
    *   Manage focus state.
    *   Handle "Enter" key to emit a `SendMsg`.
    *   Styling the prompt and input box.
*   **Events:** Returns a custom `SendRequestedMsg` when Enter is pressed, so the parent knows to process text.

### C. ChatViewport Component (`internal/tui/components/chatview`)
*   **Wraps:** `github.com/charmbracelet/bubbles/viewport`
*   **Responsibility:**
    *   Rendering a list of messages.
    *   Distinguishing between `User` (Right aligned, Blue) and `Sensei` (Left aligned, Green) messages.
    *   Auto-scrolling to bottom on new messages.
*   **Inputs:** `AddMessage(role, text)` method.

## 5. Main Model Composition

The root model (`internal/tui/model.go`) will no longer handle low-level rendering. It will coordinate:

```go
type MainModel struct {
    chatView  chatview.Model
    chatInput input.Model
    agent     Agent
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case input.SendRequestedMsg:
        // 1. Get text from input
        // 2. Add to chatView
        // 3. Trigger Agent
    }
    // Propagate standard messages to children
    // ...
}
```
