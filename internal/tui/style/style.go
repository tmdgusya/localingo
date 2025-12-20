package style

import "github.com/charmbracelet/lipgloss"

// Theme colors
var (
	ColorPrimary   = lipgloss.Color("#7D56F4") // Purple
	ColorSecondary = lipgloss.Color("#25A065") // Green
	ColorText      = lipgloss.Color("#FAFAFA") // White
	ColorSubText   = lipgloss.Color("#767676") // Grey
	ColorError     = lipgloss.Color("#FF0000") // Red
)

// Common Styles
var (
	// Base styles
	Base = lipgloss.NewStyle().
		Foreground(ColorText)

	// Sender styles
	UserLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginRight(1).
			Render("You:")

	SenseiLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			MarginRight(1).
			Render("Sensei:")

	// Message body styles
	UserMessage = lipgloss.NewStyle().
			Foreground(ColorText).
			PaddingLeft(2)

	SenseiMessage = lipgloss.NewStyle().
			Foreground(ColorText)

	// Input styles
	InputPrompt = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	InputBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)
)
