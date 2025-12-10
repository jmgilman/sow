# Task 040: Migrate State Wrapper Types

## Context

This task is part of the `libs/project` module consolidation effort. It migrates the state wrapper types (Project, Phase, Task, Artifact) and collection types from `cli/internal/sdks/project/state/` to `libs/project/state/`.

The wrapper types serve several purposes:
1. Embed CUE-generated types (`project.ProjectState`, etc.) for data storage
2. Add runtime fields (Backend reference, ProjectTypeConfig, Machine)
3. Provide typed collection access (PhaseCollection, TaskCollection, ArtifactCollection)
4. Support state machine integration

This migration also updates the Project type to use the new Backend interface instead of `sow.Context`.

## Requirements

### 1. Migrate Project Type (state/project.go)

Migrate and update the Project type:

**Current structure (cli/internal/sdks/project/state/project.go):**
```go
type Project struct {
    project.ProjectState         // Embedded CUE type
    ctx                  *sow.Context
    config               ProjectTypeConfig
    machine              *stateMachine.Machine
}
```

**New structure (libs/project/state/project.go):**
```go
// Project wraps the CUE-generated ProjectState with runtime behavior.
// It provides methods for state machine integration and storage operations.
type Project struct {
    project.ProjectState                              // Embedded CUE type
    backend              Backend                      // Storage backend
    config               *ProjectTypeConfig           // Project type configuration (from parent package)
    machine              *stateMachine.Machine        // State machine instance
}

// NewProject creates a new Project wrapper around a ProjectState.
// The config and machine fields are set during Load() or Create().
func NewProject(state project.ProjectState, backend Backend) *Project

// Config returns the project type configuration.
func (p *Project) Config() *ProjectTypeConfig

// SetConfig sets the project type configuration.
// This is called during Load/Create to attach the config from the registry.
func (p *Project) SetConfig(config *ProjectTypeConfig)

// Machine returns the state machine for this project.
func (p *Project) Machine() *stateMachine.Machine

// SetMachine sets the state machine for this project.
// This is called during Load/Create to attach the built machine.
func (p *Project) SetMachine(machine *stateMachine.Machine)

// Backend returns the storage backend for this project.
func (p *Project) Backend() Backend
```

### 2. Migrate Phase Type (state/phase.go)

Migrate the Phase wrapper:

```go
// Phase wraps the CUE-generated PhaseState.
// This is a pure data wrapper with no additional runtime fields.
type Phase struct {
    project.PhaseState
}
```

Also migrate phase helper functions:
- `IncrementPhaseIteration(p *Project, phaseName string) error`
- `MarkPhaseFailed(p *Project, phaseName string) error`
- `MarkPhaseInProgress(p *Project, phaseName string) error`
- `MarkPhaseCompleted(p *Project, phaseName string) error`
- `AddPhaseInputFromOutput(p *Project, sourcePhaseName, targetPhaseName, artifactType string, filter func(*project.ArtifactState) bool) error`

### 3. Migrate Task Type (state/task.go)

Migrate the Task wrapper:

```go
// Task wraps the CUE-generated TaskState.
// This is a pure data wrapper with no additional runtime fields.
type Task struct {
    project.TaskState
}
```

### 4. Migrate Artifact Type (state/artifact.go)

Migrate the Artifact wrapper:

```go
// Artifact wraps the CUE-generated ArtifactState.
// This is a pure data wrapper with no additional runtime fields.
type Artifact struct {
    project.ArtifactState
}
```

### 5. Migrate Collection Types (state/collections.go)

Migrate the collection types with their methods:

```go
// PhaseCollection provides map-based access to phases by name.
type PhaseCollection map[string]*Phase

func (pc PhaseCollection) Get(name string) (*Phase, error)

// ArtifactCollection provides slice-based access to artifacts.
type ArtifactCollection []Artifact

func (ac ArtifactCollection) Get(index int) (*Artifact, error)
func (ac *ArtifactCollection) Add(artifact Artifact) error
func (ac *ArtifactCollection) Remove(index int) error

// TaskCollection provides slice-based access to tasks.
type TaskCollection []Task

func (tc TaskCollection) Get(id string) (*Task, error)
func (tc *TaskCollection) Add(task Task) error
func (tc *TaskCollection) Remove(id string) error
```

### 6. Migrate Convert Functions (state/convert.go)

Migrate conversion functions between CUE types and wrapper types:

```go
func convertArtifacts(stateArtifacts []project.ArtifactState) ArtifactCollection
func convertArtifactsToState(artifacts ArtifactCollection) []project.ArtifactState
func convertTasks(stateTasks []project.TaskState) TaskCollection
func convertTasksToState(tasks TaskCollection) []project.TaskState
func convertPhases(statePhases map[string]project.PhaseState) PhaseCollection
func convertPhasesToState(phases PhaseCollection) map[string]project.PhaseState
```

### 7. Handle Forward Reference to ProjectTypeConfig

The Project type needs a reference to `ProjectTypeConfig` which will be defined in the parent `project` package (Task 060). Use a forward declaration pattern:

**Option A: Define interface in state package**
```go
// In state/types.go or state/config.go
type ProjectTypeConfig interface {
    Name() string
    // Other methods needed by Project
}
```

