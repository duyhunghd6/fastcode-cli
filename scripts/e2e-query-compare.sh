#!/bin/bash
# E2E Query Comparison: FastCode Go vs Python (Prompt structure diff)
# Usage: ./scripts/e2e-query-compare.sh /path/to/repo "your query here"
set -euo pipefail

REPO="${1:?Usage: $0 <repo-path> <query>}"
QUERY="${2:?Usage: $0 <repo-path> <query>}"
REPO=$(cd "$REPO" && pwd)

GO_CLI="$HOME/duyhunghd6/fastcode-cli"
PY_CLI="$HOME/duyhunghd6/gmind/reference/FastCode"

# Temporary files for prompt interception
GO_PROMPT="/tmp/fastcode_go_prompt.json"
PY_PROMPT="/tmp/fastcode_py_prompt.json"
rm -f "$GO_PROMPT" "$PY_PROMPT"

echo "═══════════════════════════════════════"
echo "  E2E Query Prompt Comparison: Go vs Python"
echo "═══════════════════════════════════════"
echo ""
echo "📦 Target: $REPO"
echo "❓ Query:  $QUERY"
echo ""

# --- Step 1: Index with Go ---
echo "🔵 [Go] Indexing with --no-embeddings..."
export FASTCODE_DEBUG_PROMPT_FILE=""
GO_INDEX=$("$GO_CLI/fastcode" index "$REPO" --force --no-embeddings 2>&1 || true)
GO_FILES=$(echo "$GO_INDEX" | grep "Files:" | awk '{print $2}' || echo "?")
GO_ELEMENTS=$(echo "$GO_INDEX" | grep "Elements:" | awk '{print $2}' || echo "?")
echo "   Go Index: $GO_FILES files, $GO_ELEMENTS elements"
echo ""

# --- Step 2: Query with Go (Intercept) ---
echo "🔵 [Go] Generating prompt..."
export FASTCODE_DEBUG_PROMPT_FILE="$GO_PROMPT"
# we know the answer generation will fail if it's a dummy response but the first tools call prompt will be dumped
$GO_CLI/fastcode query "$QUERY" --repo "$REPO" --no-embeddings --json >/dev/null 2>&1 || true

if [ ! -f "$GO_PROMPT" ]; then
    echo "❌ [Go] Failed to intercept prompt!"
else
    GO_MSG_COUNT=$(python3 -c "import sys,json; d=json.load(open('$GO_PROMPT')); print(len(d.get('messages',[])))" 2>/dev/null || echo "0")
    echo "   ✅ Prompt intercepted: $GO_MSG_COUNT messages"
fi
echo ""

# --- Step 3: Query with Python (Intercept) ---
echo "🟡 [Python] Generating prompt..."
export FASTCODE_NO_EMBEDDINGS=1
export FASTCODE_LOG_LEVEL=DEBUG
export FASTCODE_DEBUG_PROMPT_FILE="$PY_PROMPT"
(cd "$PY_CLI" && source .venv/bin/activate && python main.py query --repo-path "$REPO" --query "$QUERY" || true)

if [ ! -f "$PY_PROMPT" ]; then
    echo "❌ [Python] Failed to intercept prompt!"
else
    PY_MSG_COUNT=$(python3 -c "import sys,json; d=json.load(open('$PY_PROMPT')); print(len(d.get('messages',[])))" 2>/dev/null || echo "0")
    echo "   ✅ Prompt intercepted: $PY_MSG_COUNT messages"
fi
echo ""

# --- Summary & Diff ---
echo "═══════════════════════════════════════"
echo "  PROMPT COMPARISON (First Message)"
echo "═══════════════════════════════════════"

if [[ -f "$GO_PROMPT" && -f "$PY_PROMPT" ]]; then
    echo "🔵 Go System Message (Lines 1-15):"
    python3 -c "import sys,json; d=json.load(open('$GO_PROMPT')); msgs=d.get('messages',[]); print(msgs[0].get('content','') if msgs else 'None')" | head -n 15
    echo "..."
    echo ""
    echo "🟡 Python System Message (Lines 1-15):"
    python3 -c "import sys,json; d=json.load(open('$PY_PROMPT')); msgs=d.get('messages',[]); print(msgs[0].get('content','') if msgs else 'None')" | head -n 15
    echo "..."
    echo ""
    echo "📝 Compare full structured JSONs manually if needed:"
    echo "   Go:     $GO_PROMPT"
    echo "   Python: $PY_PROMPT"
else
    echo "❌ Missing prompt files. Could not compare."
fi
