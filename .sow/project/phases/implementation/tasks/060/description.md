# Task 060: Create Metadata Schemas

## Context

This task creates CUE metadata schemas for phase validation. Each phase can have custom metadata fields that are validated using embedded CUE schemas. The schemas ensure metadata structure consistency and type safety.

The exploration project type has minimal metadata requirements:
- **Exploration phase**: No required metadata (research is flexible)
- **Finalization phase**: Optional fields for PR URL and deletion flag

CUE schemas are embedded as strings and used by the SDK for runtime validation.

## Requirements

### Create CUE Directory and Schemas

1. **Create directory**:
   ```
   cli/internal/projects/exploration/cue/
   ```

2. **Create exploration metadata schema**:

   File: `cli/internal/projects/exploration/cue/exploration_metadata.cue`

   ```cue
   package exploration

   // Metadata schema for exploration phase
   {
       // No required metadata for exploration phase
       // Optional metadata can be added as needed
   }
   ```

3. **Create finalization metadata schema**:

   File: `cli/internal/projects/exploration/cue/finalization_metadata.cue`

   ```cue
   package exploration

   // Metadata schema for finalization phase
   {
       // pr_url: Optional URL of created pull request
       pr_url?: string

       // project_deleted: Flag indicating .sow/project/ has been deleted
       project_deleted?: bool
   }
   ```

### Create Metadata Embeddings File

Create `cli/internal/projects/exploration/metadata.go`:

```go
package exploration

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/exploration_metadata.cue
var explorationMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
```

### Remove Placeholder Variables

In `cli/internal/projects/exploration/exploration.go`, remove the placeholder metadata schema variables that were added in Task 030. The embedded variables from `metadata.go` will be used instead.

## Acceptance Criteria

- [ ] Directory `cli/internal/projects/exploration/cue/` exists
- [ ] File `exploration_metadata.cue` created with minimal schema
- [ ] File `finalization_metadata.cue` created with pr_url and project_deleted fields
- [ ] Both fields in finalization schema are optional (use `?`)
- [ ] File `metadata.go` created with embed directives
- [ ] Both schemas embedded as string variables
- [ ] Placeholder variables removed from `exploration.go`
- [ ] Package declaration is "exploration" in all CUE files
- [ ] CUE syntax is valid
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### CUE Embedding

Go 1.16+ supports embedding files via the `embed` package:

```go
//go:embed cue/schema.cue
var schemaVariable string
```

This reads the file at compile time and stores it as a string constant. The SDK uses these strings for runtime validation.

### CUE Schema Syntax

CUE uses structural typing with optional fields:

```cue
{
    required_field: string           // Must exist
    optional_field?: string          // May be omitted
    typed_field: int & >0            // Type with constraint
}
```

For exploration, schemas are intentionally minimal to avoid over-constraining the workflow.

### Schema Validation

The SDK validates metadata when:
- Phases are created/updated
- `sow project validate` is run
- State transitions occur (if configured)

Validation errors include:
- Field type mismatches
- Missing required fields
- Constraint violations

### Metadata Usage

**Finalization metadata fields**:
- `pr_url`: Stored when PR is created, used in prompts/logs
- `project_deleted`: Set to true by cleanup actions, guards final transition

These are used by guards (Task 040) and prompts (Task 070).

### Package Declaration

CUE files must declare their package. For embedded schemas, use the Go package name:

```cue
package exploration
```

This ensures proper CUE module structure.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/metadata.go` - Reference metadata embeddings
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/cue/implementation_metadata.cue` - Reference CUE schema
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/cue/finalize_metadata.cue` - Reference finalize schema
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/state/validate.go` - Validation logic
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 904-947)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Metadata Embeddings (Reference)

From `cli/internal/projects/standard/metadata.go`:

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

### Standard Implementation Metadata Schema (Reference)

From `cli/internal/projects/standard/cue/implementation_metadata.cue`:

```cue
package standard

// Metadata schema for implementation phase
{
    // No metadata currently required for implementation phase
    // Task list approval tracked via output artifacts
    // Task descriptions tracked via task input artifacts
}
```

### Standard Finalize Metadata Schema (Reference)

From `cli/internal/projects/standard/cue/finalize_metadata.cue`:

```cue
package standard

// Metadata schema for finalize phase
{
    // Whether project directory has been deleted
    project_deleted?: bool

    // URL of created pull request
    pr_url?: string

    // Documentation updates made during finalization
    documentation_updates?: [...string]
}
```

### Using Metadata in Guards

Pattern from standard project:

```go
func projectDeleted(p *state.Project) bool {
    phase, exists := p.Phases["finalize"]
    if !exists {
        return false
    }

    if phase.Metadata == nil {
        return false
    }

    val, ok := phase.Metadata["project_deleted"]
    if !ok {
        return false
    }

    boolVal, ok := val.(bool)
    return ok && boolVal
}
```

## Dependencies

- Task 010 (Package structure) - Provides package directory
- Task 030 (Phase configuration) - References these schema variables
- Will be used by SDK validation system at runtime
- Guards in Task 040 may reference metadata fields

## Constraints

- CUE files must use `.cue` extension
- Package name must match Go package ("exploration")
- Embed paths must be relative to Go file location
- Optional fields must use `?` syntax
- Empty schemas are valid (no required fields)
- Cannot use complex CUE features (keep schemas simple)
- Field names must match exactly what code will use (snake_case)
