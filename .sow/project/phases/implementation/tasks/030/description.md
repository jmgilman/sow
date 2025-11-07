# Task 030: AddBranch Builder Method (Core Auto-Generation Logic)

## Context

This task implements the `AddBranch()` builder method, which is the central piece of the state-determined branching API. This method auto-generates state machine transitions and event determiners from declarative branch configurations.

This is the third task in a 5-phase TDD implementation of the AddBranch API. It builds on Task 010 (transition descriptions) and Task 020 (branch data structures) to provide the complete branching functionality.

### Project Goal

Implement a declarative API for state-determined branching in the sow project SDK. The `AddBranch()` method enables project type developers to configure branching states by:
1. Specifying a discriminator function that examines project state
2. Defining branch paths for each possible discriminator value
3. Automatically generating the necessary transitions and event determiner

### Why This Task Third

Now that we have:
- Transition descriptions (Task 010)
- Branch data structures and option functions (Task 020)

We can implement the core auto-generation logic that uses these components to create a working branching system.

## Requirements

### 1. Add branches Field to ProjectTypeConfigBuilder

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go`

Extend the `ProjectTypeConfigBuilder` struct (lines 13-22):

```go
type ProjectTypeConfigBuilder struct {
    name               string
    phaseConfigs       map[string]*PhaseConfig
    initialState       sdkstate.State
    transitions        []TransitionConfig
    onAdvance          map[sdkstate.State]EventDeterminer
    prompts            map[sdkstate.State]PromptGenerator
    orchestratorPrompt PromptGenerator
    initializer        state.Initializer
    branches           map[sdkstate.State]*BranchConfig  // NEW: state -> branch config
}
```

Update `NewProjectTypeConfigBuilder()` (lines 27-35) to initialize the map:

```go
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder {
    return &ProjectTypeConfigBuilder{
        name:         name,
        phaseConfigs: make(map[string]*PhaseConfig),
        transitions:  make([]TransitionConfig, 0),
        onAdvance:    make(map[sdkstate.State]EventDeterminer),
        prompts:      make(map[sdkstate.State]PromptGenerator),
        branches:     make(map[sdkstate.State]*BranchConfig),  // NEW
    }
}
```

### 2. Implement AddBranch Builder Method

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go`

Add after the `OnAdvance()` method (around line 93):

