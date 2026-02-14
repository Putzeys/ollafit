//go:build linux

package hardware

func detectGPUs() []GPUInfo {
	// Try NVIDIA first (highest priority)
	if gpus := detectNvidiaGPUs(); len(gpus) > 0 {
		return gpus
	}

	// Try AMD via rocm-smi
	if gpus := detectAMDGPUs(); len(gpus) > 0 {
		return gpus
	}

	return nil
}
