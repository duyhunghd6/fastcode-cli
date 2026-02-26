# FastCode-CLI — Project Plan

> **Goal:** Port [HKUDS/FastCode](https://github.com/HKUDS/FastCode) from Python to Go, creating a single-binary CLI for token-efficient codebase intelligence.

---

## 1. Scope & Module Mapping

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
| 24  | `main.py` (engine)        | 1,700 | `internal/engine`          | P4       |
| 25  | `api.py`                  | 750   | `cmd/fastcode`             | P4       |
| 26  | `mcp_server.py`           | 400   | `cmd/fastcode`             | P4       |

---

## 2. Architecture

```
fastcode-cli/
├── cmd/fastcode/          # CLI entry (Cobra)
│   └── main.go
├── internal/
│   ├── types/             # Shared data structures
│   │   └── types.go       # CodeElement, FunctionInfo, ClassInfo, ImportInfo, FileParseResult
│   ├── util/              # Utilities (path, token counting, language detection)
│   ├── loader/            # Repository file loader & walker
│   ├── parser/            # Tree-sitter AST parsing + extractors
│   ├── graph/             # Call/Dependency/Inheritance graphs
│   ├── index/             # Hybrid indexer (embeddings + BM25)
│   ├── llm/               # LLM client (OpenAI-compatible)
│   ├── agent/             # Iterative retrieval agent
│   ├── cache/             # Disk cache for indexes
│   └── engine/            # Orchestrator (wires everything together)
├── pkg/treesitter/        # Tree-sitter Go bindings helper
├── reference/             # Original Python source
├── docs/                  # Documentation
└── go.mod
```

---

## 3. Phased Implementation

### Phase 1 — Core Engine (Parser + Graph) `[Week 1]`

**Goal:** Parse a Go/Python/JS repository into structured AST nodes and build relationship graphs.

#### Deliverables

| File                               | Description                                                                         |
| ---------------------------------- | ----------------------------------------------------------------------------------- |
| `internal/types/types.go`          | `CodeElement`, `FunctionInfo`, `ClassInfo`, `ImportInfo`, `FileParseResult` structs |
| `internal/util/language.go`        | Language detection from file extension                                              |
| `internal/util/path.go`            | Path normalization, module path conversion                                          |
| `internal/loader/loader.go`        | Walk repo directory, filter files, read content                                     |
| `pkg/treesitter/parser.go`         | Tree-sitter wrapper (init, parse, set language)                                     |
| `internal/parser/parser.go`        | Main parser: dispatch to language-specific extractors                               |
| `internal/parser/go_parser.go`     | Go-specific AST extraction                                                          |
| `internal/parser/python_parser.go` | Python-specific AST extraction                                                      |
| `internal/parser/js_parser.go`     | JavaScript/TypeScript AST extraction                                                |
| `internal/graph/graph.go`          | Dependency, Inheritance, Call graph builder                                         |

#### Tests

| Test File                        | Coverage                                              |
| -------------------------------- | ----------------------------------------------------- |
| `internal/loader/loader_test.go` | File walking, filtering, gitignore respect            |
| `pkg/treesitter/parser_test.go`  | Parse Go/Python/JS snippets, verify tree              |
| `internal/parser/parser_test.go` | Extract functions, classes, imports from sample files |
| `internal/graph/graph_test.go`   | Build graphs, verify edges, traverse hops             |

---

### Phase 2 — Indexing (Embeddings + BM25) `[Week 2]`

**Goal:** Index parsed code elements into a hybrid search store (dense vectors + BM25 text).

#### Deliverables

| File                             | Description                                           |
| -------------------------------- | ----------------------------------------------------- |
| `internal/llm/client.go`         | OpenAI-compatible HTTP client for chat + embeddings   |
| `internal/llm/embedder.go`       | Batch embedding generation via API                    |
| `internal/index/indexer.go`      | Multi-level indexer: file/class/function/doc elements |
| `internal/index/bm25.go`         | BM25 text search using Bleve                          |
| `internal/index/vector_store.go` | In-memory vector store with cosine similarity         |
| `internal/index/hybrid.go`       | Hybrid retrieval: merge vector + BM25 results         |
| `internal/cache/cache.go`        | Disk serialization of index (gob/JSON)                |

#### Tests

| Test File                             | Coverage                                            |
| ------------------------------------- | --------------------------------------------------- |
| `internal/llm/client_test.go`         | Mock HTTP server, verify embedding request/response |
| `internal/index/indexer_test.go`      | Index sample repo, verify element counts            |
| `internal/index/bm25_test.go`         | Index texts, search keywords, verify ranking        |
| `internal/index/vector_store_test.go` | Add vectors, search by cosine, verify top-k         |
| `internal/index/hybrid_test.go`       | Combine BM25 + vector results, verify fusion        |

---

### Phase 3 — Retrieval Agent `[Week 3]`

**Goal:** Port the iterative, budget-aware retrieval agent that gathers relevant code context.

#### Deliverables

| File                          | Description                                              |
| ----------------------------- | -------------------------------------------------------- |
| `internal/agent/retriever.go` | Hybrid retriever orchestrator                            |
| `internal/agent/iterative.go` | Multi-round iterative agent with confidence control      |
| `internal/agent/tools.go`     | Agent tool definitions (search, browse, skim)            |
| `internal/agent/query.go`     | Query processor (complexity scoring, keyword extraction) |
| `internal/agent/answer.go`    | Answer generator (context + LLM prompt)                  |

#### Tests

| Test File                          | Coverage                                      |
| ---------------------------------- | --------------------------------------------- |
| `internal/agent/query_test.go`     | Query parsing, complexity scoring             |
| `internal/agent/retriever_test.go` | End-to-end retrieval on sample index          |
| `internal/agent/iterative_test.go` | Mock LLM, verify iteration stops at threshold |

---

### Phase 4 — CLI & MCP Integration `[Week 4]`

**Goal:** Wire everything into a production CLI and MCP server.

#### Deliverables

| File                        | Description                                                      |
| --------------------------- | ---------------------------------------------------------------- |
| `internal/engine/engine.go` | Main orchestrator: init→load→index→query pipeline                |
| `cmd/fastcode/main.go`      | Cobra CLI: `index`, `query`, `summary`, `serve-mcp`, `serve-api` |
| `cmd/fastcode/mcp.go`       | MCP server (stdio + SSE transport)                               |
| `cmd/fastcode/api.go`       | REST API server (optional)                                       |

#### Tests

| Test File                        | Coverage                                                            |
| -------------------------------- | ------------------------------------------------------------------- |
| `internal/engine/engine_test.go` | Full pipeline: load→index→query on test repo                        |
| CLI integration test             | `go build && ./fastcode index ./testdata && ./fastcode query "..."` |

---

## 4. Test Strategy

### Unit Tests

- Each `internal/` package has `_test.go` files.
- Use `testdata/` directories with small sample repos (Go, Python, JS).
- Run: `go test ./... -v -cover`

### Integration Tests

- Full pipeline test: clone a small public repo → index → query → verify answer.
- Run: `go test ./internal/engine/ -v -run TestFullPipeline -count=1`

### Coverage Target

- **Phase 1:** ≥ 80% coverage on `parser`, `graph`, `loader`
- **Phase 2:** ≥ 75% coverage on `index`, `llm`
- **Phase 3:** ≥ 70% coverage on `agent`
- **Phase 4:** ≥ 60% overall project coverage

### Test Reports

- Generate: `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o docs/coverage.html`
- Summary report: `go test ./... -cover | tee docs/test-report.txt`

---

## 5. Documentation Deliverables

| Document           | Path                    | Description                                              |
| ------------------ | ----------------------- | -------------------------------------------------------- |
| Architecture Guide | `docs/architecture.md`  | System architecture, package responsibilities, data flow |
| API Specification  | `docs/api-spec.md`      | CLI commands, MCP tools, REST endpoints                  |
| Go Port Notes      | `docs/porting-notes.md` | Python→Go translation decisions and tradeoffs            |
| Test Report        | `docs/test-report.txt`  | `go test` output with coverage percentages               |
| Coverage Report    | `docs/coverage.html`    | Visual HTML coverage report                              |

---

## 6. Go Dependencies

```
github.com/spf13/cobra          # CLI framework
github.com/smacker/go-tree-sitter  # AST parsing
github.com/blevesearch/bleve/v2    # BM25 text search
github.com/dominikbraun/graph      # Graph data structures
github.com/joho/godotenv           # .env loading
```

---

## 7. Milestones & Exit Criteria

| Milestone      | Exit Criteria                                                                   |
| -------------- | ------------------------------------------------------------------------------- |
| **M1: Parser** | Can parse Go/Python/JS files and emit `CodeElement` structs. Tests pass.        |
| **M2: Graph**  | Can build Dependency + Call graphs from parsed elements. Graph traversal works. |
| **M3: Index**  | Can index a repo and search by keyword (BM25) and semantic (vector).            |
| **M4: Agent**  | Can answer a natural-language question about a codebase with relevant context.  |
| **M5: CLI**    | `fastcode index` and `fastcode query` work end-to-end. MCP server starts.       |