```go
// AddBranch configures state-determined branching from a state.
//
// This method provides a declarative API for defining multi-way branches where
// the next state can be determined by examining project state. It auto-generates:
//   1. Transitions for each When clause
//   2. Event determiner using the discriminator function
//
// The discriminator examines project state and returns a string value. This value
// is matched against the values defined in When() clauses to determine which
// event to fire and which state to transition to.
//
// Example (binary branch):
//   AddBranch(
//       sdkstate.State(ReviewActive),
//       project.BranchOn(func(p *state.Project) string {
//           // Get review assessment from latest approved review
//           assessment := getReviewAssessment(p)
//           return assessment  // "pass" or "fail"
//       }),
//       project.When("pass",
//           sdkstate.Event(EventReviewPass),
//           sdkstate.State(FinalizeChecks),
//           project.WithDescription("Review approved - proceed to finalization"),
//       ),
//       project.When("fail",
//           sdkstate.Event(EventReviewFail),
//           sdkstate.State(ImplementationPlanning),
//           project.WithDescription("Review failed - return to planning for rework"),
//           project.WithFailedPhase("review"),
//       ),
//   )
//
// This auto-generates:
//   - Two AddTransition calls (one per When clause)
//   - One OnAdvance determiner that calls discriminator and maps value to event
//
// Returns the builder for method chaining.
func (b *ProjectTypeConfigBuilder) AddBranch(
    from sdkstate.State,
    opts ...BranchOption,
) *ProjectTypeConfigBuilder {
    // Create BranchConfig from options
    bc := &BranchConfig{
        from: from,
    }

    // Apply all options to build the branch config
    for _, opt := range opts {
        opt(bc)
    }

    // Validate configuration
    if bc.discriminator == nil {
        // Panic here because this is a programming error at config time
        panic(fmt.Sprintf("AddBranch for state %s: no discriminator provided (use BranchOn)", from))
    }
    if len(bc.branches) == 0 {
        panic(fmt.Sprintf("AddBranch for state %s: no branch paths provided (use When)", from))
    }

    // Generate AddTransition calls for each branch path
    for value, path := range bc.branches {
        // Collect transition options from branch path
        var transOpts []TransitionOption

        if path.description != "" {
            transOpts = append(transOpts, WithDescription(path.description))
        }
        if path.guardTemplate.Func != nil {
            transOpts = append(transOpts, WithGuard(path.guardTemplate.Description, path.guardTemplate.Func))
        }
        if path.onEntry != nil {
            transOpts = append(transOpts, WithOnEntry(path.onEntry))
        }
        if path.onExit != nil {
            transOpts = append(transOpts, WithOnExit(path.onExit))
        }
        if path.failedPhase != "" {
            transOpts = append(transOpts, WithFailedPhase(path.failedPhase))
        }

        // Generate transition
        b.AddTransition(from, path.to, path.event, transOpts...)

        // Store value in path for error messages (used by generated OnAdvance)
        _ = value // Used in closure below
    }

    // Generate OnAdvance determiner
    b.OnAdvance(from, func(p *state.Project) (sdkstate.Event, error) {
        // Call discriminator to get branch value
        value := bc.discriminator(p)

        // Look up branch path for this value
        path, exists := bc.branches[value]
        if !exists {
            // Build helpful error message with available values
            availableValues := make([]string, 0, len(bc.branches))
            for v := range bc.branches {
                availableValues = append(availableValues, fmt.Sprintf("%q", v))
            }
            return "", fmt.Errorf(
                "no branch defined for discriminator value %q from state %s (available: %v)",
                value, from, availableValues)
        }

        // Return the event for this branch
        return path.event, nil
    })

    // Store branch config for introspection
    b.branches[from] = bc

    return b
}
```

**Key Design Decisions**:
- Validation uses panic (not error) because invalid config is a programming error, not a runtime error
- OnAdvance is auto-generated with helpful error messages listing available values
- All transition options from BranchPath are forwarded to AddTransition
- BranchConfig is stored for later introspection

### 3. Update Build() to Copy branches Map

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go`

Update the `Build()` method (lines 137-170) to copy the branches map:

```go
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig {
    // ... existing code ...

    // Copy branches map (add after prompts copy, before return)
    branches := make(map[sdkstate.State]*BranchConfig, len(b.branches))
    for k, v := range b.branches {
        branches[k] = v
    }

    return &ProjectTypeConfig{
        name:               b.name,
        phaseConfigs:       phaseConfigs,
        initialState:       b.initialState,
        transitions:        transitions,
        onAdvance:          onAdvance,
        prompts:            prompts,
        orchestratorPrompt: b.orchestratorPrompt,
        initializer:        b.initializer,
        branches:           branches,  // NEW
    }
}
```

### 4. Add branches Field to ProjectTypeConfig

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go`

Extend the `ProjectTypeConfig` struct (lines 63-92):

```go
type ProjectTypeConfig struct {
    name string
    phaseConfigs map[string]*PhaseConfig
    initialState sdkstate.State
    transitions []TransitionConfig
    onAdvance map[sdkstate.State]EventDeterminer
    prompts map[sdkstate.State]PromptGenerator
    orchestratorPrompt PromptGenerator
    initializer state.Initializer
    branches map[sdkstate.State]*BranchConfig  // NEW: for introspection
}
```

