# Task 020: Branch Data Structures and Option Functions

## Context

This task implements the core data structures and option functions for the AddBranch API. It creates the types that represent branch configurations and the functions that build them, establishing the foundation for declarative state-determined branching.

This is the second task in a 5-phase TDD implementation of the AddBranch API. It depends on Task 010 (transition descriptions) and provides the types needed by Task 030 (AddBranch builder method).

### Project Goal

Implement a declarative API for state-determined branching in the sow project SDK. These data structures capture:
- A discriminator function that examines project state and returns a value
- Multiple branch paths mapping discriminator values to transitions

### Why This Task Second

The data structures must exist before `AddBranch()` can use them, but they are independently testable. By implementing and testing them separately, we ensure the types are correct before adding the complex auto-generation logic.

## Requirements

### 1. Create Branch Configuration Types

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch.go` (NEW FILE)

Create a new file with the following types:

```go
package project

import (
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// BranchConfig represents a state-determined branch point in the state machine.
// It captures a discriminator function that examines project state and returns
// a string value, plus a map of branch paths that define what happens for each value.
//
// BranchConfig is used internally by AddBranch to auto-generate transitions and
// event determiners. It's stored in ProjectTypeConfig for introspection.
type BranchConfig struct {
    from          sdkstate.State                           // Source state
    discriminator func(*state.Project) string              // Returns branch value
    branches      map[string]*BranchPath                   // value -> branch path
}

// BranchPath represents one possible branch destination.
// Each path maps a discriminator value to a transition configuration.
//
// Example: discriminator returns "pass" → fire EventReviewPass → go to FinalizeChecks
type BranchPath struct {
    value       string              // Discriminator value that triggers this path
    event       sdkstate.Event      // Event to fire
    to          sdkstate.State      // Target state
    description string              // Human-readable description

    // Standard transition configuration (forwarded from When options)
    guardTemplate GuardTemplate     // Optional guard (in addition to discriminator)
    onEntry      Action             // OnEntry action
    onExit       Action             // OnExit action
    failedPhase  string             // Phase to mark as failed
}

// BranchOption configures a BranchConfig.
type BranchOption func(*BranchConfig)
```

**Design Rationale**:
- Fields are lowercase (internal only) - users interact via option functions
- `discriminator` separates intent determination from guard validation
- `branches` is a map for O(1) lookup during event determination
- BranchPath includes all TransitionConfig fields for full transition support

### 2. Implement BranchOn Option Function

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch.go`

```go
// BranchOn specifies the discriminator function for a branch configuration.
//
// The discriminator examines project state and returns a string value that
// determines which branch path to take. This value is matched against the
// values defined in When() clauses.
//
// The discriminator is called during Advance() to automatically determine
// which event to fire. It should be a pure function that examines project
// state but does not modify it.
//
// Example:
//   BranchOn(func(p *state.Project) string {
//       // Get review assessment from latest approved review artifact
//       phase := p.Phases["review"]
//       for i := len(phase.Outputs) - 1; i >= 0; i-- {
//           artifact := phase.Outputs[i]
//           if artifact.Type == "review" && artifact.Approved {
//               if assessment, ok := artifact.Metadata["assessment"].(string); ok {
//                   return assessment  // "pass" or "fail"
//               }
//           }
//       }
//       return ""  // No approved review yet
//   })
//
func BranchOn(discriminator func(*state.Project) string) BranchOption {
    return func(bc *BranchConfig) {
        bc.discriminator = discriminator
    }
}
```

### 3. Implement When Option Function

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch.go`

```go
// When defines a branch path based on a discriminator value.
//
// Each When clause creates one possible branch destination. When the discriminator
// returns the specified value, the corresponding event is fired and the state
// machine transitions to the target state.
//
// Standard transition options (WithGuard, WithOnEntry, WithDescription, etc.) can
// be passed to configure the generated transition.
//
// Parameters:
//   value - Discriminator value to match (e.g., "pass", "fail", "staging")
//   event - Event to fire when this branch is taken
//   to - Target state for this branch
//   opts - Standard TransitionOption functions
//
// Example:
//   When("pass",
//       sdkstate.Event(EventReviewPass),
//       sdkstate.State(FinalizeChecks),
//       WithGuard("review passed", func(p *state.Project) bool {
//           return getReviewAssessment(p) == "pass"
//       }),
//       WithDescription("Review approved - proceed to finalization"),
//   )
//
func When(
    value string,
    event sdkstate.Event,
    to sdkstate.State,
    opts ...TransitionOption,
) BranchOption {
    return func(bc *BranchConfig) {
        // Create BranchPath
        path := &BranchPath{
            value: value,
            event: event,
            to:    to,
        }

        // Apply transition options to extract configuration
        // This is a temporary TransitionConfig used to capture option values
        tc := TransitionConfig{}
        for _, opt := range opts {
            opt(&tc)
        }

        // Copy transition config fields to branch path
        path.description = tc.description
        path.guardTemplate = tc.guardTemplate
        path.onEntry = tc.onEntry
        path.onExit = tc.onExit
        path.failedPhase = tc.failedPhase

        // Initialize branches map if needed
        if bc.branches == nil {
            bc.branches = make(map[string]*BranchPath)
        }

        // Store branch path (last one wins if duplicate value)
        bc.branches[value] = path
    }
}
```

**Design Note**: `When()` accepts `TransitionOption` functions and applies them to a temporary `TransitionConfig` to extract the values, then copies them to `BranchPath`. This allows reusing all existing transition options without duplicating code.

### 4. Test Coverage

**File**: `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/branch_test.go` (NEW FILE)

Create comprehensive unit tests following TDD:

```go
package project

import (
    "testing"

    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBranchOn(t *testing.T) {
    t.Run("sets discriminator function", func(t *testing.T) {
        // Test that BranchOn stores the discriminator in BranchConfig
    })

    t.Run("discriminator is called correctly", func(t *testing.T) {
        // Test that the discriminator function can examine project state
    })
}

func TestWhen(t *testing.T) {
    t.Run("creates branch path with value, event, and target state", func(t *testing.T) {
        // Test basic When clause creates BranchPath
    })

    t.Run("stores path in branches map", func(t *testing.T) {
        // Test that When adds to bc.branches
    })

    t.Run("initializes branches map if nil", func(t *testing.T) {
        // Test that When creates the map if it doesn't exist
    })

    t.Run("forwards WithDescription option", func(t *testing.T) {
        // Test that transition options are applied to BranchPath
    })

    t.Run("forwards WithGuard option", func(t *testing.T) {
        // Test guard forwarding
    })

    t.Run("forwards WithOnEntry option", func(t *testing.T) {
        // Test onEntry forwarding
    })

    t.Run("forwards WithOnExit option", func(t *testing.T) {
        // Test onExit forwarding
    })

    t.Run("forwards WithFailedPhase option", func(t *testing.T) {
        // Test failedPhase forwarding
    })

    t.Run("multiple When clauses accumulate", func(t *testing.T) {
        // Test that multiple When calls add multiple paths
    })

    t.Run("duplicate value overwrites previous path", func(t *testing.T) {
        // Test last-one-wins behavior
    })
}

func TestBranchConfigIntegration(t *testing.T) {
    t.Run("BranchOn and When work together", func(t *testing.T) {
        // Test complete branch config creation
        bc := &BranchConfig{}

        discriminator := func(p *state.Project) string {
            return "test_value"
        }
        BranchOn(discriminator)(bc)

        When("test_value",
            sdkstate.Event("test_event"),
            sdkstate.State("test_state"),
            WithDescription("Test branch"),
        )(bc)

        require.NotNil(t, bc.discriminator)
        require.NotNil(t, bc.branches)
        assert.Len(t, bc.branches, 1)

        path := bc.branches["test_value"]
        require.NotNil(t, path)
        assert.Equal(t, "test_value", path.value)
        assert.Equal(t, sdkstate.Event("test_event"), path.event)
        assert.Equal(t, sdkstate.State("test_state"), path.to)
        assert.Equal(t, "Test branch", path.description)
    })
}
```

## Acceptance Criteria

- [ ] New file `branch.go` created in cli/internal/sdks/project/
- [ ] `BranchConfig` struct defined with all required fields
- [ ] `BranchPath` struct defined with all required fields
- [ ] `BranchOption` type defined
- [ ] `BranchOn()` function implemented and documented
- [ ] `When()` function implemented and documented
- [ ] `When()` correctly forwards all TransitionOption types:
  - WithDescription
  - WithGuard
  - WithOnEntry
  - WithOnExit
  - WithFailedPhase
- [ ] Unit tests pass for `TestBranchOn` covering:
  - Setting discriminator function
  - Discriminator can access project state
- [ ] Unit tests pass for `TestWhen` covering:
  - Basic branch path creation
  - Storage in branches map
  - Map initialization
  - All transition option forwarding
  - Multiple When clauses
  - Duplicate value handling
- [ ] Integration test shows BranchOn and When working together
- [ ] Code follows existing SDK patterns and style
- [ ] Godoc comments are clear and include examples

## Technical Details

### Option Function Pattern

This follows the SDK's established pattern:

```go
// Phase options
type PhaseOpt func(*PhaseConfig)
func WithStartState(state) PhaseOpt { ... }

// Transition options
type TransitionOption func(*TransitionConfig)
func WithGuard(desc, fn) TransitionOption { ... }

// Branch options (NEW)
type BranchOption func(*BranchConfig)
func BranchOn(discriminator) BranchOption { ... }
func When(value, event, to, opts...) BranchOption { ... }
```

### Why Map Instead of Slice?

`branches map[string]*BranchPath` provides:
- O(1) lookup during event determination
- Clear mapping from discriminator values to paths
- Automatic handling of duplicate values (last-one-wins)

### Discriminator Function Signature

The discriminator is `func(*state.Project) string` because:
- Examines project state to make decision
- Returns string for human-readable values ("pass", "fail", "staging")
- Can be easily debugged and tested
- Matches the pattern used in guards and actions

### TransitionOption Reuse

`When()` accepts `...TransitionOption` to:
- Reuse existing option functions (WithGuard, WithDescription, etc.)
- Avoid duplicating option implementations
- Maintain consistency with AddTransition API
- Support all transition features (guards, actions, descriptions)

The implementation applies options to a temporary `TransitionConfig`, then copies the fields to `BranchPath`. This is a clean way to extract option values without exposing TransitionConfig internals.

## Relevant Inputs

### Core SDK Files (Read for Patterns)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/types.go` - See GuardTemplate, Action, TransitionConfig for field types
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/options.go` - See existing option functions for pattern reference
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/config.go` - Understanding of how configs are structured

### Test Files (Pattern Reference)

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/builder_test.go` - Testing patterns for builder methods
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/cli/internal/sdks/project/integration_test.go` - Integration test patterns

### Documentation

- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/project/context/issue-77.md` - Complete requirements (section 3-4)
- `/Users/josh/code/sow/.sow/worktrees/77-sdk-branching-support-addbranch-api/.sow/knowledge/designs/sdk-addbranch-api.md` - Data structures specification (section 6)

## Examples

### Binary Branch Configuration

```go
bc := &BranchConfig{}

// Set discriminator
BranchOn(func(p *state.Project) string {
    phase := p.Phases["review"]
    for i := len(phase.Outputs) - 1; i >= 0; i-- {
        if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
            return phase.Outputs[i].Metadata["assessment"].(string)
        }
    }
    return ""
})(bc)

// Define pass branch
When("pass",
    sdkstate.Event(EventReviewPass),
    sdkstate.State(FinalizeChecks),
    WithDescription("Review approved - proceed to finalization"),
)(bc)

// Define fail branch
When("fail",
    sdkstate.Event(EventReviewFail),
    sdkstate.State(ImplementationPlanning),
    WithDescription("Review failed - return to planning for rework"),
    WithFailedPhase("review"),
)(bc)

// Result: bc.branches has 2 paths, bc.discriminator is set
```

### N-Way Branch Configuration

```go
bc := &BranchConfig{}

BranchOn(func(p *state.Project) string {
    return p.Phases["deployment"].Metadata["target"].(string)
})(bc)

When("staging", EventDeployStaging, DeployingStaging,
    WithDescription("Deploy to staging environment"))(bc)

When("production", EventDeployProduction, DeployingProduction,
    WithGuard("all tests passed", allTestsPassed),
    WithDescription("Deploy to production"))(bc)

When("canary", EventDeployCanary, DeployingCanary,
    WithDescription("Deploy to canary environment"))(bc)

// Result: bc.branches has 3 paths
```

## Dependencies

- Task 010 (Transition Descriptions) must be complete
  - `When()` uses `WithDescription()` option
  - BranchPath has description field

## Constraints

### Type Safety

- Use exported SDK types (sdkstate.State, sdkstate.Event)
- Internal fields (lowercase) to prevent direct manipulation
- Option functions are the only public API

### Memory Efficiency

- Use pointers for BranchPath to avoid copying large structs
- Map provides efficient lookup
- All configuration happens at build time (not runtime)

### Error Handling

- Option functions don't return errors (consistent with SDK pattern)
- Validation happens later in AddBranch
- Invalid configurations detected during Build()

## Success Criteria

This task is complete when:

1. `BranchConfig` and `BranchPath` types are defined
2. `BranchOn()` and `When()` option functions work correctly
3. All unit tests pass
4. Option forwarding works for all TransitionOption types
5. Code is well-documented with examples
6. No breaking changes to existing functionality

The types created here will be used by Task 030 (AddBranch builder method) to auto-generate transitions and event determiners.
