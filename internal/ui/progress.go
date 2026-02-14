package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
)

// PullProgress tracks download progress for model pulling.
type PullProgress struct {
	progress progress.Model
	status   string
	total    int64
	done     int64
	finished bool
	err      error
}

// PullProgressMsg is sent to update the progress bar.
type PullProgressMsg struct {
	Status    string
	Total     int64
	Completed int64
}

// PullDoneMsg signals the pull is complete.
type PullDoneMsg struct {
	Err error
}

func NewPullProgress() PullProgress {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50
	return PullProgress{
		progress: p,
		status:   "Starting download...",
	}
}

func (m PullProgress) Init() tea.Cmd {
	return nil
}

func (m PullProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case PullProgressMsg:
		m.status = msg.Status
		m.total = msg.Total
		m.done = msg.Completed

		if msg.Total > 0 {
			pct := float64(msg.Completed) / float64(msg.Total)
			return m, m.progress.SetPercent(pct)
		}

	case PullDoneMsg:
		m.finished = true
		m.err = msg.Err
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m PullProgress) View() string {
	if m.finished {
		if m.err != nil {
			return ErrorStyle.Render(fmt.Sprintf("\nError: %s\n", m.err))
		}
		return SuccessStyle.Render("\nDownload complete!\n")
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(LabelStyle.Render("Status: "))
	b.WriteString(m.status)
	b.WriteString("\n\n")
	b.WriteString(m.progress.View())

	if m.total > 0 {
		b.WriteString(DimStyle.Render(fmt.Sprintf(" %s / %s",
			formatSize(m.done), formatSize(m.total))))
	}

	b.WriteString("\n\n")
	b.WriteString(DimStyle.Render("Press q or ctrl+c to cancel"))
	b.WriteString("\n")

	return b.String()
}

func formatSize(bytes int64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)
	if bytes >= GB {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	}
	return fmt.Sprintf("%.0f MB", float64(bytes)/float64(MB))
}
