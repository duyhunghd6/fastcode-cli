package index

import (
	"sort"

	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// HybridRetriever combines vector semantic search and BM25 keyword search.
type HybridRetriever struct {
	vectorStore *VectorStore
	bm25        *BM25
	elements    map[string]*types.CodeElement // ID → element

	// Weights for combining scores
	SemanticWeight float64
	KeywordWeight  float64
}

// HybridResult holds a combined search result.
type HybridResult struct {
	Element *types.CodeElement
	Score   float64
	Source  string // "semantic", "keyword", or "hybrid"
}

// NewHybridRetriever creates a new hybrid retriever.
func NewHybridRetriever(vs *VectorStore, bm25 *BM25) *HybridRetriever {
	return &HybridRetriever{
		vectorStore:    vs,
		bm25:           bm25,
		elements:       make(map[string]*types.CodeElement),
		SemanticWeight: 0.6,
		KeywordWeight:  0.4,
	}
}

// IndexElements indexes code elements into both BM25 and vector stores.
// embedder may be nil if embeddings are not available.
func (hr *HybridRetriever) IndexElements(elements []types.CodeElement, embedder *llm.Embedder) error {
	// Store element references
	for i := range elements {
		elem := &elements[i]
		hr.elements[elem.ID] = elem

		// Add to BM25
		searchText := llm.BuildSearchText(elem.Name, elem.Docstring, elem.Signature, elem.Code)
		hr.bm25.AddDocument(elem.ID, searchText)
	}

	// Generate and store embeddings if embedder is available
	if embedder != nil {
		texts := make([]string, len(elements))
		for i, elem := range elements {
			texts[i] = llm.BuildSearchText(elem.Name, elem.Docstring, elem.Signature, elem.Code)
		}

		embeddings, err := embedder.EmbedTexts(texts)
		if err != nil {
			// Non-fatal: continue without vector search
			return err
		}

		for i, emb := range embeddings {
			if emb != nil {
				hr.vectorStore.Add(elements[i].ID, emb)
			}
		}
	}

	return nil
}

// Search performs hybrid search combining semantic and keyword results.
func (hr *HybridRetriever) Search(query string, queryVec []float32, topK int) []HybridResult {
	scores := make(map[string]float64)

	// BM25 keyword search
	bm25Results := hr.bm25.Search(query, topK*2)
	maxBM25 := 0.0
	for _, r := range bm25Results {
		if r.Score > maxBM25 {
			maxBM25 = r.Score
		}
	}
	for _, r := range bm25Results {
		normalized := 0.0
		if maxBM25 > 0 {
			normalized = r.Score / maxBM25
		}
		scores[r.ID] += normalized * hr.KeywordWeight
	}

	// Vector semantic search
	if queryVec != nil && hr.vectorStore.Count() > 0 {
		vecResults := hr.vectorStore.Search(queryVec, topK*2)
		for _, r := range vecResults {
			scores[r.ID] += r.Score * hr.SemanticWeight
		}
	}

	// Sort by combined score
	type scored struct {
		id    string
		score float64
	}
	var sorted_ []scored
	for id, s := range scores {
		sorted_ = append(sorted_, scored{id: id, score: s})
	}
	sort.Slice(sorted_, func(i, j int) bool {
		return sorted_[i].score > sorted_[j].score
	})

	if topK > len(sorted_) {
		topK = len(sorted_)
	}

	results := make([]HybridResult, topK)
	for i := 0; i < topK; i++ {
		elem := hr.elements[sorted_[i].id]
		source := "hybrid"
		results[i] = HybridResult{
			Element: elem,
			Score:   sorted_[i].score,
			Source:  source,
		}
	}
	return results
}

// ElementCount returns the total number of indexed elements.
func (hr *HybridRetriever) ElementCount() int {
	return len(hr.elements)
}
