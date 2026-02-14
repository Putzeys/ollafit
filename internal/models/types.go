package models

// Verdict represents the compatibility verdict for a model.
type Verdict string

const (
	VerdictCanRun    Verdict = "CAN_RUN"
	VerdictDegraded  Verdict = "DEGRADED"
	VerdictCannotRun Verdict = "CANNOT_RUN"
)

// Quantization represents a model quantization level.
type Quantization string

const (
	QuantQ4KM Quantization = "Q4_K_M"
	QuantQ8   Quantization = "Q8_0"
	QuantFP16 Quantization = "FP16"
)

// BytesPerParam returns the bytes per parameter for a quantization level.
func (q Quantization) BytesPerParam() float64 {
	switch q {
	case QuantQ4KM:
		return 0.57
	case QuantQ8:
		return 1.00
	case QuantFP16:
		return 2.00
	default:
		return 0.57 // default to Q4_K_M
	}
}

func (q Quantization) String() string {
	return string(q)
}

// ParseQuantization parses a quantization string.
func ParseQuantization(s string) Quantization {
	switch s {
	case "Q4_K_M", "q4_k_m", "Q4_K_S", "q4_k_s", "Q4_0", "q4_0":
		return QuantQ4KM
	case "Q8_0", "q8_0", "Q8", "q8":
		return QuantQ8
	case "FP16", "fp16", "F16", "f16":
		return QuantFP16
	default:
		return QuantQ4KM
	}
}

// ModelVariant represents a specific model variant with its requirements.
type ModelVariant struct {
	Name         string
	Description  string
	Parameters   float64      // in billions
	Quantization Quantization
	VRAMRequired uint64       // estimated VRAM in bytes
	IsLocal      bool         // whether it's installed locally
}

// VRAMEstimate holds the VRAM estimation breakdown.
type VRAMEstimate struct {
	ModelParams  float64 // billions
	Quantization Quantization
	BaseVRAM     uint64  // bytes (params * bytes_per_param)
	Overhead     uint64  // bytes (KV cache + runtime)
	TotalVRAM    uint64  // bytes (base + overhead)
}

// CompatResult holds the compatibility check result.
type CompatResult struct {
	Model       ModelVariant
	Estimate    VRAMEstimate
	Verdict     Verdict
	Reason      string
	GPUName     string
	GPUAvail    uint64 // best GPU VRAM in bytes
	TotalVRAM   uint64 // total VRAM across all GPUs
	TotalRAM    uint64 // total system RAM
}
