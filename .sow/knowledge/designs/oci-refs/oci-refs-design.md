# OCI Refs Distribution System Design

**Author**: Architecture Team
**Date**: 2025-10-30
**Status**: Proposed
**Related ADR**: [ADR-003: Use OCI Registries for Refs Distribution](../../adrs/003-oci-refs-distribution.md)

## Executive Summary

This design implements OCI registry-based distribution for sow refs, replacing the current git-based approach. Refs will be packaged as estargz-format OCI images containing content plus metadata (`.sow-ref.yaml`), published to standard registries (GitHub Container Registry, Docker Hub, self-hosted), and consumed via standard OCI pull operations. The `github.com/jmgilman/go/oci` library provides estargz support enabling pre-inspection of contents without downloading (via `ListFiles`), selective extraction using glob patterns, and structure validation before full pull. This reduces bandwidth, improves UX, and simplifies publishing workflows from multi-step git operations to single-command registry push.

## Overview

The current refs system uses git repositories for distributing external knowledge and code references. While functional, this approach has significant pain points: complex publishing workflows requiring full git operations, storage inefficiency from cloning entire git history, inability to inspect contents before downloading, and versioning ambiguity.

This design replaces git with OCI registry packages. Publishers run `sow refs publish <dir> <registry-url>:<tag>` to package directories as estargz OCI images and push to registries. Consumers can inspect contents without downloading (`sow refs inspect <url>` uses `ListFiles` to read only the table-of-contents), validate structure before committing to download, and selectively extract specific paths (`sow refs add <url> --path "docs/**"`). Full downloads extract to local cache and symlink into `.sow/refs/` as before, maintaining existing workspace structure.

The implementation leverages `github.com/jmgilman/go/oci` which provides production-ready OCI operations with built-in security (path traversal protection, size limits) and estargz format support added specifically for sow's selective extraction requirements.

## Goals and Non-Goals

### Goals

**G1: Simplify publishing workflow**
- Replace multi-step git workflow (init, add, commit, tag, push) with single `sow refs publish` command
- Success metric: Publishing a ref takes < 30 seconds end-to-end for 10MB ref

**G2: Enable pre-inspection without download**
- Users can view ref contents (file list) and metadata before downloading
- Success metric: `sow refs inspect` returns results in < 3 seconds without downloading ref

**G3: Support selective extraction**
- Download only needed files via glob patterns instead of entire ref
- Success metric: Downloading 10% of large ref (using `--path`) takes < 15% of full download time

**G4: Improve storage efficiency**
- OCI compression and absence of git metadata reduces bandwidth and cache size
- Success metric: Average ref size reduced by > 40% compared to git clone equivalent

**G5: Immutable versioning**
- Tags and digests provide reliable, immutable versioning
- Success metric: Digest pinning guarantees bit-for-bit identical content on repeated pulls


### Non-Goals

**NG1: Centralized registry hosting** - Sow will not operate a central OCI registry. Users bring their own (ghcr.io, Docker Hub, self-hosted).

**NG2: Registry-side search** - Discovery via marketplace system (separate design), not registry queries.

**NG3: Ref signing/verification** - Cryptographic signing deferred to future enhancement.

**NG4: Multi-platform refs** - Platform-specific content (linux/amd64, darwin/arm64) not needed for static file refs.

**NG5: Automatic ref updates** - Users explicitly update refs with `sow refs update`, no automatic polling.

## Background

### Motivation

The October 2025 refs improvement exploration (`improve-refs-2025-10`) identified critical limitations in git-based distribution:

**Complex publishing**: Publishing a ref requires initializing git repo, adding files, committing, tagging with version, pushing to remote, managing authentication. For simple documentation refs, this is heavyweight process deterring contributions.

**Storage waste**: Git clones download full repository history. A 5MB documentation ref with 50 commits might require 25MB clone. Users pay bandwidth and storage costs for history irrelevant to content consumption.

**No pre-inspection**: Users must clone entire repository to understand contents. Bad refs (wrong structure, missing docs) discovered only after full download.

**No selective access**: Need one file from large ref? Clone entire repository. No mechanism for partial extraction.

**Versioning ambiguity**: Git tags can be force-pushed (though discouraged). Branch names move. SHAs work but lack human readability.

### Current State

