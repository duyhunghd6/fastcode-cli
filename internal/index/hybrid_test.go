package index

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestHybridRetrieverBM25Only(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "parseFile", Type: "function", Code: "func parseFile(path string) error { return nil }"},
		{ID: "e2", Name: "loadConfig", Type: "function", Code: "func loadConfig(config Config) { }"},
		{ID: "e3", Name: "buildGraph", Type: "function", Code: "func buildGraph(elements []Element) Graph { }"},
	}

	err := hr.IndexElements(elements, nil) // no embedder
	if err != nil {
		t.Fatalf("IndexElements: %v", err)
	}

	results := hr.Search("parsefile path string", nil, 3)
	if len(results) == 0 {
		t.Fatal("expected results for 'parsefile path string'")
	}
	if results[0].Element.ID != "e1" {
		t.Errorf("expected e1 first, got %s", results[0].Element.ID)
	}
}

func TestHybridRetrieverElementCount(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "foo"},
		{ID: "e2", Name: "bar"},
	}

	_ = hr.IndexElements(elements, nil)
	if got := hr.ElementCount(); got != 2 {
		t.Errorf("ElementCount() = %d, want 2", got)
	}
}

func TestNewHybridRetriever(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)
	if hr == nil {
		t.Fatal("NewHybridRetriever returned nil")
	}
	if hr.SemanticWeight != 0.6 {
		t.Errorf("SemanticWeight = %f, want 0.6", hr.SemanticWeight)
	}
	if hr.KeywordWeight != 0.4 {
		t.Errorf("KeywordWeight = %f, want 0.4", hr.KeywordWeight)
	}
}

func TestHybridSearchEmpty(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	results := hr.Search("test", nil, 5)
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty index, got %d", len(results))
	}
}

func TestHybridSearchWithVectors(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "authenticate", Type: "function", Code: "func authenticate() {}"},
		{ID: "e2", Name: "authorize", Type: "function", Code: "func authorize() {}"},
	}

	_ = hr.IndexElements(elements, nil)

	// Manually add vectors
	vs.Add("e1", []float32{1, 0, 0})
	vs.Add("e2", []float32{0, 1, 0})

	// Search with vector - should combine BM25 and vector scores
	results := hr.Search("authenticate", []float32{1, 0, 0}, 5)
	if len(results) == 0 {
		t.Fatal("expected results with vector search")
	}
	if results[0].Element.ID != "e1" {
		t.Errorf("expected e1 first with vector+BM25, got %s", results[0].Element.ID)
	}
}

func TestHybridIndexElementsWithEmbedder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input []string `json:"input"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		data := make([]map[string]any, len(req.Input))
		for i := range req.Input {
			data[i] = map[string]any{
				"index":     i,
				"embedding": []float64{float64(i) * 0.1, 0.5, 0.3},
			}
		}
		resp := map[string]any{"data": data}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	embedder := llm.NewEmbedder(client, "model", 32)

	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "foo", Type: "function", Code: "func foo() {}"},
	}

	err := hr.IndexElements(elements, embedder)
	if err != nil {
		t.Fatalf("IndexElements with embedder: %v", err)
	}

	// Check vector was stored
	if vs.Count() == 0 {
		t.Error("expected vectors to be stored")
	}
}

func TestHybridIndexElementsEmbedderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	embedder := llm.NewEmbedder(client, "model", 32)

	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "foo", Type: "function"},
	}

	err := hr.IndexElements(elements, embedder)
	if err == nil {
		t.Error("expected error from failed embedder")
	}
	// BM25 should still be indexed even if embedding fails
	if hr.ElementCount() != 1 {
		t.Errorf("ElementCount = %d, want 1 (BM25 should still work)", hr.ElementCount())
	}
}

func TestHybridSearchTopKGreaterThanResults(t *testing.T) {
	vs := NewVectorStore()
	bm := NewBM25(1.5, 0.75)
	hr := NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "hello", Type: "function", Code: "func hello() {}"},
	}
	_ = hr.IndexElements(elements, nil)

	results := hr.Search("hello", nil, 100)
	if len(results) != 1 {
		t.Errorf("expected 1 result when topK > available, got %d", len(results))
	}
}
