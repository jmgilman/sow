# Task 100: Add README and Final Cleanup

## Context

This is the final task in the `libs/project` module consolidation effort. With all code migrated (Tasks 010-090), this task ensures the module is properly documented, linted, and ready for use.

This task includes:
- Creating comprehensive README.md documentation
- Ensuring all linting passes
- Verifying the full test suite
- Any final cleanup or polish

## Requirements

### 1. Create libs/project/README.md

Create a comprehensive README following the pattern in `libs/config/`:

```markdown
# libs/project

Project SDK for defining project types with phases, state machines, and persistence.

## Overview

This package provides:
- **Project SDK**: Fluent API for defining project types with phases and transitions
- **State Machine**: Generic state machine wrapper around qmuntal/stateless
- **Storage Backends**: Pluggable backends for project state persistence
- **Runtime Types**: Project, Phase, Task, and Artifact wrappers

## Installation

```go
import (
    "github.com/jmgilman/sow/libs/project"
    "github.com/jmgilman/sow/libs/project/state"
)
```

## Quick Start

### Defining a Project Type

```go
config := project.NewProjectTypeConfigBuilder("mytype").
    SetInitialState(project.State("Planning")).
    AddPhase("planning",
        project.WithStartState(project.State("Planning")),
        project.WithEndState(project.State("Planning")),
        project.WithOutputs("task_list"),
    ).
    AddTransition(
        project.State("Planning"),
        project.State("Implementation"),
        project.Event("AdvancePlanning"),
        project.WithGuard("artifacts approved", func(p *state.Project) bool {
            return allArtifactsApproved(p, "planning")
        }),
    ).
    Build()

project.Register("mytype", config)
```

### Loading a Project

```go
// From filesystem
fs := billy.NewLocal(".sow")
proj, err := state.LoadFromFS(ctx, fs)

// With custom backend
backend := state.NewYAMLBackend(fs)
proj, err := state.Load(ctx, backend)
```

### Creating a Project

```go
proj, err := state.CreateOnFS(ctx, fs, "feat/my-feature", "Description", nil)
```

### Saving Changes

```go
proj.Phases["planning"].Status = "completed"
err := state.Save(ctx, proj)
```

## Package Structure

```
libs/project/
├── types.go              # Core types (State, Event, Guard, Action)
├── machine.go            # State machine wrapper
├── builder.go            # Machine builder
├── options.go            # Transition options
├── config.go             # ProjectTypeConfig
├── config_builder.go     # ProjectTypeConfigBuilder
├── branch.go             # Branch configuration
├── registry.go           # Project type registry
└── state/
    ├── project.go        # Project wrapper type
    ├── phase.go          # Phase type and helpers
    ├── task.go           # Task type
    ├── artifact.go       # Artifact type
    ├── collections.go    # Collection types
    ├── backend.go        # Backend interface
    ├── backend_yaml.go   # YAML file backend
    ├── backend_memory.go # In-memory backend (testing)
    ├── loader.go         # Load/Create/Save functions
    └── validate.go       # CUE validation
```

## Key Concepts

### Project Types

Project types define the lifecycle of a project through:
- **Phases**: Major divisions of work (planning, implementation, review)
- **States**: Positions in the state machine
- **Events**: Triggers for state transitions
- **Guards**: Conditions that must be true for transitions
- **Actions**: Side effects that run during transitions

### Backend Interface

The Backend interface abstracts storage:

```go
type Backend interface {
    Load(ctx context.Context) (*project.ProjectState, error)
    Save(ctx context.Context, state *project.ProjectState) error
    Exists(ctx context.Context) (bool, error)
    Delete(ctx context.Context) error
}
```

Built-in implementations:
- `YAMLBackend`: File-based storage (production)
- `MemoryBackend`: In-memory storage (testing)

### State Machine

The state machine manages project lifecycle:

```go
machine := config.BuildMachine(proj, initialState)
machine.State()              // Current state
machine.Fire(event)          // Trigger transition
machine.CanFire(event)       // Check if transition allowed
machine.PermittedTriggers()  // Available events
```

## Testing

Use MemoryBackend for unit tests:

```go
func TestMyFeature(t *testing.T) {
    backend := state.NewMemoryBackend()
    proj, err := state.Create(ctx, backend, "feat/test", "Test", nil)
    require.NoError(t, err)

    // Test logic...

    loaded, err := state.Load(ctx, backend)
    require.NoError(t, err)
}
```

## API Reference

See [pkg.go.dev documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/project).
```

### 2. Create libs/project/state/README.md

Create a README for the state subpackage:

