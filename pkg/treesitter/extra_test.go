package treesitter

import (
	"testing"
)

// TestParseLanguageSwitchError tests Parse with an unsupported language switch
func TestParseLanguageSwitchError(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Try switching to unsupported language during Parse
	_, err = p.Parse([]byte("code"), "unsupported_language")
	if err == nil {
		t.Error("expected error for unsupported language during Parse")
	}
}

// TestParseSameLanguage tests Parse with same language (no switch)
func TestParseSameLanguage(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tree, err := p.Parse([]byte("package main"), "go")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if tree == nil {
		t.Fatal("tree should not be nil")
	}
	tree.Close()
}

// TestParseEmptyLanguage tests Parse with empty language (no switch)
func TestParseEmptyLanguage(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tree, err := p.Parse([]byte("package main"), "")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if tree == nil {
		t.Fatal("tree should not be nil")
	}
	tree.Close()
}

// TestParseSwitchLanguage tests Parse successfully switching languages
func TestParseSwitchLanguage(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Switch to python
	tree, err := p.Parse([]byte("def main(): pass"), "python")
	if err != nil {
		t.Fatalf("Parse python: %v", err)
	}
	if tree == nil {
		t.Fatal("tree should not be nil")
	}
	tree.Close()

	if p.Language() != "python" {
		t.Errorf("Language = %q, want python", p.Language())
	}
}

// TestNewUnsupported tests New with unsupported language
func TestNewUnsupported(t *testing.T) {
	_, err := New("brainfuck")
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

// TestSetLanguageMultiple tests switching languages multiple times
func TestSetLanguageMultiple(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	languages := []string{"python", "javascript", "typescript", "tsx", "java", "rust", "c", "cpp", "csharp", "go"}
	for _, lang := range languages {
		if err := p.SetLanguage(lang); err != nil {
			t.Errorf("SetLanguage(%q): %v", lang, err)
		}
		if p.Language() != lang {
			t.Errorf("Language = %q, want %q", p.Language(), lang)
		}
	}
}

// TestLanguageCaching tests that language cache works
func TestLanguageCaching(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Switch away and back to test cache hit
	p.SetLanguage("python")
	p.SetLanguage("go") // Should use cache
	if p.Language() != "go" {
		t.Errorf("Language = %q, want go", p.Language())
	}
}
