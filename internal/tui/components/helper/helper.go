package helper

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

type Command struct {
	Key         string
	Description string
}

var commands = []Command{
	{"/history", "View past conversations"},
	{"/report",  "Analyze linguistic patterns"},
	{"/lesson",  "Get a personalized English lesson"},
	{"/review",  "Quiz yourself on past mistakes"},
	{"/new",     "Start a fresh conversation"},
	{"/quit",    "Exit the application"},
}

type Model struct {
	filter string
	width  int
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m *Model) SetFilter(f string) {
	m.filter = f
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var suggestions []string

	// Only show suggestions if the user starts typing a command
	if strings.HasPrefix(m.filter, "/") {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd.Key, m.filter) {
				suggestion := fmt.Sprintf("%s %s", 
					style.Base.Foreground(style.ColorPrimary).Bold(true).Render(cmd.Key),
					style.Base.Foreground(style.ColorSubText).Render("- "+cmd.Description),
				)
				suggestions = append(suggestions, suggestion)
			}
		}
	} else if m.filter == "" {
		// Default hint when empty
		suggestions = append(suggestions, style.Base.Foreground(style.ColorSubText).Render("Tip: Type / to see available commands"))
	}

	if len(suggestions) == 0 {
		return ""
	}

	// Join suggestions horizontally or vertically depending on space
	// For now, let's join them with a separator
	content := strings.Join(suggestions, "   ")
	
	// Truncate if too long
	if lipgloss.Width(content) > m.width {
		content = content[:m.width-3] + "..."
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(style.ColorSubText).
		Faint(true).
		Padding(0, 1).
		Width(m.width).
		Render(content)
}
