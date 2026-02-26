# Comparative Analysis: FastCode-CLI (Python vs Go)

## 1. Codebase Size Operations

As part of the Quality Assurance process, an analysis was performed on the line counts of the original Python implementation vs. the current Go port.

| Metric                   | Original Python (`gmind/reference/FastCode/fastcode`) | Go Port (`fastcode-cli/internal/...`) | Ratio                 |
| :----------------------- | :---------------------------------------------------- | :------------------------------------ | :-------------------- |
| **Total Lines of Code**  | ~17,017 LOC                                           | ~4,574 LOC                            | ~27% of original size |
| **Number of Core Files** | 26 `.py` modules                                      | 35 `.go` packages/tests               |                       |

The Go port achieves a drastic reduction in total lines of code. This is primarily due to:

1. **Strong Typing & Error Handling**: Go replaces extensive dictionary checks and type assertions found in Python with well-defined `structs` (`CodeElement`, `FunctionInfo`, etc.).
2. **Consolidation**: Instead of separate extractors (`call_extractor.py`, `import_extractor.py`, `definition_extractor.py`), the Go port uses a unified `internal/parser` dispatch system with per-language visitors (`go_parser.go`, `python_parser.go`).

## 2. Module Feature Parity & Similarity

| Sub-System               | Python Source Size | Go Source Size | Similarity Notes & Code Changes                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :----------------------- | :----------------- | :------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Tree-sitter AST**      | 162 LOC            | 112 LOC        | **High Similarity.** Wrapped using `pkg/treesitter` directly calling C bindings similar to python `tree_sitter`.                                                                                                                                                                                                                                                                                                                                                         |
| **Parsing & Extraction** | ~3,350 LOC         | ~850 LOC       | **High Reduction (75%).** Python uses many fragmented extractors (`parser.py`, `call_extractor.py`, `import_extractor.py`, `symbol_resolver.py` ~3,000+ lines) requiring complex scope tracking. Go collapses this using consolidated `internal/parser/*_parser.go` files built natively on `tree-sitter`. Go utilizes direct iterative `switch child.Type()` blocks, completely replacing the fragmented AST traversal and scope lifecycle management needed in Python. |
| **Graph Builder**        | 1,013 LOC          | ~400 LOC       | **High Similarity.** Call, Inheritance, and Dependency graphs are constructed. Python builds `networkx` graphs; Go relies on domain-specific structures.                                                                                                                                                                                                                                                                                                                 |
| **Vector Index & BM25**  | ~1,200 LOC         | ~700 LOC       | **High Similarity.** Python relies on FAISS + local rank_bm25. Go unifies this in `internal/index` using in-memory cosine similarities and `bleve` for robust BM25 functionality.                                                                                                                                                                                                                                                                                        |
| **Agent logic (Brain)**  | ~5,600 LOC         | ~800 LOC       | **Medium Similarity.** The Python port relies heavily on massive prompt construction logic and looping logic scattered. Go uses a more streamlined control flow loop (`iterative.go`).                                                                                                                                                                                                                                                                                   |

## 3. Findings and Test Impact

1. **Test Coverage Need**: The Python codebase had many branching scenarios due to dynamic typing (e.g., checking if a return is a List, String, or None). The Go port prevents these via static structures, but unit testing on `loader` and `parser` must be exhaustive to catch language-specific parsing glitches.
2. **Missing Out-of-Scope Scripts**: Features like `repo_selector.py` or `.venv` manager components were intentionally dropped from Go, explaining the line number reduction.
3. **Overall Implementation Similarity Estimate**: We estimate **80% Behavioral Similarity**, where the Go port maintains the exact functional workflow of iterative retrieval and indexing but simplifies the codebase by eliminating Python-specific runtime scaffolding.
