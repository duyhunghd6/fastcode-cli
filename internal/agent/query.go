package agent

import (
	"strings"
	"unicode"
)

// ProcessedQuery holds the analyzed and enriched form of a user query.
type ProcessedQuery struct {
	Original   string   `json:"original"`
	Cleaned    string   `json:"cleaned"`
	Keywords   []string `json:"keywords"`
	Complexity int      `json:"complexity"` // 0-100
	QueryType  string   `json:"query_type"` // "locate", "understand", "debug", "howto", "overview"
}

// ProcessQuery analyzes a user query and extracts keywords, complexity, and type.
func ProcessQuery(query string) *ProcessedQuery {
	pq := &ProcessedQuery{
		Original: query,
		Cleaned:  strings.TrimSpace(query),
	}

	pq.Keywords = extractKeywords(pq.Cleaned)
	pq.Complexity = scoreComplexity(pq.Cleaned, pq.Keywords)
	pq.QueryType = classifyQuery(pq.Cleaned)

	return pq
}

// extractKeywords pulls meaningful terms from the query.
func extractKeywords(query string) []string {
	// Stop words to filter out
	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true,
		"a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "of": true, "to": true, "for": true, "with": true,
		"how": true, "what": true, "where": true, "when": true, "why": true,
		"does": true, "do": true, "this": true, "that": true, "it": true,
		"from": true, "are": true, "was": true, "were": true, "be": true,
		"has": true, "have": true, "had": true, "can": true, "could": true,
		"would": true, "should": true, "will": true, "i": true, "me": true,
		"my": true, "we": true, "our": true, "you": true, "your": true,
	}

	words := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.'
	})

	var keywords []string
	seen := make(map[string]bool)
	for _, w := range words {
		if len(w) < 2 || stopWords[w] || seen[w] {
			continue
		}
		seen[w] = true
		keywords = append(keywords, w)
	}
	return keywords
}

// scoreComplexity rates query complexity from 0-100.
func scoreComplexity(query string, keywords []string) int {
	score := 0

	// Length-based
	words := strings.Fields(query)
	if len(words) > 15 {
		score += 30
	} else if len(words) > 8 {
		score += 20
	} else {
		score += 10
	}

	// Keyword density
	if len(keywords) > 6 {
		score += 20
	} else if len(keywords) > 3 {
		score += 10
	}

	// Multi-concept indicators
	multiIndicators := []string{"and", "also", "both", "between", "compare", "relationship", "interact", "flow"}
	for _, ind := range multiIndicators {
		if strings.Contains(strings.ToLower(query), ind) {
			score += 10
			break
		}
	}

	// Technical depth indicators
	techIndicators := []string{"architecture", "design pattern", "inheritance", "dependency", "concurrency",
		"thread", "async", "lifecycle", "pipeline", "algorithm"}
	for _, ind := range techIndicators {
		if strings.Contains(strings.ToLower(query), ind) {
			score += 15
			break
		}
	}

	// Question complexity
	if strings.Contains(query, "?") {
		score += 5
	}

	if score > 100 {
		score = 100
	}
	return score
}

// classifyQuery determines the query type.
func classifyQuery(query string) string {
	q := strings.ToLower(query)

	switch {
	case strings.Contains(q, "where") || strings.Contains(q, "find") || strings.Contains(q, "locate"):
		return "locate"
	case strings.Contains(q, "bug") || strings.Contains(q, "error") || strings.Contains(q, "fix") || strings.Contains(q, "wrong"):
		return "debug"
	case strings.Contains(q, "how to") || strings.Contains(q, "how do") || strings.Contains(q, "implement"):
		return "howto"
	case strings.Contains(q, "overview") || strings.Contains(q, "architecture") || strings.Contains(q, "structure"):
		return "overview"
	default:
		return "understand"
	}
}
