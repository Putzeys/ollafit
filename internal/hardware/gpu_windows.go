//go:build windows

package hardware

func detectGPUs() []GPUInfo {
	// Try NVIDIA first
	if gpus := detectNvidiaGPUs(); len(gpus) > 0 {
		return gpus
	}

	// Fallback: WMI (limited to 4GB VRAM reporting)
	return detectWMIGPUs()
}

func detectWMIGPUs() []GPUInfo {
	// WMI detection via wmi package
	type win32VideoController struct {
		Name          string
		AdapterRAM    uint32
		DriverVersion string
	}

	// Use wmi.Query for Win32_VideoController
	// Note: WMI reports AdapterRAM as uint32, capping at 4GB
	// This is a known Windows limitation

	// We import wmi conditionally; if not available, return nil
	return detectWMIGPUsImpl()
}
