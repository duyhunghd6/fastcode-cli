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

// === Additional Tools Tests ===

func TestSearchCodeWithEmbedder(t *testing.T) {
	// Mock embedding server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"index": 0, "embedding": []float64{0.9, 0.1, 0.0}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	embedder := llm.NewEmbedder(client, "model", 32)

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "handleAuth", Type: "function", Code: "func handleAuth() {}"},
	}
	_ = hr.IndexElements(elements, nil)
	vs.Add("e1", []float32{1.0, 0.0, 0.0})

	te := NewToolExecutor(hr, embedder, elements)
	// search_code still works via backward compat in Execute()
	result, err := te.Execute("search_code", "auth handler")
	if err != nil {
		t.Fatalf("search_code with embedder: %v", err)
	}
	if result.ToolName != "search_codebase" {
		t.Errorf("ToolName = %q, want search_codebase", result.ToolName)
	}
}

func TestSearchCodeWithEmbedderError(t *testing.T) {
	// Mock embedding server that errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"fail"}}`))
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	embedder := llm.NewEmbedder(client, "model", 32)

	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "foo", Type: "function", Code: "func foo() {}"},
	}
	_ = hr.IndexElements(elements, nil)

	te := NewToolExecutor(hr, embedder, elements)
	// Should not error — embedder error is handled gracefully
	result, err := te.Execute("search_code", "foo")
	if err != nil {
		t.Fatalf("search_code with embedder error: %v", err)
	}
	if result.ToolName != "search_codebase" {
		t.Errorf("ToolName = %q, want search_codebase", result.ToolName)
	}
}

func TestBrowseFileSuffixMatch(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "internal/parser/go_parser.go", Code: "package parser"},
	}
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	// Match via suffix
	result, err := te.Execute("browse_file", "go_parser.go")
	if err != nil {
		t.Fatalf("browse_file suffix: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Errorf("expected 1 element via suffix match, got %d", len(result.Elements))
	}
}

func TestListDirectoryCaseInsensitive(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "f1", Type: "file", RelativePath: "Internal/Parser/GoParser.go"},
	}
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, elements)

	// list_directory (also aliased from list_files)
	result, err := te.Execute("list_directory", "parser")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Elements) != 1 {
		t.Errorf("expected case-insensitive match, got %d", len(result.Elements))
	}
}

func TestSearchGraph(t *testing.T) {
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "playAudio", Type: "function", Code: "func playAudio() {}"},
	}
	_ = hr.IndexElements(elements, nil)
	te := NewToolExecutor(hr, nil, elements)

	// search_graph is now implemented as a stub that falls back to search_codebase
	result, err := te.Execute("search_graph", "audio")
	if err != nil {
		t.Fatalf("search_graph should not error: %v", err)
	}
	if result.ToolName != "search_codebase" {
		t.Errorf("ToolName = %q, want search_codebase (fallback)", result.ToolName)
	}
}

// === Additional Query Tests ===

func TestExtractKeywords(t *testing.T) {
	keywords := extractKeywords("how does the authentication work in handleAuth?")
	if len(keywords) == 0 {
		t.Error("expected keywords")
	}
	// Stop words should be filtered
	for _, kw := range keywords {
		if kw == "how" || kw == "does" || kw == "the" || kw == "in" {
			t.Errorf("stop word %q should be filtered", kw)
		}
	}
}

func TestExtractKeywordsDuplication(t *testing.T) {
	keywords := extractKeywords("test test test unique")
	count := 0
	for _, kw := range keywords {
		if kw == "test" {
			count++
		}
	}
	if count > 1 {
		t.Error("duplicate keywords should be filtered")
	}
}

func TestExtractKeywordsShortWords(t *testing.T) {
	keywords := extractKeywords("a b c x")
	if len(keywords) != 0 {
		t.Errorf("single-char words should be filtered, got %v", keywords)
	}
}

func TestScoreComplexityLow(t *testing.T) {
	pq := ProcessQuery("find main")
	if pq.Complexity > 20 {
		t.Errorf("simple query complexity = %d, expected <= 20", pq.Complexity)
	}
}

func TestScoreComplexityMedium(t *testing.T) {
	pq := ProcessQuery("how does the application handle authentication and authorization between services?")
	if pq.Complexity < 30 {
		t.Errorf("medium query complexity = %d, expected >= 30", pq.Complexity)
	}
}

func TestScoreComplexityHigh(t *testing.T) {
	pq := ProcessQuery("explain the architecture and design pattern for the concurrent pipeline and how does the inheritance relationship between base handler and child handlers work? also compare the algorithm complexity")
	if pq.Complexity < 50 {
		t.Errorf("high complexity = %d, expected >= 50", pq.Complexity)
	}
}