**Option B: Accept interface{} and type assert at runtime**
```go
type Project struct {
    ...
    config interface{}  // *project.ProjectTypeConfig
}
```

**Recommended: Option A** - Define a minimal interface that ProjectTypeConfig must implement. This provides type safety while avoiding import cycles.

## Acceptance Criteria

1. [ ] `state/project.go` defines Project type with Backend field (not sow.Context)
2. [ ] `state/phase.go` defines Phase type and helper functions
3. [ ] `state/task.go` defines Task type
4. [ ] `state/artifact.go` defines Artifact type
5. [ ] `state/collections.go` defines all collection types with methods
6. [ ] `state/convert.go` defines all conversion functions
7. [ ] Project type has getters/setters for config and machine fields
8. [ ] All types are properly documented with doc comments
9. [ ] Code compiles without import cycles
10. [ ] `golangci-lint run ./...` passes with no issues
11. [ ] `go test -race ./...` passes with no failures
12. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
13. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write tests following TESTING.md patterns:

**state/project_test.go:**
- NewProject creates wrapper correctly
- Config/SetConfig getter/setter work
- Machine/SetMachine getter/setter work
- Backend returns the provided backend

**state/phase_test.go:**
- IncrementPhaseIteration increments correctly
- MarkPhaseFailed sets status and timestamp
- MarkPhaseInProgress sets status only if pending
- MarkPhaseCompleted sets status and timestamp
- AddPhaseInputFromOutput copies artifacts correctly

**state/collections_test.go:**
- PhaseCollection.Get returns phase or error
- ArtifactCollection.Get returns artifact or error
- ArtifactCollection.Add appends artifact
- ArtifactCollection.Remove removes at index
- TaskCollection.Get finds by ID
- TaskCollection.Add appends task
- TaskCollection.Remove removes by ID

**state/convert_test.go:**
- Verify round-trip conversion preserves all fields
- Test with empty collections
- Test with populated collections

## Technical Details

### Import Dependencies

```go
import (
    "fmt"
    "time"

    "github.com/jmgilman/sow/libs/schemas/project"
    stateMachine "github.com/qmuntal/stateless"
)
```

### Key Changes from Original

1. **Backend instead of Context**: The Project type now holds a `Backend` interface instead of `*sow.Context`. This decouples the state package from CLI-specific types.

2. **No direct Save method on Project**: In the consolidated design, Save is a standalone function that uses the backend. The Project holds the backend reference so Save can be implemented cleanly.

3. **ProjectTypeConfig interface**: Rather than importing the concrete type (which would create a cycle), define a minimal interface.

### File Organization

Following STYLE.md, organize each file as:
1. Package comment
2. Imports
3. Type declarations
4. Constructors
5. Public methods (alphabetized)
6. Private methods (alphabetized)
7. Package-level functions

## Relevant Inputs

- `cli/internal/sdks/project/state/project.go` - Current Project type (lines 1-39)
- `cli/internal/sdks/project/state/phase.go` - Phase type and helpers (lines 1-153)
- `cli/internal/sdks/project/state/task.go` - Task type
- `cli/internal/sdks/project/state/artifact.go` - Artifact type
- `cli/internal/sdks/project/state/collections.go` - Collection types
- `cli/internal/sdks/project/state/convert.go` - Conversion functions
- `.sow/knowledge/designs/sdk-consolidation-design.md` - Design specification
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Project Usage Example

```go
// Creating a project with backend
backend := state.NewMemoryBackend()
projectState := project.ProjectState{
    Name:   "test-project",
    Type:   "standard",
    Branch: "feat/test",
    Phases: map[string]project.PhaseState{
        "planning": {Status: "pending"},
    },
}

proj := state.NewProject(projectState, backend)

// Setting config (done by loader)
proj.SetConfig(config)

// Setting machine (done by loader)
machine := config.BuildMachine(proj, initialState)
proj.SetMachine(machine)

// Accessing embedded state
fmt.Println(proj.Name)  // "test-project" - from embedded ProjectState
```

### Phase Helper Usage

```go
// Mark phase in progress (only if pending)
err := state.MarkPhaseInProgress(proj, "implementation")

// Increment iteration on rework
err := state.IncrementPhaseIteration(proj, "implementation")

// Mark phase completed
err := state.MarkPhaseCompleted(proj, "planning")
```

### Collection Usage

```go
// Get phase by name
phase, err := proj.Phases.Get("planning")

// Add task to phase
task := state.Task{TaskState: project.TaskState{Id: "010", Name: "Design API"}}
phase.Tasks.Add(task)

// Get task by ID
t, err := phase.Tasks.Get("010")
```

## Dependencies

- Task 010: Module foundation with core types
- Task 020: Backend interface (for Project.backend field)
- Task 030: MemoryBackend (for testing)

## Constraints

- Do NOT import anything from `cli/internal/` - this module must be standalone
- Do NOT include Save/Load logic here - that's Task 070
- Do NOT include validation logic here - that's Task 080
- Use the Backend interface, not concrete implementations
- Minimize the ProjectTypeConfig interface to avoid leaking implementation details
