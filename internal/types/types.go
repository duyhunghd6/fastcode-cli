package types

// CodeElement represents a unified code element for indexing.
type CodeElement struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"` // "file", "class", "function", "documentation"
	Name         string         `json:"name"`
	FilePath     string         `json:"file_path"`
	RelativePath string         `json:"relative_path"`
	Language     string         `json:"language"`
	StartLine    int            `json:"start_line"`
	EndLine      int            `json:"end_line"`
	Code         string         `json:"code"`
	Signature    string         `json:"signature,omitempty"`
	Docstring    string         `json:"docstring,omitempty"`
	Summary      string         `json:"summary,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	RepoName     string         `json:"repo_name,omitempty"`
	RepoURL      string         `json:"repo_url,omitempty"`
}

// FunctionInfo holds extracted function/method metadata.
type FunctionInfo struct {
	Name       string   `json:"name"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	Docstring  string   `json:"docstring,omitempty"`
	Parameters []string `json:"parameters,omitempty"`
	ReturnType string   `json:"return_type,omitempty"`
	IsAsync    bool     `json:"is_async,omitempty"`
	IsMethod   bool     `json:"is_method,omitempty"`
	ClassName  string   `json:"class_name,omitempty"`
	Decorators []string `json:"decorators,omitempty"`
	Complexity int      `json:"complexity,omitempty"`
	Receiver   string   `json:"receiver,omitempty"` // Go-specific: method receiver
	Calls      []string `json:"calls,omitempty"`    // function/method names called within this function
}

// ClassInfo holds extracted class/struct/interface metadata.
type ClassInfo struct {
	Name       string         `json:"name"`
	StartLine  int            `json:"start_line"`
	EndLine    int            `json:"end_line"`
	Docstring  string         `json:"docstring,omitempty"`
	Bases      []string       `json:"bases,omitempty"` // parent classes / embedded types
	Methods    []FunctionInfo `json:"methods,omitempty"`
	Decorators []string       `json:"decorators,omitempty"`
	Kind       string         `json:"kind,omitempty"` // "class", "struct", "interface"
}

// ImportInfo holds extracted import statement metadata.
type ImportInfo struct {
	Module string   `json:"module"`
	Names  []string `json:"names,omitempty"`
	IsFrom bool     `json:"is_from,omitempty"` // Python: from X import Y
	Line   int      `json:"line"`
	Level  int      `json:"level,omitempty"` // Python relative import level
	Alias  string   `json:"alias,omitempty"`
}

// FileParseResult is the result of parsing a single source file.
type FileParseResult struct {
	FilePath        string         `json:"file_path"`
	Language        string         `json:"language"`
	Classes         []ClassInfo    `json:"classes,omitempty"`
	Functions       []FunctionInfo `json:"functions,omitempty"`
	Imports         []ImportInfo   `json:"imports,omitempty"`
	ModuleDocstring string         `json:"module_docstring,omitempty"`
	TotalLines      int            `json:"total_lines"`
	CodeLines       int            `json:"code_lines"`
	CommentLines    int            `json:"comment_lines"`
}
