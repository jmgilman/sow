# Arc42 Section 9: Architecture Decisions - Refs Distribution

**Document**: Architecture Decisions Index for Refs Subsystem
**Date**: 2025-10-30

## Decision Log

| ID | Title | Status | Date | Affects Sections |
|----|-------|--------|------|------------------|
| [ADR-003](../../adrs/adr-003-oci-refs-distribution.md) | Use OCI Registries for Refs Distribution | Proposed | 2025-10-30 | 3 (Context), 5 (Building Blocks), 8 (Concepts) |

---

## Key Decisions Summary

### ADR-003: Use OCI Registries for Refs Distribution

**Decision**: Replace git-based refs with OCI registry packages (estargz format, `.sow-ref.yaml` metadata, published to ghcr.io/Docker Hub/Harbor)

**Context**: Git-based refs have complex publishing, storage inefficiency, no pre-inspection, no selective extraction, versioning ambiguity. October 2025 exploration validated OCI with `github.com/jmgilman/go/oci` library providing estargz support.

**Alternatives Rejected**:
1. Keep git - doesn't solve pain points
2. Custom archive + HTTP - reinvents OCI
3. NPM-like registry - infrastructure burden
4. GitHub Releases - provider lock-in, no selective extraction

**Key Benefits**:
- Simpler publishing (single command)
- Pre-inspection via `ListFiles` (< 10KB bandwidth)
- Selective extraction with globs (90%+ savings)
- Structure validation before download
- Immutable versioning (digest pinning)
- 70-80% compression

**Key Trade-offs**:
- Registry requirement (mitigated: free ghcr.io)
- Less familiar than git
- estargz format required
- Migration effort (tooling provided)

**Impact on Architecture**:

**Section 5 (Building Blocks)**:
- Adds OCI Client Wrapper (Packager, Validator, Inspector, Installer)
- Updates Refs Core for git/OCI routing
- Maintains Git Client as legacy

**Section 8 (Concepts)**:
- OCI Distribution Pattern
- Selective Extraction with estargz
- Metadata Schema and Validation (CUE)
- Digest-Based Caching
- Coexistence and Migration Pattern

**Implementation**: 3-4 weeks across 4 phases

**Related Documents**:
- **Design**: [OCI Refs Design](../designs/oci-refs-design.md)
- **Exploration**: `.sow/knowledge/explorations/improve-refs-2025-10/`
- **Building Blocks**: [Section 5](./arc42-05-building-blocks-refs.md)
- **Concepts**: [Section 8](./arc42-08-concepts-oci-refs.md)

---

## Future Decisions

**Ref Search Architecture**: FTS5 keyword + optional Ollama semantic (after OCI complete)

**Marketplace System**: Git-based discovery (Homebrew-style taps)

---

## References

- **ADRs**: `.sow/knowledge/adrs/`
- **Explorations**: `.sow/knowledge/explorations/`
- **Design Docs**: `.sow/knowledge/designs/`
