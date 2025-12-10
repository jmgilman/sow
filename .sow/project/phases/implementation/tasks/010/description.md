# Task 010: Create Module Foundation and Core Types

## Context

This task is part of the `libs/project` module consolidation effort. The goal is to create a new standalone Go module at `libs/project/` that consolidates `cli/internal/sdks/state/` and `cli/internal/sdks/project/` into a single cohesive module.

The libs/project module will provide:
- Project SDK for defining project types with phases, transitions, and state machines
- Storage backend abstraction for project state persistence
- Runtime types for state machine execution

This task focuses on setting up the module foundation and defining core types that other tasks will build upon. It establishes the module structure and consolidates type definitions that were previously scattered across multiple packages.

## Requirements

### 1. Create Module Structure

Create the following directory structure:
```
libs/project/
├── go.mod                    # Module definition
├── go.sum                    # Dependencies
├── doc.go                    # Package documentation
└── types.go                  # Core type definitions
```

### 2. Initialize Go Module

Create `go.mod` with:
- Module path: `github.com/jmgilman/sow/libs/project`
- Go version: match other libs modules (1.25.3)
- Dependencies:
  - `github.com/jmgilman/go/fs/core` - For FS interface in backends
  - `github.com/jmgilman/sow/libs/schemas` - For CUE-generated project types
  - `github.com/qmuntal/stateless` - For state machine implementation
  - `github.com/stretchr/testify` - For testing
  - `gopkg.in/yaml.v3` - For YAML marshaling

Add replace directives for local development (matching pattern in libs/config/go.mod):
```go
replace (
    github.com/jmgilman/go/fs/billy => /Users/josh/code/go/fs/billy
    github.com/jmgilman/go/fs/core => /Users/josh/code/go/fs/core
    github.com/jmgilman/sow/libs/schemas => ../schemas
)
```

### 3. Create Package Documentation (doc.go)

Create `doc.go` with comprehensive package documentation following the style in `libs/config/doc.go`:
- Overview of what the package provides
- Key concepts: project types, phases, state machines, backends
- Example usage for common operations
- Links to subpackages (state/)

### 4. Define Core Types (types.go)

Consolidate type definitions from `cli/internal/sdks/state/states.go`, `cli/internal/sdks/state/events.go`, `cli/internal/sdks/project/types.go`:

```go
// State represents a state in the project lifecycle state machine.
type State string

const (
    // NoProject indicates no active project exists in the repository.
    NoProject State = "NoProject"
)

// String returns the string representation of the state.
func (s State) String() string

// Event represents a trigger that causes state transitions.
type Event string

// String returns the string representation of the event.
func (e Event) String() string

// Guard is a condition function that determines if a transition is allowed.
// Returns true if the transition should proceed, false otherwise.
type Guard func() bool

// GuardTemplate is a template function bound to a project instance.
// It receives the project and returns whether the transition is allowed.
// Description provides human-readable explanation for error messages.
type GuardTemplate struct {
    Description string
    Func        func(*state.Project) bool
}

// Action is a function that mutates project state during transitions.
type Action func(*state.Project) error

// EventDeterminer examines project state and determines the next event
// for the generic Advance() command.
type EventDeterminer func(*state.Project) (Event, error)

// PromptGenerator creates a contextual prompt for a given state.
type PromptGenerator func(*state.Project) string

// PromptFunc generates a prompt for a state during transitions.
// Used by the state machine builder.
type PromptFunc func(State) string
```

**Important**: The `*state.Project` references will create a forward reference to the `state` subpackage. For now, use a placeholder import and comment indicating this will be resolved when the state subpackage is created in Task 040.

### 5. Create State Subpackage Stub

Create a minimal `state/doc.go` file to establish the subpackage structure:

