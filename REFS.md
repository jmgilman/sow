# References System (Refs)

**Last Updated**: 2025-10-20
**Status**: Design Document

---

## Overview

The references (refs) system allows sow projects to integrate external GitHub repositories as knowledge sources or code examples. Refs are locally cached, indexed with rich metadata, and optionally semantically searchable.

**Core Capabilities:**
- GitHub repository caching and management
- LLM-driven metadata enrichment
- Keyword-based search
- Optional semantic code search (via Claude Context)
- Orchestrator-guided workflows

---

## Design Philosophy

### Progressive Enhancement

The refs system is built in layers:

1. **Foundation** - Simple GitHub repo caching (always available)
2. **Enrichment** - LLM-analyzed metadata and keyword search (default)
3. **Advanced** - Semantic code search with natural language queries (opt-in)

Each layer adds capability without breaking previous layers.

### GitHub-First

Focus exclusively on GitHub repositories:
- Most common case (public and private repos)
- Consistent authentication (SSH keys, tokens)
- Reliable infrastructure
- Git operations well-understood

**Removed**: Local file references (`file://`) - adds complexity for marginal benefit.

### Orchestrator-Driven

Rather than forcing users to memorize CLI flags and options, the orchestrator guides them through:
- Repository analysis and metadata generation
- Indexing decisions
- Service installation (Claude Context)
- Searching and discovery

---

## Implementation Phases

### Phase 1: Simplification and Foundation

**Goal**: Clean, simple GitHub-only refs system

**Changes:**
- Remove local file reference support
- Simplify to single ref type: `<org>/<repo>`
- Clean up CLI commands (remove unused options)
- Standardize on GitHub authentication

**Deliverables:**
- [ ] Remove `file://` protocol support
- [ ] Simplify `sow refs add` to GitHub-only
- [ ] Update index schema (remove local-specific fields)
- [ ] Clean up documentation

**CLI After Phase 1:**
```bash
# Add GitHub ref (simple)
sow refs add golang/go

# Add with subpath
sow refs add golang/go --path src/net/http

# List refs
sow refs list

# Remove ref
sow refs remove golang-go

# Info
sow refs info golang-go
```

**Index Schema v1.0:**
```json
{
  "version": "1.0.0",
  "refs": [
    {
      "id": "golang-go",
      "source": "https://github.com/golang/go",
      "branch": "master",
      "path": "src/net/http",
      "link": "golang/go"
    }
  ]
}
```

---

### Phase 2: Metadata Enrichment and Search

**Goal**: Rich, LLM-generated metadata and keyword search

**Changes:**
- Expand index schema with metadata fields
- Add orchestrator command: `/sow:refs:add`
- LLM analyzes repo and generates metadata
- Keyword-based search implementation

**Deliverables:**
- [ ] Update index schema to v2.0
- [ ] Create `/sow:refs:add` orchestrator command
- [ ] Implement LLM-driven analysis workflow
- [ ] Add `sow refs search` with keyword matching
- [ ] Create `/sow:refs:search` orchestrator command

**Orchestrator Workflow: `/sow:refs:add`**

```markdown
1. User provides: golang/go
2. Orchestrator clones to temp directory
3. Analyzes:
   - README.md (description, features)
   - Directory structure (languages, patterns)
   - Sample files (architectural style)
4. Generates metadata:
   - Description (1 sentence)
   - Summary (2-3 sentences)
   - Tags (5-10 keywords)
   - Topics (key areas)
   - Recommended for (use cases)
5. Presents for approval
6. User approves/edits
7. Executes: sow refs add golang/go --description "..." --tags ... --metadata '{...}'
```

**Index Schema v2.0:**
```json
{
  "version": "2.0.0",
  "refs": [
    {
      "id": "golang-go",
      "source": "https://github.com/golang/go",
      "branch": "master",
      "path": "src/net/http",
      "link": "golang/go",

      "description": "Go standard library HTTP implementation",
      "summary": "Reference implementation of HTTP client and server...",
      "tags": ["go", "http", "networking", "stdlib", "client", "server"],

      "metadata": {
        "languages": ["Go", "Assembly"],
        "topics": [
          "HTTP client patterns",
          "Server middleware",
          "Connection pooling"
        ],
        "key_files": [
          "client.go",
          "server.go",
          "transport.go"
        ],
        "recommended_for": [
          "Building HTTP services",
          "Understanding connection management",
          "Studying production-grade patterns"
        ],
        "architectural_patterns": [
          "Interface-based design",
          "Middleware pattern"
        ]
      },

      "indexed_at": "2024-10-20T16:30:00Z"
    }
  ]
}
```

