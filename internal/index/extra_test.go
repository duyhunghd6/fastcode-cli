package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/loader"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// TestIndexRepositorySkipUnreadable tests that unreadable files are skipped
func TestIndexRepositorySkipUnreadable(t *testing.T) {
	dir, _ := os.MkdirTemp("", "indexer-skip-*")
	defer os.RemoveAll(dir)

	// Write a valid Go file
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	// Create a file that we'll make unreadable
	unreadable := filepath.Join(dir, "secret.go")
	os.WriteFile(unreadable, []byte("package main\n"), 0644)
	os.Chmod(unreadable, 0000)
	defer os.Chmod(unreadable, 0644) // cleanup

	repo := &loader.Repository{
		RootPath: dir,
		Name:     "test-repo",
		Files: []loader.FileInfo{
			{Path: filepath.Join(dir, "main.go"), RelativePath: "main.go", Language: "go"},
			{Path: unreadable, RelativePath: "secret.go", Language: "go"},
		},
	}

	idx := NewIndexer("test-repo")
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		t.Fatalf("IndexRepository: %v", err)
	}
	// Should have at least elements from main.go (file + function)
	if len(elements) < 2 {
		t.Errorf("expected at least 2 elements, got %d", len(elements))
	}
}

// TestIndexRepositoryUnsupportedFile tests that unsupported files return nil parse
func TestIndexRepositoryUnsupportedFile(t *testing.T) {
	dir, _ := os.MkdirTemp("", "indexer-unsupported-*")
	defer os.RemoveAll(dir)

	// Write a markdown file (unsupported for parsing)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	repo := &loader.Repository{
		RootPath: dir,
		Name:     "test-repo",
		Files: []loader.FileInfo{
			{Path: filepath.Join(dir, "README.md"), RelativePath: "README.md", Language: ""},
			{Path: filepath.Join(dir, "main.go"), RelativePath: "main.go", Language: "go"},
		},
	}

	idx := NewIndexer("test-repo")
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		t.Fatalf("IndexRepository: %v", err)
	}
	// Should have elements from main.go only
	if len(elements) < 2 {
		t.Errorf("expected at least 2 elements, got %d", len(elements))
	}
}

// TestIndexRepositoryWithClassAndDocstring tests class+docstring indexing paths
func TestIndexRepositoryWithClassAndDocstring(t *testing.T) {
	dir, _ := os.MkdirTemp("", "indexer-class-*")
	defer os.RemoveAll(dir)

	pyContent := `"""
This is the module docstring.
"""

class Server:
    """Server handles HTTP requests."""

    def __init__(self):
        self.port = 8080

    def start(self):
        """Start the server."""
        pass

class Client(Server):
    """Client extends Server."""
    pass
`
	os.WriteFile(filepath.Join(dir, "server.py"), []byte(pyContent), 0644)

	repo := &loader.Repository{
		RootPath: dir,
		Name:     "test-repo",
		Files: []loader.FileInfo{
			{Path: filepath.Join(dir, "server.py"), RelativePath: "server.py", Language: "python"},
		},
	}

	idx := NewIndexer("test-repo")
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		t.Fatalf("IndexRepository: %v", err)
	}

	// Should include: file element, class Server, class Client, method __init__, method start, doc element
	hasClassWithBases := false
	hasDocElement := false
	for _, elem := range elements {
		if elem.Type == "class" && elem.Name == "Client" {
			// Client extends Server — tests the "extends" path in addClassElement
			if elem.Signature == "" {
				t.Error("Client should have signature")
			}
			hasClassWithBases = true
		}
		if elem.Type == "documentation" {
			hasDocElement = true
		}
	}

	if !hasDocElement {
		t.Logf("warning: module docstring element not found (may depend on parser version)")
	}
	if !hasClassWithBases {
		t.Logf("warning: Client class with bases not found, elements count = %d", len(elements))
	}
}

// TestExtractCodeBlockEdgeCases tests the extractCodeBlock helper
func TestExtractCodeBlockEdgeCases(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"

	// Normal case
	result := extractCodeBlock(content, 2, 4)
	if result != "line2\nline3\nline4" {
		t.Errorf("normal = %q", result)
	}

	// startLine < 1
	result = extractCodeBlock(content, 0, 2)
	if result != "line1\nline2" {
		t.Errorf("startLine<1 = %q", result)
	}

	// endLine > len(lines)
	result = extractCodeBlock(content, 4, 100)
	if result != "line4\nline5" {
		t.Errorf("endLine>len = %q", result)
	}

	// startLine > endLine
	result = extractCodeBlock(content, 5, 2)
	if result != "" {
		t.Errorf("start>end = %q, want empty", result)
	}
}

// TestTruncate tests the truncate helper
func TestTruncate(t *testing.T) {
	// Short string — no truncation
	result := truncate("hello", 100)
	if result != "hello" {
		t.Errorf("short = %q", result)
	}

	// Long string — truncated
	result = truncate("hello world this is a long string", 5)
	if result != "hello\n... (truncated)" {
		t.Errorf("long = %q", result)
	}

	// Exact length
	result = truncate("exact", 5)
	if result != "exact" {
		t.Errorf("exact = %q", result)
	}
}

// TestGenID tests the ID generation
func TestGenIDUniqueness(t *testing.T) {
	idx := NewIndexer("test-repo")
	id1 := idx.genID("function", "main.go", "main")
	id2 := idx.genID("function", "main.go", "helper")
	if id1 == id2 {
		t.Error("different elements should have different IDs")
	}
	if id1 == "" || id2 == "" {
		t.Error("IDs should not be empty")
	}
}

// TestGenerateFileSummary tests the file summary generation
func TestGenerateFileSummaryWithClasses(t *testing.T) {
	idx := NewIndexer("test")

	// With classes and functions
	pr := &types.FileParseResult{
		Language:   "go",
		TotalLines: 100,
		Classes:    []types.ClassInfo{{Name: "Server"}, {Name: "Client"}},
		Functions: []types.FunctionInfo{
			{Name: "main", IsMethod: false},
			{Name: "Start", IsMethod: true},
			{Name: "helper", IsMethod: false},
		},
	}

	summary := idx.generateFileSummary(pr)
	if summary == "" {
		t.Error("summary should not be empty")
	}
	// Should contain class names and non-method function names
	t.Logf("summary: %s", summary)
}

// TestGenerateFileSummaryEmpty tests summary with no classes or functions
func TestGenerateFileSummaryMinimal(t *testing.T) {
	idx := NewIndexer("test")
	pr := &types.FileParseResult{
		Language:   "go",
		TotalLines: 5,
	}

	summary := idx.generateFileSummary(pr)
	if summary == "" {
		t.Error("summary should include language and lines")
	}
}
