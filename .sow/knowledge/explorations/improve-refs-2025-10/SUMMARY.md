# Refs System Improvements - Exploration Summary

**Exploration**: improve-refs
**Date**: October 2025
**Branch**: explore/improve-refs
**Status**: Research complete

## What We Explored

This exploration investigated major improvements to sow's refs system to make refs more discoverable, manageable, and searchable. We researched four key areas:

1. **OCI-based distribution** - Replace git refs with OCI registry packages
2. **Marketplace for discovery** - Homebrew-style ref discovery via git manifests
3. **Semantic search** - Enable natural language queries across refs
4. **MCP integration** - Complement static refs with dynamic public data

## Reading Guide

### Start Here: Overview

**[00-overview.md](./00-overview.md)** (15 min read)

High-level summary of the entire exploration. Read this first to understand:
- What we researched and why
- Key findings and recommendations
- Architecture diagrams
- Implementation timeline (~5 weeks)
- Decision rationale

**Best for**: Getting the big picture, understanding recommendations, sharing with stakeholders

---

### Deep Dives

#### OCI-Based Refs Distribution

**[oci-refs-distribution.md](./oci-refs-distribution.md)** (30 min read)

Complete design for replacing git-based refs with OCI registry packages.

**Topics covered**:
- Why OCI over git (simpler, better versioning)
- `github.com/jmgilman/go/oci` module capabilities
- `.sow-ref.yaml` metadata schema (complete specification)
- Publishing workflow (create → publish → version)
- Consumption workflow (install → update → remove)
- Registry recommendations (ghcr.io, Docker Hub, self-hosted)
- Migration from git refs
- Security considerations

**Best for**: Understanding OCI implementation details, schema design, publishing workflows

**Key takeaways**:
- OCI is simpler than git (one command to publish)
- Metadata via `.sow-ref.yaml` with classifications and LLM hints
- GitHub Container Registry recommended (free, 500MB per package)
- Existing Go module ready to use

---

#### Marketplace Design

**[marketplace-design.md](./marketplace-design.md)** (20 min read)

Homebrew-style marketplace system for discovering and installing refs.

**Topics covered**:
- Git-based marketplace manifests (`.sow/marketplace.yml`)
- Marketplace commands (`marketplace add`, search, install)
- Version resolution from OCI registry
- Publishing and curation guidelines
- Decentralized architecture (anyone can publish)

**Best for**: Understanding ref discovery, marketplace workflows, publishing guides

**Key takeaways**:
- Decentralized like Homebrew taps (git repos with manifests)
- `sow refs marketplace add org/repo` adds marketplace
- `sow refs search <term>` searches across all marketplaces
- No central infrastructure required

---

#### Semantic Search Strategy

**[semantic-search-strategy.md](./semantic-search-strategy.md)** (25 min read)

Two-phase search implementation: FTS5 keyword search (MVP), then optional Ollama semantic search.

**Topics covered**:
- Phase 1: SQLite FTS5 (zero dependencies, ~1 week implementation)
- Phase 2: Ollama + chromem-go + LangChainGo (~1 week implementation)
- Technology stack justification (pure-Go, battle-tested)
- Complete implementation roadmap with daily breakdown
- User automation strategy
- Graceful fallback (semantic → FTS5)
- Performance characteristics

**Best for**: Understanding search implementation, library choices, user experience

**Key takeaways**:
- FTS5 first provides immediate value (keyword search, <5 days)
- Semantic search optional (requires Ollama, but automatic fallback)
- chromem-go: pure-Go vector DB (zero deps, no CGO)
- LangChainGo: markdown-aware chunking (100+ projects use it)
- Total: ~2 weeks for complete search solution

---

#### MCP Strategy

**[mcp-strategy.md](./mcp-strategy.md)** (25 min read)

Complete strategy for integrating MCP servers to complement static refs with dynamic public data.

**Topics covered**:
- MCP vs refs (when to use which)
- Selected servers: Context7 (primary) + Exa (secondary)
- Plugin-declared deployment (`.mcp.json`)
- Usage guidance for LLMs (priority: Context7 → Exa → refs)
- Integration with refs system
- Complete implementation checklist

**Best for**: Understanding MCP integration, server selection rationale, deployment strategy

**Key takeaways**:
- Ship 2 MCP servers: Context7 (library docs) + Exa (code/web search)
- Both remote HTTP (no Node.js dependency)
- Both have free tiers (works out of box)
- Context7 PRIMARY, Exa SECONDARY (prevents credit exhaustion)
- Deploy via plugin `.mcp.json` (zero sow CLI work)

---

## Quick Reference

### By Implementation Timeline

**Week 1**: OCI refs
- OCI push/pull integration
- `.sow-ref.yaml` schema
- Update `sow refs add` for OCI URLs

**Week 2**: Marketplace
- Marketplace manifest schema
- `sow refs marketplace` commands
- Ref discovery and search

**Week 3**: FTS5 search
- SQLite indexing on ref add
- `sow refs search` command
- Snippet extraction and display

**Week 4**: Semantic indexing
- Ollama detection and setup
- chromem-go + LangChainGo integration
- Embedding generation

**Week 5**: Semantic search
- chromem-go query integration
- Fallback logic (Ollama → FTS5)
- Performance optimization

**MCP**: Can ship immediately with plugin (no CLI changes)

### By Priority

**Phase 1 (Immediate - Weeks 1-3)**:
1. OCI refs distribution
2. Marketplace discovery
3. FTS5 keyword search

**Phase 2 (Near-term - Weeks 4-5)**:
4. Semantic search with Ollama

**Shipped Separately (Immediate)**:
5. MCP servers via plugin

### By Technical Area

