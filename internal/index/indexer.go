package index

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/loader"
	"github.com/duyhunghd6/fastcode-cli/internal/parser"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// Indexer indexes a code repository at multiple levels (file, class, function, documentation).
type Indexer struct {
	parser   *parser.Parser
	repoName string
	Elements []types.CodeElement
}

// NewIndexer creates a new multi-level code indexer.
func NewIndexer(repoName string) *Indexer {
	return &Indexer{
		parser:   parser.New(),
		repoName: repoName,
	}
}

// IndexRepository parses all files in a repository and produces CodeElements.
func (idx *Indexer) IndexRepository(repo *loader.Repository) ([]types.CodeElement, error) {
	idx.repoName = repo.Name
	idx.Elements = nil

	for _, fi := range repo.Files {
		content, err := loader.ReadFileContent(fi.Path)
		if err != nil {
			log.Printf("[indexer] skip %s: %v", fi.RelativePath, err)
			continue
		}

		parseResult := idx.parser.ParseFile(fi.Path, content)
		if parseResult == nil {
			continue
		}

		idx.indexFile(fi, content, parseResult)
	}

	log.Printf("[indexer] indexed %d elements from %s (%d files)",
		len(idx.Elements), repo.Name, len(repo.Files))
	return idx.Elements, nil
}

func (idx *Indexer) indexFile(fi loader.FileInfo, content string, pr *types.FileParseResult) {
	// File-level element
	idx.addFileElement(fi, content, pr)

	// Class-level elements
	for _, cls := range pr.Classes {
		idx.addClassElement(fi, content, pr, cls)
	}

	// Function-level elements
	for _, fn := range pr.Functions {
		idx.addFunctionElement(fi, content, pr, fn)
	}

	// Documentation element (if module has docstring)
	if pr.ModuleDocstring != "" {
		idx.addDocElement(fi, pr)
	}
}

func (idx *Indexer) addFileElement(fi loader.FileInfo, content string, pr *types.FileParseResult) {
	summary := idx.generateFileSummary(pr)
	elem := types.CodeElement{
		ID:           idx.genID("file", fi.RelativePath),
		Type:         "file",
		Name:         fi.RelativePath,
		FilePath:     fi.Path,
		RelativePath: fi.RelativePath,
		Language:     fi.Language,
		StartLine:    1,
		EndLine:      pr.TotalLines,
		Code:         truncate(content, 4000),
		Summary:      summary,
		RepoName:     idx.repoName,
		Metadata: map[string]any{
			"total_lines":   pr.TotalLines,
			"num_classes":   len(pr.Classes),
			"num_functions": len(pr.Functions),
			"num_imports":   len(pr.Imports),
			"imports":       pr.Imports,
		},
	}
	idx.Elements = append(idx.Elements, elem)
}

func (idx *Indexer) addClassElement(fi loader.FileInfo, content string, pr *types.FileParseResult, cls types.ClassInfo) {
	code := extractCodeBlock(content, cls.StartLine, cls.EndLine)
	sig := fmt.Sprintf("%s %s", cls.Kind, cls.Name)
	if len(cls.Bases) > 0 {
		sig += " extends " + strings.Join(cls.Bases, ", ")
	}

	elem := types.CodeElement{
		ID:           idx.genID("class", fi.RelativePath, cls.Name),
		Type:         "class",
		Name:         cls.Name,
		FilePath:     fi.Path,
		RelativePath: fi.RelativePath,
		Language:     fi.Language,
		StartLine:    cls.StartLine,
		EndLine:      cls.EndLine,
		Code:         truncate(code, 3000),
		Signature:    sig,
		Docstring:    cls.Docstring,
		RepoName:     idx.repoName,
		Metadata: map[string]any{
			"kind":        cls.Kind,
			"bases":       cls.Bases,
			"num_methods": len(cls.Methods),
			"decorators":  cls.Decorators,
		},
	}
	idx.Elements = append(idx.Elements, elem)
}

func (idx *Indexer) addFunctionElement(fi loader.FileInfo, content string, pr *types.FileParseResult, fn types.FunctionInfo) {
	code := extractCodeBlock(content, fn.StartLine, fn.EndLine)
	sig := fn.Name + "(" + strings.Join(fn.Parameters, ", ") + ")"
	if fn.ReturnType != "" {
		sig += " " + fn.ReturnType
	}
	if fn.ClassName != "" {
		sig = fn.ClassName + "." + sig
	}

	elem := types.CodeElement{
		ID:           idx.genID("function", fi.RelativePath, fn.ClassName, fn.Name),
		Type:         "function",
		Name:         fn.Name,
		FilePath:     fi.Path,
		RelativePath: fi.RelativePath,
		Language:     fi.Language,
		StartLine:    fn.StartLine,
		EndLine:      fn.EndLine,
		Code:         truncate(code, 2000),
		Signature:    sig,
		Docstring:    fn.Docstring,
		RepoName:     idx.repoName,
		Metadata: map[string]any{
			"class_name": fn.ClassName,
			"is_method":  fn.IsMethod,
			"is_async":   fn.IsAsync,
			"receiver":   fn.Receiver,
			"complexity": fn.Complexity,
		},
	}
	idx.Elements = append(idx.Elements, elem)
}

func (idx *Indexer) addDocElement(fi loader.FileInfo, pr *types.FileParseResult) {
	elem := types.CodeElement{
		ID:           idx.genID("doc", fi.RelativePath),
		Type:         "documentation",
		Name:         fi.RelativePath + " (docs)",
		FilePath:     fi.Path,
		RelativePath: fi.RelativePath,
		Language:     fi.Language,
		StartLine:    1,
		EndLine:      1,
		Code:         pr.ModuleDocstring,
		Docstring:    pr.ModuleDocstring,
		RepoName:     idx.repoName,
	}
	idx.Elements = append(idx.Elements, elem)
}

func (idx *Indexer) generateFileSummary(pr *types.FileParseResult) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Language: %s, Lines: %d", pr.Language, pr.TotalLines))
	if len(pr.Classes) > 0 {
		names := make([]string, len(pr.Classes))
		for i, c := range pr.Classes {
			names[i] = c.Name
		}
		parts = append(parts, fmt.Sprintf("Classes: %s", strings.Join(names, ", ")))
	}
	if len(pr.Functions) > 0 {
		names := make([]string, 0, len(pr.Functions))
		for _, f := range pr.Functions {
			if !f.IsMethod {
				names = append(names, f.Name)
			}
		}
		if len(names) > 0 {
			parts = append(parts, fmt.Sprintf("Functions: %s", strings.Join(names, ", ")))
		}
	}
	return strings.Join(parts, "; ")
}

func (idx *Indexer) genID(elemType string, parts ...string) string {
	h := sha256.New()
	h.Write([]byte(idx.repoName))
	for _, p := range parts {
		h.Write([]byte(p))
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))[:12]
	return fmt.Sprintf("%s_%s_%s", idx.repoName, elemType, hash)
}

func extractCodeBlock(content string, startLine, endLine int) string {
	lines := strings.Split(content, "\n")
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return ""
	}
	return strings.Join(lines[startLine-1:endLine], "\n")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}
