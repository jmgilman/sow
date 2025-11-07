# Task 010: Transition Descriptions (WithDescription Option)

## Context

This task implements the foundation for the AddBranch API by adding support for human-readable descriptions on state machine transitions. Descriptions are critical for:
- Explaining what each transition does in branching states
- Enabling the CLI to display available transitions with context
- Self-documenting state machines for orchestrator guidance

This is the first task in a 5-phase TDD implementation of the AddBranch API for state-determined branching. All subsequent tasks depend on transition descriptions being available.

### Project Goal

Implement a declarative API for state-determined branching in the sow project SDK. The `AddBranch()` API will extend the existing fluent builder pattern to allow project type developers to define branching logic explicitly rather than hiding it in OnAdvance determiners.

### Why This Task First

The `WithDescription()` option is used by:
- Branch paths defined in `When()` clauses (Phase 2)
- Introspection methods like `GetTransitionDescription()` (Phase 5)
- Manual transitions that want to document their purpose

By establishing this foundation first, all other components can use descriptions from the start.

## Requirements

### 1. Add Description Field to TransitionConfig

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/types.go`

Extend the `TransitionConfig` struct to include a description field:

```go
type TransitionConfig struct {
    From  sdkstate.State
    To    sdkstate.State
    Event sdkstate.Event

    guardTemplate GuardTemplate
    onEntry      Action
    onExit       Action
    failedPhase  string
    description  string  // NEW: Human-readable description of this transition
}
```

The description field should be:
- Exported (lowercase `description`) for internal use only
- Optional (empty string if not provided)
- Context-specific (same event from different states can have different meanings)

### 2. Implement WithDescription Option Function

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/options.go`

Create a new `TransitionOption` function that sets the description:

```go
// WithDescription adds a human-readable description to a transition.
//
// Descriptions are:
// - Context-specific (same event from different states can have different meanings)
// - Co-located with guards and actions
// - Used by CLI --list to show what each transition does
// - Visible to orchestrators for decision-making
//
// Best practice: Always add descriptions for transitions, especially in branching states.
//
// Example:
//   WithDescription("Review approved - proceed to finalization")
//
func WithDescription(description string) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.description = description
    }
}
```

The option should:
- Follow the existing option function pattern (see `WithGuard`, `WithOnEntry`, etc.)
- Accept a string parameter
- Modify the TransitionConfig's description field
- Be chainable with other TransitionOption functions

### 3. Test Coverage

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/options_test.go`

Add comprehensive unit test following TDD approach:

```go
func TestWithDescription(t *testing.T) {
    t.Run("sets description on transition config", func(t *testing.T) {
        // Test that WithDescription modifies TransitionConfig.description
    })

    t.Run("works with other options", func(t *testing.T) {
        // Test that WithDescription can be combined with WithGuard, WithOnEntry, etc.
    })

    t.Run("empty description is allowed", func(t *testing.T) {
        // Test that empty string is acceptable (optional feature)
    })
}
```

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder_test.go`

Add integration test for AddTransition with description:

```go
// Add to TestAddTransition function
t.Run("stores description in transition config", func(t *testing.T) {
    builder := NewProjectTypeConfigBuilder(testProjectName)

    builder.AddTransition(testState1, testState2, testEvent1,
        WithDescription("Test transition description"))

    require.Len(t, builder.transitions, 1)
    assert.Equal(t, "Test transition description", builder.transitions[0].description)
})
```

## Acceptance Criteria

- [ ] `TransitionConfig` struct has a `description` field of type `string`
- [ ] `WithDescription()` option function is implemented in options.go
- [ ] `WithDescription()` follows the existing option function pattern
- [ ] `WithDescription()` can be combined with other TransitionOption functions
- [ ] Unit tests pass for `TestWithDescription` covering:
  - Setting description on transition config
  - Combining with other options (guards, actions)
  - Empty descriptions are allowed
- [ ] Integration test in `TestAddTransition` verifies description is stored
- [ ] Code follows existing SDK patterns and style
- [ ] No breaking changes to existing functionality

## Technical Details

### Existing Pattern Reference

The SDK uses the **option function pattern** for configuration. Study these examples:

**From options.go** (lines 66-73):
```go
func WithGuard(description string, guardFunc func(*state.Project) bool) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.guardTemplate = GuardTemplate{
            Description: description,
            Func:        guardFunc,
        }
    }
}
```

