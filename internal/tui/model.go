package tui

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/tmdgusya/localingo/internal/agent"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui/components/chatview"
	"github.com/tmdgusya/localingo/internal/tui/components/history"
	"github.com/tmdgusya/localingo/internal/tui/components/input"
	"github.com/tmdgusya/localingo/internal/tui/components/report"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

// ViewState Enum
type ViewState int

const (
	ViewChat ViewState = iota
	ViewHistory
	ViewReport
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
	reportView  report.Model

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
		reportView:   report.New(r),
	}
}

func (m *TuiModel) Init() tea.Cmd {
	return tea.Batch(
		m.chatInput.Init(),
		m.chatView.Init(),
		m.historyView.Init(),
		m.reportView.Init(),
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
			if m.viewState == ViewHistory || m.viewState == ViewReport {
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

		// Propagate to all views to ensure they are ready
		_, hCmd := m.historyView.Update(msg)
		_, rCmd := m.reportView.Update(msg)
		cmds = append(cmds, hCmd, rCmd)

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
		log.Printf("Conversation initialized with ID: %s. Triggering pending actions.", msg.id)
		// Now that we have an ID, save the pending message and trigger agent
		return m, tea.Batch(
			m.saveMessageCmd(msg.id, "user", msg.firstMessage),
			m.sendMessage(msg.firstMessage),
		)

	// --- History Selection Handling ---
	case history.ConversationSelectedMsg:
		log.Printf("Conversation selected: %s", msg.ID)
		m.conversationID = msg.ID
		m.viewState = ViewChat
		m.chatView.Clear()
		return m, m.loadConversationMessages(msg.ID)

	case messagesLoadedMsg:
		log.Printf("Loaded %d messages for conversation", len(msg.messages))
		// Populate chat view
		for _, msg := range msg.messages {
			m.chatView.AddMessage(string(msg.Role), msg.Content)
		}
		return m, nil

	// --- Agent Response Handling ---
	case responseMsg:
		log.Printf("Received agent response. Saving to DB? (ID: %s, Repo: %v)", m.conversationID, m.repo != nil)
		m.chatView.AddMessage("sensei", msg.text)
		
		// Save assistant response
		if m.conversationID != uuid.Nil && m.repo != nil {
			var metadata map[string]interface{}
			if msg.analysis != nil {
				metadata = make(map[string]interface{})
				metadata["analysis"] = msg.analysis
				log.Println("Analysis data found, saving to metadata")
			}
			cmds = append(cmds, m.saveMessageCmd(m.conversationID, "assistant", msg.text, metadata))
		} else {
			log.Println("Skipping DB save: ConversationID is nil or Repo is nil")
		}
		return m, tea.Batch(cmds...)
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

	case ViewReport:
		var rModel tea.Model
		rModel, cmd = m.reportView.Update(msg)
		m.reportView = rModel.(report.Model)
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
	case ViewReport:
		return m.reportView.View()
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

	// Update Report Layout
	// Handled via propagation or explicit set
}

func (m *TuiModel) handleCommand(text string) tea.Cmd {
	parts := strings.Fields(text)
	command := parts[0]

	switch command {
	case "/history":
		m.viewState = ViewHistory
		return m.historyView.LoadConversations()
	case "/report":
		m.viewState = ViewReport
		return m.reportView.LoadStats()
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

	// 1. Create Conversation if needed
	if m.conversationID == uuid.Nil && m.repo != nil {
		log.Println("Starting new conversation creation...")
		// We need to create a conversation first.
		// We DO NOT trigger sendMessage here. We wait for conversationCreatedMsg.
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
				log.Printf("Error creating conversation: %v", err)
				return errorMsg{err}
			}
			log.Printf("Conversation created: %s", conv.ID)
			return conversationCreatedMsg{id: conv.ID, firstMessage: text}
		})
		return tea.Batch(cmds...)
	}

	// 2. Existing Conversation Flow
	if m.conversationID != uuid.Nil && m.repo != nil {
		log.Printf("Saving user message to conversation %s", m.conversationID)
		cmds = append(cmds, m.saveMessageCmd(m.conversationID, "user", text))
	}

	// 3. Trigger Agent (Only if conversation exists, otherwise handled in conversationCreatedMsg)
	log.Println("Triggering agent request...")
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

func (m *TuiModel) saveMessageCmd(convID uuid.UUID, role, content string, metadata ...map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			log.Println("Repo is nil, cannot save message")
			return nil
		}
		ctx := context.Background()
		r := repository.MessageRoleUser
		if role == "assistant" {
			r = repository.MessageRoleAssistant
		}
		
		model := m.defaultModel
		
		meta := make(map[string]interface{})
		if len(metadata) > 0 && metadata[0] != nil {
			meta = metadata[0]
		}

		_, err := m.repo.CreateMessage(ctx, repository.CreateMessageParams{
			ConversationID: convID,
			Role:           r,
			Content:        content,
			Model:          &model,
			Metadata:       meta,
		})
		if err != nil {
			log.Printf("Failed to save message to DB: %v", err)
			return nil
		}
		log.Printf("Message saved successfully (Role: %s)", role)
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

		log.Println("Sending request to agent...")
		// Use RephraseAndAnalyze which returns full structure including Analysis
		resp, err := m.agent.RephraseAndAnalyze(ctx, req)
		if err != nil {
			log.Printf("Agent error: %v", err)
			return errorMsg{err}
		}
		log.Println("Agent response received")
		return responseMsg{
			text:     resp.Text, 
			done:     true,
			analysis: resp.Analysis,
		}
	}
}

// --- Utils ---

type responseMsg struct {
	text     string
	done     bool
	analysis *agent.Analysis
}

type errorMsg struct {
	err error
}

func fmtError(err error) string {
	return style.Base.Foreground(style.ColorError).Render(err.Error())
}
