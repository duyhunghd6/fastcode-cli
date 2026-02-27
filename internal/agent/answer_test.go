package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestNewAnswerGenerator(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	ag := NewAnswerGenerator(client)
	if ag == nil {
		t.Fatal("NewAnswerGenerator returned nil")
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"ab", 2, "ab"},
		{"abc", 2, "ab..."},
	}
	for _, tt := range tests {
		got := truncateStr(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestAnswerSystemPrompt(t *testing.T) {
	prompt := answerSystemPrompt()
	if prompt == "" {
		t.Error("answerSystemPrompt should not be empty")
	}
	if !strings.Contains(prompt, "code understanding") {
		t.Error("system prompt should mention code understanding")
	}
}

func TestBuildPromptNoElements(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	ag := NewAnswerGenerator(client)
	pq := ProcessQuery("test query")

	result := ag.buildPrompt("test query", pq, nil)
	if !strings.Contains(result, "test query") {
		t.Error("prompt should contain the query")
	}
	// With no elements, the prompt should still contain the query
	if !strings.Contains(result, "Instructions") {
		t.Error("prompt should contain instructions")
	}
}

func TestBuildPromptWithElements(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	ag := NewAnswerGenerator(client)
	pq := ProcessQuery("find auth handler")

	elements := []types.CodeElement{
		{
			Type: "function", Name: "handleAuth", RelativePath: "auth.go",
			StartLine: 10, EndLine: 20, Language: "go",
			Signature: "func handleAuth()", Docstring: "Handles auth",
			Code: "func handleAuth() { }",
		},
	}

	result := ag.buildPrompt("find auth handler", pq, elements)
	if !strings.Contains(result, "handleAuth") {
		t.Error("prompt should contain element name")
	}
	if !strings.Contains(result, "auth.go") {
		t.Error("prompt should contain file path")
	}
	if !strings.Contains(result, "func handleAuth()") {
		t.Error("prompt should contain signature")
	}
}

func TestBuildPromptManyElements(t *testing.T) {
	client := llm.NewClientWith("key", "model", "http://localhost")
	ag := NewAnswerGenerator(client)
	pq := ProcessQuery("overview")

	// Create 20 elements to trigger truncation at 15
	var elements []types.CodeElement
	for i := 0; i < 20; i++ {
		elements = append(elements, types.CodeElement{
			Type: "function", Name: "func" + string(rune('A'+i)),
			RelativePath: "file.go", Language: "go", StartLine: 1, EndLine: 5,
		})
	}

	result := ag.buildPrompt("overview", pq, elements)
	// 20 elements: only first 15 are included in the prompt
	if !strings.Contains(result, "funcA") {
		t.Error("prompt should contain first element")
	}
}

func TestGenerateAnswer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "The auth handler is in auth.go"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClientWith("test-key", "test-model", server.URL)
	ag := NewAnswerGenerator(client)
	pq := ProcessQuery("where is auth?")

	elements := []types.CodeElement{
		{Type: "function", Name: "handleAuth", RelativePath: "auth.go", Language: "go"},
	}

	answer, err := ag.GenerateAnswer("where is auth?", pq, elements)
	if err != nil {
		t.Fatalf("GenerateAnswer error: %v", err)
	}
	if answer == "" {
		t.Error("answer should not be empty")
	}
}

func TestGenerateAnswerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer server.Close()

	client := llm.NewClientWith("test-key", "test-model", server.URL)
	ag := NewAnswerGenerator(client)
	pq := ProcessQuery("test")

	_, err := ag.GenerateAnswer("test", pq, nil)
	if err == nil {
		t.Error("expected error from failed LLM call")
	}
}
