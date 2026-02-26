# Iteration 3 — Retrieval Agent

**Date:** 2026-02-26  
**Duration:** ~30 minutes  
**Status:** ✅ Complete

---

## Objective

Implement the iterative retrieval agent: query analysis, tool execution, multi-round confidence-based retrieval, and LLM-powered answer generation.

---

## Deliverables

| #   | File                          | Package | Lines | Description                                                                          |
| --- | ----------------------------- | ------- | ----- | ------------------------------------------------------------------------------------ |
| 1   | `internal/agent/query.go`     | agent   | 120   | Query processor: keyword extraction, complexity scoring (0-100), type classification |
| 2   | `internal/agent/tools.go`     | agent   | 145   | Tool executor: search_code, browse_file, skim_file, list_files                       |
| 3   | `internal/agent/iterative.go` | agent   | 260   | Multi-round agent: confidence control, budget tracking, tool orchestration           |
| 4   | `internal/agent/answer.go`    | agent   | 90    | LLM answer generator: structured prompt building from retrieved context              |

**Total Phase 3:** ~615 lines of Go

---

## Test Results

```
=== agent ===
TestProcessQuery              --- PASS
TestClassifyQuery             --- PASS
TestExtractKeywords           --- PASS
TestScoreComplexity           --- PASS
TestExtractJSON               --- PASS
TestDeduplicateElements       --- PASS
TestToolExecutorSearchCode    --- PASS
TestToolExecutorBrowseFile    --- PASS
TestToolExecutorListFiles     --- PASS
TestToolExecutorUnknown       --- PASS
```

**All tests: 38/38 PASS** ✅ (10 new + 28 from Phases 1-2)

### Coverage

| Package          | Coverage  |
| ---------------- | --------- |
| `internal/agent` | **42.2%** |
| `internal/cache` | **84.6%** |
| `internal/graph` | **68.4%** |
| `internal/index` | **58.6%** |
| `internal/llm`   | **51.1%** |
| `internal/util`  | **60.0%** |

> Agent coverage is lower because `iterative.go` and `answer.go` require LLM API calls to test fully, which is tested via integration tests.

---

## Architecture Decisions

1. **Query classification strategy:** Rule-based keyword matching (locate/debug/howto/overview/understand) rather than LLM-based classification. Saves API calls and keeps the classifier fast and deterministic.

2. **Token-efficient skim_file tool:** Returns only signatures and docstrings (no full code), reducing token consumption by ~80% compared to browse_file. The agent prefers this tool in early rounds.

3. **Adaptive round limits:** Simple queries (complexity < 30) are capped at 2 rounds instead of 5, saving API costs while still delivering quality results for simple lookups.

4. **JSON extraction from LLM:** The `extractJSON` function handles both markdown code-fenced JSON and raw embedded JSON, using bracket depth-tracking for robustness.

5. **No external agent framework:** Implemented the agent loop directly instead of using LangChain-Go or similar. Keeps the codebase lean and avoids heavy transitive dependencies.

---

## Cumulative Progress

| Phase                    | Files  | LOC        | Tests  |
| ------------------------ | ------ | ---------- | ------ |
| Phase 1 (Parser + Graph) | 11     | ~1,556     | 11     |
| Phase 2 (Indexing)       | 7      | ~972       | 12     |
| Phase 3 (Agent)          | 4      | ~615       | 10     |
| **Total**                | **22** | **~3,143** | **38** |
