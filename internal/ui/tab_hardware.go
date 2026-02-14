package ui

import (
	"github.com/putzeys/putzeys-cli/internal/hardware"
)

// HardwareTab holds state for the hardware detection tab.
type HardwareTab struct {
	hw       *hardware.SystemInfo
	err      error
	loading  bool
	scrollY  int
	rendered string
}

func NewHardwareTab() HardwareTab {
	return HardwareTab{loading: true}
}

func (t *HardwareTab) SetHardware(hw hardware.SystemInfo) {
	t.hw = &hw
	t.loading = false
	t.err = nil
	t.rendered = RenderHardwareInfo(hw)
}

func (t *HardwareTab) SetError(err error) {
	t.err = err
	t.loading = false
}

func (t HardwareTab) View(width, height int) string {
	if t.loading {
		return DimStyle.Render("  Detecting hardware...")
	}
	if t.err != nil {
		return ErrorStyle.Render("  Error: " + t.err.Error())
	}
	if t.hw == nil {
		return DimStyle.Render("  No hardware info available.")
	}
	return t.rendered
}
