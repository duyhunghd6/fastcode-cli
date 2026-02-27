# Iteration 8 — Cross-Repository Element Parity

**Date:** 2026-02-27  
**Status:** ✅ Complete

---

## Objective

Extend strict Go==Python element parity from a single repo (music-theory) to three additional repositories: **beads** (Go/Rust project), **mcp_agent_mail** (Python project), and validate no regressions on music-theory. This required identifying and fixing four distinct behavioral differences between Go and Python parsers/indexers.

---

## Deliverables

| #   | File                               | Type | Description                                            |
| --- | ---------------------------------- | ---- | ------------------------------------------------------ |
| 1   | `internal/parser/parser.go`        | Fix  | Remove Go language from code-language detection        |
| 2   | `internal/parser/python_parser.go` | Fix  | Scan past shebangs for module docstrings               |
| 3   | `internal/parser/python_parser.go` | Fix  | Add maxFunctionLines=1000 limit                        |
| 4   | `internal/index/indexer.go`        | Fix  | Skip empty files (match Python's `if not c: continue`) |
| 5   | 7 test files                       | Test | Update assertions for new parser behavior              |

---

## Key Changes Made

### 1. Go Language Parsing Removed (`parser.go`)

- **Root Cause:** Python treats `.go` files as unsupported (`_parse_generic` → 0 functions/classes). Go had full tree-sitter extraction producing thousands of extra elements.
- **Fix:** Removed `"go"` from `isCodeLanguage()` and `case "go":` from `ParseFile` switch.
- **Impact:** beads repo dropped from 6898 → 1434 elements (exact match).

### 2. Module Docstring Shebang Handling (`python_parser.go`)

- **Root Cause:** `parsePython` only checked `root.Child(0)` for docstrings. Files starting with `#!/usr/bin/env python3` had comment nodes first, pushing docstrings to later children.
- **Fix:** Scan first 15 root children, skipping `comment` nodes, before checking for `expression_statement` with string.
- **Impact:** Fixed 3 missing documentation elements in beads (agent.py, generate-newsletter.py, test_multi_repo.py).

### 3. Max Function Lines Limit (`python_parser.go`)

- **Root Cause:** Python's `_extract_python_function` skips functions >1000 lines (`max_function_lines=1000`). Go had no such limit, extracting mega-functions like `build_http_app` (3031 lines).
- **Fix:** Added `maxFunctionLines = 1000` constant; functions exceeding this return empty and are skipped.
- **Impact:** mcp_agent_mail dropped from 2469 → 2466 elements (exact match).

### 4. Empty File Skipping (`indexer.go`)

- **Root Cause:** Python skips empty files with `if not c: continue`. Go indexed them, creating spurious file elements.
- **Fix:** Added `if content == "" { continue }` after reading file content.
- **Impact:** mcp_agent_mail `tests/__init__.py` (0 bytes) no longer indexed.

---

## Verification

All three E2E comparisons pass with strict equality:

```
═══════════════════════════════════════
🎉 PASS: Go == Python (Files=222, Elements=2466)   # mcp_agent_mail
🎉 PASS: Go == Python (Files=911, Elements=1434)   # beads
🎉 PASS: Go == Python (Files=770, Elements=1328)   # music-theory
═══════════════════════════════════════
```

Unit tests: all parser, index, and orchestrator packages pass.
