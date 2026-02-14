package hardware

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func detectAMDGPUs() []GPUInfo {
	out, err := exec.Command("rocm-smi", "--showmeminfo", "vram", "--csv").Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Parse CSV output — look for total VRAM
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "device") && strings.Contains(strings.ToLower(line), "total") {
			continue // skip header
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		vramStr := strings.TrimSpace(fields[len(fields)-1])
		vramBytes, err := strconv.ParseUint(vramStr, 10, 64)
		if err != nil {
			continue
		}

		name := detectAMDGPUName()
		gpus = append(gpus, GPUInfo{
			Name:   name,
			Vendor: "amd",
			VRAM:   vramBytes,
		})
	}

	// Fallback: parse non-CSV output
	if len(gpus) == 0 {
		gpus = parseAMDFallback(string(out))
	}

	return gpus
}

func detectAMDGPUName() string {
	out, err := exec.Command("rocm-smi", "--showproductname").Output()
	if err != nil {
		return "AMD GPU"
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Card Series") || strings.Contains(line, "Card Model") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				if name != "" {
					return "AMD " + name
				}
			}
		}
	}
	return "AMD GPU"
}

func parseAMDFallback(output string) []GPUInfo {
	re := regexp.MustCompile(`(?i)total\s+.*?(\d+)\s*bytes`)
	matches := re.FindAllStringSubmatch(output, -1)
	var gpus []GPUInfo
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		vram, err := strconv.ParseUint(m[1], 10, 64)
		if err != nil {
			continue
		}
		gpus = append(gpus, GPUInfo{
			Name:   "AMD GPU",
			Vendor: "amd",
			VRAM:   vram,
		})
	}
	return gpus
}
