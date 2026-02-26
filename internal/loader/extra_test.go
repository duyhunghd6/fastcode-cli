package loader

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRepositoryMaxFileSize tests that large files are skipped
func TestLoadRepositoryMaxFileSize(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-maxsize-*")
	defer os.RemoveAll(dir)

	// Write a large file
	bigContent := make([]byte, 2000)
	for i := range bigContent {
		bigContent[i] = 'a'
	}
	os.WriteFile(filepath.Join(dir, "big.go"), bigContent, 0644)
	os.WriteFile(filepath.Join(dir, "small.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	cfg.MaxFileSize = 1000 // Skip files > 1000 bytes

	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Should only include small.go
	for _, f := range repo.Files {
		if f.RelativePath == "big.go" {
			t.Error("big file should have been skipped due to MaxFileSize")
		}
	}
}

// TestLoadRepositoryExcludeFiles tests file exclusion by pattern
func TestLoadRepositoryExcludeFilePatterns(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-exclude-*")
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	cfg.ExcludeFiles = []string{"*_test.go"}

	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == "main_test.go" {
			t.Error("test file should have been excluded")
		}
	}
}

// TestLoadRepositoryExcludeDirs tests directory exclusion
func TestLoadRepositoryExcludeDirs(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-exdir-*")
	defer os.RemoveAll(dir)

	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "lib.go"), []byte("package vendor\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	cfg.ExcludeDirs = append(cfg.ExcludeDirs, "vendor")

	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == filepath.Join("vendor", "lib.go") {
			t.Error("vendor file should have been excluded")
		}
	}
}

// TestLoadRepositoryGitignore tests .gitignore pattern matching
func TestLoadRepositoryGitignore(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-gitignore-*")
	defer os.RemoveAll(dir)

	// Create .gitignore
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\nbuild/\n# comment\n\n"), 0644)

	// Create files
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(dir, "debug.log"), []byte("log data\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "build"), 0755)
	os.WriteFile(filepath.Join(dir, "build", "output.go"), []byte("package build\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == "debug.log" {
			t.Error("gitignored .log file should be excluded")
		}
	}
}

// TestLoadRepositoryDotDirSkipped tests that dot directories are skipped
func TestLoadRepositoryDotDirSkipped(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-dotdir-*")
	defer os.RemoveAll(dir)

	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.go"), []byte("package hidden\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == filepath.Join(".hidden", "secret.go") {
			t.Error(".hidden dir should be skipped")
		}
	}
}

// TestLoadRepositoryNotADir tests loading a file instead of directory
func TestLoadRepositoryNotADir(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "loader-notdir-*.go")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err := LoadRepository(tmpFile.Name(), DefaultConfig())
	if err == nil {
		t.Error("expected error for non-directory path")
	}
}

// TestMatchGitignoreNegation tests gitignore negation pattern
func TestMatchGitignoreNegation(t *testing.T) {
	// Negation pattern (starts with !) should not match
	result := matchGitignore("!important.go", "important.go")
	if result {
		t.Error("negation pattern should return false")
	}
}

// TestMatchGitignoreDirSuffix tests gitignore directory suffix pattern
func TestMatchGitignoreDirSuffix(t *testing.T) {
	result := matchGitignore("build/", "build")
	if !result {
		t.Error("dir pattern should match")
	}
}

// TestMatchGitignoreFullPath tests gitignore matching against full path
func TestMatchGitignoreFullPath(t *testing.T) {
	result := matchGitignore("*.log", "debug.log")
	if !result {
		t.Error("glob pattern should match filename")
	}
}

// TestLoadGitignoreNoFile tests loadGitignore with missing .gitignore
func TestLoadGitignoreNoFilePresent(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-nogitignore-*")
	defer os.RemoveAll(dir)

	patterns := loadGitignore(dir)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(patterns))
	}
}

// TestLoadGitignoreWithComments tests gitignore parsing with comments/blanks
func TestLoadGitignoreWithComments(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-gitignore-comments-*")
	defer os.RemoveAll(dir)

	content := "# comment\n\n*.log\n  \n*.tmp\n# another comment\n"
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644)

	patterns := loadGitignore(dir)
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d: %v", len(patterns), patterns)
	}
}

// TestReadFileContentError tests reading non-existent file
func TestReadFileContentError(t *testing.T) {
	_, err := ReadFileContent("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error reading nonexistent file")
	}
}

// TestReadFileContentSuccess tests reading an existing file
func TestReadFileContentSuccess(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "loader-read-*.go")
	tmpFile.WriteString("package main\n")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	content, err := ReadFileContent(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if content != "package main\n" {
		t.Errorf("content = %q", content)
	}
}
