package lesson

import (
	"strings"

tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

type Model struct {
	viewport viewport.Model
	content  string
	ready    bool
	loading  bool
	width    int
	height   int
}

func New() Model {
	return Model{
		loading: false,
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
	if m.loading {
		return style.Base.Render("Phrase Sensei is studying your habits... (Consulting with Data Analyst)")
	}
	if !m.ready || m.content == "" {
		return style.Base.Render("No lessons generated yet. Type /lesson to start.")
	}
	return m.viewport.View()
}

func (m *Model) SetDimensions(w, h int) {
	m.width = w
	m.height = h
	if !m.ready {
		m.viewport = viewport.New(w, h)
		m.ready = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = h
	}
	m.updateViewport()
}

func (m *Model) SetLoading(l bool) {
	m.loading = l
}

func (m *Model) SetContent(content string) {
	m.content = content
	m.loading = false
	m.updateViewport()
}

func (m *Model) updateViewport() {
	if !m.ready {
		return
	}
	
	// Apply some basic styling to markdown-like headers
	lines := strings.Split(m.content, "\n")
	var styledLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "##") {
			styledLines = append(styledLines, style.Base.Foreground(style.ColorPrimary).Bold(true).Render(line))
		} else {
			styledLines = append(styledLines, style.Base.Render(line))
		}
	}
	
m.viewport.SetContent(strings.Join(styledLines, "\n"))
	m.viewport.GotoTop()
}