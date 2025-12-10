# libs/project

Project SDK for defining project types with phases, state machines, and persistence.

## Quick Start

```go
import (
    "github.com/jmgilman/sow/libs/project"
    "github.com/jmgilman/sow/libs/project/state"
)

// Define a project type
config := project.NewProjectTypeConfigBuilder("mytype").
    SetInitialState(project.State("Planning")).
    WithPhase("planning",
        project.StartState(project.State("Planning")),
        project.EndState(project.State("Planning")),
        project.OutputTypes("task_list"),
    ).
    AddTransition(
        project.State("Planning"),
        project.State("Implementation"),
        project.Event("AdvancePlanning"),
    ).
    Build()

// Register it
state.Register("mytype", config)

// Load a project
backend := state.NewYAMLBackend(fs)
proj, err := state.Load(ctx, backend)
if err != nil {
    return fmt.Errorf("load project: %w", err)
}
```

## Usage

### Defining a Project Type

Use the fluent builder API to define project types:

```go
config := project.NewProjectTypeConfigBuilder("standard").
    SetInitialState(StateIdea).
    WithPhase("planning",
        project.StartState(StatePlanning),
        project.EndState(StatePlanning),
        project.SupportsTasks(),
    ).
    WithPhase("implementation",
        project.StartState(StateImplementing),
        project.EndState(StateImplementing),
        project.SupportsTasks(),
    ).
    AddTransition(
        StatePlanning,
        StateImplementing,
        EventStartImplementation,
        project.WithProjectGuard("planning approved", func(p *state.Project) bool {
            return p.PhaseOutputApproved("planning", "task_list")
        }),
    ).
    Build()
```

### Loading a Project

```go
// From filesystem with YAML backend
fs := billy.NewLocal(".sow")
backend := state.NewYAMLBackend(fs)
proj, err := state.Load(ctx, backend)
if err != nil {
    return fmt.Errorf("load project: %w", err)
}

// Convenience wrapper
proj, err := state.LoadFromFS(ctx, fs)
```

### Creating a Project

```go
proj, err := state.Create(ctx, backend, state.CreateOpts{
    Branch:      "feat/my-feature",
    Description: "Add new feature",
})
if err != nil {
    return fmt.Errorf("create project: %w", err)
}
```

### Saving Changes

```go
proj.Phases["planning"].Status = "completed"
if err := state.Save(ctx, proj); err != nil {
    return fmt.Errorf("save project: %w", err)
}
```

### Working with State Machines

```go
// Build a machine from config
machine := config.BuildProjectMachine(proj, project.State(proj.Statechart.Current_state))

// Check current state
currentState := machine.State()

// Fire an event
if err := machine.Fire(EventAdvance); err != nil {
    return fmt.Errorf("fire event: %w", err)
}

// Check what transitions are available
triggers := machine.PermittedTriggers()

// Check if a specific event can fire
if machine.CanFire(EventComplete) {
    // ...
}
```

### Testing with Memory Backend

```go
func TestMyFeature(t *testing.T) {
    backend := state.NewMemoryBackend()
    proj, err := state.Create(ctx, backend, state.CreateOpts{
        Branch:      "feat/test",
        Description: "Test project",
    })
    require.NoError(t, err)

    // Test logic...

    loaded, err := state.Load(ctx, backend)
    require.NoError(t, err)
}
```

## Package Structure

```
libs/project/
├── types.go              # Core types (State, Event, Guard, Action)
├── machine.go            # State machine wrapper
├── builder.go            # Machine builder
├── options.go            # Transition options
├── config.go             # ProjectTypeConfig
├── project_builder.go    # ProjectTypeConfigBuilder
├── config_options.go     # Project-level transition options
├── phase_config.go       # Phase configuration
├── transition_config.go  # Transition configuration
├── branch.go             # Branch configuration for state-determined branching
├── errors.go             # Error types
└── state/
    ├── project.go        # Project wrapper type
    ├── phase.go          # Phase helpers
    ├── task.go           # Task type
    ├── artifact.go       # Artifact type
    ├── collections.go    # Collection types
    ├── backend.go        # Backend interface
    ├── backend_yaml.go   # YAML file backend
    ├── backend_memory.go # In-memory backend (testing)
    ├── loader.go         # Load/Create/Save functions
    ├── registry.go       # Project type registry
    ├── validate.go       # CUE validation
    └── errors.go         # Error types
```

## Key Concepts

### Project Types

Project types define the lifecycle of a project through:
- **Phases**: Major divisions of work (planning, implementation, review)
- **States**: Positions in the state machine
- **Events**: Triggers for state transitions
- **Guards**: Conditions that must be true for transitions
- **Actions**: Side effects that run during transitions

### State Machine

The state machine wraps qmuntal/stateless and manages project lifecycle:

```go
machine := config.BuildProjectMachine(proj, initialState)
machine.State()              // Current state
machine.Fire(event)          // Trigger transition
machine.CanFire(event)       // Check if transition allowed
machine.PermittedTriggers()  // Available events
```

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

### Guards and Actions

Guards prevent transitions when conditions aren't met:

```go
project.WithProjectGuard("all tasks complete", func(p *state.Project) bool {
    return p.AllTasksComplete()
})
```

Actions run during transitions:

```go
project.WithOnEntry(func(ctx context.Context, args ...any) error {
    // Side effects on entering state
    return nil
})
```

### Branching

State-determined branching allows different paths based on project state:

```go
builder.AddBranch(StateReview,
    project.BranchOn(func(p *state.Project) string {
        if p.PhaseMetadataBool("review", "approved") {
            return "approved"
        }
        return "rejected"
    }),
    project.When("approved",
        project.To(StateComplete),
        project.Event(EventApprove),
    ),
    project.When("rejected",
        project.To(StatePlanning),
        project.Event(EventReject),
    ),
)
```

## Links

- [Go Package Documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/project)
- [state subpackage](./state/)
