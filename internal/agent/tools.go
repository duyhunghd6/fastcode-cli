package agent

import (
	"fmt"
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

// AvailableTools returns the tools the agent can use.
func AvailableTools() []Tool {
	return []Tool{
		{Name: "search_code", Description: "Semantic + keyword search for code elements matching a query"},
		{Name: "browse_file", Description: "Read the full content of a specific file"},
		{Name: "skim_file", Description: "Read only signatures and docstrings from a file (token-efficient)"},
		{Name: "search_graph", Description: "Find related code elements via dependency/call/inheritance graphs"},
		{Name: "list_files", Description: "List all files in the repository matching a pattern"},
	}
}

// ToolExecutor executes agent tools against the index.
type ToolExecutor struct {
	hybrid   *index.HybridRetriever
	embedder *llm.Embedder
	elements map[string]*types.CodeElement
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

// Execute runs a tool by name with the given argument.
func (te *ToolExecutor) Execute(toolName, arg string) (*ToolResult, error) {
	switch toolName {
	case "search_code":
		return te.searchCode(arg)
	case "browse_file":
		return te.browseFile(arg)
	case "skim_file":
		return te.skimFile(arg)
	case "list_files":
		return te.listFiles(arg)
	case "search_graph":
		// Stub: fall back to semantic search until graph index is implemented
		return te.searchCode(arg)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func (te *ToolExecutor) searchCode(query string) (*ToolResult, error) {
	var queryVec []float32
	if te.embedder != nil {
		vec, err := te.embedder.EmbedText(query)
		if err == nil {
			queryVec = vec
		}
	}

	results := te.hybrid.Search(query, queryVec, 10)
	var elements []types.CodeElement
	for _, r := range results {
		if r.Element != nil {
			elements = append(elements, *r.Element)
		}
	}

	return &ToolResult{
		ToolName: "search_code",
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
	return &ToolResult{ToolName: "list_files", Elements: files}, nil
}
