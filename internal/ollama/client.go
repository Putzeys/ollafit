package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the local Ollama API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsRunning checks if the Ollama server is reachable.
func (c *Client) IsRunning() bool {
	resp, err := c.HTTPClient.Get(c.BaseURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ListLocalModels returns all locally installed models.
func (c *Client) ListLocalModels() ([]LocalModel, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("connecting to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var tagsResp TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return tagsResp.Models, nil
}

// ShowModel returns details about a specific model.
func (c *Client) ShowModel(name string) (*ShowResponse, error) {
	body, _ := json.Marshal(ShowRequest{Name: name})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/show", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connecting to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model %q not found (status %d)", name, resp.StatusCode)
	}

	var showResp ShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&showResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &showResp, nil
}

// PullProgressFunc is called for each progress update during a pull.
type PullProgressFunc func(status string, total, completed int64)

// PullModel downloads a model with streaming progress.
func (c *Client) PullModel(name string, progressFn PullProgressFunc) error {
	body, _ := json.Marshal(PullRequest{Name: name, Stream: true})

	// Use a longer timeout for pull operations
	client := &http.Client{Timeout: 0} // no timeout for downloads
	resp, err := client.Post(c.BaseURL+"/api/pull", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("connecting to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pull failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var pullResp PullResponse
		if err := json.Unmarshal(scanner.Bytes(), &pullResp); err != nil {
			continue
		}
		if progressFn != nil {
			progressFn(pullResp.Status, pullResp.Total, pullResp.Completed)
		}
	}

	return scanner.Err()
}