Refs are declared in `.sow/refs/index.json` with git repository URLs:
```json
{
  "id": "go-standards",
  "source": "git+https://github.com/myorg/team-docs",
  "semantic": "knowledge",
  "link": "go-standards",
  "config": {"branch": "main", "path": "docs/"}
}
```

The `sow refs add` command clones the repository into `~/.cache/sow/refs/<id>-<hash>/`, checks out the specified branch, and creates symlink from `.sow/refs/<link>` to cache directory. Updates require git pull operations.

### Requirements

**Functional Requirements**:
- FR1: Package arbitrary directory as OCI image with metadata
- FR2: Publish OCI image to standard registries (ghcr.io, Docker Hub, Harbor, Nexus)
- FR3: Inspect OCI image contents without downloading (list files, read metadata)
- FR4: Validate `.sow-ref.yaml` structure before download
- FR5: Download entire OCI image to local cache
- FR6: Download subset of files using glob patterns
- FR7: Support version tags (v1.0.0) and digest pinning (@sha256:...)

**Non-Functional Requirements**:
- NFR1: Publishing 10MB ref completes in < 30 seconds
- NFR2: Inspection (ListFiles) completes in < 3 seconds
- NFR3: Full download and extraction of 10MB ref completes in < 15 seconds
- NFR4: Selective download (10% of files) completes in < 3 seconds
- NFR5: OCI images are 40-60% smaller than equivalent git clones
- NFR6: Security: path traversal protection, size limits, permission sanitization
- NFR7: Registry authentication via Docker credential chain (transparent to user)

### Constraints

**Technical Constraints**:
- TC1: Must use `github.com/jmgilman/go/oci` library (provides estargz support)
- TC2: Must produce estargz-format images (not standard tar.gz) for selective extraction
- TC3: Must be compatible with standard OCI registries (no custom server modifications)
- TC4: Symlink structure in `.sow/refs/` unchanged (maintains compatibility with existing workflows)

**Business Constraints**:
- BC1: Free tier registries (ghcr.io) must be viable for typical usage

## Design

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Publishing Flow                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Local Directory              OCI Registry                       │
│  ┌──────────────┐            ┌────────────────┐                │
│  │ docs/        │            │ ghcr.io/       │                │
│  │ examples/    │   publish  │ org/ref:v1.0.0 │                │
│  │ .sow-ref.yaml├───────────>│                │                │
│  └──────────────┘            │ (estargz)      │                │
│         │                     │ + annotations  │                │
│         │                     └────────────────┘                │
│         v                                                         │
│  1. Validate .sow-ref.yaml                                       │
│  2. Package as estargz (jmgilman/go/oci)                        │
│  3. Extract metadata -> OCI annotations                          │
│  4. Push to registry                                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Consumption Flow (Inspect)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  OCI Registry                                                     │
│  ┌────────────────┐                                             │
│  │ ghcr.io/       │                                             │
│  │ org/ref:v1.0.0 │                                             │
│  └───────┬────────┘                                             │
│          │ ListFiles (downloads TOC only, ~few KB)              │
│          v                                                        │
│  ┌──────────────────────┐                                       │
│  │ TOC (Table of         │                                       │
│  │ Contents):            │                                       │
│  │ - docs/README.md      │                                       │
│  │ - docs/guide.md       │                                       │
│  │ - examples/demo.go    │                                       │
│  │ - .sow-ref.yaml       │                                       │
│  └──────────────────────┘                                       │
│          │                                                        │
│          v                                                        │
│  Display to user: "Ref contains 15 files, 2.3MB"                │
│  Show .sow-ref.yaml metadata                                     │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Consumption Flow (Full Install)               │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  OCI Registry          Cache                    Workspace        │
│  ┌────────────┐       ┌──────────────┐        ┌──────────┐    │
│  │ ghcr.io/   │ pull  │ ~/.cache/sow/│ symlink│ .sow/    │    │
│  │ org/ref    ├──────>│ refs/ref-abc/│<───────┤ refs/    │    │
│  │ :v1.0.0    │       │   docs/      │        │ myref    │    │
│  └────────────┘       │   examples/  │        └──────────┘    │
│                        │   .sow-ref...│                         │
│                        └──────────────┘                         │
│                                                                   │
│  Flow:                                                            │
│  1. Pull OCI image (entire contents)                            │
│  2. Extract to cache: ~/.cache/sow/refs/<id>-<short-digest>/   │
│  3. Create symlink: .sow/refs/<link> -> cache                   │
│  4. Index for search (FTS5, optional semantic)                  │
│  5. Add entry to .sow/refs/index.json                           │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                Consumption Flow (Selective Install)              │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  OCI Registry                                                     │
│  ┌────────────────┐                                             │
│  │ ghcr.io/       │                                             │
│  │ org/ref:v1.0.0 │                                             │
│  └───────┬────────┘                                             │
│          │ pull --path "docs/**/*.md" --path "examples/*.go"    │
│          │ (selective extraction via estargz with multiple globs)│
│          v                                                        │
│  ┌──────────────────────┐                                       │
│  │ Download only:        │                                       │
│  │ - docs/README.md      │                                       │
│  │ - docs/guide.md       │                                       │
│  │ - examples/demo.go    │                                       │
│  │ - examples/test.go    │                                       │
│  │ - .sow-ref.yaml       │                                       │
│  │                       │                                       │
│  │ Skip:                 │                                       │
│  │ - templates/ (not     │                                       │
│  │   matched by globs)   │                                       │
│  └──────────────────────┘                                       │
│          │                                                        │
│          v                                                        │
│  Extract to cache (partial content)                              │
│  Symlink to workspace                                            │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

