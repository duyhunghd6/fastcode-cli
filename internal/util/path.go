package util

import (
	"path/filepath"
	"strings"
)

// NormalizePath cleans and normalizes a file path.
func NormalizePath(p string) string {
	return filepath.Clean(p)
}

// RelativePath returns the relative path from base to target.
func RelativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

// FilePathToModulePath converts a file path to a module-style dotted path.
// e.g., "internal/parser/go_parser.go" → "internal.parser.go_parser"
func FilePathToModulePath(filePath string) string {
	// Remove extension
	ext := filepath.Ext(filePath)
	noExt := strings.TrimSuffix(filePath, ext)
	// Convert separators to dots
	return strings.ReplaceAll(noExt, string(filepath.Separator), ".")
}

// CountLines returns the number of lines in a string.
func CountLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	// If the string doesn't end with a newline, count the last line
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

// ExtractLines extracts lines [startLine, endLine] (1-indexed, inclusive) from content.
func ExtractLines(content string, startLine, endLine int) string {
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
