package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// === buildRootCmd Tests ===

func TestBuildRootCmd(t *testing.T) {
	cmd := buildRootCmd()
	if cmd == nil {
		t.Fatal("buildRootCmd returned nil")
	}
	if cmd.Use != "fastcode" {
		t.Errorf("Use = %q, want fastcode", cmd.Use)
	}
	if cmd.Version != version {
		t.Errorf("Version = %q, want %s", cmd.Version, version)
	}
}

func TestBuildRootCmdSubcommands(t *testing.T) {
	cmd := buildRootCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, expected := range []string{"index", "query", "serve-mcp"} {
		if !names[expected] {
			t.Errorf("missing subcommand: %s", expected)
		}
	}
}

func TestRootCmdHelp(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("help output should not be empty")
	}
}

func TestRootCmdVersion(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version error: %v", err)
	}
}

// === Index Command Tests ===

func TestIndexCmdSuccess(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-idx-cmd-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-idx-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("index cmd: %v", err)
	}
}

func TestIndexCmdForce(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-idx-force-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-idx-force-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings", "--force"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("index cmd --force: %v", err)
	}
}

func TestIndexCmdJSON(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-idx-json-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-idx-json-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings", "--json"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("index cmd --json: %v", err)
	}
}

func TestIndexCmdCached(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-idx-cached-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-idx-cached-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	// First index
	cmd1 := buildRootCmd()
	cmd1.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings"})
	cmd1.Execute()

	// Second index (cached) — exercises the Cached branch
	cmd2 := buildRootCmd()
	cmd2.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd2.Execute()
	if err != nil {
		t.Fatalf("index cmd cached: %v", err)
	}
}

func TestIndexCmdInvalidPath(t *testing.T) {
	cacheDir, _ := os.MkdirTemp("", "fastcode-idx-err-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"index", "/nonexistent/path", "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestIndexCmdMissingArg(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"index"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no repo-path given")
	}
}

// === Query Command Tests ===

func TestQueryCmdWithRepo(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-qry-cmd-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qry-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"query", "what is main?", "--repo", repoDir, "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("query cmd with repo: %v", err)
	}
}

func TestQueryCmdMultiWordQuestion(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-qry-multi-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qry-multi-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"query", "how", "does", "the", "main", "function", "work?", "--repo", repoDir, "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("query cmd multi-word: %v", err)
	}
}

func TestQueryCmdNoIndex(t *testing.T) {
	cacheDir, _ := os.MkdirTemp("", "fastcode-qry-noindex-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"query", "test question", "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when querying without index")
	}
}

func TestQueryCmdInvalidRepo(t *testing.T) {
	cacheDir, _ := os.MkdirTemp("", "fastcode-qry-err-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"query", "test", "--repo", "/nonexistent", "--cache-dir", cacheDir, "--no-embeddings"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid repo")
	}
}

func TestQueryCmdMissingArgs(t *testing.T) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"query"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no question given")
	}
}

func TestQueryCmdJSONOutput(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-qry-json-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-qry-json-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"query", "what is main?", "--repo", repoDir, "--cache-dir", cacheDir, "--no-embeddings", "--json"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("query cmd --json: %v", err)
	}
}

// === Flags Tests ===

func TestPersistentFlagsCacheDir(t *testing.T) {
	cmd := buildRootCmd()
	flag := cmd.PersistentFlags().Lookup("cache-dir")
	if flag == nil {
		t.Fatal("cache-dir flag not found")
	}
	if flag.DefValue != "" {
		t.Errorf("cache-dir default = %q, want empty", flag.DefValue)
	}
}

func TestPersistentFlagsEmbeddingModel(t *testing.T) {
	cmd := buildRootCmd()
	flag := cmd.PersistentFlags().Lookup("embedding-model")
	if flag == nil {
		t.Fatal("embedding-model flag not found")
	}
	if flag.DefValue != "text-embedding-3-small" {
		t.Errorf("embedding-model default = %q", flag.DefValue)
	}
}

func TestPersistentFlagsNoEmbeddings(t *testing.T) {
	cmd := buildRootCmd()
	flag := cmd.PersistentFlags().Lookup("no-embeddings")
	if flag == nil {
		t.Fatal("no-embeddings flag not found")
	}
}

func TestIndexFlagForce(t *testing.T) {
	cmd := buildRootCmd()
	indexCmd, _, _ := cmd.Find([]string{"index"})
	if indexCmd == nil {
		t.Fatal("index command not found")
	}
	flag := indexCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not found on index command")
	}
}

func TestIndexFlagJSON(t *testing.T) {
	cmd := buildRootCmd()
	indexCmd, _, _ := cmd.Find([]string{"index"})
	if indexCmd == nil {
		t.Fatal("index command not found")
	}
	flag := indexCmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("json flag not found on index command")
	}
}

func TestQueryFlagRepo(t *testing.T) {
	cmd := buildRootCmd()
	queryCmd, _, _ := cmd.Find([]string{"query"})
	if queryCmd == nil {
		t.Fatal("query command not found")
	}
	flag := queryCmd.Flags().Lookup("repo")
	if flag == nil {
		t.Fatal("repo flag not found on query command")
	}
}

func TestServeMCPFlagPort(t *testing.T) {
	cmd := buildRootCmd()
	serveCmd, _, _ := cmd.Find([]string{"serve-mcp"})
	if serveCmd == nil {
		t.Fatal("serve-mcp command not found")
	}
	flag := serveCmd.Flags().Lookup("port")
	if flag == nil {
		t.Fatal("port flag not found on serve-mcp command")
	}
	if flag.DefValue != "9999" {
		t.Errorf("port default = %q, want 9999", flag.DefValue)
	}
}

// === Custom Embedding Model Flag ===

func TestIndexWithCustomEmbeddingModel(t *testing.T) {
	repoDir, _ := os.MkdirTemp("", "fastcode-emb-model-*")
	defer os.RemoveAll(repoDir)
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-emb-model-cache-*")
	defer os.RemoveAll(cacheDir)

	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", origKey)

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"index", repoDir, "--cache-dir", cacheDir, "--no-embeddings", "--embedding-model", "custom-model"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("index with custom model: %v", err)
	}
}

// Note: serveMCP is a thin wrapper around buildMCPMux (100% covered)
// + http.ListenAndServe which blocks and cannot be unit tested.
