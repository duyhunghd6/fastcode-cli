# FastCode CLI — Test Report

**Date:** 2026-02-26  
**Go Version:** 1.24  
**Status:** ✅ All Tests Passing

---

## 1. Coverage Summary

| Package                 | Coverage          | Status |
| :---------------------- | :---------------- | :----- |
| `internal/agent`        | **97.3%**         | ✅     |
| `internal/cache`        | **96.7%**         | ✅     |
| `internal/graph`        | **100.0%**        | ✅     |
| `internal/index`        | **99.5%**         | ✅     |
| `internal/llm`          | **97.8%**         | ✅     |
| `internal/loader`       | **92.8%**         | ✅     |
| `internal/orchestrator` | **96.6%**         | ✅     |
| `internal/parser`       | **95.5%**         | ✅     |
| `internal/util`         | **100.0%**        | ✅     |
| `pkg/treesitter`        | **97.4%**         | ✅     |
| `cmd/fastcode`          | **54.9%**         | ✅     |
| `internal/types`        | — (no test files) | N/A    |

> **Weighted Average: ~96%+ across core packages**

---

## 2. Test Execution Summary

Total test cases executed across all packages:

| Package                 | Tests | Status  |
| :---------------------- | :---- | :------ |
| `cmd/fastcode`          | 26    | ✅ PASS |
| `internal/agent`        | 29    | ✅ PASS |
| `internal/cache`        | 10    | ✅ PASS |
| `internal/graph`        | 15    | ✅ PASS |
| `internal/index`        | 22    | ✅ PASS |
| `internal/llm`          | 14    | ✅ PASS |
| `internal/loader`       | 17    | ✅ PASS |
| `internal/orchestrator` | 15    | ✅ PASS |
| `internal/parser`       | 30    | ✅ PASS |
| `internal/util`         | 10    | ✅ PASS |
| `pkg/treesitter`        | 18    | ✅ PASS |

---

## 3. Artifacts Generated

All coverage artifacts are consolidated in `docs/test_report/`:

| File                   | Description                       |
| :--------------------- | :-------------------------------- |
| `coverage.out`         | Raw Go coverage profile           |
| `coverage.html`        | Interactive HTML coverage browser |
| `coverage_verbose.txt` | Full verbose test output log      |
| `test_report.md`       | This report                       |

---

## 4. Bugs Found & Fixed During Testing

| #   | Bug                                    | Root Cause                                                                                | Fix                                                                |
| --- | -------------------------------------- | ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| 1   | `ParseCtx()` nil pointer panic         | Tree-sitter binding expected non-nil `context.Context`                                    | Pass `context.Background()` in `pkg/treesitter/parser.go`          |
| 2   | Gob serialization crash on cache write | `[]ImportInfo`, `[]FunctionInfo`, `[]ClassInfo`, `map[string]any` not registered with gob | Added `gob.Register()` calls in `internal/cache/cache.go` `init()` |
| 3   | `.env` not loaded by CLI binary        | `os.Getenv()` doesn't read `.env` files                                                   | Added `godotenv.Load()` in `cmd/fastcode/main.go`                  |

---

## 5. E2E Test

The `run_e2e.sh` script performs an automated end-to-end validation:

1. Builds the `fastcode` binary from source
2. Creates a sandbox test repository with real source files
3. Runs `fastcode index` to parse, graph, and embed the repository
4. Runs `fastcode query` with a semantic question
5. Asserts the LLM-powered response contains expected keywords

**Result:** ✅ Passed — 93% confidence, 2 agent rounds, 9 elements retrieved
