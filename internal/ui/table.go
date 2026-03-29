package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/putzeys/ollafit/internal/hardware"
	"github.com/putzeys/ollafit/internal/models"
)

// RenderHardwareInfo renders the hardware detection results.
func RenderHardwareInfo(info hardware.SystemInfo) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Hardware Detection"))
	b.WriteString("\n\n")

	// CPU
	b.WriteString(LabelStyle.Render("CPU: "))
	b.WriteString(ValueStyle.Render(info.CPU.Model))
	b.WriteString(DimStyle.Render(fmt.Sprintf(" (%d cores, %d threads)", info.CPU.Cores, info.CPU.Threads)))
	b.WriteString("\n")

	// RAM
	b.WriteString(LabelStyle.Render("RAM: "))
	b.WriteString(ValueStyle.Render(hardware.FormatBytes(info.Memory.Total)))
	b.WriteString(DimStyle.Render(fmt.Sprintf(" (%s available)", hardware.FormatBytes(info.Memory.Available))))
	b.WriteString("\n")

	// GPUs
	if len(info.GPUs) == 0 {
		b.WriteString(LabelStyle.Render("GPU: "))
		b.WriteString(WarningStyle.Render("None detected"))
		b.WriteString("\n")
	} else {
		for i, gpu := range info.GPUs {
			label := "GPU: "
			if len(info.GPUs) > 1 {
				label = fmt.Sprintf("GPU %d: ", i+1)
			}
			b.WriteString(LabelStyle.Render(label))
			b.WriteString(ValueStyle.Render(gpu.Name))

			details := hardware.FormatBytes(gpu.VRAM)
			if gpu.Unified {
				details += " (unified memory)"
			}
			if gpu.Driver != "" {
				details += fmt.Sprintf(", driver %s", gpu.Driver)
			}
			b.WriteString(DimStyle.Render(fmt.Sprintf(" [%s]", details)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RenderCompatTable renders the compatibility results as a colored table.
func RenderCompatTable(results []models.CompatResult) string {
	if len(results) == 0 {
		return WarningStyle.Render("No models to check.")
	}

	var b strings.Builder

	b.WriteString(TitleStyle.Render("Model Compatibility"))
	b.WriteString("\n\n")

	// Calculate column widths
	nameW, quantW, vramW, verdictW, reasonW := 25, 8, 12, 12, 40

	// Header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		nameW, "MODEL",
		quantW, "QUANT",
		vramW, "VRAM NEEDED",
		verdictW, "STATUS",
		reasonW, "DETAILS",
	)
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Separator
	sep := strings.Repeat("─", nameW+quantW+vramW+verdictW+reasonW+4)
	b.WriteString(DimStyle.Render(sep))
	b.WriteString("\n")

	// Rows
	for _, r := range results {
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

		row := fmt.Sprintf("%-*s %-*s %-*s %s %-*s",
			nameW, name,
			quantW, quant,
			vramW, vram,
			verdictStr,
			reasonW, reason,
		)
		b.WriteString(CellStyle.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}

// RenderCompatJSON renders results as JSON-friendly output.
func RenderCompatJSON(results []models.CompatResult) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		out = append(out, map[string]interface{}{
			"model":        r.Model.Name,
			"parameters":   r.Model.Parameters,
			"quantization": r.Model.Quantization.String(),
			"vram_needed":  r.Estimate.TotalVRAM,
			"verdict":      string(r.Verdict),
			"reason":       r.Reason,
			"gpu":          r.GPUName,
		})
	}
	return out
}

// RenderModelList renders a list of models.
func RenderModelList(modelNames []string, descriptions []string) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Available Models"))
	b.WriteString("\n\n")

	nameW := 30
	header := fmt.Sprintf("%-*s %s", nameW, "MODEL", "DESCRIPTION")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	sep := strings.Repeat("─", nameW+50)
	b.WriteString(DimStyle.Render(sep))
	b.WriteString("\n")

	for i, name := range modelNames {
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		row := fmt.Sprintf("%-*s %s", nameW, truncate(name, nameW), desc)
		b.WriteString(CellStyle.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}

// Legend renders the color legend.
func Legend() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		VerdictCanRun.Render("● CAN RUN"),
		"  ",
		VerdictDegraded.Render("● DEGRADED"),
		"  ",
		VerdictCannotRun.Render("● CAN'T RUN"),
	)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