**Search Implementation:**

Keyword-based ranking algorithm:
- Tag exact match: 10 points
- Tag fuzzy match: 5 points
- Description match: 8 points
- Topic match: 7 points
- Recommended-for match: 9 points (high value)

```bash
# CLI search (returns JSON)
sow refs search "jwt authentication" --format json

# Orchestrator search (conversational)
/sow:refs:search
> What are you looking for?
>> JWT authentication with RS256

Found 2 relevant refs:
  1. auth-examples/jwt (High: tags match, recommended_for match)
  2. golang/go (Medium: topic match)
```

---

### Phase 3: Semantic Code Search (Optional)

**Goal**: Natural language code search with Claude Context

**Changes:**
- Add `sow services` command for managing optional services
- Claude Context installation and configuration
- Optional semantic indexing workflow
- MCP tool integration for code search

**Deliverables:**
- [ ] Create `sow services` command structure
- [ ] Implement `sow services install claude-context`
- [ ] Update `/sow:refs:add` with optional indexing
- [ ] Add indexing metadata to schema
- [ ] Create agent instructions for using MCP search tool
- [ ] Update `/sow:refs:search` to use semantic search

**New Command: `sow services`**

```bash
# List available services
sow services list

# Show service info
sow services info claude-context

# Install service
sow services install claude-context

# Check service status
sow services status claude-context

# Uninstall service
sow services uninstall claude-context
```

**Service: Claude Context**

```bash
$ sow services install claude-context

Installing Claude Context (semantic code search)...

Prerequisites:
  ✓ Node.js installed
  ✓ npm available

Installing:
  1. npm install -g claude-context-local
  2. Configuring MCP server in .claude/mcp.json
  3. Testing installation

✓ Installation complete!

Claude Context is now available as an MCP tool.
Use /sow:refs:add to index repositories with semantic search.

First-time indexing will download embeddings model (~1.3GB).
```

**Implementation:**

The `sow services install claude-context` command:
1. Checks Node.js availability
2. Installs `claude-context-local` globally via npm
3. Creates/updates `.claude/mcp.json` in repository:

```json
{
  "mcpServers": {
    "claude-context": {
      "command": "npx",
      "args": ["-y", "claude-context-local"],
      "description": "Local semantic code search"
    }
  }
}
```

4. Verifies MCP server responds
5. Returns success/failure

**Enhanced Workflow: `/sow:refs:add` (Phase 3)**

```
/sow:refs:add

Orchestrator: What GitHub reference would you like to add?

User: golang/go

[Performs standard indexing from Phase 2]

✓ Reference added: golang/go

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

This is a code reference. I can configure advanced semantic indexing
that makes searching code-aware with natural language queries.

Examples:
  • "Find connection pooling with graceful shutdown"
  • "Show me JWT RS256 signing implementations"
  • "Locate middleware error handling patterns"

Would you like me to configure semantic indexing?

User: yes

Orchestrator: Checking for Claude Context...

[Calls MCP tool - fails if not installed]

I see that Claude Context isn't available locally.
Claude Context provides local, private semantic code search.

Requirements:
  • ~1.3GB one-time model download
  • Node.js (already installed ✓)

Would you like me to install it?

User: yes

Orchestrator: Installing Claude Context...

[Executes CLI]
sow services install claude-context

✓ Installation complete!

Now indexing golang/go with semantic search...

[Calls MCP tool]
MCP: index_codebase
  path: /Users/josh/code/myproject/.sow/refs/golang/go
  recursive: true

Progress: 15% (142/940 files)
Progress: 42% (395/940 files)
Progress: 78% (733/940 files)
Progress: 100% (940/940 files)

✓ Semantic indexing complete!
  Files: 940
  Code chunks: 4,238
  Storage: ~/.claude_code_search/golang-go/

[Updates metadata via CLI]
sow refs update-metadata golang-go \
  --semantic-indexed true \
  --semantic-indexed-at "2024-10-20T17:00:00Z" \
  --semantic-stats '{"files":940,"chunks":4238}'

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Reference ready with semantic search!

You can now ask questions like:
  "Find graceful shutdown patterns in golang/go"

Try: /sow:refs:search
```

