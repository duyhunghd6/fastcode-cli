package agent

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// Tool represents an agent action that can be invoked during retrieval.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToolResult holds the output of a tool execution.
type ToolResult struct {
	ToolName string              `json:"tool_name"`
	Elements []types.CodeElement `json:"elements,omitempty"`
	Text     string              `json:"text,omitempty"`
}

// FileCandidate represents a file found by search_codebase with match metadata.
// This mirrors Python's search result dict from agent_tools.py.
type FileCandidate struct {
	FilePath   string `json:"file_path"`
	MatchCount int    `json:"match_count"`
	RepoName   string `json:"repo_name"`
}

// AvailableTools returns the tools the agent can use (matching Python's tool schema).
func AvailableTools() []Tool {
	return []Tool{
		{Name: "search_codebase", Description: "Search for specific terms, classes, functions in file contents"},
		{Name: "list_directory", Description: "Explore directory structure by listing contents of a path"},
		{Name: "browse_file", Description: "Read the full content of a specific file"},
		{Name: "skim_file", Description: "Read only signatures and docstrings from a file (token-efficient)"},
	}
}

// ToolExecutor executes agent tools against the index.
type ToolExecutor struct {
	hybrid   *index.HybridRetriever
	embedder *llm.Embedder
	elements map[string]*types.CodeElement
	repoRoot string // Absolute path to the repository root (for filesystem search)
	repoName string // Name of the repository
}

// NewToolExecutor creates a new tool executor.
func NewToolExecutor(hybrid *index.HybridRetriever, embedder *llm.Embedder, elements []types.CodeElement) *ToolExecutor {
	elemMap := make(map[string]*types.CodeElement, len(elements))
	for i := range elements {
		elemMap[elements[i].ID] = &elements[i]
	}
	return &ToolExecutor{
		hybrid:   hybrid,
		embedder: embedder,
		elements: elemMap,
	}
}

// SetRepoRoot sets the repository root path for filesystem-based search.
func (te *ToolExecutor) SetRepoRoot(repoRoot, repoName string) {
	te.repoRoot = repoRoot
	te.repoName = repoName
}

// Execute runs a tool by name with the given argument.
func (te *ToolExecutor) Execute(toolName, arg string) (*ToolResult, error) {
	switch toolName {
	case "search_codebase", "search_code":
		return te.searchCode(arg)
	case "list_directory", "list_files":
		return te.listFiles(arg)
	case "browse_file":
		return te.browseFile(arg)
	case "skim_file":
		return te.skimFile(arg)
	case "search_graph":
		// Stub: fall back to semantic search until graph index is implemented
		return te.searchCode(arg)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ExecuteSearchCodebase performs real filesystem content search like Python's agent_tools.py.
// Returns FileCandidate list for LLM file selection. This is separate from the BM25 Execute path.
func (te *ToolExecutor) ExecuteSearchCodebase(searchTerm, filePattern string, useRegex bool) []FileCandidate {
	if te.repoRoot == "" || searchTerm == "" {
		return nil
	}

	// Build content search pattern
	var contentPattern *regexp.Regexp
	flags := "(?i)" // case-insensitive by default
	if useRegex {
		compiled, err := regexp.Compile(flags + searchTerm)
		if err != nil {
			log.Printf("[tools] invalid regex: %v", err)
			return nil
		}
		contentPattern = compiled
	} else {
		// Literal search — escape special characters
		escaped := regexp.QuoteMeta(searchTerm)
		contentPattern = regexp.MustCompile(flags + escaped)
	}

	// Directories to skip (matching Python's exclusions)
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"__pycache__":  true,
		"dist":         true,
		"build":        true,
		"venv":         true,
		"coverage":     true,
		".claude":      true,
		".kilocode":    true,
		".agent":       true,
		".agents":      true,
		".claude-flow": true,
	}

	var candidates []FileCandidate
	maxResults := 30

	_ = filepath.WalkDir(te.repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}

		// Skip directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Get relative path
		relPath, _ := filepath.Rel(te.repoRoot, path)
		relPath = filepath.ToSlash(relPath) // normalize to forward slashes

		// File pattern matching (simple glob on filename or path)
		if filePattern != "" && filePattern != "*" {
			matched, _ := filepath.Match(filePattern, d.Name())
			if !matched {
				// Also try matching against full relative path
				matched, _ = filepath.Match(filePattern, relPath)
			}
			if !matched {
				// Try glob-style extension match (e.g., "**/*.ts")
				parts := strings.Split(filePattern, "/")
				lastPart := parts[len(parts)-1]
				matched, _ = filepath.Match(lastPart, d.Name())
			}
			if !matched {
				return nil
			}
		}

		// Read file and search content
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)

		// Count matches
		matches := contentPattern.FindAllStringIndex(content, -1)
		matchCount := len(matches)

		// Also check filename/path match
		filenameMatch := contentPattern.MatchString(d.Name()) || contentPattern.MatchString(relPath)

		if matchCount > 0 || filenameMatch {
			candidates = append(candidates, FileCandidate{
				FilePath:   relPath,
				MatchCount: matchCount,
				RepoName:   te.repoName,
			})
		}

		if len(candidates) >= maxResults {
			return filepath.SkipAll
		}
		return nil
	})

	return candidates
}

