package ollama

import "time"

// LocalModel represents a model installed locally in Ollama.
type LocalModel struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// TagsResponse is the response from GET /api/tags.
type TagsResponse struct {
	Models []LocalModel `json:"models"`
}

// PullRequest is the request body for POST /api/pull.
type PullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

// PullResponse is a single streaming response from POST /api/pull.
type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
}

// ShowRequest is the request body for POST /api/show.
type ShowRequest struct {
	Name string `json:"name"`
}

// ShowResponse is the response from POST /api/show.
type ShowResponse struct {
	ModelFile  string       `json:"modelfile"`
	Parameters string      `json:"parameters"`
	Template   string       `json:"template"`
	Details    ModelDetails `json:"details"`
}

// RemoteModel represents a model from ollamadb.dev.
type RemoteModel struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Pulls       int64    `json:"pulls"`
	Updated     string   `json:"updated"`
}

// RemoteModelsResponse is the response from ollamadb.dev API.
type RemoteModelsResponse struct {
	Models     []RemoteModel `json:"models"`
	Total      int           `json:"total"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
}