### 5. Test Coverage

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch_test.go`

Add comprehensive unit tests:

```go
func TestAddBranchGeneratesTransitions(t *testing.T) {
    t.Run("creates transitions for each When clause", func(t *testing.T) {
        // Verify that AddBranch generates AddTransition calls
        builder := NewProjectTypeConfigBuilder("test")

        builder.AddBranch(
            sdkstate.State("BranchState"),
            BranchOn(func(_ *state.Project) string { return "value1" }),
            When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
            When("value2", sdkstate.Event("Event2"), sdkstate.State("State2")),
        )

        // Should have generated 2 transitions
        assert.Len(t, builder.transitions, 2)
    })

    t.Run("forwards transition options correctly", func(t *testing.T) {
        // Verify that options in When clauses are applied to transitions
    })
}

func TestAddBranchGeneratesOnAdvance(t *testing.T) {
    t.Run("creates event determiner using discriminator", func(t *testing.T) {
        // Verify that OnAdvance determiner is created
        builder := NewProjectTypeConfigBuilder("test")

        discriminator := func(p *state.Project) string {
            return "test_value"
        }

        builder.AddBranch(
            sdkstate.State("BranchState"),
            BranchOn(discriminator),
            When("test_value", sdkstate.Event("TestEvent"), sdkstate.State("NextState")),
        )

        // Should have OnAdvance registered
        assert.Contains(t, builder.onAdvance, sdkstate.State("BranchState"))
    })

    t.Run("determiner returns correct event for discriminator value", func(t *testing.T) {
        // Test that the generated determiner calls discriminator and returns matching event
    })
}

func TestAddBranchBinary(t *testing.T) {
    t.Run("binary branch workflow", func(t *testing.T) {
        // Complete workflow test: create config, build machine, test both branches
        builder := NewProjectTypeConfigBuilder("test")

        builder.AddBranch(
            sdkstate.State("ReviewState"),
            BranchOn(func(p *state.Project) string {
                // Mock: return "pass" or "fail" based on project state
                if val, ok := p.Metadata["review_result"].(string); ok {
                    return val
                }
                return "pass"
            }),
            When("pass",
                sdkstate.Event("PassEvent"),
                sdkstate.State("PassState"),
                WithDescription("Test passed"),
            ),
            When("fail",
                sdkstate.Event("FailEvent"),
                sdkstate.State("FailState"),
                WithDescription("Test failed"),
            ),
        )

        config := builder.Build()

        // Test with "pass" discriminator value
        // Test with "fail" discriminator value
    })
}

func TestAddBranchNWay(t *testing.T) {
    t.Run("N-way branch workflow (3+ branches)", func(t *testing.T) {
        // Test with 3 or more branch paths
        builder := NewProjectTypeConfigBuilder("test")

        builder.AddBranch(
            sdkstate.State("DeployState"),
            BranchOn(func(p *state.Project) string {
                return p.Metadata["target"].(string)
            }),
            When("staging", sdkstate.Event("DeployStaging"), sdkstate.State("StagingState")),
            When("production", sdkstate.Event("DeployProd"), sdkstate.State("ProdState")),
            When("canary", sdkstate.Event("DeployCanary"), sdkstate.State("CanaryState")),
        )

        config := builder.Build()

        // Test all 3 paths
    })
}
```

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/integration_test.go`

Add integration test:

