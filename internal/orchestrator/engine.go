package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/agent"
	"github.com/duyhunghd6/fastcode-cli/internal/cache"
	"github.com/duyhunghd6/fastcode-cli/internal/graph"
	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/loader"
	"github.com/duyhunghd6/fastcode-cli/internal/logger"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// Engine is the top-level orchestrator connecting all FastCode modules.
type Engine struct {
	client   *llm.Client
	embedder *llm.Embedder
	cache    *cache.IndexCache
	graphs   *graph.CodeGraphs
	hybrid   *index.HybridRetriever
	elements []types.CodeElement
	repoName string
	repoPath string // Absolute path to the repo root
	cacheDir string
}

// Config holds engine configuration.
type Config struct {
	CacheDir       string
	EmbeddingModel string
	BatchSize      int
	NoEmbeddings   bool // If true, skip embedding generation (BM25 only)
}

// DefaultConfig returns the default engine configuration.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	embeddingModel := os.Getenv("EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	return Config{
		CacheDir:       filepath.Join(home, ".fastcode", "cache"),
		EmbeddingModel: embeddingModel,
		BatchSize:      32,
		NoEmbeddings:   false,
	}
}

// NewEngine creates a new FastCode engine.
func NewEngine(cfg Config) *Engine {
	client := llm.NewClient()
	var embedder *llm.Embedder
	if !cfg.NoEmbeddings && client.APIKey != "" {
		embedder = llm.NewEmbedder(client, cfg.EmbeddingModel, cfg.BatchSize)
	}

	return &Engine{
		client:   client,
		embedder: embedder,
		cache:    cache.NewIndexCache(cfg.CacheDir),
		cacheDir: cfg.CacheDir,
	}
}

// IndexResult holds the result of an indexing operation.
type IndexResult struct {
	RepoName      string         `json:"repo_name"`
	TotalFiles    int            `json:"total_files"`
	TotalElements int            `json:"total_elements"`
	GraphStats    map[string]any `json:"graph_stats"`
	Cached        bool           `json:"cached"`
}

// Index parses, indexes, and optionally embeds a repository.
func (e *Engine) Index(repoPath string, forceReindex bool) (*IndexResult, error) {
	// Load repository
	loaderCfg := loader.DefaultConfig()
	repo, err := loader.LoadRepository(repoPath, loaderCfg)
	if err != nil {
		return nil, fmt.Errorf("load repository: %w", err)
	}
	e.repoName = repo.Name
	e.repoPath, _ = filepath.Abs(repoPath)
	logger.Debugf("[engine] loaded %d files from %s", len(repo.Files), repo.Name)

	// Check cache
	if !forceReindex && e.cache.Exists(repo.Name) {
		cached, err := e.cache.Load(repo.Name)
		if err == nil {
			logger.Debugf("[engine] loaded %d elements from cache", len(cached.Elements))
			e.elements = cached.Elements
			e.rebuildFromCache(cached)
			return &IndexResult{
				RepoName:      repo.Name,
				TotalFiles:    len(repo.Files),
				TotalElements: len(e.elements),
				GraphStats:    e.graphs.Stats(),
				Cached:        true,
			}, nil
		}
		logger.Debugf("[engine] cache load failed, re-indexing: %v", err)
	}

	// Parse and index
	indexer := index.NewIndexer(repo.Name)
	elements, err := indexer.IndexRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("index repository: %w", err)
	}
	e.elements = elements

	// Build graphs
	e.graphs = graph.NewCodeGraphs()
	e.graphs.BuildGraphs(elements)

	// Build hybrid search index
	vs := index.NewVectorStore()
	bm := index.NewBM25(1.5, 0.75)
	e.hybrid = index.NewHybridRetriever(vs, bm)

	err = e.hybrid.IndexElements(elements, e.embedder)
	if err != nil {
		logger.Debugf("[engine] embedding failed (BM25 only): %v", err)
	}

	// Cache results
	cachedData := &cache.CachedIndex{
		RepoName: repo.Name,
		Elements: elements,
		Vectors:  make(map[string][]float32),
	}
	// Store vectors if available
	for _, elem := range elements {
		if vec := vs.Get(elem.ID); vec != nil {
			cachedData.Vectors[elem.ID] = vec
		}
	}
	if err := e.cache.Save(repo.Name, cachedData); err != nil {
		logger.Debugf("[engine] cache save failed: %v", err)
	}

	return &IndexResult{
		RepoName:      repo.Name,
		TotalFiles:    len(repo.Files),
		TotalElements: len(elements),
		GraphStats:    e.graphs.Stats(),
		Cached:        false,
	}, nil
}

