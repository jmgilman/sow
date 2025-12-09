# Migrate Project Schemas to libs/schemas/project

## Context

This task is part of the `libs/schemas` module migration project. After the root `libs/schemas` module is created (Task 010), this task migrates the project-specific schemas to a `project/` subpackage.

The project schemas define the core data model for all sow projects:
- `ProjectState` - Complete project state with phases and statechart
- `PhaseState` - State of a single phase within a project
- `TaskState` - Discrete unit of work within a phase
- `ArtifactState` - File or document produced/consumed by phases or tasks
- `StatechartState` - Current state machine position

These schemas are the most heavily used in the codebase (~56 consumer files).

## Requirements

### 1. Create Project Subpackage Structure

Create the following in `libs/schemas/project/`:

```
libs/schemas/project/
├── project.cue          # ProjectState, StatechartState
├── phase.cue            # PhaseState
├── task.cue             # TaskState
├── artifact.cue         # ArtifactState
└── test_helper.cue      # Helper for CUE tests
```

### 2. Copy CUE Schema Files

Copy from `cli/schemas/project/` to `libs/schemas/project/`:
- `project.cue`
- `phase.cue`
- `task.cue`
- `artifact.cue`
- `test_helper.cue`

**Important**: Do NOT modify the CUE files' content. Only copy them.

### 3. Update Package Declaration

The CUE files in `cli/schemas/project/` declare `package project`. This should remain the same since the subpackage name is `project`.

### 4. Generate Go Types

Run CUE code generation from the `libs/schemas` root:

```bash
cd libs/schemas
cue exp gengotypes ./...
```

This will generate:
- `libs/schemas/cue_types_gen.go` (root types)
- `libs/schemas/project/cue_types_gen.go` (project types)

### 5. Copy Schema Tests

Copy the test file from `cli/schemas/project/schemas_test.go` to `libs/schemas/project/schemas_test.go`.

**Update the test file**:
- No import path changes needed (tests are in same package)
- Tests validate CUE schemas against Go data structures

### 6. Verify Generation and Tests

```bash
cd libs/schemas
go build ./...
go test ./...
```

## Acceptance Criteria

1. [ ] `libs/schemas/project/` directory exists with all CUE files
2. [ ] All 5 CUE files are copied and unmodified from source:
   - `project.cue`
   - `phase.cue`
   - `task.cue`
   - `artifact.cue`
   - `test_helper.cue`
3. [ ] `libs/schemas/project/cue_types_gen.go` is generated with types:
   - `ProjectState`
   - `StatechartState`
   - `PhaseState`
   - `TaskState`
   - `ArtifactState`
4. [ ] `libs/schemas/project/schemas_test.go` exists and passes
5. [ ] `go build ./...` succeeds from `libs/schemas`
6. [ ] `go test ./...` succeeds from `libs/schemas`

### Test Requirements

The copied tests should verify:
- Valid project state creation
- Project name pattern validation (`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
- Project type pattern validation (`^[a-z0-9_]+$`)
- Task ID pattern validation (`^[0-9]{3}$`)
- Task status enum validation
- Required field validation
- Optional field handling

All existing tests from `cli/schemas/project/schemas_test.go` should pass without modification.

## Technical Details

### Package Import Path
After migration, consumers will import:
```go
import "github.com/jmgilman/sow/libs/schemas/project"
```

### Generated Type Structure
The generated types should match the existing types in `cli/schemas/project/cue_types_gen.go`:

```go
// ProjectState represents the complete state of a project.
type ProjectState struct {
    Name           string                  `json:"name"`
    Type           string                  `json:"type"`
    Branch         string                  `json:"branch"`
    Description    string                  `json:"description,omitempty"`
    Created_at     time.Time               `json:"created_at"`
    Updated_at     time.Time               `json:"updated_at"`
    Phases         map[string]PhaseState   `json:"phases"`
    Statechart     StatechartState         `json:"statechart"`
    Agent_sessions map[string]string       `json:"agent_sessions,omitempty"`
}
```

### Test Helper CUE File
The `test_helper.cue` file ensures all schema definitions are included during CUE validation tests. It contains references to ensure the CUE tool loads all schemas.

## Relevant Inputs

- `libs/schemas/` - Parent module created in Task 010
- `cli/schemas/project/project.cue` - ProjectState and StatechartState definitions
- `cli/schemas/project/phase.cue` - PhaseState definition
- `cli/schemas/project/task.cue` - TaskState definition
- `cli/schemas/project/artifact.cue` - ArtifactState definition
- `cli/schemas/project/test_helper.cue` - Test helper definitions
- `cli/schemas/project/cue_types_gen.go` - Reference for expected generated types
- `cli/schemas/project/schemas_test.go` - Test file to copy

## Examples

### Expected Directory Structure After Task
```
libs/schemas/
├── go.mod
├── go.sum
├── embed.go
├── cue.mod/module.cue
├── config.cue
├── user_config.cue
├── refs_cache.cue
├── refs_committed.cue
├── refs_local.cue
├── knowledge_index.cue
├── cue_types_gen.go
└── project/
    ├── project.cue
    ├── phase.cue
    ├── task.cue
    ├── artifact.cue
    ├── test_helper.cue
    ├── cue_types_gen.go
    └── schemas_test.go
```

### Test Execution
```bash
cd libs/schemas
go test ./... -v

# Expected output:
# === RUN   TestValidProjectState
# --- PASS: TestValidProjectState (0.05s)
# === RUN   TestValidProjectState_WithDescription
# --- PASS: TestValidProjectState_WithDescription (0.03s)
# ... (53 tests total)
```

## Dependencies

- Task 010: Create libs/schemas Go module structure (must be completed first)

## Constraints

- Do NOT modify the content of CUE schema files (exact copy)
- Do NOT modify the test file logic (exact copy, with any necessary import updates)
- Package name in CUE files MUST remain `project`
- Generated Go types must match existing types exactly (field names, types, json tags)
