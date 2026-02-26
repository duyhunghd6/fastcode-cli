package loader

import (
	"os"
	"path/filepath"
	"testing"
)

// === LoadRepository: WalkDir entry error (inaccessible path) ===

func TestLoadRepositoryWalkDirError(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-walkdir-err-*")
	defer os.RemoveAll(dir)

	// Create a subdirectory and make it inaccessible
	subDir := filepath.Join(dir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main\n"), 0644)
	os.Chmod(subDir, 0000)
	defer os.Chmod(subDir, 0755) // cleanup

	// Create a readable file in root
	os.WriteFile(filepath.Join(dir, "root.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository should not error for inaccessible subdir: %v", err)
	}
	// Should have the root.go file but not the inaccessible subdir
	if len(repo.Files) < 1 {
		t.Error("should include root.go")
	}
}

// === LoadRepository: gitignore directory match ===

func TestLoadRepositoryGitignoreSkipsDir(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-gitignore-dir-*")
	defer os.RemoveAll(dir)

	// Create .gitignore with a directory pattern
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("build\nvendor\n"), 0644)

	// Create files in ignored directories
	os.MkdirAll(filepath.Join(dir, "build"), 0755)
	os.WriteFile(filepath.Join(dir, "build", "output.go"), []byte("package build\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "dep.go"), []byte("package dep\n"), 0644)

	// Create a file in a non-ignored directory
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	// Should not include build/ or vendor/ files
	for _, f := range repo.Files {
		if f.RelativePath == "build/output.go" || f.RelativePath == "vendor/dep.go" {
			t.Errorf("gitignored file should be excluded: %s", f.RelativePath)
		}
	}
}

// === LoadRepository: file gitignore match (not dir) ===

func TestLoadRepositoryGitignoreSkipsFile(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-gitignore-file-*")
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n*.tmp\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(dir, "debug.log"), []byte("log data\n"), 0644)
	os.WriteFile(filepath.Join(dir, "temp.tmp"), []byte("temp\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == "debug.log" || f.RelativePath == "temp.tmp" {
			t.Errorf("gitignored file should be excluded: %s", f.RelativePath)
		}
	}
}

// === LoadRepository: ExcludeDirs actually skips subdirectory ===

func TestLoadRepositoryExcludeDirsSkip(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-excludedirs-*")
	defer os.RemoveAll(dir)

	os.MkdirAll(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "dep.js"), []byte("exports={}"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	cfg.ExcludeDirs = append(cfg.ExcludeDirs, "node_modules")
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == "node_modules/dep.js" {
			t.Error("ExcludeDirs should skip node_modules")
		}
	}
}

// === LoadRepository: IsSupportedFile filters unsupported ===

func TestLoadRepositoryUnsupportedFileSkipped(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-unsupported-*")
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47}, 0644) // PNG header
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	for _, f := range repo.Files {
		if f.RelativePath == "image.png" {
			t.Error("unsupported file (PNG) should be skipped")
		}
	}
}

// === LoadRepository: absolute path resolution ===

func TestLoadRepositoryAbsResolution(t *testing.T) {
	dir, _ := os.MkdirTemp("", "loader-abs-*")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	cfg := DefaultConfig()
	repo, err := LoadRepository(dir, cfg)
	if err != nil {
		t.Fatalf("LoadRepository: %v", err)
	}

	if !filepath.IsAbs(repo.RootPath) {
		t.Errorf("RootPath should be absolute, got %q", repo.RootPath)
	}
	if repo.Name == "" {
		t.Error("repo name should not be empty")
	}
}
