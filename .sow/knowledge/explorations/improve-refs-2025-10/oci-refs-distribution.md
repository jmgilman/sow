# OCI-Based Refs Distribution

## Summary

Use OCI registries (ghcr.io, Docker Hub, etc.) to distribute refs as immutable, versioned packages. The `github.com/jmgilman/go/oci` module provides production-ready functionality with security built-in. Refs include a `.sow-ref.yaml` manifest with metadata, classifications, and LLM hints. Publishing workflow: package directory → push to registry. Consumption: pull from registry → extract to cache → symlink to workspace.

## Why OCI for Refs?

**Current git-based approach limitations**:
- Complex workflow (clone, checkout branch, track commits)
- Large clones for small content (full git history)
- Difficult versioning (tags, branches, SHAs)
- Authentication complexity (SSH keys, tokens)

**OCI advantages**:
- **Immutable versioning**: Tags like `v1.0.0`, `v2.1.0`, `latest`
- **Digest pinning**: Lock to specific content (`@sha256:...`)
- **Efficient storage**: Optimized compression, no git overhead
- **Native auth**: Docker credentials work automatically
- **Simple API**: Push and pull, that's it
- **Registry ecosystem**: Free hosting (ghcr.io), self-hosting options

## OCI Module Capabilities

**Package**: `github.com/jmgilman/go/oci`

### Core Operations

```go
import "github.com/jmgilman/go/oci"

// Publish ref
client, _ := oci.New()
err := client.Push(ctx, "./team-docs", "ghcr.io/myorg/team-docs:v1.0.0")

// Install ref
err := client.Pull(ctx, "ghcr.io/myorg/team-docs:v1.0.0", cacheDir)
```

### Security Features (Built-in)

**Path traversal protection**:
- Blocks `..`, absolute paths, symlink escapes
- Prevents malicious archives from escaping target directory

**Size limits**:
- Per-file limit: 100MB default
- Total size limit: 1GB default
- File count limit: 10k files default
- Prevents decompression bombs

**Permission sanitization**:
- Strips setuid/setgid bits
- Safe file permissions on extraction

**Authentication**:
- Docker credential chain (`~/.docker/config.json`)
- Credential helper integration (macOS Keychain, etc.)
- Static credentials override for specific registries

**Reliability**:
- Automatic retry with exponential backoff
- Atomic extraction (temp dir → target, rollback on failure)
- Progress callbacks for long operations

### Metadata via OCI Annotations

OCI images support arbitrary key-value annotations. During publishing, sow copies `.sow-ref.yaml` fields to standard OCI annotations:

```go
annotations := map[string]string{
    "org.opencontainers.image.title": "Go Team Standards",
    "org.opencontainers.image.version": "v1.0.0",
    "org.opencontainers.image.description": "Team Go coding conventions",
    "org.opencontainers.image.authors": "Platform Team",
    "org.opencontainers.image.source": "https://github.com/myorg/team-docs",
    "org.opencontainers.image.licenses": "MIT",

    // Sow-specific annotations
    "com.sow.ref.classifications": `[{"type":"guidelines","description":"Go coding standards"}]`,
    "com.sow.ref.tags": "golang,conventions,standards",
    "com.sow.ref.link": "go-standards",
}

client.Push(ctx, refPath, registryURL, oci.WithAnnotations(annotations))
```

**Benefits**:
- Queryable without pulling image
- Standard OCI tools can display metadata
- Backward compatible with non-sow tools

## Ref Metadata Schema

**File**: `.sow-ref.yaml` (included in OCI image)

### Complete Schema

```yaml
# Schema version for future migrations
schema_version: "1.0.0"

# Core identification
ref:
  # Human-readable name
  title: "Go Team Standards"

  # Symlink name in .sow/refs/ (kebab-case)
  link: "go-standards"

# Content description
content:
  # One-sentence summary (50-150 chars)
  description: "Comprehensive Go development guidelines for the engineering team."

  # Detailed multi-line description (2-5 sentences, markdown supported)
  summary: |
    Complete reference for Go development covering coding standards,
    testing practices, code review guidelines, and architecture patterns.
    Includes runnable examples and templates.

  # Content classifications (at least one required)
  classifications:
    - type: guidelines
      description: "Contains coding standards and conventions for Go development"

    - type: code-examples
      description: "Includes runnable Go examples demonstrating testing patterns"

    - type: code-templates
      description: "Provides service templates and boilerplate for new projects"

  # Topic keywords (at least one required)
  tags:
    - golang
    - conventions
    - testing
    - code-review
    - architecture

# Authorship and provenance (optional)
provenance:
  authors:
    - "Platform Team"
  created: "2024-01-15T10:00:00Z"
  updated: "2025-01-30T15:30:00Z"  # Auto-updated on publish
  source: "https://github.com/myorg/team-docs"
  license: "MIT"

# Publishing configuration (optional)
packaging:
  exclude:
    - "*.draft.md"
    - ".DS_Store"
    - "tmp/"
    - "**/.git"

# LLM usage hints (optional)
hints:
  # Example questions LLM can answer
  suggested_queries:
    - "How should I structure error handling?"
    - "What are the testing requirements?"
    - "Show me the code review checklist"

  # Key files to read first
  primary_files:
    - "README.md"
    - "standards/golang.md"
    - "templates/service.go"

# Organization-specific metadata (optional, freeform)
metadata:
  team: "platform"
  compliance: "sox-approved"
  audience: "all-engineers"
```

