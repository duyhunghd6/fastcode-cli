package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// TestSaveMkdirError tests Save when cache dir parent is not writable
func TestSaveMkdirError(t *testing.T) {
	c := NewIndexCache("/dev/null/impossible/path")
	data := &CachedIndex{RepoName: "test"}
	err := c.Save("test", data)
	if err == nil {
		t.Error("expected error when creating cache dir fails")
	}
}

// TestSaveCreateFileError tests Save when file cannot be created
func TestSaveCreateFileError(t *testing.T) {
	// Create cache dir as a file to make file creation fail
	tmpDir, _ := os.MkdirTemp("", "cache-file-err-*")
	defer os.RemoveAll(tmpDir)

	// Create a file where the gob file should be
	cachePath := filepath.Join(tmpDir, "test.gob")
	os.MkdirAll(cachePath, 0755) // Make it a directory so os.Create fails

	c := NewIndexCache(tmpDir)
	data := &CachedIndex{RepoName: "test"}
	err := c.Save("test", data)
	if err == nil {
		t.Error("expected error when creating cache file fails")
	}
}

// TestSaveAndLoadRoundTrip tests full save-load cycle with complex data
func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cache-roundtrip-*")
	defer os.RemoveAll(tmpDir)

	c := NewIndexCache(tmpDir)

	data := &CachedIndex{
		RepoName: "my-project",
		Elements: []types.CodeElement{
			{ID: "e1", Name: "main", Type: "function", Language: "go", Code: "func main() {}"},
			{ID: "e2", Name: "Server", Type: "class", Language: "go", Code: "type Server struct{}"},
		},
		Vectors: map[string][]float32{
			"e1": {0.1, 0.2, 0.3},
			"e2": {0.4, 0.5, 0.6},
		},
	}

	err := c.Save("my-project", data)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := c.Load("my-project")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.RepoName != "my-project" {
		t.Errorf("RepoName = %q", loaded.RepoName)
	}
	if len(loaded.Elements) != 2 {
		t.Errorf("Elements = %d, want 2", len(loaded.Elements))
	}
	if len(loaded.Vectors) != 2 {
		t.Errorf("Vectors = %d, want 2", len(loaded.Vectors))
	}
}

// TestLoadCorruptFile tests Load with a corrupted file
func TestLoadCorruptFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cache-corrupt-*")
	defer os.RemoveAll(tmpDir)

	// Write corrupt data
	os.WriteFile(filepath.Join(tmpDir, "corrupt.gob"), []byte("not valid gob data"), 0644)

	c := NewIndexCache(tmpDir)
	_, err := c.Load("corrupt")
	if err == nil {
		t.Error("expected error loading corrupt cache file")
	}
}

// TestLoadNonexistent tests Load when file doesn't exist
func TestLoadNonexistent(t *testing.T) {
	c := NewIndexCache("/tmp/nonexistent-cache-dir")
	_, err := c.Load("nonexistent")
	if err == nil {
		t.Error("expected error loading nonexistent cache")
	}
}

// TestDeleteNonexistent tests Delete when file doesn't exist
func TestDeleteNonexistent(t *testing.T) {
	c := NewIndexCache("/tmp/nonexistent-cache-dir")
	err := c.Delete("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent cache")
	}
}

// TestDeleteExisting tests Delete successfully
func TestDeleteExisting(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cache-delete-*")
	defer os.RemoveAll(tmpDir)

	c := NewIndexCache(tmpDir)
	c.Save("test", &CachedIndex{RepoName: "test"})

	if !c.Exists("test") {
		t.Fatal("cache should exist after save")
	}

	err := c.Delete("test")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if c.Exists("test") {
		t.Error("cache should not exist after delete")
	}
}

// TestCachePath tests the internal path generation
func TestCachePathGeneration(t *testing.T) {
	c := NewIndexCache("/tmp/test-cache")
	path := c.cachePath("my-repo")
	expected := filepath.Join("/tmp/test-cache", "my-repo.gob")
	if path != expected {
		t.Errorf("cachePath = %q, want %q", path, expected)
	}
}

// TestSaveEmptyData tests saving empty CachedIndex
func TestSaveEmptyData(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "cache-empty-*")
	defer os.RemoveAll(tmpDir)

	c := NewIndexCache(tmpDir)
	err := c.Save("empty", &CachedIndex{})
	if err != nil {
		t.Fatalf("Save empty: %v", err)
	}

	loaded, err := c.Load("empty")
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if loaded.RepoName != "" {
		t.Error("empty data should have empty repo name")
	}
}
