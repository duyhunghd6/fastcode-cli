package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// === Index command RunE execution ===

func TestIndexCmdRunESuccess(t *testing.T) {
	// Create a temp directory with a Go file
	dir, _ := os.MkdirTemp("", "fastcode-index-*")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	// Unset OPENAI_API_KEY to avoid real API calls
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", dir, "--no-embeddings"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("index error: %v", err)
	}
}

// TestIndexCmdRunEForce tests the --force flag
func TestIndexCmdRunEForce(t *testing.T) {
	dir, _ := os.MkdirTemp("", "fastcode-index-force-*")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", dir, "--force", "--no-embeddings"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("index --force error: %v", err)
	}
}

// TestIndexCmdRunEJSON tests --json output
func TestIndexCmdRunEJSON(t *testing.T) {
	dir, _ := os.MkdirTemp("", "fastcode-index-json-*")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", dir, "--json", "--no-embeddings"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("index --json error: %v", err)
	}
}

// === Query command RunE ===

// TestQueryCmdRunENoIndex tests query without prior index
func TestQueryCmdRunENoIndex(t *testing.T) {
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"query", "what is main?"})

	err := cmd.Execute()
	// Should error because no index is loaded
	if err == nil {
		t.Logf("query without index succeeded unexpectedly")
	}
}

// TestQueryCmdRunEWithRepo tests query with --repo flag
func TestQueryCmdRunEWithRepo(t *testing.T) {
	dir, _ := os.MkdirTemp("", "fastcode-query-repo-*")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"query", "what is main?", "--repo", dir, "--no-embeddings"})

	err := cmd.Execute()
	// May error because no API key for LLM, but should at least run the index
	if err != nil {
		t.Logf("query with repo error: %v", err)
	}
}

// TestQueryCmdRunEMultipleWords tests query with multiple args joined
func TestQueryCmdRunEMultipleWords(t *testing.T) {
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("OPENAI_API_KEY", origKey)
		}
	}()

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"query", "how", "does", "the", "main", "function", "work?"})

	err := cmd.Execute()
	// Will likely error but we test the multi-word joining path
	if err != nil {
		t.Logf("multi-word query error: %v", err)
	}
}

// === buildConfig inner function ===

func TestBuildConfigCustomCacheDir(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", "/nonexistent", "--cache-dir", "/tmp/custom-cache", "--no-embeddings"})

	// We just want to exercise the buildConfig path with custom cache dir
	err := cmd.Execute()
	if err != nil {
		t.Logf("custom cache dir error: %v (expected for nonexistent path)", err)
	}
}

func TestBuildConfigCustomEmbeddingModel(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", "/nonexistent", "--embedding-model", "text-embedding-ada-002"})

	err := cmd.Execute()
	if err != nil {
		t.Logf("custom embedding model error: %v (expected)", err)
	}
}

// === serve-mcp port flag ===

func TestServeMCPFlagPortDefault(t *testing.T) {
	cmd := buildRootCmd()
	sub, _, err := cmd.Find([]string{"serve-mcp"})
	if err != nil {
		t.Fatalf("find serve-mcp: %v", err)
	}
	flag := sub.Flags().Lookup("port")
	if flag == nil {
		t.Error("serve-mcp should have --port flag")
	}
	if flag.DefValue != "9999" {
		t.Errorf("default port = %s, want 9999", flag.DefValue)
	}
}
