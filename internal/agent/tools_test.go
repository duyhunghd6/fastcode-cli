package agent

import (
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestAvailableTools(t *testing.T) {
	tools := AvailableTools()
	if len(tools) == 0 {
		t.Fatal("expected available tools")
	}
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
	for _, expected := range []string{"search_codebase", "browse_file", "skim_file", "list_directory"} {
		if !names[expected] {
			t.Errorf("missing expected tool: %s", expected)
		}
	}
}

func TestNewToolExecutor(t *testing.T) {
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	elements := []types.CodeElement{
		{ID: "e1", Name: "foo", Type: "function"},
	}
	te := NewToolExecutor(hr, nil, elements)
	if te == nil {
		t.Fatal("NewToolExecutor returned nil")
	}
	if len(te.elements) != 1 {
		t.Errorf("elements map size = %d, want 1", len(te.elements))
	}
}

func TestToolExecutorSearchCode(t *testing.T) {
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "handleAuth", Type: "function", Code: "func handleAuth() { authenticate user }"},
		{ID: "e2", Name: "loadDB", Type: "function", Code: "func loadDB() { connect database }"},
	}
	_ = hr.IndexElements(elements, nil)

	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("search_code", "authenticate user")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.ToolName != "search_codebase" {
		t.Errorf("ToolName = %s, want search_codebase", result.ToolName)
	}
}

func TestToolExecutorBrowseFile(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "internal/main.go", Code: "package main\nfunc main() {}"},
	}

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("browse_file", "internal/main.go")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(result.Elements))
	}
	if result.Text == "" {
		t.Error("browse_file should populate Text with code")
	}
}

func TestToolExecutorBrowseFileNotFound(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "internal/main.go"},
	}
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("browse_file", "nonexistent/file.go")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.Elements) != 0 {
		t.Errorf("expected 0 elements for not found, got %d", len(result.Elements))
	}
	if result.Text == "" {
		t.Error("browse_file not found should set text message")
	}
}

func TestToolExecutorSkimFile(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "function", Name: "handleAuth", RelativePath: "auth.go", Code: "func handleAuth() {}"},
		{ID: "f2", Type: "class", Name: "AuthService", RelativePath: "auth.go", Code: "type AuthService struct{}"},
		{ID: "f3", Type: "file", RelativePath: "auth.go", Code: "package auth"},
	}

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("skim_file", "auth.go")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.ToolName != "skim_file" {
		t.Errorf("ToolName = %s, want skim_file", result.ToolName)
	}
	// Should find function and class, not file
	if len(result.Elements) != 2 {
		t.Errorf("expected 2 elements (function + class), got %d", len(result.Elements))
	}
	// Skim should have empty code
	for _, elem := range result.Elements {
		if elem.Code != "" {
			t.Errorf("skim should omit code, but element %q has code", elem.Name)
		}
	}
}

func TestToolExecutorSkimFileNotFound(t *testing.T) {
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	result, err := te.Execute("skim_file", "nonexistent.go")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Text == "" {
		t.Error("skim_file not found should set text message")
	}
}

func TestToolExecutorListFiles(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "internal/parser/go_parser.go"},
		{ID: "f2", Type: "file", RelativePath: "internal/parser/py_parser.go"},
		{ID: "f3", Type: "file", RelativePath: "internal/graph/graph.go"},
	}

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("list_files", "parser")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.Elements) != 2 {
		t.Errorf("expected 2 parser files, got %d", len(result.Elements))
	}
}

func TestToolExecutorListFilesNoMatch(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "main.go"},
	}

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	result, err := te.Execute("list_files", "nonexistent")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.Elements) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Elements))
	}
}

func TestToolExecutorUnknown(t *testing.T) {
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	_, err := te.Execute("nonexistent", "arg")
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestProcessQueryEmpty(t *testing.T) {
	pq := ProcessQuery("")
	if pq == nil {
		t.Fatal("ProcessQuery returned nil for empty")
	}
	if pq.Original != "" {
		t.Errorf("Original should be empty, got %q", pq.Original)
	}
}
