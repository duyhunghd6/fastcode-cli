package loader

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/util"
)

// FileInfo represents a loaded file from the repository.
type FileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Language     string `json:"language"`
	Size         int64  `json:"size"`
}

// Config holds loader configuration.
type Config struct {
	MaxFileSize  int64    // Maximum file size in bytes (default: 1MB)
	ExcludeDirs  []string // Directories to exclude
	ExcludeFiles []string // File patterns to exclude
}

// DefaultConfig returns the default loader configuration.
func DefaultConfig() Config {
	return Config{
		MaxFileSize: 1024 * 1024, // 1MB
		ExcludeDirs: []string{
			".git", ".svn", ".hg", "node_modules", "__pycache__",
			".venv", "venv", ".env", "vendor", "dist", "build",
			".idea", ".vscode", ".DS_Store", "target",
		},
		ExcludeFiles: []string{
			"*.min.js", "*.min.css", "*.map", "*.lock",
			"package-lock.json", "yarn.lock", "go.sum",
		},
	}
}

// Repository represents a loaded code repository.
type Repository struct {
	RootPath string
	Name     string
	Files    []FileInfo
}

// LoadRepository walks a repository directory and returns all supported source files.
func LoadRepository(rootPath string, cfg Config) (*Repository, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q: %w", rootPath, err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot access %q: %w", absRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", absRoot)
	}

	repo := &Repository{
		RootPath: absRoot,
		Name:     filepath.Base(absRoot),
	}

	// Load .gitignore patterns
	gitignorePatterns := loadGitignore(absRoot)

	excludeDirSet := make(map[string]bool, len(cfg.ExcludeDirs))
	for _, d := range cfg.ExcludeDirs {
		excludeDirSet[d] = true
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		relPath, _ := filepath.Rel(absRoot, path)

		// Skip excluded directories
		if d.IsDir() {
			dirName := d.Name()
			if excludeDirSet[dirName] || strings.HasPrefix(dirName, ".") {
				return filepath.SkipDir
			}
			// Check gitignore
			for _, pat := range gitignorePatterns {
				if matchGitignore(pat, relPath+"/") {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check file support
		if !util.IsSupportedFile(path) {
			return nil
		}

		// Check file size
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		if cfg.MaxFileSize > 0 && fi.Size() > cfg.MaxFileSize {
			return nil
		}

		// Check exclude patterns
		for _, pat := range cfg.ExcludeFiles {
			matched, _ := filepath.Match(pat, d.Name())
			if matched {
				return nil
			}
		}

		// Check gitignore
		for _, pat := range gitignorePatterns {
			if matchGitignore(pat, relPath) {
				return nil
			}
		}

		repo.Files = append(repo.Files, FileInfo{
			Path:         path,
			RelativePath: relPath,
			Language:     util.GetLanguageFromPath(path),
			Size:         fi.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk error: %w", err)
	}

	return repo, nil
}

// ReadFileContent reads the content of a file.
func ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// loadGitignore reads .gitignore patterns from the repository root.
func loadGitignore(rootPath string) []string {
	f, err := os.Open(filepath.Join(rootPath, ".gitignore"))
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// matchGitignore performs a simplified gitignore pattern match.
func matchGitignore(pattern, path string) bool {
	// Handle negation
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	// Handle directory-only patterns
	pattern = strings.TrimSuffix(pattern, "/")
	// Simple glob match
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}
	// Try matching against the full relative path
	matched, _ = filepath.Match(pattern, path)
	return matched
}
