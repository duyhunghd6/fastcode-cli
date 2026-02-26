# FastCode CLI — Comprehensive Test Plan

## 1. Objective

Achieve **100% test coverage** across all packages in `fastcode-cli` through comprehensive unit tests and an E2E integration test.

## 2. Current Baseline (Before This Plan)

| Package                 | Coverage | Status                            |
| ----------------------- | -------- | --------------------------------- |
| `cmd/fastcode`          | 0.0%     | No tests                          |
| `internal/agent`        | 42.2%    | Partial (query + tools only)      |
| `internal/cache`        | 84.6%    | Good, needs edge cases            |
| `internal/graph`        | 68.4%    | Partial, missing graph builders   |
| `internal/index`        | 58.6%    | Missing indexer.go entirely       |
| `internal/llm`          | 51.1%    | Missing embedder.go + error paths |
| `internal/loader`       | 81.2%    | Missing gitignore edge cases      |
| `internal/orchestrator` | 6.9%     | Only init tested                  |
| `internal/parser`       | 10.9%    | Only Go parsing, no Python/JS     |
| `internal/types`        | N/A      | Types-only, no logic              |
| `internal/util`         | 60.0%    | Missing several functions         |
| `pkg/treesitter`        | 66.7%    | Missing multi-language tests      |

## 3. Unit Test Strategy

### `internal/agent`

- **query.go**: `ProcessQuery` (empty input), `extractKeywords`, `classifyQuery`, `scoreComplexity` (all branches)
- **tools.go**: `AvailableTools`, `Execute` all tools including `skim_file`, `browse_file` (not found), `search_code`
- **answer.go**: `NewAnswerGenerator`, `GenerateAnswer` (mock LLM), `buildPrompt` (0/1/16+ elements), `truncateStr`, `answerSystemPrompt`
- **iterative.go**: `DefaultAgentConfig`, `NewIterativeAgent`, `Retrieve` (mock LLM), `buildRoundPrompt`, `parseRoundResponse` (valid/invalid JSON), `systemPrompt`, `extractJSON`, `deduplicateElements`, `min`

### `internal/cache`

- `Save`/`Load`/`Exists`/`Delete` including error paths (read-only dir, non-existent file)

### `internal/graph`

- `NewGraph`, `AddEdge` (dedup, self-loop), `Successors`/`Predecessors`, `NodeCount`/`EdgeCount`
- `BuildGraphs` with dependency/inheritance/call edges, `Stats`, `resolveImport` (direct, module-style, no match)
- `GetRelatedElements` across all three graphs, `GenerateElementID`

### `internal/index`

- **bm25.go**: `NewBM25` (default params), `AddDocument`, `Search` (ranked, empty, no match, topK), `DocCount`, `tokenize`
- **vector_store.go**: `Add`, `Search`, `Count`, `Dimension`, `Get`, `cosineSimilarity` (same, orthogonal, mismatch)
- **hybrid.go**: `NewHybridRetriever`, `IndexElements` (with/without embedder), `Search` (BM25 only + hybrid), `ElementCount`
- **indexer.go**: `NewIndexer`, `IndexRepository`, `extractCodeBlock` (edge cases), `truncate`, `generateFileSummary`

### `internal/llm`

- **client.go**: `NewClient`, `NewClientWith`, `ChatCompletion` (success, API error, no choices, parse error), `Embed` (success, API error, default model), `post`, `getEnvOr`
- **embedder.go**: `NewEmbedder` (defaults), `EmbedTexts` (empty, batching), `EmbedText` (single, error), `BuildSearchText` (edge cases)

### `internal/loader`

- `LoadRepository` (success, non-existent, not-a-dir), `ReadFileContent` (success, missing)
- `DefaultConfig`, `loadGitignore`, `matchGitignore` (negation, dir-only, full path)

### `internal/orchestrator`

- `DefaultConfig`, `NewEngine`, `Index` (with temp repo, force reindex, cached)
- `Query` (no index error, direct search), `queryDirect`, `rebuildFromCache`, `simpleAnswer`

### `internal/parser`

- **Go**: functions, methods, structs, interfaces, imports (single/grouped), comments/docstrings
- **Python**: classes, functions, imports, from-imports, decorators, async, module docstring
- **JavaScript**: classes, arrow functions, regular functions, imports, exports
- **Java/Rust/C**: generic visitor (functions, structs)

### `internal/util`

- `GetLanguageFromPath`/`GetLanguageFromExtension`, `IsSupportedFile`, `SupportedExtensions`
- `CountLines`, `ExtractLines` (edge cases), `FilePathToModulePath`, `NormalizePath`, `RelativePath`

### `pkg/treesitter`

- `New`, `SetLanguage` (multiple languages), `Parse` (multi-lang, code), `Language`, language cache

### `cmd/fastcode`

- MCP handlers via `httptest`: `/mcp/initialize`, `/mcp/tools/list`, `/mcp/tools/call`, `/health`
- Helper functions: `writeJSON`, `writeError`, `writeToolResult`

## 4. E2E Test Strategy

### Full Pipeline E2E Test (`internal/orchestrator/e2e_test.go`)

1. Create temp directory with sample Go + Python files
2. `NewEngine(cfg)` with `NoEmbeddings=true`
3. `engine.Index(tempDir, true)` — verify files and elements
4. `engine.Query("how does the loader work?")` — verify answer returned
5. Verify graph stats, element counts, and cache behavior

## 5. Execution & Reporting

```bash
# Run all tests with coverage
go test ./... -coverprofile=coverage.out -count=1 -v

# Per-package coverage summary
go tool cover -func=coverage.out

# HTML coverage report
go tool cover -html=coverage.out -o docs/test_report/coverage.html
```

Results will be saved to `docs/test_report/coverage.txt`.

## 6. Target

**100% statement coverage** on all packages (excluding `internal/types` which has no logic).
