package hardware

import "fmt"

type GPUInfo struct {
	Name     string
	Vendor   string // nvidia, amd, apple, intel, unknown
	VRAM     uint64 // bytes
	Driver   string
	Unified  bool   // true for Apple Silicon (shared memory)
}

type CPUInfo struct {
	Model   string
	Cores   int
	Threads int
}

type MemoryInfo struct {
	Total     uint64 // bytes
	Available uint64 // bytes
}

type SystemInfo struct {
	GPUs   []GPUInfo
	CPU    CPUInfo
	Memory MemoryInfo
}

func (g GPUInfo) VRAMMiB() uint64 {
	return g.VRAM / (1024 * 1024)
}

func (m MemoryInfo) TotalGiB() float64 {
	return float64(m.Total) / (1024 * 1024 * 1024)
}

func (m MemoryInfo) AvailableGiB() float64 {
	return float64(m.Available) / (1024 * 1024 * 1024)
}

func FormatBytes(b uint64) string {
	const (
		MiB = 1024 * 1024
		GiB = 1024 * MiB
	)
	if b >= GiB {
		return fmt.Sprintf("%.1f GiB", float64(b)/float64(GiB))
	}
	return fmt.Sprintf("%d MiB", b/MiB)
}

// TotalVRAM returns the sum of VRAM across all GPUs.
func (s SystemInfo) TotalVRAM() uint64 {
	var total uint64
	for _, g := range s.GPUs {
		total += g.VRAM
	}
	return total
}

// BestGPU returns the GPU with the most VRAM, or nil if no GPUs.
func (s SystemInfo) BestGPU() *GPUInfo {
	if len(s.GPUs) == 0 {
		return nil
	}
	best := &s.GPUs[0]
	for i := 1; i < len(s.GPUs); i++ {
		if s.GPUs[i].VRAM > best.VRAM {
			best = &s.GPUs[i]
		}
	}
	return best
}

// DetectAll gathers all hardware information.
func DetectAll() (SystemInfo, error) {
	var info SystemInfo

	cpuInfo, err := detectCPU()
	if err != nil {
		return info, fmt.Errorf("detecting CPU: %w", err)
	}
	info.CPU = cpuInfo

	memInfo, err := detectMemory()
	if err != nil {
		return info, fmt.Errorf("detecting memory: %w", err)
	}
	info.Memory = memInfo

	gpus := detectGPUs()
	info.GPUs = gpus

	return info, nil
}
