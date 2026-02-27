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
		MaxFileSize: 5 * 1024 * 1024, // 5MB (matches Python)
		ExcludeDirs: []string{
			".git", "node_modules", "__pycache__",
			"dist", "build",
		},
		ExcludeFiles: []string{
			"*.pyc", "*.min.js", "*.bundle.js", "*.lock",
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

	// Separate negation patterns from normal patterns
	hasNegation := false
	for _, pat := range gitignorePatterns {
		if strings.HasPrefix(pat, "!") {
			hasNegation = true
			break
		}
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		relPath, _ := filepath.Rel(absRoot, path)

		// Skip excluded directories
		if d.IsDir() {
			dirName := d.Name()
			if excludeDirSet[dirName] {
				return filepath.SkipDir
			}
			// Check gitignore for directories — only SkipDir if there are
			// NO negation patterns (negation patterns require entering the
			// directory to check individual files)
			if !hasNegation {
				for _, pat := range gitignorePatterns {
					if matchGitignore(pat, relPath+"/") {
						return filepath.SkipDir
					}
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

		// Check gitignore (with negation support)
		if isGitignored(gitignorePatterns, relPath) {
			return nil
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

// isGitignored checks if a path is ignored by the gitignore patterns,
// properly handling negation patterns (lines starting with !).
// Patterns are evaluated in order; the last matching pattern wins.
func isGitignored(patterns []string, path string) bool {
	ignored := false
	for _, pat := range patterns {
		if strings.HasPrefix(pat, "!") {
			// Negation pattern: if it matches, UN-ignore the file
			negPat := strings.TrimPrefix(pat, "!")
			if matchGitignorePattern(negPat, path) {
				ignored = false
			}
		} else {
			if matchGitignorePattern(pat, path) {
				ignored = true
			}
		}
	}
	return ignored
}

// matchGitignore performs gitignore-compatible pattern matching (single pattern).
// Supports: simple globs, directory patterns (trailing /), path prefixes,
// wildcard subdirectory patterns (dir/*), and negation (!).
func matchGitignore(pattern, path string) bool {
	// Handle negation — treat as non-match for backward compatibility
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	return matchGitignorePattern(pattern, path)
}

// matchGitignorePattern performs the actual gitignore pattern matching.
func matchGitignorePattern(pattern, path string) bool {
	// Determine if this is a directory-only pattern
	dirOnly := strings.HasSuffix(pattern, "/")
	isDir := strings.HasSuffix(path, "/")
	cleanPattern := strings.TrimSuffix(pattern, "/")
	cleanPath := strings.TrimSuffix(path, "/")

	// Directory-only patterns should only match directories
	if dirOnly && !isDir {
		// Match files INSIDE the directory at any depth
		// e.g., pattern "perf/" should match "docs/perf/baseline.md"
		if strings.HasPrefix(cleanPath, cleanPattern+"/") || cleanPath == cleanPattern {
			return true
		}
		// Check if any path component matches the pattern (basename-level matching)
		// e.g., "perf/" matches "docs/perf/file.md" because "perf" appears as a dir component
		if !strings.Contains(cleanPattern, "/") {
			if strings.Contains(cleanPath, "/"+cleanPattern+"/") || strings.HasPrefix(cleanPath, cleanPattern+"/") {
				return true
			}
		}
		return false
	}

	// 1. Try matching against basename (e.g., "*.log" matches "debug.log")
	baseName := filepath.Base(cleanPath)
	matched, _ := filepath.Match(cleanPattern, baseName)
	if matched {
		return true
	}

	// 2. Try matching against full relative path (e.g., "dist/*" matches "dist/main.js")
	matched, _ = filepath.Match(cleanPattern, cleanPath)
	if matched {
		return true
	}

	// 3. Handle directory prefix patterns (e.g., "coverage/lcov-report/data/")
	//    These match any path that starts with the pattern prefix.
	if strings.Contains(cleanPattern, "/") {
		// Exact directory match
		if cleanPath == cleanPattern {
			return true
		}
		// Path is inside the pattern directory
		if strings.HasPrefix(cleanPath, cleanPattern+"/") {
			return true
		}
	}

	// 4. Patterns without "/" are basename-level: match any path component
	//    e.g., "thirdparty" matches "thirdparty/CMakeLists.txt" and "src/thirdparty/foo.h"
	if !strings.Contains(cleanPattern, "/") && !strings.ContainsAny(cleanPattern, "*?[") {
		// Check if the pattern matches as a directory component in the path
		if strings.HasPrefix(cleanPath, cleanPattern+"/") {
			return true
		}
		if strings.Contains(cleanPath, "/"+cleanPattern+"/") {
			return true
		}
	}

	return false
}