```markdown
# libs/project/state

Project state management with persistence backends.

## Overview

This package provides:
- Project wrapper type with runtime behavior
- Backend interface for pluggable storage
- YAML and memory backend implementations
- Load/Save operations with CUE validation

## Usage

See [parent package README](../README.md) for examples.

## Types

- `Project`: Wrapper around `project.ProjectState` with runtime fields
- `Backend`: Interface for storage backends
- `YAMLBackend`: File-based storage
- `MemoryBackend`: In-memory storage for testing
```

### 3. Run Full Lint Check

Run linting across the entire libs/project module:

```bash
cd libs/project
golangci-lint run ./...
```

Fix any issues:
- Missing doc comments
- Unused code
- Import organization
- Error handling issues
- Style violations

### 4. Run Full Test Suite

Run all tests with race detection:

```bash
cd libs/project
go test -race -v ./...
```

Ensure:
- All tests pass
- No race conditions
- Adequate coverage for key paths

### 5. Verify CLI Integration

Run the full CLI test suite to ensure integration works:

```bash
cd cli
go test -race -v ./...
```

### 6. Update CHANGELOG (if exists)

If the repository has a CHANGELOG, add an entry for this feature:

```markdown
## [Unreleased]

### Added
- `libs/project` module: Consolidated project SDK with Backend interface abstraction
- `libs/project/state`: Project state management with pluggable backends
- `MemoryBackend` for improved testability
- Context-based API with cancellation support

### Changed
- Project state operations now use `Backend` interface instead of `sow.Context`
- Moved project SDK from `cli/internal/sdks/` to `libs/project/`

### Removed
- `cli/internal/sdks/state/` package (consolidated into libs/project)
- `cli/internal/sdks/project/` package (consolidated into libs/project)
```

### 7. Final Code Review Checklist

Review all new code for:

- [ ] All exported types have doc comments
- [ ] All exported functions have doc comments
- [ ] Error messages are clear and actionable
- [ ] No unused code or imports
- [ ] Consistent code style throughout
- [ ] No hardcoded paths or magic strings
- [ ] Thread safety where needed
- [ ] Proper error wrapping with `%w`
- [ ] Tests cover key behaviors
- [ ] No TODO comments without issue references

## Acceptance Criteria

1. [ ] `libs/project/README.md` exists with comprehensive documentation
2. [ ] `libs/project/state/README.md` exists
3. [ ] `golangci-lint run ./...` passes with no issues in libs/project
4. [ ] `golangci-lint run ./...` passes with no issues in cli
5. [ ] `go test -race ./...` passes in libs/project
6. [ ] `go test -race ./...` passes in cli directory
7. [ ] No unused code remains
8. [ ] All doc comments are complete and follow STYLE.md format
9. [ ] CHANGELOG updated (if applicable)
10. [ ] Final code review checklist complete
11. [ ] All code adheres to `.standards/STYLE.md`
12. [ ] All tests adhere to `.standards/TESTING.md`

### Test Requirements

No new tests for this task - focus is on:
- Verifying existing tests pass
- Fixing any test failures discovered
- Ensuring test coverage is adequate

## Technical Details

### Common Lint Issues to Fix

1. **Missing doc comments**:
```go
// BAD
func DoSomething() {}

// GOOD
// DoSomething performs the something operation.
func DoSomething() {}
```

2. **Import organization**:
```go
// GOOD: stdlib, external, internal
import (
    "context"
    "fmt"

    "github.com/external/pkg"

    "github.com/jmgilman/sow/libs/project/state"
)
```

3. **Error handling**:
```go
// BAD
result, _ := doSomething()

// GOOD
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}
```

### Coverage Targets

Aim for:
- 80% behavioral coverage on core types
- 100% coverage on exported functions
- Key error paths tested

Check coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Relevant Inputs

- `libs/config/README.md` - Reference for README style
- `.standards/STYLE.md` - Code style requirements
- `.standards/READMES.md` - README requirements (if exists)
- All new files in `libs/project/`

## Examples

### README Code Examples

Ensure all code examples in READMEs:
- Compile correctly
- Follow project style
- Demonstrate key use cases
- Include error handling

### Lint Command Output

Example successful lint run:
```bash
$ cd libs/project
$ golangci-lint run ./...
$ # No output = no issues
```

## Dependencies

- Tasks 010-090: All previous tasks must be complete

## Constraints

- Do NOT add new features in this task
- Do NOT change behavior
- Focus only on documentation, linting, and cleanup
- Keep READMEs concise but comprehensive
- Follow existing README patterns in the repository
