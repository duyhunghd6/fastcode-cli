#!/bin/bash
# E2E Comparison: FastCode Go vs Python (Strict Equality)
# Usage: ./scripts/e2e-compare.sh /path/to/repo
#
# Pass/Fail Criteria (STRICT):
#   Files:    Go_files    == Python_files    → ✅ PASS
#   Elements: Go_elements == Python_elements → ✅ PASS
#   Any mismatch → ❌ FAIL
set -euo pipefail

REPO="${1:?Usage: $0 <repo-path>}"
REPO=$(cd "$REPO" && pwd)  # resolve to absolute path

GO_CLI="$HOME/duyhunghd6/fastcode-cli"
PY_CLI="$HOME/duyhunghd6/gmind/reference/FastCode"

echo "═══════════════════════════════════════"
echo "  E2E Comparison: Go vs Python"
echo "  Criteria: EXACT MATCH (==)"
echo "═══════════════════════════════════════"
echo ""
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
echo "═══════════════════════════════════════"
PASS=true

# Check files
if [ "$GO_FILES" -ne "$PY_FILES" ]; then
  echo "❌ FAIL: File count mismatch (Go=$GO_FILES, Python=$PY_FILES)"
  PASS=false
else
  echo "✅ Files match: $GO_FILES == $PY_FILES"
fi

# Check elements (strict equality)
if [ "$GO_ELEMENTS" -ne "$PY_ELEMENTS" ]; then
  DIFF=$((GO_ELEMENTS - PY_ELEMENTS))
  if [ "$DIFF" -gt 0 ]; then
    echo "❌ FAIL: Element count mismatch (Go=$GO_ELEMENTS != Python=$PY_ELEMENTS, Go has +$DIFF extra)"
  else
    ABS_DIFF=$(( -DIFF ))
    echo "❌ FAIL: Element count mismatch (Go=$GO_ELEMENTS != Python=$PY_ELEMENTS, Go is missing $ABS_DIFF)"
  fi
  PASS=false
else
  echo "✅ Elements match: $GO_ELEMENTS == $PY_ELEMENTS"
fi

echo "═══════════════════════════════════════"
if [ "$PASS" = true ]; then
  echo "🎉 PASS: Go == Python (Files=$GO_FILES, Elements=$GO_ELEMENTS)"
  exit 0
else
  echo "💥 OVERALL: FAIL"
  exit 1
fi
