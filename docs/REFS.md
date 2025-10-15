# External References System

**Last Updated**: 2025-10-15
**Purpose**: External knowledge and code reference system

The refs system provides a unified approach to external git resources, enabling teams to centralize and share knowledge across repositories without duplication or staleness.

---

## Table of Contents

- [Overview](#overview)
- [Purpose](#purpose)
- [Architecture](#architecture)
- [Reference Types](#reference-types)
- [CLI Commands](#cli-commands)
- [Workflows](#workflows)
- [Platform Differences](#platform-differences)
- [Orchestrator Integration](#orchestrator-integration)
- [Related Documentation](#related-documentation)

---

## Overview

Unified system for external git resources with semantic types (knowledge vs code). Replaces previous "sinks" (knowledge) and "repos" (code) systems with single consistent mechanism. Remote repositories cloned once to local cache and symlinked (or copied on Windows) to `.sow/refs/`. Two-index system separates categorical metadata (committed) from transient data (cache).

---

## Purpose

### Problem

Without refs: teams copy style guides, conventions, policies into every repo; documentation becomes stale as originals update; no central source of truth; AI agents lack context about team standards.

### Solution

With refs: centralize knowledge in dedicated repositories, install references into any repo needing them, automatic staleness detection, AI agents consult refs when making decisions.

### Use Cases

**Knowledge References**: Style guides (Python conventions, Go idioms), API standards (REST conventions, GraphQL patterns), team policies (code review checklist, testing requirements), security guidelines (authentication patterns, encryption standards).

**Code References**: Implementation examples from other services, shared libraries for pattern reference, reference architectures, working code to study and adapt (not direct imports).

---

## Architecture

### Core Concept

All remote repositories cached once per machine at `~/.cache/sow/repos/`. Symlinks (or copies on Windows) created from cache to `.sow/refs/` pointing to relevant paths. Metadata tracked in two separate indexes: committed index (categorical info shared with team) and cache index (transient data per-machine).

### Local Caching

Benefits: clone once and reference from multiple repos, efficient disk usage, single update point for all consuming repos, enables subpath references without sparse checkouts.

Cache organization: organized by git host and namespace, full repository clones (not sparse), one clone per repository-branch combination, shared across all local repositories.

### Symlink Strategy

Symlinks connect cache to repository refs. Symlinks not committed to git (only index files committed). Each developer runs `sow refs init` after cloning. Updates to cache automatically reflect via symlinks on Unix platforms.

Windows exception: Copy instead of symlink when symlink privileges unavailable (requires rsync on updates).

### Two-Index System

**Committed Index** (`.sow/refs/index.json`): What refs exist, where they come from (URL, branch, path), what they contain (tags, description, summary). Shared across team via git.

**Cache Index** (`~/.cache/sow/index.json`): When repos were cached, current commit SHAs, staleness status, which local repos use which refs. Per-machine never shared.

**Local Index** (`.sow/refs/index.local.json`): Local-only references (not shared with team). Uses `file://` protocol for local paths. Never committed to git.

**Separation Rationale**: Avoid git conflicts from timestamps/SHAs, enable team sharing without coupling to specific commits, each developer can be at different cache states, clean separation of concerns.

---

## Reference Types

### Knowledge (`--type knowledge`)

Documentation, guides, policies, standards. Typically markdown files. Read for principles, rules, conventions. Consulted when making decisions. Agents check these when implementing features to ensure compliance with team standards.

### Code (`--type code`)

Implementation examples, patterns, reference code. Working source code. Read for patterns and approaches. Studied and adapted (not directly imported). Agents examine these to understand implementation patterns when building similar functionality.

---

## CLI Commands

### Core Operations

**`sow refs add`**: Add remote or local reference. Clones to cache if needed, creates entry in committed or local index, creates symlink or copy to `.sow/refs/`.

**`sow refs init`**: Initialize references after cloning repository. Reads committed and local indexes, clones repos to cache if needed, creates symlinks or copies.

**`sow refs status`**: Check if references up to date with remote. Fetches from remote, compares SHAs, reports staleness.

**`sow refs update`**: Pull latest changes from remote. Updates cache with git pull, updates cache index, rsyncs to consuming repos on Windows.

**`sow refs list`**: Display configured references. Shows remote refs (shared with team) and local refs (per-developer).

**`sow refs remove`**: Remove reference from repository. Removes index entry, removes symlink or copy, updates cache index.

### Cache Management

**`sow cache status`**: Show cache usage and statistics. Lists cached repositories, sizes, usage by consuming repos, identifies orphaned caches.

**`sow cache prune`**: Remove cached repos not used by any repository. Requires confirmation, shows what will be removed.

**`sow cache clear`**: Remove all cached repositories. Warning that symlinks will break until `sow refs init` run.

**See Also**: [CLI_REFERENCE.md](./CLI_REFERENCE.md) for complete command documentation.

---

## Workflows

### Adding Remote Reference

User requests via natural language or slash command → Orchestrator examines content and analyzes to determine tags/description → Orchestrator calls `sow refs add` with appropriate parameters → CLI clones to cache if needed, adds entry to index, creates symlink → Orchestrator confirms installation.

### Fresh Clone Setup

Developer clones repository with existing refs → Orchestrator detects uninitialized refs when user starts work → Orchestrator prompts to initialize → User confirms → Orchestrator calls `sow refs init` → CLI clones repos to cache and creates symlinks.

### Checking for Updates

User requests update check → Orchestrator calls `sow refs status` → CLI fetches from remotes and compares SHAs → Orchestrator reports staleness and suggests updates for behind refs.

### Updating References

User requests update → Orchestrator calls `sow refs update` → CLI pulls in cache and updates index → On Windows: rsyncs to all consuming repos → Orchestrator reports changes pulled.

### Adding Multiple Paths from Same Repo

First path added creates new ref entry → Second path from same repo appends to existing ref's paths array → Both symlinks point to different subdirectories of same cached repo → Single cache serves multiple ref links.

---

## Platform Differences

### Unix/Linux/macOS

Native symlink support. Updates to cache automatically visible via symlinks (no additional sync needed). `sow refs update` only needs git pull in cache.

### Windows

Symlink requires Developer Mode or Administrator privileges. Fallback strategy: copy instead of symlink when privileges unavailable. Updates must rsync from cache to all consuming repos. CLI finds consuming repos in cache index and syncs updated paths. Detection automatic (CLI detects platform and chooses strategy).

---

## Orchestrator Integration

### Discovery Phase

Researcher agent consults knowledge refs when performing focused research. Refs provide team context grounding discussions in established standards.

### Design Phase

Architect agent references knowledge refs for design standards and code refs for implementation patterns. Design decisions align with team conventions documented in refs.

### Implementation Phase

Implementer agent consults knowledge refs for coding standards and code refs for implementation examples. Ensures code follows team conventions and patterns.

### Context Compilation

Orchestrator determines relevant refs for each task. Includes ref paths in task state references list. Workers read referenced files during execution. Focused context prevents bloat while ensuring standards compliance.

**See Also**: [AGENTS.md](./AGENTS.md#context-compilation)

---

## Related Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - External references system rationale and design
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Refs directory structure and git versioning
- **[PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md)** - Researcher agent usage of refs
- **[PHASES/DESIGN.md](./PHASES/DESIGN.md)** - Architect agent usage of refs
- **[PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)** - Implementer agent usage of refs
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - Complete refs and cache command reference
