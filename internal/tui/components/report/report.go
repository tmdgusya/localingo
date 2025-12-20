package report

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

// Messages
type StatsLoadedMsg struct {
	Stats *repository.ErrorStats
}

type ErrorMsg struct {
	Err error
}

// Model
type Model struct {
	repo    repository.ChatHistoryRepository
	stats   *repository.ErrorStats
	width   int
	height  int
	loading bool
}

func New(repo repository.ChatHistoryRepository) Model {
	return Model{
		repo:    repo,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case StatsLoadedMsg:
		m.stats = msg.Stats
		m.loading = false
		return m, nil

	case ErrorMsg:
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return style.Base.Render("Analyzing your patterns... (Collecting data from heart)")
	}

	if m.stats == nil || len(m.stats.CategoryCount) == 0 {
		return style.Base.Render("No patterns detected yet. Keep chatting with Sensei!")
	}

	var sb strings.Builder

	header := style.Base.Foreground(style.ColorPrimary).Bold(true).Render("Linguistic Pattern Report")
	sb.WriteString(header + "\n")
	sb.WriteString(style.Base.Foreground(style.ColorSubText).Render(fmt.Sprintf("Total analyzed messages: %d", m.stats.TotalAnalyzed)) + "\n\n")

	// Sort categories by count
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	maxVal := 0
	for k, v := range m.stats.CategoryCount {
		sorted = append(sorted, kv{k, v})
		if v > maxVal {
			maxVal = v
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	// Render bar chart
	chartWidth := m.width - 20
	if chartWidth < 10 {
		chartWidth = 10
	}

	for _, item := range sorted {
		label := lipgloss.NewStyle().Width(12).Render(item.Key)
		barLen := 0
		if maxVal > 0 {
			barLen = (item.Value * chartWidth) / maxVal
		}
		bar := lipgloss.NewStyle(
			).
			Foreground(style.ColorSecondary).
			Render(strings.Repeat("█", barLen))
		
		count := style.Base.Foreground(style.ColorSubText).Render(fmt.Sprintf(" (%d)", item.Value))
		
		sb.WriteString(fmt.Sprintf("%s %s%s\n", label, bar, count))
	}

	sb.WriteString("\n" + style.Base.Foreground(style.ColorSubText).Render("Press [Esc] to go back to chat."))

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// Commands
func (m *Model) LoadStats() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx := context.Background()
		stats, err := m.repo.GetErrorStats(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return StatsLoadedMsg{Stats: stats}
	}
}
