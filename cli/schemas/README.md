# sow CLI Schemas

This package contains CUE schema definitions and auto-generated Go types for all sow state and index files.

## Contents

### CUE Schemas (Source of Truth)

#### Universal Project Schemas (`project/`)
- **project.cue** - ProjectState and StatechartState definitions
- **phase.cue** - PhaseState definition with artifact and task collections
- **artifact.cue** - ArtifactState definition for tracking work products
- **task.cue** - TaskState definition with iteration support

#### Refs Indexes
- **refs_committed.cue** - Committed refs index (`.sow/refs/index.json`)
- **refs_cache.cue** - Cache index (`~/.cache/sow/index.json`)
- **refs_local.cue** - Local refs index (`.sow/refs/index.local.json`)

### Generated Go Types

- **project/cue_types_gen.go** - Auto-generated from project schemas with full GoDoc comments
- **cue_types_gen.go** - Auto-generated from refs schemas

### Supporting Files

- **doc.go** - Package documentation + `go:generate` directive
- **example_test.go** - Usage examples and verification tests

## Regenerating Go Types

When you modify any `.cue` file, regenerate the Go types:

```bash
go generate ./cli/schemas
```

This runs:
```bash
go run cuelang.org/go/cmd/cue@v0.12.0 exp gengotypes ./...
```

## Generated Types

### Universal Project Types (`project.ProjectState`, etc.)
- **ProjectState** - Complete project state with metadata, phases map, and statechart
- **PhaseState** - Universal phase state (replaces concrete phase types like DiscoveryPhase)
- **ArtifactState** - Tracks work products (design docs, task lists, reviews, ADRs)
- **TaskState** - Discrete unit of work with iteration support and artifact I/O
- **StatechartState** - Current state machine position

### Refs Indexes
- `RefsCommittedIndex` + `RemoteRef`, `RefPath`
- `RefsCacheIndex` + `CachedRepo`, `CacheUsage`
- `RefsLocalIndex` + `LocalRef`

## Usage Example

```go
import (
    "time"
    "github.com/jmgilman/sow/cli/schemas/project"
)

// Create a project state with universal types
state := project.ProjectState{
    Name:        "my-feature",
    Type:        "standard",
    Branch:      "feat/my-feature",
    Description: "Implement new feature",
    Created_at:  time.Now(),
    Updated_at:  time.Now(),
    Phases: map[string]project.PhaseState{
        "planning": {
            Status:     "completed",
            Enabled:    true,
            Created_at: time.Now(),
            Inputs:     []project.ArtifactState{},
            Outputs:    []project.ArtifactState{},
            Tasks:      []project.TaskState{},
        },
    },
    Statechart: project.StatechartState{
        Current_state: "ImplementationPlanning",
        Updated_at:    time.Now(),
    },
}
```

## Validation

Use the generated types with the CUE library for validation:

```go
import (
    "cuelang.org/go/cue"
    "cuelang.org/go/cue/cuecontext"
    "github.com/jmgilman/sow/cli/schemas/project"
)

// 1. Create CUE context
ctx := cuecontext.New()

// 2. Load CUE schema (example: project state)
schemaValue := ctx.CompileString(`
    #include "schemas/project/project.cue"
`)

// 3. Load data and unify with schema
dataValue := ctx.CompileString(`...your YAML data...`)
unified := schemaValue.Unify(dataValue)

// 4. Validate
if err := unified.Validate(cue.Concrete(true)); err != nil {
    // validation failed
}

// 5. Decode into Go type
var state project.ProjectState
unified.Decode(&state)
```

## Notes on Optional Fields

Optional fields in CUE (marked with `?`) are represented in Go with `omitempty` tags:

```go
// CUE definition
started_at?: time.Time

// Generated Go type
Started_at time.Time `json:"started_at,omitempty"`
```

Zero values are omitted during JSON marshaling. Use time.IsZero() to check if optional timestamps were set.

## Schema Constraints

### Universal Project Schemas

The project schemas enforce:
- **Project names**: `^[a-z0-9][a-z0-9-]*[a-z0-9]$` (kebab-case, no leading/trailing hyphens)
- **Project types**: `^[a-z0-9_]+$` (lowercase alphanumeric with underscores)
- **Task IDs**: `^[0-9]{3}$` (exactly 3 digits: 001, 010, 042, 999)
- **Task status**: `"pending" | "in_progress" | "completed" | "abandoned"`
- **Phase status**: Defined by project type (e.g., "pending", "in_progress", "completed")
- **Non-empty strings**: All required text fields must be non-empty
- **Iteration numbers**: Must be >= 1
- **Timestamps**: ISO 8601 via `time.Time`

### Refs Schemas

The refs schemas enforce:
- **Semantic versioning**: `^[0-9]+\.[0-9]+\.[0-9]+$`
- **SHA hashes**: `^[a-f0-9]{40}$`
- **File protocol**: `^file:///.+` for local refs

## Testing

Run tests to verify schemas and generated types:

```bash
# Test universal project schemas (53 tests)
go test ./schemas/project

# Test all schemas
go test ./schemas/...
```

The project schema tests verify:
- Valid state creation and modification
- Constraint enforcement (patterns, enums, required fields)
- Optional field handling
- Nested structure validation (phases, tasks, artifacts)

## Universal Data Model

The `project/` schemas define a **universal data model** that all project types share:

### Key Concepts

1. **Single Type System**: All project types (standard, exploration, design, etc.) use the same `ProjectState`, `PhaseState`, `TaskState`, and `ArtifactState` types.

2. **Flexible Metadata**: Common fields (status, timestamps, collections) are strongly typed. Project-type-specific data goes in flexible `metadata` maps.

3. **Phase Map**: Instead of concrete types (`DiscoveryPhase`, `DesignPhase`), there's a single `map[string]PhaseState`. Project types define which phases they have.

4. **Artifact I/O**: Phases and tasks track inputs and outputs as `ArtifactState` collections, making work products explicit.

5. **Iteration Support**: Tasks have iteration numbers for refinement cycles (task → review → revise → review again).

6. **State Machine Integration**: The `statechart` field tracks current state machine position for workflow enforcement.

### Benefits

- **Extensibility**: Add new project types without changing the type system
- **Consistency**: All projects follow the same structure
- **Tooling**: Write tools once, work with all project types
- **Migration**: Easy to restructure phases without breaking changes

See also: `internal/sdks/state/` for the state machine SDK that powers project workflows.
