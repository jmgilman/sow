# Task 010: Core Configuration Types and Options Pattern

# Task 010: Core Configuration Types and Options Pattern

## Objective

Implement the foundational types and options pattern for project type configuration. This includes configuration structures (PhaseConfig, TransitionConfig, ProjectTypeConfig), option functions for flexible configuration, and function type definitions used throughout the SDK.

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "Component Breakdown - Project SDK" (lines 387-428) for SDK structure overview
- Section "APIs and Interfaces - Project SDK Builder API" (lines 579-639) for complete API
- Section "Data Models" (lines 439-575) for type requirements

**Existing Foundation:**
- State machine SDK exists at `cli/internal/sdks/state/` with Builder, Machine, Guards, Events
- State types exist at `cli/internal/sdks/project/state/` with Project, Phase, Artifact, Task wrappers
- CUE schemas at `cli/schemas/project/` define data structures

**What This Task Builds:**
The configuration layer that project types use to define their behavior declaratively.

## Requirements

### 1. Configuration Structures

Create `cli/internal/sdks/project/config.go`:

**PhaseConfig** - Configuration for a single phase:
- `name` (string) - Phase identifier
- `startState` (State) - State when phase begins
- `endState` (State) - State when phase ends
- `allowedInputTypes` ([]string) - Artifact types allowed as inputs (empty = allow all)
- `allowedOutputTypes` ([]string) - Artifact types allowed as outputs (empty = allow all)
- `supportsTasks` (bool) - Whether phase can have tasks
- `metadataSchema` (string) - Embedded CUE schema for metadata validation

**TransitionConfig** - Configuration for a state machine transition:
- `From` (State) - Source state
- `To` (State) - Target state
- `Event` (Event) - Event that triggers transition
- `guardTemplate` (GuardTemplate) - Function template that becomes bound guard
- `onEntry` (Action) - Action to execute when entering target state
- `onExit` (Action) - Action to execute when exiting source state

**ProjectTypeConfig** - Complete configuration for a project type:
- `name` (string) - Project type identifier
- `phaseConfigs` (map[string]*PhaseConfig) - Phase configurations by name
- `initialState` (State) - Starting state of state machine
- `transitions` ([]TransitionConfig) - All state transitions
- `onAdvance` (map[State]EventDeterminer) - Event determiners per state
- `prompts` (map[State]PromptGenerator) - Prompt generators per state

### 2. Function Type Definitions

Create `cli/internal/sdks/project/types.go`:

```go
// GuardTemplate is a template function that gets bound to a project instance
// via closure. It receives the project and returns whether transition is allowed.
type GuardTemplate func(*state.Project) bool

// Action is a function that mutates project state during transitions.
// Returns error if action fails.
type Action func(*state.Project) error

// EventDeterminer examines project state and determines the next event
// for the generic Advance() command. Returns the event or error if unable
// to determine (e.g., missing required state).
type EventDeterminer func(*state.Project) (Event, error)

// PromptGenerator creates a contextual prompt for a given state.
// Returns markdown-formatted string to display to user.
type PromptGenerator func(*state.Project) string
```

Note: Import `State` and `Event` from `cli/internal/sdks/state` package.

### 3. Options Pattern

Create `cli/internal/sdks/project/options.go`:

**Phase Options (type `PhaseOpt func(*PhaseConfig)`):**
- `WithStartState(state State) PhaseOpt` - Sets phase start state
- `WithEndState(state State) PhaseOpt` - Sets phase end state
- `WithInputs(types ...string) PhaseOpt` - Sets allowed input artifact types
- `WithOutputs(types ...string) PhaseOpt` - Sets allowed output artifact types
- `WithTasks() PhaseOpt` - Enables task support for phase
- `WithMetadataSchema(schema string) PhaseOpt` - Sets embedded CUE metadata schema

**Transition Options (type `TransitionOption func(*TransitionConfig)`):**
- `WithGuard(guardFunc GuardTemplate) TransitionOption` - Sets guard template
- `WithOnEntry(action Action) TransitionOption` - Adds entry action
- `WithOnExit(action Action) TransitionOption` - Adds exit action

Each option function should modify the config struct and return the option type.

## Files to Create

1. `cli/internal/sdks/project/config.go` - Configuration structures
2. `cli/internal/sdks/project/types.go` - Function type definitions
3. `cli/internal/sdks/project/options.go` - Option functions
4. `cli/internal/sdks/project/options_test.go` - Behavioral tests

## Testing Requirements (TDD)

Create `cli/internal/sdks/project/options_test.go` with behavior tests:

**Phase Options Tests:**
- WithStartState() sets the start state on PhaseConfig
- WithEndState() sets the end state on PhaseConfig
- WithInputs() sets allowed input types (multiple types work)
- WithOutputs() sets allowed output types (multiple types work)
- WithTasks() enables supportsTasks flag
- WithMetadataSchema() sets schema string
- Multiple options can be applied to single PhaseConfig

**Transition Options Tests:**
- WithGuard() sets guard template function
- WithOnEntry() sets entry action function
- WithOnExit() sets exit action function
- Multiple options can be applied to single TransitionConfig

**Test Pattern:**
```go
func TestWithStartState(t *testing.T) {
    config := &PhaseConfig{}
    opt := WithStartState(State("TestState"))
    opt(config)

    if config.startState != State("TestState") {
        t.Errorf("expected startState TestState, got %v", config.startState)
    }
}
```

## Acceptance Criteria

- [ ] PhaseConfig struct with all required fields defined
- [ ] TransitionConfig struct with all required fields defined
- [ ] ProjectTypeConfig struct with all required fields defined
- [ ] All function types defined (GuardTemplate, Action, EventDeterminer, PromptGenerator)
- [ ] All PhaseOpt functions implemented and working
- [ ] All TransitionOption functions implemented and working
- [ ] Options can be applied in any order
- [ ] Multiple options can be applied to same config
- [ ] All tests pass (100% coverage of option behavior)
- [ ] Code compiles without errors

## Dependencies

None - this is the foundation task.

## Technical Notes

- Use `cli/internal/sdks/state.State` and `cli/internal/sdks/state.Event` types (import from state machine SDK)
- Use `cli/internal/sdks/project/state.Project` type for function signatures
- Options pattern provides flexibility without parameter explosion
- Each option is a function that modifies config in place
- No validation needed in options (validation happens in Build() later)

## Estimated Time

1.5 hours
