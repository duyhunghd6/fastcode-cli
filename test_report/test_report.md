# FastCode CLI - QA Test Report

**Date:** 2026-02-26  
**Focus:** Core Pipeline Modules & End-to-End Execution

## 1. Executive Summary

This test report summarizes the Quality Assurance (QA) enhancement phase for the FastCode CLI Go port. Testing focused heavily on areas previously lacking coverage: `loader`, `parser`, `treesitter`, and the `orchestrator` engine.

Alongside unit test expansion, an End-to-End (E2E) testing suite was developed to validate that the complete data flow (Parsing → Graphing → Embedding → Querying) functions smoothly against a live environment utilizing hybrid LLM+BM25 retrieval mechanisms.

## 2. Unit Testing Coverage

New test files were systematically implemented across the core indexing domains:

- **`internal/loader/loader_test.go`**: Validates comprehensive repository boundary loading (e.g. nested directories, `.gitignore` parsing logic, and `MaxFileSize` exclusions).
- **`pkg/treesitter/parser_test.go`**: Verifies underlying language dispatcher context initialization logic inside the wrapper.
- **`internal/parser/parser_test.go`**: Validates the correct parsing logic boundary utilizing standard Go code snippet setups.
- **`internal/orchestrator/engine_test.go`**: Checks stable struct configurations when invoking `NewEngine()`.

### Sub-System Coverage Gains

- `internal/loader`: Achieved **81.2%** statement coverage.
- `pkg/treesitter`: Achieved **66.7%** statement coverage.
- `internal/cache`: Achieved **84.6%** statement coverage.

_(Full HTML metrics can be viewed in `docs/test_report/coverage.html`)_

## 3. End-to-End (E2E) Verification

A fully automated End-to-End bash script (`run_e2e.sh`) was introduced. It compiles the FastCode CLI, initializes a mock test repository consisting of real parsed modules, and runs semantic search queries against the hybrid index.

### Validated E2E Flow:

1. `go build` the `fastcode` binary.
2. Delete isolated sandbox cache configurations (`~/.fastcode/cache/test_repo_snapshot.gob`).
3. Form a mocked `test_repo_snapshot` utilizing files from `loader.go` and `parser.go`.
4. Run `./fastcode index test_repo_snapshot` utilizing `.env` LLM embedding keys to form real cosine vectors.
5. Run `./fastcode query "how does the file loader filter by file size?"`.
6. Assert the retrieved context natively retrieves chunk data referencing `Config.MaxFileSize`.

**Result:** ✅ E2E Flow Passed successfully with >90% Confidence LLM assertions.

## 4. Critical Bugs Discovered & Resolved

During the implementation of these QA suites, several hidden runtime panics were discovered and patched:

1. **AST Context Panic:**
   - **Issue**: `ParseCtx()` inside `pkg/treesitter` occasionally triggered nil pointer dereference crashes since Python bindings didn't require explicit HTTP/time contexts.
   - **Fix**: Passed `context.Background()` universally across tree-node injections.
2. **Gob Serialization Crash on Cache Store:**
   - **Issue**: Standard Go serialization via `encoding/gob` panicked when the engine attempted to commit dynamic generic map structs (`Metadata`) and specific arrays like `[]ImportInfo` to the SQLite/Disk fallback file.
   - **Fix**: Explicitly initialized `gob.Register()` for all index map/slice interfaces during `cache.go` module boot.
3. **Missing System Environment Variable Loading**:
   - **Issue**: Using compiled CLI commands failed to automatically extract local `.env` values (like `OPENAI_API_KEY`), defaulting queries back to BM25 solely.
   - **Fix**: Injected `godotenv.Load()` initialization unconditionally in `cmd/fastcode/main.go`.
