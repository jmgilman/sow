# C4 Container Diagram: Refs Subsystem with OCI

**Level**: C4 Level 2 (Container)
**Related ADR**: [ADR-003](../adrs/adr-003-oci-refs-distribution.md)
**Diagram**: [c4-container-refs-oci.mmd](./c4-container-refs-oci.mmd)
**Date**: 2025-10-30

## Overview

C4 Container diagram showing deployment architecture and technology choices for OCI-based refs distribution. Illustrates publishing/consumption flows, estargz selective extraction, and git/OCI coexistence during transition.

---

## External Actors

**Developer**: Publishes refs to registries, installs refs to workspaces via `sow refs` CLI

**AI Agent**: Reads ref contents as context via workspace symlinks (unaware of OCI backing)

---

## External Systems

**OCI Registries** (ghcr.io, Docker Hub, Harbor, Nexus):
- Protocol: HTTPS (OCI Distribution Spec)
- Format: estargz OCI images with annotations
- Auth: Docker credential chain
- Recommended: ghcr.io (free tier, 500MB/package, no rate limits)

**Git Repositories** (Deprecated):
- Legacy refs during transition
- HTTPS/SSH protocols

**Docker Credential Chain**:
- Location: `~/.docker/config.json`
- Transparent authentication (no credential duplication)

---

## Sow CLI Components

**Refs CLI Commands** (Cobra): `publish`, `inspect`, `add`, `update`, `remove`, `list`, `prune`, `migrate`

**Refs Core Logic**: Orchestration, URL format detection (git vs OCI), security enforcement

**OCI Client Wrapper** (`github.com/jmgilman/go/oci`):
- Push/pull estargz images
- ListFiles (TOC-only download)
- Selective extraction with multiple glob patterns (OR logic)
- Security: path traversal, size limits, permission sanitization

**Git Client** (Deprecated): Legacy git operations (go-git)

**Ref Packager**: Create estargz archives, map `.sow-ref.yaml` to OCI annotations

**Schema Validator** (CUE): Validate `cli/schemas/ref_manifest.cue`

**Inspector**: Pre-inspection via `ListFiles` (< 10KB bandwidth)

**Ref Installer**: Pull full or selective (multiple globs, OR logic), atomic extraction, symlink creation

**Cache Manager**: Digest-based deduplication at `~/.cache/sow/refs/`

**Index Manager**: Maintain `.sow/refs/index.json` with file locking - Stores installation metadata and ref metadata (title, description, classifications, tags, authors, license) for fast querying

---

## Local Storage

**CUE Schema**: `cli/schemas/ref_manifest.cue` (embedded in binary)

**Ref Cache**: `~/.cache/sow/refs/<id>-<short-digest>/` (content-addressable)

**Workspace Refs**: `.sow/refs/<link>` → cache (symlinks)

**Refs Index**: `.sow/refs/index.json` - Catalog with installation metadata (id, source, digest, link, selective.globs[]) and ref metadata from `.sow-ref.yaml` (title, description, classifications, tags, authors, license) for fast querying. `selective.globs` is array of glob patterns when installed with multiple `--path` flags.

---

## Data Flows

### Publishing
1. `sow refs publish <dir> <url>:<tag>`
2. Validate `.sow-ref.yaml` (CUE schema)
3. Package as estargz, map annotations
4. Push to registry via Docker credentials

### Inspection (Pre-Download)
1. `sow refs inspect <url>:<tag>`
2. `ListFiles` downloads TOC only (~5KB)
3. Selectively extract `.sow-ref.yaml` (~2KB)
4. Display file list + metadata
5. **Total: < 10KB bandwidth**

### Full Installation
1. `sow refs add <url>:<tag> --link <name>`
2. Pull full image from registry
3. Extract to cache (`<id>-<short-digest>`)
4. Create symlink (`.sow/refs/<link>`)
5. Update index

### Selective Installation
1. `sow refs add <url>:<tag> --path "docs/**/*.md" --path "examples/*.go"`
2. Pull only files matching any of the glob patterns via estargz
3. **90%+ bandwidth savings** for small subsets
4. Always includes `.sow-ref.yaml`

### Legacy Git (Transition)
1. `sow refs add git+https://...`
2. Refs Core routes to Git Client
3. Clone, extract to cache, symlink

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| CLI Framework | Go Cobra |
| OCI Operations | github.com/jmgilman/go/oci (estargz) |
| Schema Validation | CUE (`cli/schemas/`) |
| Protocols | HTTPS (OCI, git) |
| Auth | Docker credential chain |
| Image Format | estargz (seekable tar.gz) |

---

## Performance

| Operation | Bandwidth | Time |
|-----------|-----------|------|
| Publish 10MB | 10MB | ~12-15s |
| Inspect | < 10KB | < 3s |
| Full install 10MB | 10MB | ~8-13s |
| Selective install (10%) | ~1MB | ~2-4s (85% faster) |
| Cache hit | 0KB | < 1s |

---

## Security

- **Path traversal protection**: Rejects malicious paths
- **Size limits**: 100MB/file, 1GB total
- **Permission sanitization**: Strips setuid/setgid
- **Docker credentials**: Transparent auth
- **Digest verification**: Tamper detection

---

## References

- **ADR-003**: [Use OCI Registries](../adrs/adr-003-oci-refs-distribution.md)
- **Design**: [OCI Refs Design](../designs/oci-refs-design.md)
- **Building Blocks**: [Arc42 Section 5](./arc42-05-building-blocks-refs.md)
- **Concepts**: [Arc42 Section 8](./arc42-08-concepts-oci-refs.md)
