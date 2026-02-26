package util

import (
	"path/filepath"
	"strings"
)

// Supported language extensions.
var languageExtensions = map[string]string{
	".go":    "go",
	".py":    "python",
	".js":    "javascript",
	".jsx":   "javascript",
	".ts":    "typescript",
	".tsx":   "tsx",
	".java":  "java",
	".rs":    "rust",
	".c":     "c",
	".h":     "c",
	".cpp":   "cpp",
	".cc":    "cpp",
	".cxx":   "cpp",
	".hpp":   "cpp",
	".cs":    "csharp",
	".rb":    "ruby",
	".php":   "php",
	".swift": "swift",
	".kt":    "kotlin",
	".scala": "scala",
}

// GetLanguageFromExtension returns the language name for a file extension.
// Returns empty string if unsupported.
func GetLanguageFromExtension(ext string) string {
	return languageExtensions[strings.ToLower(ext)]
}

// GetLanguageFromPath returns the language name for a file path.
func GetLanguageFromPath(filePath string) string {
	ext := filepath.Ext(filePath)
	return GetLanguageFromExtension(ext)
}

// IsSupportedFile returns true if the file extension is a supported language.
func IsSupportedFile(filePath string) bool {
	return GetLanguageFromPath(filePath) != ""
}

// SupportedExtensions returns all supported file extensions.
func SupportedExtensions() []string {
	exts := make([]string, 0, len(languageExtensions))
	for ext := range languageExtensions {
		exts = append(exts, ext)
	}
	return exts
}
