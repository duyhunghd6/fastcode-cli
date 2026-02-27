#!/bin/bash
# E2E Full-Flow Query Comparison: FastCode Go vs Python
# Captures ALL multi-step LLM calls (request + response) during the full iterative agent loop.
# Usage: ./scripts/e2e-fullflow-query-compare.sh /path/to/repo "your query here"
set -euo pipefail

REPO="${1:?Usage: $0 <repo-path> <query>}"
QUERY="${2:?Usage: $0 <repo-path> <query>}"
REPO=$(cd "$REPO" && pwd)

GO_CLI="$HOME/duyhunghd6/fastcode-cli"
PY_CLI="$HOME/duyhunghd6/gmind/reference/FastCode"

# Directories for full-flow prompt/response logs
GO_DIR="/tmp/fastcode_go_fullflow"
PY_DIR="/tmp/fastcode_py_fullflow"
rm -rf "$GO_DIR" "$PY_DIR"
mkdir -p "$GO_DIR" "$PY_DIR"

echo "════════════════════════════════════════════════"
echo "  E2E Full-Flow Query Comparison: Go vs Python"
echo "════════════════════════════════════════════════"
echo ""
echo "📦 Target: $REPO"
echo "❓ Query:  $QUERY"
echo ""

# ─── Step 1: Index with Go ───────────────────────────
echo "🔵 [Go] Indexing with --no-embeddings..."
unset FASTCODE_DEBUG_PROMPT_FILE 2>/dev/null || true
unset FASTCODE_DEBUG_PROMPT_DIR 2>/dev/null || true
GO_INDEX=$("$GO_CLI/fastcode" index "$REPO" --force --no-embeddings 2>&1 || true)
GO_FILES=$(echo "$GO_INDEX" | grep "Files:" | awk '{print $2}' || echo "?")
GO_ELEMENTS=$(echo "$GO_INDEX" | grep "Elements:" | awk '{print $2}' || echo "?")
echo "   Go Index: $GO_FILES files, $GO_ELEMENTS elements"
echo ""

# ─── Step 2: Full-flow query with Go ─────────────────
echo "🔵 [Go] Running full-flow query (logging ALL LLM calls)..."
export FASTCODE_DEBUG_PROMPT_DIR="$GO_DIR"
unset FASTCODE_DEBUG_PROMPT_FILE 2>/dev/null || true
$GO_CLI/fastcode query "$QUERY" --repo "$REPO" --no-embeddings 2>&1 || true
unset FASTCODE_DEBUG_PROMPT_DIR

GO_REQ_COUNT=$(ls "$GO_DIR"/call_*_request.json 2>/dev/null | wc -l | tr -d ' ')
GO_RESP_COUNT=$(ls "$GO_DIR"/call_*_response.json 2>/dev/null | wc -l | tr -d ' ')
echo "   ✅ Go captured: $GO_REQ_COUNT requests, $GO_RESP_COUNT responses"
echo ""

# ─── Step 3: Index with Python ───────────────────────
echo "🟡 [Python] Indexing with FASTCODE_NO_EMBEDDINGS=1..."
export FASTCODE_NO_EMBEDDINGS=1
unset FASTCODE_DEBUG_PROMPT_DIR 2>/dev/null || true
unset FASTCODE_DEBUG_PROMPT_FILE 2>/dev/null || true
(cd "$PY_CLI" && source .venv/bin/activate && python main.py index --repo-path "$REPO" --reindex </dev/null >/dev/null 2>&1 || true)
echo "   Python Index completed"
echo ""

# ─── Step 4: Full-flow query with Python ────────────
echo "🟡 [Python] Running full-flow query (logging ALL LLM calls)..."
export FASTCODE_NO_EMBEDDINGS=1
export FASTCODE_DEBUG_PROMPT_DIR="$PY_DIR"
unset FASTCODE_DEBUG_PROMPT_FILE 2>/dev/null || true
(cd "$PY_CLI" && source .venv/bin/activate && python main.py query --repo-path "$REPO" --query "$QUERY" </dev/null 2>&1 || true)
unset FASTCODE_DEBUG_PROMPT_DIR

PY_REQ_COUNT=$(ls "$PY_DIR"/call_*_request.json 2>/dev/null | wc -l | tr -d ' ')
PY_RESP_COUNT=$(ls "$PY_DIR"/call_*_response.json 2>/dev/null | wc -l | tr -d ' ')
echo "   ✅ Python captured: $PY_REQ_COUNT requests, $PY_RESP_COUNT responses"
echo ""

