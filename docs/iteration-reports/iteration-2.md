# Iteration 2 — Indexing (Embeddings + BM25)

**Date:** 2026-02-26  
**Duration:** ~45 minutes  
**Status:** ✅ Complete

---

## Objective

Implement the indexing layer: multi-level code element indexing, BM25 keyword search, vector similarity search, hybrid retrieval, LLM API client, and disk caching.

---

## Deliverables

| #   | File                             | Package | Lines | Description                                                   |
| --- | -------------------------------- | ------- | ----- | ------------------------------------------------------------- |
| 1   | `internal/llm/client.go`         | llm     | 185   | OpenAI-compatible HTTP client (chat + embeddings)             |
| 2   | `internal/llm/embedder.go`       | llm     | 90    | Batch embedding generator with configurable batch sizes       |
| 3   | `internal/index/indexer.go`      | index   | 200   | Multi-level indexer (file / class / function / documentation) |
| 4   | `internal/index/bm25.go`         | index   | 180   | Pure-Go BM25+ keyword search with snake_case tokenizer        |
| 5   | `internal/index/vector_store.go` | index   | 105   | In-memory cosine similarity vector search                     |
| 6   | `internal/index/hybrid.go`       | index   | 130   | Score fusion: BM25 × 0.4 + vector × 0.6                       |
| 7   | `internal/cache/cache.go`        | cache   | 82    | Gob-encoded disk serialization for index persistence          |

**Total Phase 2:** ~972 lines of Go

---

## Test Results

```
=== cache ===
TestCacheSaveAndLoad          --- PASS
TestCacheLoadNotExists        --- PASS

=== graph (Phase 1) ===
TestGraphAddEdge              --- PASS
TestGraphNoDuplicateEdges     --- PASS
TestGraphNoSelfLoop           --- PASS
TestGetRelatedElements        --- PASS
TestGenerateElementID         --- PASS
TestBuildInheritanceGraph     --- PASS

=== index ===
TestBM25AddAndSearch          --- PASS
TestBM25EmptyQuery            --- PASS
TestBM25NoMatch               --- PASS
TestBM25DocCount              --- PASS
TestTokenize                  --- PASS
TestHybridRetrieverBM25Only   --- PASS
TestHybridRetrieverElementCount --- PASS
TestVectorStoreAddAndSearch   --- PASS
TestVectorStoreEmpty          --- PASS
TestCosineSimilarity          --- PASS
TestVectorStoreCount          --- PASS

=== llm ===
TestChatCompletion            --- PASS
TestEmbed                     --- PASS
TestChatCompletionAPIError    --- PASS
TestBuildSearchText           --- PASS

=== util (Phase 1) ===
TestGetLanguageFromPath       --- PASS
TestIsSupportedFile           --- PASS
TestCountLines                --- PASS
TestExtractLines              --- PASS
TestFilePathToModulePath      --- PASS
```

**Result: 28/28 PASS** ✅

### Coverage

| Package          | Coverage  | Target   |
| ---------------- | --------- | -------- |
| `internal/cache` | **84.6%** | ≥ 75% ✅ |
| `internal/graph` | **68.4%** | ≥ 75%    |
| `internal/index` | **58.6%** | ≥ 75%    |
| `internal/llm`   | **51.1%** | ≥ 75%    |
| `internal/util`  | **60.0%** | ≥ 80%    |

---

## Bugs Fixed During Implementation

### 1. BM25 IDF Zero-Score Bug

**Problem:** Standard BM25 IDF formula `log((N-df+0.5)/(df+0.5))` returns negative values when >50% of documents contain a term, which was being clamped to 0. This caused zero scores for common terms in small collections.  
**Fix:** Switched to BM25+ variant: `log(1 + (N-df+0.5)/(df+0.5))` — always positive.

### 2. Tokenizer Snake_Case Splitting

**Problem:** The BM25 tokenizer treated `_` as a regular character, producing `"build_graph"` as one token instead of `["build", "graph"]`.  
**Fix:** Treat `_` as a separator to properly split snake_case identifiers.

---

## Architecture Decisions

1. **Pure-Go BM25:** Implemented BM25+ from scratch instead of importing Bleve, keeping dependencies minimal. The entire BM25 engine is ~180 lines.

2. **In-memory vector store:** Chose simple `map[string][]float32` + cosine similarity over FAISS/Annoy. For codebases under 100k elements this is fast enough, and avoids CGO dependencies.

3. **Hybrid score fusion:** Using weighted linear combination (semantic 0.6 + keyword 0.4) with BM25 scores normalized to [0,1]. This is the same approach as the Python FastCode retriever.

4. **Gob caching:** Go's native `encoding/gob` for disk persistence instead of JSON. Faster serialization and supports all Go types natively.

5. **Mock HTTP server tests:** LLM client tests use `httptest.NewServer` to simulate OpenAI API responses, ensuring tests run without real API keys.
