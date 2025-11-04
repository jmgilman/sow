# Task 030: Create Metadata Schemas

# Task 030: Create Metadata Schemas

## Overview

Define minimal CUE schemas for phase-specific metadata validation and create Go embeddings. These schemas validate custom metadata fields for implementation, review, and finalize phases.

## Context

**Design Reference**: `.sow/knowledge/designs/project-sdk-implementation.md` (lines 564-577) for metadata schema examples

**Why Metadata Schemas**: The SDK uses a two-tier validation approach:
1. **Structure validation** (CUE) - Universal fields like `name`, `type`, `status` (handled by SDK)
2. **Metadata validation** (embedded CUE) - Project-type-specific fields in the `metadata` map (this task)

**Current Usage**: Review existing guards in `cli/internal/project/standard/guards.go` to understand which metadata fields are accessed:
- `tasks_approved` (bool) - Implementation phase
- `iteration` (int) - Review phase
- `project_deleted` (bool) - Finalize phase
- `pr_url` (string) - Finalize phase

## Requirements

### CUE Schema Files

Create three CUE schema files in `cli/internal/projects/standard/cue/`:

**1. implementation_metadata.cue**

```cue
package standard

// Metadata schema for implementation phase
{
	// tasks_approved indicates human approval of task breakdown
	tasks_approved?: bool
}
```

**2. review_metadata.cue**

```cue
package standard

// Metadata schema for review phase
{
	// iteration tracks review cycles (starts at 1)
	iteration?: int & >=1
}
```

**3. finalize_metadata.cue**

```cue
package standard

// Metadata schema for finalize phase
{
	// project_deleted indicates project directory removed
	project_deleted?: bool

	// pr_url is the pull request URL if created
	pr_url?: string
}
```

### Go Embeddings File

Create `cli/internal/projects/standard/metadata.go`:

```go
package standard

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/implementation_metadata.cue
var implementationMetadataSchema string

//go:embed cue/review_metadata.cue
var reviewMetadataSchema string

//go:embed cue/finalize_metadata.cue
var finalizeMetadataSchema string
```

## Acceptance Criteria

- [ ] Directory `cli/internal/projects/standard/cue/` exists (created in Task 010)
- [ ] File `cli/internal/projects/standard/cue/implementation_metadata.cue` created
- [ ] File `cli/internal/projects/standard/cue/review_metadata.cue` created
- [ ] File `cli/internal/projects/standard/cue/finalize_metadata.cue` created
- [ ] File `cli/internal/projects/standard/metadata.go` created with embeddings
- [ ] All CUE files use `package standard`
- [ ] All fields are optional (`?` suffix) to support incremental updates
- [ ] CUE schemas compile: `cue vet cli/internal/projects/standard/cue/*.cue`
- [ ] Go file compiles: `go build ./cli/internal/projects/standard/...`
- [ ] Embedded variables are package-level and lowercase (not exported)
- [ ] Old package `cli/internal/project/standard/` untouched

## Validation Commands

```bash
# Verify CUE files exist
ls cli/internal/projects/standard/cue/

# Verify CUE syntax is valid
cue vet cli/internal/projects/standard/cue/*.cue

# Verify Go embeddings compile
go build ./cli/internal/projects/standard/...

# Verify schemas are embedded (check binary)
go build -o /tmp/test-embed ./cli/internal/projects/standard/
strings /tmp/test-embed | grep "tasks_approved"
strings /tmp/test-embed | grep "iteration"
strings /tmp/test-embed | grep "project_deleted"

# Verify old package untouched
git diff cli/internal/project/standard/
```

## Dependencies

- Task 010 (cue/ directory exists)

## Standards

- Keep schemas minimal - only fields currently used by guards and prompts
- All fields optional (support incremental metadata updates)
- Use descriptive comments explaining each field's purpose
- Follow CUE naming conventions (snake_case for fields)
- Use Go `//go:embed` directive for clean embeddings

## Notes

**Why Keep Schemas Minimal**: We only define fields that are actually used. This avoids over-engineering and keeps validation focused. Fields can be added later as needed.

**Validation Timing**: These schemas are validated during `project.Save()` operations in the SDK. See `.sow/knowledge/designs/project-sdk-implementation.md` (lines 1319-1340) for validation implementation details.

**Schema Evolution**: Since all fields are optional, adding new fields later is backwards-compatible. Existing project states will still validate successfully.

**Usage in Task 6**: These embedded schema strings will be passed to `WithMetadataSchema()` when configuring phases in the SDK builder.