func TestScoreComplexityCapsAt100(t *testing.T) {
	// Very very complex query with all indicators
	query := "explain the architecture and design pattern for the concurrent thread and async pipeline and how does the inheritance relationship and dependency flow between all components interact? also compare the algorithm complexity?"
	pq := ProcessQuery(query)
	if pq.Complexity > 100 {
		t.Errorf("complexity should cap at 100, got %d", pq.Complexity)
	}
}

func TestClassifyQueryLocate(t *testing.T) {
	tests := []string{"where is the main function?", "find the config", "locate the handler"}
	for _, q := range tests {
		pq := ProcessQuery(q)
		if pq.QueryType != "locate" {
			t.Errorf("ProcessQuery(%q) type = %q, want locate", q, pq.QueryType)
		}
	}
}

func TestClassifyQueryDebug(t *testing.T) {
	tests := []string{"there's a bug in auth", "how to fix this error", "something is wrong with the cache"}
	for _, q := range tests {
		pq := ProcessQuery(q)
		if pq.QueryType != "debug" {
			t.Errorf("ProcessQuery(%q) type = %q, want debug", q, pq.QueryType)
		}
	}
}

func TestClassifyQueryHowto(t *testing.T) {
	tests := []string{"how to add a new handler", "how do I implement caching", "implement a retry mechanism"}
	for _, q := range tests {
		pq := ProcessQuery(q)
		if pq.QueryType != "howto" {
			t.Errorf("ProcessQuery(%q) type = %q, want howto", q, pq.QueryType)
		}
	}
}

func TestClassifyQueryOverview(t *testing.T) {
	tests := []string{"project overview", "what is the architecture", "explain the structure"}
	for _, q := range tests {
		pq := ProcessQuery(q)
		if pq.QueryType != "overview" {
			t.Errorf("ProcessQuery(%q) type = %q, want overview", q, pq.QueryType)
		}
	}
}

func TestClassifyQueryUnderstand(t *testing.T) {
	pq := ProcessQuery("explain auth token validation")
	if pq.QueryType != "understand" {
		t.Errorf("ProcessQuery(explain...) type = %q, want understand", pq.QueryType)
	}
}

// === Additional Iterative Agent Tests ===

func TestBuildRoundNPromptWithGatheredElements(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	// Init adaptive params
	agent.initializeAdaptiveParams(50)

	// Set gathered elements to cover the context section
	agent.gatheredElements = []types.CodeElement{
		{Type: "function", Name: "handleAuth", RelativePath: "auth.go", StartLine: 10, EndLine: 20, Signature: "func handleAuth()"},
		{Type: "class", Name: "Server", RelativePath: "server.go", StartLine: 1, EndLine: 50},
	}

	pq := ProcessQuery("how does auth work?")
	prompt := agent.buildRoundNPrompt("how does auth work?", pq, 2)

	if !strings.Contains(prompt, "Retrieved Elements") {
		t.Error("prompt should contain retrieved elements section")
	}
	if !strings.Contains(prompt, "handleAuth") {
		t.Error("prompt should reference gathered element names")
	}
	if !strings.Contains(prompt, "keep_files") {
		t.Error("prompt should mention keep_files")
	}
}

func TestBuildRoundNPromptManyGatheredElements(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	// Init adaptive params
	agent.initializeAdaptiveParams(50)

	// Set 25 gathered elements to trigger truncation at 20
	for i := 0; i < 25; i++ {
		agent.gatheredElements = append(agent.gatheredElements, types.CodeElement{
			Type: "function", Name: "func" + string(rune('A'+i)),
			RelativePath: "file.go", StartLine: 1, EndLine: 5,
		})
	}

	pq := ProcessQuery("overview")
	prompt := agent.buildRoundNPrompt("overview", pq, 2)

	if !strings.Contains(prompt, "more elements") {
		t.Error("prompt should indicate truncated elements when > 20")
	}
}

func TestNewIterativeAgentWithZeroConfig(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	// Zero config should use defaults
	agent := NewIterativeAgent(client, te, AgentConfig{})
	if agent.config.MaxRounds != 4 {
		t.Errorf("zero config MaxRounds = %d, want 4 (default)", agent.config.MaxRounds)
	}
}

