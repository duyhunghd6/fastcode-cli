# E2E Comparison: FastCode-CLI (Go) vs CodeGraph (Rust)

Comprehensive feature and architecture comparison between our **Go** implementation and the **CodeGraph-Rust** reference project.

---

## At a Glance

| Dimension         | FastCode-CLI (Go)            | CodeGraph (Rust)            |
| ----------------- | ---------------------------- | --------------------------- |
| **Language**      | Go 1.21+                     | Rust (edition 2021)         |
| **Codebase Size** | ~13,353 LOC / 52 files       | ~99,918 LOC / 260 files     |
| **Architecture**  | Single binary, flat packages | 14-crate workspace monorepo |
| **License**       | MIT                          | MIT / Apache-2.0            |

---

## 1. Parsing & Language Support

| Capability                          | FastCode-CLI (Go)                                                     | CodeGraph (Rust)                                                   | Verdict       |
| ----------------------------------- | --------------------------------------------------------------------- | ------------------------------------------------------------------ | ------------- |
| Parser Backend                      | `go-tree-sitter` bindings                                             | `tree-sitter` Rust bindings                                        | 🤝 Same       |
| **Active Languages**                | Go, Python, JS/TS/TSX, Java, Rust, C/C++ (**8**)                      | Rust, TS, JS, Python, Go, Java, C++, Swift, C#, Ruby, PHP (**11**) | ⚠️ Go lacks 3 |
| Disabled Languages                  | C#, Ruby, PHP, Swift, Kotlin, Scala listed as "code" but no extractor | Kotlin, Dart (tree-sitter version conflict)                        | —             |
| Non-code file indexing              | ✅ Markdown, JSON, YAML, etc. as file-level elements                  | ✅ Docs/spec nodes linked to symbols in README & `docs/**/*.md`    | 🤝 Both       |
| FastML (heuristic pattern matching) | ❌ None                                                               | ✅ `fast_ml/` pattern matcher + symbol resolver + enhancer         | 🔴 **Go gap** |
| AST Visitors                        | Single-pass per-language extractors                                   | `visitor.rs` (26 KB) + per-language extractors                     | 🤝 Similar    |
| Complexity Analysis                 | ❌                                                                    | ✅ `complexity.rs` — cyclomatic/cognitive metrics                  | 🔴 **Go gap** |
| Diff / Incremental Parsing          | ❌                                                                    | ✅ `diff.rs` (23 KB) + `watcher.rs` (35 KB)                        | 🔴 **Go gap** |
| Semantic Analysis                   | ❌                                                                    | ✅ `semantic.rs` (31 KB) — cross-file semantic linking             | 🔴 **Go gap** |

---

## 2. Indexing & Storage

| Capability                  | FastCode-CLI (Go)                    | CodeGraph (Rust)                                                                                             | Verdict           |
| --------------------------- | ------------------------------------ | ------------------------------------------------------------------------------------------------------------ | ----------------- |
| Index Store                 | **BoltDB** (embedded key-value)      | **SurrealDB** (graph DB + HNSW vector index)                                                                 | 🔴 **Go simpler** |
| Index Tiers                 | Single mode (BM25 ± embeddings)      | `fast` / `balanced` / `full` (progressively richer)                                                          | 🔴 **Go gap**     |
| Graph Database              | ❌ No graph                          | ✅ Full knowledge graph — nodes, edges, graph traversals                                                     | 🔴 **Go gap**     |
| Graph Functions             | Basic adjacency via `graph/` package | SurrealQL: `fn::get_transitive_dependencies`, `fn::trace_call_chain`, `fn::calculate_coupling_metrics`, etc. | 🔴 **Go gap**     |
| Incremental Indexing        | ✅ File-hash cache                   | ✅ Incremental module (`incremental/`) + file watcher daemon                                                 | 🤝 Basic parity   |
| LSP Integration             | ❌                                   | ✅ `balanced`/`full` tiers use LSP (rust-analyzer, pyright, gopls, etc.)                                     | 🔴 **Go gap**     |
| Module / Dataflow Edges     | ❌                                   | ✅ Module nodes, import/containment edges, `defines`/`uses`/`flows_to`/`mutates`                             | 🔴 **Go gap**     |
| Architecture Boundary Rules | ❌                                   | ✅ `codegraph.boundaries.toml` — `violates_boundary` edges                                                   | 🔴 **Go gap**     |

