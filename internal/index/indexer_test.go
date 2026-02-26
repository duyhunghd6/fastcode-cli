package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/loader"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestNewIndexer(t *testing.T) {
	idx := NewIndexer("test-repo")
	if idx == nil {
		t.Fatal("NewIndexer returned nil")
	}
	if idx.repoName != "test-repo" {
		t.Errorf("repoName = %q, want test-repo", idx.repoName)
	}
}

func TestIndexRepository(t *testing.T) {
	// Create a temporary test repository
	dir, err := os.MkdirTemp("", "fastcode-indexer-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	// Write a Go file with functions, struct, and imports
	goContent := `package main

import "fmt"

// Server handles HTTP requests.
type Server struct {
	Port int
}

// Start starts the server.
func (s *Server) Start() error {
	fmt.Println("starting")
	return nil
}

func main() {
	s := &Server{Port: 8080}
	s.Start()
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Write a Python file
	pyContent := `"""Module documentation"""

class Calculator:
    """A simple calculator"""
    def add(self, a, b):
        return a + b

def main():
    c = Calculator()
    print(c.add(1, 2))
`
	if err := os.WriteFile(filepath.Join(dir, "calc.py"), []byte(pyContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := loader.DefaultConfig()
	repo, err := loader.LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	idx := NewIndexer("test-repo")
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		t.Fatalf("IndexRepository: %v", err)
	}

	if len(elements) == 0 {
		t.Fatal("expected elements from indexing")
	}

	// Check we have file, function, class elements
	elemTypes := make(map[string]int)
	for _, elem := range elements {
		elemTypes[elem.Type]++
	}

	if elemTypes["file"] < 2 {
		t.Errorf("expected at least 2 file elements, got %d", elemTypes["file"])
	}
	if elemTypes["function"] < 2 {
		t.Errorf("expected at least 2 function elements, got %d", elemTypes["function"])
	}

	// Verify elements have correct fields
	// Note: IndexRepository sets repoName to repo.Name (dir basename)
	for _, elem := range elements {
		if elem.ID == "" {
			t.Error("element has empty ID")
		}
		if elem.RepoName != repo.Name {
			t.Errorf("element RepoName = %q, want %q", elem.RepoName, repo.Name)
		}
	}
}

func TestIndexRepositoryWithDocElement(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-indexer-doc-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	pyContent := `"""This is a module docstring"""

def hello():
    pass
`
	if err := os.WriteFile(filepath.Join(dir, "module.py"), []byte(pyContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := loader.DefaultConfig()
	repo, err := loader.LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	idx := NewIndexer("test-repo")
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		t.Fatalf("IndexRepository: %v", err)
	}

	// Should have a documentation element
	foundDoc := false
	for _, elem := range elements {
		if elem.Type == "documentation" {
			foundDoc = true
		}
	}
	if !foundDoc {
		t.Error("expected documentation element for module with docstring")
	}
}

func TestExtractCodeBlock(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"

	tests := []struct {
		start, end int
		want       string
	}{
		{2, 4, "line2\nline3\nline4"},
		{1, 1, "line1"},
		{5, 5, "line5"},
		{0, 2, "line1\nline2"},  // start < 1 → clamped to 1
		{4, 10, "line4\nline5"}, // end > lines → clamped
		{3, 2, ""},              // start > end
	}

	for _, tt := range tests {
		got := extractCodeBlock(content, tt.start, tt.end)
		if got != tt.want {
			t.Errorf("extractCodeBlock(_, %d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestTruncateIndex(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello\n... (truncated)"},
		{"", 5, ""},
		{"ab", 2, "ab"},
		{"abc", 2, "ab\n... (truncated)"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestGenerateFileSummary(t *testing.T) {
	idx := NewIndexer("test-repo")

	// Test with all sections populated
	pr := &types.FileParseResult{
		Language:   "go",
		TotalLines: 100,
		Classes: []types.ClassInfo{
			{Name: "Server"},
			{Name: "Client"},
		},
		Functions: []types.FunctionInfo{
			{Name: "main", IsMethod: false},
			{Name: "Start", IsMethod: true, ClassName: "Server"},
			{Name: "helper", IsMethod: false},
		},
	}

	summary := idx.generateFileSummary(pr)
	if !strings.Contains(summary, "go") {
		t.Error("summary should contain language")
	}
	if !strings.Contains(summary, "Server") {
		t.Error("summary should contain class name")
	}
	if !strings.Contains(summary, "main") {
		t.Error("summary should contain function name")
	}
	// Methods should NOT appear in top-level functions list
	if strings.Contains(summary, "Start") {
		t.Error("summary should not list methods in Functions section")
	}
}

func TestGenerateFileSummaryEmpty(t *testing.T) {
	idx := NewIndexer("test-repo")
	pr := &types.FileParseResult{
		Language:   "go",
		TotalLines: 1,
	}

	summary := idx.generateFileSummary(pr)
	if summary == "" {
		t.Error("summary should not be empty even for minimal file")
	}
}

func TestGenID(t *testing.T) {
	idx := NewIndexer("test-repo")
	id1 := idx.genID("file", "a.go")
	id2 := idx.genID("file", "a.go")
	if id1 != id2 {
		t.Error("genID should be deterministic")
	}

	id3 := idx.genID("function", "main")
	if id1 == id3 {
		t.Error("different inputs should produce different IDs")
	}
}
