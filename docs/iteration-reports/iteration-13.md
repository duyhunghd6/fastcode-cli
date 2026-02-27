# Iteration 13: E2E Full-Flow Query Parity & Architectural Divergence Analysis

## Overview

This iteration focused on analyzing the full query execution pipeline comparing FastCode Go vs FastCode Python to understand why Go occasionally executed more LLM calls (4-5) compared to Python (strictly 3). We ran a full-flow prompt capture script (`e2e-fullflow-query-compare.sh`) across several large projects (`beads`, `coding_agent_session_search`, `mcp_agent_mail`, `zvec`) and analyzed the exact prompt payloads.

## Key Changes

1. **Answer Generation Parity**: We aligned Go's final answer generation LLM call to perfectly match Python. Go now bundles the system prompt directly into the user message, and uses `temperature=0.4` and `max_tokens=20000` (up from 0.2 and 8000).
2. **E2E Test Robustness**: Fixed TTY hang issues in the comparison script where background Python executions would stall waiting for standard input.

## Architectural Discovery: AST Elements vs Regex Files

We experimentally bypassed Go's `llmSelectFiles` call in an attempt to perfectly replicate Python's "3-call pipeline" (Initial -> Round 2 -> Answer). However, Go still invoked 4-5 calls even with the bypass on large repositories.

**The Finding:**

- Python achieves a 3-call pipeline because its retrieval engine parses the codebase into an Abstract Syntax Tree (AST). It performs BM25 keyword searches against highly specific _Elements_ (classes/functions) and utilizes "Graph Expansion" to automatically pull in mathematical dependencies (callers/callees). This perfectly filters context before the LLM ever sees it.
- Go's initial retrieval relies heavily on text-based (regex) searches returning massive, noisy _whole files_ (e.g., 74 files for a single search).
- _Conclusion:_ Go's `llmSelectFiles` call acts as a vital "Semantic Bridge". It leverages the LLM's intuition to filter noisy regex hits down to the 5-10 structurally vital files. Removing this call forced the iterative agent into extra loops due to poor confidence. We reverted the bypass, cementing `llmSelectFiles` as a necessary architectural optimization for Go's current stack.
