# FastCode Research & Integration Proposal for Gmind

## 1. FastCode Overview

**FastCode** is a token-efficient, high-performance framework designed for comprehensive codebase understanding, architecture exploration, and QA. It acts as a powerful structural and semantic search engine for the codebase, significantly outperforming traditional Vector RAG systems and context-heavy IDE agents (like Cursor or Claude Code).

## 2. Core Capabilities

FastCode relies on a three-phase architecture that perfectly aligns with modern multi-agent coding workflows:

1. **Semantic-Structural Representation (Graph RAG & Hybrid Index):**
   - **AST Parsing:** Multi-level indexing (files, classes, functions, docs) via AST parsing (supports Python, JS, Go, Rust, C++, C#, Java).
   - **Hybrid Retrieval:** Combines semantic embedding search with BM25 keyword search.
   - **Multi-Layer Graph Modeling:** Maintains 3 interconnected graphs: Call Graph, Dependency Graph, and Inheritance Graph, allowing structural code navigation (following imports, calls, etc.).
2. **Lightning-Fast Navigation:** Two-stage smart search that limits context size by browsing files safely and reading only "headlines" (signatures/stubs) instead of full files.
3. **Versatile Interfaces:**
   - **MCP Server:** Provides robust tools (`code_qa`, `list_indexed_repos`, etc.) for Claude/Cursor out of the box.
   - **REST API:** Easily callable by external CLI tools (like `gmind`).
   - **Web UI & Feishu Nanobot:** Excellent for human-in-the-loop and team collaboration.
   - **Local Model Support:** Deployable with open-source/local models (e.g., Qwen3 Coder).

## 3. Why Replace `codegraph` / `codegraph-rust` with FastCode?

The current Gmind PRDs (PRD-01/02/03) describe a vision where **Zvec DB** combined with **Tree-sitter AST Nodes** acts as a Graph RAG system. Building this from scratch (`codegraph` or `codegraph-rust`) is a monumental task requiring parsing logic, hybrid indexing algorithms, and graph traversal schemas.

**FastCode provides this exact architecture out of the box.**

- **Avoid reinventing the wheel:** FastCode already implements Tree-sitter AST to graph extraction + BM25 + Embeddings.
- **MCP Native:** FastCode exposes an MCP server ready to be consumed by `mcp_agent_mail` or Sub-agents.
- **Cost & Token Optimized:** Its "budget-aware decision making" prevents sending the entire AST to the context window—a known challenge in `gmind context` (noted by the PO in PRD-03).

## 4. Integration Plan to Gmind PRDs

To integrate FastCode into the Gmind architecture, we propose the following updates to the PRDs:

### **Update to PRD-01-Overview.md**

- **Architecture Diagram Update:** Replace `Tree-sitter AST` in the Storage Layer with `FastCode Server (AST + Hybrid Index + Graph RAG)`.
- **System Layer Adjustment:** Instead of purely splitting into `Zvec` and `AST`, the Storage Layer will leverage FastCode as the dedicated Codebase RAG engine, while Zvec handles purely conversational history and generic docs.

### **Update to PRD-02-Storage.md**

- **Refining "Bộ nhớ Cấu trúc / Structural Memory":** Update this section to specify **FastCode** as the primary engine instead of a custom Tree-sitter script.
- FastCode handles the AST tree creation, Call/Dependency generation, and provides the API to query graph relationships.
- Zvec DB will still be used to store Chat Logs and generic Markdown, while FastCode's internal vector/FAISS engine + BM25 will manage the code components.

### **Update to PRD-03-CLI-and-Workflow.md**

- **`gmind search <query>`:** Under the hood, this command will proxy API requests to the `FastCode REST API` (port 8000) for code insights, retrieving precision AST snippets instead of raw Zvec queries.
- **`gmind context <beads-id>`:** Will merge Dolt Issue Status + Chat logs from Zvec + **FastCode's context summaries**. The PO's concern regarding Token overloading is directly resolved by FastCode's `budget-aware` contextualization engine.
- **MCP Expansion:** Add `FastCode MCP` as a core capability within the Execution Layer alongside `mcp_agent_mail`. Agents can run `<call_mcp name="fastcode">` to query specific dependencies before executing file writes.

## 5. Next Steps

1. **Review & Approve:** Review this integration research document.
2. **Execute PRD Updates:** If approved, I will implement these structural changes directly into `PRD-01-Overview.md`, `PRD-02-Storage.md`, and `PRD-03-CLI-and-Workflow.md`.
3. **Deploy FastCode Instance:** Start a prototype FastCode server in the `Gmind` repo to test API compatibility with the `gmind` CLI framework.
