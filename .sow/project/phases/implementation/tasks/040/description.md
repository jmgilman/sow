# Task 040: BuildMachine with Closure Binding

# Task 040: BuildMachine with Closure Binding

## Objective

Implement `BuildMachine()` that creates a state machine with guards bound to project instances via closures. This enables guards to access live project state while being defined declaratively in project type configs.

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "BuildMachine Implementation" (lines 717-763) for complete implementation details
- Section "Template binding pattern" (line 224) for architectural decision
- Section "State Management - Event Determination" (lines 305-327) for OnAdvance context

**Existing Code:**
- `cli/internal/sdks/project/state/registry.go` has stub `BuildMachine()` method that returns nil
- `cli/internal/sdks/state/` contains MachineBuilder with `NewBuilder()`, `AddTransition()`, `WithGuard()`, etc.
- Guard type in state machine SDK: `type GuardFunc func() bool`

**Prerequisite:** Task 030 completed (Registry and ProjectTypeConfig accessible)

**What This Task Builds:**
The bridge between declarative configuration (GuardTemplate functions) and runtime state machines (bound GuardFunc closures).

## Requirements

### Core Implementation

Modify `cli/internal/sdks/project/state/registry.go`:

Replace the stub:
```go
func (ptc *ProjectTypeConfig) BuildMachine(_ *Project, _ State) *Machine {
    return nil
}
```

With full implementation:
```go
func (ptc *ProjectTypeConfig) BuildMachine(
    project *Project,
    initialState State,
) *stateMachine.Machine {
    builder := stateMachine.NewBuilder(initialState)

    // Add all transitions with guards bound to project instance
    for _, tc := range ptc.transitions {
        var boundGuard func() bool
        if tc.guardTemplate != nil {
            // Closure captures project - guard can access live state
            boundGuard = func() bool {
                return tc.guardTemplate(project)
            }
        }

        var boundOnEntry func() error
        if tc.onEntry != nil {
            boundOnEntry = func() error {
                return tc.onEntry(project)
            }
        }

        var boundOnExit func() error
        if tc.onExit != nil {
            boundOnExit = func() error {
                return tc.onExit(project)
            }
        }

        builder.AddTransition(
            tc.From,
            tc.To,
            tc.Event,
            stateMachine.WithGuard(boundGuard),
            stateMachine.WithOnEntry(boundOnEntry),
            stateMachine.WithOnExit(boundOnExit),
        )
    }

    return builder.Build()
}
```

### Key Concepts

**Guard Templates vs Bound Guards:**
- **GuardTemplate:** `func(*Project) bool` - Defined in config, receives project explicitly
- **Bound Guard:** `func() bool` - Captures project in closure, matches state machine SDK signature

**Closure Binding:**
```go
// Template defined in config
guardTemplate := func(p *Project) bool {
    return p.PhaseOutputApproved("planning", "task_list")
}

// Bound to project instance via closure
boundGuard := func() bool {
    return guardTemplate(project) // project captured in closure
}
```

This allows:
- Guards defined declaratively in project type config
- Guards accessing live project state during evaluation
- Guards matching state machine SDK's expected signature

## Files to Modify

1. `cli/internal/sdks/project/state/registry.go` - Implement BuildMachine() method
2. `cli/internal/sdks/project/state/machine_test.go` (create) - Behavioral tests

## Testing Requirements (TDD)

Create `cli/internal/sdks/project/state/machine_test.go`:

**BuildMachine Tests:**
- BuildMachine() creates state machine initialized with given state
- Machine has transitions from ProjectTypeConfig
- Guards are properly bound (can access project state)
- Bound guards return correct values based on project state
- Machine.CanFire() respects guard results
- OnEntry actions execute and can mutate project
- OnExit actions execute and can mutate project
- Transitions without guards work (no guard = always allowed)
- Transitions without actions work (actions are optional)

