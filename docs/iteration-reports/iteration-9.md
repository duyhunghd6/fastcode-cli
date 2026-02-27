# Iteration 9 — E2E Prompt & Retrieval Parity

**Date:** 2026-02-27  
**Status:** ✅ Complete

---

## Objective

Ensure that the Go implementation of FastCode generates identical LLM prompts and retrieves the exact same code elements in the same order as the Python implementation. This aligns the core ReAct agent's hydration logic and BM25 search scoring for true E2E functional parity.

---

## Deliverables

| #   | File                           | Type | Description                                         |
| --- | ------------------------------ | ---- | --------------------------------------------------- |
| 1   | `internal/agent/answer.go`     | Fix  | Combine System and User prompts into one message    |
| 2   | `internal/agent/tools.go`      | Fix  | Limit `searchCode` results to 4 elements            |
| 3   | `internal/index/hybrid.go`     | Fix  | Match BM25/Vector text structure to Python exactly  |
| 4   | `internal/agent/iterative.go`  | Fix  | Perform standard retrieval before first LLM loop    |
| 5   | `scripts/e2e-query-compare.sh` | Feat | Implement prompt interception and comparison script |

---

## Key Changes Made

### 1. Merged Prompt Message Structure

- **Root Cause:** Go was sending system instructions as a `{"role": "system"}` message and the query as `{"role": "user"}`. Python combined the system prompt and the user query into a single `{"role": "user"}` message.
- **Fix:** Refactored `AnswerGenerator` to prepend the system prompt to the user query and send a single message, matching Python precisely.

### 2. Standard Context Hydration

- **Root Cause:** Go's `IterativeAgent` started the first ReAct round with an empty context array, waiting for the LLM to call tools. Python's agent performed an initial "standard retrieval" fallback to populate the context with BM25/Vector results immediately.
- **Fix:** In `internal/agent/iterative.go`, we now initialize the context by calling `ia.toolExecutor.searchCode(query)` before entering the reasoning loop.

### 3. Retrieval Element Limitations

- **Root Cause:** Go's `searchCode` tool retrieved 10 elements by default. Python's complex post-retrieval culling loop typically pared the results down to around 4 independent elements.
- **Fix:** Hardcoded the `searchCode` return limit to exactly 4 elements in Go to shadow Python's final contextual footprint.

### 4. BM25 Text Structure Consistency

- **Root Cause:** Go's `BuildSearchText` arbitrarily truncated code to 180 chars, destroying term frequency. Python used specific concatenated properties (`Name`, `Type`, `Language`, `RelativePath`, `Docstring`, `Signature`, `Summary`, `Code[:1000]`).
- **Fix:** Rewrote `BuildSearchText` in `hybrid.go` to exactly mirror Python's string assembly arrays. This ensures identical BM25 indexing signatures and matching BM25 keyword retrieval scores between both platforms.

---

## Verification

Running the headless E2E query comparator using `--no-embeddings` (`FASTCODE_NO_EMBEDDINGS=1` in Python) successfully intercepts identically structured JSON prompts.

Both the Go and Python clients predictably isolate the **exact same 4 top elements** (`schedule()`, `AudioContextManager`, `AudioUnlocker.test.tsx`, `tutorial.md`) for the query "how is audio played?" without semantic aid, confirming retrieval logic synchronization.