### Classification Types

**Knowledge Classifications**:
- `tutorial` - Step-by-step learning content
- `api-reference` - API documentation
- `guidelines` - Standards, conventions, style guides
- `architecture` - Design docs, ADRs, diagrams
- `runbook` - Operational procedures, troubleshooting
- `specification` - Technical specs, RFCs
- `reference` - Quick reference, cheat sheets

**Code Classifications**:
- `code-examples` - Runnable examples, snippets
- `code-templates` - Boilerplate, scaffolding
- `code-library` - Reusable library code

**Fallback**:
- `uncategorized` - Cannot be reasonably classified

### Required vs Optional Fields

**Required**:
- `schema_version`
- `ref.title`
- `ref.link`
- `content.description`
- `content.classifications` (at least one)
- `content.tags` (at least one)

**Optional**:
- `content.summary`
- `provenance.*`
- `packaging.exclude`
- `hints.*`
- `metadata.*`

### Validation Rules

**Format validation**:
- `schema_version`: Must match semver pattern (`^[0-9]+\.[0-9]+\.[0-9]+$`)
- `ref.link`: Must match `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- `content.classifications[].type`: Must be predefined classification type
- `provenance.created`, `provenance.updated`: Must be valid RFC 3339 if present

**Semantic validation**:
- `ref.title`: 5-100 characters recommended
- `content.description`: 50-200 characters recommended
- `content.classifications[].description`: 20-200 characters, required
- `content.tags`: Lowercase, alphanumeric with hyphens recommended

### OCI Annotation Mapping

| .sow-ref.yaml Field | OCI Annotation Key |
|---------------------|-------------------|
| `ref.title` | `org.opencontainers.image.title` |
| `content.description` | `org.opencontainers.image.description` |
| `provenance.authors` | `org.opencontainers.image.authors` |
| `provenance.created` | `org.opencontainers.image.created` |
| `provenance.source` | `org.opencontainers.image.source` |
| `provenance.license` | `org.opencontainers.image.licenses` |
| `content.classifications` | `com.sow.ref.classifications` (JSON) |
| `content.tags` | `com.sow.ref.tags` (comma-separated) |
| `ref.link` | `com.sow.ref.link` |

## Publishing Workflow

### 1. Create Ref Content

```bash
mkdir team-docs
cd team-docs

# Create content
cat > README.md <<EOF
# Team Go Standards
Our team's Go development guidelines.
EOF

mkdir standards
cat > standards/golang.md <<EOF
# Go Coding Standards
## Error Handling
Always wrap errors with context...
EOF
```

### 2. Generate Metadata

**Manual creation**:
```bash
cat > .sow-ref.yaml <<EOF
schema_version: "1.0.0"
ref:
  title: "Go Team Standards"
  link: "go-standards"
content:
  description: "Team Go coding conventions and best practices."
  classifications:
    - type: guidelines
      description: "Go coding standards for the team"
  tags:
    - golang
    - conventions
provenance:
  authors: ["Platform Team"]
  license: "MIT"
EOF
```

**LLM-assisted generation**:
```bash
sow refs index ./team-docs

# Claude analyzes content and generates metadata:
# - Scans directory structure
# - Reads representative files (README, main docs)
# - Detects topics, languages, frameworks
# - Infers classifications from content patterns
# - Generates suggested_queries based on structure
# - Interactive refinement with user
```

### 3. Publish to Registry

```bash
# Package and publish
sow refs publish ./team-docs ghcr.io/myorg/go-standards:v1.0.0

# Internally:
# 1. Validate .sow-ref.yaml schema
# 2. Apply packaging.exclude patterns
# 3. Create tar.gz archive
# 4. Extract annotations from .sow-ref.yaml
# 5. Push to OCI registry with annotations
```

**GitHub Actions workflow**:
```yaml
name: Publish Ref

on:
  push:
    tags:
      - 'v*'

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install sow
        run: |
          curl -fsSL https://sow.sh/install.sh | sh

      - name: Login to GitHub Container Registry
        run: echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Publish ref
        run: |
          sow refs publish . ghcr.io/${{ github.repository }}:${{ github.ref_name }}
