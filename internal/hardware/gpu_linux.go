//go:build linux

package hardware

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func detectGPUs() []GPUInfo {
	// Try NVIDIA first (highest priority)
	if gpus := detectNvidiaGPUs(); len(gpus) > 0 {
		return gpus
	}

	// Try AMD via rocm-smi
	if gpus := detectAMDGPUs(); len(gpus) > 0 {
		return gpus
	}

	// Fallback: detect GPUs via lspci (works without vendor drivers)
	if gpus := detectGPUsLspci(); len(gpus) > 0 {
		return gpus
	}

	return nil
}

// detectGPUsLspci parses lspci output to find VGA/3D controllers.
// This detects GPUs even when vendor tools (nvidia-smi, rocm-smi) are absent.
func detectGPUsLspci() []GPUInfo {
	out, err := exec.Command("lspci").Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	re := regexp.MustCompile(`(?i)(VGA compatible controller|3D controller|Display controller):\s*(.+)`)

	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := strings.TrimSpace(m[2])

		vendor := "unknown"
		nameLower := strings.ToLower(name)
		switch {
		case strings.Contains(nameLower, "nvidia"):
			vendor = "nvidia"
		case strings.Contains(nameLower, "amd") || strings.Contains(nameLower, "radeon") || strings.Contains(nameLower, "advanced micro"):
			vendor = "amd"
		case strings.Contains(nameLower, "intel"):
			vendor = "intel"
		}

		gpu := GPUInfo{
			Name:   name,
			Vendor: vendor,
		}

		// Try to read VRAM from sysfs
		if vram := readDRMVRAM(vendor); vram > 0 {
			gpu.VRAM = vram
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// readDRMVRAM reads VRAM size from sysfs DRM nodes.
func readDRMVRAM(vendor string) uint64 {
	// Check /sys/class/drm/card*/device/mem_info_vram_total (AMD)
	matches, _ := filepath.Glob("/sys/class/drm/card[0-9]*/device/mem_info_vram_total")
	for _, path := range matches {
		out, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		val, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err == nil && val > 0 {
			return val
		}
	}

	// Fallback: parse resource file for BAR sizes
	resMatches, _ := filepath.Glob("/sys/class/drm/card[0-9]*/device/resource")
	for _, path := range resMatches {
		out, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if vram := parsePCIResource(string(out)); vram > 0 {
			return vram
		}
	}

	return 0
}

// parsePCIResource parses a PCI resource file to find the largest BAR (likely VRAM).
func parsePCIResource(content string) uint64 {
	var maxSize uint64
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		start, err1 := strconv.ParseUint(strings.TrimPrefix(fields[0], "0x"), 16, 64)
		end, err2 := strconv.ParseUint(strings.TrimPrefix(fields[1], "0x"), 16, 64)
		if err1 != nil || err2 != nil || start == 0 || end <= start {
			continue
		}
		size := end - start + 1
		if size > maxSize {
			maxSize = size
		}
	}
	return maxSize
}
