# Refs System Improvements - Exploration Overview

**Exploration**: improve-refs
**Branch**: explore/improve-refs
**Date**: October 2025
**Status**: Research complete

## Context

The current refs system allows linking external git repositories and files as references, cached locally and symlinked into `.sow/refs/` for AI agent access. This exploration investigated improvements to make refs more discoverable, manageable, and searchable.

## What We Researched

### 1. OCI-Based Ref Distribution
**Research focus**: Replace git-based refs with OCI registry packages

**Key findings**:
- `github.com/jmgilman/go/oci` module provides production-ready OCI client
- Security built-in (path traversal protection, size limits, permission sanitization)
- Seamless Docker credential integration
- Simple Push/Pull API for packaging and distribution

**Outcome**: Technically viable, simpler than git-based distribution

### 2. Marketplace for Ref Discovery
**Research focus**: How users discover and install refs

**Key findings**:
- Homebrew-style "tap" system: git repos with manifest files
- Decentralized (anyone can publish marketplace)
- Version-agnostic (marketplace lists refs, versions queried from registry)
- Familiar UX (`sow refs marketplace add org/repo`)

**Outcome**: Git-based marketplace manifests + OCI registry storage

### 3. Semantic Search for Refs
**Research focus**: Enable natural language queries across installed refs

**Key findings**:
- Two-phase approach: Start with FTS5 keyword search (zero deps), add optional semantic
- Mature Go ecosystem: chromem-go (pure-Go vector DB), LangChainGo (chunking), Ollama SDK
- Implementation: ~2 weeks total (1 week per phase)
- Precomputed embeddings investigated but rejected (still requires Ollama for queries)

**Outcome**: Phased implementation with graceful fallback

### 4. MCP Server Integration
**Research focus**: How MCP servers complement static refs

**Key findings**:
- MCP for dynamic data (API docs, live services)
- Refs for static, versioned knowledge (team docs, templates)
- Plugin-declared MCP servers simpler than proxy architecture
- Clear separation of concerns

**Outcome**: Complementary systems, not competing

## High-Level Recommendations

### Immediate (Phase 1)
1. **OCI-based refs** - Replace git refs with OCI registry distribution
   - Simpler publishing workflow
   - Better versioning (OCI tags)
   - Metadata via `.sow-ref.yaml` manifest
   - Use existing `jmgilman/go/oci` module

2. **FTS5 keyword search** - Enable basic ref search immediately
   - Zero dependencies
   - Works offline
   - Fast indexing and queries
   - 2-3 days implementation

### Near-term (Phase 2)
3. **Marketplace system** - Git-based ref discovery
   - `.sow/marketplace.yml` manifests in git repos
   - `sow refs marketplace add org/repo`
   - `sow refs search <term>` across marketplaces
   - Simple validation and caching

4. **Semantic search** - Optional Ollama-based semantic search
   - Natural language queries
   - Automatic fallback to FTS5
   - User installs Ollama (one-time)
   - 5-7 days implementation

### Future Considerations
5. **MCP integration** - Ship MCP servers in sow plugin
   - Declared in plugin `.mcp.json`
   - Provides dynamic data access (Context7, GitHub, etc.)
   - Complements static refs

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Ref Distribution                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Publisher                    OCI Registry                  │
│  ┌──────────┐                ┌──────────┐                  │
│  │ Ref      │  sow refs      │ ghcr.io/ │                  │
│  │ Content  │─────publish────>│ org/ref  │                  │
│  │ + Metadata│                │ :v1.0.0  │                  │
│  └──────────┘                └──────────┘                  │
│                                     │                        │
│                                     │ pull                   │
│                                     ↓                        │
│  Consumer                    Local Cache                    │
│  ┌──────────┐                ┌──────────┐                  │
│  │ .sow/    │<───symlink─────│ ~/.cache/│                  │
│  │ refs/    │                │ sow/refs/│                  │
│  └──────────┘                └──────────┘                  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     Ref Discovery                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Git Repository (Marketplace)                               │
│  ┌──────────────────────────────┐                          │
│  │ .sow/marketplace.yml         │                          │
│  │                              │                          │
│  │ refs:                        │                          │
│  │   go-standards:              │                          │
│  │     registry: ghcr.io/...    │                          │
│  │     title: "Go Standards"    │                          │
│  │     tags: [golang, guide]    │                          │
│  └──────────────────────────────┘                          │
│           │                                                 │
│           │ sow refs marketplace add org/marketplace       │
│           ↓                                                 │
│  ┌──────────────────────────────┐                          │
│  │ ~/.config/sow/marketplaces/  │                          │
│  │ └─ org-marketplace/          │                          │
│  │    └─ .sow/marketplace.yml   │                          │
│  └──────────────────────────────┘                          │
│           │                                                 │
│           │ sow refs search golang                         │
│           ↓                                                 │
│  Search results with registry URLs                         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     Ref Search                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Phase 1: Keyword Search (FTS5)                             │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐           │
│  │ Query    │────>│ SQLite   │────>│ Results  │           │
│  │ "error"  │     │ FTS5     │     │ Snippets │           │
│  └──────────┘     └──────────┘     └──────────┘           │
│                                                              │
│  Phase 2: Semantic Search (Ollama + chromem-go)            │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐           │
│  │ Query    │────>│ Ollama   │────>│ chromem  │           │
│  │ "how to  │     │ Embed    │     │ Vector   │           │
│  │ handle   │     │          │     │ Search   │           │
│  │ errors?" │     └──────────┘     └──────────┘           │
│  └──────────┘            │                │                │
│                          │ Unavailable    │                │
│                          ↓                ↓                │
│                     ┌──────────┐     ┌──────────┐         │
│                     │ Fallback │────>│ FTS5     │         │
│                     │ to FTS5  │     │ Search   │         │
│                     └──────────┘     └──────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Key Decisions Made

