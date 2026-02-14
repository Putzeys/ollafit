package models

import (
	"testing"
)

func TestEstimateVRAM(t *testing.T) {
	tests := []struct {
		name     string
		params   float64
		quant    Quantization
		overhead float64
		wantMiB  uint64 // approximate, allow ±5%
	}{
		{"7B Q4_K_M", 7, QuantQ4KM, 20.0, 4564},
		{"13B Q4_K_M", 13, QuantQ4KM, 20.0, 8482},
		{"70B Q4_K_M", 70, QuantQ4KM, 20.0, 45638},
		{"7B Q8_0", 7, QuantQ8, 20.0, 8003},
		{"7B FP16", 7, QuantFP16, 20.0, 16006},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			est := EstimateVRAM(tt.params, tt.quant, tt.overhead)
			gotMiB := est.TotalVRAM / (1024 * 1024)

			// Allow 5% tolerance
			lower := uint64(float64(tt.wantMiB) * 0.95)
			upper := uint64(float64(tt.wantMiB) * 1.05)

			if gotMiB < lower || gotMiB > upper {
				t.Errorf("EstimateVRAM(%v, %v, %v) = %d MiB, want ~%d MiB",
					tt.params, tt.quant, tt.overhead, gotMiB, tt.wantMiB)
			}
		})
	}
}

func TestParseQuantization(t *testing.T) {
	tests := []struct {
		input string
		want  Quantization
	}{
		{"Q4_K_M", QuantQ4KM},
		{"q4_k_m", QuantQ4KM},
		{"Q8_0", QuantQ8},
		{"FP16", QuantFP16},
		{"f16", QuantFP16},
		{"unknown", QuantQ4KM}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseQuantization(tt.input)
			if got != tt.want {
				t.Errorf("ParseQuantization(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