```

### 4. Version Management

**Semantic versioning**:
```bash
# Publish specific versions
sow refs publish . ghcr.io/myorg/go-standards:v1.0.0
sow refs publish . ghcr.io/myorg/go-standards:v1.1.0
sow refs publish . ghcr.io/myorg/go-standards:v2.0.0

# Update latest tag
sow refs publish . ghcr.io/myorg/go-standards:latest
```

**Digest pinning** (reproducibility):
```bash
# Query digest
digest=$(sow refs inspect ghcr.io/myorg/go-standards:v1.0.0 --format '{{.Digest}}')

# Reference by digest (immutable)
sow refs add ghcr.io/myorg/go-standards@sha256:abc123...
```

## Consumption Workflow

### 1. Install Ref

```bash
sow refs add ghcr.io/myorg/go-standards:v1.0.0 --link team-go

# Internally:
# 1. Pull OCI image from registry
# 2. Extract to ~/.cache/sow/refs/go-standards-abc123/
# 3. Load .sow-ref.yaml metadata
# 4. Create symlink: .sow/refs/team-go → cache
# 5. Add to .sow/refs/index.json
# 6. Index for search (FTS5 + optional semantic)
```

**Authentication** (private registries):
```bash
# Use Docker credentials
docker login ghcr.io

# sow automatically uses Docker credential chain
sow refs add ghcr.io/myorg/private-ref:v1.0.0
```

### 2. Update Ref

```bash
sow refs update team-go

# Internally:
# 1. Query registry for latest digest
# 2. Compare with cached digest
# 3. If different, pull new version
# 4. Update cache and symlink
# 5. Re-index for search
```

### 3. List Installed Refs

```bash
sow refs list

# Output:
# NAME        SOURCE                                   VERSION  STATUS
# team-go     ghcr.io/myorg/go-standards              v1.0.0   current
# api-guide   ghcr.io/myorg/api-patterns              v2.1.0   stale (v2.2.0 available)
```

### 4. Remove Ref

```bash
sow refs remove team-go --prune-cache

# Removes symlink and optionally cache
```

## Storage Architecture

### Directory Structure

```
# OCI image contents
ref-package.tar.gz
├── README.md
├── standards/
│   ├── golang.md
│   ├── testing.md
│   └── reviews.md
├── examples/
│   └── error-handling/
│       └── main.go
└── .sow-ref.yaml          # Metadata manifest

# Local cache after pull
~/.cache/sow/refs/
├── go-standards-abc123/   # Content extracted from OCI
│   ├── README.md
│   ├── standards/
│   └── .sow-ref.yaml
└── api-patterns-def456/

# Workspace (committed to project)
.sow/refs/
├── team-go -> ~/.cache/sow/refs/go-standards-abc123
└── api-guide -> ~/.cache/sow/refs/api-patterns-def456

# Index (committed to project)
.sow/refs/index.json
{
  "version": "1.0.0",
  "refs": [
    {
      "id": "team-go",
      "source": "ghcr.io/myorg/go-standards:v1.0.0",
      "digest": "sha256:abc123...",
      "link": "team-go",
      "semantic": "knowledge",
      "tags": ["golang", "conventions"],
      "description": "Team Go coding conventions"
    }
  ]
}
```

### Caching Strategy

**Cache key**: `{ref-id}-{short-digest}`

**Benefits**:
- Multiple versions can coexist
- Rollback without re-download
- Shared cache across projects

**Cache eviction**:
```bash
# Prune unused refs
sow refs prune

# Clear all cache
sow refs prune --all
```

## Registry Recommendations

### GitHub Container Registry (ghcr.io)

**Pros**:
- Free for public repos
- 500MB per package (generous)
- Integrated with GitHub Actions
- Team-based access control
- Familiar authentication (GitHub tokens)

**Cons**:
- Requires GitHub account
- Public packages are, well, public

**Setup**:
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
sow refs publish . ghcr.io/myorg/ref:v1.0.0
```

### Docker Hub

**Pros**:
- Well-known, mature
- Free tier available

**Cons**:
- Rate limits on free tier (100 pulls / 6 hours)
- Smaller free storage

**Use case**: Public, popular refs with broad distribution

### Self-Hosted (Harbor, Nexus)

**Pros**:
- Complete control
- Airgapped environments
- No rate limits
- Private by default

**Cons**:
- Infrastructure overhead
- Maintenance burden

**Use case**: Enterprise with strict security requirements

### Recommendations

**Open source refs**: ghcr.io (free, generous limits)

**Private team refs**: ghcr.io with GitHub org