### 1. OCI Over Git for Ref Distribution

**Why**:
- Simpler publishing (single `push` command vs git workflow)
- Better versioning (immutable tags, digest pinning)
- Registry authentication built-in (Docker credentials)
- Smaller storage overhead (optimized compression)
- Existing tooling (`jmgilman/go/oci`)

**Trade-offs accepted**:
- Requires OCI registry access (but GitHub provides free ghcr.io)
- Less familiar than git (but simpler API compensates)

### 2. Git-Based Marketplace (Not Centralized Registry)

**Why**:
- Decentralized (anyone can publish marketplace)
- Familiar model (Homebrew taps)
- No infrastructure required (just git hosting)
- Easy curation (team maintains own marketplace repo)

**Trade-offs accepted**:
- Name collisions possible (mitigated with `marketplace/name` syntax)
- Marketplace updates require git pull (acceptable, infrequent)

### 3. Two-Phase Search (FTS5 Then Semantic)

**Why**:
- Phase 1 provides immediate value (keyword search, zero deps)
- Validates search UX before investing in semantic
- Graceful fallback (FTS5 always available)
- User choice (can skip Ollama if keyword sufficient)

**Trade-offs accepted**:
- Two implementations to maintain
- Some complexity in fallback logic
- Ollama installation burden on users (but optional)

### 4. chromem-go Over sqlite-vec

**Why**:
- Pure Go (no CGO, simpler cross-compilation)
- Zero dependencies (easier distribution)
- Built-in Ollama integration
- Active development (2025)

**Trade-offs accepted**:
- In-memory (requires export/import for persistence)
- Exhaustive search (slower at >1M vectors, but sow won't reach that)

### 5. Precomputed Embeddings Rejected

**Why not**:
- User still needs Ollama to embed queries (no dependency savings)
- 10-30% larger OCI images permanently
- Model lock-in (can't change without republishing)

**Alternatives tried**: None better than local generation

## Technical Stack Selected

### OCI Refs
- **Distribution**: `github.com/jmgilman/go/oci`
- **Registry**: GitHub Container Registry (ghcr.io) recommended
- **Metadata**: `.sow-ref.yaml` manifest (YAML format)

### Marketplace
- **Storage**: Git repositories
- **Format**: `.sow/marketplace.yml` manifest
- **Discovery**: GitHub topics, docs, word-of-mouth

### Search
- **Phase 1**: SQLite FTS5 (built-in)
- **Phase 2**:
  - chromem-go (vector DB)
  - LangChainGo textsplitter (chunking)
  - Ollama (embeddings)

## Implementation Timeline

### Phase 1 (Weeks 1-3)
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

### Phase 2 (Weeks 4-5)
**Week 4**: Semantic indexing
- Ollama detection and setup
- chromem-go integration
- LangChainGo chunking
- Embedding generation

**Week 5**: Semantic search
- chromem-go query integration
- Fallback logic (Ollama → FTS5)
- Performance optimization
- User documentation

**Total**: 5 weeks for complete implementation

## Open Questions

1. **Migration strategy**: How to transition existing git-based refs to OCI?
   - Provide migration tool (`sow refs migrate git-to-oci`)
   - Support both formats during transition period
   - Document migration guide

2. **OCI registry hosting**: Recommend specific registries?
   - GitHub Container Registry (free, 500MB per package)
   - Docker Hub (rate limits on free tier)
   - Self-hosted (Harbor, Nexus) for enterprises

3. **Marketplace curation**: Who maintains official marketplace?
   - sow project maintains `sow-project/refs`
   - Community can create alternatives
   - Quality guidelines for inclusion

4. **Search scope**: What to index?
   - All markdown content (documentation)
   - Code comments (if code-classified refs)
   - Metadata (title, description, tags)

5. **Embedding model versioning**: Handle model updates?
   - Track model version in index metadata
   - Warn on version mismatch
   - Provide reindex command

## Success Criteria

### OCI Refs
- ✅ Publish ref to OCI registry in <5 minutes
- ✅ Install ref from registry in <30 seconds (10MB ref)
- ✅ Metadata extraction works without cloning entire repo
- ✅ Works with GitHub Container Registry

### Marketplace
- ✅ Add marketplace in <1 minute
- ✅ Search across marketplaces in <1 second
- ✅ Resolve ref to OCI URL successfully
- ✅ Handle name collisions gracefully

### Search
- ✅ Keyword search works offline with zero setup
- ✅ Search latency <50ms (Phase 1)
- ✅ Semantic search finds relevant results for natural language queries (Phase 2)
- ✅ Graceful fallback to FTS5 when Ollama unavailable
- ✅ Clear user guidance for Ollama installation

## Next Steps

1. **Review findings**: Validate approach with stakeholders
2. **Prioritize phases**: Confirm Phase 1 scope
3. **Design implementation**: Create detailed specs for each component
4. **Build prototypes**: Test OCI workflow and search UX
5. **Implement incrementally**: Ship Phase 1, gather feedback, iterate

## Related Documents

- **[OCI Refs Distribution](./oci-refs-distribution.md)** - Detailed OCI implementation
- **[Marketplace Design](./marketplace-design.md)** - Ref discovery system
- **[Semantic Search Strategy](./semantic-search-strategy.md)** - Two-phase search approach
- **[MCP Strategy](./mcp-strategy.md)** - Complete MCP server strategy (Context7 + Exa)
