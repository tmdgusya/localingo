package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui/components/chatview"
	"github.com/tmdgusya/localingo/internal/tui/components/input"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

type TuiModel struct {
	agent    *agent.PhraseSenseiAgent
	repo     repository.ChatHistoryRepository
	
	// Configuration
	defaultModel string

	// Child Components
	chatInput input.Model
	chatView  chatview.Model
	
	// Layout
	width  int
	height int
	err    error
}

func NewModel(a *agent.PhraseSenseiAgent, r repository.ChatHistoryRepository, defaultModel string) tea.Model {
	return &TuiModel{
		agent:        a,
		repo:         r,
		defaultModel: defaultModel,
		chatInput:    input.New(),
		chatView:     chatview.New(),
	}
}

func (m *TuiModel) Init() tea.Cmd {
	return m.chatInput.Init()
}

func (m *TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Calculate available height for chat view
		// Input height (approx 3 lines with border)
		inputHeight := 3
		chatHeight := msg.Height - inputHeight
		
		m.chatView.SetDimensions(msg.Width, chatHeight)
		m.chatInput.SetWidth(msg.Width)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case input.SendRequestedMsg:
		// Handle message sent from Input component
		userInput := msg.Message
		m.chatView.AddMessage("user", userInput)
		
		// Trigger Agent
		return m, m.sendMessage(userInput)

	case responseMsg:
		m.chatView.AddMessage("sensei", msg.text)
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.chatView.AddMessage("sensei", fmtError(msg.err))
		return m, nil
	}

	// Propagate to children
	// Note: We cast to appropriate model types if Update returns a generic tea.Model
	
	// Update ChatView
	var cvModel tea.Model
	cvModel, cmd = m.chatView.Update(msg)
	m.chatView = cvModel.(chatview.Model)
	cmds = append(cmds, cmd)

	// Update Input
	var ciModel tea.Model
	ciModel, cmd = m.chatInput.Update(msg)
	m.chatInput = ciModel.(input.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *TuiModel) View() string {
	if m.height == 0 {
		return "Initializing..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.chatView.View(),
		m.chatInput.View(),
	)
}

// --- Commands ---

type responseMsg struct {
	text string
	done bool
}

type errorMsg struct {
	err error
}

func fmtError(err error) string {
	return style.Base.Foreground(style.ColorError).Render(err.Error())
}

func (m *TuiModel) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		req := &agent.GenerateRequest{
			Model:  m.defaultModel,
			Prompt: text,
		}

		resp, err := m.agent.Rephrase(ctx, req)
		if err != nil {
			return errorMsg{err}
		}
		return responseMsg{text: resp.Text, done: true}
	}
}