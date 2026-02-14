package ui

import (
	"fmt"
	"strings"

	"github.com/putzeys/putzeys-cli/internal/hardware"
	"github.com/putzeys/putzeys-cli/internal/models"
	"github.com/putzeys/putzeys-cli/internal/ollama"
)

// CompatTab holds state for the compatibility check tab.
type CompatTab struct {
	hw              *hardware.SystemInfo
	remoteModels    []ollama.RemoteModel
	results         []models.CompatResult
	filtered        []models.CompatResult
	quant           models.Quantization
	quantIdx        int
	searching       bool
	searchQuery     string
	cursor          int
	loading         bool
	overheadPercent float64
	gpuMemFraction  float64
}

var quantCycle = []models.Quantization{models.QuantQ4KM, models.QuantQ8, models.QuantFP16}

func NewCompatTab(overheadPercent, gpuMemFraction float64) CompatTab {
	return CompatTab{
		quant:           models.QuantQ4KM,
		quantIdx:        0,
		loading:         true,
		overheadPercent: overheadPercent,
		gpuMemFraction:  gpuMemFraction,
	}
}

func (t *CompatTab) SetHardware(hw hardware.SystemInfo) {
	t.hw = &hw
	t.recompute()
}

func (t *CompatTab) SetRemoteModels(remote []ollama.RemoteModel) {
	t.remoteModels = remote
	t.recompute()
}

func (t *CompatTab) CycleQuant() {
	t.quantIdx = (t.quantIdx + 1) % len(quantCycle)
	t.quant = quantCycle[t.quantIdx]
	t.recompute()
}

func (t *CompatTab) StartSearch() {
	t.searching = true
	t.searchQuery = ""
}

func (t *CompatTab) StopSearch() {
	t.searching = false
	t.searchQuery = ""
	t.applyFilter()
}

func (t *CompatTab) SetSearchQuery(q string) {
	t.searchQuery = q
	t.applyFilter()
}

func (t *CompatTab) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
	}
}

func (t *CompatTab) MoveDown() {
	max := len(t.filtered) - 1
	if t.cursor < max {
		t.cursor++
	}
}

func (t *CompatTab) recompute() {
	if t.hw == nil {
		return
	}
	t.loading = false

	// Use all available models
	remoteModels := t.remoteModels
	if len(remoteModels) == 0 {
		remoteModels = ollama.GetPopularModels()
	}
	variants := make([]models.ModelVariant, 0, len(remoteModels))
	for _, m := range remoteModels {
		params := ollama.ParseModelParams(m.Name)
		est := models.EstimateVRAM(params, t.quant, t.overheadPercent)
		variants = append(variants, models.ModelVariant{
			Name:         m.Name,
			Description:  m.Description,
			Parameters:   params,
			Quantization: t.quant,
			VRAMRequired: est.TotalVRAM,
		})
	}

	t.results = models.CheckAll(variants, *t.hw, t.overheadPercent, t.gpuMemFraction)
	t.applyFilter()
}

func (t *CompatTab) applyFilter() {
	if t.searchQuery == "" {
		t.filtered = t.results
	} else {
		query := strings.ToLower(t.searchQuery)
		t.filtered = nil
		for _, r := range t.results {
			if strings.Contains(strings.ToLower(r.Model.Name), query) {
				t.filtered = append(t.filtered, r)
			}
		}
	}
	if t.cursor >= len(t.filtered) {
		t.cursor = max(0, len(t.filtered)-1)
	}
}

func (t CompatTab) View(width, height int) string {
	if t.loading {
		return DimStyle.Render("  Loading compatibility data...")
	}

	var b strings.Builder

	// Quantization indicator
	b.WriteString("  ")
	b.WriteString(LabelStyle.Render("Quantization: "))
	b.WriteString(ValueStyle.Render(t.quant.String()))
	b.WriteString(DimStyle.Render("  (press q to cycle)"))
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

	if len(t.filtered) == 0 {
		b.WriteString(WarningStyle.Render("  No models match the filter."))
		return b.String()
	}

	// Table header
	nameW, quantW, vramW, verdictW, reasonW := 25, 8, 12, 12, 35
	header := fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s",
		nameW, "MODEL",
		quantW, "QUANT",
		vramW, "VRAM NEEDED",
		verdictW, "STATUS",
		reasonW, "DETAILS",
	)
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Rows
	maxRows := height - 6 // leave room for header/footer
	if maxRows < 1 {
		maxRows = len(t.filtered)
	}

	// Scroll window around cursor
	start := 0
	if t.cursor >= maxRows {
		start = t.cursor - maxRows + 1
	}
	end := start + maxRows
	if end > len(t.filtered) {
		end = len(t.filtered)
	}

	for i := start; i < end; i++ {
		r := t.filtered[i]
		prefix := "  "
		if i == t.cursor {
			prefix = "> "
		}

		name := truncate(r.Model.Name, nameW)
		quant := truncate(r.Model.Quantization.String(), quantW)
		vram := truncate(hardware.FormatBytes(r.Estimate.TotalVRAM), vramW)
		reason := truncate(r.Reason, reasonW)

		var verdictStr string
		switch r.Verdict {
		case models.VerdictCanRun:
			verdictStr = VerdictCanRun.Render(pad("CAN RUN", verdictW))
		case models.VerdictDegraded:
			verdictStr = VerdictDegraded.Render(pad("DEGRADED", verdictW))
		case models.VerdictCannotRun:
			verdictStr = VerdictCannotRun.Render(pad("CAN'T RUN", verdictW))
		}

		row := fmt.Sprintf("%s%-*s %-*s %-*s %s %-*s",
			prefix,
			nameW, name,
			quantW, quant,
			vramW, vram,
			verdictStr,
			reasonW, reason,
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(Legend())

	return b.String()
}
