# FastCode-Go (fastcode-cli) Porting Plan

## 1. Objective

To rewrite the Python-based `FastCode` framework into a high-performance, statically compiled Go-based CLI tool (`fastcode-cli` or integrated into `gmind`). This aligns completely with the project's Go tech stack, eliminating the need to maintain a separate Python environment and Docker container for code comprehension.

## 2. Python to Go Architecture Mapping

The current Python FastCode architecture is highly modular. Here is how we map its core files to Go packages:

| Python Module (`fastcode/`)                                   | Go Package / Library Strategy (`internal/...`) | Description                                                                                                                                                                                                                                       |
| :------------------------------------------------------------ | :--------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `tree_sitter_parser.py`                                       | `github.com/smacker/go-tree-sitter`            | Core AST parsing. We will need the C-bindings for Go tree-sitter, along with grammar support for target languages (Go, Python, Java, JS, Rust, etc.).                                                                                             |
| `parser.py`, `definition_extractor.py`, `import_extractor.py` | `internal/parser`                              | Traverse the AST nodes to extract classes, functions, variables, and imports using tree-sitter queries.                                                                                                                                           |
| `graph_builder.py`, `module_resolver.py`                      | `internal/graph`                               | Build the Call Graph, Dependency Graph, and Inheritance Graph. This can be stored in-memory using a Go graph library (e.g., `github.com/dominikbraun/graph`).                                                                                     |
| `embedder.py`                                                 | Local LLM Embeddings via API                   | Python uses `sentence-transformers` locally. In Go, we should rely on external embedding APIs via the `gmind` LLM integration (e.g., OpenAI `text-embedding-3-small` or local Ollama endpoints) to avoid bundling heavy ML frameworks like Torch. |
| `indexer.py`, `vector_store.py`                               | `internal/index` (Zvec DB integration)         | Python uses FAISS + BM25(Pickle). In Go, we will push the AST embedded nodes directly into our **Zvec DB** (C++ core / local vector store) and implement a simple Go-based BM25 text indexer.                                                     |
| `iterative_agent.py`, `retriever.py`, `query_processor.py`    | `internal/agent`                               | The core "budget-aware algorithm" that decides how to search the codebase. This will be pure Go business logic utilizing LLM API calls.                                                                                                           |
| `api.py`, `mcp_server.py`                                     | `cmd/fastcode-cli` or `cmd/gmind`              | A CLI interface (using `cobra`) and potentially a fast Go MCP server using existing Go MCP SDKs.                                                                                                                                                  |

## 3. Core Go Dependencies

To achieve the port, we will pull in the following reliable Go frameworks:

1. **CLI & Routing**: `github.com/spf13/cobra` (for `fastcode-cli`)
2. **AST Parsing**: `github.com/smacker/go-tree-sitter`
   - Need parsers: `go-tree-sitter/golang`, `go-tree-sitter/python`, `go-tree-sitter/javascript`, etc.
3. **Graph Algorithms**: `github.com/dominikbraun/graph` (For Call/Dependency graphs)
4. **Vector Storage**: Direct integration with the project's **Zvec DB** (via CGO or local socket) OR pure Go alternative like `github.com/milvus-io/milvus-sdk-go` if applicable.
5. **Text Search (BM25)**: `github.com/blevesearch/bleve/v2` (A powerful Go-native search and indexing library, an excellent replacement for Python's `rank_bm25`).
6. **LLM/Embeddings Interface**: Use standard REST/HTTP clients to talk to OpenRouter/OpenAI/Ollama format endpoints.

## 4. Phased Implementation Strategy

### Phase 1: Core Parsing & Graphing (The Engine)

1. Initialize the Go module (`go mod init fastcode`).
2. Implement `internal/parser`: Wrap `go-tree-sitter` to read a codebase directory and emit strongly-typed Go structs (`ASTNode`, `FunctionDef`, `ClassDef`).
3. Implement `internal/graph`: Connect the emitted nodes into Dependency and Call Graphs.

### Phase 2: Indexing (The Database)

1. Implement embedding calls (HTTP out to LLM) for each `ASTNode`.
2. Connect to `Zvec DB` to store the dense vectors.
3. Integrate `bleve` to index the raw source texts and docstrings to provide the BM25 capability.

### Phase 3: The Retrieval Agent (The Brain)

1. Port the Python `IterativeAgent` logic into Go.
2. Implement the token-limiting and context-gathering heuristics (the "budget-aware" and "file-skimming" functions).

### Phase 4: CLI & Integration

1. Build `cmd/fastcode-cli` with commands like `index`, `query`, and `serve-mcp`.
2. Integrate this functionality natively into `gmind` so that Agent workflows can call it.

## 5. Why Go is Better for FastCode

- **Speed & Concurrency**: Go's goroutines will heavily parallelize the codebase AST parsing and HTTP embedding calls, turning a 20-second Python index into a 2-second Go index.
- **Single Binary**: No need to manage `uv`, `pip`, `venv`, or large Docker images natively running PyTorch and CUDA just for embeddings—just one fast Go binary (`gmind` / `fastcode-cli`).
- **Memory Footprint**: Python FAISS + Pickle + Torch consumes GBs of RAM. Go + Bleve + Zvec will require a fraction of that, keeping local dev environments snappy.
