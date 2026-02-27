# Iteration 10 — BM25 IDF Math Fix & Broad E2E Verification

**Date:** 2026-02-27  
**Status:** ✅ Complete

---

## Objective

Fix remaining BM25 IDF calculation discrepancies between Go and Python that caused ranking divergences on certain repos. Verify exact prompt parity across all 6 reference repositories with the E2E query comparison script.

---

## Deliverables

| #   | File                        | Type    | Description                                      |
| --- | --------------------------- | ------- | ------------------------------------------------ |
| 1   | `internal/index/bm25.go`    | Fix     | Okapi IDF math, epsilon floor, stable sort       |
| 2   | `internal/index/indexer.go` | Fix     | File-element Docstring field for TF double-count |
| 3   | `internal/index/hybrid.go`  | Cleanup | Remove debug print statements                    |

---

## Key Changes Made

### 1. BM25 IDF & Scoring (`bm25.go`)

- **IDF:** Switched to `log((N - df + 0.5) / (df + 0.5))` to exactly match Python's `BM25Okapi`.
- **Epsilon floor:** Common terms with negative IDF now receive `epsilon * average_idf`, matching Python.
- **Stable sort:** Added index-based tie-breaking so identical scores resolve in the same order as Python.

### 2. File-Element Docstring (`indexer.go`)

Added `Docstring: pr.ModuleDocstring` in `addFileElement`. Python effectively double-counts module docstring tokens in BM25 by including them in both `Code` and `Docstring` fields. Go now mirrors this.

### 3. Debug Cleanup (`hybrid.go`)

Removed all `fmt.Printf` debug statements that were added during the BM25 investigation phase.

---

## Verification

E2E prompt comparison across all 6 reference repositories:

| Repository                    | Language  | Result            |
| ----------------------------- | --------- | ----------------- |
| `mcp_agent_mail`              | Python/TS | ✅ Perfect Match  |
| `beads`                       | Rust      | ✅ Perfect Match  |
| `beads_viewer`                | JS/TS     | ✅ Perfect Match  |
| `coding_agent_session_search` | Go/JS     | ✅ Perfect Match  |
| `zvec`                        | Go        | ✅ Perfect Match  |
| `FastCode`                    | Python    | ⚠️ Expected delta |

`FastCode` diverges because Python's `_expand_with_graph(max_hops=2)` only activates on Python codebases. Go's graph expansion is not yet implemented. All 5 non-Python repos confirmed zero-diff prompt parity.
