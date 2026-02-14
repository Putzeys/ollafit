package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/putzeys/putzeys-cli/internal/hardware"
	"github.com/putzeys/putzeys-cli/internal/models"
	"github.com/putzeys/putzeys-cli/internal/ollama"
)

// PullTab holds state for the pull/download tab.
type PullTab struct {
	hw              *hardware.SystemInfo
	input           string
	editing         bool
	compatResult    *models.CompatResult
	pulling         bool
	pullStatus      string
	pullTotal       int64
	pullDone        int64
	pullFinished    bool
	pullErr         error
	progress        progress.Model
	overheadPercent float64
	gpuMemFraction  float64
	ollamaHost      string
}

func NewPullTab(overheadPercent, gpuMemFraction float64, ollamaHost string) PullTab {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50
	return PullTab{
		editing:         true,
		progress:        p,
		overheadPercent: overheadPercent,
		gpuMemFraction:  gpuMemFraction,
		ollamaHost:      ollamaHost,
	}
}

func (t *PullTab) SetHardware(hw hardware.SystemInfo) {
	t.hw = &hw
}

func (t *PullTab) SetInput(s string) {
	t.input = s
	t.editing = true
	t.compatResult = nil
	t.pullFinished = false
	t.pullErr = nil
	t.pulling = false
}

func (t *PullTab) CheckCompat() {
	if t.hw == nil || t.input == "" {
		return
	}
	params := ollama.ParseModelParams(t.input)
	quant := models.QuantQ4KM
	result := models.CheckCompatibility(
		models.ModelVariant{
			Name:         t.input,
			Parameters:   params,
			Quantization: quant,
		},
		*t.hw,
		t.overheadPercent,
		t.gpuMemFraction,
	)
	t.compatResult = &result
}

func (t *PullTab) StartPull() {
	t.pulling = true
	t.pullStatus = "Starting download..."
	t.pullTotal = 0
	t.pullDone = 0
	t.pullFinished = false
	t.pullErr = nil
	t.editing = false
}

func (t *PullTab) UpdateProgress(status string, total, completed int64) {
	t.pullStatus = status
	t.pullTotal = total
	t.pullDone = completed
}

func (t *PullTab) FinishPull(err error) {
	t.pullFinished = true
	t.pullErr = err
	t.pulling = false
}

func (t *PullTab) Reset() {
	t.input = ""
	t.editing = true
	t.compatResult = nil
	t.pulling = false
	t.pullFinished = false
	t.pullErr = nil
}

func (t PullTab) View(width, height int) string {
	var b strings.Builder

	// Input
	b.WriteString("  ")
	b.WriteString(LabelStyle.Render("Model name: "))
	if t.editing {
		b.WriteString(ValueStyle.Render(t.input))
		b.WriteString(DimStyle.Render("_"))
		b.WriteString("\n")
		b.WriteString(DimStyle.Render("  Type a model name and press Enter (e.g. llama3.2:3b)"))
	} else {
		b.WriteString(ValueStyle.Render(t.input))
	}
	b.WriteString("\n\n")

	// Compat result
	if t.compatResult != nil {
		b.WriteString("  ")
		b.WriteString(LabelStyle.Render("Compatibility: "))
		switch t.compatResult.Verdict {
		case models.VerdictCanRun:
			b.WriteString(VerdictCanRun.Render("CAN RUN"))
		case models.VerdictDegraded:
			b.WriteString(VerdictDegraded.Render("DEGRADED"))
		case models.VerdictCannotRun:
			b.WriteString(VerdictCannotRun.Render("CAN'T RUN"))
		}
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(DimStyle.Render(t.compatResult.Reason))
		b.WriteString("\n")

		b.WriteString("  ")
		b.WriteString(LabelStyle.Render("VRAM needed: "))
		b.WriteString(ValueStyle.Render(hardware.FormatBytes(t.compatResult.Estimate.TotalVRAM)))
		b.WriteString("\n\n")

		if !t.pulling && !t.pullFinished {
			b.WriteString(DimStyle.Render("  Press Enter to start download, Esc to cancel"))
			b.WriteString("\n")
		}
	}

	// Pull progress
	if t.pulling {
		b.WriteString("  ")
		b.WriteString(LabelStyle.Render("Status: "))
		b.WriteString(t.pullStatus)
		b.WriteString("\n\n")
		b.WriteString("  ")
		b.WriteString(t.progress.View())

		if t.pullTotal > 0 {
			b.WriteString(DimStyle.Render(fmt.Sprintf(" %s / %s",
				formatSize(t.pullDone), formatSize(t.pullTotal))))
		}
		b.WriteString("\n")
	}

	// Pull finished
	if t.pullFinished {
		if t.pullErr != nil {
			b.WriteString("  ")
			b.WriteString(ErrorStyle.Render("Error: " + t.pullErr.Error()))
		} else {
			b.WriteString("  ")
			b.WriteString(SuccessStyle.Render("Download complete!"))
		}
		b.WriteString("\n\n")
		b.WriteString(DimStyle.Render("  Press Enter to pull another model"))
		b.WriteString("\n")
	}

	return b.String()
}
