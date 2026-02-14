package models

// EstimateVRAM calculates the estimated VRAM needed for a model.
// Formula: VRAM = parameters × bytes_per_param × (1 + overhead_percent/100)
func EstimateVRAM(paramsBillions float64, quant Quantization, overheadPercent float64) VRAMEstimate {
	bytesPerParam := quant.BytesPerParam()

	// Base VRAM: parameters * bytes_per_param
	// Parameters are in billions, convert to actual count
	baseBytes := uint64(paramsBillions * 1e9 * bytesPerParam)

	// Overhead for KV cache and runtime
	overheadBytes := uint64(float64(baseBytes) * (overheadPercent / 100.0))

	return VRAMEstimate{
		ModelParams:  paramsBillions,
		Quantization: quant,
		BaseVRAM:     baseBytes,
		Overhead:     overheadBytes,
		TotalVRAM:    baseBytes + overheadBytes,
	}
}

// EstimateVRAMMiB returns estimated VRAM in MiB.
func EstimateVRAMMiB(paramsBillions float64, quant Quantization, overheadPercent float64) uint64 {
	est := EstimateVRAM(paramsBillions, quant, overheadPercent)
	return est.TotalVRAM / (1024 * 1024)
}
