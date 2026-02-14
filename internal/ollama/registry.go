package ollama

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultRegistryURL = "https://ollamadb.dev/api/v1/models"

// Registry discovers models from the remote ollamadb.dev registry.
type Registry struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewRegistry() *Registry {
	return &Registry{
		BaseURL: defaultRegistryURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SearchModels searches for models matching the query (single page).
func (r *Registry) SearchModels(query string, page int) ([]RemoteModel, error) {
	u, err := url.Parse(r.BaseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if query != "" {
		q.Set("q", query)
	}
	if page > 0 {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	u.RawQuery = q.Encode()

	resp, err := r.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("fetching models from registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var modelsResp RemoteModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("decoding registry response: %w", err)
	}

	return modelsResp.Models, nil
}

// FetchAllModels fetches all models from the registry using limit/skip pagination.
// Returns all available models sorted by popularity (pulls).
func (r *Registry) FetchAllModels(batchSize int) ([]RemoteModel, error) {
	if batchSize <= 0 {
		batchSize = 200
	}

	var all []RemoteModel
	skip := 0

	for {
		u, err := url.Parse(r.BaseURL)
		if err != nil {
			return nil, err
		}

		q := u.Query()
		q.Set("limit", strconv.Itoa(batchSize))
		q.Set("skip", strconv.Itoa(skip))
		q.Set("sort", "pulls")
		u.RawQuery = q.Encode()

		resp, err := r.HTTPClient.Get(u.String())
		if err != nil {
			return nil, fmt.Errorf("fetching models from registry: %w", err)
		}

		var modelsResp RemoteModelsResponse
		if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding registry response: %w", err)
		}
		resp.Body.Close()

		if len(modelsResp.Models) == 0 {
			break
		}

		all = append(all, modelsResp.Models...)
		skip += len(modelsResp.Models)

		// Stop if we got everything
		total := modelsResp.TotalCount
		if total == 0 {
			total = modelsResp.Total
		}
		if total > 0 && skip >= total {
			break
		}

		// Safety: stop if batch returned fewer than requested
		if len(modelsResp.Models) < batchSize {
			break
		}
	}

	return all, nil
}

var (
	reModelLink = regexp.MustCompile(`href="/library/([^"]+)"`)
	reModelDesc = regexp.MustCompile(`<p class="max-w-lg break-words[^"]*">([^<]+)</p>`)
	reModelTags = regexp.MustCompile(`<span[^>]*>(\d+(?:\.\d+)?[bBmM])</span>`)
	reModelBlock = regexp.MustCompile(`(?s)<li x-test-model.*?</li>`)
)

// FetchOllamaLibrary scrapes ollama.com/library to get all available models.
// This is used as primary source since the ollamadb.dev API may be unavailable.
func FetchOllamaLibrary() ([]RemoteModel, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://ollama.com/library")
	if err != nil {
		return nil, fmt.Errorf("fetching ollama library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama.com returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	html := string(body)
	blocks := reModelBlock.FindAllString(html, -1)

	var models []RemoteModel
	for _, block := range blocks {
		nameMatch := reModelLink.FindStringSubmatch(block)
		if nameMatch == nil {
			continue
		}
		name := nameMatch[1]

		desc := ""
		descMatch := reModelDesc.FindStringSubmatch(block)
		if descMatch != nil {
			desc = strings.TrimSpace(descMatch[1])
			desc = strings.ReplaceAll(desc, "&#39;", "'")
			desc = strings.ReplaceAll(desc, "&amp;", "&")
			desc = strings.ReplaceAll(desc, "&quot;", "\"")
		}

		tagMatches := reModelTags.FindAllStringSubmatch(block, -1)
		var tags []string
		for _, tm := range tagMatches {
			tag := strings.ToLower(tm[1])
			// Only keep parameter size tags (ending in 'b'), skip pull counts (ending in 'm')
			if strings.HasSuffix(tag, "b") {
				tags = append(tags, tag)
			}
		}

		// Expand each model into variants by tag (e.g. llama3.1 → llama3.1:8b, llama3.1:70b, llama3.1:405b)
		if len(tags) > 0 {
			for _, tag := range tags {
				models = append(models, RemoteModel{
					Name:        name + ":" + tag,
					Description: desc,
					Tags:        tags,
				})
			}
		} else {
			// No size tags found, add as-is
			models = append(models, RemoteModel{
				Name:        name,
				Description: desc,
			})
		}
	}

	return models, nil
}

// GetPopularModels returns a curated list of popular models with their known parameter sizes.
// Used as fallback when the registry is unavailable.
func GetPopularModels() []RemoteModel {
	return []RemoteModel{
		{Name: "llama3.2:1b", Description: "Llama 3.2 1B - Compact and fast", Tags: []string{"1b", "q4_k_m"}},
		{Name: "llama3.2:3b", Description: "Llama 3.2 3B - Good balance", Tags: []string{"3b", "q4_k_m"}},
		{Name: "llama3.1:8b", Description: "Llama 3.1 8B - Great general purpose", Tags: []string{"8b", "q4_k_m"}},
		{Name: "llama3.1:70b", Description: "Llama 3.1 70B - High quality", Tags: []string{"70b", "q4_k_m"}},
		{Name: "mistral:7b", Description: "Mistral 7B - Fast and capable", Tags: []string{"7b", "q4_k_m"}},
		{Name: "mixtral:8x7b", Description: "Mixtral 8x7B - MoE model", Tags: []string{"47b", "q4_k_m"}},
		{Name: "codellama:7b", Description: "Code Llama 7B - Code generation", Tags: []string{"7b", "q4_k_m"}},
		{Name: "codellama:13b", Description: "Code Llama 13B - Better code generation", Tags: []string{"13b", "q4_k_m"}},
		{Name: "phi3:3.8b", Description: "Phi-3 3.8B - Microsoft's compact model", Tags: []string{"3.8b", "q4_k_m"}},
		{Name: "gemma2:9b", Description: "Gemma 2 9B - Google's model", Tags: []string{"9b", "q4_k_m"}},
		{Name: "gemma2:27b", Description: "Gemma 2 27B - Larger Google model", Tags: []string{"27b", "q4_k_m"}},
		{Name: "qwen2.5:7b", Description: "Qwen 2.5 7B - Alibaba's model", Tags: []string{"7b", "q4_k_m"}},
		{Name: "qwen2.5:72b", Description: "Qwen 2.5 72B - Large Alibaba model", Tags: []string{"72b", "q4_k_m"}},
		{Name: "deepseek-r1:7b", Description: "DeepSeek R1 7B - Reasoning model", Tags: []string{"7b", "q4_k_m"}},
		{Name: "deepseek-r1:70b", Description: "DeepSeek R1 70B - Large reasoning", Tags: []string{"70b", "q4_k_m"}},
	}
}

// ParseModelParams extracts parameter count in billions from a model name or tag.
func ParseModelParams(name string) float64 {
	lower := strings.ToLower(name)

	// Handle MoE patterns like "8x7b" → 8*7 = 56 effective params (but ~47B active)
	if idx := strings.LastIndex(lower, "x"); idx > 0 {
		parts := strings.FieldsFunc(lower, func(r rune) bool {
			return r == ':' || r == '-'
		})
		for _, p := range parts {
			if xIdx := strings.Index(p, "x"); xIdx > 0 && strings.HasSuffix(p, "b") {
				// e.g., "8x7b"
				numStr := strings.TrimSuffix(p, "b")
				xParts := strings.SplitN(numStr, "x", 2)
				if len(xParts) == 2 {
					n1, e1 := strconv.ParseFloat(xParts[0], 64)
					n2, e2 := strconv.ParseFloat(xParts[1], 64)
					if e1 == nil && e2 == nil {
						return n1 * n2
					}
				}
			}
		}
	}

	// Extract the tag part after ":"
	tag := lower
	if colonIdx := strings.LastIndex(lower, ":"); colonIdx >= 0 {
		tag = lower[colonIdx+1:]
	}

	// Try to parse tag directly as parameter size (e.g., "7b", "70b", "3.8b")
	if strings.HasSuffix(tag, "b") {
		numStr := strings.TrimSuffix(tag, "b")
		// Strip quantization suffixes like "-q4_k_m"
		if dashIdx := strings.Index(numStr, "-"); dashIdx >= 0 {
			numStr = numStr[:dashIdx]
		}
		if val, err := strconv.ParseFloat(numStr, 64); err == nil {
			return val
		}
	}

	// Fallback: scan the whole name for known sizes
	knownSizes := []struct {
		suffix string
		params float64
	}{
		{"405b", 405}, {"72b", 72}, {"70b", 70},
		{"47b", 47}, {"34b", 34}, {"32b", 32}, {"27b", 27},
		{"14b", 14}, {"13b", 13}, {"9b", 9}, {"8b", 8}, {"7b", 7},
		{"3.8b", 3.8}, {"3b", 3}, {"2b", 2}, {"1.5b", 1.5}, {"1b", 1},
	}

	for _, ks := range knownSizes {
		if strings.HasSuffix(lower, ":"+ks.suffix) {
			return ks.params
		}
	}

	return 7 // default assumption
}