The system consists of four primary workflows:

**Publishing**: Validate `.sow-ref.yaml`, package directory as estargz OCI image with `github.com/jmgilman/go/oci`, map metadata to OCI annotations, push to registry.

**Inspection**: Use `ListFiles` to download only the estargz table-of-contents (few KB), display file list and metadata to user without downloading actual content.

**Full installation**: Pull entire OCI image, extract to cache, create symlink, index for search.

**Selective installation**: Use estargz selective extraction with multiple glob patterns to download only matching files, extract to cache, create symlink.

### Component Breakdown

#### OCI Client Wrapper

**Responsibility**: Interface to `github.com/jmgilman/go/oci` library

**Key Behaviors**:
- Initialize OCI client with Docker credential chain
- Push estargz images to registries
- Pull images (full or selective) from registries
- List files via estargz TOC without full download
- Query image metadata (annotations) without download

**Dependencies**: `github.com/jmgilman/go/oci`, Docker credential chain

**Implementation Notes**:
- Wrap library calls in sow error types for consistent error handling
- Configure security features: max file size 100MB, max total size 1GB, 10k file limit
- Enable retry with exponential backoff (library built-in)

#### Ref Packager

**Responsibility**: Package local directory as estargz OCI image

**Key Behaviors**:
- Validate `.sow-ref.yaml` schema before packaging
- Apply exclusion patterns from `packaging.exclude` in manifest
- Create estargz archive (not standard tar.gz)
- Generate OCI annotations from `.sow-ref.yaml` fields
- Calculate content digest

**Dependencies**: OCI Client Wrapper, Schema Validator

**Implementation Notes**:
- Default exclusions: `.git/`, `.DS_Store`, `node_modules/`
- Preserve Unix permissions (sanitized by OCI client: no setuid/setgid)
- Include `.sow-ref.yaml` in root of image

#### Schema Validator

**Responsibility**: Validate `.sow-ref.yaml` conforms to schema

**Key Behaviors**:
- Validate required fields present (`schema_version`, `ref.title`, `ref.link`, `content.description`, `content.classifications`, `content.tags`)
- Validate field formats (semver for schema_version, kebab-case for ref.link, valid classification types)
- Validate semantic constraints (field length recommendations, date formats if present)
- Return detailed validation errors for user correction

**Dependencies**: CUE schema (`cli/schemas/ref_manifest.cue`), CUE validator

**Implementation Notes**:
- Use CUE schema for validation (consistent with existing sow schemas: `project_state.cue`, `refs_cache.cue`, etc.)
- Schema location: `cli/schemas/ref_manifest.cue` (embedded in CLI binary)
- Classification types defined in CUE enum: `tutorial`, `api-reference`, `guidelines`, `architecture`, `runbook`, `specification`, `reference`, `code-examples`, `code-templates`, `code-library`, `uncategorized`
- Link validation regex in CUE: `=~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"`
- Leverage existing `schemas.Validator` infrastructure from `cli/schemas/validator.go`

