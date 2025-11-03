# State Machine SDK

This package provides a fluent state machine SDK for building project workflows in sow.

## Overview

The state SDK allows project types to define state machines that control workflow progression. It supports:

- **Fluent Builder API**: Define states, transitions, guards, and actions with method chaining
- **Guard Conditions**: Control when transitions are allowed based on project state
- **Event System**: Trigger state changes via named events
- **Prompt Generation**: Auto-generate prompts based on current state
- **State Persistence**: Save and restore machine state

## Origin

This SDK was forked from `internal/project/statechart/` to provide a reusable foundation for all project types. The original implementation was specific to the standard project type.

## Usage Example

```go
import (
    "github.com/jmgilman/sow/cli/internal/sdks/state"
    "github.com/jmgilman/sow/cli/schemas/project"
)

// Define guard conditions
planningComplete := false
designApproved := false

// Build a state machine
machine := state.NewBuilder(&projectState, promptGenerator).
    State("PlanningActive").
        OnEntry(func(ctx context.Context, s *project.ProjectState) error {
            // Initialize planning phase
            return nil
        }).
        Permit("complete", "DesignActive").
            When(func(ctx context.Context, s *project.ProjectState) bool {
                return planningComplete
            }).
        Build().
    State("DesignActive").
        OnEntry(func(ctx context.Context, s *project.ProjectState) error {
            // Initialize design phase
            return nil
        }).
        Permit("approve", "ImplementationPlanning").
            When(func(ctx context.Context, s *project.ProjectState) bool {
                return designApproved
            }).
        Build().
    InitialState("PlanningActive").
    Build()

// Fire events to transition states
if err := machine.Fire(ctx, "complete"); err != nil {
    // handle error
}
```

## Key Components

### Builder (`builder.go`)
Fluent API for defining state machines with states, transitions, guards, and actions.

### Machine (`machine.go`)
Runtime state machine that:
- Tracks current state
- Fires events to trigger transitions
- Validates transitions with guards
- Executes entry/exit actions
- Generates prompts for current state

### Guards (`guards.go`)
Boolean functions that control transition availability:
```go
type GuardFunc func(context.Context, *project.ProjectState) bool
```

### Events (`events.go`)
Named triggers that cause state transitions:
```go
type Event string
```

### Prompt Generator (`prompt_generator.go`)
Generates prompts based on current state and available transitions.

## Testing

The SDK has comprehensive behavioral tests (42 tests, >90% coverage):

```bash
go test ./internal/sdks/state -v
```

Tests verify:
- State transitions and guard evaluation
- Entry/exit action execution
- Prompt generation for different states
- Complex workflows (12 states, 20+ transitions)
- Error handling and validation

## Design Philosophy

This SDK follows **behavioral testing** principles:
- Tests verify **what happens** (outcomes), not **how it happens** (internals)
- High coverage of real-world workflows
- Tests serve as living documentation

## Future Enhancements

This SDK is feature-frozen until we identify specific needs:
- Current functionality is sufficient for all project types
- Changes will be driven by actual requirements
- Focus is on stability and reliability