---

## 3. Search & Retrieval

| Capability                | FastCode-CLI (Go)                                  | CodeGraph (Rust)                                                   | Verdict       |
| ------------------------- | -------------------------------------------------- | ------------------------------------------------------------------ | ------------- |
| BM25 Keyword Search       | ✅ Custom in-memory BM25                           | ✅ Lexical component (30% weight)                                  | 🤝 Both       |
| Vector / Embedding Search | ✅ Optional (OpenAI-compatible API)                | ✅ HNSW vector index in SurrealDB (70% weight)                     | 🤝 Both       |
| **Hybrid Search**         | BM25 + cosine similarity (when embeddings enabled) | 70% vector + 30% lexical + graph traversal + optional reranking    | ⚠️ Go simpler |
| Reranking                 | ❌                                                 | ✅ Cross-encoder reranker (`reranker.rs`, `reranking/`)            | 🔴 **Go gap** |
| Graph Traversal in Search | ❌                                                 | ✅ Relationship-aware results (callers, dependencies, containment) | 🔴 **Go gap** |

---

## 4. Embedding Providers

| Provider              | FastCode-CLI (Go) | CodeGraph (Rust) |
| --------------------- | ----------------- | ---------------- |
| OpenAI-compatible API | ✅                | ✅               |
| Ollama                | ❌                | ✅               |
| Jina AI               | ❌                | ✅               |
| LM Studio             | ❌                | ✅               |
| ONNX Runtime (local)  | ❌                | ✅               |

**Verdict:** 🔴 Go supports only 1 provider vs Rust's 5.

---

## 5. LLM / Agent Architecture

| Capability                  | FastCode-CLI (Go)                        | CodeGraph (Rust)                                                                                                               | Verdict           |
| --------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ----------------- |
| Agent Framework             | Custom iterative retrieval agent         | **Rig** (native Rust) + legacy `react`/`lats` via AutoAgents                                                                   | 🔴 **Go simpler** |
| Agent Tools                 | `search`, `browse`, `skim`, `list_files` | 4 agentic tools (`agentic_context`, `agentic_impact`, `agentic_architecture`, `agentic_quality`) backed by 6 inner graph tools | 🔴 **Go gap**     |
| Agent Strategies            | Single ReAct-style loop                  | **LATS** (tree search), **ReAct** (linear), **Reflexion** (auto-recovery) — auto-selected per task                             | 🔴 **Go gap**     |
| Context Window Awareness    | ❌                                       | ✅ Tier-aware prompting (3–8 steps based on context window size)                                                               | 🔴 **Go gap**     |
| Context Overflow Protection | ❌                                       | ✅ Per-tool truncation + accumulation guard                                                                                    | 🔴 **Go gap**     |
| LLM Providers               | OpenAI-compatible (1)                    | Anthropic, OpenAI, xAI Grok, Ollama, LM Studio, OpenAI-compatible (**6+**)                                                     | 🔴 **Go gap**     |
| Answer Generation           | LLM synthesis from retrieved context     | Multi-step agent reasoning → structured JSON output with `highlights`, `next_steps`                                            | 🔴 **Go gap**     |

---

## 6. MCP / Integration

| Capability              | FastCode-CLI (Go)                          | CodeGraph (Rust)                                                 | Verdict       |
| ----------------------- | ------------------------------------------ | ---------------------------------------------------------------- | ------------- |
| MCP Protocol            | ✅ `serve-mcp` command (JSON-RPC over TCP) | ✅ `start stdio` / `start http` — full MCP server (`rmcp` crate) | 🤝 Both       |
| Daemon Mode             | ❌                                         | ✅ File watcher daemon with debounced re-indexing                | 🔴 **Go gap** |
| HTTP / GraphQL API      | ❌                                         | ✅ Axum HTTP server + async-graphql + Swagger/OpenAPI docs       | 🔴 **Go gap** |
| IDE Integration Targets | Generic MCP                                | Claude Code, Cursor, and any MCP-compatible client               | 🤝 Similar    |

---

## 7. Performance & Engineering

