# ADR-003: Use OCI Registries for Refs Distribution

**Status**: Proposed
**Date**: 2025-10-30
**Deciders**: Architecture Team
**Technical Story**: Refs system redesign (improve-refs exploration, October 2025)

## Context

The current refs system uses git repositories for distributing external knowledge and code references. Users add refs by specifying git repository URLs, which sow clones into local cache and symlinks into `.sow/refs/`. While functional, this approach has significant limitations:

**Complex workflow**: Publishing refs requires full git workflow (init, add, commit, push, tag management). Consumers must specify branches, commits, or tags. Authentication requires SSH keys or personal access tokens.

**Storage inefficiency**: Git clones download full repository history even when only current content is needed. For documentation-heavy refs, this means downloading megabytes of git metadata for kilobytes of actual content.

**Versioning challenges**: Git tags and branches provide versioning, but lack immutability guarantees. Tags can be force-pushed. SHAs work but are not human-readable. No standard metadata format exists for describing ref contents.

**No pre-inspection**: Users must download the entire repository to understand its contents. No way to validate structure or metadata before committing to full clone.

The October 2025 exploration of refs improvements (`improve-refs-2025-10`) investigated OCI registries as an alternative distribution mechanism. Key findings validated that OCI provides simpler workflows, better versioning, efficient storage, and the `github.com/jmgilman/go/oci` Go module offers production-ready functionality with estargz support for selective file extraction.

Additionally, the library's estargz capability enables inspecting OCI image contents without downloading (via `ListFiles` reading only the table-of-contents), validating structure before download, and selectively extracting specific paths using glob patterns. This fundamentally changes the consumption experience.

## Decision

Replace git-based refs distribution with OCI registry packages.

Refs will be packaged as OCI images and published to standard OCI registries (GitHub Container Registry, Docker Hub, self-hosted Harbor/Nexus). Each ref image contains content files plus a `.sow-ref.yaml` manifest describing metadata, classifications, and LLM usage hints. The manifest fields are also mapped to standard OCI annotations for queryability.

Use `github.com/jmgilman/go/oci` as the OCI client library, leveraging its built-in security features (path traversal protection, size limits, permission sanitization) and estargz support for selective extraction.

Publishing workflow: `sow refs publish <dir> <registry-url>:<tag>` packages the directory and pushes to registry.

Consumption workflow: `sow refs add <registry-url>:<tag>` pulls the image (optionally inspecting first with `sow refs inspect`), extracts to cache, and symlinks into `.sow/refs/`. Users can specify subpaths via globs to download only specific files: `sow refs add <registry-url>:<tag> --path "docs/**/*.md"`.

## Consequences

### Positive

- **Simpler publishing**: Single command replaces multi-step git workflow. No branches, commits, or force-push concerns.
- **Immutable versioning**: OCI tags are best-practice immutable. Digest pinning (`@sha256:...`) provides cryptographic guarantee of content.
- **Storage efficiency**: OCI compression (70-80% for markdown) and absence of git history metadata significantly reduces bandwidth and storage.
- **Pre-inspection capability**: estargz `ListFiles` allows seeing ref contents before download. Users can verify structure and read `.sow-ref.yaml` metadata without committing to full download.
- **Selective extraction**: Download only needed files via glob patterns. For large refs, users can pull specific subdirectories instead of entire content.
- **Structure validation**: CLI can verify `.sow-ref.yaml` exists and validates before downloading full image, preventing bad ref installations.
- **Native authentication**: Docker credential chain integration means users already logged in to registries have automatic authentication. No separate credential management.
- **Registry ecosystem**: Free hosting (GitHub Container Registry provides 500MB per package), mature tooling, enterprise self-hosting options.
- **Standard metadata**: OCI annotations provide queryable metadata without pulling image. Standard tools (Skopeo, crane) can inspect sow refs.

### Negative

- **Registry requirement**: Publishing requires access to OCI registry. While free options exist (ghcr.io), this is additional infrastructure. Offline publishing requires workaround (export/import).
- **Less familiar**: OCI registries are less universally known than git. Some users will need to learn registry concepts (tags, digests, authentication).
- **estargz format required**: Publishers must use estargz format (not standard tar.gz) to enable selective extraction. The `github.com/jmgilman/go/oci` library handles this, but it's a specific format requirement.
- **Registry rate limits**: Public registries (Docker Hub) have pull rate limits on free tiers. Mitigation: recommend GitHub Container Registry (no rate limits) or self-hosted for heavy usage.
- **Tool dependency**: Using `github.com/jmgilman/go/oci` creates dependency on external library. Mitigation: library is production-ready and actively maintained, with estargz support added per sow requirements.

### Neutral

