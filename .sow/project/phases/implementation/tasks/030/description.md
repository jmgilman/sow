# Task 030: Implement MemoryBackend for Testing

## Context

This task is part of the `libs/project` module consolidation effort. The `MemoryBackend` provides an in-memory implementation of the `Backend` interface for unit testing purposes.

Currently, testing project state operations requires either:
- Real filesystem operations (slow, requires cleanup)
- Complex mocking of the `sow.Context` type

The `MemoryBackend` enables fast, isolated unit tests by storing project state in memory. This follows the pattern established in `libs/exec` which provides test-friendly implementations alongside production code.

## Requirements

### 1. Implement MemoryBackend (state/backend_memory.go)

Create an in-memory backend for testing:

```go
// MemoryBackend implements Backend using in-memory storage.
// This is primarily intended for unit tests and development.
type MemoryBackend struct {
    state *project.ProjectState
    mu    sync.RWMutex
}

// NewMemoryBackend creates an empty in-memory backend.
func NewMemoryBackend() *MemoryBackend

// NewMemoryBackendWithState creates a backend pre-populated with state.
// Useful for testing scenarios that require existing project state.
func NewMemoryBackendWithState(state *project.ProjectState) *MemoryBackend

// Load returns the stored project state.
func (b *MemoryBackend) Load(ctx context.Context) (*project.ProjectState, error)

// Save stores the project state in memory.
func (b *MemoryBackend) Save(ctx context.Context, state *project.ProjectState) error

// Exists returns whether state is currently stored.
func (b *MemoryBackend) Exists(ctx context.Context) (bool, error)

// Delete clears the stored state.
func (b *MemoryBackend) Delete(ctx context.Context) error

// State returns the raw stored state for test assertions.
// This is NOT part of the Backend interface - it's a test helper.
func (b *MemoryBackend) State() *project.ProjectState
```

### 2. Implementation Requirements

**Thread Safety:**
- Use `sync.RWMutex` to allow concurrent read access
- Load and Exists use RLock
- Save and Delete use Lock
- This enables safe parallel test execution

**Load operation:**
- Return `ErrNotFound` if state is nil
- Return a deep copy of the state to prevent test interference
- Deep copy should copy all nested structures (Phases map, Tasks, Artifacts, etc.)

**Save operation:**
- Store a deep copy of the input state
- Never modify the caller's state directly
- Replace any existing state completely

**Exists operation:**
- Return true if state is not nil
- Always succeeds (returns nil error)

**Delete operation:**
- Set state to nil
- Always succeeds even if state was already nil

**State helper method:**
- Returns the raw pointer for test assertions
- Should NOT be used in production code
- Document that this may return nil

### 3. Deep Copy Implementation

Create a helper function for deep copying ProjectState:

```go
// copyProjectState creates a deep copy of a ProjectState.
// This is used to ensure the backend's internal state is isolated
// from caller modifications.
func copyProjectState(src *project.ProjectState) *project.ProjectState
```