| Capability              | FastCode-CLI (Go)   | CodeGraph (Rust)                                                          | Verdict        |
| ----------------------- | ------------------- | ------------------------------------------------------------------------- | -------------- |
| Memory Allocator        | Go GC (default)     | **jemalloc** (tikv-jemallocator)                                          | Rust advantage |
| Parallelism             | `goroutines`        | `rayon` + `crossbeam` + `tokio` (async)                                   | 🤝 Both strong |
| Zero-copy Serialization | ❌                  | ✅ `rkyv` via `codegraph-zerocopy` crate                                  | 🔴 **Go gap**  |
| SIMD Vector Operations  | ❌                  | ✅ `simd_ops.rs` — hardware-accelerated similarity                        | 🔴 **Go gap**  |
| Memory-mapped I/O       | ❌                  | ✅ `mmap.rs`, `memmap2`                                                   | 🔴 **Go gap**  |
| Compression             | ❌                  | ✅ zstd + lz4 + flate2                                                    | 🔴 **Go gap**  |
| Build Profiles          | Standard `go build` | 6 profiles: `dev`, `fast-dev`, `test`, `bench`, `release`, `release-size` | —              |
| GPU Support             | ❌                  | ✅ `gpu.rs` for embedding acceleration                                    | 🔴 **Go gap**  |

---

## 8. Configuration & Security

| Capability           | FastCode-CLI (Go)                  | CodeGraph (Rust)                                                  | Verdict       |
| -------------------- | ---------------------------------- | ----------------------------------------------------------------- | ------------- |
| Config Format        | `~/.fastcode/config.yaml` + `.env` | `~/.codegraph/config.toml` + env vars + project `.codegraph.toml` | 🤝 Both       |
| Secrets Management   | Env vars only                      | ✅ `chacha20poly1305` encryption + `argon2` + `secrecy` crate     | 🔴 **Go gap** |
| `.gitignore` Respect | ✅                                 | ✅ + additional secrets pattern filtering                         | 🤝 Both       |

---

## Summary Scorecard

| Category            | Go Features             | Rust Features                            | Gap                   |
| ------------------- | ----------------------- | ---------------------------------------- | --------------------- |
| Parsing & Languages | 8 active                | 11 active + FastML                       | −3 languages, −FastML |
| Storage & Indexing  | BoltDB + BM25           | SurrealDB graph + HNSW + 3 tiers + LSP   | Major                 |
| Search & Retrieval  | BM25 + optional vectors | Hybrid 70/30 + graph + reranking         | Significant           |
| Embedding Providers | 1                       | 5                                        | −4                    |
| Agent Architecture  | Single loop, 4 tools    | 3 strategies, 4 agentic tools + 6 inner  | Major                 |
| MCP / Integration   | TCP server              | stdio + HTTP + daemon + GraphQL          | Significant           |
| Performance         | Standard Go             | jemalloc + SIMD + zero-copy + mmap + GPU | Major                 |
| Codebase Scale      | ~13K LOC                | ~100K LOC (**7.5× larger**)              | —                     |

---

## Key Takeaway

> **FastCode-CLI (Go)** is a lean, focused tool (~13K LOC) that covers the **core indexing + querying workflow** well — parse code, build BM25 index, optionally generate embeddings, and answer questions via an iterative retrieval agent.
>
> **CodeGraph (Rust)** is a **full-stack code intelligence platform** (~100K LOC) that adds a real knowledge graph, multi-strategy agentic reasoning, 5+ embedding providers, LSP integration, daemon mode, and extensive performance engineering (SIMD, zero-copy, mmap, GPU).
>
> The Go version covers roughly **30–40% of CodeGraph's feature surface**, focused on the critical path. The biggest gaps are: **graph database**, **multi-strategy agents**, **LSP integration**, and **performance engineering**.

---

## Previous E2E Test: Go vs Python (Reference)

The original E2E comparison between Go and the Python reference implementation is preserved below for reference.

### Pass/Fail Criteria

```
IF Go_elements >= Python_elements → ✅ PASS
IF Go_elements <  Python_elements → ❌ FAIL
Files must match exactly: Go_files == Python_files
```

### Example Run (music-theory)

```
📦 Target: /Users/steve/duyhunghd6/music-theory

🔵 Indexing with Go...
   Go: 770 files, 1499 elements
🟡 Indexing with Python...
   Python: 770 files, 1328 elements

═══════════════════════════════════
✅ PASS: Files=770, Go=1499 >= Python=1328 (+171)
```

