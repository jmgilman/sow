# Issue #124: Work Unit 002: CUE Schema for .sow-ref.yaml Manifest

**URL**: https://github.com/jmgilman/sow/issues/124
**State**: OPEN

## Description

# Work Unit 002: CUE Schema for .sow-ref.yaml Manifest

**Status**: Specification
**Estimated Effort**: 2-3 days
**Dependencies**: None (greenfield)

---

## Behavioral Goal

**As a** ref publisher,
**I need** a validated schema for the `.sow-ref.yaml` manifest file,
**So that** I can define standardized metadata for my refs, receive clear validation errors before publishing, and ensure consumers can reliably interpret ref contents.

### Success Criteria

1. Publishers creating `.sow-ref.yaml` files receive clear, actionable validation errors for invalid manifests
2. Valid manifests pass schema validation with no false positives
3. Classification types are constrained to a predefined enum (preventing typos like "guidlines")
4. Link names are validated against kebab-case pattern (ensuring valid symlink names)
5. Generated Go types provide type-safe access to manifest fields
6. Validation function is callable from packaging code to validate before OCI push

---

## Existing Code Context

### Explanatory Context

The sow codebase has a mature CUE schema infrastructure in `libs/schemas/`. All CUE schemas are embedded into the binary via `//go:embed` and Go types are auto-generated using `cue exp gengotypes`. The existing `refs_committed.cue` schema defines the current index structure with similar patterns we'll follow (regex validation for IDs, enum constraints for types, optional fields with `?`).

The validation pattern in `libs/project/state/validate.go` demonstrates the CUE-based validation approach: schemas are loaded from embedded FS at init time, values are encoded to CUE, then unified with the schema and validated. This pattern will be reused for `.sow-ref.yaml` validation.

The new schema belongs in `libs/schemas/` (not `cli/schemas/` as mentioned in early design drafts) because:
1. All existing schemas live in `libs/schemas/`
2. The `embed.go` already embeds `*.cue` from that location
3. Discovery analysis (Section 10.1) explicitly recommended this location

### Key Files

| File | Lines | Purpose |
|------|-------|---------|
| `libs/schemas/refs_committed.cue` | 1-68 | Pattern for ref-related schema with regex validation |
| `libs/schemas/refs_cache.cue` | 1-82 | Pattern for enum types and optional fields |
| `libs/schemas/embed.go` | 1-11 | Embeds all `*.cue` files; no changes needed |
| `libs/schemas/cue_types_gen.go` | 1-275 | Generated Go types; regenerated after schema changes |
| `libs/project/state/validate.go` | 46-68 | CUE validation pattern to follow |
| `libs/schemas/project/project.cue` | 1-58 | Pattern for nested struct definitions |

---

## Existing Documentation Context

### Design Document (Primary Reference)

The OCI Refs Design Document (`.sow/knowledge/designs/oci-refs/oci-refs-design.md`, lines 369-436) defines the complete `.sow-ref.yaml` schema:

- **Required fields** are explicitly enumerated: `schema_version`, `ref.title`, `ref.link`, `content.description`, `content.classifications`, `content.tags`
- **Classification types enum** is defined with 11 valid values from `tutorial` to `uncategorized`
- **Validation constraints** specify regex patterns (link format: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`) and semver validation for `schema_version`
- **Optional sections** include provenance (authors, dates, source, license), packaging (exclude globs), hints (suggested_queries, primary_files), and freeform metadata

This schema is NOT just a data format - it's a contract. The OCI annotation mapping (lines 424-437) shows how fields map to OCI annotations during packaging, making schema consistency critical.

### Discovery Analysis

Section 4 of the discovery analysis (`.sow/project/discovery/analysis.md`) confirms:
- CUE is the established schema technology
- Generated Go types via `cue exp gengotypes` is the standard pattern
- Validation should follow the `validateStructure` pattern from state validation
- Schema location should be `libs/schemas/ref_manifest.cue` (Section 10.1 recommendation)

---

## Detailed Requirements

### Schema Definition (ref_manifest.cue)

Create `libs/schemas/ref_manifest.cue` with the following definitions:

```cue
package schemas

import "time"

// RefManifest defines the schema for .sow-ref.yaml
// This manifest is required in all OCI-distributed refs
#RefManifest: {
    // schema_version uses semver format
    schema_version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

    // Core identification
    ref: #RefIdentification

    // Content description
    content: #RefContent

    // Optional sections
    provenance?: #RefProvenance
    packaging?: #RefPackaging
    hints?: #RefHints
    metadata?: {...}  // Freeform
}

#RefIdentification: {
    // Human-readable name (5-100 chars)
    title: string & =~".{5,100}"

    // Symlink name (kebab-case)
    link: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"
}

#RefContent: {
    // Required description (50-200 chars recommended)
    description: string & !=""

    // Optional longer summary (markdown allowed)
    summary?: string

    // At least one classification required
    classifications: [#RefClassification, ...#RefClassification]

    // At least one tag required
    tags: [string, ...string]
}

#RefClassification: {
    type: #ClassificationType
    description?: string
}

#ClassificationType: "tutorial" | "api-reference" | "guidelines" |
    "architecture" | "runbook" | "specification" | "reference" |
    "code-examples" | "code-templates" | "code-library" | "uncategorized"

