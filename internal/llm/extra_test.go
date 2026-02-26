package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestPostMarshalError tests post with unmarshalable payload
func TestPostMarshalError(t *testing.T) {
	client := &Client{
		BaseURL: "http://localhost:9999",
		APIKey:  "test",
		HTTP:    http.DefaultClient,
	}

	// A channel is not JSON-marshalable
	_, err := client.post("/test", make(chan int))
	if err == nil {
		t.Error("expected marshal error")
	}
}

// TestPostHTTPConnectionError tests post when HTTP request fails
func TestPostHTTPConnectionError(t *testing.T) {
	client := &Client{
		BaseURL: "http://localhost:1", // unreachable port
		APIKey:  "test",
		HTTP:    http.DefaultClient,
	}

	_, err := client.post("/test", map[string]string{"key": "val"})
	if err == nil {
		t.Error("expected HTTP connection error")
	}
}

// TestPostHTTPStatusError tests post with 4xx/5xx response
func TestPostHTTPStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    http.DefaultClient,
	}

	_, err := client.post("/test", map[string]string{"key": "val"})
	if err == nil {
		t.Error("expected HTTP 429 error")
	}
}

// TestPostSuccessWithAuth tests successful post with auth header
func TestPostSuccessWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong auth header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing content-type")
		}
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test-key",
		HTTP:    http.DefaultClient,
	}

	body, err := client.post("/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if string(body) != `{"result":"ok"}` {
		t.Errorf("body = %q", string(body))
	}
}

// TestPostNoAuthHeader tests post without API key (no auth header)
func TestPostNoAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("should not have auth header when no API key")
		}
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "",
		HTTP:    http.DefaultClient,
	}

	_, err := client.post("/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("post no key: %v", err)
	}
}

// TestGetEnvOrExistsExtra tests getEnvOr when env var exists
func TestGetEnvOrExistsExtra(t *testing.T) {
	os.Setenv("TEST_LLM_EXTRA_VAR", "custom_value")
	defer os.Unsetenv("TEST_LLM_EXTRA_VAR")

	result := getEnvOr("TEST_LLM_EXTRA_VAR", "fallback")
	if result != "custom_value" {
		t.Errorf("getEnvOr = %q, want custom_value", result)
	}
}

// TestGetEnvOrFallbackExtra tests getEnvOr when env var is missing
func TestGetEnvOrFallbackExtra(t *testing.T) {
	os.Unsetenv("TEST_LLM_EXTRA_NONEXISTENT")

	result := getEnvOr("TEST_LLM_EXTRA_NONEXISTENT", "fallback")
	if result != "fallback" {
		t.Errorf("getEnvOr = %q, want fallback", result)
	}
}

// TestChatCompletionServerError tests ChatCompletion with server errors
func TestChatCompletionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
		HTTP:    http.DefaultClient,
	}

	messages := []ChatMessage{{Role: "user", Content: "test"}}
	_, err := client.ChatCompletion(messages, 0.7, 1000)
	if err == nil {
		t.Error("expected error from 500 response")
	}
}

// TestChatCompletionEmptyChoicesExtra tests ChatCompletion with empty choices
func TestChatCompletionEmptyChoicesExtra(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"choices": []any{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		Model:   "test-model",
		HTTP:    http.DefaultClient,
	}

	messages := []ChatMessage{{Role: "user", Content: "test"}}
	_, err := client.ChatCompletion(messages, 0.7, 1000)
	if err == nil {
		t.Error("expected error for empty choices")
	}
}
