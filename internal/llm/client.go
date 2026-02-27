package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// debugCallCounter tracks the number of LLM calls for FASTCODE_DEBUG_PROMPT_DIR logging.
var debugCallCounter uint64

// Client is an OpenAI-compatible LLM API client.
type Client struct {
	APIKey           string
	Model            string
	BaseURL          string
	EmbeddingBaseURL string // Separate base URL for embeddings (optional)
	HTTP             *http.Client
}

// NewClient creates a new LLM client from environment variables.
func NewClient() *Client {
	baseURL := getEnvOr("BASE_URL", "https://api.openai.com/v1")
	return &Client{
		APIKey:           os.Getenv("OPENAI_API_KEY"),
		Model:            getEnvOr("MODEL", "gpt-4o"),
		BaseURL:          baseURL,
		EmbeddingBaseURL: getEnvOr("EMBEDDING_URL", baseURL),
		HTTP: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// NewClientWith creates a client with explicit parameters.
func NewClientWith(apiKey, model, baseURL string) *Client {
	return &Client{
		APIKey:           apiKey,
		Model:            model,
		BaseURL:          baseURL,
		EmbeddingBaseURL: baseURL,
		HTTP:             &http.Client{Timeout: 120 * time.Second},
	}
}

// --- Chat Completion ---

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ChatCompletion sends a chat completion request and returns the response text.
func (c *Client) ChatCompletion(messages []ChatMessage, temperature float64, maxTokens int) (string, error) {
	req := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// --- Mode 1: Single-prompt abort (existing behaviour) ---
	if dumpFile := os.Getenv("FASTCODE_DEBUG_PROMPT_FILE"); dumpFile != "" {
		data, err := json.MarshalIndent(req, "", "  ")
		if err == nil {
			_ = os.WriteFile(dumpFile, data, 0644)
		}
		return "DEBUG_PROMPT_WRITTEN", nil
	}

	// --- Mode 2: Full-flow logging (log every call, don't abort) ---
	dumpDir := os.Getenv("FASTCODE_DEBUG_PROMPT_DIR")
	var callNum uint64
	if dumpDir != "" {
		callNum = atomic.AddUint64(&debugCallCounter, 1)
		_ = os.MkdirAll(dumpDir, 0755)
		reqPath := filepath.Join(dumpDir, fmt.Sprintf("call_%03d_request.json", callNum))
		data, err := json.MarshalIndent(req, "", "  ")
		if err == nil {
			_ = os.WriteFile(reqPath, data, 0644)
		}
	}

	body, err := c.post("/chat/completions", req)
	if err != nil {
		return "", err
	}

	var resp chatResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse chat response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("API error: %s", resp.Error.Message)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	// Log response in full-flow mode
	if dumpDir != "" {
		respPath := filepath.Join(dumpDir, fmt.Sprintf("call_%03d_response.json", callNum))
		respData, err := json.MarshalIndent(resp, "", "  ")
		if err == nil {
			_ = os.WriteFile(respPath, respData, 0644)
		}
	}

	return resp.Choices[0].Message.Content, nil
}

// --- Embeddings ---

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed generates embedding vectors for the given texts.
func (c *Client) Embed(texts []string, model string) ([][]float32, error) {
	if model == "" {
		model = "text-embedding-3-small"
	}

	req := embeddingRequest{
		Model: model,
		Input: texts,
	}

	var url string
	if strings.HasSuffix(c.EmbeddingBaseURL, "/embeddings") {
		url = c.EmbeddingBaseURL
	} else {
		url = strings.TrimSuffix(c.EmbeddingBaseURL, "/") + "/embeddings"
	}

	body, err := c.postTo(url, "", req)
	if err != nil {
		return nil, err
	}

	var resp embeddingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse embedding response: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("API error: %s", resp.Error.Message)
	}

	// Sort by index to maintain order
	result := make([][]float32, len(texts))
	for _, d := range resp.Data {
		if d.Index < len(result) {
			result[d.Index] = d.Embedding
		}
	}

	return result, nil
}

// --- HTTP helper ---

func (c *Client) post(path string, payload any) ([]byte, error) {
	return c.postTo(c.BaseURL, path, payload)
}

func (c *Client) postTo(baseURL, path string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