### One-Liner Script

Save as `scripts/e2e-compare.sh`:

```bash
#!/bin/bash
# Usage: ./scripts/e2e-compare.sh /path/to/repo
set -euo pipefail

REPO="${1:?Usage: $0 <repo-path>}"
REPO=$(cd "$REPO" && pwd)  # resolve to absolute path

GO_CLI="$HOME/duyhunghd6/fastcode-cli"
PY_CLI="$HOME/duyhunghd6/gmind/reference/FastCode"

echo "📦 Target: $REPO"
echo ""

# --- Go ---
echo "🔵 Indexing with Go..."
GO_OUT=$("$GO_CLI/fastcode" index "$REPO" --force --no-embeddings 2>&1)
GO_FILES=$(echo "$GO_OUT" | grep "Files:" | awk '{print $2}')
GO_ELEMENTS=$(echo "$GO_OUT" | grep "Elements:" | awk '{print $2}')
echo "   Go: $GO_FILES files, $GO_ELEMENTS elements"

# --- Python ---
echo "🟡 Indexing with Python..."
PY_OUT=$(cd "$PY_CLI" && source .venv/bin/activate && python -c "
import yaml, logging
from fastcode.loader import RepositoryLoader
from fastcode.parser import CodeParser
from fastcode.indexer import CodeIndexer
logging.disable(logging.CRITICAL)
with open('config/config.yaml') as f:
    config = yaml.safe_load(f)
loader = RepositoryLoader(config)
loader.load_from_path('$REPO')
files = loader.scan_files()
parser = CodeParser(config.get('parser', {}))
indexer = CodeIndexer.__new__(CodeIndexer)
indexer.config = config
indexer.loader = loader
indexer.parser = parser
indexer.embedder = None
indexer.vector_store = None
indexer.logger = logging.getLogger(__name__)
indexer.levels = config.get('indexing', {}).get('levels', ['file', 'class', 'function', 'documentation'])
indexer.include_imports = config.get('indexing', {}).get('include_imports', True)
indexer.include_class_context = config.get('indexing', {}).get('include_class_context', True)
indexer.generate_overview = False
indexer.elements = []
indexer.current_repo_name = 'test'
indexer.current_repo_url = None
for fi in files:
    c = loader.read_file_content(fi['path'])
    if not c: continue
    pr = parser.parse_file(fi['path'], c)
    if pr: indexer._index_file(fi, c, pr)
print(f'{len(files)} {len(indexer.elements)}')
" 2>&1 | tail -1)
PY_FILES=$(echo "$PY_OUT" | awk '{print $1}')
PY_ELEMENTS=$(echo "$PY_OUT" | awk '{print $2}')
echo "   Python: $PY_FILES files, $PY_ELEMENTS elements"

# --- Judge ---
echo ""
echo "═══════════════════════════════════"
if [ "$GO_FILES" -ne "$PY_FILES" ]; then
  echo "❌ FAIL: File count mismatch (Go=$GO_FILES, Python=$PY_FILES)"
  exit 1
elif [ "$GO_ELEMENTS" -lt "$PY_ELEMENTS" ]; then
  echo "❌ FAIL: Go elements < Python ($GO_ELEMENTS < $PY_ELEMENTS)"
  exit 1
else
  DIFF=$((GO_ELEMENTS - PY_ELEMENTS))
  echo "✅ PASS: Files=$GO_FILES, Go=$GO_ELEMENTS >= Python=$PY_ELEMENTS (+$DIFF)"
  exit 0
fi
```

### Troubleshooting

| Issue                                    | Fix                                                                                                                                                            |
| ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Python `.venv` not found                 | `cd ~/duyhunghd6/gmind/reference/FastCode && pyenv local 3.11 && python -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt`         |
| Go build fails                           | `cd ~/duyhunghd6/fastcode-cli && go mod tidy && go build -o fastcode ./cmd/fastcode/`                                                                          |
| `AttributeError: 'NoneType'` on embedder | The script above bypasses the embedder. If using `indexer.index_repository()` directly, pass `--no-embeddings` or set `embedder=None`.                         |
| Element count < Python                   | Check: missing extensions in `language.go`? Gitignore excluding too much? Run `./fastcode index <repo> --force --no-embeddings` and compare file counts first. |