```go
func TestReviewBranchingWorkflow(t *testing.T) {
    // Create complete project type with AddBranch
    config := NewProjectTypeConfigBuilder("review-test").
        WithPhase("review",
            WithOutputs("review"),
        ).
        SetInitialState(sdkstate.State("ReviewActive")).
        AddBranch(
            sdkstate.State("ReviewActive"),
            BranchOn(func(p *state.Project) string {
                phase := p.Phases["review"]
                for i := len(phase.Outputs) - 1; i >= 0; i-- {
                    artifact := phase.Outputs[i]
                    if artifact.Type == "review" && artifact.Approved {
                        if assessment, ok := artifact.Metadata["assessment"].(string); ok {
                            return assessment
                        }
                    }
                }
                return ""
            }),
            When("pass",
                sdkstate.Event("ReviewPass"),
                sdkstate.State("FinalizeState"),
                WithDescription("Review approved - proceed"),
            ),
            When("fail",
                sdkstate.Event("ReviewFail"),
                sdkstate.State("ReworkState"),
                WithDescription("Review failed - rework"),
            ),
        ).
        Build()

    // Create test project
    proj := createTestProject(t, "review-test")

    // Build machine starting in ReviewActive
    machine := config.BuildMachine(proj, sdkstate.State("ReviewActive"))

    // Test 1: Set review assessment to "pass"
    // Add review artifact with assessment="pass", approved=true
    // Determine event (should return ReviewPass)
    // Fire event (should transition to FinalizeState)

    // Test 2: Reset to ReviewActive, set assessment to "fail"
    // Determine event (should return ReviewFail)
    // Fire event (should transition to ReworkState)
}
```

## Acceptance Criteria

- [ ] `branches` field added to ProjectTypeConfigBuilder and ProjectTypeConfig
- [ ] `NewProjectTypeConfigBuilder()` initializes branches map
- [ ] `AddBranch()` method implemented with:
  - Full godoc with examples
  - Validation for missing discriminator and branches
  - Auto-generation of transitions via AddTransition
  - Auto-generation of OnAdvance determiner
  - Storage of BranchConfig for introspection
  - Chainable return (returns *ProjectTypeConfigBuilder)
- [ ] `Build()` copies branches map to config
- [ ] Unit tests pass for:
  - TestAddBranchGeneratesTransitions
  - TestAddBranchGeneratesOnAdvance
  - TestAddBranchBinary
  - TestAddBranchNWay
- [ ] Integration test `TestReviewBranchingWorkflow` passes
- [ ] Error messages are helpful (list available values)
- [ ] Code follows existing SDK patterns
- [ ] No breaking changes to existing functionality

## Technical Details

### Auto-Generation Logic

`AddBranch()` performs three main operations:

1. **Generate Transitions**: For each `When()` clause, call `AddTransition()` with:
   - from: The branch source state
   - to: The path's target state
   - event: The path's event
   - opts: All options from the BranchPath

2. **Generate OnAdvance Determiner**: Create a function that:
   - Calls the discriminator to get a value
   - Looks up the value in bc.branches
   - Returns the matching event
   - Returns helpful error if value not found

3. **Store BranchConfig**: Save in builder.branches for introspection

### Error Handling Strategy

**Build-time validation** (uses panic):
- No discriminator provided → panic
- No When clauses provided → panic
- Rationale: These are programming errors that should be caught during development

**Runtime errors** (returns error):
- Discriminator returns unmapped value → error with available values
- Rationale: This could happen in production if state data is unexpected

### Closure Binding Pattern

The generated OnAdvance determiner captures `bc` (BranchConfig) in a closure:

```go
b.OnAdvance(from, func(p *state.Project) (sdkstate.Event, error) {
    value := bc.discriminator(p)  // bc captured from outer scope
    path := bc.branches[value]
    return path.event, nil
})
```

This matches the existing pattern used in `BuildMachine()` where template functions are bound to project instances via closures.

### Integration with Existing SDK

AddBranch uses existing builder methods:
- Calls `b.AddTransition()` directly (no need to duplicate logic)
- Calls `b.OnAdvance()` directly (no need to duplicate logic)
- Stores in `b.branches` for introspection (new feature)

This ensures:
- Consistency with existing behavior
- Reuse of tested code
- No special cases in other parts of the SDK

### Why Store BranchConfig?

The branches map is needed for:
- `IsBranchingState()` introspection method (Task 050)
- `GetAvailableTransitions()` to include branch info (Task 050)
- Future tooling that wants to understand branching structure

## Relevant Inputs

### Implementation Files (Modify These)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go` - Add AddBranch method, branches field
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go` - Add branches field

### Branch Types (Created in Task 020)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch.go` - BranchConfig, BranchPath, option functions

