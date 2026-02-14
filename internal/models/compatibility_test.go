package models

import (
	"testing"

	"github.com/putzeys/putzeys-cli/internal/hardware"
)

const GiB = 1024 * 1024 * 1024

func TestCheckCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		model       ModelVariant
		hw          hardware.SystemInfo
		wantVerdict Verdict
	}{
		{
			name: "fits in single GPU",
			model: ModelVariant{
				Name:         "llama3.1:8b",
				Parameters:   8,
				Quantization: QuantQ4KM,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "RTX 4090", Vendor: "nvidia", VRAM: 24 * GiB},
				},
				Memory: hardware.MemoryInfo{Total: 32 * GiB, Available: 16 * GiB},
			},
			wantVerdict: VerdictCanRun,
		},
		{
			name: "fits across multi-GPU",
			model: ModelVariant{
				Name:         "llama3.1:70b",
				Parameters:   70,
				Quantization: QuantQ4KM,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "RTX 4090", Vendor: "nvidia", VRAM: 24 * GiB},
					{Name: "RTX 4090", Vendor: "nvidia", VRAM: 24 * GiB},
				},
				Memory: hardware.MemoryInfo{Total: 64 * GiB, Available: 32 * GiB},
			},
			wantVerdict: VerdictCanRun,
		},
		{
			name: "degraded with offload",
			model: ModelVariant{
				Name:         "llama3.1:70b",
				Parameters:   70,
				Quantization: QuantQ4KM,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "RTX 3060", Vendor: "nvidia", VRAM: 12 * GiB},
				},
				Memory: hardware.MemoryInfo{Total: 64 * GiB, Available: 32 * GiB},
			},
			wantVerdict: VerdictDegraded,
		},
		{
			name: "cannot run insufficient memory",
			model: ModelVariant{
				Name:         "llama3.1:70b",
				Parameters:   70,
				Quantization: QuantFP16,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "RTX 3060", Vendor: "nvidia", VRAM: 12 * GiB},
				},
				Memory: hardware.MemoryInfo{Total: 16 * GiB, Available: 8 * GiB},
			},
			wantVerdict: VerdictCannotRun,
		},
		{
			name: "CPU only degraded",
			model: ModelVariant{
				Name:         "llama3.2:1b",
				Parameters:   1,
				Quantization: QuantQ4KM,
			},
			hw: hardware.SystemInfo{
				GPUs:   nil,
				Memory: hardware.MemoryInfo{Total: 16 * GiB, Available: 8 * GiB},
			},
			wantVerdict: VerdictDegraded,
		},
		{
			name: "Apple Silicon unified memory fits",
			model: ModelVariant{
				Name:         "llama3.1:8b",
				Parameters:   8,
				Quantization: QuantQ4KM,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "Apple M2 Pro", Vendor: "apple", VRAM: 32 * GiB, Unified: true},
				},
				Memory: hardware.MemoryInfo{Total: 32 * GiB, Available: 16 * GiB},
			},
			wantVerdict: VerdictCanRun,
		},
		{
			name: "Apple Silicon exceeds threshold degraded",
			model: ModelVariant{
				Name:         "gemma2:27b",
				Parameters:   27,
				Quantization: QuantQ8,
			},
			hw: hardware.SystemInfo{
				GPUs: []hardware.GPUInfo{
					{Name: "Apple M2 Pro", Vendor: "apple", VRAM: 32 * GiB, Unified: true},
				},
				Memory: hardware.MemoryInfo{Total: 32 * GiB, Available: 16 * GiB},
			},
			wantVerdict: VerdictDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckCompatibility(tt.model, tt.hw, 20.0, 0.75)
			if result.Verdict != tt.wantVerdict {
				t.Errorf("CheckCompatibility() verdict = %v, want %v (reason: %s)",
					result.Verdict, tt.wantVerdict, result.Reason)
			}
		})
	}
}
