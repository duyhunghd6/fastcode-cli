package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("fastcode-cli %s\n", version)
		return
	}

	fmt.Println("⚡ FastCode-CLI — Codebase Intelligence Engine (Go)")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  fastcode index <repo-path>     Index a local repository")
	fmt.Println("  fastcode query <question>       Query the indexed codebase")
	fmt.Println("  fastcode serve-mcp              Start MCP server")
	fmt.Println("  fastcode version                Show version")
	fmt.Println()
	fmt.Println("Get started: https://github.com/duyhunghd6/fastcode-cli")
}