### Test Files (Extend These)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch_test.go` - Add AddBranch tests
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/integration_test.go` - Add workflow test

### Pattern Reference

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go` - See AddTransition (lines 64-82), OnAdvance (lines 84-93) for patterns
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/machine.go` - See BuildMachine for closure binding pattern

### Reference Implementation (Current Workaround)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/projects/standard/standard.go` - Lines 128-175 (ReviewActive transitions), lines 220-255 (ReviewActive OnAdvance)

### Documentation

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/project/context/issue-77.md` - Complete requirements
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/knowledge/designs/sdk-addbranch-api.md` - Sections 5-6 for API spec and auto-generation logic

## Examples

### Usage Example (Binary Branch)

```go
config := NewProjectTypeConfigBuilder("standard").
    // ... phase configs ...
    AddBranch(
        sdkstate.State(ReviewActive),
        project.BranchOn(func(p *state.Project) string {
            phase := p.Phases["review"]
            for i := len(phase.Outputs) - 1; i >= 0; i-- {
                artifact := phase.Outputs[i]
                if artifact.Type == "review" && artifact.Approved {
                    return artifact.Metadata["assessment"].(string)
                }
            }
            return ""
        }),
        project.When("pass",
            sdkstate.Event(EventReviewPass),
            sdkstate.State(FinalizeChecks),
            project.WithDescription("Review approved - proceed to finalization"),
        ),
        project.When("fail",
            sdkstate.Event(EventReviewFail),
            sdkstate.State(ImplementationPlanning),
            project.WithDescription("Review failed - return to planning for rework"),
            project.WithFailedPhase("review"),
        ),
    ).
    Build()
```

This auto-generates:
- Two transitions (pass and fail)
- One OnAdvance determiner
- Stores BranchConfig for introspection

### Usage Example (N-Way Branch)

```go
builder.AddBranch(
    sdkstate.State(DeploymentReady),
    project.BranchOn(func(p *state.Project) string {
        return p.Phases["deployment"].Metadata["target"].(string)
    }),
    project.When("staging",
        sdkstate.Event(EventDeployStaging),
        sdkstate.State(DeployingStaging),
        project.WithDescription("Deploy to staging"),
    ),
    project.When("production",
        sdkstate.Event(EventDeployProduction),
        sdkstate.State(DeployingProduction),
        project.WithGuard("all tests passed", allTestsPassed),
        project.WithDescription("Deploy to production"),
    ),
    project.When("canary",
        sdkstate.Event(EventDeployCanary),
        sdkstate.State(DeployingCanary),
        project.WithDescription("Deploy to canary"),
    ),
)
```

## Dependencies

- Task 010 (Transition Descriptions) - MUST be complete
  - AddBranch uses WithDescription option
- Task 020 (Branch Data Structures) - MUST be complete
  - AddBranch uses BranchConfig, BranchPath, BranchOn, When

## Constraints

### Validation Strategy

- Use panic for config errors (missing discriminator, no branches)
- Use error returns for runtime issues (unmapped discriminator value)
- Provide helpful error messages with available values

### Performance

- All auto-generation happens during Build() (once per project type)
- No runtime overhead compared to manual AddTransition + OnAdvance
- Map lookup is O(1) for discriminator value → event mapping

### Thread Safety

- Not required (builders are not thread-safe by design)
- Configs are immutable after Build()

## Success Criteria

This task is complete when:

1. `AddBranch()` method is fully implemented and tested
2. Binary branching works (2 paths)
3. N-way branching works (3+ paths)
4. Transitions are auto-generated correctly
5. OnAdvance determiner is auto-generated correctly
6. Error messages are helpful
7. All unit tests pass
8. Integration test demonstrates complete workflow
9. Code is well-documented
10. No breaking changes to existing functionality

After this task, the core AddBranch API is functional. Task 040 will add error case testing, and Task 050 will add introspection methods.
