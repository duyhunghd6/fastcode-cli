# FastCode-CLI — Project Plan

> **Goal:** Rewrite [HKUDS/FastCode](https://github.com/HKUDS/FastCode) from Python to Go, creating a single-binary CLI for token-efficient codebase intelligence.

---

## 1. Objective

To rewrite the Python-based `FastCode` framework into a high-performance, statically compiled Go-based CLI tool (`fastcode-cli`). This aligns completely with the project's Go tech stack, eliminating the need to maintain a separate Python environment and Docker container for code comprehension.

### Why Go is Better for FastCode

- **Speed & Concurrency**: Go's goroutines will heavily parallelize the codebase AST parsing and HTTP embedding calls, turning a 20-second Python index into a ~2-second Go index.
- **Single Binary**: No need to manage `uv`, `pip`, `venv`, or large Docker images natively running PyTorch and CUDA just for embeddings — just one fast Go binary.
- **Memory Footprint**: Python FAISS + Pickle + Torch consumes GBs of RAM. Go + Bleve + in-memory vector store will require a fraction of that, keeping local dev environments snappy.

---

## 2. Scope & Module Mapping

The Python FastCode consists of ~8,600 lines across 26 modules. Below is the mapping to Go packages:

| #   | Python Module             | Lines | Go Target                  | Priority |
| --- | ------------------------- | ----- | -------------------------- | -------- |
| 1   | `tree_sitter_parser.py`   | 163   | `pkg/treesitter`           | P1       |
| 2   | `parser.py`               | 1,703 | `internal/parser`          | P1       |
| 3   | `definition_extractor.py` | 280   | `internal/parser` (merged) | P1       |
| 4   | `import_extractor.py`     | 160   | `internal/parser` (merged) | P1       |
| 5   | `call_extractor.py`       | 760   | `internal/parser` (merged) | P1       |
| 6   | `symbol_resolver.py`      | 250   | `internal/parser` (merged) | P1       |
| 7   | `module_resolver.py`      | 100   | `internal/parser` (merged) | P1       |
| 8   | `graph_builder.py`        | 1,014 | `internal/graph`           | P1       |
| 9   | `indexer.py`              | 457   | `internal/index`           | P2       |
| 10  | `embedder.py`             | 190   | `internal/llm`             | P2       |
| 11  | `vector_store.py`         | 830   | `internal/index`           | P2       |
| 12  | `retriever.py`            | 1,444 | `internal/agent`           | P3       |
| 13  | `iterative_agent.py`      | 3,336 | `internal/agent`           | P3       |
| 14  | `agent_tools.py`          | 620   | `internal/agent`           | P3       |
| 15  | `query_processor.py`      | 1,010 | `internal/agent`           | P3       |
| 16  | `answer_generator.py`     | 1,110 | `internal/agent`           | P3       |
| 17  | `loader.py`               | 350   | `internal/loader`          | P1       |
| 18  | `repo_overview.py`        | 410   | `internal/index`           | P2       |
| 19  | `repo_selector.py`        | 540   | `internal/agent`           | P3       |
| 20  | `cache.py`                | 460   | `internal/cache`           | P2       |
| 21  | `utils.py`                | 230   | `internal/util`            | P1       |
| 22  | `path_utils.py`           | 620   | `internal/util`            | P1       |
| 23  | `llm_utils.py`            | 20    | `internal/llm`             | P2       |
| 24  | `main.py` (engine)        | 1,700 | `internal/orchestrator`    | P4       |
| 25  | `api.py`                  | 750   | `cmd/fastcode`             | P4       |
| 26  | `mcp_server.py`           | 400   | `cmd/fastcode`             | P4       |

### Python → Go Architecture Mapping Detail

| Python Module (`fastcode/`)                                   | Go Package / Strategy               | Description                                                                                                                                |
| :------------------------------------------------------------ | :---------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------- |
| `tree_sitter_parser.py`                                       | `github.com/smacker/go-tree-sitter` | Core AST parsing via CGo bindings with grammar support for target languages (Go, Python, Java, JS, Rust, C/C++, C#).                       |
| `parser.py`, `definition_extractor.py`, `import_extractor.py` | `internal/parser`                   | Traverse AST nodes to extract classes, functions, variables, and imports using tree-sitter queries.                                        |
| `graph_builder.py`, `module_resolver.py`                      | `internal/graph`                    | Build Call Graph, Dependency Graph, and Inheritance Graph stored in-memory using native Go graph data structures.                          |
| `embedder.py`                                                 | `internal/llm` (External API)       | Python uses `sentence-transformers` locally. Go relies on external embedding APIs (OpenAI / Ollama) to avoid bundling heavy ML frameworks. |
| `indexer.py`, `vector_store.py`                               | `internal/index`                    | Python uses FAISS + BM25 (Pickle). Go uses in-memory cosine similarity vector store + BM25 via native tokenizer.                           |
| `iterative_agent.py`, `retriever.py`, `query_processor.py`    | `internal/agent`                    | The core "budget-aware algorithm" — pure Go business logic utilizing LLM API calls.                                                        |
| `api.py`, `mcp_server.py`                                     | `cmd/fastcode`                      | Cobra CLI interface + JSON-RPC MCP server over stdio.                                                                                      |

---

## 3. Architecture

```
fastcode-cli/
├── cmd/fastcode/              # CLI entry (Cobra)
│   ├── main.go                # Root command + index/query subcommands
│   └── mcp.go                 # MCP server (JSON-RPC over stdio)
├── internal/
│   ├── types/                 # Shared data structures
│   │   └── types.go           # CodeElement, FunctionInfo, ClassInfo, ImportInfo, FileParseResult
│   ├── util/                  # Utilities (path, token counting, language detection)
│   │   ├── language.go        # Language detection from file extension
│   │   └── path.go            # Path normalization, module path conversion
│   ├── loader/                # Repository file loader & walker
│   │   └── loader.go          # Walk repo directory, filter files, read content
│   ├── parser/                # Tree-sitter AST parsing + extractors
│   │   ├── parser.go          # Main parser: dispatch to language-specific extractors
│   │   ├── go_parser.go       # Go-specific AST extraction
│   │   ├── python_parser.go   # Python-specific AST extraction
│   │   ├── js_parser.go       # JavaScript/TypeScript AST extraction
│   │   └── other_parsers.go   # Java, Rust, C/C++, C# parsers
│   ├── graph/                 # Call/Dependency/Inheritance graphs
│   │   └── graph.go           # Graph construction & traversal
│   ├── index/                 # Hybrid indexer (embeddings + BM25)
│   │   ├── indexer.go         # Multi-level indexer: file/class/function/doc elements
│   │   ├── bm25.go            # BM25 text search
│   │   ├── vector_store.go    # In-memory vector store with cosine similarity
│   │   └── hybrid.go          # Hybrid retrieval: merge vector + BM25 results
│   ├── llm/                   # LLM client (OpenAI-compatible)
│   │   ├── client.go          # HTTP client for chat completions
│   │   └── embedder.go        # Batch embedding generation via API
│   ├── agent/                 # Iterative retrieval agent
│   │   ├── query.go           # Query processor (complexity scoring, keyword extraction)
│   │   ├── tools.go           # Agent tool definitions (search, browse, skim, list)
│   │   ├── iterative.go       # Multi-round iterative agent with confidence control
│   │   └── answer.go          # Answer generator (context + LLM prompt)
│   ├── cache/                 # Disk cache for indexes
│   │   └── cache.go           # Gob serialization of index
│   └── orchestrator/          # Orchestrator (wires everything together)
│       └── engine.go          # init → load → index → query pipeline
├── pkg/treesitter/            # Tree-sitter Go bindings helper
├── reference/                 # Original Python source
├── docs/                      # Documentation
└── go.mod
```

---

## 4. Go Dependencies

| Dependency                          | Purpose                    |
| ----------------------------------- | -------------------------- |
| `github.com/spf13/cobra`            | CLI framework              |
| `github.com/smacker/go-tree-sitter` | AST parsing (CGo bindings) |
| `github.com/joho/godotenv`          | `.env` file loading        |

> **Note:** Graph data structures and BM25 text search are implemented natively in Go without external libraries to minimize dependencies and binary size.

---

## 5. Phased Implementation

### Phase 1 — Core Engine (Parser + Graph) `[Week 1]` ✅

**Goal:** Parse a Go/Python/JS repository into structured AST nodes and build relationship graphs.

#### Deliverables

| File                               | Description                                                                         | Status |
| ---------------------------------- | ----------------------------------------------------------------------------------- | ------ |
| `internal/types/types.go`          | `CodeElement`, `FunctionInfo`, `ClassInfo`, `ImportInfo`, `FileParseResult` structs | ✅     |
| `internal/util/language.go`        | Language detection from file extension                                              | ✅     |
| `internal/util/path.go`            | Path normalization, module path conversion                                          | ✅     |
| `internal/loader/loader.go`        | Walk repo directory, filter files, read content                                     | ✅     |
| `pkg/treesitter/parser.go`         | Tree-sitter wrapper (init, parse, set language)                                     | ✅     |
| `internal/parser/parser.go`        | Main parser: dispatch to language-specific extractors                               | ✅     |
| `internal/parser/go_parser.go`     | Go-specific AST extraction                                                          | ✅     |
| `internal/parser/python_parser.go` | Python-specific AST extraction                                                      | ✅     |
| `internal/parser/js_parser.go`     | JavaScript/TypeScript AST extraction                                                | ✅     |
| `internal/parser/other_parsers.go` | Java, Rust, C/C++, C# parsers                                                       | ✅     |
| `internal/graph/graph.go`          | Dependency, Inheritance, Call graph builder                                         | ✅     |

#### Tests

| Test File                        | Coverage                                              | Status |
| -------------------------------- | ----------------------------------------------------- | ------ |
| `internal/loader/loader_test.go` | File walking, filtering, gitignore respect            | ✅     |
| `pkg/treesitter/parser_test.go`  | Parse Go/Python/JS snippets, verify tree              | ✅     |
| `internal/parser/parser_test.go` | Extract functions, classes, imports from sample files | ✅     |
| `internal/graph/graph_test.go`   | Build graphs, verify edges, traverse hops             | ✅     |

---

### Phase 2 — Indexing (Embeddings + BM25) `[Week 2]` ✅

**Goal:** Index parsed code elements into a hybrid search store (dense vectors + BM25 text).

#### Deliverables

| File                             | Description                                           | Status |
| -------------------------------- | ----------------------------------------------------- | ------ |
| `internal/llm/client.go`         | OpenAI-compatible HTTP client for chat + embeddings   | ✅     |
| `internal/llm/embedder.go`       | Batch embedding generation via API                    | ✅     |
| `internal/index/indexer.go`      | Multi-level indexer: file/class/function/doc elements | ✅     |
| `internal/index/bm25.go`         | BM25 text search                                      | ✅     |
| `internal/index/vector_store.go` | In-memory vector store with cosine similarity         | ✅     |
| `internal/index/hybrid.go`       | Hybrid retrieval: merge vector + BM25 results         | ✅     |
| `internal/cache/cache.go`        | Disk serialization of index (gob encoding)            | ✅     |

#### Tests

| Test File                             | Coverage                                            | Status |
| ------------------------------------- | --------------------------------------------------- | ------ |
| `internal/llm/client_test.go`         | Mock HTTP server, verify embedding request/response | ✅     |
| `internal/llm/embedder_test.go`       | Batch embedding, retry logic                        | ✅     |
| `internal/index/indexer_test.go`      | Index sample repo, verify element counts            | ✅     |
| `internal/index/bm25_test.go`         | Index texts, search keywords, verify ranking        | ✅     |
| `internal/index/vector_store_test.go` | Add vectors, search by cosine, verify top-k         | ✅     |
| `internal/index/hybrid_test.go`       | Combine BM25 + vector results, verify fusion        | ✅     |
| `internal/cache/cache_test.go`        | Serialize/deserialize index                         | ✅     |

---

### Phase 3 — Retrieval Agent `[Week 3]` ✅

**Goal:** Port the iterative, budget-aware retrieval agent that gathers relevant code context.

#### Deliverables

| File                          | Description                                              | Status |
| ----------------------------- | -------------------------------------------------------- | ------ |
| `internal/agent/query.go`     | Query processor (complexity scoring, keyword extraction) | ✅     |
| `internal/agent/tools.go`     | Agent tool definitions (search, browse, skim, list)      | ✅     |
| `internal/agent/iterative.go` | Multi-round iterative agent with confidence control      | ✅     |
| `internal/agent/answer.go`    | Answer generator (context + LLM prompt)                  | ✅     |

#### Tests

| Test File                          | Coverage                                      | Status |
| ---------------------------------- | --------------------------------------------- | ------ |
| `internal/agent/query_test.go`     | Query parsing, complexity scoring             | ✅     |
| `internal/agent/tools_test.go`     | Agent tool execution                          | ✅     |
| `internal/agent/iterative_test.go` | Mock LLM, verify iteration stops at threshold | ✅     |
| `internal/agent/answer_test.go`    | Answer generation with context                | ✅     |

---

### Phase 4 — CLI & MCP Integration `[Week 4]` ✅

**Goal:** Wire everything into a production CLI and MCP server.

#### Deliverables

| File                              | Description                                       | Status |
| --------------------------------- | ------------------------------------------------- | ------ |
| `internal/orchestrator/engine.go` | Main orchestrator: init→load→index→query pipeline | ✅     |
| `cmd/fastcode/main.go`            | Cobra CLI: `index`, `query`, `serve-mcp`          | ✅     |
| `cmd/fastcode/mcp.go`             | MCP server (JSON-RPC over stdio)                  | ✅     |

#### Tests

| Test File                              | Coverage                                     | Status |
| -------------------------------------- | -------------------------------------------- | ------ |
| `internal/orchestrator/engine_test.go` | Full pipeline: load→index→query on test repo | ✅     |
| `internal/orchestrator/e2e_test.go`    | End-to-end integration test                  | ✅     |
| `cmd/fastcode/main_test.go`            | CLI command parsing and execution            | ✅     |
| `cmd/fastcode/mcp_test.go`             | MCP protocol compliance                      | ✅     |

---

### Phase 5 — Ecosystem _(Planned)_

**Goal:** REST API, multi-repo support, and distribution.

| Deliverable                     | Description                             | Status |
| ------------------------------- | --------------------------------------- | ------ |
| `cmd/fastcode/api.go`           | REST API server (optional)              | ⬜     |
| Multi-repo query                | Cross-repository reasoning              | ⬜     |
| Pre-built binaries              | GitHub Releases for Linux/macOS/Windows | ⬜     |
| `go install` / Homebrew formula | Easy installation via package managers  | ⬜     |

---

## 6. Test Strategy

### Unit Tests

- Each `internal/` package has `_test.go` files.
- Use `testdata/` directories with small sample repos (Go, Python, JS).
- Run: `go test ./... -v -cover`

### Integration Tests

- Full pipeline test: load a repo → index → query → verify answer.
- Run: `go test ./internal/orchestrator/ -v -run TestFullPipeline -count=1`

### End-to-End Tests

- Build binary and run CLI commands:
  ```bash
  go build -o fastcode ./cmd/fastcode
  ./fastcode index ./testdata/sample-repo
  ./fastcode query --repo ./testdata/sample-repo "What does this project do?"
  ```

### Coverage Targets

- **Phase 1:** ≥ 80% coverage on `parser`, `graph`, `loader`
- **Phase 2:** ≥ 75% coverage on `index`, `llm`
- **Phase 3:** ≥ 70% coverage on `agent`
- **Phase 4:** ≥ 60% overall project coverage

### Test Commands

```bash
# Run all tests with coverage
go test ./... -v -cover

# Generate coverage profile
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o docs/coverage.html

# Summary report
go test ./... -cover | tee docs/test-report.txt
```

---

## 7. Documentation Deliverables

| Document            | Path                   | Description                                             |
| ------------------- | ---------------------- | ------------------------------------------------------- |
| README              | `README.md`            | Project overview, features, quick start, architecture   |
| README (Tiếng Việt) | `README-vi.md`         | Vietnamese version of README                            |
| Project Plan        | `PLAN.md`              | This document — phased plan, module mapping, milestones |
| Test Report         | `docs/test-report.txt` | `go test` output with coverage percentages              |
| Coverage Report     | `docs/coverage.html`   | Visual HTML coverage report                             |

---

## 8. Milestones & Exit Criteria

| Milestone      | Exit Criteria                                                                   | Status |
| -------------- | ------------------------------------------------------------------------------- | ------ |
| **M1: Parser** | Can parse Go/Python/JS files and emit `CodeElement` structs. Tests pass.        | ✅     |
| **M2: Graph**  | Can build Dependency + Call graphs from parsed elements. Graph traversal works. | ✅     |
| **M3: Index**  | Can index a repo and search by keyword (BM25) and semantic (vector).            | ✅     |
| **M4: Agent**  | Can answer a natural-language question about a codebase with relevant context.  | ✅     |
| **M5: CLI**    | `fastcode index` and `fastcode query` work end-to-end. MCP server starts.       | ✅     |
| **M6: API**    | REST API server and multi-repo support.                                         | ⬜     |
