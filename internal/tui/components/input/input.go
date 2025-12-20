package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

// SendRequestedMsg is a command that indicates the user wants to send a message
type SendRequestedMsg struct {
	Message string
}

type Model struct {
	textInput textinput.Model
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Type your phrase to rephrase..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 30
	ti.PromptStyle = style.InputPrompt
	
	return Model{
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			val := m.textInput.Value()
			if val == "" {
				return m, nil
			}
			m.textInput.SetValue("")
			return m, func() tea.Msg {
				return SendRequestedMsg{Message: val}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return style.InputBorder.Render(m.textInput.View())
}

func (m Model) Value() string {
	return m.textInput.Value()
}

func (m *Model) SetWidth(w int) {
	m.textInput.Width = w - 4 // Account for border/padding
}

func (m *Model) Focus() tea.Cmd {
	return m.textInput.Focus()
}
