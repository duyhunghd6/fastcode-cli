package orchestrator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/cache"
	"github.com/duyhunghd6/fastcode-cli/internal/graph"
	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// TestIndexCacheLoadError tests the cache load error fallback (L100 in engine.go)
func TestIndexCacheLoadError(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "fastcode-cache-err-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-cache-err-cache-*")
	defer os.RemoveAll(cacheDir)

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)

	// First: index normally to create cache
	_, err = engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("first Index: %v", err)
	}

	// Corrupt the cache file
	repoName := filepath.Base(repoDir)
	cachePath := filepath.Join(cacheDir, repoName+".gob")
	os.WriteFile(cachePath, []byte("corrupted data"), 0644)

	// Second: index without force — should detect corrupt cache, fall back to reindex
	engine2 := NewEngine(cfg)
	result, err := engine2.Index(repoDir, false)
	if err != nil {
		t.Fatalf("Index with corrupt cache: %v", err)
	}
	// Should have re-indexed (not from cache, despite cache existing)
	if result.Cached {
		t.Error("expected non-cached result when cache is corrupt")
	}
	if result.TotalElements == 0 {
		t.Error("expected elements from reindexing")
	}
}

// TestRebuildFromCacheWithVectors tests rebuildFromCache with vector data
func TestRebuildFromCacheWithVectors(t *testing.T) {
	engine := &Engine{}

	cached := &cache.CachedIndex{
		RepoName: "test-repo",
		Elements: []types.CodeElement{
			{ID: "e1", Name: "foo", Type: "function", Code: "func foo() {}"},
			{ID: "e2", Name: "bar", Type: "function", Code: "func bar() {}"},
		},
		Vectors: map[string][]float32{
			"e1": {0.1, 0.2, 0.3},
			"e2": {0.4, 0.5, 0.6},
		},
	}

	engine.rebuildFromCache(cached)

	if engine.graphs == nil {
		t.Error("graphs should be initialized")
	}
	if engine.hybrid == nil {
		t.Error("hybrid should be initialized")
	}

	// Verify vector store has the vectors
	results := engine.hybrid.Search("foo", []float32{0.1, 0.2, 0.3}, 5)
	if len(results) == 0 {
		t.Error("expected search results after cache rebuild with vectors")
	}
}

// TestRebuildFromCacheEmptyVectors tests rebuildFromCache with no vectors
func TestRebuildFromCacheEmptyVectors(t *testing.T) {
	engine := &Engine{}

	cached := &cache.CachedIndex{
		RepoName: "test-repo",
		Elements: []types.CodeElement{
			{ID: "e1", Name: "main", Type: "function", Code: "func main() {}"},
		},
		Vectors: map[string][]float32{}, // empty vectors
	}

	engine.rebuildFromCache(cached)

	if engine.hybrid == nil {
		t.Error("hybrid should be initialized even with empty vectors")
	}
}

// TestQueryDirectWithEmbedderSuccess tests queryDirect where embedder successfully embeds
func TestQueryDirectWithEmbedderSuccess(t *testing.T) {
	// Mock embedding server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{0.1, 0.2, 0.3}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	repoDir, _ := os.MkdirTemp("", "fastcode-qd-emb-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qd-emb-cache-*")
	defer os.RemoveAll(cacheDir)

	// Set API key to enable embedder, but then clear it before query so we go direct
	origKey := os.Getenv("OPENAI_API_KEY")
	origBase := os.Getenv("BASE_URL")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("BASE_URL", mockServer.URL)
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("BASE_URL", origBase)
	}()

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: false}
	engine := NewEngine(cfg)

	// Index with embeddings enabled (embedder != nil)
	_, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index: %v", err)
	}

	// Now clear API key on client to force direct search path
	engine.client.APIKey = ""

	result, err := engine.Query("main function")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if result.StopReason != "direct_search" {
		t.Errorf("StopReason = %q, want direct_search", result.StopReason)
	}
}

// TestQueryWithAgentRetrievalError tests queryWithAgent when retrieval fails
func TestQueryWithAgentRetrievalError(t *testing.T) {
	// Mock LLM that always returns 500
	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"internal error"}}`))
	}))
	defer mockLLM.Close()

	repoDir, _ := os.MkdirTemp("", "fastcode-qaerr-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qaerr-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	origBase := os.Getenv("BASE_URL")
	origModel := os.Getenv("MODEL")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("BASE_URL", mockLLM.URL)
	os.Setenv("MODEL", "test-model")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("BASE_URL", origBase)
		os.Setenv("MODEL", origModel)
	}()

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)

	_, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index: %v", err)
	}

	// Query should fail because LLM returns 500
	_, err = engine.Query("test query")
	if err == nil {
		t.Error("expected error from failed agent retrieval")
	}
}

// TestQueryWithAgentAnswerError tests queryWithAgent when answer generation fails
func TestQueryWithAgentAnswerError(t *testing.T) {
	callCount := 0
	mockLLM := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 1 {
			// Agent retrieval succeeds with high confidence
			resp := map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{
						"role":    "assistant",
						"content": `{"confidence": 95, "reasoning": "done", "tool_calls": []}`,
					}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			// Answer generation fails
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"message":"answer gen failed"}}`))
		}
	}))
	defer mockLLM.Close()

	repoDir, _ := os.MkdirTemp("", "fastcode-qaans-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qaans-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	origBase := os.Getenv("BASE_URL")
	origModel := os.Getenv("MODEL")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("BASE_URL", mockLLM.URL)
	os.Setenv("MODEL", "test-model")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("BASE_URL", origBase)
		os.Setenv("MODEL", origModel)
	}()

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)

	_, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index: %v", err)
	}

	_, err = engine.Query("test query")
	if err == nil {
		t.Error("expected error from failed answer generation")
	}
}