// QueryResult holds the result of a query operation.
type QueryResult struct {
	Answer     string `json:"answer"`
	Confidence int    `json:"confidence"`
	Rounds     int    `json:"rounds"`
	StopReason string `json:"stop_reason"`
	Elements   int    `json:"elements_used"`
}

// AnalyzeResult holds the result of a direct hybrid search (no LLM).
type AnalyzeResult struct {
	Query    string      `json:"query"`
	RepoName string      `json:"repo_name"`
	Total    int         `json:"total_results"`
	Results  []SearchHit `json:"results"`
}

// SearchHit represents a single search result with relevance score.
type SearchHit struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	FilePath  string  `json:"file_path"`
	StartLine int     `json:"start_line"`
	EndLine   int     `json:"end_line"`
	Score     float64 `json:"relevance_score"`
	Signature string  `json:"signature,omitempty"`
	Docstring string  `json:"docstring,omitempty"`
	Code      string  `json:"code,omitempty"`
}

// Analyze performs a direct hybrid search without LLM, returning scored results.
// This is the script-callable entry point for agent coding tools.
func (e *Engine) Analyze(question string, topK int, includeCode bool) (*AnalyzeResult, error) {
	if e.hybrid == nil || len(e.elements) == 0 {
		return nil, fmt.Errorf("no repository indexed — run 'fastcode index <path>' first")
	}

	// Optionally embed the query for semantic search
	var queryVec []float32
	if e.embedder != nil {
		vec, err := e.embedder.EmbedText(question)
		if err == nil {
			queryVec = vec
		}
	}

	results := e.hybrid.Search(question, queryVec, topK)

	hits := make([]SearchHit, 0, len(results))
	for _, r := range results {
		if r.Element == nil {
			continue
		}
		hit := SearchHit{
			Name:      r.Element.Name,
			Type:      r.Element.Type,
			FilePath:  r.Element.RelativePath,
			StartLine: r.Element.StartLine,
			EndLine:   r.Element.EndLine,
			Score:     r.Score,
			Signature: r.Element.Signature,
			Docstring: r.Element.Docstring,
		}
		if includeCode {
			hit.Code = r.Element.Code
		}
		hits = append(hits, hit)
	}

	return &AnalyzeResult{
		Query:    question,
		RepoName: e.repoName,
		Total:    len(hits),
		Results:  hits,
	}, nil
}

// ToTOON converts AnalyzeResult to TOON (Token-Oriented Object Notation) format.
// TOON uses tabular headers for homogeneous arrays, saving ~40% tokens vs JSON.
func (ar *AnalyzeResult) ToTOON(includeCode bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("query: %s\n", ar.Query))
	sb.WriteString(fmt.Sprintf("repo_name: %s\n", ar.RepoName))
	sb.WriteString(fmt.Sprintf("total_results: %d\n", ar.Total))

	if len(ar.Results) == 0 {
		sb.WriteString("results:\n  (empty)\n")
		return sb.String()
	}

	// Tabular TOON: declare fields once, then rows
	if includeCode {
		sb.WriteString("results[" + fmt.Sprintf("%d", len(ar.Results)) + "]{name,type,file_path,start_line,end_line,relevance_score,signature,docstring,code}:\n")
	} else {
		sb.WriteString("results[" + fmt.Sprintf("%d", len(ar.Results)) + "]{name,type,file_path,start_line,end_line,relevance_score,signature,docstring}:\n")
	}

	for _, hit := range ar.Results {
		sig := hit.Signature
		if sig == "" {
			sig = "-"
		}
		doc := hit.Docstring
		if doc == "" {
			doc = "-"
		}
		// Escape commas in values
		sig = strings.ReplaceAll(sig, ",", "\\,")
		doc = strings.ReplaceAll(doc, ",", "\\,")

		if includeCode {
			code := hit.Code
			if code == "" {
				code = "-"
			} else {
				// Encode newlines for single-line TOON row
				code = strings.ReplaceAll(code, "\n", "\\n")
				code = strings.ReplaceAll(code, ",", "\\,")
				// Truncate very long code to keep TOON usable
				if len(code) > 2000 {
					code = code[:2000] + "...(truncated)"
				}
			}
			sb.WriteString(fmt.Sprintf("  %s,%s,%s,%d,%d,%.4f,%s,%s,%s\n",
				hit.Name, hit.Type, hit.FilePath, hit.StartLine, hit.EndLine, hit.Score, sig, doc, code))
		} else {
			sb.WriteString(fmt.Sprintf("  %s,%s,%s,%d,%d,%.4f,%s,%s\n",
				hit.Name, hit.Type, hit.FilePath, hit.StartLine, hit.EndLine, hit.Score, sig, doc))
		}
	}

	return sb.String()
}

