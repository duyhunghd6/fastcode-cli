package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	agent := NewIterativeAgent(client, te, cfg)
	if agent == nil {
		t.Fatal("NewIterativeAgent returned nil")
	}
}

func TestSystemPrompt(t *testing.T) {
	prompt := systemPrompt()
	if prompt == "" {
		t.Error("systemPrompt should not be empty")
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

func TestParseRoundResponseValidJSON(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	response := `{"confidence": 85, "reasoning": "Found the handler", "tool_calls": [{"name": "search_code", "arg": "auth handler"}]}`
	result, err := agent.parseRoundResponse(response, 1)
	if err != nil {
		t.Fatalf("parseRoundResponse error: %v", err)
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
}

func TestParseRoundResponseCodeFence(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	response := "Here is my response:\n```json\n{\"confidence\": 90, \"reasoning\": \"done\"}\n```"
	result, err := agent.parseRoundResponse(response, 1)
	if err != nil {
		t.Fatalf("parseRoundResponse code fence error: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("confidence = %d, want 90", result.Confidence)
	}
}

func TestParseRoundResponseBadJSON(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	// parseRoundResponse gracefully handles bad JSON: returns confidence=90 fallback
	result, err := agent.parseRoundResponse("this is not json at all", 1)
	if err != nil {
		t.Fatalf("parseRoundResponse should not error on bad JSON: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("expected fallback confidence 90, got %d", result.Confidence)
	}
}

func TestBuildRoundPrompt(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	hr := index.NewHybridRetriever(vs, bm)
	te := NewToolExecutor(hr, nil, nil)
	cfg := DefaultAgentConfig()
	agent := NewIterativeAgent(client, te, cfg)

	pq := ProcessQuery("how does auth work?")
	prompt := agent.buildRoundPrompt("how does auth work?", pq, 1)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
}

func TestRetrieveHighConfidence(t *testing.T) {
	// Mock LLM that returns high confidence immediately
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": `{"confidence": 95, "reasoning": "Found everything needed", "tool_calls": [{"name": "search_code", "arg": "main"}]}`}},
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
	cfg.MaxRounds = 2
	agent := NewIterativeAgent(client, te, cfg)

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
	agent := NewIterativeAgent(client, te, cfg)

	pq := ProcessQuery("test")
	result, err := agent.Retrieve("test", pq)
	// Should not crash, returns with error or partial result
	if err != nil && result != nil {
		t.Log("Got error and result, which is acceptable")
	}
	// Just verify no panic
}
