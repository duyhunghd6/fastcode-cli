package orchestrator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/duyhunghd6/fastcode-cli/internal/agent"
	"github.com/duyhunghd6/fastcode-cli/internal/cache"
	"github.com/duyhunghd6/fastcode-cli/internal/graph"
	"github.com/duyhunghd6/fastcode-cli/internal/index"
	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/loader"
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
	log.Printf("[engine] loaded %d files from %s", len(repo.Files), repo.Name)

	// Check cache
	if !forceReindex && e.cache.Exists(repo.Name) {
		cached, err := e.cache.Load(repo.Name)
		if err == nil {
			log.Printf("[engine] loaded %d elements from cache", len(cached.Elements))
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
		log.Printf("[engine] cache load failed, re-indexing: %v", err)
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
		log.Printf("[engine] embedding failed (BM25 only): %v", err)
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
		log.Printf("[engine] cache save failed: %v", err)
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

// Query performs a full query pipeline: search → agent → answer.
func (e *Engine) Query(question string) (*QueryResult, error) {
	if e.hybrid == nil || len(e.elements) == 0 {
		return nil, fmt.Errorf("no repository indexed — run 'fastcode index <path>' first")
	}

	// Process query
	pq := agent.ProcessQuery(question)
	log.Printf("[engine] query type=%s complexity=%d keywords=%v", pq.QueryType, pq.Complexity, pq.Keywords)

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
