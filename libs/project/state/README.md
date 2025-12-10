# libs/project/state

Project state management with persistence backends.

## Overview

This package provides:
- Project wrapper type with runtime behavior
- Backend interface for pluggable storage
- YAML and memory backend implementations
- Load/Save operations with CUE validation
- Phase, Task, and Artifact types
- Project type registry

## Usage

See [parent package README](../README.md) for comprehensive examples.

### Loading a Project

```go
backend := state.NewYAMLBackend(fs)
proj, err := state.Load(ctx, backend)
if err != nil {
    return fmt.Errorf("load project: %w", err)
}
```

### Creating a Project

```go
proj, err := state.Create(ctx, backend, state.CreateOpts{
    Branch:      "feat/my-feature",
    Description: "Add new feature",
})
```

### Saving a Project

```go
if err := state.Save(ctx, proj); err != nil {
    return fmt.Errorf("save project: %w", err)
}
```

### Registering Project Types

```go
state.Register("mytype", config)

cfg, exists := state.GetConfig("mytype")
if !exists {
    return errors.New("unknown project type")
}
```

## Types

### Project

Wrapper around the CUE-generated `ProjectState` with runtime fields:

```go
type Project struct {
    project.ProjectState  // Embedded serializable state
    // Runtime fields (not persisted)
}
```

Methods:
- `AllTasksComplete()` - Check if all tasks are completed
- `PhaseMetadataBool(phase, key)` - Read boolean from phase metadata
- `PhaseOutputApproved(phase, type)` - Check if output artifact is approved
- `Save(ctx)` - Save project state to backend

### Backend

Interface for storage abstraction:

```go
type Backend interface {
    Load(ctx context.Context) (*project.ProjectState, error)
    Save(ctx context.Context, state *project.ProjectState) error
    Exists(ctx context.Context) (bool, error)
    Delete(ctx context.Context) error
}
```

Implementations:
- `YAMLBackend` - File-based storage for production
- `MemoryBackend` - In-memory storage for testing

### Phase Helpers

```go
state.MarkPhaseInProgress(proj, "planning")
state.MarkPhaseCompleted(proj, "planning")
state.MarkPhaseFailed(proj, "planning")
```

## Testing

Use `MemoryBackend` for isolated unit tests:

```go
func TestProject(t *testing.T) {
    backend := state.NewMemoryBackend()
    proj, err := state.Create(ctx, backend, state.CreateOpts{
        Branch:      "feat/test",
        Description: "Test",
    })
    require.NoError(t, err)

    // Test logic...
}
```

## Links

- [Parent Package](../)
- [Go Package Documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/project/state)
