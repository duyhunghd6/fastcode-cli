# Iteration 11 — Iterative Agent Parity & Retrieval Quality Fix

**Date:** 2026-02-27  
**Status:** ✅ Complete

---

## Objective

Achieve functional parity between Go and Python iterative agents, then fix a critical retrieval quality bug where Go retrieved wrong files (coverage HTMLs, `.claude/` docs) instead of actual source code.

---

## Deliverables

| #   | File                                    | Type    | Description                                          |
| --- | --------------------------------------- | ------- | ---------------------------------------------------- |
| 1   | `internal/agent/iterative.go`           | Major   | 2-phase prompt architecture, LLM file selection      |
| 2   | `internal/agent/tools.go`               | Major   | Filesystem-based search_codebase, FileCandidate type |
| 3   | `internal/agent/answer.go`              | Fix     | System/user message split, tuned LLM params          |
| 4   | `internal/llm/client.go`                | Fix     | Multi-message support, temperature/max_tokens parity |
| 5   | `internal/orchestrator/engine.go`       | Wire    | Pass repoRoot to ToolExecutor                        |
| 6   | `cmd/fastcode/main.go`                  | Feature | `--version` with build time (GMT+7) and git commit   |
| 7   | `scripts/e2e-fullflow-query-compare.sh` | New     | Full-flow Go vs Python LLM call comparison           |

---

## Key Changes Made

### 1. Two-Phase Prompt Architecture (`iterative.go`)

- Split `buildRoundPrompt` → `buildRound1Prompt` (assessment, no code context) + `buildRoundNPrompt` (cost-aware with elements)
- Split `parseRoundResponse` → `parseRound1Response` + `parseRoundNResponse`
- Added `initializeAdaptiveParams()` matching Python's adaptive confidence/budget logic
- Added `keep_files` filtering and tool call history tracking
- LLM params: temperature 0.1→0.2, max_tokens 2000→8000, confidence threshold 85→95

### 2. Real Filesystem Search (`tools.go`)

- **Root cause fix**: Go's `search_codebase` was doing BM25 on indexed elements (returning coverage HTML). Now does real `os.Walk` + regex grep like Python's `agent_tools.py`
- Added `ExecuteSearchCodebase()`, `ExecuteListDirectory()`, `FindElementsForFile()`
- Skips `.git`, `node_modules`, `coverage/`, `.claude/`, `.kilocode/`

### 3. LLM File Selection (`iterative.go`)

- Added `llmSelectFiles()` — sends file candidates to LLM to pick most relevant files
- Matches Python's `_llm_select_elements_with_granularity()` pipeline
- Flow: tool calls → filesystem search → LLM selection → indexed elements

### 4. Version Flag (`main.go`)

- `fastcode --version` → `0.1.0 (built: 2026-02-27 19:25:03 GMT+7, commit: c52a7a2)`
- Build time embedded via `-ldflags` at compile time

---

## Verification

### Unit Tests

All agent tests pass: `go test ./internal/agent/... -count=1` ✅

### E2E Index Parity (Go == Python)

| Repository                    | Go Files | Go Elems | Py Files | Py Elems | Result  |
| ----------------------------- | -------- | -------- | -------- | -------- | ------- |
| `music-theory`                | 770      | 1,328    | 770      | 1,328    | ✅ PASS |
| `beads`                       | 934      | 1,457    | 934      | 1,457    | ✅ PASS |
| `mcp_agent_mail`              | 222      | 2,466    | 222      | 2,466    | ✅ PASS |
| `beads_viewer`                | 3,613    | 4,544    | 3,613    | 4,570    | ❌ -26  |
| `coding_agent_session_search` | 500      | 9,961    | 499      | 10,146   | ❌ -185 |
| `zvec`                        | 787      | 2,602    | 431      | 3,132    | ❌ -530 |

### E2E Query Quality ("how is audio played?" on music-theory)

| Metric          | Go (Before)      | Go (After)           | Python     |
| --------------- | ---------------- | -------------------- | ---------- |
| Confidence      | 62%              | **96%**              | 96%        |
| Rounds          | 5                | 3                    | 2          |
| Stop            | max_rounds       | confidence           | confidence |
| Files Retrieved | coverage HTML ❌ | `audio-engine.ts` ✅ | Same ✅    |
| Answer Quality  | Wrong            | Correct              | Correct    |
