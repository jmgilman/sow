# Task 050: Migrate State Machine and Builder

## Context

This task is part of the `libs/project` module consolidation effort. It migrates the state machine integration from `cli/internal/sdks/state/` to `libs/project/`.

The state machine provides:
- A wrapper around `qmuntal/stateless` for project lifecycle management
- A builder API for configuring transitions, guards, and actions
- Support for prompts and event determination

This migration consolidates the state machine code into the `libs/project` module, keeping it at the top level (not in `state/` subpackage) since it's part of the SDK's public API for defining project types.

## Requirements

### 1. Migrate Machine Type (machine.go)

Migrate the Machine wrapper from `cli/internal/sdks/state/machine.go`:

```go
// Machine wraps a qmuntal/stateless state machine with project-specific behavior.
type Machine struct {
    fsm       *stateless.StateMachine
    promptGen PromptFunc
}

// NewMachine creates a new Machine wrapper.
// This is typically called by MachineBuilder.Build().
func NewMachine(fsm *stateless.StateMachine, promptGen PromptFunc) *Machine

// State returns the current state.
func (m *Machine) State() State

// Fire triggers a state transition with the given event.
// Returns an error if the transition is not allowed.
func (m *Machine) Fire(event Event) error

// CanFire returns true if the given event can be fired from the current state.
func (m *Machine) CanFire(event Event) bool

// PermittedTriggers returns all events that can be fired from the current state.
func (m *Machine) PermittedTriggers() []Event

// Prompt returns the prompt for the current state.
// Returns empty string if no prompt is configured for the state.
func (m *Machine) Prompt() string
```

### 2. Migrate MachineBuilder (builder.go)

Migrate the builder from `cli/internal/sdks/state/builder.go`:

```go
// MachineBuilder provides a fluent API for building state machines.
type MachineBuilder struct {
    initialState State
    promptFunc   PromptFunc
    transitions  []transitionDef
}

// transitionDef holds the configuration for a single transition.
type transitionDef struct {
    from    State
    to      State
    event   Event
    options []TransitionOption
}

// NewBuilder creates a new MachineBuilder with the given initial state.
// The promptFunc parameter is optional and provides state-specific prompts.
func NewBuilder(initialState State, promptFunc PromptFunc) *MachineBuilder

// AddTransition adds a transition from one state to another.
func (b *MachineBuilder) AddTransition(from, to State, event Event, opts ...TransitionOption) *MachineBuilder

// Build creates the state machine with all configured transitions.
func (b *MachineBuilder) Build() *Machine
```

### 3. Migrate Transition Options (options.go)

Migrate transition option functions:

```go
// TransitionOption configures a transition.
type TransitionOption func(*transitionConfig)

// transitionConfig holds internal configuration for a transition.
type transitionConfig struct {
    guard            Guard
    guardDescription string
    onEntry          func(context.Context, ...any) error
    onExit           func(context.Context, ...any) error
}

// WithGuard sets a guard function that must return true for the transition to proceed.
func WithGuard(guard Guard) TransitionOption

// WithGuardDescription sets a guard with a human-readable description.
// The description is used in error messages when the guard fails.
func WithGuardDescription(description string, guard Guard) TransitionOption

// WithOnEntry sets an action to run when entering the target state.
func WithOnEntry(action func(context.Context, ...any) error) TransitionOption

// WithOnExit sets an action to run when leaving the source state.
func WithOnExit(action func(context.Context, ...any) error) TransitionOption
```

### 4. Integration with stateless Library

The Machine type wraps `github.com/qmuntal/stateless`. Key integration points:

```go
// In builder.go Build() method
func (b *MachineBuilder) Build() *Machine {
    fsm := stateless.NewStateMachine(b.initialState)

    for _, t := range b.transitions {
        cfg := fsm.Configure(t.from).Permit(t.event, t.to)

        // Apply options
        config := &transitionConfig{}
        for _, opt := range t.options {
            opt(config)
        }

        if config.guard != nil {
            if config.guardDescription != "" {
                cfg.PermitIf(t.event, t.to, func(_ context.Context, _ ...any) bool {
                    return config.guard()
                }, config.guardDescription)
            } else {
                cfg.PermitIf(t.event, t.to, func(_ context.Context, _ ...any) bool {
                    return config.guard()
                })
            }
        }

        if config.onEntry != nil {
            fsm.Configure(t.to).OnEntry(config.onEntry)
        }

        if config.onExit != nil {
            fsm.Configure(t.from).OnExit(config.onExit)
        }
    }

    return &Machine{fsm: fsm, promptGen: b.promptFunc}
}
```

### 5. Error Handling

When `Fire()` fails, return descriptive errors:
- "transition not allowed: cannot fire event X from state Y"
- Include guard failure descriptions if available

## Acceptance Criteria