#### Inspector

**Responsibility**: Inspect OCI ref without downloading full content

**Key Behaviors**:
- Call `ListFiles` on OCI client to retrieve file list from estargz TOC
- Parse `.sow-ref.yaml` from TOC (read file content via selective extraction)
- Display file count, total size estimate, directory tree
- Show metadata (title, description, classifications, tags)
- Validate structure (`.sow-ref.yaml` exists and valid) before user commits to download

**Dependencies**: OCI Client Wrapper, Schema Validator

**Implementation Notes**:
- `ListFiles` downloads only TOC (~few KB), not file contents
- Reading `.sow-ref.yaml` via selective extraction downloads only that file (~1-5KB)
- Total bandwidth for inspection: < 10KB for typical ref

#### Ref Installer

**Responsibility**: Install ref to local cache and workspace

**Key Behaviors**:
- Pull OCI image (full or selective based on `--path` flags)
- Extract to cache directory: `~/.cache/sow/refs/<id>-<short-digest>/`
- Create symlink: `.sow/refs/<link>` -> cache directory
- Add entry to `.sow/refs/index.json` with metadata
- Trigger search indexing (FTS5, optional semantic)

**Dependencies**: OCI Client Wrapper, Cache Manager, Index Manager

**Implementation Notes**:
- Atomic extraction: use temp directory, move to final location on success
- Short digest: first 7 characters of SHA256 digest (for human-readable cache paths)
- Selective extraction: pass multiple glob patterns to OCI client, files matching any pattern extracted
- Always extract `.sow-ref.yaml` even with selective extraction (needed for metadata)

#### Cache Manager

**Responsibility**: Manage local ref cache

**Key Behaviors**:
- Allocate cache directories with naming convention: `<id>-<short-digest>`
- Check if ref already cached (digest-based)
- Prune unused cache entries (refs not symlinked from any workspace)
- Calculate cache statistics (total size, ref count)

**Dependencies**: Filesystem operations

**Implementation Notes**:
- Cache location: `~/.cache/sow/refs/` (respects `XDG_CACHE_HOME` on Linux)
- Pruning: `sow refs prune` removes refs not referenced by any workspace
- Aggressive pruning: `sow refs prune --all` removes entire cache

#### Index Manager

**Responsibility**: Maintain `.sow/refs/index.json`

**Key Behaviors**:
- Read/write `.sow/refs/index.json` with file locking
- Add ref entries with metadata extracted from `.sow-ref.yaml`:
  - Installation: `id`, `source`, `digest`, `link`, `installed_at`, `selective`
  - Ref metadata: `title`, `description`, `classifications`, `tags`, `authors`, `license`
- Update ref entries (e.g., version update changes digest and metadata)
- Remove ref entries
- Query refs by ID, link, tags, or classifications

**Dependencies**: Filesystem operations, JSON marshaling

**Implementation Notes**:
- File locking prevents concurrent modification races
- Schema version in index for future migrations
- Metadata cached in index enables fast queries without reading ref files


### Data Models

#### .sow-ref.yaml Schema

```yaml
# Schema version (semver)
schema_version: "1.0.0"

# Core identification
ref:
  title: "Go Team Standards"              # Human-readable name (5-100 chars)
  link: "go-standards"                     # Symlink name (kebab-case, alphanumeric+hyphens)

# Content description
content:
  description: "Team Go coding conventions and best practices."  # 50-200 chars

  summary: |                               # Optional: 2-5 sentences, markdown
    Complete reference for Go development covering coding standards,
    testing practices, code review guidelines, and architecture patterns.

  classifications:                         # At least one required
    - type: guidelines                     # Predefined type
      description: "Go coding standards for the team"  # 20-200 chars

  tags:                                    # At least one required
    - golang
    - conventions
    - testing

# Authorship and provenance (optional)
provenance:
  authors: ["Platform Team"]
  created: "2024-01-15T10:00:00Z"         # RFC 3339
  updated: "2025-01-30T15:30:00Z"         # Auto-updated on publish
  source: "https://github.com/myorg/team-docs"
  license: "MIT"

# Publishing configuration (optional)
packaging:
  exclude:
    - "*.draft.md"
    - ".DS_Store"
    - "tmp/"

# LLM usage hints (optional)
hints:
  suggested_queries:
    - "How should I structure error handling?"
  primary_files:
    - "README.md"
    - "standards/golang.md"

# Organization-specific metadata (optional, freeform)
metadata:
  team: "platform"
  audience: "all-engineers"
```

