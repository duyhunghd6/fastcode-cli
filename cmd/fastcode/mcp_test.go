package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/orchestrator"
)

// === Helper function tests (writeJSON, writeError, writeToolResult) ===

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]string{"key": "value"})
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
	}
	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["key"] != "value" {
		t.Errorf("body key = %q, want value", result["key"])
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		msg  string
		code int
	}{
		{"bad request", 400},
		{"not found", 404},
		{"server error", 500},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		writeError(w, tt.msg, tt.code)
		if w.Code != tt.code {
			t.Errorf("writeError(%q, %d): status = %d", tt.msg, tt.code, w.Code)
		}
		var result map[string]any
		json.Unmarshal(w.Body.Bytes(), &result)
		errMap, ok := result["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected error field")
		}
		if errMap["message"] != tt.msg {
			t.Errorf("error message = %q, want %q", errMap["message"], tt.msg)
		}
	}
}

func TestWriteToolResult(t *testing.T) {
	w := httptest.NewRecorder()
	writeToolResult(w, map[string]string{"answer": "hello"})
	var result map[string]any
	json.Unmarshal(w.Body.Bytes(), &result)
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("expected content array")
	}
	first, ok := content[0].(map[string]any)
	if !ok {
		t.Fatal("expected map in content[0]")
	}
	if first["type"] != "text" {
		t.Errorf("content type = %q, want text", first["type"])
	}
	if first["text"] == nil || first["text"] == "" {
		t.Error("expected text content")
	}
}

func TestVersionVar(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}

// === Full MCP server integration tests ===

func setupTestServer(t *testing.T) (*httptest.Server, string, func()) {
	t.Helper()

	// Create a temp repo with code
	repoDir, err := os.MkdirTemp("", "fastcode-mcp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	goContent := `package main

import "fmt"

type Server struct {
	Port int
}

func (s *Server) Start() error {
	fmt.Println("starting")
	return nil
}

func main() {
	s := &Server{Port: 8080}
	s.Start()
}
`
	os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644)

	cacheDir, _ := os.MkdirTemp("", "fastcode-mcp-cache-*")

	// Clear API key to avoid real LLM calls
	origKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	cfg := orchestrator.Config{
		CacheDir:     cacheDir,
		BatchSize:    32,
		NoEmbeddings: true,
	}
	handler := buildMCPMux(orchestrator.NewEngine(cfg))
	server := httptest.NewServer(handler)

	cleanup := func() {
		server.Close()
		os.RemoveAll(repoDir)
		os.RemoveAll(cacheDir)
		os.Setenv("OPENAI_API_KEY", origKey)
	}

	return server, repoDir, cleanup
}

func TestMCPInitialize(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/mcp/initialize")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v", result["protocolVersion"])
	}
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("expected serverInfo")
	}
	if serverInfo["name"] != "fastcode-cli" {
		t.Errorf("server name = %v", serverInfo["name"])
	}
}

func TestMCPToolsList(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/mcp/tools/list")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("expected tools array")
	}
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]any)
		toolNames[toolMap["name"].(string)] = true
	}
	for _, expected := range []string{"index_repository", "query_codebase", "search_code"} {
		if !toolNames[expected] {
			t.Errorf("missing tool: %s", expected)
		}
	}
}

func TestMCPHealth(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
	if result["version"] != version {
		t.Errorf("version = %q, want %s", result["version"], version)
	}
}

func TestMCPToolsCallIndexRepository(t *testing.T) {
	server, repoDir, cleanup := setupTestServer(t)
	defer cleanup()

	body := fmt.Sprintf(`{"name":"index_repository","arguments":{"path":"%s","force":true}}`, repoDir)
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("expected content array in tool result")
	}
}

func TestMCPToolsCallIndexMissingPath(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"index_repository","arguments":{}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestMCPToolsCallIndexInvalidPath(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"index_repository","arguments":{"path":"/nonexistent/path"}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500 for invalid path", resp.StatusCode)
	}
}

func TestMCPToolsCallQueryMissingQuestion(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"query_codebase","arguments":{}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestMCPToolsCallQueryWithoutIndex(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"query_codebase","arguments":{"question":"what is main?"}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Query without index should return 500 error
	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500 for query without index", resp.StatusCode)
	}
}

func TestMCPToolsCallQueryWithRepo(t *testing.T) {
	server, repoDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Query with repo path — should auto-index then query
	body := fmt.Sprintf(`{"name":"query_codebase","arguments":{"question":"what functions exist?","repo":"%s"}}`, repoDir)
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200 for query with repo", resp.StatusCode)
	}
}

func TestMCPToolsCallQueryWithInvalidRepo(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"query_codebase","arguments":{"question":"test","repo":"/nonexistent"}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500 for invalid repo", resp.StatusCode)
	}
}

func TestMCPToolsCallUnknownTool(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	body := `{"name":"unknown_tool","arguments":{}}`
	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMCPToolsCallInvalidBody(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader("not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestMCPToolsCallIndexThenQuery(t *testing.T) {
	server, repoDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Step 1: Index
	indexBody := fmt.Sprintf(`{"name":"index_repository","arguments":{"path":"%s"}}`, repoDir)
	resp1, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(indexBody))
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Fatalf("index status = %d", resp1.StatusCode)
	}

	// Step 2: Query (without repo — uses already-indexed data)
	queryBody := `{"name":"query_codebase","arguments":{"question":"what does the Server struct do?"}}`
	resp2, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(queryBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("query status = %d, want 200", resp2.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp2.Body).Decode(&result)
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Error("expected content in query response")
	}
}

func TestMCPToolsCallIndexWithForce(t *testing.T) {
	server, repoDir, cleanup := setupTestServer(t)
	defer cleanup()

	// Index twice: first normal, then with force
	body1 := fmt.Sprintf(`{"name":"index_repository","arguments":{"path":"%s"}}`, repoDir)
	resp1, _ := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body1))
	resp1.Body.Close()

	body2 := fmt.Sprintf(`{"name":"index_repository","arguments":{"path":"%s","force":true}}`, repoDir)
	resp2, err := http.Post(server.URL+"/mcp/tools/call", "application/json", strings.NewReader(body2))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("force reindex status = %d", resp2.StatusCode)
	}
}
