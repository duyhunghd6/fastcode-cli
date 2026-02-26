package parser

import (
	"log"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
	"github.com/duyhunghd6/fastcode-cli/internal/util"
	ts "github.com/duyhunghd6/fastcode-cli/pkg/treesitter"
)

// Parser dispatches parsing to language-specific extractors.
type Parser struct {
	tsParser *ts.Parser
}

// New creates a new code parser.
func New() *Parser {
	// Initialize with Go as default; will switch per file
	p, err := ts.New("go")
	if err != nil {
		log.Printf("[parser] warning: failed to initialize tree-sitter: %v", err)
	}
	return &Parser{tsParser: p}
}

// ParseFile parses a source file and extracts structured information.
func (p *Parser) ParseFile(filePath, content string) *types.FileParseResult {
	language := util.GetLanguageFromPath(filePath)
	if language == "" {
		return nil
	}

	result := &types.FileParseResult{
		FilePath:   filePath,
		Language:   language,
		TotalLines: util.CountLines(content),
	}

	// Non-code files (markdown, json, yaml, etc.) don't need tree-sitter parsing.
	// They're indexed as file-level elements for BM25 keyword search.
	if !isCodeLanguage(language) {
		return result
	}

	code := []byte(content)

	tree, err := p.tsParser.Parse(code, language)
	if err != nil {
		log.Printf("[parser] failed to parse %s: %v", filePath, err)
		return result
	}
	defer tree.Close()

	rootNode := tree.RootNode()

	switch language {
	case "go":
		parseGo(rootNode, code, result)
	case "python":
		parsePython(rootNode, code, result)
	case "javascript", "typescript", "tsx":
		parseJS(rootNode, code, result)
	case "java":
		parseJava(rootNode, code, result)
	case "rust":
		parseRust(rootNode, code, result)
	case "c", "cpp":
		parseC(rootNode, code, result)
	default:
		// Fallback for code languages without a dedicated parser
	}

	return result
}

// isCodeLanguage returns true if the language has a tree-sitter grammar
// and should be parsed for classes, functions, and imports.
func isCodeLanguage(lang string) bool {
	switch lang {
	case "go", "python", "javascript", "typescript", "tsx",
		"java", "rust", "c", "cpp", "csharp", "ruby", "php",
		"swift", "kotlin", "scala":
		return true
	}
	return false
}

// nodeText returns the UTF-8 text content of a tree-sitter node.
func nodeText(node interface{ Content([]byte) string }, code []byte) string {
	return node.Content(code)
}
