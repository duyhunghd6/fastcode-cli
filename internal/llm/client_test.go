package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Save and restore env
	origKey := os.Getenv("OPENAI_API_KEY")
	origModel := os.Getenv("MODEL")
	origBase := os.Getenv("BASE_URL")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("MODEL", origModel)
		os.Setenv("BASE_URL", origBase)
	}()

	os.Setenv("OPENAI_API_KEY", "test-key-123")
	os.Setenv("MODEL", "test-model")
	os.Setenv("BASE_URL", "http://test.local")

	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.APIKey != "test-key-123" {
		t.Errorf("APIKey = %q, want test-key-123", client.APIKey)
	}
	if client.Model != "test-model" {
		t.Errorf("Model = %q, want test-model", client.Model)
	}
	if client.BaseURL != "http://test.local" {
		t.Errorf("BaseURL = %q, want http://test.local", client.BaseURL)
	}
}

func TestNewClientDefaults(t *testing.T) {
	origKey := os.Getenv("OPENAI_API_KEY")
	origModel := os.Getenv("MODEL")
	origBase := os.Getenv("BASE_URL")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("MODEL", origModel)
		os.Setenv("BASE_URL", origBase)
	}()

	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("MODEL")
	os.Unsetenv("BASE_URL")

	client := NewClient()
	if client.Model != "gpt-4o" {
		t.Errorf("default model = %q, want gpt-4o", client.Model)
	}
	if client.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("default baseURL = %q", client.BaseURL)
	}
}

func TestNewClientWith(t *testing.T) {
	client := NewClientWith("key", "model", "http://base")
	if client.APIKey != "key" {
		t.Errorf("APIKey = %q", client.APIKey)
	}
	if client.Model != "model" {
		t.Errorf("Model = %q", client.Model)
	}
	if client.BaseURL != "http://base" {
		t.Errorf("BaseURL = %q", client.BaseURL)
	}
	if client.HTTP == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header")
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "Hello from mock!"}},
			},
			"usage": map[string]int{"total_tokens": 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("test-key", "test-model", server.URL)
	result, err := client.ChatCompletion([]ChatMessage{
		{Role: "user", Content: "Hello"},
	}, 0.7, 100)

	if err != nil {
		t.Fatalf("ChatCompletion error: %v", err)
	}
	if result != "Hello from mock!" {
		t.Errorf("result = %q, want 'Hello from mock!'", result)
	}
}

func TestChatCompletionNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("test-key", "m", server.URL)
	_, err := client.ChatCompletion([]ChatMessage{{Role: "user", Content: "hi"}}, 0.7, 10)
	if err == nil {
		t.Error("expected error for no choices")
	}
}

func TestChatCompletionAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":{"message":"Invalid API key"}}`))
	}))
	defer server.Close()

	client := NewClientWith("bad-key", "m", server.URL)
	_, err := client.ChatCompletion([]ChatMessage{{Role: "user", Content: "hi"}}, 0.7, 10)
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestChatCompletionAPIErrorInBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]string{"message": "rate limited"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "m", server.URL)
	_, err := client.ChatCompletion([]ChatMessage{{Role: "user", Content: "hi"}}, 0.7, 10)
	if err == nil {
		t.Error("expected error for API error in body")
	}
}

func TestChatCompletionParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClientWith("key", "m", server.URL)
	_, err := client.ChatCompletion([]ChatMessage{{Role: "user", Content: "hi"}}, 0.7, 10)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{0.1, 0.2, 0.3}},
				{"index": 1, "embedding": []float64{0.4, 0.5, 0.6}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("test-key", "test-model", server.URL)
	embeddings, err := client.Embed([]string{"hello", "world"}, "test-embed")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
	if len(embeddings) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(embeddings))
	}
	if len(embeddings[0]) != 3 {
		t.Errorf("expected dim 3, got %d", len(embeddings[0]))
	}
}

func TestEmbedDefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "text-embedding-3-small" {
			t.Errorf("default model = %q, want text-embedding-3-small", req.Model)
		}
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{1.0}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "m", server.URL)
	_, err := client.Embed([]string{"test"}, "")
	if err != nil {
		t.Fatalf("Embed error: %v", err)
	}
}

func TestEmbedAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]string{"message": "quota exceeded"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "m", server.URL)
	_, err := client.Embed([]string{"test"}, "model")
	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestEmbedParseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClientWith("key", "m", server.URL)
	_, err := client.Embed([]string{"test"}, "model")
	if err == nil {
		t.Error("expected error for parse error")
	}
}

func TestPostNoAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("should not set Authorization when no key")
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("", "m", server.URL)
	_, err := client.ChatCompletion([]ChatMessage{{Role: "user", Content: "hi"}}, 0.7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildSearchText(t *testing.T) {
	text := BuildSearchText("parseFile", "Parses a file", "parseFile(path string)", "")
	if text == "" {
		t.Error("expected non-empty search text")
	}
	if !strings.Contains(text, "parseFile") {
		t.Error("should contain name")
	}
}

func TestBuildSearchTextAllEmpty(t *testing.T) {
	text := BuildSearchText("", "", "", "")
	if text != "" {
		t.Errorf("expected empty search text for all empty, got %q", text)
	}
}

func TestBuildSearchTextCodeTruncation(t *testing.T) {
	longCode := strings.Repeat("x", 3000)
	text := BuildSearchText("name", "", "", longCode)
	if len(text) > 2100 {
		t.Errorf("text too long: %d, code should be truncated at 2000", len(text))
	}
}

func TestGetEnvOr(t *testing.T) {
	orig := os.Getenv("TEST_GETENVVAR_XYZ")
	defer os.Setenv("TEST_GETENVVAR_XYZ", orig)

	os.Unsetenv("TEST_GETENVVAR_XYZ")
	if got := getEnvOr("TEST_GETENVVAR_XYZ", "fallback"); got != "fallback" {
		t.Errorf("getEnvOr(unset) = %q, want fallback", got)
	}

	os.Setenv("TEST_GETENVVAR_XYZ", "present")
	if got := getEnvOr("TEST_GETENVVAR_XYZ", "fallback"); got != "present" {
		t.Errorf("getEnvOr(set) = %q, want present", got)
	}
}
