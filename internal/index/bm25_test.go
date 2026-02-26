package index

import (
	"testing"
)

func TestBM25AddAndSearch(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "the quick brown fox jumps over the lazy dog")
	bm.AddDocument("d2", "the lazy cat sleeps on the mat")
	bm.AddDocument("d3", "a quick red fox runs through the forest")

	results := bm.Search("quick fox", 3)
	if len(results) == 0 {
		t.Fatal("expected results for 'quick fox'")
	}
	if results[0].ID != "d1" && results[0].ID != "d3" {
		t.Errorf("expected d1 or d3 first, got %s", results[0].ID)
	}
}

func TestBM25EmptyQuery(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "hello world")
	results := bm.Search("", 5)
	if len(results) != 0 {
		t.Error("expected no results for empty query")
	}
}

func TestBM25NoMatch(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "hello world")
	results := bm.Search("xyzzyx", 5)
	if len(results) != 0 {
		t.Error("expected no results for non-matching query")
	}
}

func TestBM25DocCount(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "one")
	bm.AddDocument("d2", "two")
	if got := bm.DocCount(); got != 2 {
		t.Errorf("DocCount() = %d, want 2", got)
	}
}

func TestBM25DefaultParams(t *testing.T) {
	bm := NewBM25(0, 0) // Should use defaults
	if bm.k1 != 1.5 {
		t.Errorf("default k1 = %f, want 1.5", bm.k1)
	}
	if bm.b != 0.75 {
		t.Errorf("default b = %f, want 0.75", bm.b)
	}
}

func TestBM25SearchEmptyIndex(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	results := bm.Search("hello", 5)
	if len(results) != 0 {
		t.Error("expected no results from empty index")
	}
}

func TestBM25SearchTopKGreaterThanResults(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "hello world")

	results := bm.Search("hello", 100)
	if len(results) != 1 {
		t.Errorf("expected 1 result when topK > available, got %d", len(results))
	}
}

func TestBM25SearchSingleDoc(t *testing.T) {
	bm := NewBM25(1.5, 0.75)
	bm.AddDocument("d1", "the quick brown fox")

	results := bm.Search("quick brown", 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "d1" {
		t.Errorf("expected d1, got %s", results[0].ID)
	}
	if results[0].Score <= 0 {
		t.Errorf("score should be > 0, got %f", results[0].Score)
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("func ParseFile(path string) *Result")
	expected := []string{"func", "parsefile", "path", "string", "result"}
	if len(tokens) != len(expected) {
		t.Errorf("tokenize: got %d tokens %v, want %d: %v", len(tokens), tokens, len(expected), expected)
	}

	// Test underscore splitting
	tokens2 := tokenize("build_graph_call")
	expected2 := []string{"build", "graph", "call"}
	if len(tokens2) != len(expected2) {
		t.Errorf("tokenize snake_case: got %v, want %v", tokens2, expected2)
	}
}

func TestTokenizeEmpty(t *testing.T) {
	tokens := tokenize("")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty, got %d", len(tokens))
	}
}

func TestTokenizeShortTokens(t *testing.T) {
	tokens := tokenize("a b c x y z")
	if len(tokens) != 0 {
		t.Errorf("single-char tokens should be filtered, got %v", tokens)
	}
}

func TestTokenizeSpecialChars(t *testing.T) {
	tokens := tokenize("hello-world foo.bar baz123")
	if len(tokens) < 3 {
		t.Errorf("expected tokens from special chars, got %v", tokens)
	}
}
