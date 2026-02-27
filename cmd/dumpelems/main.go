package main

import (
"fmt"
"os"
"sort"

"github.com/duyhunghd6/fastcode-cli/internal/index"
"github.com/duyhunghd6/fastcode-cli/internal/loader"
)

func main() {
	repoPath := os.Args[1]
	cfg := loader.DefaultConfig()
	repo, err := loader.LoadRepository(repoPath, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	indexer := index.NewIndexer(repo.Name)
	elements, err := indexer.IndexRepository(repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Count per-file element breakdown
	fileCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	for _, e := range elements {
		fileCounts[e.RelativePath]++
		typeCounts[e.Type]++
	}

	paths := make([]string, 0, len(fileCounts))
	for p := range fileCounts {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		fmt.Printf("%d\t%s\n", fileCounts[p], p)
	}

	fmt.Fprintf(os.Stderr, "Total elements: %d\n", len(elements))
	for t, c := range typeCounts {
		fmt.Fprintf(os.Stderr, "  %s: %d\n", t, c)
	}
}
