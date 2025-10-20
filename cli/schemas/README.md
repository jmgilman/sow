# sow CLI Schemas

This package contains CUE schema definitions and auto-generated Go types for all sow state and index files.

## Contents

### CUE Schemas (Source of Truth)

- **project_state.cue** - Project state schema (`.sow/project/state.yaml`)
- **task_state.cue** - Task state schema (`.sow/project/phases/implementation/tasks/<id>/state.yaml`)
- **refs_committed.cue** - Committed refs index (`.sow/refs/index.json`)
- **refs_cache.cue** - Cache index (`~/.cache/sow/index.json`)
- **refs_local.cue** - Local refs index (`.sow/refs/index.local.json`)

### Generated Go Types

- **cue_types_gen.go** - Auto-generated from CUE schemas (426 lines)

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

### Project State
- `ProjectState` - Root type for project state
- `DiscoveryPhase`, `DesignPhase`, `ImplementationPhase`, `ReviewPhase`, `FinalizePhase`
- `Phase` (base type), `Artifact`, `Task`, `ReviewReport`

### Task State
- `TaskState` - Root type for task state
- `Feedback`

### Refs Indexes
- `RefsCommittedIndex` + `RemoteRef`, `RefPath`
- `RefsCacheIndex` + `CachedRepo`, `CacheUsage`
- `RefsLocalIndex` + `LocalRef`

## Usage Example

```go
import "github.com/jmgilman/sow/cli/schemas"

// Create a project state
state := schemas.ProjectState{
    Project: struct {
        Name        string    `json:"name"`
        Branch      string    `json:"branch"`
        Description string    `json:"description"`
        Created_at  time.Time `json:"created_at"`
        Updated_at  time.Time `json:"updated_at"`
    }{
        Name:        "my-feature",
        Branch:      "feat/my-feature",
        Description: "Implement new feature",
        Created_at:  time.Now(),
        Updated_at:  time.Now(),
    },
}
```

## Validation

Use the generated types with the CUE library for validation:

```go
import (
    "github.com/jmgilman/go/cue"
    "github.com/jmgilman/sow/cli/schemas"
)

// 1. Load CUE schema
loader := cue.NewLoader(schemaFS)
schema, _ := loader.LoadFile(ctx, "schemas/project_state.cue")

// 2. Load data file
dataLoader := cue.NewLoader(dataFS)
data, _ := dataLoader.LoadFile(ctx, ".sow/project/state.yaml")

// 3. Validate against schema
err := cue.Validate(ctx, schema, data)

// 4. Decode into Go type
var state schemas.ProjectState
cue.Decode(ctx, data, &state)
```

## Notes on Nullable Fields

CUE's `*null | T` pattern is represented as `any` in Go with a comment:

```go
Started_at any/* CUE disjunction: (null|string) */ `json:"started_at"`
```

This is because Go's type system can't perfectly represent CUE's union types. You'll need to handle null checks at runtime.

## Schema Constraints

The CUE schemas enforce:
- **Kebab-case naming**: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- **Gap-numbered IDs**: `^[0-9]{3,}$` (010, 020, 030...)
- **Semantic versioning**: `^[0-9]+\.[0-9]+\.[0-9]+$`
- **SHA hashes**: `^[a-f0-9]{40}$`
- **File protocol**: `^file:///. +` for local refs
- **Status enums**: "pending" | "in_progress" | "completed" | etc.
- **Timestamp validation**: ISO 8601 via `time.Time`

## Testing

Run tests to verify generated types:

```bash
go test ./cli/schemas
```

See `example_test.go` for usage patterns.
