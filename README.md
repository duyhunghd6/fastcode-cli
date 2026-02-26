<div align="center">

# ⚡ FastCode-CLI

### A Go-native Codebase Intelligence Engine

Inspired by [HKUDS/FastCode](https://github.com/HKUDS/FastCode) — Rewritten in Go for speed, portability, and single-binary deployment.

[![Go 1.23+](https://img.shields.io/badge/go-1.23+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tiếng Việt](https://img.shields.io/badge/🇻🇳_Tiếng_Việt-README--vi.md-red)](README-vi.md)

**[Features](#-features)** • **[Quick Start](#-quick-start)** • **[Architecture](#-architecture)** • **[Roadmap](#-roadmap)** • **[Credits](#-credits)**

</div>

---

## 🎯 What is FastCode-CLI?

FastCode-CLI is a **high-performance, token-efficient code understanding engine** written in Go. It parses, indexes, and navigates large codebases using AST analysis, hybrid search (semantic + BM25), and multi-layer graph modeling — all from a single compiled binary.

It is designed for:

- **AI Agent workflows** — Provide structured, budget-aware code context to LLMs without overwhelming context windows.
- **Developer tooling** — Quickly understand unfamiliar codebases, trace dependencies, and locate code.
- **MCP Server integration** — Plug directly into Cursor, Claude Code, Windsurf, or any MCP-compatible client.

---

## ✨ Features

### 🏗️ Semantic-Structural Code Representation

- **AST Parsing** via [go-tree-sitter](https://github.com/smacker/go-tree-sitter) — Multi-level indexing across files, classes, functions, and documentation for **8+ languages** (Go, Python, JavaScript, TypeScript, Java, Rust, C/C++, C#).
- **Hybrid Index** — Combines dense vector embeddings with [Bleve](https://github.com/blevesearch/bleve) BM25 keyword search for precise and robust code retrieval.
- **Multi-Layer Graph Modeling** — Three interconnected relationship graphs (Call Graph, Dependency Graph, Inheritance Graph) for structural navigation.

### 🧭 Lightning-Fast Navigation

- **Two-Stage Smart Search** — First finds potentially relevant code, then ranks the best matches for your specific query.
- **Code Skimming** — Reads only function signatures, class definitions, and type hints instead of full files, saving massive amounts of tokens.
- **Graph Traversal** — Traces code connections up to N hops away, following imports, calls, and inheritance chains.

### 💰 Cost-Efficient Context Management

- **Budget-Aware Decision Making** — Weighs confidence, query complexity, codebase size, and token cost before processing.
- **Value-First Selection** — Prioritizes high-impact, low-cost information first, like picking the ripest fruit at the best price.

### 🚀 Go Advantages

- **Single Binary** — No Python, no pip, no venv, no Docker. Just one fast binary.
- **Goroutine Concurrency** — Parallel AST parsing and HTTP embedding calls turn a 20s Python index into a 2s Go index.
- **Tiny Memory Footprint** — No PyTorch, no FAISS pickle blobs. Just lean Go + Bleve.

---

## 🚀 Quick Start

### Install from Source

```bash
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli
go build -o fastcode ./cmd/fastcode

# Configure your LLM endpoint
export OPENAI_API_KEY="your-key"
export MODEL="gpt-4o"
export BASE_URL="https://api.openai.com/v1"
```

### Usage

```bash
# Index a local repository
fastcode index /path/to/your/repo

# Query the indexed codebase
fastcode query "How does the authentication flow work?"

# Multi-repo query
fastcode query --repos /path/repo1,/path/repo2 "Where is the payment logic?"

# Start as MCP server (for Cursor / Claude Code)
fastcode serve-mcp --port 8080
```

---

## 🏗 Architecture

```
┌─────────────────────────────────────────────────┐
│                  fastcode-cli                    │
├─────────────┬───────────────┬───────────────────┤
│  cmd/       │  internal/    │  pkg/             │
│  fastcode   │  parser       │  treesitter       │
│  (Cobra)    │  graph        │                   │
│             │  index        │                   │
│             │  agent        │                   │
│             │  llm          │                   │
└─────────────┴───────────────┴───────────────────┘
        │              │               │
   CLI/MCP      AST + Graph      Tree-sitter
   Interface    Engine           Go Bindings
        │              │               │
        ▼              ▼               ▼
   ┌─────────┐  ┌───────────┐  ┌─────────────┐
   │ LLM API │  │ Bleve BM25│  │ Vector Store│
   │ (OpenAI │  │ (Keyword  │  │ (Embeddings)│
   │ /Ollama)│  │  Search)  │  │             │
   └─────────┘  └───────────┘  └─────────────┘
```

### Package Layout

| Package           | Description                                                                 |
| ----------------- | --------------------------------------------------------------------------- |
| `cmd/fastcode`    | CLI entry point (Cobra), subcommands: `index`, `query`, `serve-mcp`         |
| `internal/parser` | Tree-sitter AST parsing, code unit extraction (functions, classes, imports) |
| `internal/graph`  | Call Graph, Dependency Graph, Inheritance Graph construction & traversal    |
| `internal/index`  | Hybrid indexing engine (vector embeddings + BM25 via Bleve)                 |
| `internal/agent`  | Iterative retrieval agent with budget-aware context gathering               |
| `internal/llm`    | LLM client abstraction (OpenAI-compatible API)                              |
| `pkg/treesitter`  | Tree-sitter Go bindings and language grammar helpers                        |
| `reference/`      | Original Python FastCode source code for reference during porting           |
| `docs/`           | Research documents, analysis, and porting plans                             |

---

## 🗺 Roadmap

### Phase 1: Core Engine _(In Progress)_

- [ ] Tree-sitter AST parsing for Go, Python, JS/TS, Java, Rust
- [ ] Code unit extraction (functions, classes, imports, types)
- [ ] Call Graph and Dependency Graph construction

### Phase 2: Indexing

- [ ] LLM-based embedding generation (via OpenAI / Ollama API)
- [ ] Bleve BM25 text indexing for keyword search
- [ ] Hybrid retrieval (vector + BM25 fusion)

### Phase 3: Retrieval Agent

- [ ] Budget-aware iterative agent (port from Python `IterativeAgent`)
- [ ] Code skimming and smart file browsing
- [ ] Multi-repo query support

### Phase 4: Integration

- [ ] CLI commands: `index`, `query`, `summary`
- [ ] MCP Server mode (`serve-mcp`)
- [ ] REST API server mode

---

## 🙏 Credits

This project is a **Go rewrite** inspired by [**FastCode**](https://github.com/HKUDS/FastCode) by the [HKUDS Lab](https://github.com/HKUDS) at The University of Hong Kong. The original Python implementation introduced the groundbreaking three-phase framework for token-efficient code understanding.

We gratefully acknowledge the original authors and their research contributions.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

<div align="center">

**Built with ❤️ in Go**

_Part of the [Gmind](https://github.com/duyhunghd6/gmind) ecosystem — Memory Management for Agentic Coding_

</div>
