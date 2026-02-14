//go:build darwin

package hardware

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

func detectGPUs() []GPUInfo {
	// Try NVIDIA first (external GPU)
	if gpus := detectNvidiaGPUs(); len(gpus) > 0 {
		return gpus
	}

	// Check for Apple Silicon
	if gpu := detectAppleSilicon(); gpu != nil {
		return []GPUInfo{*gpu}
	}

	// Fallback: system_profiler for Intel Macs with discrete GPUs
	return detectMacGPUsViaProfiler()
}

func detectAppleSilicon() *GPUInfo {
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err != nil {
		return nil
	}

	brand := strings.TrimSpace(string(out))
	if !strings.HasPrefix(brand, "Apple") {
		return nil
	}

	// Get total memory
	memOut, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return nil
	}

	totalMem, err := strconv.ParseUint(strings.TrimSpace(string(memOut)), 10, 64)
	if err != nil {
		return nil
	}

	return &GPUInfo{
		Name:    brand,
		Vendor:  "apple",
		VRAM:    totalMem, // Unified memory — full system RAM
		Unified: true,
	}
}

type systemProfilerOutput struct {
	SPDisplaysDataType []spDisplay `json:"SPDisplaysDataType"`
}

type spDisplay struct {
	Name       string `json:"_name"`
	ChipsetModel string `json:"sppci_model"`
	VRAM       string `json:"spdisplays_vram"`
	VRAMShared string `json:"spdisplays_vram_shared"`
	Vendor     string `json:"spdisplays_vendor"`
}

func detectMacGPUsViaProfiler() []GPUInfo {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return nil
	}

	var result []systemProfilerOutput
	if err := json.Unmarshal(out, &result); err != nil {
		// Try alternate format
		var single systemProfilerOutput
		if err := json.Unmarshal(out, &single); err != nil {
			return nil
		}
		result = []systemProfilerOutput{single}
	}

	var gpus []GPUInfo
	for _, r := range result {
		for _, d := range r.SPDisplaysDataType {
			vramStr := d.VRAM
			if vramStr == "" {
				vramStr = d.VRAMShared
			}
			vram := parseVRAMString(vramStr)

			name := d.ChipsetModel
			if name == "" {
				name = d.Name
			}

			gpus = append(gpus, GPUInfo{
				Name:   name,
				Vendor: classifyVendor(name),
				VRAM:   vram,
			})
		}
	}

	return gpus
}

func parseVRAMString(s string) uint64 {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Handle formats like "4 GB", "4096 MB", "4GB"
	s = strings.ReplaceAll(s, " ", "")

	if strings.HasSuffix(s, "gb") {
		numStr := strings.TrimSuffix(s, "gb")
		if n, err := strconv.ParseFloat(numStr, 64); err == nil {
			return uint64(n * 1024 * 1024 * 1024)
		}
	}
	if strings.HasSuffix(s, "mb") {
		numStr := strings.TrimSuffix(s, "mb")
		if n, err := strconv.ParseFloat(numStr, 64); err == nil {
			return uint64(n * 1024 * 1024)
		}
	}

	return 0
}

func classifyVendor(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce"):
		return "nvidia"
	case strings.Contains(lower, "amd") || strings.Contains(lower, "radeon"):
		return "amd"
	case strings.Contains(lower, "intel"):
		return "intel"
	case strings.Contains(lower, "apple"):
		return "apple"
	default:
		return "unknown"
	}
}
