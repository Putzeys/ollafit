package ollama

import (
	"testing"
)

func TestParseModelParams(t *testing.T) {
	tests := []struct {
		name string
		want float64
	}{
		{"llama3.2:1b", 1},
		{"llama3.2:3b", 3},
		{"llama3.1:8b", 8},
		{"llama3.1:70b", 70},
		{"mistral:7b", 7},
		{"mixtral:8x7b", 56},
		{"phi3:3.8b", 3.8},
		{"gemma2:9b", 9},
		{"gemma2:27b", 27},
		{"qwen2.5:72b", 72},
		{"codellama:13b", 13},
		{"deepseek-r1:7b", 7},
		{"deepseek-r1:70b", 70},
		{"unknown-model", 7}, // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseModelParams(tt.name)
			if got != tt.want {
				t.Errorf("ParseModelParams(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetPopularModels(t *testing.T) {
	models := GetPopularModels()
	if len(models) == 0 {
		t.Error("GetPopularModels() returned empty list")
	}
	for _, m := range models {
		if m.Name == "" {
			t.Error("model with empty name")
		}
	}
}
