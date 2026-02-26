package agent

import (
	"testing"
)

func TestProcessQuery(t *testing.T) {
	pq := ProcessQuery("How does the authentication flow work in the API?")

	if pq.Original == "" {
		t.Error("Original should not be empty")
	}
	if len(pq.Keywords) == 0 {
		t.Error("Keywords should not be empty")
	}
	if pq.Complexity == 0 {
		t.Error("Complexity should be > 0")
	}
	if pq.QueryType == "" {
		t.Error("QueryType should not be empty")
	}
}

func TestClassifyQuery(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"Where is the payment handler defined?", "locate"},
		{"Explain the authentication flow", "understand"},
		{"How to implement a new endpoint?", "howto"},
		{"There's a bug in the login function", "debug"},
		{"Give me an overview of the architecture", "overview"},
	}
	for _, tt := range tests {
		got := classifyQuery(tt.query)
		if got != tt.want {
			t.Errorf("classifyQuery(%q) = %q, want %q", tt.query, got, tt.want)
		}
	}
}

func TestScoreComplexity(t *testing.T) {
	// Simple query
	simple := scoreComplexity("where is main?", []string{"main"})
	// Complex query
	complex := scoreComplexity(
		"How do the authentication and authorization modules interact with the database layer and what design patterns are used?",
		[]string{"authentication", "authorization", "modules", "interact", "database", "layer", "design", "patterns"},
	)
	if complex <= simple {
		t.Errorf("complex (%d) should score higher than simple (%d)", complex, simple)
	}
}