- **Registry choice flexibility**: Users can choose any OCI-compatible registry (ghcr.io, Docker Hub, Harbor, Nexus). No lock-in to specific provider.
- **Metadata schema introduction**: `.sow-ref.yaml` is new schema requiring documentation. However, this also brings structured metadata that git refs lacked.

## Alternatives Considered

### Option 1: Keep Git-Based Refs (Status Quo)

**Description**: Continue using git repositories for refs distribution, potentially adding metadata schema (`.sow-ref.yaml`) to git repos.

**Pros**:
- No migration required
- Familiar to all developers
- Decentralized by nature

**Cons**:
- Complex publishing workflow persists
- Cannot pre-inspect contents before full clone
- No selective extraction (must clone entire repo)
- Storage inefficiency from git metadata
- Versioning ambiguity (tags are mutable)
- No standard for metadata querying

**Why not chosen**: Does not address core pain points identified in exploration. Publishing complexity and storage inefficiency remain. No path to selective extraction or pre-inspection.

### Option 2: Custom Archive Format + HTTP Server

**Description**: Define custom archive format (tar.gz with manifest), distribute via HTTP server. Publishers upload archives, consumers download via HTTP.

**Pros**:
- Full control over format
- Simple HTTP distribution (no registry concepts)
- Could support selective extraction with custom format

**Cons**:
- Requires custom archive format design and implementation
- No existing ecosystem (must build all tooling)
- Must implement authentication, versioning, metadata querying from scratch
- No existing client libraries
- Hosting infrastructure not standardized (each publisher finds their own solution)
- Reinventing OCI without ecosystem benefits

**Why not chosen**: Significant engineering effort to rebuild what OCI provides. No ecosystem leverage. OCI is proven, standardized, and has mature tooling.

### Option 3: NPM-Like Package Registry

**Description**: Implement npm-style package registry with custom server. Publish using `sow refs publish`, server stores packages and provides API for queries.

**Pros**:
- Familiar to JavaScript developers
- Centralized discovery built-in
- Rich metadata querying

**Cons**:
- Requires operating central registry infrastructure (hosting, uptime, scaling)
- Single point of failure unless self-hosted options provided
- Must implement entire registry protocol
- Package format still needs definition
- Authentication system required
- No existing Go libraries for client/server

**Why not chosen**: Infrastructure burden outweighs benefits. OCI registries already exist with free tiers. Central registry creates governance and maintenance burden. Self-hosting still requires full registry implementation.

### Option 4: GitHub Releases as Distribution

**Description**: Use GitHub releases to attach ref archives. Download via GitHub API.

**Pros**:
- Free hosting
- Familiar git-based workflow
- Versioning via release tags

**Cons**:
- Tied to GitHub (no registry flexibility)
- No selective extraction capability
- Limited metadata (release notes only)
- GitHub API rate limits aggressive (60 requests/hour unauthenticated)
- No OCI ecosystem benefits
- Must download full archive every time

**Why not chosen**: Tied to single provider (GitHub). Cannot support self-hosted or alternative registries. No selective extraction or pre-inspection. OCI provides all benefits without GitHub lock-in.

## Implementation Notes

**Phase 1: OCI integration** (~1 week)
- Integrate `github.com/jmgilman/go/oci` module with estargz support
- Implement `.sow-ref.yaml` schema validation
- Implement `sow refs publish` command (validate manifest, package with estargz, push to registry)
- Implement `sow refs inspect` command (use `ListFiles` to show contents, display metadata without downloading)
- Update `sow refs add` command to detect and handle OCI URLs
- Update `sow refs add` to support `--path` for selective extraction

**Phase 2: Management commands** (~3-5 days)
- Implement `sow refs list`, `sow refs update`, `sow refs remove`
- Implement cache management (`sow refs prune`, `sow refs cache-info`)
- Polish error handling and user feedback

**Registry recommendations**:
- Default: GitHub Container Registry (ghcr.io) - free, 500MB per package, no rate limits
- Alternative: Docker Hub (public refs, rate limits apply)
- Enterprise: Self-hosted Harbor or Nexus

## References

- Design doc: `../designs/oci-refs/oci-refs-design.md` (implementation details)
- Exploration findings: `.sow/knowledge/explorations/improve-refs-2025-10/` (October 2025)
  - Overview: `00-overview.md`
  - OCI research: `oci-refs-distribution.md`
  - Summary: `SUMMARY.md`
- OCI module: [`github.com/jmgilman/go/oci`](https://github.com/jmgilman/go) (estargz support)
- OCI Specification: [Open Container Initiative](https://opencontainers.org/)
- Arc42 updates: Sections 3 (Context), 5 (Building Blocks), 8 (Concepts), 9 (Decisions)
- C4 diagram: Container view update showing OCI registry integration
