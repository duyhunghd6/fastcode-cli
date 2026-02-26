package index

import (
	"math"
	"sort"
)

// VectorStore is an in-memory vector store for embedding-based similarity search.
type VectorStore struct {
	vectors map[string][]float32 // elementID → embedding vector
	dim     int
}

// NewVectorStore creates a new empty vector store.
func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors: make(map[string][]float32),
	}
}

// Add stores an embedding vector for the given element ID.
func (vs *VectorStore) Add(id string, vector []float32) {
	vs.vectors[id] = vector
	if vs.dim == 0 && len(vector) > 0 {
		vs.dim = len(vector)
	}
}

// VectorResult holds a similarity search result.
type VectorResult struct {
	ID    string
	Score float64
}

// Search finds the top-k most similar vectors to the query vector.
func (vs *VectorStore) Search(queryVec []float32, topK int) []VectorResult {
	if len(vs.vectors) == 0 || len(queryVec) == 0 {
		return nil
	}

	type scored struct {
		id    string
		score float64
	}
	var results []scored

	for id, vec := range vs.vectors {
		sim := cosineSimilarity(queryVec, vec)
		if sim > 0 {
			results = append(results, scored{id: id, score: sim})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if topK > len(results) {
		topK = len(results)
	}

	out := make([]VectorResult, topK)
	for i := 0; i < topK; i++ {
		out[i] = VectorResult{
			ID:    results[i].id,
			Score: results[i].score,
		}
	}
	return out
}

// Count returns the number of stored vectors.
func (vs *VectorStore) Count() int {
	return len(vs.vectors)
}

// Dimension returns the dimension of stored vectors.
func (vs *VectorStore) Dimension() int {
	return vs.dim
}

// Get returns the stored vector for an ID, or nil.
func (vs *VectorStore) Get(id string) []float32 {
	return vs.vectors[id]
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
