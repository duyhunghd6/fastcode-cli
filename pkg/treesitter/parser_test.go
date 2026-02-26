package treesitter

import (
	"testing"
)

func TestParserInitGo(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("Failed to initialize go parser: %v", err)
	}
	if p.Language() != "go" {
		t.Errorf("Expected language go, got %s", p.Language())
	}

	code := []byte("package main\n\nvar x = 1\n")
	tree, err := p.Parse(code, "")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if tree == nil {
		t.Fatalf("Expected tree, got nil")
	}
	defer tree.Close()

	root := tree.RootNode()
	if root == nil || root.Type() != "source_file" {
		t.Errorf("Expected root node type source_file, got %v", root.Type())
	}
}

func TestParserUnsupportedLanguage(t *testing.T) {
	_, err := New("unsupported_language_123")
	if err == nil {
		t.Errorf("Expected error for unsupported language")
	}
}

func TestSetLanguagePython(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := p.SetLanguage("python"); err != nil {
		t.Fatalf("SetLanguage(python): %v", err)
	}
	if p.Language() != "python" {
		t.Errorf("Language() = %q, want python", p.Language())
	}
}

func TestSetLanguageJavaScript(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := p.SetLanguage("javascript"); err != nil {
		t.Fatalf("SetLanguage(javascript): %v", err)
	}
	if p.Language() != "javascript" {
		t.Errorf("Language() = %q, want javascript", p.Language())
	}
}

func TestSetLanguageUnsupported(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := p.SetLanguage("nonexistent_lang"); err == nil {
		t.Error("expected error for unsupported language")
	}
	// Language should remain as before
	if p.Language() != "go" {
		t.Errorf("Language() should remain go after failed switch, got %q", p.Language())
	}
}

func TestParseMultipleLanguages(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Parse Go
	goCode := []byte("package main\nfunc main() {}\n")
	tree, err := p.Parse(goCode, "go")
	if err != nil {
		t.Fatalf("Parse Go: %v", err)
	}
	tree.Close()

	// Parse Python (auto-switch)
	pyCode := []byte("def hello():\n    pass\n")
	tree, err = p.Parse(pyCode, "python")
	if err != nil {
		t.Fatalf("Parse Python: %v", err)
	}
	tree.Close()
	if p.Language() != "python" {
		t.Errorf("Language after Python parse = %q, want python", p.Language())
	}

	// Parse JavaScript
	jsCode := []byte("function hello() { return 1; }\n")
	tree, err = p.Parse(jsCode, "javascript")
	if err != nil {
		t.Fatalf("Parse JS: %v", err)
	}
	tree.Close()
	if p.Language() != "javascript" {
		t.Errorf("Language after JS parse = %q, want javascript", p.Language())
	}
}

func TestParseSameLanguageNoSwitch(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	code := []byte("package main\n")
	tree, err := p.Parse(code, "go")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	tree.Close()

	// Parse again with same language
	tree, err = p.Parse(code, "go")
	if err != nil {
		t.Fatalf("Parse again: %v", err)
	}
	tree.Close()
}

func TestParseEmptyLanguageUsesCurrent(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	code := []byte("package main\n")
	tree, err := p.Parse(code, "")
	if err != nil {
		t.Fatalf("Parse with empty lang: %v", err)
	}
	tree.Close()
	if p.Language() != "go" {
		t.Errorf("Language should still be go, got %q", p.Language())
	}
}

func TestLanguageCacheReuse(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Switch to python twice - second time should use cache
	if err := p.SetLanguage("python"); err != nil {
		t.Fatalf("SetLanguage python: %v", err)
	}
	if err := p.SetLanguage("go"); err != nil {
		t.Fatalf("SetLanguage go: %v", err)
	}
	if err := p.SetLanguage("python"); err != nil {
		t.Fatalf("SetLanguage python (cached): %v", err)
	}

	if len(p.cache) < 2 {
		t.Errorf("cache should have at least 2 entries, got %d", len(p.cache))
	}
}

func TestAllSupportedLanguages(t *testing.T) {
	languages := []string{
		"go", "python", "javascript", "typescript", "tsx",
		"java", "rust", "c", "cpp", "csharp",
	}

	for _, lang := range languages {
		p, err := New(lang)
		if err != nil {
			t.Errorf("New(%q) error: %v", lang, err)
			continue
		}
		if p.Language() != lang {
			t.Errorf("Language() = %q, want %q", p.Language(), lang)
		}
	}
}

func TestParseUnsupportedLanguageInParse(t *testing.T) {
	p, err := New("go")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	code := []byte("some code")
	_, err = p.Parse(code, "unsupported_lang_xyz")
	if err == nil {
		t.Error("expected error for unsupported language in Parse")
	}
}
