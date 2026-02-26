package treesitter

import (
	"context"
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Parser wraps go-tree-sitter with multi-language support and caching.
type Parser struct {
	mu       sync.Mutex
	parser   *sitter.Parser
	langName string
	cache    map[string]*sitter.Language
}

// New creates a new Parser initialized for the given language.
func New(language string) (*Parser, error) {
	p := &Parser{
		parser: sitter.NewParser(),
		cache:  make(map[string]*sitter.Language),
	}
	if err := p.SetLanguage(language); err != nil {
		return nil, err
	}
	return p, nil
}

// SetLanguage switches the parser to a different language.
func (p *Parser) SetLanguage(language string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	lang, err := p.getLanguage(language)
	if err != nil {
		return err
	}
	p.parser.SetLanguage(lang)
	p.langName = language
	return nil
}

// Parse parses source code and returns a tree-sitter Tree.
// If language is non-empty and different from the current one, it switches first.
func (p *Parser) Parse(code []byte, language string) (*sitter.Tree, error) {
	if language != "" && language != p.langName {
		if err := p.SetLanguage(language); err != nil {
			return nil, err
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	tree, err := p.parser.ParseCtx(context.Background(), nil, code)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return tree, nil
}

// Language returns the current language name.
func (p *Parser) Language() string {
	return p.langName
}

// getLanguage returns the sitter.Language for the given name, using cache.
func (p *Parser) getLanguage(name string) (*sitter.Language, error) {
	if lang, ok := p.cache[name]; ok {
		return lang, nil
	}

	var lang *sitter.Language
	switch name {
	case "go":
		lang = golang.GetLanguage()
	case "python":
		lang = python.GetLanguage()
	case "javascript":
		lang = javascript.GetLanguage()
	case "typescript":
		lang = typescript.GetLanguage()
	case "tsx":
		lang = tsx.GetLanguage()
	case "java":
		lang = java.GetLanguage()
	case "rust":
		lang = rust.GetLanguage()
	case "c":
		lang = c.GetLanguage()
	case "cpp":
		lang = cpp.GetLanguage()
	case "csharp":
		lang = csharp.GetLanguage()
	default:
		return nil, fmt.Errorf("unsupported language: %s", name)
	}

	p.cache[name] = lang
	return lang, nil
}
