package hardware

import (
	"github.com/shirou/gopsutil/v4/mem"
)

func detectMemory() (MemoryInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}, err
	}

	return MemoryInfo{
		Total:     v.Total,
		Available: v.Available,
	}, nil
}
