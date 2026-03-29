package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/putzeys/ollafit/internal/config"
	"github.com/putzeys/ollafit/internal/hardware"
	"github.com/putzeys/ollafit/internal/ollama"
)

const (
	tabHardware = iota
	tabCompat
	tabModels
	tabPull
)

var tabNames = []string{"Hardware", "Compatibilidade", "Modelos", "Pull"}

// --- Messages ---

type hwDetectedMsg struct {
	hw  hardware.SystemInfo
	err error
}

type modelsLoadedMsg struct {
	remote []ollama.RemoteModel
	local  []ollama.LocalModel
	err    error
}

type pullProgressMsg struct {
	status    string
	total     int64
	completed int64
}

type pullDoneMsg struct {
	err error
}

// --- TUI Model ---

// TUIModel is the top-level bubbletea model for the interactive TUI.
type TUIModel struct {
	cfg       config.Config
	activeTab int
	width     int
	height    int

	hwTab     HardwareTab
	compatTab CompatTab
	modelsTab ModelsTab
	pullTab   PullTab
}

// NewTUIModel creates a new TUI model with the given config.
func NewTUIModel(cfg config.Config) TUIModel {
	return TUIModel{
		cfg:       cfg,
		activeTab: tabHardware,
		hwTab:     NewHardwareTab(),
		compatTab: NewCompatTab(cfg.VRAMOverheadPercent, cfg.GPUMemoryFraction),
		modelsTab: NewModelsTab(cfg.OllamaHost),
		pullTab:   NewPullTab(cfg.VRAMOverheadPercent, cfg.GPUMemoryFraction, cfg.OllamaHost),
	}
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(detectHardwareCmd(), loadModelsCmd(m.cfg.OllamaHost))
}

func detectHardwareCmd() tea.Cmd {
	return func() tea.Msg {
		hw, err := hardware.DetectAll()
		return hwDetectedMsg{hw: hw, err: err}
	}
}

func loadModelsCmd(ollamaHost string) tea.Cmd {
	return func() tea.Msg {
		// Try ollama.com/library first (scraping), then ollamadb.dev API, then fallback
		remote, err := ollama.FetchOllamaLibrary()
		if err != nil || len(remote) == 0 {
			registry := ollama.NewRegistry()
			remote, err = registry.FetchAllModels(200)
		}
		if err != nil || len(remote) == 0 {
			remote = ollama.GetPopularModels()
		}

		var local []ollama.LocalModel
		client := ollama.NewClient(ollamaHost)
		if client.IsRunning() {
			local, _ = client.ListLocalModels()
		}

		return modelsLoadedMsg{remote: remote, local: local, err: err}
	}
}