**OCI Annotation Mapping**:

| .sow-ref.yaml Field | OCI Annotation Key |
|---------------------|-------------------|
| `ref.title` | `org.opencontainers.image.title` |
| `content.description` | `org.opencontainers.image.description` |
| `provenance.authors` | `org.opencontainers.image.authors` (JSON array) |
| `provenance.created` | `org.opencontainers.image.created` |
| `provenance.source` | `org.opencontainers.image.source` |
| `provenance.license` | `org.opencontainers.image.licenses` |
| `content.classifications` | `com.sow.ref.classifications` (JSON) |
| `content.tags` | `com.sow.ref.tags` (comma-separated) |
| `ref.link` | `com.sow.ref.link` |

#### .sow/refs/index.json Schema

```json
{
  "version": "1.0.0",
  "refs": [
    {
      "id": "go-standards",                          // Unique ID (user-provided or generated)
      "source": "ghcr.io/myorg/go-standards:v1.0.0", // OCI URL with tag
      "digest": "sha256:abc123...",                  // Full digest (for verification)
      "link": "go-standards",                        // Symlink name in .sow/refs/

      // Metadata from .sow-ref.yaml
      "title": "Go Team Standards",                  // From ref.title
      "description": "Team Go coding conventions and best practices.", // From content.description
      "classifications": [                           // From content.classifications
        {
          "type": "guidelines",
          "description": "Go coding standards for the team"
        }
      ],
      "tags": ["golang", "conventions", "testing"], // From content.tags
      "authors": ["Platform Team"],                  // From provenance.authors (optional)
      "license": "MIT",                              // From provenance.license (optional)

      // Installation metadata
      "installed_at": "2025-01-30T15:30:00Z",       // Installation timestamp
      "selective": {                                  // Present if installed with --path
        "globs": ["docs/**/*.md", "examples/*.go"], // Array of glob patterns
        "partial": true
      }
    }
  ]
}
```

### APIs and Interfaces

#### CLI Commands

**Publishing**:
```bash
# Publish ref to registry
sow refs publish <directory> <registry-url>:<tag>

# Example
sow refs publish ./team-docs ghcr.io/myorg/go-standards:v1.0.0

# Options
  --registry-auth <credentials>   # Override Docker credential chain
  --latest                        # Also push :latest tag
```

**Inspection**:
```bash
# Inspect ref without downloading
sow refs inspect <registry-url>:<tag>

# Example
sow refs inspect ghcr.io/myorg/go-standards:v1.0.0

# Output shows:
# - File list and directory tree
# - Total size estimate
# - .sow-ref.yaml metadata (title, description, classifications, tags)
# - Validation status
```

**Installation**:
```bash
# Install full ref
sow refs add <registry-url>:<tag> [--link <name>]

# Example
sow refs add ghcr.io/myorg/go-standards:v1.0.0 --link team-go

# Selective install (glob patterns - can specify multiple)
sow refs add <registry-url>:<tag> --path <glob> [--path <glob2>...] [--link <name>]

# Example: Only markdown docs and Go examples
sow refs add ghcr.io/myorg/go-standards:v1.0.0 --path "docs/**/*.md" --path "examples/*.go" --link team-go

# Digest pinning
sow refs add ghcr.io/myorg/go-standards@sha256:abc123...

# Options
  --link <name>        # Custom symlink name (default: derived from URL)
  --path <glob>        # Glob pattern for selective extraction (can be specified multiple times)
  --force              # Overwrite existing ref
```

**Management**:
```bash
# List installed refs
sow refs list
# Output: table with ID, SOURCE, VERSION, LINK, STATUS

# Update ref to latest version
sow refs update <id>

# Remove ref
sow refs remove <id> [--prune-cache]

# Cache management
sow refs prune              # Remove unused cache entries
sow refs prune --all        # Remove entire cache
sow refs cache-info         # Show cache size and stats
```

## Testing Strategy

