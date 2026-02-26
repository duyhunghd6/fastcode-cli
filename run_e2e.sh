#!/bin/bash

# Exit on error
set -e

echo "==========================================="
echo "   FastCode CLI - End to End Test Suite    "
echo "==========================================="

test_dir="test_repo_snapshot"
cache_dir="$HOME/.fastcode/cache"

# Build the fastcode binary
echo "[1] Building the fastcode binary..."
go build -o fastcode cmd/fastcode/*.go
if [ ! -f "fastcode" ]; then
  echo "Error: Binary build failed."
  exit 1
fi

# Clean previous cache specifically for the test repo
echo "[2] Cleaning old index cache..."
rm -f "$cache_dir/${test_dir}.gob"

# Create a tiny test repository snapshot (the codebase itself might be too big for a quick E2E)
echo "[3] Preparing tiny test repository..."
mkdir -p "$test_dir"
cp internal/loader/loader.go "$test_dir/loader.go"
cp internal/parser/parser.go "$test_dir/parser.go"

# 1. Test Indexing
echo "[4] Running Indexing on test repository..."
./fastcode index "$test_dir"

# 2. Test Querying
echo "[5] Running Semantic Query..."
query_string="how does the file loader filter by file size?"
echo "Query: $query_string"
./fastcode query "$query_string" --repo "$test_dir" > query_result.txt

# Display results
cat query_result.txt

# Validation
if grep -qi "MaxFileSize" query_result.txt; then
  echo "✅ E2E Test Passed: Semantic query successfully retrieved information."
else
  echo "❌ E2E Test Failed: Expected keyword 'MaxFileSize' not found in the output."
  exit 1
fi

# Clean up test artifacts
rm -rf "$test_dir"
rm query_result.txt
echo "==========================================="
echo "   E2E Test Suite Completed Successfully   "
echo "==========================================="
