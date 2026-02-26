# FastCode Analysis: `steveyegge/beads` Repository

Based on the FastCode analysis of the `steveyegge/beads` repository (loaded, indexed 916 files, 1439 elements), here is the technical overview:

## 1. Main Purpose

The repository’s main purpose is to provide a developer-oriented issue tracker that is:

- **Local-first/offline-first**
- **Version-controlled at the data layer**
- **Syncable across machines/clones**
- **Scriptable via CLI**

Beads uses **Dolt** as a storage backend with Git-like semantics (branch/merge/diff/push/pull) at the DB level.

## 2. Architecture (High-Level)

1. **CLI/User Layer** (`bd` commands built with Cobra)
2. **Local Database Layer** (Dolt-backed persistent store under `.beads/...`)
3. **Remote Sync Layer** (Dolt remotes such as DoltHub/S3/GCS)

### Data flow patterns

- **Write path:** `bd create` → DB update → auto-commit to version history
- **Read path:** `bd list` → SQL query → immediate results
- **Sync path:** `bd dolt push/pull` ↔ remote commits

## 3. Core Components

- **CLI Command Layer (Cobra-based):** Registers commands (e.g., `restore <issue-id>`), handles CLI errors, provides JSON output, and integrates with a shared store.
- **Storage Abstraction & Domain Model:** Includes `types.Issue` with fields like `Description`, `Dependencies`, `CompactionLevel`. Provides point reads (`GetIssue`), historical queries, and commits.
- **Runtime Modes:** Dolt server mode (for concurrency/background operations) and Embedded mode (for CI/container use).
- **Git/Worktree Integration:** Each Git worktree is initialized with its own `.beads` directory and DB file (`.beads/beads.db`).

## 4. Error Recovery & Compaction

- Issues may be compacted for size management.
- The `restore` command searches historical versions and selects the fullest version by content size to output read-only restored data.

## 5. Architectural Characteristics

- **Local-first:** operations run against local DB.
- **Explicit sync:** no always-on cloud dependency.
- **History-rich:** commit-level audit/recovery at DB layer.
- **CLI-centric:** command-driven workflows for developers/automation.
- **Worktree-aware:** separate state per checkout/worktree.

---

_Source: Generated via FastCode remote server (`http://10.0.1.42:9990/query`) processing AST nodes and Hybrid Indexing._
