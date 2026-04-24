---
description: Run E2E comparison tests between FastCode-CLI (Go) and CodeGraph (Rust) or Python reference
---

# E2E Comparison Workflow

This workflow runs end-to-end comparison tests between FastCode-CLI implementations.

## Prerequisites

- Go 1.21+ installed
- FastCode-CLI Go binary buildable at `~/duyhunghd6/fastcode-cli`
- Python reference at `~/duyhunghd6/gmind/reference/FastCode` (for Go-vs-Python comparison)
- A target repository to index (e.g., `~/duyhunghd6/music-theory`)

---

## Part A: Go vs Python E2E Comparison

### 1. Build the Go binary

```bash
cd /Users/steve/duyhunghd6/fastcode-cli && go build -o fastcode ./cmd/fastcode/
```

### 2. Set the target repository

Ask the user which repository to test against. Default: `/Users/steve/duyhunghd6/music-theory`

### 3. Run the E2E comparison script

```bash
/Users/steve/duyhunghd6/fastcode-cli/scripts/e2e-compare.sh <repo-path>
```

If the script doesn't exist, create it using the template in the **Script Reference** section below.

### 4. Evaluate results

**Pass criteria (STRICT EQUALITY):**

| Metric   | Condition                                  |
| -------- | ------------------------------------------ |
| Files    | `Go_files == Python_files` → ✅ PASS       |
| Elements | `Go_elements == Python_elements` → ✅ PASS |

Any mismatch → ❌ FAIL. Investigate using the Troubleshooting section.

### 5. If tests fail — investigate element breakdown

Run each indexer separately and compare element types (file, class, function, documentation):

- **Go:** `./fastcode index <repo> --force --no-embeddings --json`
- **Python:** Run Python indexer inline (see script reference)

Report differences per element type in a table.

---

## Part B: Go vs Rust (CodeGraph) Feature Comparison

This is a **manual feature audit**, not a script. Compare capabilities across these dimensions:

1. **Parsing & Language Support** — active languages, AST visitors, complexity analysis
2. **Indexing & Storage** — store type, graph DB, incremental indexing, LSP
3. **Search & Retrieval** — BM25, vector, hybrid, reranking, graph traversal
4. **Embedding Providers** — count and variety
5. **LLM / Agent Architecture** — strategies, tools, context handling
6. **MCP / Integration** — protocol, daemon, HTTP/GraphQL
7. **Performance & Engineering** — allocator, SIMD, zero-copy, mmap, GPU
8. **Configuration & Security** — config format, secrets management

Refer to the **Feature Matrix** section below for the full comparison table.

---

## Troubleshooting

| Issue                                    | Fix                                                                                                                                                    |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Python `.venv` not found                 | `cd ~/duyhunghd6/gmind/reference/FastCode && pyenv local 3.11 && python -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt` |
| Go build fails                           | `cd ~/duyhunghd6/fastcode-cli && go mod tidy && go build -o fastcode ./cmd/fastcode/`                                                                  |
| `AttributeError: 'NoneType'` on embedder | The script bypasses the embedder. If using `indexer.index_repository()` directly, set `embedder=None`.                                                 |
| Element count < Python                   | Check: missing extensions in `language.go`? Gitignore excluding too much? Compare file counts first.                                                   |

---

## Script Reference: `scripts/e2e-compare.sh`

// turbo-all

```bash
#!/bin/bash
# Usage: ./scripts/e2e-compare.sh /path/to/repo
set -euo pipefail

REPO="${1:?Usage: $0 <repo-path>}"
REPO=$(cd "$REPO" && pwd)

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

# --- Judge (STRICT EQUALITY) ---
echo ""
echo "═══════════════════════════════════"
PASS=true
if [ "$GO_FILES" -ne "$PY_FILES" ]; then
  echo "❌ FAIL: File count mismatch (Go=$GO_FILES, Python=$PY_FILES)"
  PASS=false
else
  echo "✅ Files match: $GO_FILES == $PY_FILES"
fi
if [ "$GO_ELEMENTS" -ne "$PY_ELEMENTS" ]; then
  DIFF=$((GO_ELEMENTS - PY_ELEMENTS))
  echo "❌ FAIL: Element count mismatch (Go=$GO_ELEMENTS != Python=$PY_ELEMENTS, diff=$DIFF)"
  PASS=false
else
  echo "✅ Elements match: $GO_ELEMENTS == $PY_ELEMENTS"
fi
echo "═══════════════════════════════════"
if [ "$PASS" = true ]; then
  echo "🎉 PASS: Go == Python"
  exit 0
else
  echo "💥 OVERALL: FAIL"
  exit 1
fi
```

---

## Feature Matrix: Go vs Rust

> [!NOTE]
> Go covers roughly **30–40%** of CodeGraph-Rust's feature surface, focused on the critical indexing + querying path.

| Category            | FastCode-CLI (Go)          | CodeGraph (Rust)                              | Gap         |
| ------------------- | -------------------------- | --------------------------------------------- | ----------- |
| Parsing & Languages | 8 active langs             | 11 active + FastML                            | −3, −FastML |
| Storage & Indexing  | BoltDB + BM25              | SurrealDB graph + HNSW + 3 tiers + LSP        | Major       |
| Search & Retrieval  | BM25 + optional vectors    | Hybrid 70/30 + graph + reranking              | Significant |
| Embedding Providers | 1 (OpenAI-compatible)      | 5 (OpenAI, Ollama, Jina, LM Studio, ONNX)     | −4          |
| Agent Architecture  | Single ReAct loop, 4 tools | 3 strategies (LATS/ReAct/Reflexion), 10 tools | Major       |
| MCP / Integration   | TCP server                 | stdio + HTTP + daemon + GraphQL               | Significant |
| Performance         | Standard Go                | jemalloc + SIMD + zero-copy + mmap + GPU      | Major       |
| Codebase Scale      | ~13K LOC                   | ~100K LOC (7.5×)                              | —           |
