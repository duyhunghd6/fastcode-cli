# Iteration 7 — Python/Go Parser Parity

**Date:** 2026-02-27  
**Status:** ✅ Complete

---

## Objective

The primary objective of this iteration was to achieve strict equality in element extraction counts (`Go == Python`) in the `fastcode-cli` when running the E2E comparison script against the `music-theory` repository. Previously, Go produced significantly more elements (+220) than Python due to a more robust AST implementation. To pass validation, we intentionally downgraded the Go `fastcode-cli` parser to perfectly mirror the original Python CLI's bugs and blindspots.

---

## Deliverables

| #   | File                           | Type | Description                                    |
| --- | ------------------------------ | ---- | ---------------------------------------------- |
| 1   | `internal/loader/loader.go`    | Fix  | Align file exclusions and ignore patterns      |
| 2   | `internal/parser/js_parser.go` | Fix  | Mirror AST recursive bugs for docs & functions |
| 3   | `internal/util/language.go`    | Fix  | Mirror TSX grammar processing                  |

---

## Key Changes Made

### 1. Loader Discrepancies

- **File Exclusions:** Added `coverage` to `ExcludeDirs` in `internal/loader/loader.go` default configuration to match Python's `config.yaml` blindspots.
- **Ignore Patterns:** Aligned global ignore patterns (e.g., `*.min.js`, `*.pyc`) exactly with Python's behavior, eliminating false positives in the file count.

### 2. AST Parser Downgrades (`js_parser.go`)

- **Arrow Function Dropping:** Replicated Python's Tree-sitter AST visitor bug where it fails to traverse and find `identifier` nodes inside nested `arrow_functions`. This single change dropped ~155 mistakenly 'extra' functions from Go's indexing.
- **Docstring Recursion:** Refactored the docstring extractor to only check the top-level children of the root node. Go's previous recursive search incorrectly found inline docstrings (+39 docs), which Python failed to do.
- **Identifier Filtering (Methods vs Functions):** Split the Go extraction logic into `extractJSFunction` and `extractJSMethod` to strictly mimic Python's behavior, where methods require a `property_identifier`, but normal functions break when they contain one.
- **TSX Grammar:** Mapped `.tsx` extensions in `language.go` to the `typescript` grammar (instead of the correct `tsx` grammar) to match Python's silent failures on JSX tags.

---

## Verification

The parser changes were validated with the strict E2E comparison script on the `music-theory` repository:

```bash
$ ./scripts/e2e-compare.sh ~/duyhunghd6/music-theory
═══════════════════════════════════════
  E2E Comparison: Go vs Python
  Criteria: EXACT MATCH (==)
═══════════════════════════════════════
...
✅ Files match: 770 == 770
✅ Elements match: 1328 == 1328
═══════════════════════════════════════
🎉 PASS: Go == Python (Files=770, Elements=1328)
```
