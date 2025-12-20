package helper

import (
	"fmt"
	"strings"

		"github.com/charmbracelet/bubbles/spinner"
		"github.com/charmbracelet/bubbletea"
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
			{"/level",   "Set correction strictness (1:Gentle, 2:Std, 3:Strict)"},
			{"/new",     "Start a fresh conversation"},
			{"/quit",    "Exit the application"},
		}
			type Model struct {
		filter     string
		width      int
		spinner    spinner.Model
		isThinking bool
	}
	
	func New() Model {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(style.ColorSecondary)
		
		return Model{
			spinner: s,
		}
	}
	
	func (m Model) Init() tea.Cmd {
		return m.spinner.Tick
	}
	
	func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd
	}
	
	func (m *Model) SetFilter(f string) {
		m.filter = f
	}
	
	func (m *Model) SetThinking(t bool) {
		m.isThinking = t
	}
	
	func (m Model) View() string {
		if m.width == 0 {
			return ""
		}
	
		var content string
	
		if m.isThinking {
			content = fmt.Sprintf("%s %s", m.spinner.View(), style.Base.Foreground(style.ColorSecondary).Render("Sensei is rephrasing..."))
		} else if strings.HasPrefix(m.filter, "/") {
			var suggestions []string
			for _, cmd := range commands {
				if strings.HasPrefix(cmd.Key, m.filter) {
					suggestion := fmt.Sprintf("%s %s", 
						style.Base.Foreground(style.ColorPrimary).Bold(true).Render(cmd.Key),
						style.Base.Foreground(style.ColorSubText).Render("- "+cmd.Description),
					)
					suggestions = append(suggestions, suggestion)
				}
			}
			if len(suggestions) > 0 {
				content = strings.Join(suggestions, "   ")
			}
		} else if m.filter == "" {
			content = style.Base.Foreground(style.ColorSubText).Render("Tip: Type / to see available commands")
		}
	
		if content == "" {
			return ""
		}
		
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
