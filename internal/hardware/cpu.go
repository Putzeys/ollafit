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

		// Physical cores
		physical, err := cpu.Counts(false)
		if err == nil && physical > 0 {
			info.Cores = physical
		}

		// Logical processors (threads)
		logical, err := cpu.Counts(true)
		if err == nil && logical > 0 {
			info.Threads = logical
		}
	}

	return info, nil
}