```go
// Package state provides project state types and persistence.
//
// This package contains:
//   - Project wrapper type with runtime behavior
//   - Phase, Task, and Artifact types
//   - Collection types for phases, tasks, and artifacts
//   - Backend interface for storage abstraction
//   - YAML and memory backend implementations
//   - Load/Save operations with CUE validation
package state
```

## Acceptance Criteria

1. [ ] `libs/project/go.mod` exists with correct module path and dependencies
2. [ ] `go mod tidy` succeeds in the libs/project directory
3. [ ] `libs/project/doc.go` provides comprehensive package documentation
4. [ ] `libs/project/types.go` defines all core types (State, Event, Guard, GuardTemplate, Action, EventDeterminer, PromptGenerator, PromptFunc)
5. [ ] `libs/project/state/doc.go` exists with subpackage documentation
6. [ ] All exported types have doc comments following STYLE.md guidelines
7. [ ] Code compiles without errors
8. [ ] `golangci-lint run ./...` passes with no issues
9. [ ] `go test -race ./...` passes with no failures
10. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
11. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)
- Write unit tests for State.String() and Event.String() methods
- Use table-driven tests with `t.Run()` following TESTING.md patterns
- Use `testify/assert` for assertions
- Tests should verify the String() methods return the expected values
- Focus on behavioral coverage, not line coverage

## Technical Details

### Module Path
```
github.com/jmgilman/sow/libs/project
```

### Import Patterns
Future files in this module will use:
```go
import (
    "github.com/jmgilman/sow/libs/project"
    "github.com/jmgilman/sow/libs/project/state"
)
```

### File Organization (per STYLE.md)
In types.go, organize code in this order:
1. Imports
2. Constants (NoProject)
3. Type declarations (State, Event)
4. Type methods (String())
5. Struct types (GuardTemplate)
6. Function types (Guard, Action, EventDeterminer, PromptGenerator, PromptFunc)

## Relevant Inputs

- `cli/internal/sdks/state/states.go` - Original State type definition
- `cli/internal/sdks/state/events.go` - Original Event type definition
- `cli/internal/sdks/state/builder.go` - Contains PromptFunc and GuardFunc types
- `cli/internal/sdks/project/types.go` - GuardTemplate, Action, EventDeterminer, PromptGenerator types
- `libs/config/go.mod` - Reference for go.mod structure and replace directives
- `libs/config/doc.go` - Reference for doc.go style
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### types.go Structure
```go
package project

// State represents a state in the project lifecycle state machine.
// States are string constants defined by project type configurations.
// All project types share the NoProject state for when no project exists.
type State string

const (
    // NoProject indicates no active project exists in the repository.
    // This is a shared state used by all project types.
    NoProject State = "NoProject"
)

// String returns the string representation of the state.
func (s State) String() string {
    return string(s)
}

// Event represents a trigger that causes state transitions.
// Events are defined by project type configurations and fired
// when advancing through the project lifecycle.
type Event string

// String returns the string representation of the event.
func (e Event) String() string {
    return string(e)
}

// ... rest of types
```

### types_test.go Structure
```go
package project

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestState_String(t *testing.T) {
    tests := []struct {
        name  string
        state State
        want  string
    }{
        {name: "NoProject constant", state: NoProject, want: "NoProject"},
        {name: "custom state", state: State("PlanningActive"), want: "PlanningActive"},
        {name: "empty state", state: State(""), want: ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.state.String()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestEvent_String(t *testing.T) {
    tests := []struct {
        name  string
        event Event
        want  string
    }{
        {name: "custom event", event: Event("AdvancePlanning"), want: "AdvancePlanning"},
        {name: "empty event", event: Event(""), want: ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.event.String()
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Dependencies

- None - this is the foundation task

## Constraints

- Do NOT implement full functionality for types that reference `*state.Project` - those will be completed in Task 040
- Do NOT copy test files yet - write new tests following TESTING.md
- Do NOT add any CLI-specific code or dependencies
- Keep module isolated from `cli/internal/sow` package