**Index Schema v3.0 (with semantic indexing):**

```json
{
  "version": "3.0.0",
  "refs": [
    {
      "id": "golang-go",
      "source": "https://github.com/golang/go",

      // Phase 2 metadata
      "description": "...",
      "tags": [...],
      "metadata": {...},

      // Phase 3 semantic indexing
      "semantic_indexed": true,
      "semantic_indexed_at": "2024-10-20T17:00:00Z",
      "semantic_stats": {
        "files": 940,
        "chunks": 4238,
        "languages": ["Go"]
      }
    }
  ]
}
```

**Agent Instructions (for workers):**

When workers need to research refs with semantic indexing:

```markdown
## Using Semantic Search on References

If a task references an indexed ref (check state.yaml), you can search it semantically.

**Example:**

Task references: refs/golang/go
Task description: "Implement graceful HTTP server shutdown"

You can search the ref:

Use MCP tool: search_code
Parameters:
  query: "graceful shutdown http server"
  paths: ["/absolute/path/to/.sow/refs/golang/go"]
  top_k: 5

Returns file paths, line numbers, and code snippets matching your query.

This is more effective than browsing directories manually.
```

---

## Final System Architecture

### Component Responsibilities

**CLI (`sow` binary):**
- GitHub repository operations (clone, update)
- Metadata management (CRUD on index.json)
- Service installation and configuration
- Context detection and session info

**Orchestrator Agent:**
- LLM-driven metadata generation
- User guidance and approval flows
- MCP tool orchestration (indexing, searching)
- Conversational search interface

**MCP Server (Claude Context):**
- AST-based code parsing
- Embedding generation (local model)
- Vector storage (FAISS + SQLite)
- Semantic search queries

**Workers (Implementer, etc.):**
- Research using semantic search
- Study referenced code examples
- Apply patterns from refs

### Data Flow

```
User Request
    ↓
Orchestrator (Conversational Interface)
    ↓
├─→ CLI (Metadata Operations)
│   └─→ .sow/refs/index.json
│
└─→ MCP Tool (Semantic Operations)
    └─→ ~/.claude_code_search/ (FAISS indexes)
```

### Directory Structure

```
repository/
├── .claude/
│   └── mcp.json                    # MCP server config (services add here)
│
└── .sow/
    └── refs/
        ├── index.json              # Ref metadata (v3.0 schema)
        ├── golang/
        │   └── go/                 # Cached GitHub repo
        └── auth-examples/
            └── jwt/                # Cached GitHub repo

~/.claude_code_search/              # Global semantic indexes
├── golang-go/
│   ├── faiss.index
│   └── metadata.db
└── auth-examples-jwt/
    ├── faiss.index
    └── metadata.db
```

---

## CLI Reference

### Phase 1 Commands

```bash
sow refs add <org>/<repo> [--path <subpath>] [--branch <branch>]
sow refs list
sow refs remove <id>
sow refs info <id>
sow refs update <id>              # Git pull latest
```

### Phase 2 Commands

```bash
sow refs search <query> [--format json|table]
sow refs update-metadata <id> --field value
```

### Phase 3 Commands

```bash
# Service management
sow services list
sow services info <service>
sow services install <service>
sow services status <service>
sow services uninstall <service>

# Metadata for semantic indexing
sow refs update-metadata <id> \
  --semantic-indexed true \
  --semantic-stats '{...}'
```

---

## Orchestrator Commands

### `/sow:refs:add`

Conversational workflow for adding refs with:
- GitHub repo analysis
- LLM-generated metadata
- Optional semantic indexing
- Service installation if needed

### `/sow:refs:search`

Search interface supporting:
- Keyword search (Phase 2)
- Semantic search (Phase 3, if indexed)
- Combined ranking
- Conversational results

### `/sow:refs:reindex`

Re-index a ref after updates:
- Clear old semantic index
- Re-analyze with MCP tool
- Update metadata

---

## Migration Guide

### v1.0 → v2.0 (Phase 1 → Phase 2)

**Changes:**
- Added metadata fields
- Removed local file references