// TestIndexEmbeddingError tests the embedding error log path (L122 in engine.go)
func TestIndexEmbeddingError(t *testing.T) {
	// Mock server that fails on embeddings
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"embed failed"}}`))
	}))
	defer mockServer.Close()

	repoDir, _ := os.MkdirTemp("", "fastcode-emb-err-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-emb-err-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	origBase := os.Getenv("BASE_URL")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("BASE_URL", mockServer.URL)
	defer func() {
		os.Setenv("OPENAI_API_KEY", origKey)
		os.Setenv("BASE_URL", origBase)
	}()

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: false}
	engine := NewEngine(cfg)

	// Should succeed even though embeddings fail (falls back to BM25)
	result, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index with embedding error: %v", err)
	}
	if result.TotalElements == 0 {
		t.Error("should still have elements even if embeddings fail")
	}
}

// TestIndexCacheSaveError tests when cache save fails (e.g., read-only dir)
func TestIndexCacheSaveError(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-save-err-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	// Use non-writable cache dir
	cfg := Config{CacheDir: "/dev/null/impossible", BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)

	// Should NOT error — cache save failure is logged, not returned
	result, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index with cache save error: %v", err)
	}
	if result.TotalElements == 0 {
		t.Error("should have elements even if cache save fails")
	}
}

// TestGraphsStatsContent verifies graph stats in index result
func TestGraphsStatsContent(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-stats-*")
	defer os.RemoveAll(repoDir)

	goContent := `package main

import "fmt"

type Handler struct{}

func (h *Handler) Handle() { fmt.Println("handled") }

func main() {
	h := &Handler{}
	h.Handle()
}
`
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-stats-cache-*")
	defer os.RemoveAll(cacheDir)

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)

	result, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatal(err)
	}

	stats := result.GraphStats
	if stats == nil {
		t.Fatal("GraphStats should not be nil")
	}
	// Stats should include graph type names
	t.Logf("GraphStats: %v", stats)
}

// TestEngineDirectVsAgentPath tests that direct/agent path is selected by API key
func TestEngineDirectVsAgentPath(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-path-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-path-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cfg := Config{CacheDir: cacheDir, BatchSize: 32, NoEmbeddings: true}
	engine := NewEngine(cfg)
	engine.Index(repoDir, true)

	// No API key → direct path
	result, err := engine.Query("test")
	if err != nil {
		t.Fatal(err)
	}
	if result.StopReason != "direct_search" {
		t.Errorf("no API key should use direct_search, got %q", result.StopReason)
	}
}

// TestSimpleAnswerMultipleResults tests simpleAnswer with multiple results
func TestSimpleAnswerMultipleResults(t *testing.T) {
	sa := &simpleAnswer{}
	sa.addResult(&types.CodeElement{
		Type: "function", Name: "foo", RelativePath: "a.go",
		StartLine: 1, EndLine: 5, Signature: "func foo()",
	})
	sa.addResult(&types.CodeElement{
		Type: "class", Name: "Bar", RelativePath: "b.go",
		StartLine: 10, EndLine: 30, Signature: "type Bar struct",
	})

	result := sa.String()
	if result == "" {
		t.Error("expected non-empty result")
	}
	if len(sa.lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(sa.lines))
	}
}

// TestCodeGraphsIntegration tests direct CodeGraphs usage from engine context
func TestCodeGraphsIntegration(t *testing.T) {
	g := graph.NewCodeGraphs()
	elements := []types.CodeElement{
		{ID: "e1", Name: "Server", Type: "class", Language: "go", RelativePath: "server.go"},
		{ID: "e2", Name: "Start", Type: "function", Language: "go", RelativePath: "server.go"},
	}
	g.BuildGraphs(elements)
	stats := g.Stats()
	if stats == nil {
		t.Error("stats should not be nil")
	}
}

// TestVectorStoreInRebuild tests vector store Get after rebuild
func TestVectorStoreInRebuild(t *testing.T) {
	vs := index.NewVectorStore()
	vs.Add("e1", []float32{0.1, 0.2})
	vs.Add("e2", []float32{0.3, 0.4})

	got := vs.Get("e1")
	if got == nil {
		t.Error("expected vector for e1")
	}
	got2 := vs.Get("nonexistent")
	if got2 != nil {
		t.Error("expected nil for nonexistent")
	}
}
