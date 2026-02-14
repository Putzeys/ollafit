package models

import (
	"fmt"

	"github.com/putzeys/putzeys-cli/internal/hardware"
)

// CheckCompatibility evaluates whether a model can run on the given hardware.
func CheckCompatibility(model ModelVariant, hw hardware.SystemInfo, overheadPercent float64, gpuMemFraction float64) CompatResult {
	estimate := EstimateVRAM(model.Parameters, model.Quantization, overheadPercent)
	required := estimate.TotalVRAM

	bestGPU := hw.BestGPU()
	totalVRAM := hw.TotalVRAM()
	totalRAM := hw.Memory.Total

	result := CompatResult{
		Model:     model,
		Estimate:  estimate,
		TotalVRAM: totalVRAM,
		TotalRAM:  totalRAM,
	}

	if bestGPU != nil {
		result.GPUName = bestGPU.Name
		result.GPUAvail = bestGPU.VRAM
	}

	// Apple Silicon: unified memory, use gpu_memory_fraction
	if bestGPU != nil && bestGPU.Unified {
		effectiveVRAM := uint64(float64(bestGPU.VRAM) * gpuMemFraction)

		if required <= effectiveVRAM {
			result.Verdict = VerdictCanRun
			result.Reason = fmt.Sprintf("Fits in unified memory (%.0f%% of %.1f GiB)",
				gpuMemFraction*100, float64(bestGPU.VRAM)/(1024*1024*1024))
			return result
		}

		if required <= bestGPU.VRAM {
			result.Verdict = VerdictDegraded
			result.Reason = fmt.Sprintf("Fits in unified memory but exceeds %.0f%% threshold (may use swap)",
				gpuMemFraction*100)
			return result
		}

		result.Verdict = VerdictCannotRun
		result.Reason = fmt.Sprintf("Exceeds total unified memory (%.1f GiB)",
			float64(bestGPU.VRAM)/(1024*1024*1024))
		return result
	}

	// 1. Check if fits in best single GPU
	if bestGPU != nil && required <= bestGPU.VRAM {
		result.Verdict = VerdictCanRun
		result.Reason = fmt.Sprintf("Fits in %s VRAM (%s)", bestGPU.Name, hardware.FormatBytes(bestGPU.VRAM))
		return result
	}

	// 2. Check if fits distributed across multiple GPUs
	if len(hw.GPUs) > 1 && required <= totalVRAM {
		result.Verdict = VerdictCanRun
		result.Reason = fmt.Sprintf("Fits across %d GPUs (total %s VRAM)", len(hw.GPUs), hardware.FormatBytes(totalVRAM))
		return result
	}

	// 3. Check if fits with VRAM + RAM offload
	if totalVRAM > 0 && required <= totalVRAM+totalRAM {
		result.Verdict = VerdictDegraded
		result.Reason = "Requires CPU/RAM offload (slower performance)"
		return result
	}

	// 4. No GPU — CPU-only mode
	if totalVRAM == 0 && required <= totalRAM {
		result.Verdict = VerdictDegraded
		result.Reason = "No GPU detected — CPU-only mode (very slow)"
		return result
	}

	// 5. Cannot run
	result.Verdict = VerdictCannotRun
	totalAvail := totalVRAM + totalRAM
	result.Reason = fmt.Sprintf("Insufficient memory: needs %s, have %s total",
		hardware.FormatBytes(required), hardware.FormatBytes(totalAvail))
	return result
}

// CheckAll evaluates multiple models against the hardware.
func CheckAll(models []ModelVariant, hw hardware.SystemInfo, overheadPercent float64, gpuMemFraction float64) []CompatResult {
	results := make([]CompatResult, 0, len(models))
	for _, m := range models {
		results = append(results, CheckCompatibility(m, hw, overheadPercent, gpuMemFraction))
	}
	return results
}
