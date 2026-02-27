package main

import (
"fmt"
"os"
"sort"

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
	paths := make([]string, len(repo.Files))
	for i, f := range repo.Files {
		paths[i] = f.RelativePath
	}
	sort.Strings(paths)
	for _, p := range paths {
		fmt.Println(p)
	}
}
