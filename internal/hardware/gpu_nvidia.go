package hardware

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func detectNvidiaGPUs() []GPUInfo {
	out, err := exec.Command("nvidia-smi",
		"--query-gpu=name,memory.total,driver_version",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ", ", 3)
		if len(parts) < 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		vramMiB, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			continue
		}

		driver := ""
		if len(parts) >= 3 {
			driver = strings.TrimSpace(parts[2])
		}

		gpus = append(gpus, GPUInfo{
			Name:   fmt.Sprintf("NVIDIA %s", name),
			Vendor: "nvidia",
			VRAM:   vramMiB * 1024 * 1024,
			Driver: driver,
		})
	}

	return gpus
}
