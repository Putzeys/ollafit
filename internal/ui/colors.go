package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")). // bright blue
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // gray
			Italic(true)

	// Verdict styles
	VerdictCanRun = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")) // green

	VerdictDegraded = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11")) // yellow

	VerdictCannotRun = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("9")) // red

	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")). // white
			Background(lipgloss.Color("8")).   // gray bg
			Padding(0, 1)

	CellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Info labels
	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")) // cyan

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")) // white

	// Warning/Error
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // yellow

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")) // red

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")) // green

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // dim gray

	// TUI tab styles
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("12")).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).
				Background(lipgloss.Color("0")).
				Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("0")).
			Padding(0, 1).
			Width(80)
)
