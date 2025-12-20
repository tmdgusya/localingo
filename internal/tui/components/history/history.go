package history

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

// Messages
type ConversationSelectedMsg struct {
	ID uuid.UUID
}

type ConversationsLoadedMsg struct {
	Conversations []*repository.Conversation
}

type ErrorMsg struct {
	Err error
}

// Item implements list.Item
type Item struct {
	id        uuid.UUID
	title     string
	createdAt time.Time
}

func (i Item) FilterValue() string { return i.title }
func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.createdAt.Format(time.RFC822) }

// Model
type Model struct {
	list  list.Model
	repo  repository.ChatHistoryRepository
	ready bool
}

func New(repo repository.ChatHistoryRepository) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Chat History"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.Styles.Title = style.Base.Foreground(style.ColorPrimary).Bold(true)

	return Model{
		list: l,
		repo: repo,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)

	case ConversationsLoadedMsg:
		items := make([]list.Item, len(msg.Conversations))
		for i, c := range msg.Conversations {
			title := "Untitled Conversation"
			if c.Title != nil && *c.Title != "" {
				title = *c.Title
			} else {
				title = fmt.Sprintf("Conversation %s", c.ID.String()[:8])
			}
			items[i] = Item{
				id:        c.ID,
				title:     title,
				createdAt: c.CreatedAt,
			}
		}
		m.list.SetItems(items)
		m.ready = true

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.Type {
		case tea.KeyEnter:
			if selectedItem, ok := m.list.SelectedItem().(Item); ok {
				return m, func() tea.Msg {
					return ConversationSelectedMsg{ID: selectedItem.id}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading history..."
	}
	return m.list.View()
}

// Commands
func (m *Model) LoadConversations() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// Fetch last 20 conversations
		convs, err := m.repo.ListConversations(ctx, 20, 0)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return ConversationsLoadedMsg{Conversations: convs}
	}
}

func (m *Model) SetSize(width, height int) {
	m.list.SetSize(width, height)
}
