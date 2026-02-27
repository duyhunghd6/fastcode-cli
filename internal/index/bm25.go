package index

import (
	"math"
	"sort"
	"strings"
)

// BM25 implements the Okapi BM25 ranking algorithm, matching python's rank_bm25 exactly.
type BM25 struct {
	k1         float64
	b          float64
	epsilon    float64
	docs       []bm25Doc
	df         map[string]int // document frequency per term
	idf        map[string]float64
	avgDL      float64
	averageIdf float64
	totalDocs  int
}

type bm25Doc struct {
	ID     string
	Tokens []string
	Length int
	TF     map[string]float64
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
		k1:      k1,
		b:       b,
		epsilon: 0.25, // Python's BM25Okapi default epsilon
		df:      make(map[string]int),
		idf:     make(map[string]float64),
	}
}

// AddDocument adds a document to the BM25 index.
func (bm *BM25) AddDocument(id, text string) {
	tokens := tokenize(text)
	tf := make(map[string]float64)
	for _, t := range tokens {
		tf[t]++
	}

	doc := bm25Doc{
		ID:     id,
		Tokens: tokens,
		Length: len(tokens),
		TF:     tf,
	}
	bm.docs = append(bm.docs, doc)

	// Update DF
	seen := make(map[string]bool)
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			bm.df[t]++
		}
	}

	bm.totalDocs++
	// Recalculate avgDL
	totalLen := 0
	for _, d := range bm.docs {
		totalLen += d.Length
	}
	bm.avgDL = float64(totalLen) / float64(bm.totalDocs)

	bm.calcIDF()
}

// calcIDF recalculates the IDF for all terms in df exactly like python's rank_bm25
func (bm *BM25) calcIDF() {
	var idfSum float64
	var negativeIdfs []string

	for word, freq := range bm.df {
		// Python: math.log(self.corpus_size - freq + 0.5) - math.log(freq + 0.5)
		idf := math.Log(float64(bm.totalDocs)-float64(freq)+0.5) - math.Log(float64(freq)+0.5)
		bm.idf[word] = idf
		idfSum += idf
		if idf < 0 {
			negativeIdfs = append(negativeIdfs, word)
		}
	}

	if len(bm.idf) > 0 {
		bm.averageIdf = idfSum / float64(len(bm.idf))
	} else {
		bm.averageIdf = 0
	}

	eps := bm.epsilon * bm.averageIdf
	for _, word := range negativeIdfs {
		bm.idf[word] = eps
	}
}

// BM25Result holds a scored document ID.
type BM25Result struct {
	ID    string
	Score float64
}

type scored struct {
	idx   int
	score float64
}

// Search returns the top-k documents for a query text.
func (bm *BM25) Search(query string, topK int) []BM25Result {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 || bm.totalDocs == 0 {
		return nil
	}

	var results []scored
	for i, doc := range bm.docs {
		var score float64

		for _, token := range queryTokens {
			termFreq := doc.TF[token]
			if termFreq == 0 {
				continue
			}

			idf := bm.idf[token]
			// Python's TF normalization implementation
			tfNorm := (termFreq * (bm.k1 + 1)) / (termFreq + bm.k1*(1-bm.b+bm.b*float64(doc.Length)/bm.avgDL))

			score += idf * tfNorm
		}

		if score > 0 {
			results = append(results, scored{idx: i, score: score})
		}
	}

	// Sort by score descending. For ties, Python rank_bm25 preserves original order (mostly).
	// To exactly mirror python, we use a stable sort and tie-break on index.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score == results[j].score {
			return results[i].idx < results[j].idx
		}
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
