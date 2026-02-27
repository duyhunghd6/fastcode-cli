# Iteration 12 — File & Element Parity (3 New Repos)

**Date:** 2026-02-27  
**Status:** ✅ 5/6 PASS, zvec +68 (irreducible grammar diff)

---

## Objective

Fix file and element count mismatches identified in iteration 11 for three new repositories: **beads_viewer**, **coding_agent_session_search**, and **zvec**. Ensure no regressions on previously passing repos (music-theory, beads, mcp_agent_mail).

---

## Deliverables

| #   | File                               | Type | Description                                                              |
| --- | ---------------------------------- | ---- | ------------------------------------------------------------------------ |
| 1   | `internal/util/language.go`        | Fix  | Remove `.cc`, `.cxx`, `.hpp`, `.scala` extensions                        |
| 2   | `internal/loader/loader.go`        | Fix  | Gitignore basename-level directory matching                              |
| 3   | `internal/parser/parser.go`        | Fix  | Pass language string to `parseC` for C vs C++ differentiation            |
| 4   | `internal/parser/other_parsers.go` | Fix  | Full C/C++ parser: module docstrings, function_declarator, ERROR skip    |
| 5   | `internal/parser/other_parsers.go` | Fix  | Full Rust parser: doc-comment-only module docstrings, impl_item handling |

---

## Key Changes Made

### 1. Extension Alignment (`language.go`)

- **Root Cause:** Go supported `.cc`, `.cxx`, `.hpp`, `.scala` which Python does not.
- **Fix:** Removed these extensions from `languageExtensions` map.
- **Impact:** zvec dropped from 787 → 431 files (exact match with Python).

### 2. Gitignore Basename-Level Directory Matching (`loader.go`)

- **Root Cause:** Patterns like `thirdparty` and `perf/` only matched at path prefix level.
- **Fix:** Enhanced `matchGitignorePattern` with basename-level matching at any depth.
- **Impact:** zvec excludes 14 `thirdparty/` files; coding_agent_session_search excludes `docs/perf/baseline_round1.md`.

### 3. C/C++ Parser Rewrite (`other_parsers.go`)

- Dedicated `visitCNode`/`visitCNodeAtDepth`, `extractCClass`, `extractCFunction`
- Module docstrings from leading `//` or `/* */` comments
- `function_definition` requires `function_declarator` child (matching Python)
- Classes extract methods from `field_declaration_list`
- **Root-level ERROR node skip**: Go's tree-sitter C grammar wraps C++ content in `ERROR` nodes (recursed into), while Python wraps in `function_definition` (elif stops recursion). Skipping root ERROR nodes reduces zvec from +437 → +68.

### 4. Rust Parser Rewrite (`other_parsers.go`)

- Dedicated `visitRustNode`, `extractRustType`, `extractRustFunction`
- Doc-comment-only module docstrings (`///`, `//!`, `/* */`) — skips regular `//` comments
- `struct_item`/`trait_item`/`impl_item` → class (no recursion, elif chain)
- Methods embedded in class info; only top-level `function_item` → function elements
- **Impact:** coding_agent_session_search fixed from -14 → **0** (exact match).

---

## Verification

```
═══════════════════════════════════════
🎉 PASS: Go == Python (Files=770, Elements=1328)      # music-theory
🎉 PASS: Go == Python (Files=934, Elements=1457)      # beads
🎉 PASS: Go == Python (Files=222, Elements=2466)      # mcp_agent_mail
🎉 PASS: Go == Python (Files=3613, Elements=4570)     # beads_viewer ← NEW
🎉 PASS: Go == Python (Files=499, Elements=10146)     # coding_agent_session_search ← NEW
❌ FAIL: Go=3200 != Python=3132 (Go has +68 extra)    # zvec (+437→+68, ↓85%)
═══════════════════════════════════════
```

---

## Remaining: zvec +68

Irreducible tree-sitter C grammar version mismatch between Go's `smacker/go-tree-sitter` (v2024-08) and Python's `tree-sitter-c`.

**Verified root cause:** Go and Python produce different parse trees for C++ `.h` files parsed with C grammar:

- `vector_array.h` ERROR node: **0 fn_defs in Python, 23 in Go** — same source, different ASTs
- `flat_searcher.h` ERROR node: **16 fn_defs in Python, 18 in Go** — similar but not identical

Would require matching exact `tree-sitter-c` grammar versions to fully eliminate.