**From builder.go** (lines 64-82):
```go
func (b *ProjectTypeConfigBuilder) AddTransition(
    from, to sdkstate.State,
    event sdkstate.Event,
    opts ...TransitionOption,
) *ProjectTypeConfigBuilder {
    tc := TransitionConfig{
        From:  from,
        To:    to,
        Event: event,
    }

    // Apply all options to the transition config
    for _, opt := range opts {
        opt(&tc)
    }

    b.transitions = append(b.transitions, tc)
    return b
}
```

### Testing Pattern Reference

**From builder_test.go** (lines 138-149):
```go
t.Run("applies options correctly", func(t *testing.T) {
    builder := NewProjectTypeConfigBuilder(testProjectName)

    guardFunc := func(_ *state.Project) bool { return true }
    entryAction := func(_ *state.Project) error { return nil }
    exitAction := func(_ *state.Project) error { return nil }

    builder.AddTransition(testState1, testState2, testEvent1,
        WithGuard("test guard", guardFunc),
        WithOnEntry(entryAction),
        WithOnExit(exitAction),
    )

    // Assert options were applied...
})
```

### Implementation Order (TDD)

1. **Write test first**: Create `TestWithDescription` in options_test.go
2. **Run test (should fail)**: Verify test fails because WithDescription doesn't exist
3. **Add description field**: Modify TransitionConfig struct in types.go
4. **Implement WithDescription**: Create function in options.go
5. **Run test (should pass)**: Verify implementation is correct
6. **Add integration test**: Extend TestAddTransition in builder_test.go
7. **Verify all tests pass**: Run full test suite

### Why Not Auto-Generate from Guards?

Guards and descriptions serve different purposes:
- **Guards**: Validate preconditions (e.g., "all tasks complete")
- **Descriptions**: Explain intent (e.g., "Review approved - proceed to finalization")

A transition might have no guard but still need a description, or have a guard with a technical description that differs from the user-facing description of what the transition means.

## Relevant Inputs

### Core SDK Files (Read for Context)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/types.go` - TransitionConfig struct definition (modify this)
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/options.go` - Existing option functions (add WithDescription here)
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder.go` - How AddTransition applies options (understand pattern)

### Test Files (Extend These)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/options_test.go` - Add TestWithDescription here
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder_test.go` - Add description test to TestAddTransition

### Documentation (Context)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/project/context/issue-77.md` - Complete project requirements
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/knowledge/designs/sdk-addbranch-api.md` - API specification (section 5.3 for WithDescription)

## Examples

### Usage Example

```go
// In a project type definition
builder.AddTransition(
    sdkstate.State(ReviewActive),
    sdkstate.State(FinalizeChecks),
    sdkstate.Event(EventReviewPass),
    project.WithGuard("latest review approved", latestReviewApproved),
    project.WithDescription("Review approved - proceed to finalization"),
)
```

### Test Example

```go
func TestWithDescription(t *testing.T) {
    t.Run("sets description on transition config", func(t *testing.T) {
        tc := TransitionConfig{}
        opt := WithDescription("Test description")

        opt(&tc)

        assert.Equal(t, "Test description", tc.description)
    })

    t.Run("works with other options", func(t *testing.T) {
        tc := TransitionConfig{}
        guardFunc := func(_ *state.Project) bool { return true }

        WithGuard("test guard", guardFunc)(&tc)
        WithDescription("Test description")(&tc)

        assert.Equal(t, "test guard", tc.guardTemplate.Description)
        assert.Equal(t, "Test description", tc.description)
        assert.NotNil(t, tc.guardTemplate.Func)
    })
}
```

## Dependencies

None - this is the foundation task.

## Constraints

### Must Follow Existing Patterns

- Use lowercase `description` field (internal only, not exported)
- Follow the `func(*TransitionConfig)` signature for option functions
- Include godoc comments explaining purpose and usage
- Use existing test helpers and assertions (testify/assert, testify/require)

### No Breaking Changes

- Existing transitions without descriptions must continue working
- Description is optional (empty string is valid)
- No changes to public API signatures

### Performance

- Option functions are called at builder time (not runtime)
- Description storage has negligible memory impact
- No runtime overhead

## Success Criteria

This task is complete when:

1. A developer can add `WithDescription()` to any `AddTransition()` call
2. The description is stored in the TransitionConfig
3. All unit tests pass
4. All integration tests pass
5. Code follows existing SDK patterns and style
6. No existing tests are broken

The ultimate validation will be using `WithDescription()` in the AddBranch implementation (Task 030).
