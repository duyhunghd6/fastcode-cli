# Iteration 4 — CLI & MCP Integration

**Date:** 2026-02-26  
**Duration:** ~25 minutes  
**Status:** ✅ Complete

---

## Objective

Wire up all modules into a production CLI binary with Cobra commands and an MCP server for IDE integration.

---

## Deliverables

| #   | File                              | Package      | Lines | Description                                                         |
| --- | --------------------------------- | ------------ | ----- | ------------------------------------------------------------------- |
| 1   | `internal/orchestrator/engine.go` | orchestrator | 230   | End-to-end pipeline: load → index → cache → search → agent → answer |
| 2   | `cmd/fastcode/main.go`            | main         | 155   | Cobra CLI: `index`, `query`, `serve-mcp` commands                   |
| 3   | `cmd/fastcode/mcp.go`             | main         | 155   | MCP HTTP server: initialize, tools/list, tools/call endpoints       |

**Total Phase 4:** ~540 lines of Go

---

## CLI Output

```
$ ./fastcode --help
⚡ FastCode-CLI — Codebase Intelligence Engine

Usage:
  fastcode [command]

Available Commands:
  index       Index a local repository
  query       Query the indexed codebase
  serve-mcp   Start MCP (Model Context Protocol) server

Flags:
      --cache-dir string         Cache directory (default: ~/.fastcode/cache)
      --embedding-model string   Embedding model name (default "text-embedding-3-small")
      --no-embeddings            Skip embedding generation (BM25 only)
  -v, --version                  version for fastcode
```

## MCP Server Endpoints

| Endpoint          | Method | Description                                                    |
| ----------------- | ------ | -------------------------------------------------------------- |
| `/mcp/initialize` | POST   | Protocol handshake                                             |
| `/mcp/tools/list` | GET    | List available tools                                           |
| `/mcp/tools/call` | POST   | Execute a tool (index_repository, query_codebase, search_code) |
| `/health`         | GET    | Health check                                                   |

---

## Test Results

**All tests: 38/38 PASS** ✅ (unchanged from Phase 3 — Phase 4 adds integration points without new unit tests)

### Coverage

| Package          | Coverage  |
| ---------------- | --------- |
| `internal/agent` | **42.2%** |
| `internal/cache` | **84.6%** |
| `internal/graph` | **68.4%** |
| `internal/index` | **58.6%** |
| `internal/llm`   | **51.1%** |
| `internal/util`  | **60.0%** |

---

## Architecture Decisions

1. **Cobra CLI framework:** Standard Go CLI library. Provides auto-completion, help generation, flag parsing, and subcommand routing out of the box.

2. **Graceful fallback without API key:** When `OPENAI_API_KEY` is not set, the engine falls back to BM25-only search without LLM agent rounds, still providing useful results.

3. **MCP over HTTP (not stdio):** Chose HTTP transport for the MCP server instead of stdio for easier debugging, load balancing, and remote access. The protocol follows MCP spec 2024-11-05.

4. **Single binary:** All modules compile into one `fastcode` binary — no Python, no Docker, no runtime dependencies.

---

## Cumulative Project Summary

| Phase                    | Files  | LOC        | Tests  |
| ------------------------ | ------ | ---------- | ------ |
| Phase 1 (Parser + Graph) | 11     | ~1,556     | 11     |
| Phase 2 (Indexing)       | 7      | ~972       | 17     |
| Phase 3 (Agent)          | 4      | ~615       | 10     |
| Phase 4 (CLI + MCP)      | 3      | ~540       | 0      |
| **Total**                | **25** | **~3,683** | **38** |

### Dependencies

- `github.com/smacker/go-tree-sitter` — AST parsing
- `github.com/spf13/cobra` — CLI framework
- Go standard library (net/http, encoding/json, encoding/gob, crypto/sha256)
