# Arc42 Section 8: Cross-cutting Concepts - OCI Refs Architecture

**Document**: Cross-cutting Concepts for OCI-Based Refs Distribution
**Related ADR**: [ADR-003](../adrs/adr-003-oci-refs-distribution.md)
**Date**: 2025-10-30

## Overview

Architectural patterns and concepts that apply across the OCI refs subsystem.

---

## 1. OCI Distribution Pattern

**Concept**: Use OCI registries as distribution mechanism, treating refs as immutable packages

**Rationale**: Standards-based, existing infrastructure, mature tooling, immutability, compression

**Application**:
- Package refs as estargz OCI images with `.sow-ref.yaml` + OCI annotations
- Publish via standard OCI push
- Consume via standard OCI pull
- Version with semantic tags + optional digest pinning

---

## 2. Selective Extraction with estargz

**Concept**: Use estargz (seekable tar.gz) to download subsets without retrieving entire images

**Rationale**: Bandwidth efficiency, pre-inspection, improved UX

**Application**:

**Inspection** (`sow refs inspect`):
- Download TOC only (~5-20KB) via `ListFiles`
- Display directory tree, file count
- Selectively extract `.sow-ref.yaml` (~1-5KB)
- Total bandwidth: < 10KB

**Selective install** (`--path "glob1" --path "glob2"`):
- Download files matching any of the glob patterns via estargz (OR logic)
- Always include `.sow-ref.yaml`
- Bandwidth savings: 90%+ for small subsets

**Structure validation**: Inspect validates `.sow-ref.yaml` before suggesting full download

---

## 3. Metadata Schema and Validation

**Concept**: Structured `.sow-ref.yaml` schema validated with CUE

**Rationale**: Structured metadata enables search/querying, classification guides LLM usage, early validation prevents bad refs

**Schema Structure**:
- Required: `schema_version`, `ref.{title,link}`, `content.{description,classifications,tags}`
- Optional: `content.summary`, `provenance.*`, `packaging.exclude`, `hints.*`
- Classifications: `tutorial`, `api-reference`, `guidelines`, `architecture`, `code-examples`, etc.

**CUE Validation**:
- Schema: `cli/schemas/ref_manifest.cue` (embedded)
- Validates formats, enums, constraints
- Leverages existing `schemas.Validator` infrastructure

---

## 4. Digest-Based Caching and Deduplication

**Concept**: Content-addressable storage keyed by SHA256 digest

**Rationale**: Deduplication, version coexistence, integrity, rollback

**Cache Structure**:
```
~/.cache/sow/refs/
├── go-standards-abc1234/  # <id>-<short-digest>
└── api-patterns-def5678/

.sow/refs/
├── team-go -> ~/.cache/sow/refs/go-standards-abc1234
└── apis -> ~/.cache/sow/refs/api-patterns-def5678
```

**Benefits**: Same digest = stored once, fast rollback, integrity verification

---

## 5. Coexistence and Migration Pattern

**Concept**: Support git and OCI refs simultaneously during transition

**URL Detection**:
- `git+https://...` → Git client
- `ghcr.io/...` → OCI client

**Migration**: `sow refs migrate git-to-oci` converts refs, updates index

**Timeline**: 4-phase rollout over 10+ months (parallel → deprecation → removal)

---

## 6. Security Model

**Multi-layer Security**:

1. **Publisher Validation**: CUE schema prevents malformed refs, exclusion patterns
2. **Registry Auth**: Docker credential chain (transparent)
3. **OCI Client Security**: Path traversal protection, size limits (100MB/file, 1GB total), permission sanitization
4. **Digest Verification**: Tamper detection via digest mismatch
5. **Structure Validation**: Pre-installation check validates `.sow-ref.yaml`

---

## 7. Error Handling and Observability

**Fail Fast**: Abort on validation errors, retry network errors (3x), atomic operations

**Error Categories**:
- Validation: Field-level errors, exit code 1
- Network: Auto-retry, exit code 2
- Auth: Show `docker login` command, exit code 3
- Registry: Display registry error, exit code 4

**Observability**:
- Structured JSON logs
- Progress bars for long operations
- Local metrics only (no telemetry)
- Verbose mode for debugging

---

## 8. Versioning and Immutability

**Semantic Version Tags**: `v1.0.0`, `v1.1.0` (immutable best practice)

**Digest Pinning**: `@sha256:...` for cryptographic guarantee

**Update Semantics**: Explicit `sow refs update` (no automatic polling)

**Rollback**: Previous versions remain in cache

---

## Summary

Key principles:
1. Standards over custom (OCI ecosystem)
2. Efficiency over completeness (selective extraction, 90%+ savings)
3. Validation over trust (CUE schema, structure checks)
4. Immutability over mutability (digest-based)
5. Security by default (multi-layer)

## References

- **ADR-003**: [Use OCI Registries](../adrs/adr-003-oci-refs-distribution.md)
- **Design**: [OCI Refs Design](../designs/oci-refs-design.md)