#RefProvenance: {
    authors?: [...string]
    created?: time.Time
    updated?: time.Time
    source?: string
    license?: string
}

#RefPackaging: {
    exclude?: [...string]  // Glob patterns
}

#RefHints: {
    suggested_queries?: [...string]
    primary_files?: [...string]
}
```

### Validation Constraints

1. **schema_version**: Must be valid semver (e.g., "1.0.0", "0.1.0-beta")
2. **ref.title**: 5-100 characters (informative constraint)
3. **ref.link**: Kebab-case alphanumeric with hyphens (pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
4. **content.classifications**: Array must have at least one element
5. **content.tags**: Array must have at least one element
6. **classification.type**: Must be one of the 11 enum values
7. **provenance.created/updated**: RFC 3339 format (time.Time in CUE)

### Go Type Generation

After creating the CUE schema, run:
```bash
go generate ./libs/schemas/...
```

This will update `cue_types_gen.go` with new types:
- `RefManifest`
- `RefIdentification`
- `RefContent`
- `RefClassification`
- `RefProvenance`
- `RefPackaging`
- `RefHints`

### Validation Function

Create validation logic following the pattern in `libs/project/state/validate.go`. The function should:

1. Load embedded schema using `cuepkg.NewLoader`
2. Encode input struct to CUE value
3. Unify with `#RefManifest` schema
4. Return detailed error messages for validation failures

Recommended location: Either inline in future `libs/refs/` module, or as utility function in `libs/schemas/` if reuse is expected. The packaging work unit (004) will consume this.

---

## Testing Requirements

### Unit Tests

1. **Valid manifest tests**:
   - Minimal valid manifest (only required fields)
   - Full manifest with all optional sections
   - All classification types individually
   - Edge cases: minimum length title (5 chars), maximum length (100 chars)

2. **Invalid manifest tests**:
   - Missing required field: `schema_version`
   - Missing required field: `ref.title`
   - Missing required field: `ref.link`
   - Missing required field: `content.description`
   - Empty `classifications` array
   - Empty `tags` array
   - Invalid classification type (e.g., "guidlines" typo)
   - Invalid link format (spaces, uppercase, special chars)
   - Invalid semver format

3. **Error message quality tests**:
   - Verify error messages include field path
   - Verify error messages are actionable (suggest fix)

### Test File Location

`libs/schemas/ref_manifest_test.go` or wherever validation function is placed.

---

## Implementation Notes

### Embed.go Update

The existing `//go:embed *.cue` directive in `libs/schemas/embed.go` will automatically include the new `ref_manifest.cue` file. No changes needed to embed.go.

### CUE Module Configuration

The `libs/schemas/cue.mod/module.cue` should already be configured. Verify it can import `time` package for timestamp validation.

### Integration Points

This schema will be consumed by:
- **Work Unit 004 (Packaging)**: Validates `.sow-ref.yaml` before OCI push
- **Work Unit 005 (Inspection)**: Validates remote manifest before displaying to user
- **Work Unit 006 (Installation)**: Validates manifest after extraction

These work units depend on this schema being complete.

---

## Out of Scope

- **OCI annotation mapping**: Handled in Work Unit 004 (Packaging)
- **Index schema updates**: Handled in Work Unit 007 (CLI Integration)
- **CLI validation commands**: Handled in Work Unit 007 (CLI Integration)
- **Validation error formatting for CLI output**: Deferred to consuming work units

---

## Implementation Standards

All code produced in this work unit MUST adhere to the following standards:

### Code Quality Standards
- **STYLE.md Compliance**: All Go code must follow the conventions documented in `.standards/STYLE.md`
- **TESTING.md Compliance**: All tests must follow the patterns documented in `.standards/TESTING.md`
- **golangci-lint**: Code must pass `golangci-lint run` with zero errors before completion

### Required Dependencies
- **OCI Operations**: Use `github.com/jmgilman/go/oci` for all OCI registry operations (established in Work Unit 003)
- **Filesystem Abstractions**: Use `github.com/jmgilman/go/fs/core` and `github.com/jmgilman/go/fs/billy` for all file system operations requiring abstraction (enables testability via memory FS)

### Verification Checklist
Before marking this work unit complete, verify:
- [ ] `golangci-lint run ./libs/schemas/...` passes with zero errors
- [ ] All code follows STYLE.md conventions (functional options, error wrapping, etc.)
- [ ] All tests follow TESTING.md patterns (table-driven tests, test helpers, etc.)

---

## Acceptance Criteria

- [ ] `libs/schemas/ref_manifest.cue` exists and compiles without errors
- [ ] Running `go generate ./libs/schemas/...` produces updated `cue_types_gen.go`
- [ ] New Go types are present: `RefManifest`, `RefIdentification`, `RefContent`, `RefClassification`, `RefProvenance`, `RefPackaging`, `RefHints`
- [ ] Unit tests pass for valid manifests (minimal and full)
- [ ] Unit tests pass for invalid manifests (all validation constraints)
- [ ] Classification type enum rejects typos and unlisted values
- [ ] Link format regex rejects uppercase, spaces, and special characters
- [ ] A validation function exists that can be called from packaging code
