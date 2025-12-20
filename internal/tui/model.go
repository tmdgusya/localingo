package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui/components/chatview"
	"github.com/tmdgusya/localingo/internal/tui/components/history"
	"github.com/tmdgusya/localingo/internal/tui/components/input"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

// ViewState Enum
type ViewState int

const (
	ViewChat ViewState = iota
	ViewHistory
)

type TuiModel struct {
	agent        *agent.PhraseSenseiAgent
	repo         repository.ChatHistoryRepository
	defaultModel string

	// State
	viewState      ViewState
	conversationID uuid.UUID // Current active conversation

	// Child Components
	chatInput   input.Model
	chatView    chatview.Model
	historyView history.Model

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
		viewState:    ViewChat,
		chatInput:    input.New(),
		chatView:     chatview.New(),
		historyView:  history.New(r),
	}
}

func (m *TuiModel) Init() tea.Cmd {
	return tea.Batch(
		m.chatInput.Init(),
		m.chatView.Init(),
		m.historyView.Init(),
	)
}

func (m *TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	// Global Key Handling (e.g. Quit)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		// We can add Esc to go back from History to Chat if needed
		case tea.KeyEsc:
			if m.viewState == ViewHistory {
				m.viewState = ViewChat
				return m, nil
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout(msg.Width, msg.Height)

	// --- Command & Input Handling ---
	case input.SendRequestedMsg:
		text := strings.TrimSpace(msg.Message)
		if strings.HasPrefix(text, "/") {
			// Handle Command
			return m, m.handleCommand(text)
		} else {
			// Handle Chat Message
			return m, m.handleUserMessage(text)
		}

	// --- Conversation Creation Handling ---
	case conversationCreatedMsg:
		m.conversationID = msg.id
		// Now that we have an ID, save the pending message and trigger agent
		return m, tea.Batch(
			m.saveMessageCmd(msg.id, "user", msg.firstMessage),
			m.sendMessage(msg.firstMessage),
		)

	// --- History Selection Handling ---
	case history.ConversationSelectedMsg:
		m.conversationID = msg.ID
		m.viewState = ViewChat
		m.chatView.Clear() // Implement Clear in chatview
		return m, m.loadConversationMessages(msg.ID)

	case messagesLoadedMsg:
		// Populate chat view
		for _, msg := range msg.messages {
			m.chatView.AddMessage(string(msg.Role), msg.Content)
		}
		return m, nil

	// --- Agent Response Handling ---
	case responseMsg:
		m.chatView.AddMessage("sensei", msg.text)
		// Save assistant response
		if m.conversationID != uuid.Nil && m.repo != nil {
			cmds = append(cmds, m.saveMessageCmd(m.conversationID, "assistant", msg.text))
		}
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.chatView.AddMessage("sensei", fmtError(msg.err))
		return m, nil
	}

	// --- Component Updates based on ViewState ---

	switch m.viewState {
	case ViewChat:
		var ciModel tea.Model
		ciModel, cmd = m.chatInput.Update(msg)
		m.chatInput = ciModel.(input.Model)
		cmds = append(cmds, cmd)

		var cvModel tea.Model
		cvModel, cmd = m.chatView.Update(msg)
		m.chatView = cvModel.(chatview.Model)
		cmds = append(cmds, cmd)

	case ViewHistory:
		var hModel tea.Model
		hModel, cmd = m.historyView.Update(msg)
		m.historyView = hModel.(history.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *TuiModel) View() string {
	if m.height == 0 {
		return "Initializing..."
	}

	switch m.viewState {
	case ViewHistory:
		return m.historyView.View()
	default:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.chatView.View(),
			m.chatInput.View(),
		)
	}
}

// --- Helper Logic ---

func (m *TuiModel) layout(width, height int) {
	// Update Chat Layout
	inputHeight := 3
	chatHeight := height - inputHeight
	m.chatView.SetDimensions(width, chatHeight)
	m.chatInput.SetWidth(width)

	// Update History Layout
	m.historyView.SetSize(width, height)
}

func (m *TuiModel) handleCommand(text string) tea.Cmd {
	parts := strings.Fields(text)
	command := parts[0]

	switch command {
	case "/history":
		m.viewState = ViewHistory
		return m.historyView.LoadConversations()
	case "/new":
		m.conversationID = uuid.Nil
		m.chatView.Clear() // Need to implement Clear
		return nil
	case "/quit":
		return tea.Quit
	default:
		m.chatView.AddMessage("sensei", style.Base.Foreground(style.ColorSubText).Render(fmt.Sprintf("Unknown command: %s", command)))
		return nil
	}
}

func (m *TuiModel) handleUserMessage(text string) tea.Cmd {
	m.chatView.AddMessage("user", text)

	var cmds []tea.Cmd

	// 1. Create Conversation if needed (Optimistic UI: do it in background or blocking?)
	// For simplicity in TUI, we might want to do it before triggering agent
	if m.conversationID == uuid.Nil && m.repo != nil {
		// We need to create a conversation first.
		// NOTE: This should ideally be a Msg flow, but for simplicity we wrap in a Cmd
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			// Create conversation
			title := text
			if len(title) > 30 {
				title = title[:30] + "..."
			}
			conv, err := m.repo.CreateConversation(ctx, repository.CreateConversationParams{
				Title: &title,
			})
			if err != nil {
				return errorMsg{err}
			}
			
			// We need to update the model's conversationID. 
			// But we can't update model from Cmd. We must return a Msg.
			return conversationCreatedMsg{id: conv.ID, firstMessage: text}
		})
		return tea.Batch(cmds...)
	}

	// 2. Save User Message
	if m.conversationID != uuid.Nil && m.repo != nil {
		cmds = append(cmds, m.saveMessageCmd(m.conversationID, "user", text))
	}

	// 3. Trigger Agent
	cmds = append(cmds, m.sendMessage(text))

	return tea.Batch(cmds...)
}

// --- Internal Messages & Commands ---

type conversationCreatedMsg struct {
	id           uuid.UUID
	firstMessage string
}

type messagesLoadedMsg struct {
	messages []*repository.Message
}

func (m *TuiModel) loadConversationMessages(id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return nil
		}
		ctx := context.Background()
		msgs, err := m.repo.ListMessagesByConversation(ctx, id)
		if err != nil {
			return errorMsg{err}
		}
		return messagesLoadedMsg{messages: msgs}
	}
}

func (m *TuiModel) saveMessageCmd(convID uuid.UUID, role, content string) tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return nil
		}
		ctx := context.Background()
		r := repository.MessageRoleUser
		if role == "assistant" {
			r = repository.MessageRoleAssistant
		}
		
		model := m.defaultModel
		
		_, err := m.repo.CreateMessage(ctx, repository.CreateMessageParams{
			ConversationID: convID,
			Role:           r,
			Content:        content,
			Model:          &model,
		})
		if err != nil {
			// Log error or ignore? 
			// return errorMsg{err} 
			// Silent fail for history save for now to not interrupt chat
			return nil
		}
		return nil
	}
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

// --- Utils ---

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
