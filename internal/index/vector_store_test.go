package index

import (
	"math"
	"testing"
)

func TestVectorStoreAddAndSearch(t *testing.T) {
	vs := NewVectorStore()
	vs.Add("a", []float32{1, 0, 0})
	vs.Add("b", []float32{0, 1, 0})
	vs.Add("c", []float32{0.9, 0.1, 0})

	results := vs.Search([]float32{1, 0, 0}, 2)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	if results[0].ID != "a" {
		t.Errorf("expected 'a' first, got %s", results[0].ID)
	}
	if results[1].ID != "c" {
		t.Errorf("expected 'c' second, got %s", results[1].ID)
	}
}

func TestVectorStoreEmpty(t *testing.T) {
	vs := NewVectorStore()
	results := vs.Search([]float32{1, 0}, 5)
	if len(results) != 0 {
		t.Error("expected no results from empty store")
	}
}

func TestVectorStoreSearchEmptyQuery(t *testing.T) {
	vs := NewVectorStore()
	vs.Add("a", []float32{1, 0})
	results := vs.Search(nil, 5)
	if len(results) != 0 {
		t.Error("expected no results for nil query")
	}
	results = vs.Search([]float32{}, 5)
	if len(results) != 0 {
		t.Error("expected no results for empty query")
	}
}

func TestVectorStoreSearchTopKExceedsResults(t *testing.T) {
	vs := NewVectorStore()
	vs.Add("a", []float32{1, 0})

	results := vs.Search([]float32{1, 0}, 100)
	if len(results) != 1 {
		t.Errorf("expected 1 result when topK > available, got %d", len(results))
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Same vector → 1.0
	s := cosineSimilarity([]float32{1, 2, 3}, []float32{1, 2, 3})
	if math.Abs(s-1.0) > 0.001 {
		t.Errorf("same vector similarity = %f, want ~1.0", s)
	}

	// Orthogonal → 0.0
	s = cosineSimilarity([]float32{1, 0}, []float32{0, 1})
	if math.Abs(s) > 0.001 {
		t.Errorf("orthogonal similarity = %f, want ~0.0", s)
	}
}

func TestCosineSimilarityLengthMismatch(t *testing.T) {
	s := cosineSimilarity([]float32{1, 0}, []float32{1, 0, 0})
	if s != 0 {
		t.Errorf("length mismatch similarity = %f, want 0", s)
	}
}

func TestCosineSimilarityEmpty(t *testing.T) {
	s := cosineSimilarity([]float32{}, []float32{})
	if s != 0 {
		t.Errorf("empty similarity = %f, want 0", s)
	}
}

func TestCosineSimilarityZeroVector(t *testing.T) {
	s := cosineSimilarity([]float32{0, 0, 0}, []float32{1, 2, 3})
	if s != 0 {
		t.Errorf("zero vector similarity = %f, want 0", s)
	}
}

func TestVectorStoreCount(t *testing.T) {
	vs := NewVectorStore()
	vs.Add("a", []float32{1, 0})
	vs.Add("b", []float32{0, 1})
	if got := vs.Count(); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
	if got := vs.Dimension(); got != 2 {
		t.Errorf("Dimension() = %d, want 2", got)
	}
}

func TestVectorStoreDimensionEmpty(t *testing.T) {
	vs := NewVectorStore()
	if got := vs.Dimension(); got != 0 {
		t.Errorf("Dimension() = %d for empty store, want 0", got)
	}
}

func TestVectorStoreGet(t *testing.T) {
	vs := NewVectorStore()
	vs.Add("a", []float32{1, 2, 3})

	got := vs.Get("a")
	if got == nil {
		t.Fatal("Get(a) returned nil")
	}
	if len(got) != 3 {
		t.Errorf("Get(a) len = %d, want 3", len(got))
	}
}

func TestVectorStoreGetNotFound(t *testing.T) {
	vs := NewVectorStore()
	got := vs.Get("nonexistent")
	if got != nil {
		t.Errorf("Get(nonexistent) should return nil, got %v", got)
	}
}

func TestNewVectorStore(t *testing.T) {
	vs := NewVectorStore()
	if vs == nil {
		t.Fatal("NewVectorStore returned nil")
	}
	if vs.Count() != 0 {
		t.Error("new store should be empty")
	}
}
