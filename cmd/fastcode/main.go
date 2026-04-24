package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/duyhunghd6/fastcode-cli/internal/config"
	"github.com/duyhunghd6/fastcode-cli/internal/logger"
	"github.com/duyhunghd6/fastcode-cli/internal/orchestrator"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var version = "0.1.0-dev"
var buildTime = "unknown"
var gitCommit = "unknown"

func main() {
	fmt.Println("[init] Starting fastcode main execution...")
	// Load global config from ~/.fastcode/config.yaml first
	if _, err := config.Load(); err != nil {
		log.Printf("warning: config load: %v", err)
	}
	// Then load local .env (overrides YAML since env vars take precedence)
	_ = godotenv.Load()

	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// buildRootCmd creates the root cobra command with all subcommands.
func buildRootCmd() *cobra.Command {
	versionStr := fmt.Sprintf("%s (built: %s, commit: %s)", version, buildTime, gitCommit)

	rootCmd := &cobra.Command{
		Use:   "fastcode",
		Short: "⚡ FastCode-CLI — Codebase Intelligence Engine",
		Long: `FastCode-CLI is a Go-based code understanding tool that combines
AST parsing, graph analysis, BM25 keyword search, vector embeddings,
and LLM-powered iterative retrieval to answer questions about codebases.`,
		Version: versionStr,
	}

	// Shared flags
	var cacheDir string
	var embeddingModel string
	var noEmbeddings bool
	var debug bool

	rootCmd.Flags().BoolP("version", "V", false, "Print version information and quit")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "Cache directory (default: ~/.fastcode/cache)")
	rootCmd.PersistentFlags().StringVar(&embeddingModel, "embedding-model", "", "Embedding model name (default: from config)")
	rootCmd.PersistentFlags().BoolVar(&noEmbeddings, "no-embeddings", true, "Skip embedding generation (BM25 only, defaults to true)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")

	// Set debug mode globally when root command runs
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logger.DebugEnabled = debug
	}

	buildConfig := func() orchestrator.Config {
		cfg := orchestrator.DefaultConfig()
		if cacheDir != "" {
			cfg.CacheDir = cacheDir
		}
		if embeddingModel != "" {
			cfg.EmbeddingModel = embeddingModel
		}
		cfg.NoEmbeddings = noEmbeddings
		return cfg
	}

	// --- index command ---
	var forceReindex bool
	var jsonOutput bool

	indexCmd := &cobra.Command{
		Use:   "index <repo-path>",
		Short: "Index a local repository",
		Long:  "Parse, analyze, and index a code repository for querying.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := args[0]
			cfg := buildConfig()
			engine := orchestrator.NewEngine(cfg)

			if debug {
				fmt.Printf("⚡ Indexing %s...\n", repoPath)
			}
			start := time.Now()

			result, err := engine.Index(repoPath, forceReindex)
			if err != nil {
				return fmt.Errorf("indexing failed: %w", err)
			}

			elapsed := time.Since(start)

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("\n✅ Indexed %s in %s\n", result.RepoName, elapsed.Round(time.Millisecond))
			fmt.Printf("   Files:    %d\n", result.TotalFiles)
			fmt.Printf("   Elements: %d\n", result.TotalElements)
			if result.Cached {
				fmt.Println("   Source:   cache (use --force to reindex)")
			}
			if result.GraphStats != nil {
				fmt.Printf("   Graphs:   %v\n", result.GraphStats)
			}
			return nil
		},
	}
	indexCmd.Flags().BoolVar(&forceReindex, "force", false, "Force re-indexing (ignore cache)")
	indexCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.AddCommand(indexCmd)

	// --- query command ---
	queryCmd := &cobra.Command{
		Use:   "query <question>",
		Short: "Query the indexed codebase",
		Long:  "Ask a question about a previously indexed codebase.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Join remaining args as the question
			question := ""
			for i, arg := range args {
				if i > 0 {
					question += " "
				}
				question += arg
			}

			repoPath, _ := cmd.Flags().GetString("repo")
			cfg := buildConfig()
			engine := orchestrator.NewEngine(cfg)

			// Index first if repo is specified
			if repoPath != "" {
				if debug {
					fmt.Printf("⚡ Loading index for %s...\n", repoPath)
				}
				_, err := engine.Index(repoPath, false)
				if err != nil {
					return fmt.Errorf("index load failed: %w", err)
				}
			}

			if debug {
				fmt.Printf("🔍 Querying: %s\n\n", question)
			}
			start := time.Now()

			result, err := engine.Query(question)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}

			elapsed := time.Since(start)

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Println(result.Answer)
			fmt.Printf("\n---\n")
			fmt.Printf("⏱  %s | 🎯 Confidence: %d%% | 🔄 Rounds: %d | 📦 Elements: %d | Stop: %s\n",
				elapsed.Round(time.Millisecond), result.Confidence, result.Rounds, result.Elements, result.StopReason)
			return nil
		},
	}
	queryCmd.Flags().String("repo", "", "Repository path to index/load")
	queryCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.AddCommand(queryCmd)

	// --- analyze command ---
	var topK int
	var includeCode bool
	var outputFormat string
	var reIndex bool

	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze codebase with hybrid search (no LLM needed)",
		Long: `Perform direct hybrid search (BM25 + optional vector) against an indexed
repository. Returns scored results as structured JSON for script consumption.
No LLM API key required.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")
			queryStr, _ := cmd.Flags().GetString("query")

			if repoPath == "" {
				return fmt.Errorf("--repo is required")
			}
			if queryStr == "" {
				return fmt.Errorf("--query is required")
			}

			cfg := buildConfig()
			engine := orchestrator.NewEngine(cfg)

			// Index: force re-index only when --re-index is passed
			if reIndex {
				fmt.Fprintln(os.Stderr, "Firstcall analyze -> force reindex the code")
			}
			_, err := engine.Index(repoPath, reIndex)
			if err != nil {
				return fmt.Errorf("index failed: %w", err)
			}

			start := time.Now()
			result, err := engine.Analyze(queryStr, topK, includeCode)
			if err != nil {
				return fmt.Errorf("analyze failed: %w", err)
			}
			elapsed := time.Since(start)

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if outputFormat == "toon" {
				fmt.Print(result.ToTOON(includeCode))
				return nil
			}

			// Text format
			fmt.Printf("🔍 Query: %s\n", result.Query)
			fmt.Printf("📦 Repo: %s | Results: %d | ⏱ %s\n\n", result.RepoName, result.Total, elapsed.Round(time.Millisecond))
			for i, hit := range result.Results {
				fmt.Printf("%d. [%.2f] %s (%s) — %s:L%d-%d\n",
					i+1, hit.Score, hit.Name, hit.Type, hit.FilePath, hit.StartLine, hit.EndLine)
				if hit.Signature != "" {
					fmt.Printf("   %s\n", hit.Signature)
				}
				if includeCode && hit.Code != "" {
					lines := strings.Split(hit.Code, "\n")
					preview := hit.Code
					if len(lines) > 10 {
						preview = strings.Join(lines[:10], "\n") + "\n   ... (truncated)"
					}
					fmt.Printf("   ```\n   %s\n   ```\n", preview)
				}
				fmt.Println()
			}
			return nil
		},
	}
	analyzeCmd.Flags().String("repo", "", "Repository path (required)")
	analyzeCmd.Flags().String("query", "", "Search query (required)")
	analyzeCmd.Flags().IntVar(&topK, "top-k", 10, "Number of results to return")
	analyzeCmd.Flags().StringVar(&outputFormat, "format", "toon", "Output format: json, toon, or text")
	analyzeCmd.Flags().BoolVar(&includeCode, "include-code", true, "Include source code in results")
	analyzeCmd.Flags().BoolVar(&reIndex, "reindex", false, "Force re-index before searching (use on first call)")
	analyzeCmd.Flags().BoolVar(&reIndex, "re-index", false, "Force re-index before searching (alias)")
	_ = analyzeCmd.Flags().MarkHidden("re-index")
	rootCmd.AddCommand(analyzeCmd)

	// --- serve-mcp command ---
	serveMCPCmd := &cobra.Command{
		Use:   "serve-mcp",
		Short: "Start MCP (Model Context Protocol) server",
		Long:  "Start a JSON-RPC server implementing the Model Context Protocol for IDE integration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			cfg := buildConfig()
			return serveMCP(cfg, port)
		},
	}
	serveMCPCmd.Flags().Int("port", 9999, "Port to listen on")
	rootCmd.AddCommand(serveMCPCmd)

	// --- completion command ---
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for fastcode.

To load completions:

Bash:
  $ source <(fastcode completion bash)

Zsh:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc  # once
  $ fastcode completion zsh > "${fpath[1]}/_fastcode"
  $ exec zsh

Fish:
  $ fastcode completion fish | source
  $ fastcode completion fish > ~/.config/fish/completions/fastcode.fish

PowerShell:
  PS> fastcode completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	rootCmd.AddCommand(completionCmd)

	return rootCmd
}