**Enterprise**: Self-hosted Harbor/Nexus

## Migration from Git Refs

### Current Git-Based Ref

```json
{
  "id": "go-standards",
  "source": "git+https://github.com/myorg/team-docs",
  "semantic": "knowledge",
  "link": "go-standards",
  "config": {
    "branch": "main",
    "path": "docs/"
  }
}
```

### OCI-Based Ref

```json
{
  "id": "go-standards",
  "source": "ghcr.io/myorg/go-standards:v1.0.0",
  "digest": "sha256:abc123...",
  "semantic": "knowledge",
  "link": "go-standards",
  "config": {}
}
```

### Migration Tool

```bash
sow refs migrate git-to-oci go-standards

# Prompts:
# - OCI registry URL (default: ghcr.io/myorg/go-standards)
# - Version tag (default: v1.0.0)
# - Preserve git ref? (default: no)

# Actions:
# 1. Clone git repo
# 2. Generate .sow-ref.yaml from existing metadata
# 3. Publish to OCI registry
# 4. Update .sow/refs/index.json
# 5. Remove git cache (optional)
```

### Transition Period

**Support both formats** during migration:
- Git refs: Continue working as-is
- OCI refs: New format, gradual adoption
- `sow refs add` detects format from URL scheme
  - `git+https://...` → git ref
  - `ghcr.io/...` → OCI ref

## Security Considerations

### Publisher Side

**Content safety**:
- Don't include secrets (`.env`, credentials)
- Use `packaging.exclude` to filter sensitive files
- Review `.sow-ref.yaml` metadata before publishing

**Registry authentication**:
- Use GitHub tokens (not personal passwords)
- Rotate tokens regularly
- Limit token scope to package write only

### Consumer Side

**Registry trust**:
- Verify registry source (trusted org?)
- Check ref metadata before installing
- Use digest pinning for critical refs

**Sandbox execution**:
- Refs are static files (no code execution)
- LLM reads content (not executes)
- Low risk compared to executable packages

**Private registries**:
- Use authenticated registries for sensitive content
- Don't rely on security-through-obscurity
- Audit ref access logs

## Performance Characteristics

### Publishing

**10MB ref**:
- Package: ~2 seconds
- Push to registry: ~5-10 seconds (network-dependent)
- Total: ~10-15 seconds

**100MB ref**:
- Package: ~10 seconds
- Push: ~30-60 seconds
- Total: ~1 minute

### Consumption

**10MB ref**:
- Pull from registry: ~5-10 seconds
- Extract: ~2 seconds
- Index (FTS5): ~1 second
- Index (semantic): ~30 seconds (Ollama embeddings)
- Total: ~8-13 seconds (FTS5), ~38-43 seconds (semantic)

**Cache hit** (already downloaded):
- Symlink creation: <1 second
- Index refresh: ~1 second
- Total: ~1 second

### Storage Overhead

**OCI compression**:
- Markdown/text: ~70-80% compression
- Code: ~60-70% compression
- Mixed: ~70% average

**Example**: 50MB of markdown → ~15MB OCI image

## Open Questions

1. **Registry quotas**: What happens when user exceeds GitHub's 500MB per package limit?
   - Split large refs into multiple packages?
   - Recommend self-hosted registry?

2. **Offline publishing**: How to publish without registry access initially?
   - Export OCI image as tarball
   - Import on machine with registry access
   - `sow refs export` / `sow refs import` commands

3. **Ref versioning semantics**: What constitutes a major vs minor version?
   - Breaking changes to structure (major)
   - Content additions (minor)
   - Typo fixes (patch)
   - Document in publishing guide

4. **Multi-platform refs**: Support platform-specific content (linux/amd64, darwin/arm64)?
   - OCI supports multi-platform manifests
   - Use case unclear (refs are static files, not binaries)
   - Skip for now, add if needed

5. **Ref signing**: Should refs be cryptographically signed?
   - OCI supports image signing (Cosign, Notation)
   - Adds complexity
   - Defer to future enhancement

## Success Criteria

- ✅ Publish 10MB ref to ghcr.io in <15 seconds
- ✅ Install ref from registry in <10 seconds
- ✅ `.sow-ref.yaml` metadata extracted and validated
- ✅ Automatic Docker credential usage
- ✅ Digest pinning for reproducibility
- ✅ Works with GitHub Container Registry
- ✅ Migration tool converts git refs to OCI

## Next Steps

1. **Prototype**: Build `sow refs publish` and `sow refs add` for OCI
2. **Test registries**: Validate with ghcr.io, Docker Hub
3. **Schema validation**: Implement CUE validation for `.sow-ref.yaml`
4. **Documentation**: Publishing guide, registry recommendations
5. **Migration tool**: Assist users in converting existing git refs
