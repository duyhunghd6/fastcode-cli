package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestDefaultAgentConfig(t *testing.T) {
	cfg := DefaultAgentConfig()
	if cfg.MaxRounds <= 0 {
		t.Error("MaxRounds should be > 0")
	}
	if cfg.ConfidenceThreshold <= 0 {
		t.Error("ConfidenceThreshold should be > 0")
	}
	if cfg.MaxTokenBudget <= 0 {
		t.Error("MaxTokenBudget should be > 0")
	}
	if cfg.Temperature <= 0 {
		t.Error("Temperature should be > 0")
	}
}

func TestNewIterativeAgent(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()

	agent := NewIterativeAgent(client, te, nil, cfg)
	if agent == nil {
		t.Fatal("NewIterativeAgent returned nil")
	}
}

func TestMinFunc(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3,5) should be 3")
	}
	if min(5, 3) != 3 {
		t.Error("min(5,3) should be 3")
	}
	if min(4, 4) != 4 {
		t.Error("min(4,4) should be 4")
	}
}

func TestParseRound1ResponseValidJSON(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, nil, cfg)

	response := `{"confidence": 85, "reasoning": "Found the handler", "query_complexity": 45, "tool_calls": [{"tool": "search_codebase", "parameters": {"search_term": "auth handler"}}]}`
	result, err := agent.parseRound1Response(response)
	if err != nil {
		t.Fatalf("parseRound1Response error: %v", err)
	}
	if result.Confidence != 85 {
		t.Errorf("confidence = %d, want 85", result.Confidence)
	}
	if result.Reasoning != "Found the handler" {
		t.Errorf("reasoning = %q", result.Reasoning)
	}
	if len(result.ToolCalls) != 1 {
		t.Errorf("tool_calls = %d, want 1", len(result.ToolCalls))
	}
	if result.QueryComplexity != 45 {
		t.Errorf("query_complexity = %d, want 45", result.QueryComplexity)
	}
}

func TestParseRound1ResponseCodeFence(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, nil, cfg)

	response := "Here is my response:\n```json\n{\"confidence\": 90, \"reasoning\": \"done\"}\n```"
	result, err := agent.parseRound1Response(response)
	if err != nil {
		t.Fatalf("parseRound1Response code fence error: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("confidence = %d, want 90", result.Confidence)
	}
}

func TestParseRound1ResponseBadJSON(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, nil, cfg)

	result, err := agent.parseRound1Response("this is not json at all")
	if err != nil {
		t.Fatalf("parseRound1Response should not error on bad JSON: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("expected fallback confidence 90, got %d", result.Confidence)
	}
}

func TestBuildRound1Prompt(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, nil, cfg)

	pq := ProcessQuery("how does auth work?")
	prompt := agent.buildRound1Prompt("how does auth work?", pq)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
	if !strings.Contains(prompt, "how does auth work?") {
		t.Error("prompt should contain the query")
	}
	if !strings.Contains(prompt, "search_codebase") {
		t.Error("prompt should mention search_codebase tool")
	}
	if !strings.Contains(prompt, "list_directory") {
		t.Error("prompt should mention list_directory tool")
	}
}

func TestBuildRoundNPrompt(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, nil, cfg)

	// Init adaptive params for proper budget display
	agent.initializeAdaptiveParams(50)

	// Set gathered elements
	agent.gatheredElements = []types.CodeElement{
		{Type: "function", Name: "handleAuth", RelativePath: "auth.go", StartLine: 10, EndLine: 20, Signature: "func handleAuth()"},
		{Type: "class", Name: "Server", RelativePath: "server.go", StartLine: 1, EndLine: 50},
	}

	pq := ProcessQuery("how does auth work?")
	prompt := agent.buildRoundNPrompt("how does auth work?", pq, 2)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
	if !strings.Contains(prompt, "cost-aware") {
		t.Error("round N prompt should contain cost-aware instructions")
	}
	if !strings.Contains(prompt, "keep_files") {
		t.Error("round N prompt should mention keep_files")
	}
	if !strings.Contains(prompt, "handleAuth") {
		t.Error("prompt should reference gathered element names")
	}
}

func TestRetrieveHighConfidence(t *testing.T) {
	// Mock LLM that returns high confidence
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var content string
		if callCount <= 1 {
			// Round 1: assessment with tool calls
			content = `{"confidence": 60, "query_complexity": 30, "reasoning": "need to search", "tool_calls": [{"tool": "search_codebase", "parameters": {"search_term": "main"}}]}`
		} else {
			// Round 2+: keep_files with high confidence
			content = `{"confidence": 97, "reasoning": "Found everything needed", "keep_files": ["main.go"]}`
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": content}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("test-key", "test-model", server.URL)
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "main", Type: "function", Code: "func main() {}"},
	}
	_ = hr.IndexElements(elements, nil)
	te := NewToolExecutor(hr, nil, elements)

	cfg := DefaultAgentConfig()
	cfg.MaxRounds = 3
	agent := NewIterativeAgent(client, te, nil, cfg)

	pq := ProcessQuery("where is main?")
	result, err := agent.Retrieve("where is main?", pq)
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	if result.Confidence < 80 {
		t.Errorf("confidence = %d, expected >= 80", result.Confidence)
	}
	if result.StopReason == "" {
		t.Error("stop reason should not be empty")
	}
}

func TestRetrieveLLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"server down"}}`))
	}))
	defer server.Close()

	client := llm.NewClientWith("test-key", "test-model", server.URL)
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	cfg := DefaultAgentConfig()
	cfg.MaxRounds = 1
	agent := NewIterativeAgent(client, te, nil, cfg)

	pq := ProcessQuery("test")
	result, err := agent.Retrieve("test", pq)
	// Should not crash, returns with error or partial result
	if err != nil && result != nil {
		t.Log("Got error and result, which is acceptable")
	}
	// Just verify no panic
}
