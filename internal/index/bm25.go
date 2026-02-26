package index

import (
	"math"
	"sort"
	"strings"
)

// BM25 implements the Okapi BM25 ranking algorithm for keyword search.
type BM25 struct {
	k1        float64
	b         float64
	docs      []bm25Doc
	df        map[string]int // document frequency per term
	avgDL     float64
	totalDocs int
}

type bm25Doc struct {
	ID     string
	Tokens []string
	Length int
}

// NewBM25 creates a new BM25 index with standard parameters.
func NewBM25(k1, b float64) *BM25 {
	if k1 == 0 {
		k1 = 1.5
	}
	if b == 0 {
		b = 0.75
	}
	return &BM25{
		k1: k1,
		b:  b,
		df: make(map[string]int),
	}
}

// AddDocument adds a document to the BM25 index.
func (bm *BM25) AddDocument(id, text string) {
	tokens := tokenize(text)
	doc := bm25Doc{
		ID:     id,
		Tokens: tokens,
		Length: len(tokens),
	}
	bm.docs = append(bm.docs, doc)

	// Update document frequency
	seen := make(map[string]bool)
	for _, tok := range tokens {
		if !seen[tok] {
			bm.df[tok]++
			seen[tok] = true
		}
	}

	// Recalculate average document length
	bm.totalDocs = len(bm.docs)
	totalLen := 0
	for _, d := range bm.docs {
		totalLen += d.Length
	}
	bm.avgDL = float64(totalLen) / float64(bm.totalDocs)
}

// BM25Result holds a search result.
type BM25Result struct {
	ID    string
	Score float64
}

// Search returns the top-k documents matching the query, ranked by BM25 score.
func (bm *BM25) Search(query string, topK int) []BM25Result {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 || bm.totalDocs == 0 {
		return nil
	}

	type scored struct {
		idx   int
		score float64
	}

	var results []scored

	for i, doc := range bm.docs {
		score := 0.0
		// Build term frequency map for this document
		tf := make(map[string]int)
		for _, tok := range doc.Tokens {
			tf[tok]++
		}

		for _, qt := range queryTokens {
			docFreq, exists := bm.df[qt]
			if !exists {
				continue
			}
			termFreq := float64(tf[qt])

			// IDF component — BM25+ variant: always positive even for common terms
			idf := math.Log(1 + (float64(bm.totalDocs)-float64(docFreq)+0.5)/(float64(docFreq)+0.5))

			// TF component with length normalization
			tfNorm := (termFreq * (bm.k1 + 1)) /
				(termFreq + bm.k1*(1-bm.b+bm.b*float64(doc.Length)/bm.avgDL))

			score += idf * tfNorm
		}

		if score > 0 {
			results = append(results, scored{idx: i, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Return top-k
	if topK > len(results) {
		topK = len(results)
	}

	out := make([]BM25Result, topK)
	for i := 0; i < topK; i++ {
		out[i] = BM25Result{
			ID:    bm.docs[results[i].idx].ID,
			Score: results[i].score,
		}
	}
	return out
}

// DocCount returns the number of documents in the index.
func (bm *BM25) DocCount() int {
	return bm.totalDocs
}

// tokenize splits text into lowercase tokens, handling camelCase and snake_case.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Split on non-alphanumeric characters
	var raw []string
	var current strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			current.WriteRune(r)
		} else if r == '_' {
			// underscore is a separator
			if current.Len() > 0 {
				raw = append(raw, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				raw = append(raw, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		raw = append(raw, current.String())
	}

	// Filter short tokens
	var tokens []string
	for _, tok := range raw {
		if len(tok) > 1 {
			tokens = append(tokens, tok)
		}
	}
	return tokens
}
