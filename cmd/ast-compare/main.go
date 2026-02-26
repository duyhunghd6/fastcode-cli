package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/duyhunghd6/fastcode-cli/internal/graph"
	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/loader"
)

func getStats(path string) (map[string]any, error) {
	config := loader.DefaultConfig()
	repo, err := loader.LoadRepository(path, config)
	if err != nil {
		return nil, err
	}

	idx := index.NewIndexer(repo.Name)
	elements, err := idx.IndexRepository(repo)
	if err != nil {
		return nil, err
	}

	g := graph.NewCodeGraphs()
	g.BuildGraphs(elements)

	counts := make(map[string]int)
	totalComplexity := 0
	functionsCount := 0

	for _, e := range elements {
		counts[e.Type]++

		// Attempt to extract complexity if available in Metadata
		if comp, ok := e.Metadata["complexity"].(float64); ok {
			totalComplexity += int(comp)
			if e.Type == "function" || e.Type == "method" {
				functionsCount++
			}
		} else if comp, ok := e.Metadata["complexity"].(int); ok {
			totalComplexity += comp
			if e.Type == "function" || e.Type == "method" {
				functionsCount++
			}
		}
	}

	avgComplexity := 0.0
	if functionsCount > 0 {
		avgComplexity = float64(totalComplexity) / float64(functionsCount)
	}

	return map[string]any{
		"files":          len(repo.Files),
		"total_elements": len(elements),
		"element_types":  counts,
		"avg_complexity": avgComplexity,
		"graph_stats":    g.Stats(),
	}, nil
}

func main() {
	pStats, err := getStats("/Users/steve/duyhunghd6/gmind/reference/FastCode/fastcode")
	if err != nil {
		fmt.Printf("Error extracting Python AST: %v\n", err)
		os.Exit(1)
	}

	gStats, err := getStats("/Users/steve/duyhunghd6/fastcode-cli")
	if err != nil {
		fmt.Printf("Error extracting Go AST: %v\n", err)
		os.Exit(1)
	}

	res := map[string]any{
		"Python_AST": pStats,
		"Go_AST":     gStats,
	}

	out, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(out))
}
