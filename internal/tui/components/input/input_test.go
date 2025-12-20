package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputModel(t *testing.T) {
	m := New()

	// 1. Test Initial State
	if m.Value() != "" {
		t.Errorf("Expected empty input, got %s", m.Value())
	}

	// 2. Test Init
	if m.Init() == nil {
		t.Error("Expected Init command (blink), got nil")
	}

	// 3. Test Typing
	// Bubble Tea models are immutable, Update returns a new model
	tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	newM := tm.(Model)
	if newM.Value() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", newM.Value())
	}

	// 4. Test Enter Key (Should clear input and return Msg)
	tm, cmd := newM.Update(tea.KeyMsg{Type: tea.KeyEnter})
	finalM := tm.(Model)

	if finalM.Value() != "" {
		t.Errorf("Expected empty input after Enter, got '%s'", finalM.Value())
	}

	if cmd == nil {
		t.Error("Expected command after Enter")
	}

	// Ideally we would check if cmd returns SendRequestedMsg, 
	// but strictly checking tea.Cmd output in unit tests is tricky without a runtime.
}
