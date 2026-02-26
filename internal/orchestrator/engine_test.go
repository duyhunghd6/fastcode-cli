package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CacheDir == "" {
		t.Error("CacheDir should not be empty")
	}
	if cfg.EmbeddingModel == "" {
		t.Error("EmbeddingModel should not be empty")
	}
	if cfg.BatchSize <= 0 {
		t.Error("BatchSize should be > 0")
	}
}

func TestEngineInit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fastcode-engine-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfig()
	cfg.CacheDir = filepath.Join(tempDir, "cache")
	cfg.NoEmbeddings = true

	engine := NewEngine(cfg)
	if engine == nil {
		t.Fatalf("Engine was nil")
	}

	if engine.cacheDir != cfg.CacheDir {
		t.Errorf("Expected cache dir %s, got %s", cfg.CacheDir, engine.cacheDir)
	}
	if engine.client == nil {
		t.Errorf("Expected client to be initialized")
	}
	if engine.cache == nil {
		t.Errorf("Expected cache to be initialized")
	}
	if engine.embedder != nil {
		t.Error("embedder should be nil when NoEmbeddings=true")
	}
}

func TestEngineIndex(t *testing.T) {
	// Create temp repo
	repoDir, err := os.MkdirTemp("", "fastcode-repo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	// Write a Go file
	goContent := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}

func helper() string {
	return "help"
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a Python file
	pyContent := `def greet(name):
    return f"Hello, {name}"
`
	if err := os.WriteFile(filepath.Join(repoDir, "app.py"), []byte(pyContent), 0644); err != nil {
		t.Fatal(err)
	}

	cacheDir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	cfg := Config{
		CacheDir:       cacheDir,
		EmbeddingModel: "test",
		BatchSize:      32,
		NoEmbeddings:   true,
	}
	engine := NewEngine(cfg)

	// Index
	result, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index: %v", err)
	}

	if result.TotalFiles < 2 {
		t.Errorf("TotalFiles = %d, want >= 2", result.TotalFiles)
	}
	if result.TotalElements < 4 {
		t.Errorf("TotalElements = %d, want >= 4", result.TotalElements)
	}
	if result.Cached {
		t.Error("first index should not be cached")
	}
	if result.GraphStats == nil {
		t.Error("GraphStats should not be nil")
	}
}

func TestEngineIndexCached(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "fastcode-repo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cacheDir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	cfg := Config{
		CacheDir:     cacheDir,
		BatchSize:    32,
		NoEmbeddings: true,
	}
	engine := NewEngine(cfg)

	// First index
	_, err = engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("First Index: %v", err)
	}

	// Second index (should use cache)
	engine2 := NewEngine(cfg)
	result, err := engine2.Index(repoDir, false)
	if err != nil {
		t.Fatalf("Second Index: %v", err)
	}
	if !result.Cached {
		t.Error("second index should be cached")
	}
}

func TestEngineQueryWithoutIndex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fastcode-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := Config{
		CacheDir:     filepath.Join(tempDir, "cache"),
		NoEmbeddings: true,
	}
	engine := NewEngine(cfg)

	_, err = engine.Query("test question")
	if err == nil {
		t.Error("expected error when querying without index")
	}
}

func TestEngineQueryDirect(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "fastcode-repo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	goContent := `package main

import "fmt"

// LoadConfig reads configuration
func LoadConfig(path string) error {
	fmt.Println("loading config from", path)
	return nil
}

func main() {
	LoadConfig("config.yaml")
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	cacheDir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	// Use no API key to trigger direct search
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cfg := Config{
		CacheDir:     cacheDir,
		BatchSize:    32,
		NoEmbeddings: true,
	}
	engine := NewEngine(cfg)

	_, err = engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index: %v", err)
	}

	result, err := engine.Query("how does config loading work?")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	if result.Answer == "" {
		t.Error("answer should not be empty")
	}
	if result.StopReason != "direct_search" {
		t.Errorf("StopReason = %q, want direct_search", result.StopReason)
	}
}

func TestSimpleAnswerEmpty(t *testing.T) {
	sa := &simpleAnswer{}
	result := sa.String()
	if result == "" {
		t.Error("empty simpleAnswer should still return text")
	}
}

func TestSimpleAnswerWithResults(t *testing.T) {
	sa := &simpleAnswer{}
	sa.addResult(&types.CodeElement{
		Type: "function", Name: "handleAuth",
		RelativePath: "auth.go", StartLine: 10, EndLine: 20,
		Signature: "func handleAuth()",
	})
	result := sa.String()
	if result == "" {
		t.Error("simpleAnswer with results should return text")
	}
}

func TestEngineIndexInvalidPath(t *testing.T) {
	cacheDir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	cfg := Config{CacheDir: cacheDir, NoEmbeddings: true}
	engine := NewEngine(cfg)

	_, err = engine.Index("/nonexistent/path/that/does/not/exist", false)
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
