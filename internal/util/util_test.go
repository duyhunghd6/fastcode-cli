package util

import (
	"path/filepath"
	"testing"
)

func TestGetLanguageFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"component.tsx", "tsx"},
		{"main.rs", "rust"},
		{"Main.java", "java"},
		{"file.ts", "typescript"},
		{"file.jsx", "javascript"},
		{"file.c", "c"},
		{"file.h", "c"},
		{"file.cpp", "cpp"},
		{"file.cc", "cpp"},
		{"file.cxx", "cpp"},
		{"file.hpp", "cpp"},
		{"file.cs", "csharp"},
		{"file.rb", "ruby"},
		{"file.php", "php"},
		{"file.swift", "swift"},
		{"file.kt", "kotlin"},
		{"file.scala", "scala"},
		{"README.md", ""},
		{"Makefile", ""},
		{"styles.css", ""},
	}
	for _, tt := range tests {
		got := GetLanguageFromPath(tt.path)
		if got != tt.want {
			t.Errorf("GetLanguageFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestGetLanguageFromExtension(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "go"},
		{".py", "python"},
		{".GO", "go"}, // case insensitive
		{".Py", "python"},
		{".xyz", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := GetLanguageFromExtension(tt.ext)
		if got != tt.want {
			t.Errorf("GetLanguageFromExtension(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestIsSupportedFile(t *testing.T) {
	if !IsSupportedFile("main.go") {
		t.Error("expected main.go to be supported")
	}
	if IsSupportedFile("README.md") {
		t.Error("expected README.md to be unsupported")
	}
	if !IsSupportedFile("test.py") {
		t.Error("expected test.py to be supported")
	}
}

func TestSupportedExtensions(t *testing.T) {
	exts := SupportedExtensions()
	if len(exts) == 0 {
		t.Fatal("expected at least 1 supported extension")
	}
	// Verify .go is in the list
	found := false
	for _, ext := range exts {
		if ext == ".go" {
			found = true
		}
	}
	if !found {
		t.Error("expected .go in supported extensions")
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello\nworld", 2},
		{"a\nb\nc\n", 3},
		{"\n", 1},
		{"\n\n", 2},
	}
	for _, tt := range tests {
		got := CountLines(tt.input)
		if got != tt.want {
			t.Errorf("CountLines(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestExtractLines(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"

	tests := []struct {
		start, end int
		want       string
	}{
		{2, 4, "line2\nline3\nline4"},
		{1, 1, "line1"},
		{5, 5, "line5"},
		{0, 2, "line1\nline2"},  // start < 1 → clamped
		{4, 10, "line4\nline5"}, // end > lines → clamped
		{3, 2, ""},              // start > end
		{1, 5, content},         // full range
	}
	for _, tt := range tests {
		got := ExtractLines(content, tt.start, tt.end)
		if got != tt.want {
			t.Errorf("ExtractLines(_, %d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestFilePathToModulePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"internal/parser/go_parser.go", "internal.parser.go_parser"},
		{"main.go", "main"},
		{"a/b/c.py", "a.b.c"},
	}
	for _, tt := range tests {
		got := FilePathToModulePath(tt.input)
		if got != tt.want {
			t.Errorf("FilePathToModulePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/a/b/../c", "/a/c"},
		{"/a/b/./c", "/a/b/c"},
		{"./test", "test"},
	}
	for _, tt := range tests {
		got := NormalizePath(tt.input)
		if got != tt.want {
			t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRelativePath(t *testing.T) {
	base := "/home/user/project"
	target := "/home/user/project/internal/main.go"
	got := RelativePath(base, target)
	want := filepath.Join("internal", "main.go")
	if got != want {
		t.Errorf("RelativePath(%q, %q) = %q, want %q", base, target, got, want)
	}
}

func TestRelativePathError(t *testing.T) {
	// On some systems, Rel can return target if it can't compute relative
	got := RelativePath("", "/absolute/path")
	if got == "" {
		t.Error("RelativePath should return something")
	}
}
