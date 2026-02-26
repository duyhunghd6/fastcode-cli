package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestRepo(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "fastcode-loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp repo: %v", err)
	}

	// Create a .gitignore
	err = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules/\n*.log\n!important.log\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}
	// Create a normal go file
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}
	// Create a normal python file
	err = os.WriteFile(filepath.Join(dir, "app.py"), []byte("def hello():\n  pass\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create app.py: %v", err)
	}
	// Create an ignored directory and file
	nodeModules := filepath.Join(dir, "node_modules")
	err = os.MkdirAll(nodeModules, 0755)
	if err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}
	err = os.WriteFile(filepath.Join(nodeModules, "ignored.js"), []byte("console.log('test')\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignored.js: %v", err)
	}
	// Create an ignored by extension file
	err = os.WriteFile(filepath.Join(dir, "system.log"), []byte("Log entry"), 0644)
	if err != nil {
		t.Fatalf("Failed to create system.log: %v", err)
	}
	// Create a huge file to test max file size logic
	hugeFile := make([]byte, 1024*1024+100) // Slightly above 1MB
	err = os.WriteFile(filepath.Join(dir, "huge.go"), hugeFile, 0644)
	if err != nil {
		t.Fatalf("Failed to create huge file: %v", err)
	}

	// Create a subdirectory with dot prefix (should be excluded)
	dotDir := filepath.Join(dir, ".hidden")
	os.MkdirAll(dotDir, 0755)
	os.WriteFile(filepath.Join(dotDir, "secret.go"), []byte("package secret\n"), 0644)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxFileSize <= 0 {
		t.Error("MaxFileSize should be > 0")
	}
	if len(cfg.ExcludeDirs) == 0 {
		t.Error("ExcludeDirs should not be empty")
	}
	if len(cfg.ExcludeFiles) == 0 {
		t.Error("ExcludeFiles should not be empty")
	}
}

func TestLoadRepository(t *testing.T) {
	dir, cleanup := createTestRepo(t)
	defer cleanup()

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}

	if repo.Name != filepath.Base(dir) {
		t.Errorf("Expected repo name %q, got %q", filepath.Base(dir), repo.Name)
	}

	expectedFiles := map[string]bool{
		"main.go": false,
		"app.py":  false,
	}
	unexpectedFiles := []string{
		"node_modules/ignored.js",
		"system.log",
	}

	for _, fi := range repo.Files {
		if _, ok := expectedFiles[fi.RelativePath]; ok {
			expectedFiles[fi.RelativePath] = true
		}
		for _, unexpected := range unexpectedFiles {
			if fi.RelativePath == unexpected || fi.RelativePath == filepath.FromSlash(unexpected) {
				t.Errorf("Found unexpectedly loaded file: %v", fi.RelativePath)
			}
		}
	}

	for exp, found := range expectedFiles {
		if !found {
			t.Errorf("Expected to load file %s, but it was not found", exp)
		}
	}
}

func TestLoadRepositoryNonExistent(t *testing.T) {
	_, err := LoadRepository("/nonexistent/path/xyz123", DefaultConfig())
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadRepositoryNotDir(t *testing.T) {
	tempFile, err := os.CreateTemp("", "fastcode-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	_, err = LoadRepository(tempFile.Name(), DefaultConfig())
	if err == nil {
		t.Error("expected error for non-directory")
	}
}

func TestLoadRepositoryNoGitignore(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-no-gitignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	repo, err := LoadRepository(dir, DefaultConfig())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(repo.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(repo.Files))
	}
}

func TestLoadRepositoryExcludeFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-exclude-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(dir, "app.js.map"), []byte("sourcemap\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for _, fi := range repo.Files {
		if fi.RelativePath == "app.js.map" {
			t.Error(".map file should be excluded by default config")
		}
	}
}

func TestReadFileContent(t *testing.T) {
	dir, cleanup := createTestRepo(t)
	defer cleanup()

	content, err := ReadFileContent(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	expected := "package main\nfunc main() {}\n"
	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}

	_, err = ReadFileContent(filepath.Join(dir, "does_not_exist.go"))
	if err == nil {
		t.Errorf("Expected an err when reading a missing file")
	}
}

func TestMatchGitignore(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.log", "error.log", true},
		{"*.log", "error.txt", false},
		{"node_modules/", "node_modules/", true},
		{"!important.log", "important.log", false}, // negation
		{"build", "build", true},
		{"*.py", "app.py", true},
	}
	for _, tt := range tests {
		got := matchGitignore(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchGitignore(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}

func TestLoadGitignoreNoFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-no-gi-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	patterns := loadGitignore(dir)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns when no .gitignore, got %d", len(patterns))
	}
}

func TestLoadGitignoreComments(t *testing.T) {
	dir, err := os.MkdirTemp("", "fastcode-gi-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	content := "# this is a comment\n\n*.log\nbuild/\n"
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644)

	patterns := loadGitignore(dir)
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns (excluding comment and blank), got %d: %v", len(patterns), patterns)
	}
}

func TestFileInfoLanguage(t *testing.T) {
	dir, cleanup := createTestRepo(t)
	defer cleanup()

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range repo.Files {
		if fi.Language == "" {
			t.Errorf("file %q has empty language", fi.RelativePath)
		}
	}
}

func TestLoadRepositoryDotDir(t *testing.T) {
	dir, cleanup := createTestRepo(t)
	defer cleanup()

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Dot-prefixed dirs are now loaded (matching Python behavior)
	// Only .git (in ExcludeDirs) should be excluded
	foundHidden := false
	for _, fi := range repo.Files {
		if fi.RelativePath == ".hidden/secret.go" || fi.RelativePath == filepath.Join(".hidden", "secret.go") {
			foundHidden = true
		}
	}
	if !foundHidden {
		t.Error(".hidden/secret.go should be loaded (dot dirs are no longer blanket-excluded)")
	}
}
