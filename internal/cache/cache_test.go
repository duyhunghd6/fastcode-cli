package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestNewIndexCache(t *testing.T) {
	c := NewIndexCache("/tmp/test-cache")
	if c == nil {
		t.Fatal("NewIndexCache returned nil")
	}
	if c.CacheDir != "/tmp/test-cache" {
		t.Errorf("CacheDir = %q", c.CacheDir)
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	c := NewIndexCache(dir)

	data := &CachedIndex{
		RepoName: "test-repo",
		Elements: []types.CodeElement{
			{ID: "e1", Name: "foo", Type: "function", Language: "go"},
			{ID: "e2", Name: "bar", Type: "class", Language: "python"},
		},
		Vectors: map[string][]float32{
			"e1": {0.1, 0.2, 0.3},
			"e2": {0.4, 0.5, 0.6},
		},
	}

	if err := c.Save("test-repo", data); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !c.Exists("test-repo") {
		t.Error("Exists() = false after save")
	}

	loaded, err := c.Load("test-repo")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.RepoName != "test-repo" {
		t.Errorf("RepoName = %s, want test-repo", loaded.RepoName)
	}
	if len(loaded.Elements) != 2 {
		t.Errorf("elements = %d, want 2", len(loaded.Elements))
	}
	if loaded.Elements[0].Name != "foo" {
		t.Errorf("first element = %s, want foo", loaded.Elements[0].Name)
	}
	if len(loaded.Vectors) != 2 {
		t.Errorf("vectors = %d, want 2", len(loaded.Vectors))
	}

	if err := c.Delete("test-repo"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if c.Exists("test-repo") {
		t.Error("Exists() = true after delete")
	}
}

func TestCacheLoadNotExists(t *testing.T) {
	c := NewIndexCache("/tmp/nonexistent-cache-dir")
	_, err := c.Load("nonexistent")
	if err == nil {
		t.Error("expected error loading nonexistent cache")
	}
}

func TestCacheExistsNotExists(t *testing.T) {
	c := NewIndexCache("/tmp/nonexistent-cache-dir-xyz")
	if c.Exists("nonexistent") {
		t.Error("Exists should return false for nonexistent")
	}
}

func TestCacheDeleteNotExists(t *testing.T) {
	c := NewIndexCache("/tmp/nonexistent-cache-dir-xyz")
	err := c.Delete("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent cache")
	}
}

func TestCachePath(t *testing.T) {
	c := NewIndexCache("/tmp/cache")
	path := c.cachePath("my-repo")
	expected := filepath.Join("/tmp/cache", "my-repo.gob")
	if path != expected {
		t.Errorf("cachePath = %q, want %q", path, expected)
	}
}

func TestCacheSaveCreatesDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	subdir := filepath.Join(dir, "sub", "deep")
	c := NewIndexCache(subdir)

	data := &CachedIndex{
		RepoName: "test",
		Elements: nil,
		Vectors:  nil,
	}

	if err := c.Save("test", data); err != nil {
		t.Fatalf("Save to deep dir: %v", err)
	}

	if !c.Exists("test") {
		t.Error("should exist after save")
	}
}

func TestCacheSaveEmptyData(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-cache-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	c := NewIndexCache(dir)

	data := &CachedIndex{
		RepoName: "empty-repo",
	}

	if err := c.Save("empty-repo", data); err != nil {
		t.Fatalf("Save empty: %v", err)
	}

	loaded, err := c.Load("empty-repo")
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if loaded.RepoName != "empty-repo" {
		t.Errorf("RepoName = %q", loaded.RepoName)
	}
}