**Migration:**
```bash
# Automatic migration on first run
sow refs migrate

# Manually:
# 1. Remove any file:// refs (no longer supported)
# 2. Run /sow:refs:add for existing refs to generate metadata
```

### v2.0 → v3.0 (Phase 2 → Phase 3)

**Changes:**
- Added semantic indexing fields
- Added services system

**Migration:**
```bash
# No breaking changes - v3.0 is additive
# Existing refs work without semantic indexing
# Optionally install Claude Context and reindex
```

---

## Best Practices

### When to Add Refs

**Good candidates:**
- Style guides and conventions (knowledge refs)
- Reference implementations you want to study
- Libraries demonstrating patterns you'll use
- Production-grade code examples

**Not ideal:**
- Dependencies (use package managers)
- Your own code (already in repo)
- Temporary/experimental projects

### Indexing Decisions

**Use keyword search when:**
- Ref is documentation/knowledge
- Ref is small (<100 files)
- You know exact terminology

**Use semantic search when:**
- Large codebase (1000+ files)
- Complex patterns to study
- Natural language queries needed
- Research-intensive work

### Ref Organization

**Naming convention:**
- Use GitHub format: `org/repo` → `org-repo` ID
- Subpaths: `org/repo/path` → `org-repo-path` ID

**Tagging strategy:**
- Language tags: `go`, `python`, `rust`
- Domain tags: `http`, `auth`, `database`
- Pattern tags: `middleware`, `repository`, `factory`
- Use 5-10 tags per ref

---

## Performance Considerations

### Caching

- GitHub repos cached in `.sow/refs/`
- Semantic indexes in `~/.claude_code_search/`
- Updates are incremental (only changed files)

### Storage

- GitHub repo: ~10-100 MB typical
- Semantic index: ~5-20 MB per repo
- Model download: 1.3 GB one-time

### Speed

- Keyword search: <100ms
- Semantic search: <500ms
- First-time indexing: 1-5 min per 1000 files
- Re-indexing: Incremental, only changed files

---

## Troubleshooting

### MCP Tool Not Found

```
Error: MCP tool 'claude-context' not available
```

**Solution:**
```bash
sow services install claude-context
```

### Indexing Failed

```
Error: Failed to index repository
```

**Check:**
1. Node.js installed: `node --version`
2. Service status: `sow services status claude-context`
3. Disk space for model download
4. Repository size (<10GB recommended)

### Search Returns No Results

**Keyword search:**
- Try different terms
- Check tags with `sow refs list`
- Use `/sow:refs:search` for suggestions

**Semantic search:**
- Verify indexed: `sow refs info <id>`
- Check `semantic_indexed: true`
- Try re-indexing: `/sow:refs:reindex <id>`

---

## Future Enhancements

### Phase 4: Multi-Repository Search

Search across all indexed refs simultaneously:
```bash
sow refs search-all "JWT authentication"
# Returns results from all indexed refs, ranked globally
```

### Phase 5: Ref Collections

Group related refs:
```bash
sow refs collections create go-stdlib
sow refs collections add go-stdlib golang/go
sow refs collections search go-stdlib "http client"
```

### Phase 6: Ref Suggestions

Automatic ref discovery:
```
Task: "Implement OAuth2 authentication"

Orchestrator: I notice this task involves OAuth2. Let me search for relevant refs...
Found: No OAuth2 refs in your workspace.

Popular suggestions:
  1. golang/oauth2 - Go OAuth2 client library
  2. auth0/go-jwt-middleware - JWT middleware patterns

Would you like me to add these?
```

---

## Related Documentation

- **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - Overall sow architecture
- **[CLI_REFERENCE.md](./docs/CLI_REFERENCE.md)** - Complete CLI command reference
- **[USER_GUIDE.md](./docs/USER_GUIDE.md)** - Day-to-day workflows

---

## Contributing

When working on the refs system:

1. **Phase 1**: Focus on simplicity and GitHub-only support
2. **Phase 2**: LLM analysis quality is critical for good metadata
3. **Phase 3**: Keep semantic indexing optional and well-tested

**Testing checklist:**
- [ ] Add GitHub repo (public and private)
- [ ] Metadata generation quality
- [ ] Keyword search accuracy
- [ ] Service installation (clean and idempotent)
- [ ] Semantic indexing (first-time and updates)
- [ ] Search result ranking

---

**Status**: Phase 1 planned, implementation pending
