package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListLocalModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			resp := TagsResponse{
				Models: []LocalModel{
					{
						Name: "llama3.1:8b",
						Size: 4661224676,
						Details: ModelDetails{
							ParameterSize:     "8.0B",
							QuantizationLevel: "Q4_K_M",
							Family:            "llama",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	if !client.IsRunning() {
		t.Error("expected server to be running")
	}

	models, err := client.ListLocalModels()
	if err != nil {
		t.Fatalf("ListLocalModels() error: %v", err)
	}

	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}

	if models[0].Name != "llama3.1:8b" {
		t.Errorf("expected model name llama3.1:8b, got %s", models[0].Name)
	}
}

func TestIsRunningFalse(t *testing.T) {
	client := NewClient("http://localhost:1") // port unlikely to be used
	if client.IsRunning() {
		t.Error("expected server to not be running")
	}
}