The deep copy must handle:
- `Phases` map - create new map with copied PhaseState values
- `PhaseState.Inputs/Outputs` slices - create new slices with copied ArtifactState
- `PhaseState.Tasks` slice - create new slice with copied TaskState
- `TaskState.Inputs/Outputs` slices - create new slices with copied ArtifactState
- `Metadata` maps - create new maps with copied values (shallow copy acceptable for map values since they're `any`)
- `Agent_sessions` map - create new map with copied values

## Acceptance Criteria

1. [ ] `state/backend_memory.go` implements MemoryBackend with all methods
2. [ ] NewMemoryBackend creates empty backend
3. [ ] NewMemoryBackendWithState creates pre-populated backend
4. [ ] All Backend interface methods are implemented correctly
5. [ ] Thread safety is implemented with RWMutex
6. [ ] Deep copy prevents state interference between operations
7. [ ] State() helper method exposes raw state for tests
8. [ ] `golangci-lint run ./...` passes with no issues
9. [ ] `go test -race ./...` passes with no failures (race detector verifies thread safety)
10. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
11. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write comprehensive tests in `state/backend_memory_test.go`:

**Constructor tests:**
- NewMemoryBackend creates empty backend where Exists returns false
- NewMemoryBackendWithState stores provided state

**Load tests:**
- Load returns ErrNotFound when empty
- Load returns state when populated
- Load returns deep copy (modifying returned state doesn't affect backend)

**Save tests:**
- Save stores state that can be retrieved via Load
- Save creates deep copy (modifying original after save doesn't affect backend)
- Save replaces existing state

**Exists tests:**
- Exists returns false when empty
- Exists returns true after Save

**Delete tests:**
- Delete clears state (subsequent Exists returns false)
- Delete on empty backend succeeds

**Thread safety tests:**
- Concurrent reads don't block each other
- Concurrent Save and Load operations complete without data races
- Run with `go test -race` to verify

**Deep copy tests:**
- Verify nested structures are properly copied
- Modify copied state and verify original is unchanged

## Technical Details

### Import Dependencies

```go
import (
    "context"
    "sync"
    "time"

    "github.com/jmgilman/sow/libs/schemas/project"
)
```

### Deep Copy Pattern

Example deep copy implementation:

```go
func copyProjectState(src *project.ProjectState) *project.ProjectState {
    if src == nil {
        return nil
    }

    dst := &project.ProjectState{
        Name:        src.Name,
        Type:        src.Type,
        Branch:      src.Branch,
        Description: src.Description,
        Created_at:  src.Created_at,
        Updated_at:  src.Updated_at,
        Statechart: project.StatechartState{
            Current_state: src.Statechart.Current_state,
            Updated_at:    src.Statechart.Updated_at,
        },
    }

    // Copy phases map
    if src.Phases != nil {
        dst.Phases = make(map[string]project.PhaseState, len(src.Phases))
        for k, v := range src.Phases {
            dst.Phases[k] = copyPhaseState(v)
        }
    }

    // Copy agent_sessions map
    if src.Agent_sessions != nil {
        dst.Agent_sessions = make(map[string]string, len(src.Agent_sessions))
        for k, v := range src.Agent_sessions {
            dst.Agent_sessions[k] = v
        }
    }

    return dst
}

func copyPhaseState(src project.PhaseState) project.PhaseState {
    dst := project.PhaseState{
        Status:       src.Status,
        Enabled:      src.Enabled,
        Created_at:   src.Created_at,
        Started_at:   src.Started_at,
        Completed_at: src.Completed_at,
        Failed_at:    src.Failed_at,
        Iteration:    src.Iteration,
    }

    // Copy inputs
    if src.Inputs != nil {
        dst.Inputs = make([]project.ArtifactState, len(src.Inputs))
        for i, a := range src.Inputs {
            dst.Inputs[i] = copyArtifactState(a)
        }
    }

    // Copy outputs
    if src.Outputs != nil {
        dst.Outputs = make([]project.ArtifactState, len(src.Outputs))
        for i, a := range src.Outputs {
            dst.Outputs[i] = copyArtifactState(a)
        }
    }

    // Copy tasks
    if src.Tasks != nil {
        dst.Tasks = make([]project.TaskState, len(src.Tasks))
        for i, t := range src.Tasks {
            dst.Tasks[i] = copyTaskState(t)
        }
    }

    // Copy metadata (shallow copy of map values)
    if src.Metadata != nil {
        dst.Metadata = make(map[string]any, len(src.Metadata))
        for k, v := range src.Metadata {
            dst.Metadata[k] = v
        }
    }

    return dst
}

// Similar implementations for copyArtifactState and copyTaskState
```

## Relevant Inputs

- `.sow/knowledge/designs/sdk-consolidation-design.md` - MemoryBackend design (lines 230-253)
- `libs/schemas/project/cue_types_gen.go` - ProjectState and nested types
- `state/errors.go` from Task 020 - ErrNotFound error
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Test Usage Example

```go
func TestProjectWorkflow(t *testing.T) {
    // Create backend with initial state
    backend := state.NewMemoryBackendWithState(&project.ProjectState{
        Name:   "test-project",
        Type:   "standard",
        Branch: "feat/test",
        Phases: map[string]project.PhaseState{
            "planning": {Status: "in_progress"},
        },
    })

    // Load project
    ctx := context.Background()
    projectState, err := backend.Load(ctx)
    require.NoError(t, err)
    assert.Equal(t, "test-project", projectState.Name)

    // Modify and save
    projectState.Phases["planning"] = project.PhaseState{Status: "completed"}
    err = backend.Save(ctx, projectState)
    require.NoError(t, err)

    // Verify persistence
    updated, err := backend.Load(ctx)
    require.NoError(t, err)
    assert.Equal(t, "completed", updated.Phases["planning"].Status)
}
```

### Race Condition Test Example

```go
func TestMemoryBackend_ConcurrentAccess(t *testing.T) {
    backend := state.NewMemoryBackend()
    ctx := context.Background()

    var wg sync.WaitGroup

    // Concurrent writers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            state := &project.ProjectState{
                Name: fmt.Sprintf("project-%d", n),
            }
            backend.Save(ctx, state)
        }(i)
    }

    // Concurrent readers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            backend.Load(ctx)
            backend.Exists(ctx)
        }()
    }

    wg.Wait()
    // If we get here without race detector errors, test passes
}
```

## Dependencies

- Task 020: Backend interface and errors must exist

## Constraints

- Do NOT use any filesystem operations - this is purely in-memory
- Do NOT add any external dependencies - only standard library and libs/schemas
- Always perform deep copies to ensure test isolation
- Use RWMutex for performance (multiple concurrent reads)
- This backend is NOT intended for production use (document this clearly)
