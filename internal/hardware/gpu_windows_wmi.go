//go:build windows

package hardware

import (
	"strings"

	"github.com/yusufpapurcu/wmi"
)

type win32VideoController struct {
	Name          string
	AdapterRAM    uint32
	DriverVersion string
}

func detectWMIGPUsImpl() []GPUInfo {
	var controllers []win32VideoController
	err := wmi.Query("SELECT Name, AdapterRAM, DriverVersion FROM Win32_VideoController", &controllers)
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	for _, c := range controllers {
		if c.AdapterRAM == 0 {
			continue
		}

		vendor := "unknown"
		lower := strings.ToLower(c.Name)
		switch {
		case strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce"):
			vendor = "nvidia"
		case strings.Contains(lower, "amd") || strings.Contains(lower, "radeon"):
			vendor = "amd"
		case strings.Contains(lower, "intel"):
			vendor = "intel"
		}

		gpus = append(gpus, GPUInfo{
			Name:   c.Name,
			Vendor: vendor,
			VRAM:   uint64(c.AdapterRAM),
			Driver: c.DriverVersion,
		})
	}

	return gpus
}
