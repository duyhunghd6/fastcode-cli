package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEmbedder(t *testing.T) {
	client := NewClientWith("key", "model", "http://localhost")
	e := NewEmbedder(client, "", 0)
	if e == nil {
		t.Fatal("NewEmbedder returned nil")
	}
	if e.model != "text-embedding-3-small" {
		t.Errorf("default model = %q, want text-embedding-3-small", e.model)
	}
	if e.batchSize != 32 {
		t.Errorf("default batchSize = %d, want 32", e.batchSize)
	}
}

func TestNewEmbedderCustom(t *testing.T) {
	client := NewClientWith("key", "model", "http://localhost")
	e := NewEmbedder(client, "my-model", 16)
	if e.model != "my-model" {
		t.Errorf("model = %q, want my-model", e.model)
	}
	if e.batchSize != 16 {
		t.Errorf("batchSize = %d, want 16", e.batchSize)
	}
}

func TestEmbedTextsEmpty(t *testing.T) {
	client := NewClientWith("key", "model", "http://localhost")
	e := NewEmbedder(client, "", 32)

	result, err := e.EmbedTexts(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestEmbedTextsSingle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{0.1, 0.2}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 32)

	result, err := e.EmbedTexts([]string{"hello"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Errorf("expected dim 2, got %d", len(result[0]))
	}
}

func TestEmbedTextsBatching(t *testing.T) {
	batchCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batchCount++
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)

		data := make([]map[string]any, len(req.Input))
		for i := range req.Input {
			data[i] = map[string]any{
				"index":     i,
				"embedding": []float64{float64(i) * 0.1},
			}
		}
		resp := map[string]any{"data": data}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 2) // batchSize=2

	texts := []string{"a", "b", "c", "d", "e"} // 5 texts, 3 batches
	result, err := e.EmbedTexts(texts)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 results, got %d", len(result))
	}
	if batchCount != 3 {
		t.Errorf("expected 3 batches, got %d", batchCount)
	}
}

func TestEmbedTextsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 32)

	_, err := e.EmbedTexts([]string{"hello"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestEmbedTextSingle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{0.5, 0.6, 0.7}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 32)

	vec, err := e.EmbedText("hello")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("expected dim 3, got %d", len(vec))
	}
}

func TestEmbedTextError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"error"}}`))
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 32)

	_, err := e.EmbedText("hello")
	if err == nil {
		t.Error("expected error")
	}
}

func TestEmbedTextNilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWith("key", "model", server.URL)
	e := NewEmbedder(client, "model", 32)

	_, err := e.EmbedText("hello")
	if err == nil {
		t.Error("expected error for nil embedding result")
	}
}
