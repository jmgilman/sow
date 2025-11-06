# Task 010: Create Exploration Package Structure

## Context

This task creates the foundational package structure for the exploration project type. The exploration project type is a new workflow for research, investigation, and knowledge gathering that follows a 2-phase model: Exploration → Finalization.

The sow framework uses a Project SDK that enables declarative configuration of project types using a builder pattern. Each project type is defined in its own package under `cli/internal/projects/` and registers itself with the global registry during package initialization.

This task creates the package directory and core registration file following the established pattern from the standard project type.

## Requirements

### Package Location

Create the exploration package at:
```
cli/internal/projects/exploration/
```

### Core Registration File

Create `cli/internal/projects/exploration/exploration.go` with the following structure:

1. **Package declaration and imports**:
   - Import `github.com/jmgilman/sow/cli/internal/sdks/project`
   - Import `github.com/jmgilman/sow/cli/internal/sdks/project/state`
   - Import `sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"`
   - Import `projschema "github.com/jmgilman/sow/cli/schemas/project"`

2. **init() function**:
   - Register the exploration project type with the global registry
   - Pattern: `state.Register("exploration", NewExplorationProjectConfig())`

3. **NewExplorationProjectConfig() function**:
   - Create and return `*project.ProjectTypeConfig`
   - Use builder pattern: `project.NewProjectTypeConfigBuilder("exploration")`
   - Call configuration helper functions (to be implemented in other tasks):
     - `configurePhases(builder)`
     - `configureTransitions(builder)`
     - `configureEventDeterminers(builder)`
     - `configurePrompts(builder)`
   - Set initializer: `builder.WithInitializer(initializeExplorationProject)`
   - Return: `builder.Build()`

4. **initializeExplorationProject() function**:
   - Signature: `func(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error`
   - Create both phases: "exploration" and "finalization"
   - For exploration phase:
     - Set `status = "active"` (starts immediately)
     - Set `enabled = true`
   - For finalization phase:
     - Set `status = "pending"`
     - Set `enabled = false`
   - Use `p.Created_at` for timestamps
   - Initialize with empty outputs, tasks, and metadata map
   - Handle initial inputs if provided (from initialInputs map)

5. **Configuration helper function stubs**:
   - `configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder`
   - `configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder`
   - `configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder`
   - `configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder`
   - These will return the builder unchanged for now (stub implementations)

## Test-Driven Development

This task follows TDD methodology:

1. **Write tests first** for:
   - Package registration (verify "exploration" appears in registry)
   - `initializeExplorationProject()` behavior:
     - Creates both phases
     - Exploration phase has status="active" and enabled=true
     - Finalization phase has status="pending" and enabled=false
     - Handles initial inputs correctly
   - `NewExplorationProjectConfig()` returns valid config

2. **Run tests** - they should fail initially (red phase)

3. **Implement functionality** - write minimum code to pass tests (green phase)

4. **Refactor** - improve code quality while keeping tests passing

Place tests in `cli/internal/projects/exploration/exploration_test.go`.

## Acceptance Criteria

- [ ] Package directory `cli/internal/projects/exploration/` exists
- [ ] File `exploration.go` exists with correct package declaration
- [ ] **Unit tests written before implementation**
- [ ] Tests cover registration, initialization, and phase creation
- [ ] All tests pass
- [ ] `init()` function registers "exploration" with global registry
- [ ] `NewExplorationProjectConfig()` creates builder and calls all configuration helpers
- [ ] `initializeExplorationProject()` creates both phases with correct initial state
- [ ] Exploration phase starts in "active" status with enabled=true
- [ ] Finalization phase starts in "pending" status with enabled=false
- [ ] All configuration helper functions defined as stubs
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### Package Registration Pattern

The registration happens via `init()` which runs when the package is imported. The main application imports the package with a blank identifier to trigger registration:

```go
import _ "github.com/jmgilman/sow/cli/internal/projects/exploration"
```

This pattern is used by the standard project type in `cli/cmd/root.go:18`.

### Phase Initialization

Phases are stored in `p.Phases` map with keys "exploration" and "finalization". Each phase is a `projschema.PhaseState` struct with fields:
- `Status`: Current phase status
- `Enabled`: Whether phase is active
- `Created_at`: Timestamp
- `Started_at`: Timestamp (zero value initially)
- `Completed_at`: Timestamp (zero value initially)
- `Failed_at`: Timestamp (zero value initially)
- `Iteration`: Counter (0 initially)
- `Metadata`: map[string]interface{}
- `Inputs`: []projschema.ArtifactState
- `Outputs`: []projschema.ArtifactState
- `Tasks`: []projschema.TaskState

### Builder Pattern

The Project SDK uses a fluent builder API where configuration methods return the builder for chaining. The pattern:
```go
builder.
    WithPhase("name", options...).
    WithPhase("another", options...).
    AddTransition(from, to, event, options...).
    Build()
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/standard.go` - Reference implementation showing registration and builder pattern
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/builder.go` - Builder API documentation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/state/registry.go` - Registry implementation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Original requirements
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Complete design specification

## Examples

### Standard Project Registration (Reference)

From `cli/internal/projects/standard/standard.go`:

```go
func init() {
    state.Register("standard", NewStandardProjectConfig())
}

func NewStandardProjectConfig() *project.ProjectTypeConfig {
    builder := project.NewProjectTypeConfigBuilder("standard")
    builder = configurePhases(builder)
    builder = configureTransitions(builder)
    builder = configureEventDeterminers(builder)
    builder = configurePrompts(builder)
    builder = builder.WithInitializer(initializeStandardProject)
    return builder.Build()
}
```

### Phase Initialization Pattern

From `cli/internal/projects/standard/standard.go:28-59`:

```go
func initializeStandardProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
    now := p.Created_at
    phaseNames := []string{"implementation", "review", "finalize"}

    for _, phaseName := range phaseNames {
        inputs := []projschema.ArtifactState{}
        if initialInputs != nil {
            if phaseInputs, exists := initialInputs[phaseName]; exists {
                inputs = phaseInputs
            }
        }

        p.Phases[phaseName] = projschema.PhaseState{
            Status:     "pending",
            Enabled:    false,
            Created_at: now,
            Inputs:     inputs,
            Outputs:    []projschema.ArtifactState{},
            Tasks:      []projschema.TaskState{},
            Metadata:   make(map[string]interface{}),
        }
    }

    return nil
}
```

## Dependencies

None - this is the first task and creates the foundation for all other tasks.

## Constraints

- Must follow existing project type patterns exactly for consistency
- All stub functions must have correct signatures to avoid compilation errors in later tasks
- Phase names must be exactly "exploration" and "finalization" (lowercase)
- Exploration phase must start active (this is unique to exploration project type)
- Cannot implement actual configuration logic yet - that comes in later tasks
