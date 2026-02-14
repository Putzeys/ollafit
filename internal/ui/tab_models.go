package ui

import (
	"fmt"
	"strings"

	"github.com/putzeys/putzeys-cli/internal/ollama"
)

// ModelsTab holds state for the models listing tab.
type ModelsTab struct {
	remote      []ollama.RemoteModel
	local       []ollama.LocalModel
	showLocal   bool
	searching   bool
	searchQuery string
	cursor      int
	loading     bool
	err         error
	ollamaHost  string
}

func NewModelsTab(ollamaHost string) ModelsTab {
	return ModelsTab{
		loading:    true,
		ollamaHost: ollamaHost,
	}
}

func (t *ModelsTab) SetRemote(models []ollama.RemoteModel) {
	t.remote = models
	t.loading = false
}

func (t *ModelsTab) SetLocal(models []ollama.LocalModel) {
	t.local = models
}

func (t *ModelsTab) SetError(err error) {
	t.err = err
	t.loading = false
}

func (t *ModelsTab) ToggleLocal() {
	t.showLocal = !t.showLocal
	t.cursor = 0
}

func (t *ModelsTab) StartSearch() {
	t.searching = true
	t.searchQuery = ""
}

func (t *ModelsTab) StopSearch() {
	t.searching = false
	t.searchQuery = ""
}

func (t *ModelsTab) SetSearchQuery(q string) {
	t.searchQuery = q
}

func (t *ModelsTab) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
	}
}

func (t *ModelsTab) MoveDown() {
	max := t.listLen() - 1
	if t.cursor < max {
		t.cursor++
	}
}

// SelectedModelName returns the name of the currently selected model.
func (t *ModelsTab) SelectedModelName() string {
	if t.showLocal {
		filtered := t.filteredLocal()
		if t.cursor < len(filtered) {
			return filtered[t.cursor].Name
		}
	} else {
		filtered := t.filteredRemote()
		if t.cursor < len(filtered) {
			return filtered[t.cursor].Name
		}
	}
	return ""
}

func (t *ModelsTab) listLen() int {
	if t.showLocal {
		return len(t.filteredLocal())
	}
	return len(t.filteredRemote())
}

func (t *ModelsTab) filteredRemote() []ollama.RemoteModel {
	if t.searchQuery == "" {
		return t.remote
	}
	query := strings.ToLower(t.searchQuery)
	var out []ollama.RemoteModel
	for _, m := range t.remote {
		if strings.Contains(strings.ToLower(m.Name), query) ||
			strings.Contains(strings.ToLower(m.Description), query) {
			out = append(out, m)
		}
	}
	return out
}

func (t *ModelsTab) filteredLocal() []ollama.LocalModel {
	if t.searchQuery == "" {
		return t.local
	}
	query := strings.ToLower(t.searchQuery)
	var out []ollama.LocalModel
	for _, m := range t.local {
		if strings.Contains(strings.ToLower(m.Name), query) {
			out = append(out, m)
		}
	}
	return out
}

func (t ModelsTab) View(width, height int) string {
	if t.loading {
		return DimStyle.Render("  Loading models...")
	}

	var b strings.Builder

	// Mode indicator
	b.WriteString("  ")
	if t.showLocal {
		b.WriteString(LabelStyle.Render("Local Models"))
	} else {
		b.WriteString(LabelStyle.Render("Remote Models"))
	}
	b.WriteString(DimStyle.Render("  (press l to toggle)"))
	b.WriteString("\n")

	// Search indicator
	if t.searching {
		b.WriteString("  ")
		b.WriteString(LabelStyle.Render("/"))
		b.WriteString(ValueStyle.Render(t.searchQuery))
		b.WriteString(DimStyle.Render("_"))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if t.showLocal {
		t.renderLocal(&b, height)
	} else {
		t.renderRemote(&b, height)
	}

	return b.String()
}

func (t ModelsTab) renderRemote(b *strings.Builder, height int) {
	filtered := t.filteredRemote()
	if len(filtered) == 0 {
		b.WriteString(WarningStyle.Render("  No models found."))
		return
	}

	nameW := 30
	header := fmt.Sprintf("  %-*s %s", nameW, "MODEL", "DESCRIPTION")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	maxRows := height - 6
	if maxRows < 1 {
		maxRows = len(filtered)
	}

	start := 0
	if t.cursor >= maxRows {
		start = t.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := start; i < end; i++ {
		m := filtered[i]
		prefix := "  "
		if i == t.cursor {
			prefix = "> "
		}
		row := fmt.Sprintf("%s%-*s %s", prefix, nameW, truncate(m.Name, nameW), m.Description)
		b.WriteString(row)
		b.WriteString("\n")
	}
}

func (t ModelsTab) renderLocal(b *strings.Builder, height int) {
	filtered := t.filteredLocal()
	if len(filtered) == 0 {
		b.WriteString(WarningStyle.Render("  No local models found."))
		b.WriteString("\n")
		b.WriteString(DimStyle.Render("  Pull a model from the Pull tab or run: ollama pull <model>"))
		return
	}

	nameW := 30
	header := fmt.Sprintf("  %-*s %-12s %-10s %s", nameW, "MODEL", "PARAMS", "QUANT", "SIZE")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	maxRows := height - 6
	if maxRows < 1 {
		maxRows = len(filtered)
	}

	start := 0
	if t.cursor >= maxRows {
		start = t.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := start; i < end; i++ {
		m := filtered[i]
		prefix := "  "
		if i == t.cursor {
			prefix = "> "
		}
		size := formatSize(m.Size)
		row := fmt.Sprintf("%s%-*s %-12s %-10s %s",
			prefix,
			nameW, truncate(m.Name, nameW),
			m.Details.ParameterSize,
			m.Details.QuantizationLevel,
			size,
		)
		b.WriteString(row)
		b.WriteString("\n")
	}
}