// Query performs a full query pipeline: search → agent → answer.
func (e *Engine) Query(question string) (*QueryResult, error) {
	if e.hybrid == nil || len(e.elements) == 0 {
		return nil, fmt.Errorf("no repository indexed — run 'fastcode index <path>' first")
	}

	// Process query
	pq := agent.ProcessQuery(question)
	logger.Debugf("[engine] query type=%s complexity=%d keywords=%v", pq.QueryType, pq.Complexity, pq.Keywords)

	// If we have an API key, use the iterative agent
	if e.client.APIKey != "" {
		return e.queryWithAgent(question, pq)
	}

	// Fallback: direct search without LLM
	return e.queryDirect(question, pq)
}

func (e *Engine) queryWithAgent(question string, pq *agent.ProcessedQuery) (*QueryResult, error) {
	// Set up agent
	toolExec := agent.NewToolExecutor(e.hybrid, e.embedder, e.elements)
	toolExec.SetRepoRoot(e.repoPath, e.repoName)
	agentCfg := agent.DefaultAgentConfig()
	iterAgent := agent.NewIterativeAgent(e.client, toolExec, e.graphs, agentCfg)

	// Run retrieval
	retrieval, err := iterAgent.Retrieve(question, pq)
	if err != nil {
		return nil, fmt.Errorf("agent retrieval: %w", err)
	}

	// Generate answer
	gen := agent.NewAnswerGenerator(e.client)
	answer, err := gen.GenerateAnswer(question, pq, retrieval.Elements)
	if err != nil {
		return nil, fmt.Errorf("answer generation: %w", err)
	}

	return &QueryResult{
		Answer:     answer,
		Confidence: retrieval.Confidence,
		Rounds:     retrieval.Rounds,
		StopReason: retrieval.StopReason,
		Elements:   len(retrieval.Elements),
	}, nil
}

func (e *Engine) queryDirect(question string, pq *agent.ProcessedQuery) (*QueryResult, error) {
	// Direct hybrid search without LLM agent
	var queryVec []float32
	if e.embedder != nil {
		vec, err := e.embedder.EmbedText(question)
		if err == nil {
			queryVec = vec
		}
	}

	results := e.hybrid.Search(question, queryVec, 10)
	var sb fmt.Stringer = &simpleAnswer{}
	answer := &simpleAnswer{}
	for _, r := range results {
		if r.Element != nil {
			answer.addResult(r.Element)
		}
	}
	_ = sb // suppress unused

	return &QueryResult{
		Answer:     answer.String(),
		Confidence: 50,
		Rounds:     1,
		StopReason: "direct_search",
		Elements:   len(results),
	}, nil
}

func (e *Engine) rebuildFromCache(cached *cache.CachedIndex) {
	e.graphs = graph.NewCodeGraphs()
	e.graphs.BuildGraphs(cached.Elements)

	vs := index.NewVectorStore()
	for id, vec := range cached.Vectors {
		vs.Add(id, vec)
	}
	bm := index.NewBM25(1.5, 0.75)
	e.hybrid = index.NewHybridRetriever(vs, bm)
	_ = e.hybrid.IndexElements(cached.Elements, nil)
}

// simpleAnswer builds a text answer from search results without LLM.
type simpleAnswer struct {
	lines []string
}

func (sa *simpleAnswer) addResult(elem *types.CodeElement) {
	sa.lines = append(sa.lines, fmt.Sprintf("[%s] %s (%s:L%d-%d)\n  %s",
		elem.Type, elem.Name, elem.RelativePath, elem.StartLine, elem.EndLine, elem.Signature))
}

func (sa *simpleAnswer) String() string {
	if len(sa.lines) == 0 {
		return "No matching code elements found."
	}
	result := "Found matching code elements:\n\n"
	for _, l := range sa.lines {
		result += l + "\n\n"
	}
	return result
}
