package hardware

import (
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"
)

func detectCPU() (CPUInfo, error) {
	info := CPUInfo{
		Cores:   runtime.NumCPU(),
		Threads: runtime.NumCPU(),
	}

	cpus, err := cpu.Info()
	if err != nil {
		return info, nil // fallback to partial info
	}

	if len(cpus) > 0 {
		info.Model = cpus[0].ModelName
		info.Cores = int(cpus[0].Cores)

		// Count logical processors (threads)
		counts, err := cpu.Counts(true)
		if err == nil {
			info.Threads = counts
		}
	}

	return info, nil
}
