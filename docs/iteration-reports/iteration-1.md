# Iteration 1 — Core Engine (Parser + Graph)

**Date:** 2026-02-26  
**Duration:** ~1 hour  
**Status:** ✅ Complete

---

## Objective

Port the foundational layer of [HKUDS/FastCode](https://github.com/HKUDS/FastCode) to Go: tree-sitter-based AST parsing, code element extraction, and relationship graph construction.

---

## Deliverables

| #   | File                               | Package    | Lines | Description                                                                               |
| --- | ---------------------------------- | ---------- | ----- | ----------------------------------------------------------------------------------------- |
| 1   | `internal/types/types.go`          | types      | 75    | Core structs: `CodeElement`, `FunctionInfo`, `ClassInfo`, `ImportInfo`, `FileParseResult` |
| 2   | `internal/util/language.go`        | util       | 57    | Language detection from file extensions (10+ languages)                                   |
| 3   | `internal/util/path.go`            | util       | 60    | Path normalization, module path conversion, line extraction                               |
| 4   | `internal/loader/loader.go`        | loader     | 170   | Repository directory walker with `.gitignore`, size filtering, extension filtering        |
| 5   | `pkg/treesitter/parser.go`         | treesitter | 112   | Thread-safe tree-sitter wrapper with multi-language caching                               |
| 6   | `internal/parser/parser.go`        | parser     | 70    | Main dispatcher routing files to language-specific extractors                             |
| 7   | `internal/parser/go_parser.go`     | parser     | 240   | Go AST extraction: functions, methods, structs, interfaces, imports, docstrings           |
| 8   | `internal/parser/python_parser.go` | parser     | 205   | Python AST extraction: classes, functions, imports, decorators, docstrings                |
| 9   | `internal/parser/js_parser.go`     | parser     | 225   | JS/TS AST extraction: classes, arrow functions, imports, exports                          |
| 10  | `internal/parser/other_parsers.go` | parser     | 82    | Generic recursive visitor stubs for Java, Rust, C/C++                                     |
| 11  | `internal/graph/graph.go`          | graph      | 260   | Dependency, Inheritance, Call graph builder with multi-hop traversal                      |

**Total:** ~1,556 lines of Go (source only, excluding tests)

---

## Test Results

```
=== RUN   TestGraphAddEdge          --- PASS (0.00s)
=== RUN   TestGraphNoDuplicateEdges --- PASS (0.00s)
=== RUN   TestGraphNoSelfLoop       --- PASS (0.00s)
=== RUN   TestGetRelatedElements    --- PASS (0.00s)
=== RUN   TestGenerateElementID     --- PASS (0.00s)
=== RUN   TestBuildInheritanceGraph --- PASS (0.00s)
=== RUN   TestGetLanguageFromPath   --- PASS (0.00s)
=== RUN   TestIsSupportedFile       --- PASS (0.00s)
=== RUN   TestCountLines            --- PASS (0.00s)
=== RUN   TestExtractLines          --- PASS (0.00s)
=== RUN   TestFilePathToModulePath  --- PASS (0.00s)
```

**Result: 11/11 PASS** ✅

### Coverage

| Package           | Coverage            | Target |
| ----------------- | ------------------- | ------ |
| `internal/graph`  | **68.4%**           | ≥ 80%  |
| `internal/util`   | **60.0%**           | ≥ 80%  |
| `internal/loader` | 0.0% (no tests yet) | ≥ 80%  |
| `internal/parser` | 0.0% (no tests yet) | ≥ 80%  |
| `pkg/treesitter`  | 0.0% (no tests yet) | ≥ 80%  |

> **Note:** Parser and loader tests require `testdata/` sample files which will be added in the next iteration.

### Build

```
$ go build ./...
# Success — zero errors, zero warnings
```

---

## Architecture Decisions

1. **go-tree-sitter module pinning:** The `github.com/smacker/go-tree-sitter/javascript` submodule conflicted with the main module. Resolved by excluding `v0.0.1` and using the bundled version from the main `go-tree-sitter` module.

2. **Generic parser fallback:** Instead of leaving Java/Rust/C parsers empty, a `visitGenericNode()` recursive visitor was implemented. It catches `function_definition`, `class_declaration`, `struct_item` etc. across languages — providing partial extraction even before dedicated parsers are ported.

3. **Go-specific receiver parsing:** The Go parser handles method receivers (`func (s *Server) Start()`) by extracting the receiver type separately, which Python's AST doesn't need. This is stored in `FunctionInfo.Receiver`.

4. **Graph without external library:** Instead of using `dominikbraun/graph`, the graph was implemented with simple adjacency lists (`map[string][]string`). This eliminates a dependency and keeps the code minimal for the current needs (BFS traversal, edge lookup).

---

## Known Gaps (To Address in Iteration 2+)

| Gap                                                 | Priority | Planned Iteration |
| --------------------------------------------------- | -------- | ----------------- |
| No `testdata/` sample repos for parser/loader tests | P1       | Iteration 2       |
| Java/Rust/C parsers are generic stubs only          | P2       | Iteration 3       |
| No CLI commands wired yet (stub `main.go` only)     | P2       | Iteration 4       |
| No embedding/indexing/BM25 search                   | P1       | Iteration 2       |
| No LLM client for agent queries                     | P2       | Iteration 3       |

---

## Python → Go Porting Notes

| Python Concept          | Go Equivalent              | Notes                                                        |
| ----------------------- | -------------------------- | ------------------------------------------------------------ |
| `@dataclass`            | `struct` with JSON tags    | Go structs with `json:"..."` tags replace Python dataclasses |
| `Dict[str, Any]`        | `map[string]any`           | Go 1.18+ `any` alias for `interface{}`                       |
| `Optional[str]`         | `string` (zero value `""`) | Go uses zero values instead of `None`; `omitempty` in JSON   |
| `networkx.DiGraph`      | Custom `Graph` struct      | Simple adjacency list; no need for full networkx in Go       |
| `tree_sitter.Parser`    | `sitter.Parser` via CGO    | `go-tree-sitter` wraps the C library via CGO                 |
| `rank_bm25`             | Bleve (planned)            | Will use `blevesearch/bleve` for BM25 in Iteration 2         |
| `sentence-transformers` | HTTP API call              | Go won't embed PyTorch; will call embedding APIs externally  |
