# External References System

**Last Updated**: 2025-10-16
**Purpose**: External knowledge and code reference system

The refs system provides a unified, extensible approach to external resources. Using URL schemes to identify reference types, it enables teams to centralize and share knowledge across repositories without duplication or staleness.

---

## Table of Contents

- [Overview](#overview)
- [Purpose](#purpose)
- [Architecture](#architecture)
- [Reference Types](#reference-types)
- [Type System](#type-system)
- [CLI Commands](#cli-commands)
- [Workflows](#workflows)
- [Platform Differences](#platform-differences)
- [Orchestrator Integration](#orchestrator-integration)
- [Related Documentation](#related-documentation)

---

## Overview

Extensible system for external resources using URL schemes to determine type. References are cached locally at `~/.cache/sow/refs/{type}/{id}/` and symlinked (or copied on Windows) to `.sow/refs/`. Two-index system separates categorical metadata (committed) from transient data (cache). Types can be enabled/disabled based on local system capabilities.

**Key Features**:
- URL scheme-based type inference (`git+https://`, `file://`, etc.)
- Local caching enforced for all types
- Extensible type system via interface
- Semantic classification (knowledge vs code) orthogonal to structural type
- Type-specific CLI commands and slash command documentation

---

## Purpose

### Problem

Without refs: teams copy style guides, conventions, policies into every repo; documentation becomes stale as originals update; no central source of truth; AI agents lack context about team standards.

### Solution

With refs: centralize resources in dedicated repositories or locations, install references into any repo needing them, automatic staleness detection, AI agents consult refs when making decisions.

### Use Cases

**Knowledge References**: Style guides (Python conventions, Go idioms), API standards (REST conventions, GraphQL patterns), team policies (code review checklist, testing requirements), security guidelines (authentication patterns, encryption standards).

**Code References**: Implementation examples from other services, shared libraries for pattern reference, reference architectures, working code to study and adapt (not direct imports).

---

## Architecture

### Core Concept

All external resources cached locally at `~/.cache/sow/refs/{type}/{id}/`. Symlinks (or copies on Windows) created from cache to `.sow/refs/{link}/` pointing to cached content. Metadata tracked in two separate indexes: committed index (categorical info shared with team) and cache index (transient data per-machine).

### Type System

Each reference type implements a common interface:
- **IsEnabled()** - Check if type available (tools installed)
- **Init()** - One-time initialization
- **Cache()** - Fetch and cache content locally
- **Update()** - Update cached content
- **IsStale()** - Check if cache outdated
- **Cleanup()** - Remove cached content
- **ValidateConfig()** - Validate type-specific configuration

Types registered in a central registry. Disabled types are skipped during operations (warnings shown).

### URL-Based Type Inference

Reference type determined from URL scheme:
- `git+https://github.com/org/repo` → git type (HTTPS transport)
- `git+ssh://git@github.com/org/repo` → git type (SSH transport)
- `git@github.com:org/repo` → git type (auto-converted to git+ssh://)
- `file:///path/to/local` → file type
- Future: `web+https://docs.example.com` → web type

### Local Caching

**Benefits**: Fetch once and reference from multiple repos, efficient disk usage, single update point for all consuming repos, enables subpath references without sparse checkouts, guaranteed filesystem access for AI.

**Cache Organization**:
```
~/.cache/sow/refs/
├── git/
│   ├── abc123/          # Cached git repo
│   └── def456/          # Another cached git repo
└── file/
    └── xyz789/          # Cached file ref (symlink or copy)
```

Each ref assigned unique ID (auto-generated). Cache path: `~/.cache/sow/refs/{type}/{id}/`.

### Symlink Strategy

Symlinks connect cache to repository refs. Symlinks not committed to git (only index files committed). Each developer runs `sow refs init` after cloning. Updates to cache automatically reflect via symlinks on Unix platforms.

**Windows Exception**: Copy instead of symlink when symlink privileges unavailable (requires rsync on updates).

### Two-Index System

**Committed Index** (`.sow/refs/index.json`):
- What refs exist
- Source URL with scheme
- Semantic type (knowledge/code)
- Link name, tags, description
- Type-specific config
- Shared across team via git

**Cache Index** (`~/.cache/sow/index.json`):
- Reference ID
- Type and cache path
- Type-specific metadata (SHAs, timestamps, etc.)
- Last updated timestamp
- Which local repos use this ref
- Per-machine, never shared

**Local Index** (`.sow/refs/index.local.json`):
- Local-only references (not shared with team)
- Uses `file://` URLs for local paths
- Never committed to git
- Same structure as committed index

**Separation Rationale**: Avoid git conflicts from timestamps/SHAs, enable team sharing without coupling to specific commits, each developer can be at different cache states, clean separation of concerns.

---

## Reference Types

### Semantic Types (What Content Represents)

**Knowledge** (`--semantic knowledge`):
- Documentation, guides, policies, standards
- Typically markdown files
- Consulted when making decisions
- Agents check these to ensure compliance with team standards

**Code** (`--semantic code`):
- Implementation examples, patterns, reference code
- Working source code
- Studied and adapted (not directly imported)
- Agents examine these to understand implementation patterns

### Structural Types (How Content is Accessed)

#### Git Type

**URL Schemes**: `git+https://`, `git+ssh://`, or shorthand `git@host:path`

**Purpose**: Clone git repositories for style guides, conventions, shared code examples.

**Config Options**:
- `branch` - Branch name (optional, defaults to repo default)
- `path` - Subpath within repo (optional, defaults to root)

**Requirements**: Git binary in PATH

**Operations**:
- Cache via `git clone`
- Update via `git pull`
- Staleness via commit SHA comparison

**Examples**:
```bash
# HTTPS transport
git+https://github.com/acme/style-guides

# SSH transport (full)
git+ssh://git@github.com/acme/style-guides

# SSH transport (shorthand, auto-converted)
git@github.com:acme/style-guides
```

#### File Type

**URL Scheme**: `file:///` (absolute path)

**Purpose**: Reference local directories for WIP documentation, experiments, personal notes.

**Config Options**: None (path in URL)

**Requirements**: None (always enabled)

**Operations**:
- Cache via symlink or copy
- No updates (local files)
- Never stale

**Examples**:
```bash
file:///Users/josh/docs
file:///home/user/team-wiki
```

#### Future Types

The type system is extensible. Future types might include:
- `web+https://` - Scraped web documentation
- `api+https://` - OpenAPI specs fetched from APIs
- `s3://` - Content from S3 buckets
- `http+archive://` - Archived documentation sites

Each type implements the same interface and follows same caching requirements.

---

## Type System

### Type Interface

All types implement:

```go
type RefType interface {
    // Check if type available on local system
    IsEnabled(ctx context.Context) (bool, error)

    // One-time initialization for the type
    Init(ctx context.Context) error

    // Fetch and cache content locally
    Cache(ctx context.Context, ref Ref) (string, error)

    // Update existing cached content
    Update(ctx context.Context, ref Ref) error

    // Check if cached version outdated
    IsStale(ctx context.Context, ref Ref) (bool, error)

    // Get cache path for filesystem access
    CachePath(ref Ref) string

    // Remove cached content
    Cleanup(ctx context.Context, ref Ref) error

    // Validate type-specific config before adding
    ValidateConfig(config map[string]any) error
}
```

### Type Registry

Central registry maps type names to implementations:

```go
var registry = map[string]RefType{
    "git":  &GitType{},
    "file": &FileType{},
}
```

Types can be queried for enablement:

```bash
sow refs init
# Output:
# ✓ Git type enabled
# ✓ File type enabled
# ⚠ Skipped 2 refs requiring git (git binary not found)
```

### Type Enablement

Types declare requirements and check if system meets them:
- **git type**: Checks for `git` binary in PATH
- **file type**: Always enabled (native filesystem)

Disabled types:
- Skipped during `sow refs init` (with warnings)
- Commands for disabled types show helpful errors
- Refs requiring disabled types remain in index (portable across machines)

---

## CLI Commands

### Command Structure

Type-specific commands follow pattern: `sow refs <type> <command>`

```bash
sow refs git add <url> [flags]
sow refs git status [id]
sow refs git update [id]

sow refs file add <path> [flags]

sow refs init              # Initialize all enabled types
sow refs list [filters]    # List across all types
sow refs remove <id>       # Remove any type
sow refs status            # Status across all types
```

### Git Type Commands

#### sow refs git add

Add git repository reference.

**Usage**:
```bash
sow refs git add <url> [flags]
```

**Arguments**:
- `url` - Git repository URL with scheme:
  - `git+https://github.com/org/repo`
  - `git+ssh://git@github.com/org/repo`
  - `git@github.com:org/repo` (auto-converted)

**Flags**:
- `--branch <name>` - Branch name (optional, defaults to repo default)
- `--path <subpath>` - Subdirectory within repo (optional, "" for root)
- `--link <name>` - Symlink name in `.sow/refs/` (required)
- `--semantic <type>` - knowledge or code (required)
- `--tags <tag>...` - Topic tags (repeat for multiple)
- `--description "<text>"` - One-sentence description (required)
- `--summary "<text>"` - 2-3 sentence summary (optional)
- `--local` - Add to local index only (not shared with team)

**Example**:
```bash
sow refs git add git+https://github.com/acme/style-guides \
  --branch main \
  --path python/ \
  --link python-style \
  --semantic knowledge \
  --tags formatting,naming,testing \
  --description "Python coding standards" \
  --summary "PEP 8 formatting, naming conventions, and testing patterns"
```

**Behavior**:
1. Parse URL and infer type (git)
2. Validate URL format and config
3. Clone to `~/.cache/sow/refs/git/{id}/` if needed
4. Checkout specified branch
5. Add entry to committed or local index
6. Create symlink from `.sow/refs/{link}/` to cache path
7. Update cache index

#### sow refs git status

Check if git refs up to date.

**Usage**:
```bash
sow refs git status [id]
```

**Arguments**:
- `id` - Optional specific ref to check (checks all git refs if omitted)

**Output**:
- Current commit SHA
- Remote commit SHA
- Status: current, behind by N commits, or error
- Recent commits if behind

**Example**:
```bash
sow refs git status python-style
# Output:
# python-style (git+https://github.com/acme/style-guides)
#   Local:  a1b2c3d (2024-10-10)
#   Remote: e4f5g6h (2024-10-15)
#   Status: Behind by 3 commits
#     - e4f5g6h: Update Python 3.12 patterns
#     - d4e5f6g: Add async testing guidelines
#     - c3d4e5f: Fix typo in naming section
```

#### sow refs git update

Pull latest changes from git remote.

**Usage**:
```bash
sow refs git update [id]
```

**Arguments**:
- `id` - Optional specific ref to update (updates all behind git refs if omitted)

**Behavior**:
1. Git pull in cache
2. Update cache index with new SHA
3. On Windows: rsync to all consuming repos

**Example**:
```bash
sow refs git update python-style
# Output:
# ✓ Updated python-style
#   Pulled 3 commits (a1b2c3d → e4f5g6h)
#   Files changed: 5 modified, 2 added
```

### File Type Commands

#### sow refs file add

Add local directory reference.

**Usage**:
```bash
sow refs file add <path> [flags]
```

**Arguments**:
- `path` - Local filesystem path (converted to file:// URL)

**Flags**:
- `--link <name>` - Symlink name in `.sow/refs/` (required)
- `--semantic <type>` - knowledge or code (required)
- `--tags <tag>...` - Topic tags (repeat for multiple)
- `--description "<text>"` - One-sentence description (required)
- `--summary "<text>"` - 2-3 sentence summary (optional)

**Note**: File refs are always local (--local flag implied)

**Example**:
```bash
sow refs file add /Users/josh/docs \
  --link local-docs \
  --semantic knowledge \
  --tags wip \
  --description "Local work-in-progress documentation"
```

**Behavior**:
1. Validate path exists
2. Convert to file:// URL
3. Add entry to local index
4. Create symlink from `.sow/refs/{link}/` to path

### Global Commands

#### sow refs init

Initialize refs structure and cache all references.

**Usage**:
```bash
sow refs init
```

**Behavior**:
1. Check which types are enabled
2. Read committed and local indexes
3. For each ref with enabled type:
   - Clone/cache if not already cached
   - Create symlink to `.sow/refs/{link}/`
   - Update cache index
4. Warn about skipped refs (disabled types)

**Output**:
```
✓ Git type enabled
✓ File type enabled
✓ Cached 3 git refs
✓ Cached 1 file ref
⚠ Skipped 1 web ref (web type not enabled)
```

**Use Case**: Run after cloning repository to set up refs

#### sow refs list

Display configured references.

**Usage**:
```bash
sow refs list [flags]
```

**Flags**:
- `--type <type>` - Filter by structural type (git, file)
- `--semantic <type>` - Filter by semantic type (knowledge, code)
- `--tags <tag>...` - Filter by tags
- `--local` - Show only local refs
- `--remote` - Show only committed refs (default: both)
- `--format <fmt>` - Output format: table (default), json, yaml

**Output** (table format):
```
ID       TYPE  SEMANTIC   LINK           SOURCE
abc123   git   knowledge  python-style   git+https://github.com/acme/style-guides
def456   file  knowledge  local-docs     file:///Users/josh/docs
```

#### sow refs remove

Remove reference from repository.

**Usage**:
```bash
sow refs remove <id> [flags]
```

**Arguments**:
- `id` - Reference ID to remove

**Flags**:
- `--prune-cache` - Also remove from cache if no other repos use it

**Behavior**:
1. Confirm with user (unless `--force`)
2. Remove entry from index (committed or local)
3. Remove symlink from `.sow/refs/{link}/`
4. Update cache index (remove from used_by)
5. Optionally prune cache if unused

**Example**:
```bash
sow refs remove python-style
# Output:
# Remove ref 'python-style' (git+https://github.com/acme/style-guides)?
# This will remove the symlink but preserve the cache. [y/N]: y
# ✓ Removed python-style
# Note: Cache preserved (used by 2 other repositories)
```

#### sow refs status

Show status across all ref types.

**Usage**:
```bash
sow refs status
```

**Output**: Aggregated status from all enabled types showing current vs remote state.

---

## Workflows

### Adding Remote Git Reference

User requests via natural language or slash command → Orchestrator examines content and analyzes to determine tags/description → Orchestrator calls `sow refs git add` with appropriate parameters → CLI clones to cache, adds entry to index, creates symlink → Orchestrator confirms installation.

### Adding Local File Reference

User provides local path → Orchestrator calls `sow refs file add` → CLI validates path exists, adds entry to local index, creates symlink → Orchestrator confirms installation.

### Fresh Clone Setup

Developer clones repository with existing refs → Run `sow refs init` → CLI checks enabled types, clones/caches refs, creates symlinks → Warns about any refs requiring disabled types.

### Checking for Updates

User requests update check → Orchestrator calls `sow refs status` or `sow refs git status` → CLI fetches from remotes and compares state → Orchestrator reports staleness and suggests updates.

### Updating References

User requests update → Orchestrator calls `sow refs git update` → CLI pulls in cache and updates index → On Windows: rsyncs to all consuming repos → Orchestrator reports changes pulled.

### Working with Disabled Types

Developer on machine without git binary → `sow refs init` skips git refs with warning → User installs git → Run `sow refs init` again → Git refs now cached and available.

---

## Platform Differences

### Unix/Linux/macOS

Native symlink support. Updates to cache automatically visible via symlinks (no additional sync needed). Type commands only need to update cache.

### Windows

Symlink requires Developer Mode or Administrator privileges. Fallback strategy: copy instead of symlink when privileges unavailable. Updates must rsync from cache to all consuming repos. CLI finds consuming repos in cache index and syncs updated paths. Detection automatic (CLI detects platform and chooses strategy).

---

## Orchestrator Integration

### Type-Specific Slash Commands

Each ref type has a slash command explaining usage to AI:
- `.claude/commands/refs/overview.md` - Overall refs system
- `.claude/commands/refs/git.md` - How to use git refs
- `.claude/commands/refs/file.md` - How to use file refs

Orchestrator loads appropriate command when helping user manage refs.

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
