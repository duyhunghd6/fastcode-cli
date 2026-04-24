# TEST-RULES

## Codebase Directory Structure

```
fastcode-cli/
├── cmd/
│   ├── fastcode/           # CLI entrypoint (main.go, mcp.go)
│   └── ast-compare/        # AST comparison utility
├── internal/
│   ├── agent/              # Query processor, tools, iterative agent, answer generator
│   ├── cache/              # Gob-based index cache (save/load/delete)
│   ├── graph/              # Call, dependency, inheritance graphs
│   ├── index/              # BM25, vector store, hybrid retriever, indexer
│   ├── llm/                # OpenAI-compatible client + embedder
│   ├── loader/             # Repository file walker + gitignore filtering
│   ├── orchestrator/       # Engine: load → index → cache → search → agent → answer
│   ├── parser/             # Per-language AST extractors (Go, Python, JS, Java, Rust, C)
│   ├── types/              # Shared structs (CodeElement, FunctionInfo, ClassInfo, etc.)
│   └── util/               # Language detection, path helpers
├── pkg/
│   └── treesitter/         # Tree-sitter wrapper (multi-language parsing)
├── docs/
│   ├── iteration-reports/  # iteration-{N}.md per development phase
│   ├── test_report/        # coverage.out, coverage.html, test_report.md
│   ├── researches/         # Architecture research documents
│   ├── test_plan.md        # QA strategy and gap analysis
│   ├── similarity_report.md
│   └── ast_similarity_report.md
├── reference/              # Original Python codebase (read-only reference)
├── run_e2e.sh              # Automated E2E test script
├── .env                    # API keys (never commit values)
└── PLAN.md                 # Project roadmap
```

---

## Iteration Report Rules

File: `docs/iteration-reports/iteration-{N}.md`

1. **Header**: Title, date, duration, status (✅ Complete / 🔄 In Progress).
2. **Objective**: One sentence describing the iteration goal.
3. **Deliverables table**: `#`, `File`, `Package/Type`, `Lines`, `Description`.
4. **Test Results**: Pass/fail count + per-package coverage table.
5. **Architecture Decisions**: Numbered list. Why, not what.
6. **Cumulative Project Summary**: Running totals (phases, files, LOC, tests).
7. Keep each iteration ≤ 100 lines. No prose padding.

---

## Test Plan Rules

File: `docs/test_plan.md`

1. List every package under `internal/` and `pkg/` with current coverage %.
2. For each package, state:
   - **Target coverage**: minimum 80%, stretch 95%.
   - **Key scenarios**: list edge cases by name, not description paragraphs.
   - **Gaps**: what is NOT tested and why.
3. Update `test_plan.md` before starting any new test implementation.

---

## Acceptance Criteria

A feature/fix is accepted when:

1. `go test ./... -race` passes with **zero failures**.
2. Every touched package maintains **≥ 80% statement coverage**.
3. New public functions have at least one test exercising the happy path and one error path.
4. E2E script (`run_e2e.sh`) passes: build → index → query → assert keyword.
5. `go vet ./...` reports zero issues.

---

## QA Rules

1. **Test file naming**: `{name}_test.go` in the same package directory. No `_test` package suffix unless testing unexported behavior.
2. **No test pollution**: Each test creates its own `t.TempDir()`. Never write to the working directory.
3. **Table-driven tests**: Use `[]struct{ name string; ... }` for ≥ 3 similar cases.
4. **Mock external calls**: LLM client, HTTP, filesystem — use interfaces or test servers. Never call live APIs in unit tests.
5. **Coverage artifacts**: Always output to `docs/test_report/`. Commands:
   ```sh
   go test ./... -coverprofile=docs/test_report/coverage.out -v | tee docs/test_report/coverage_verbose.txt
   go tool cover -html=docs/test_report/coverage.out -o docs/test_report/coverage.html
   ```
6. **Bug → test**: Every bug fix must include a regression test proving the fix before closing.
7. **No skipped tests**: `t.Skip()` is only allowed with a linked issue tracking re-enablement.