**Test Pattern - Guard Binding:**
```go
func TestBuildMachineGuardsAccessProjectState(t *testing.T) {
    // Create project with known state
    project := &Project{
        ProjectState: schemas.ProjectState{
            Phases: map[string]schemas.PhaseState{
                "test": {
                    Outputs: []schemas.ArtifactState{
                        {Type: "result", Approved: true},
                    },
                },
            },
        },
    }

    // Create config with guard that checks project state
    config := &ProjectTypeConfig{
        transitions: []TransitionConfig{
            {
                From:  State("start"),
                To:    State("end"),
                Event: Event("advance"),
                guardTemplate: func(p *Project) bool {
                    return p.PhaseOutputApproved("test", "result")
                },
            },
        },
    }

    // Build machine
    machine := config.BuildMachine(project, State("start"))

    // Verify guard works
    can, err := machine.CanFire(Event("advance"))
    if err != nil {
        t.Fatal(err)
    }
    if !can {
        t.Error("expected transition to be allowed (guard should pass)")
    }

    // Change project state
    project.Phases["test"].Outputs[0].Approved = false

    // Verify guard reflects new state
    can, _ = machine.CanFire(Event("advance"))
    if can {
        t.Error("expected transition to be blocked (guard should fail)")
    }
}
```

**Test Pattern - Actions:**
```go
func TestBuildMachineActionsCanMutateProject(t *testing.T) {
    project := &Project{
        ProjectState: schemas.ProjectState{
            Phases: map[string]schemas.PhaseState{
                "test": {Status: "pending"},
            },
        },
    }

    config := &ProjectTypeConfig{
        transitions: []TransitionConfig{
            {
                From:  State("start"),
                To:    State("end"),
                Event: Event("advance"),
                onEntry: func(p *Project) error {
                    p.Phases["test"].Status = "active"
                    return nil
                },
            },
        },
    }

    machine := config.BuildMachine(project, State("start"))
    machine.Fire(Event("advance"))

    if project.Phases["test"].Status != "active" {
        t.Errorf("expected onEntry to set status=active, got %s",
            project.Phases["test"].Status)
    }
}
```

## Acceptance Criteria

- [ ] BuildMachine() creates state machine initialized with initialState
- [ ] All transitions from config are added to machine
- [ ] Guards are bound to project instance via closures
- [ ] Bound guards can access project state and return correct bool
- [ ] Machine.CanFire() correctly evaluates bound guards
- [ ] OnEntry actions are bound and execute on transition
- [ ] OnExit actions are bound and execute on transition
- [ ] Actions can mutate project state
- [ ] Transitions without guards work (always allowed)
- [ ] Transitions without actions work (no-op)
- [ ] All tests pass (100% coverage of BuildMachine behavior)
- [ ] Code compiles without errors

## Dependencies

**Required:** Task 030 (ProjectTypeConfig with transitions)

## Technical Notes

- Import state machine SDK: `import stateMachine "github.com/jmgilman/sow/cli/internal/sdks/state"`
- Use `stateMachine.NewBuilder(initialState)` to create builder
- Use `stateMachine.WithGuard()`, `stateMachine.WithOnEntry()`, `stateMachine.WithOnExit()` options
- Closure binding pattern: `func() { return template(captured) }`
- Guards/actions are optional (nil check before binding)
- The Machine type returned is from the state machine SDK, not the stub in project.go

**State Machine SDK Reference:**
- `NewBuilder(initialState State) *MachineBuilder` - Creates builder
- `AddTransition(from, to State, event Event, opts ...TransitionOption)` - Adds transition
- `WithGuard(guard func() bool)` - Guard option
- `WithOnEntry(action func(context.Context, ...any) error)` - Entry action option
- `WithOnExit(action func(context.Context, ...any) error)` - Exit action option
- `Build() *Machine` - Builds machine

Note: The SDK actions take `context.Context` but we're using simpler signatures. Wrap if needed:
```go
if tc.onEntry != nil {
    action := tc.onEntry
    boundOnEntry = func(_ context.Context, _ ...any) error {
        return action(project)
    }
}
```

## Estimated Time

2 hours