// ExecuteListDirectory performs real filesystem directory listing.
// Returns FileCandidate list of files in the directory.
func (te *ToolExecutor) ExecuteListDirectory(dirPath string) []FileCandidate {
	if te.repoRoot == "" {
		return nil
	}

	targetDir := filepath.Join(te.repoRoot, dirPath)
	if dirPath == "." || dirPath == "" {
		targetDir = te.repoRoot
	}

	var candidates []FileCandidate

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		relPath, _ := filepath.Rel(te.repoRoot, filepath.Join(targetDir, entry.Name()))
		relPath = filepath.ToSlash(relPath)

		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}
		_ = entryType

		candidates = append(candidates, FileCandidate{
			FilePath: relPath,
			RepoName: te.repoName,
		})
	}

	return candidates
}

// FindElementsForFile retrieves all indexed elements for a given file path.
// This is used after LLM file selection to get actual code elements.
func (te *ToolExecutor) FindElementsForFile(filePath string) []types.CodeElement {
	var result []types.CodeElement
	for _, elem := range te.elements {
		if elem.RelativePath == filePath ||
			strings.HasSuffix(elem.RelativePath, filePath) ||
			strings.HasSuffix(filePath, elem.RelativePath) {
			result = append(result, *elem)
		}
	}
	return result
}

// Original BM25-based search (kept as fallback)
func (te *ToolExecutor) searchCode(query string) (*ToolResult, error) {
	var queryVec []float32
	if te.embedder != nil {
		vec, err := te.embedder.EmbedText(query)
		if err == nil {
			queryVec = vec
		}
	}

	results := te.hybrid.Search(query, queryVec, 5)
	var elements []types.CodeElement
	for _, r := range results {
		if r.Element != nil {
			elements = append(elements, *r.Element)
		}
	}

	return &ToolResult{
		ToolName: "search_codebase",
		Elements: elements,
	}, nil
}

func (te *ToolExecutor) browseFile(filePath string) (*ToolResult, error) {
	// Find the file element
	for _, elem := range te.elements {
		if elem.Type == "file" && (elem.RelativePath == filePath || strings.HasSuffix(elem.RelativePath, filePath)) {
			return &ToolResult{
				ToolName: "browse_file",
				Elements: []types.CodeElement{*elem},
				Text:     elem.Code,
			}, nil
		}
	}
	return &ToolResult{ToolName: "browse_file", Text: fmt.Sprintf("File not found: %s", filePath)}, nil
}

func (te *ToolExecutor) skimFile(filePath string) (*ToolResult, error) {
	// Find all elements from that file (functions, classes) — signatures only
	var elements []types.CodeElement
	for _, elem := range te.elements {
		if (elem.Type == "function" || elem.Type == "class") &&
			(elem.RelativePath == filePath || strings.HasSuffix(elem.RelativePath, filePath)) {
			// Create a skim copy with signature only (no full code)
			skim := *elem
			skim.Code = "" // token-efficient: omit full code
			elements = append(elements, skim)
		}
	}
	if len(elements) == 0 {
		return &ToolResult{ToolName: "skim_file", Text: fmt.Sprintf("No elements found in: %s", filePath)}, nil
	}
	return &ToolResult{ToolName: "skim_file", Elements: elements}, nil
}

func (te *ToolExecutor) listFiles(pattern string) (*ToolResult, error) {
	var files []types.CodeElement
	pattern = strings.ToLower(pattern)
	for _, elem := range te.elements {
		if elem.Type == "file" && strings.Contains(strings.ToLower(elem.RelativePath), pattern) {
			files = append(files, *elem)
		}
	}
	return &ToolResult{ToolName: "list_directory", Elements: files}, nil
}