1. [ ] `machine.go` defines Machine type with all methods
2. [ ] `builder.go` defines MachineBuilder with fluent API
3. [ ] `options.go` defines TransitionOption functions
4. [ ] Machine wraps stateless correctly
5. [ ] Guards and actions are properly integrated
6. [ ] Prompts are retrievable per state
7. [ ] Error messages are descriptive with context
8. [ ] All types are properly documented with doc comments
9. [ ] Code compiles without errors
10. [ ] `golangci-lint run ./...` passes with no issues
11. [ ] `go test -race ./...` passes with no failures
12. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
13. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write comprehensive tests:

**machine_test.go:**
- State returns current state
- Fire transitions to new state
- Fire returns error for invalid transition
- CanFire returns true/false correctly
- PermittedTriggers returns available events
- Prompt returns configured prompt

**builder_test.go:**
- NewBuilder sets initial state
- AddTransition adds transition correctly
- Build creates working machine
- Multiple transitions from same state work
- Guards block transitions when false
- Guards allow transitions when true
- OnEntry actions run on transition
- OnExit actions run on transition

**Integration tests:**
- Build complex machine with multiple states and transitions
- Verify guard descriptions appear in errors
- Verify prompt generation works correctly

## Technical Details

### Import Dependencies

```go
import (
    "context"
    "fmt"

    "github.com/qmuntal/stateless"
)
```

### stateless Library Usage

Key `qmuntal/stateless` features used:
- `stateless.NewStateMachine(initialState)` - Create machine
- `fsm.Configure(state)` - Configure state
- `.Permit(trigger, destState)` - Add transition
- `.PermitIf(trigger, destState, guard, description)` - Guarded transition
- `.OnEntry(action)` - Entry action
- `.OnExit(action)` - Exit action
- `fsm.Fire(trigger)` - Fire event
- `fsm.CanFire(trigger)` - Check if event can fire
- `fsm.MustState()` - Get current state

### Type Conversions

The stateless library uses `interface{}` for states and triggers. Handle conversions:

```go
// State() returns project.State
func (m *Machine) State() State {
    return State(m.fsm.MustState().(string))
}

// Fire converts Event to trigger
func (m *Machine) Fire(event Event) error {
    return m.fsm.Fire(stateless.Trigger(event))
}

// PermittedTriggers converts triggers to Events
func (m *Machine) PermittedTriggers() []Event {
    triggers := m.fsm.PermittedTriggers()
    events := make([]Event, len(triggers))
    for i, t := range triggers {
        events[i] = Event(t.(string))
    }
    return events
}
```

## Relevant Inputs

- `cli/internal/sdks/state/machine.go` - Current Machine implementation
- `cli/internal/sdks/state/builder.go` - Current MachineBuilder implementation
- `cli/internal/sdks/state/states.go` - State type definition
- `cli/internal/sdks/state/events.go` - Event type definition
- `libs/project/types.go` from Task 010 - State, Event, Guard, PromptFunc types
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Building a Simple State Machine

```go
builder := project.NewBuilder(project.State("Pending"), nil)

builder.
    AddTransition(
        project.State("Pending"),
        project.State("Active"),
        project.Event("Start"),
    ).
    AddTransition(
        project.State("Active"),
        project.State("Completed"),
        project.Event("Finish"),
        project.WithGuard(func() bool {
            return allTasksComplete()
        }),
    )

machine := builder.Build()

// Use the machine
fmt.Println(machine.State())  // "Pending"
machine.Fire(project.Event("Start"))
fmt.Println(machine.State())  // "Active"
```

### With Prompts and Actions

```go
promptFunc := func(s project.State) string {
    switch s {
    case project.State("PlanningActive"):
        return "Planning phase is active. Create task breakdown."
    default:
        return ""
    }
}

builder := project.NewBuilder(project.State("PlanningActive"), promptFunc)

builder.AddTransition(
    project.State("PlanningActive"),
    project.State("ImplementationReady"),
    project.Event("AdvancePlanning"),
    project.WithGuardDescription("all planning tasks complete", func() bool {
        return planningComplete()
    }),
    project.WithOnExit(func(ctx context.Context, args ...any) error {
        return finalizePlanning()
    }),
)

machine := builder.Build()
fmt.Println(machine.Prompt())  // "Planning phase is active..."
```

### Error Handling

```go
machine := builder.Build()

// Try to fire invalid event
err := machine.Fire(project.Event("InvalidEvent"))
if err != nil {
    // Error includes state and attempted event
    fmt.Println(err)  // "transition not allowed: cannot fire 'InvalidEvent' from state 'Pending'"
}

// Guard failure
err = machine.Fire(project.Event("Finish"))
if err != nil {
    // Error includes guard description
    fmt.Println(err)  // "guard 'all tasks complete' failed for event 'Finish'"
}
```

## Dependencies

- Task 010: Core types (State, Event, Guard, PromptFunc)

## Constraints

- Do NOT change the stateless library behavior - wrap it cleanly
- Do NOT include project-specific logic - keep machine generic
- Keep the builder fluent API intact
- All state/event operations should be type-safe (no raw strings in public API)
- Guard descriptions are optional but encouraged
