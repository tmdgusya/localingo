package review

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmdgusya/localingo/internal/repository"
	"github.com/tmdgusya/localingo/internal/srs"
	"github.com/tmdgusya/localingo/internal/tui/style"
)

type State int

const (
	StateLoading State = iota
	StateQuestion
	StateResult
	StateDone
)

type Model struct {
	repo         repository.ChatHistoryRepository
	pairs        []repository.CorrectionPair
	currentIndex int
	state        State
	input        textinput.Model
	feedback     string
	width        int
}

type PairsLoadedMsg struct {
	Pairs []repository.CorrectionPair
}

type ErrorMsg struct {
	Err error
}

type SRSUpdateMsg struct {
	Err error
}

func New(repo repository.ChatHistoryRepository) Model {

ti := textinput.New()

ti.Placeholder = "Type the corrected sentence..."
ti.Focus()
ti.Width = 50

return Model{
	repo:  repo,
	state: StateLoading,
	input: ti,
}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.input.Width = msg.Width - 10

	case PairsLoadedMsg:
		if len(msg.Pairs) == 0 {
			m.state = StateDone
			return m, nil
		}
		m.pairs = msg.Pairs
		m.currentIndex = 0
		m.state = StateQuestion
		m.input.SetValue("")
		m.input.Focus()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.state == StateQuestion {
				// Submit answer
				userAnswer := strings.TrimSpace(m.input.Value())
				correctAnswer := strings.TrimSpace(m.pairs[m.currentIndex].Corrected)
				
				grade := srs.GradeIncorrect
				
				// Simple case-insensitive check (can be improved)
				if strings.EqualFold(userAnswer, correctAnswer) {
					m.feedback = style.Base.Foreground(style.ColorSecondary).Bold(true).Render("✨ Perfect! You got it right.")
					grade = srs.GradePerfect
				} else {
					m.feedback = fmt.Sprintf("%s\n\nCorrect answer: %s", 
						style.Base.Foreground(style.ColorError).Render("Not quite."),
						style.Base.Foreground(style.ColorSecondary).Bold(true).Render(correctAnswer),
					)
				}
				
				// Trigger SRS update
				cmd = m.updateSRS(m.pairs[m.currentIndex].ID, grade)
				
				m.state = StateResult
				return m, cmd
			} else if m.state == StateResult {
				// Next question
				if m.currentIndex < len(m.pairs)-1 {
					m.currentIndex++
					m.state = StateQuestion
					m.input.SetValue("")
					m.feedback = ""
					return m, nil
				} else {
					m.state = StateDone
					return m, nil
				}
			}
		}
	}

	if m.state == StateQuestion {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var sb strings.Builder

	header := style.Base.Foreground(style.ColorPrimary).Bold(true).Render("Fossil Breaker: Review Quiz")
	sb.WriteString(header + "\n\n")

	if m.state == StateLoading {
		sb.WriteString("Loading your past mistakes...")
		return sb.String()
	}

	if m.state == StateDone {
		if len(m.pairs) == 0 {
			sb.WriteString("No correction history found to review. Chat more to generate data!")
		} else {
			sb.WriteString(style.Base.Foreground(style.ColorSecondary).Render("🎉 Review session complete! Good job."))
		}
		sb.WriteString("\n\nPress [Esc] to return to chat.")
		return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
	}

	// Progress
	progress := fmt.Sprintf("Question %d/%d", m.currentIndex+1, len(m.pairs))
	sb.WriteString(style.Base.Foreground(style.ColorSubText).Render(progress) + "\n\n")

	// Original Sentence (The Problem)
	currentPair := m.pairs[m.currentIndex]
	sb.WriteString(style.Base.Foreground(style.ColorText).Render("Original (Incorrect/Awkward):") + "\n")
	sb.WriteString(style.Base.Background(style.ColorSubText).Foreground(style.ColorText).Padding(0, 1).Render(currentPair.Original) + "\n\n")

	if m.state == StateQuestion {
		sb.WriteString(style.Base.Foreground(style.ColorText).Render("Your Correction:") + "\n")
		sb.WriteString(m.input.View())
		sb.WriteString("\n\n" + style.Base.Foreground(style.ColorSubText).Render("[Enter] to submit"))
	} else if m.state == StateResult {
		sb.WriteString(style.Base.Foreground(style.ColorText).Render("Your Answer: ") + m.input.Value() + "\n\n")
		sb.WriteString(m.feedback + "\n\n")
		sb.WriteString(style.Base.Foreground(style.ColorSubText).Render("[Enter] for next question"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// Commands
func (m *Model) LoadReviewPairs() tea.Cmd {
	m.state = StateLoading
	return func() tea.Msg {
		// Fetch items due for review
		pairs, err := m.repo.GetDueReviewPairs(context.Background(), 10)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		
		return PairsLoadedMsg{Pairs: pairs}
	}
}

func (m *Model) updateSRS(id [16]byte, grade srs.Grade) tea.Cmd {
	// uuid.UUID is [16]byte
	// We need to convert it properly if needed, but repository expects uuid.UUID which is [16]byte
	return func() tea.Msg {
		err := m.repo.UpdateSRSStatus(context.Background(), id, int(grade))
		if err != nil {
			return SRSUpdateMsg{Err: err}
		}
		return nil
	}
}