**Unit Tests**:
- CUE schema: verify schema compiles, test valid/invalid manifests
- Schema validator: all field combinations, error message formatting
- Annotation mapper: verify `.sow-ref.yaml` to OCI annotations mapping
- Index manager: read/write, concurrent access (file locking)
- Cache manager: path generation, pruning logic, stats calculation

**Integration Tests**:
- Full publishing workflow: directory -> OCI registry (test registry or Docker local registry)
- Full installation workflow: OCI registry -> cache -> symlink
- Selective installation: verify only matching files extracted (multiple globs with OR logic)
- Inspection workflow: verify ListFiles returns correct data without full download

**End-to-End Tests**:
- Publish to ghcr.io test account, install in clean environment
- Test authentication flows (authenticated and anonymous)
- Test error scenarios (network failure, invalid manifest, registry errors)
- Test update workflow (publish v1, install, publish v2, update)

**Performance Tests**:
- Benchmark publishing: 10MB, 100MB refs (measure package + push time)
- Benchmark inspection: measure ListFiles latency for various ref sizes
- Benchmark installation: full and selective, measure pull + extract time
- Benchmark compression ratios: various content types (markdown, code, mixed)

**Security Tests**:
- Path traversal: attempt malicious refs with `../`, verify rejection
- Size limits: attempt oversized files/refs, verify rejection
- Permission sanitization: verify setuid/setgid stripped on extraction

## Implementation Plan

### Phase 1: Core OCI Integration (Week 1, 5 days)

**Deliverables**:
- OCI client wrapper integrated with `github.com/jmgilman/go/oci`
- CUE schema for `.sow-ref.yaml` (`cli/schemas/ref_manifest.cue`)
- Schema validator using CUE
- Ref packager with estargz support
- `sow refs publish` command functional
- `sow refs inspect` command functional (ListFiles + metadata display)

**Milestones**:
- Day 1: OCI client wrapper + CUE schema + validator (leverage existing `schemas.Validator`)
- Day 2: Ref packager with annotation mapping
- Day 3: Publish command implementation
- Day 4: Inspect command implementation
- Day 5: Testing and bug fixes

**Dependencies**: None (greenfield)

**Risks**: `jmgilman/go/oci` library API changes. Mitigation: Pin version, test thoroughly.

### Phase 2: Consumption Workflow (Week 2, 5 days)

**Deliverables**:
- Installer component with full and selective extraction (multiple glob support)
- Cache manager
- Index manager updates for OCI refs
- `sow refs add` command (full and selective with `--path` flags)
- Management commands: `sow refs list`, `sow refs update`, `sow refs remove`, `sow refs prune`

**Milestones**:
- Day 1: Installer core logic (pull + extract)
- Day 2: Selective extraction with multiple glob support (OR logic)
- Day 3: Cache and index management
- Day 4: `sow refs add` and management commands
- Day 5: Testing full consumption flow, error handling polish

**Dependencies**: Phase 1 (publishing must work to test consumption)

**Risks**: Estargz selective extraction edge cases with multiple globs. Mitigation: Comprehensive test cases for glob combinations.

**Timeline**: 2 weeks total for complete implementation (10 working days).

## References

- **ADR-003**: [Use OCI Registries for Refs Distribution](../../adrs/003-oci-refs-distribution.md)
- **Exploration**: `../../explorations/improve-refs-2025-10/` (October 2025)
  - Overview: `00-overview.md`
  - OCI research: `oci-refs-distribution.md`
  - Summary: `SUMMARY.md`
- **OCI Library**: [`github.com/jmgilman/go/oci`](https://github.com/jmgilman/go) (estargz support)
- **OCI Specification**: [Open Container Initiative](https://opencontainers.org/)
- **Estargz Specification**: [Stargz Snapshotter](https://github.com/containerd/stargz-snapshotter)
- **Arc42 Sections**: Building Blocks (Section 5), Concepts (Section 8), Decisions (Section 9)
- **C4 Diagram**: Container view with OCI integration

## Future Considerations

**Ref signing**: Cryptographic verification via Cosign or Notation for supply chain security

**Registry mirrors**: Configure fallback registries for air-gapped environments

**Ref collections**: Group related refs, install with single command

**Offline mode**: Export/import refs as tarballs for fully offline operation

**Delta pulls**: Download only changed layers for updates (bandwidth optimization)

**Ref templates**: Generate new refs from templates with `sow refs init <template>`
