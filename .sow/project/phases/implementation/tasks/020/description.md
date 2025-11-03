# Task 020: Builder API Implementation

# Task 020: Builder API Implementation

## Objective

Implement the fluent builder API for defining project types. The builder provides a declarative interface for configuring all aspects of a project type: phases, state machine, event determination, and prompts.

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "APIs and Interfaces - Project SDK Builder API" (lines 579-639) for complete API specification
- Section "Complete Usage Example" (lines 673-831) for real-world usage patterns

**Prerequisite:** Task 010 completed (config types and options pattern exist)

**What This Task Builds:**
The fluent builder API that project type authors use to configure their types declaratively.

## Requirements

### Builder Structure

Create `cli/internal/sdks/project/builder.go`:

**ProjectTypeConfigBuilder** - Fluent API for building project type configs:
```go
type ProjectTypeConfigBuilder struct {
    name            string
    phaseConfigs    map[string]*PhaseConfig
    initialState    State
    transitions     []TransitionConfig
    onAdvance       map[State]EventDeterminer
    prompts         map[State]PromptGenerator
}
```

### Builder Methods

**Constructor:**
```go
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder
```
- Creates builder with name
- Initializes empty collections (maps, slices)
- Returns pointer to builder

**Phase Configuration:**
```go
func (b *ProjectTypeConfigBuilder) WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder
```
- Creates PhaseConfig with given name
- Applies all opts to the config
- Stores in phaseConfigs map
- Returns builder (chainable)

**State Machine Configuration:**
```go
func (b *ProjectTypeConfigBuilder) SetInitialState(state State) *ProjectTypeConfigBuilder
```
- Sets initialState field
- Returns builder (chainable)

```go
func (b *ProjectTypeConfigBuilder) AddTransition(
    from, to State,
    event Event,
    opts ...TransitionOption,
) *ProjectTypeConfigBuilder
```
- Creates TransitionConfig with from/to/event
- Applies all opts to the config
- Appends to transitions slice
- Returns builder (chainable)

**Event Determination:**
```go
func (b *ProjectTypeConfigBuilder) OnAdvance(
    state State,
    determiner EventDeterminer,
) *ProjectTypeConfigBuilder
```
- Stores determiner in onAdvance map for given state
- Returns builder (chainable)

**Prompt Configuration:**
```go
func (b *ProjectTypeConfigBuilder) WithPrompt(
    state State,
    generator PromptGenerator,
) *ProjectTypeConfigBuilder
```
- Stores generator in prompts map for given state
- Returns builder (chainable)

**Build:**
```go
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig
```
- Creates ProjectTypeConfig with all accumulated data
- Copies all fields from builder
- Returns pointer to new config
- Does NOT reset builder (builder remains reusable)

## Files to Create

1. `cli/internal/sdks/project/builder.go` - Builder implementation
2. `cli/internal/sdks/project/builder_test.go` - Behavioral tests

## Testing Requirements (TDD)

Create `cli/internal/sdks/project/builder_test.go` with behavior tests:

**Constructor Tests:**
- NewProjectTypeConfigBuilder() creates builder with given name
- Builder has empty collections initialized

**Phase Configuration Tests:**
- WithPhase() adds phase config
- WithPhase() applies options correctly
- WithPhase() returns builder (chainable)
- Multiple phases can be added

**State Machine Tests:**
- SetInitialState() sets initial state
- SetInitialState() returns builder (chainable)
- AddTransition() adds transition with from/to/event
- AddTransition() applies options correctly
- AddTransition() returns builder (chainable)
- Multiple transitions can be added

**Event Determination Tests:**
- OnAdvance() stores event determiner for state
- OnAdvance() returns builder (chainable)
- Multiple states can have determiners

**Prompt Configuration Tests:**
- WithPrompt() stores prompt generator for state
- WithPrompt() returns builder (chainable)
- Multiple states can have prompts

**Build Tests:**
- Build() creates ProjectTypeConfig with all data
- Build() copies name from builder
- Build() copies phase configs
- Build() copies initial state
- Build() copies transitions
- Build() copies onAdvance map
- Build() copies prompts map
- Builder is reusable (can call Build() multiple times)

**Chaining Test:**
- All methods can be chained in single expression
- Example: `NewBuilder("test").WithPhase("p1").SetInitialState(s).Build()`

## Acceptance Criteria

- [ ] NewProjectTypeConfigBuilder() creates valid builder
- [ ] WithPhase() adds phase and applies options correctly
- [ ] SetInitialState() sets initial state
- [ ] AddTransition() adds transitions and applies options
- [ ] OnAdvance() configures event determiners
- [ ] WithPrompt() configures prompt generators
- [ ] Build() creates complete ProjectTypeConfig
- [ ] All methods return builder (chainable)
- [ ] Builder is reusable (Build() doesn't reset state)
- [ ] Multiple phases/transitions can be added
- [ ] All tests pass (100% coverage of builder behavior)
- [ ] Code compiles without errors

## Dependencies

**Required:** Task 010 (config types, options pattern, function types)

## Technical Notes

- Import types from `cli/internal/sdks/project` (PhaseConfig, TransitionConfig, etc.)
- Import State/Event from `cli/internal/sdks/state`
- Builder uses pointer receiver for all methods (allows chaining)
- Build() should copy data (not return references to internal maps/slices)
- No validation in builder (validation happens later in Validate())
- Design pattern: Fluent Interface / Method Chaining

**Example Usage Pattern:**
```go
config := NewProjectTypeConfigBuilder("standard").
    WithPhase("planning", WithOutputs("task_list")).
    SetInitialState(NoProject).
    AddTransition(NoProject, PlanningActive, EventProjectInit).
    OnAdvance(PlanningActive, func(p *state.Project) (Event, error) {
        return EventCompletePlanning, nil
    }).
    Build()
```

## Estimated Time

2 hours