# ─── Summary ────────────────────────────────────────
echo "════════════════════════════════════════════════"
echo "  FULL-FLOW COMPARISON SUMMARY"
echo "════════════════════════════════════════════════"
echo ""
printf "%-12s  %-10s  %-10s\n" "" "Requests" "Responses"
printf "%-12s  %-10s  %-10s\n" "Go" "$GO_REQ_COUNT" "$GO_RESP_COUNT"
printf "%-12s  %-10s  %-10s\n" "Python" "$PY_REQ_COUNT" "$PY_RESP_COUNT"
echo ""

MAX_PAIRS=$(( GO_REQ_COUNT > PY_REQ_COUNT ? GO_REQ_COUNT : PY_REQ_COUNT ))
if [ "$MAX_PAIRS" -eq 0 ]; then
    echo "❌ No LLM calls captured. Check API key and configuration."
    exit 1
fi

echo "════════════════════════════════════════════════"
echo "  PAIR-BY-PAIR CALL COMPARISON"
echo "════════════════════════════════════════════════"

for i in $(seq 1 "$MAX_PAIRS"); do
    NUM=$(printf "%03d" "$i")
    GO_REQ="$GO_DIR/call_${NUM}_request.json"
    PY_REQ="$PY_DIR/call_${NUM}_request.json"
    GO_RESP="$GO_DIR/call_${NUM}_response.json"
    PY_RESP="$PY_DIR/call_${NUM}_response.json"

    echo ""
    echo "──── Call #$i ────────────────────────────────"

    # Check existence
    GO_REQ_EXISTS="❌"
    PY_REQ_EXISTS="❌"
    [ -f "$GO_REQ" ] && GO_REQ_EXISTS="✅"
    [ -f "$PY_REQ" ] && PY_REQ_EXISTS="✅"
    echo "  Request:  Go=$GO_REQ_EXISTS  Python=$PY_REQ_EXISTS"

    GO_RESP_EXISTS="❌"
    PY_RESP_EXISTS="❌"
    [ -f "$GO_RESP" ] && GO_RESP_EXISTS="✅"
    [ -f "$PY_RESP" ] && PY_RESP_EXISTS="✅"
    echo "  Response: Go=$GO_RESP_EXISTS  Python=$PY_RESP_EXISTS"

    # Show message counts for requests
    if [ -f "$GO_REQ" ]; then
        GO_MSGS=$(python3 -c "import json; d=json.load(open('$GO_REQ')); print(len(d.get('messages',[])))" 2>/dev/null || echo "?")
        GO_MODEL=$(python3 -c "import json; d=json.load(open('$GO_REQ')); print(d.get('model','?'))" 2>/dev/null || echo "?")
        echo "  Go Request:     model=$GO_MODEL, messages=$GO_MSGS"
    fi
    if [ -f "$PY_REQ" ]; then
        PY_MSGS=$(python3 -c "import json; d=json.load(open('$PY_REQ')); print(len(d.get('messages',[])))" 2>/dev/null || echo "?")
        PY_MODEL=$(python3 -c "import json; d=json.load(open('$PY_REQ')); print(d.get('model','?'))" 2>/dev/null || echo "?")
        PY_TOOLS=$(python3 -c "import json; d=json.load(open('$PY_REQ')); print(len(d.get('tools',[])))" 2>/dev/null || echo "0")
        echo "  Py Request:     model=$PY_MODEL, messages=$PY_MSGS, tools=$PY_TOOLS"
    fi

    # Show response preview
    if [ -f "$GO_RESP" ]; then
        GO_RESP_LEN=$(python3 -c "import json; d=json.load(open('$GO_RESP')); c=d.get('choices',[{}]); m=c[0].get('message',{}).get('content','') if c else d.get('content',''); print(len(str(m)))" 2>/dev/null || echo "?")
        echo "  Go Response:    $GO_RESP_LEN chars"
    fi
    if [ -f "$PY_RESP" ]; then
        PY_RESP_LEN=$(python3 -c "import json; d=json.load(open('$PY_RESP')); print(len(str(d.get('content',''))))" 2>/dev/null || echo "?")
        echo "  Py Response:    $PY_RESP_LEN chars"
    fi
done

echo ""
echo "════════════════════════════════════════════════"
echo "  FILE LOCATIONS"
echo "════════════════════════════════════════════════"
echo "  Go full-flow logs:     $GO_DIR/"
echo "  Python full-flow logs: $PY_DIR/"
echo ""
echo "📝 To deep-compare a specific call pair:"
echo "   diff <(python3 -m json.tool $GO_DIR/call_001_request.json) <(python3 -m json.tool $PY_DIR/call_001_request.json)"
echo ""
