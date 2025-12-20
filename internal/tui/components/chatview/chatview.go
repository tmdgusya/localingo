package chatview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

type Model struct {
	viewport viewport.Model
	messages []string
	ready    bool
}

func New() Model {
	return Model{
		messages: []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return ""
	}
	return m.viewport.View()
}

func (m *Model) SetDimensions(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(width, height)
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = height
	}
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
}

func (m *Model) AddMessage(role, content string) {
	var formatted string
	if role == "user" {
		formatted = fmt.Sprintf("%s %s", style.UserLabel, style.UserMessage.Render(content))
	} else {
		formatted = fmt.Sprintf("%s %s", style.SenseiLabel, style.SenseiMessage.Render(content))
	}

	m.messages = append(m.messages, formatted)
	
	if m.ready {
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	}
}

// AppendToLastMessage appends text to the last message (useful for streaming)
func (m *Model) AppendToLastMessage(text string) {
	if len(m.messages) == 0 {
		return
	}
	m.messages[len(m.messages)-1] += text
	
	if m.ready {
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	}
}