func TestRetrieveNoMoreActions(t *testing.T) {
	// Mock LLM that returns no tool calls on round 1 and no_more_actions on round 2
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var content string
		if callCount <= 1 {
			// Round 1: low confidence, no tool calls
			content = `{"confidence": 40, "query_complexity": 30, "reasoning": "low confidence", "tool_calls": []}`
		} else {
			// Round 2: low confidence, no tool calls → triggers no_more_actions
			content = `{"confidence": 40, "reasoning": "still low", "keep_files": [], "tool_calls": []}`
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{
					"role":    "assistant",
					"content": content,
				}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	cfg := DefaultAgentConfig()
	cfg.MaxRounds = 4
	agent := NewIterativeAgent(client, te, cfg)

	pq := ProcessQuery("test")
	result, err := agent.Retrieve("test", pq)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if result.StopReason != "no_more_actions" {
		t.Errorf("StopReason = %q, want no_more_actions", result.StopReason)
	}
}

func TestRetrieveLowComplexityFewRounds(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var content string
		if callCount <= 1 {
			content = `{"confidence": 60, "query_complexity": 15, "reasoning": "simple", "tool_calls": []}`
		} else {
			content = `{"confidence": 97, "reasoning": "done", "keep_files": []}`
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": content}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)

	cfg := DefaultAgentConfig()
	cfg.MaxRounds = 5
	agent := NewIterativeAgent(client, te, cfg)

	// Simple query → complexity < 30 → maxRounds capped at 2
	pq := &ProcessedQuery{Original: "find main", Cleaned: "find main", Complexity: 15, QueryType: "locate", Keywords: []string{"main"}}
	result, err := agent.Retrieve("find main", pq)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if result.Rounds > 3 {
		t.Errorf("low complexity should limit rounds, got %d", result.Rounds)
	}
}

func TestRetrieveToolCallExecution(t *testing.T) {
	roundCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roundCount++
		var content string
		if roundCount == 1 {
			// Round 1: assessment with tool calls
			content = `{"confidence": 50, "query_complexity": 50, "reasoning": "need more", "tool_calls": [{"tool": "search_codebase", "parameters": {"search_term": "main"}}]}`
		} else {
			// Round 2: high confidence with keep_files
			content = `{"confidence": 97, "reasoning": "found", "keep_files": ["main.go"]}`
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": content}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("key", "model", server.URL)
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)

	elements := []types.CodeElement{
		{ID: "e1", Name: "main", Type: "function", Code: "func main() {}"},
	}
	_ = hr.IndexElements(elements, nil)
	te := NewToolExecutor(hr, nil, elements)

	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	pq := &ProcessedQuery{Original: "find main", Cleaned: "find main", Complexity: 50, QueryType: "locate", Keywords: []string{"main"}}
	result, err := agent.Retrieve("find main", pq)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if result.StopReason != "confidence_threshold_reached" {
		t.Errorf("StopReason = %q, want confidence_threshold_reached", result.StopReason)
	}
	// Elements may or may not be present depending on keep_files filtering
	// The important thing is we reached high confidence
	if result.Confidence < 90 {
		t.Errorf("expected high confidence, got %d", result.Confidence)
	}
}

func TestDeduplicateElements(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "e1", Name: "foo"},
		{ID: "e2", Name: "bar"},
		{ID: "e1", Name: "foo"}, // duplicate
		{ID: "e3", Name: "baz"},
		{ID: "e2", Name: "bar"}, // duplicate
	}
	deduped := deduplicateElements(elements)
	if len(deduped) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(deduped))
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"key": "value"}`, `{"key": "value"}`},
		{"some text ```json\n{\"key\": \"value\"}\n``` after", `{"key": "value"}`},
		{"no json here", ""},
		{`text {"nested": {"a": 1}} end`, `{"nested": {"a": 1}}`},
	}
	for _, tt := range tests {
		got := extractJSON(tt.input)
		if got != tt.want {
			t.Errorf("extractJSON(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractJSONUnterminatedBrace(t *testing.T) {
	got := extractJSON(`{"key": "value"`)
	if got != "" {
		t.Errorf("unterminated brace should return empty, got %q", got)
	}
}

func TestToolCallGetToolName(t *testing.T) {
	// Test "tool" field (Python-style)
	tc := ToolCall{Tool: "search_codebase", Parameters: map[string]any{"search_term": "main"}}
	if tc.GetToolName() != "search_codebase" {
		t.Errorf("GetToolName = %q, want search_codebase", tc.GetToolName())
	}

	// Test "name" field (Go-style)
	tc2 := ToolCall{Name: "search_code", Arg: "main"}
	if tc2.GetToolName() != "search_code" {
		t.Errorf("GetToolName = %q, want search_code", tc2.GetToolName())
	}

	// Test GetArg from parameters
	tc3 := ToolCall{Tool: "search_codebase", Parameters: map[string]any{"search_term": "audio"}}
	if tc3.GetArg() != "audio" {
		t.Errorf("GetArg = %q, want audio", tc3.GetArg())
	}
}