func startPullCmd(ollamaHost, modelName string, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		client := ollama.NewClient(ollamaHost)
		if !client.IsRunning() {
			return pullDoneMsg{err: fmt.Errorf("Ollama is not running at %s", ollamaHost)}
		}
		err := client.PullModel(modelName, func(status string, total, completed int64) {
			p.Send(pullProgressMsg{status: status, total: total, completed: completed})
		})
		return pullDoneMsg{err: err}
	}
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case hwDetectedMsg:
		if msg.err != nil {
			m.hwTab.SetError(msg.err)
		} else {
			m.hwTab.SetHardware(msg.hw)
			m.compatTab.SetHardware(msg.hw)
			m.pullTab.SetHardware(msg.hw)
		}
		return m, nil

	case modelsLoadedMsg:
		if msg.err != nil {
			m.modelsTab.SetError(msg.err)
		}
		m.modelsTab.SetRemote(msg.remote)
		m.modelsTab.SetLocal(msg.local)
		m.compatTab.SetRemoteModels(msg.remote)
		return m, nil

	case pullProgressMsg:
		m.pullTab.UpdateProgress(msg.status, msg.total, msg.completed)
		if msg.total > 0 {
			pct := float64(msg.completed) / float64(msg.total)
			return m, m.pullTab.progress.SetPercent(pct)
		}
		return m, nil

	case pullDoneMsg:
		m.pullTab.FinishPull(msg.err)
		// Reload local models after pull
		return m, loadModelsCmd(m.cfg.OllamaHost)

	case progress.FrameMsg:
		if m.activeTab == tabPull {
			progressModel, cmd := m.pullTab.progress.Update(msg)
			m.pullTab.progress = progressModel.(progress.Model)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m TUIModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// If in search mode on compat tab
	if m.activeTab == tabCompat && m.compatTab.searching {
		return m.handleCompatSearch(msg)
	}

	// If in search mode on models tab
	if m.activeTab == tabModels && m.modelsTab.searching {
		return m.handleModelsSearch(msg)
	}

	// If editing input on pull tab
	if m.activeTab == tabPull && m.pullTab.editing {
		return m.handlePullInput(msg)
	}

	// If pull tab has compat result and waiting for confirmation
	if m.activeTab == tabPull && m.pullTab.compatResult != nil && !m.pullTab.pulling && !m.pullTab.pullFinished {
		switch key {
		case "enter":
			m.pullTab.StartPull()
			return m, func() tea.Msg {
				// We need the program reference for streaming; use a goroutine approach
				return nil
			}
		case "esc":
			m.pullTab.Reset()
			return m, nil
		}
	}

	// If pull finished, enter to reset
	if m.activeTab == tabPull && m.pullTab.pullFinished {
		if key == "enter" {
			m.pullTab.Reset()
			return m, nil
		}
	}

	// Tab navigation
	switch key {
	case "tab":
		m.activeTab = (m.activeTab + 1) % 4
		return m, nil
	case "shift+tab":
		m.activeTab = (m.activeTab + 3) % 4
		return m, nil
	case "1":
		m.activeTab = tabHardware
		return m, nil
	case "2":
		m.activeTab = tabCompat
		return m, nil
	case "3":
		m.activeTab = tabModels
		return m, nil
	case "4":
		m.activeTab = tabPull
		return m, nil
	}

	// Tab-specific keys
	switch m.activeTab {
	case tabCompat:
		switch key {
		case "up", "k":
			m.compatTab.MoveUp()
		case "down", "j":
			m.compatTab.MoveDown()
		case "q":
			m.compatTab.CycleQuant()
		case "/":
			m.compatTab.StartSearch()
		}
	case tabModels:
		switch key {
		case "up", "k":
			m.modelsTab.MoveUp()
		case "down", "j":
			m.modelsTab.MoveDown()
		case "l":
			m.modelsTab.ToggleLocal()
		case "/":
			m.modelsTab.StartSearch()
		case "enter":
			// Send selected model to pull tab
			name := m.modelsTab.SelectedModelName()
			if name != "" {
				m.pullTab.SetInput(name)
				m.activeTab = tabPull
			}
		}
	case tabHardware:
		if key == "q" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m TUIModel) handleCompatSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter", "esc":
		if key == "esc" {
			m.compatTab.StopSearch()
		} else {
			m.compatTab.searching = false
		}
	case "backspace":
		if len(m.compatTab.searchQuery) > 0 {
			m.compatTab.SetSearchQuery(m.compatTab.searchQuery[:len(m.compatTab.searchQuery)-1])
		}
	default:
		if len(key) == 1 {
			m.compatTab.SetSearchQuery(m.compatTab.searchQuery + key)
		}
	}
	return m, nil
}

func (m TUIModel) handleModelsSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter", "esc":
		if key == "esc" {
			m.modelsTab.StopSearch()
		} else {
			m.modelsTab.searching = false
		}
	case "backspace":
		if len(m.modelsTab.searchQuery) > 0 {
			m.modelsTab.SetSearchQuery(m.modelsTab.searchQuery[:len(m.modelsTab.searchQuery)-1])
		}
	default:
		if len(key) == 1 {
			m.modelsTab.SetSearchQuery(m.modelsTab.searchQuery + key)
		}
	}
	return m, nil
}

func (m TUIModel) handlePullInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		if m.pullTab.input != "" {
			m.pullTab.editing = false
			m.pullTab.CheckCompat()
		}
	case "esc":
		m.pullTab.Reset()
	case "backspace":
		if len(m.pullTab.input) > 0 {
			m.pullTab.input = m.pullTab.input[:len(m.pullTab.input)-1]
		}
	case "tab", "shift+tab", "1", "2", "3", "4":
		// Allow tab switching even while editing
		return m.handleTabSwitch(key)
	default:
		if len(key) == 1 {
			m.pullTab.input += key
		}
	}
	return m, nil
}

