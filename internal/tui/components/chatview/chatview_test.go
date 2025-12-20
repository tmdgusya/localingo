package chatview

import (
	"strings"
	"testing"
)

func TestChatViewModel(t *testing.T) {
	m := New()

	// 1. Initial state
	if m.View() != "" {
		t.Error("Expected empty view initially")
	}

	// Set dimensions so viewport renders
	m.SetDimensions(20, 10)

	// 2. Add User Message
	m.AddMessage("user", "Hello World")
	view := m.View()
	if !strings.Contains(view, "You:") {
		t.Error("Expected 'You:' label in view")
	}
	if !strings.Contains(view, "Hello World") {
		t.Error("Expected message content in view")
	}

	// 3. Add Sensei Message
	m.AddMessage("sensei", "Welcome")
	view = m.View()
	if !strings.Contains(view, "Sensei:") {
		t.Error("Expected 'Sensei:' label in view")
	}
}