**Distribution**: [oci-refs-distribution.md](./oci-refs-distribution.md)
**Discovery**: [marketplace-design.md](./marketplace-design.md)
**Search**: [semantic-search-strategy.md](./semantic-search-strategy.md)
**Dynamic Data**: [mcp-strategy.md](./mcp-strategy.md)

---

## Key Decisions Made

### 1. OCI Over Git for Distribution
- **Why**: Simpler publishing, better versioning, existing Go module
- **Trade-off**: Less familiar than git, but API is simpler
- **Verdict**: Production-ready, clear win

### 2. Git-Based Marketplace (Not Centralized)
- **Why**: Decentralized, no infrastructure, familiar Homebrew model
- **Trade-off**: Name collisions possible (mitigated with namespace)
- **Verdict**: Simpler than centralized registry

### 3. Two-Phase Search (FTS5 Then Semantic)
- **Why**: FTS5 provides immediate value, validates UX before semantic investment
- **Trade-off**: Two implementations to maintain
- **Verdict**: Graceful approach, semantic is optional

### 4. chromem-go Over sqlite-vec
- **Why**: Pure Go (no CGO), zero dependencies, built-in Ollama integration
- **Trade-off**: In-memory (vs persistent), exhaustive search (vs HNSW)
- **Verdict**: Simpler deployment, sufficient for sow's scale

### 5. Context7 + Exa (Not GitHub/Fetch)
- **Why**: Both remote HTTP (no Node.js), complement each other perfectly
- **Trade-off**: GitHub/Fetch valuable but require npx
- **Verdict**: Simplicity wins, remote-only is cleaner

### 6. Plugin-Declared MCP (Not Proxy)
- **Why**: Simplest deployment, Claude Code handles everything
- **Trade-off**: Can't add routing/caching logic
- **Verdict**: 2 servers don't need proxy complexity

---

## Technology Stack

### OCI Refs
- **Distribution**: `github.com/jmgilman/go/oci`
- **Registry**: GitHub Container Registry (ghcr.io) recommended
- **Metadata**: `.sow-ref.yaml` manifest (YAML format)

### Marketplace
- **Storage**: Git repositories
- **Format**: `.sow/marketplace.yml` manifest
- **Discovery**: GitHub, docs, word-of-mouth

### Search
- **Phase 1**: SQLite FTS5 (built-in)
- **Phase 2**:
  - `github.com/philippgille/chromem-go` - Vector DB (pure Go)
  - `github.com/tmc/langchaingo/textsplitter` - Chunking (100+ users)
  - `github.com/ollama/ollama/api` - Embeddings (via chromem-go)

### MCP Servers
- **Context7** (Upstash) - Library docs (primary)
- **Exa** (Exa Labs) - Code/web search (secondary)

---

## Success Criteria

### OCI Refs
- ✅ Publish ref to OCI registry in <5 minutes
- ✅ Install ref from registry in <30 seconds (10MB ref)
- ✅ Metadata extraction works without cloning
- ✅ Works with GitHub Container Registry

### Marketplace
- ✅ Add marketplace in <1 minute
- ✅ Search across marketplaces in <1 second
- ✅ Resolve ref to OCI URL successfully
- ✅ Handle name collisions gracefully

### Search
- ✅ Keyword search works offline with zero setup (FTS5)
- ✅ Search latency <50ms (Phase 1)
- ✅ Semantic search finds relevant results (Phase 2)
- ✅ Graceful fallback to FTS5 when Ollama unavailable
- ✅ Clear user guidance for Ollama setup

### MCP
- ✅ Context7 and Exa auto-configure via plugin
- ✅ Both work immediately (free tiers)
- ✅ Optional API keys documented
- ✅ Context7 primary, Exa secondary enforced
- ✅ Graceful degradation when servers unavailable

---

## Open Questions

### OCI Refs
1. Registry quotas: What if user exceeds 500MB GitHub limit?
2. Offline publishing: Export/import tarball workflow?
3. Ref versioning semantics: Major/minor/patch guidelines?
4. Ref signing: Should refs be cryptographically signed?

### Marketplace
1. Official marketplace: Who maintains `sow-project/refs`?
2. Quality guidelines: Criteria for inclusion?
3. Conflict resolution: Multiple marketplaces with same ref name?

### Search
1. Incremental indexing: Delta detection vs full re-index?
2. Model versioning: Handle embedding model updates?
3. Background indexing: Block on add or spawn worker?
4. Result ranking: Balance similarity, recency, priority?

### MCP
1. API key management: Should sow provide `sow mcp configure`?
2. Usage monitoring: How to ensure Context7 preferred over Exa?
3. Credit exhaustion: Proactive warnings when Exa credits low?
4. Future additions: What triggers adding more servers?

---

## Next Steps

1. **Review findings**: Validate approach with stakeholders
2. **Prioritize phases**: Confirm Phase 1 scope (OCI + Marketplace + FTS5)
3. **Design specs**: Create detailed implementation specs
4. **Build prototypes**: Test OCI workflow, FTS5 search, MCP integration
5. **Implement incrementally**: Ship Phase 1, gather feedback, iterate

---

## Participants

**Conducted**: October 2025
**Participants**: Josh Gilman, Claude
**Branch**: explore/improve-refs

---

## Document Index

1. **[00-overview.md](./00-overview.md)** - High-level exploration summary (START HERE)
2. **[oci-refs-distribution.md](./oci-refs-distribution.md)** - OCI-based ref distribution design
3. **[marketplace-design.md](./marketplace-design.md)** - Ref discovery marketplace system
4. **[semantic-search-strategy.md](./semantic-search-strategy.md)** - Two-phase search implementation
5. **[mcp-strategy.md](./mcp-strategy.md)** - MCP server integration strategy

**Total reading time**: ~2 hours for complete understanding