func (m TUIModel) handleTabSwitch(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "tab":
		m.activeTab = (m.activeTab + 1) % 4
	case "shift+tab":
		m.activeTab = (m.activeTab + 3) % 4
	case "1":
		m.activeTab = tabHardware
	case "2":
		m.activeTab = tabCompat
	case "3":
		m.activeTab = tabModels
	case "4":
		m.activeTab = tabPull
	}
	return m, nil
}

func (m TUIModel) View() string {
	var b strings.Builder

	// Title bar
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("  OLLAFIT")
	b.WriteString(title)
	b.WriteString("\n")

	// Tab bar
	b.WriteString("  ")
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d %s ", i+1, name)
		if i == m.activeTab {
			b.WriteString(TabActiveStyle.Render(label))
		} else {
			b.WriteString(TabInactiveStyle.Render(label))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n")

	// Separator
	sep := strings.Repeat("─", m.width)
	b.WriteString(DimStyle.Render(sep))
	b.WriteString("\n")

	// Content area
	contentHeight := m.height - 6 // title + tabs + sep + help + sep
	if contentHeight < 5 {
		contentHeight = 5
	}

	var content string
	switch m.activeTab {
	case tabHardware:
		content = m.hwTab.View(m.width, contentHeight)
	case tabCompat:
		content = m.compatTab.View(m.width, contentHeight)
	case tabModels:
		content = m.modelsTab.View(m.width, contentHeight)
	case tabPull:
		content = m.pullTab.View(m.width, contentHeight)
	}
	b.WriteString(content)

	// Pad to fill content area
	contentLines := strings.Count(content, "\n") + 1
	for i := contentLines; i < contentHeight; i++ {
		b.WriteString("\n")
	}

	// Bottom separator
	b.WriteString(DimStyle.Render(sep))
	b.WriteString("\n")

	// Help bar
	help := m.helpText()
	b.WriteString(StatusBarStyle.Render(help))

	return b.String()
}

func (m TUIModel) helpText() string {
	base := "  tab/1-4: trocar aba"

	switch m.activeTab {
	case tabHardware:
		return base + " | q: sair"
	case tabCompat:
		if m.compatTab.searching {
			return "  type to search | enter: confirm | esc: cancel"
		}
		return base + " | j/k: navegar | q: ciclar quant | /: buscar"
	case tabModels:
		if m.modelsTab.searching {
			return "  type to search | enter: confirm | esc: cancel"
		}
		return base + " | j/k: navegar | l: local/remote | /: buscar | enter: pull"
	case tabPull:
		if m.pullTab.editing {
			return "  type model name | enter: check | esc: cancel"
		}
		if m.pullTab.pulling {
			return "  downloading... | ctrl+c: sair"
		}
		return base + " | enter: confirmar | esc: cancelar"
	}

	return base + " | ctrl+c: sair"
}

// RunTUI starts the interactive TUI.
func RunTUI(cfg config.Config) error {
	m := NewTUIModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// We need to handle the pull command with access to the program.
	// Override the pull tab's start logic to use the program reference.
	// This is done by wrapping the model.
	wrapper := &tuiWrapper{model: m, program: p}
	p = tea.NewProgram(wrapper, tea.WithAltScreen())
	wrapper.program = p

	_, err := p.Run()
	return err
}

// tuiWrapper wraps TUIModel to inject the program reference for pull operations.
type tuiWrapper struct {
	model   TUIModel
	program *tea.Program
}

func (w *tuiWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *tuiWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Intercept pull confirmation
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if w.model.activeTab == tabPull &&
			w.model.pullTab.compatResult != nil &&
			!w.model.pullTab.pulling &&
			!w.model.pullTab.pullFinished &&
			keyMsg.String() == "enter" {

			w.model.pullTab.StartPull()
			modelName := w.model.pullTab.input
			ollamaHost := w.model.cfg.OllamaHost
			prog := w.program
			return w, func() tea.Msg {
				client := ollama.NewClient(ollamaHost)
				if !client.IsRunning() {
					return pullDoneMsg{err: fmt.Errorf("Ollama is not running at %s", ollamaHost)}
				}
				err := client.PullModel(modelName, func(status string, total, completed int64) {
					prog.Send(pullProgressMsg{status: status, total: total, completed: completed})
				})
				return pullDoneMsg{err: err}
			}
		}
	}

	updated, cmd := w.model.Update(msg)
	w.model = updated.(TUIModel)
	return w, cmd
}

func (w *tuiWrapper) View() string {
	return w.model.View()
}
