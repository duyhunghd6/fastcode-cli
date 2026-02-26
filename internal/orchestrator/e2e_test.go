package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
)

// TestE2EFullPipeline tests the complete pipeline: create repo → index → query → answer
func TestE2EFullPipeline(t *testing.T) {
	// Create a test repository with multiple languages
	repoDir, err := os.MkdirTemp("", "fastcode-e2e-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	// Go file with functions and struct
	goContent := `package main

import "fmt"

// Config holds application configuration.
type Config struct {
	Port     int
	Host     string
	LogLevel string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:     8080,
		Host:     "localhost",
		LogLevel: "info",
	}
}

// LoadConfig reads configuration from environment.
func LoadConfig() Config {
	cfg := DefaultConfig()
	fmt.Println("loading config")
	return cfg
}

func main() {
	cfg := LoadConfig()
	fmt.Printf("Starting on %s:%d\n", cfg.Host, cfg.Port)
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Python file with class hierarchy
	pyContent := `"""Repository loader module"""

class BaseLoader:
    """Base class for file loading"""
    def __init__(self, path):
        self.path = path

    def load(self):
        raise NotImplementedError

class FileLoader(BaseLoader):
    """Loads individual files"""
    def load(self):
        with open(self.path) as f:
            return f.read()

def create_loader(path):
    """Factory function for creating loaders"""
    return FileLoader(path)
`
	if err := os.WriteFile(filepath.Join(repoDir, "loader.py"), []byte(pyContent), 0644); err != nil {
		t.Fatal(err)
	}

	// JavaScript file with imports and functions
	jsContent := `function processData(items) {
  return items.filter(item => item.active).map(item => item.name);
}

function formatOutput(data) {
  return JSON.stringify(data, null, 2);
}

class DataProcessor {
  constructor(config) {
    this.config = config;
  }

  process(input) {
    const data = processData(input);
    return formatOutput(data);
  }
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "processor.js"), []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup cache
	cacheDir, err := os.MkdirTemp("", "fastcode-e2e-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	// Step 1: Create engine
	cfg := Config{
		CacheDir:     cacheDir,
		BatchSize:    32,
		NoEmbeddings: true, // Offline test
	}
	engine := NewEngine(cfg)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	// Step 2: Index the repository
	indexResult, err := engine.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	// Verify indexing results
	if indexResult.TotalFiles < 3 {
		t.Errorf("expected at least 3 files, got %d", indexResult.TotalFiles)
	}
	if indexResult.TotalElements < 8 {
		t.Errorf("expected at least 8 elements, got %d", indexResult.TotalElements)
	}
	if indexResult.GraphStats == nil {
		t.Error("expected graph stats")
	}
	if indexResult.RepoName == "" {
		t.Error("expected repo name")
	}

	// Step 3: Query the codebase (direct search, no API key)
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	queryResult, err := engine.Query("how does the configuration work?")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if queryResult.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if queryResult.StopReason != "direct_search" {
		t.Errorf("stop_reason = %q, want direct_search", queryResult.StopReason)
	}
	if queryResult.Elements == 0 {
		t.Error("expected at least 1 element in results")
	}

	// Step 4: Test cache (re-index should be cached)
	engine2 := NewEngine(cfg)
	indexResult2, err := engine2.Index(repoDir, false)
	if err != nil {
		t.Fatalf("Cached Index failed: %v", err)
	}
	if !indexResult2.Cached {
		t.Error("second index should use cache")
	}
	if indexResult2.TotalElements != indexResult.TotalElements {
		t.Errorf("cached elements = %d, want %d", indexResult2.TotalElements, indexResult.TotalElements)
	}

	// Step 5: Query after cache rebuild
	queryResult2, err := engine2.Query("what classes exist in the codebase?")
	if err != nil {
		t.Fatalf("Query after cache failed: %v", err)
	}
	if queryResult2.Answer == "" {
		t.Error("expected non-empty answer after cache rebuild")
	}

	// Step 6: Force reindex
	indexResult3, err := engine2.Index(repoDir, true)
	if err != nil {
		t.Fatalf("Force reindex failed: %v", err)
	}
	if indexResult3.Cached {
		t.Error("force reindex should not be cached")
	}
}

// TestE2EEmptyRepository tests indexing an empty repository
func TestE2EEmptyRepository(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "fastcode-e2e-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repoDir)

	cacheDir, err := os.MkdirTemp("", "fastcode-e2e-cache-*")
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

	result, err := engine.Index(repoDir, false)
	if err != nil {
		t.Fatalf("Index empty repo: %v", err)
	}
	if result.TotalFiles != 0 {
		t.Errorf("expected 0 files for empty repo, got %d", result.TotalFiles)
	}
}
