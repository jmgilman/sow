# Issue #77: SDK Branching Support (AddBranch API)

**URL**: https://github.com/jmgilman/sow/issues/77
**State**: OPEN

## Description

# Work Unit 002: SDK Branching Support (AddBranch API)

**Status**: Ready for Implementation
**Estimated Duration**: 4-5 days
**Implementation Approach**: Test-Driven Development (TDD)

---

## 1. Behavioral Goal

**As a** project type developer,
**I need** a declarative API for configuring state-determined branching in project state machines,
**So that** branching logic is explicit, discoverable, and self-documenting rather than hidden in OnAdvance determiners.

### Success Criteria for Reviewers

When this work unit is complete, reviewers will be able to:

1. Define binary branches (e.g., review pass/fail) using `AddBranch()` with `BranchOn()` and `When()`
2. Define N-way branches (3+ paths) using multiple `When()` clauses
3. Retrieve transition descriptions and target states through introspection methods
4. Identify branching states via `IsBranchingState()`
5. Observe that the discriminator function automatically determines which event fires
6. Verify that guards can be added to individual branch paths for additional validation
7. Confirm that the system errors meaningfully when a discriminator returns an unmapped value

---

## 2. Existing Code Context

### Integration Overview

This work unit extends the existing Project SDK's fluent builder pattern (introduced in ADR-004) to support declarative branching. The current SDK provides `AddTransition()` for defining individual state transitions and `OnAdvance()` for event determination, but these require workarounds for branching (as seen in the standard project's ReviewActive state).

The `AddBranch()` API builds on top of these existing primitives by auto-generating both transitions and event determiners from a high-level declarative configuration. It captures a discriminator function (which examines project state) and branch paths (which define target states and options), then generates the appropriate `AddTransition()` and `OnAdvance()` calls internally.

This approach follows the SDK's established patterns:
- **Fluent builder methods** that return `*ProjectTypeConfigBuilder` for chaining
- **Option functions** for configuring transitions (guards, actions, descriptions)
- **Closure binding** in `BuildMachine()` to bind template functions to project instances
- **Separation between builder and config** where the builder accumulates configuration and `Build()` copies data to an immutable config

The introspection methods integrate with the CLI's advance command (to be implemented in work unit 003), enabling discovery of available transitions through `GetAvailableTransitions()` and related methods.

### Key Files to Extend

**Core SDK Files**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/builder.go:13-171`
  - ProjectTypeConfigBuilder struct (lines 13-22): Add `branches` field
  - Build() method (lines 137-170): Copy branches map to config
  - New method location: `AddBranch()` around line 172+

- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/config.go:63-92`
  - ProjectTypeConfig struct (lines 63-92): Add `branches` field
  - New methods section: Introspection methods (GetAvailableTransitions, etc.)

- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/types.go:1-29`
  - Add TransitionConfig.description field
  - New file to create: `branch.go` for BranchConfig, BranchPath types

- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/options.go:54-106`
  - TransitionOption type (line 54): Already defined
  - New option: `WithDescription()` around line 107+

**State Machine Integration**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/machine.go:43-105`
  - BuildMachine() method: Already handles closure binding for guards/actions
  - No modifications needed (AddBranch generates standard transitions)

**Testing Patterns**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/builder_test.go:1-429`
  - Established pattern: One test function per builder method with sub-tests
  - Example: TestAddTransition (tests chainability, options, multiple calls)

- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/project/integration_test.go:1-100+`
  - Pattern: End-to-end workflow tests creating configs, building machines, firing events

**Reference Implementation (Current Workaround)**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/standard.go:118-241`
  - ReviewActive transitions (lines 118-165): Two transitions with identical guards
  - ReviewActive OnAdvance (lines 207-241): Real branching logic in discriminator
  - This is the pattern AddBranch should replace

---

## 3. Existing Documentation Context

### ADR-003: State Machine Branching Support

This ADR (`.sow/knowledge/adrs/003-state-machine-branching.md`) establishes the architectural decision to support **state-determined branching** as a first-class concept in the SDK. It identifies two types of branching:

1. **State-determined**: Decision discoverable from project state (e.g., review assessment already in metadata) - addressed by this work unit
2. **Intent-based**: Decision requires orchestrator choice (e.g., "add more research" vs "finalize") - handled by multiple `AddTransition()` calls

The ADR chose Option 1 (AddBranch API with Discriminator) over alternatives like enhanced OnAdvance or DSL approaches because it provides:
- Declarative configuration that's explicit and visible
- Type-safe Go functions with compile-time checking
- Discoverability through introspection methods
- Backward compatibility (opt-in feature)

**Section 9 (Implementation Details)** provides the exact method signatures and auto-generation logic that this work unit must implement. The decision to auto-generate guards from discriminator values was rejected (section 11.2) - discriminators determine intent while guards validate preconditions, serving different purposes.

### SDK AddBranch API Design Document

The complete API specification (`.sow/knowledge/designs/sdk-addbranch-api.md`) provides:

**Section 5 (API Specification)**: Exact signatures for `AddBranch()`, `BranchOn()`, `When()`, and `WithDescription()` - these are the contracts to implement.

**Section 6 (Internal Architecture)**: Defines the data structures (`BranchConfig`, `BranchPath`) and auto-generation logic. The builder must:
1. Create BranchConfig from options
2. Generate `AddTransition()` calls for each `When()` clause
3. Generate `OnAdvance()` determiner that calls discriminator and maps values to events
4. Store BranchConfig for introspection

**Section 7 (Usage Examples)**: Shows binary branching (review pass/fail), N-way branching (deployment targets), and mixing with intent-based branching. Example 1 demonstrates the exact closure pattern for discriminators that access phase outputs and metadata.

**Section 11 (Open Questions - Resolved)**:
- Error handling: Return error with list of available values when no match (section 11.1)
- Guards are explicit via `WithGuard()` option, not auto-generated (section 11.2)
- Descriptions are optional but recommended (section 11.4)

### ADR-004: Project SDK Architecture

This ADR (`.sow/knowledge/adrs/004-introduce-project-sdk-architecture.md`) established the SDK's architectural patterns that AddBranch must follow:

- **Builder pattern** with fluent chaining (all methods return `*Builder`)
- **Option functions** for configuration (PhaseOpt, TransitionOption, etc.)
- **Closure binding** in BuildMachine() to bind template functions to project instances
- **CUE validation** maintained for state files (not relevant to this work unit)

The SDK was designed to be extensible through new builder methods and options - AddBranch is exactly this kind of extension.

---

## 4. Implementation Approach

### Recommended Order (TDD Progression)

This work unit has 5 sub-components that build on each other. Implement in this order to enable continuous testing:

#### Phase 1: WithDescription for Transitions (Foundation)
**Duration**: 0.5 days

Start here because all other components depend on transition descriptions:
1. Write test: `TestWithDescription` - verify description stored in TransitionConfig
2. Add `description string` field to TransitionConfig
3. Implement `WithDescription()` option function
4. Verify test passes

**Why first**: Descriptions are used by branch paths and introspection methods, so establish this foundation early.

#### Phase 2: Branch Types (Data Structures)
**Duration**: 1 day

Define the data structures before the logic that uses them:
1. Write tests: `TestBranchOn`, `TestWhen` - verify option functions create correct configs
2. Create `branch.go` with `BranchConfig` and `BranchPath` structs
3. Implement `BranchOn()` option function (captures discriminator)
4. Implement `When()` option function (creates BranchPath, forwards TransitionOptions)
5. Verify tests pass

**Why second**: Types must exist before AddBranch can use them, but they're testable independently.

#### Phase 3: AddBranch Builder Method (Core Logic)
**Duration**: 1.5-2 days

Implement the auto-generation logic:
1. Write tests:
   - `TestAddBranchGeneratesTransitions` - verify AddTransition calls
   - `TestAddBranchGeneratesOnAdvance` - verify event determiner created
   - `TestAddBranchBinary` - binary branching workflow
   - `TestAddBranchNWay` - N-way branching workflow
2. Add `branches map[State]*BranchConfig` to ProjectTypeConfigBuilder
3. Implement `AddBranch()` method:
   - Create BranchConfig from options
   - Iterate BranchPaths, generate AddTransition() calls
   - Generate OnAdvance() determiner
   - Store BranchConfig in builder
4. Update `Build()` to copy branches map
5. Add `branches` field to ProjectTypeConfig
6. Verify tests pass

**Why third**: This is the core functionality that users interact with, built on phases 1-2.

#### Phase 4: Error Cases (Robustness)
**Duration**: 0.5 days

Test error handling:
1. Write tests:
   - `TestDiscriminatorNoMatch` - error when discriminator returns unknown value
   - `TestAddBranchNoDiscriminator` - error when BranchOn not provided
   - `TestAddBranchNoBranches` - error when no When clauses provided
2. Add validation logic to AddBranch
3. Enhance OnAdvance error message to list available values
4. Verify tests pass

**Why fourth**: Error cases are easier to test after happy path works.

#### Phase 5: Introspection Methods (Discoverability)
**Duration**: 1 day

Enable CLI integration:
1. Write tests: `TestIntrospectionMethods` covering all 5 methods
2. Define `TransitionInfo` struct
3. Implement introspection methods in ProjectTypeConfig:
   - `GetAvailableTransitions(from State) []TransitionInfo`
   - `GetTransitionDescription(from, event) string`
   - `GetTargetState(from, event) State`
   - `GetGuardDescription(from, event) string`
   - `IsBranchingState(state) bool`
4. Verify tests pass

**Why fifth**: Introspection depends on branches being stored by AddBranch.

### Integration Testing Strategy

After all phases, add integration tests in `integration_test.go`:
- `TestReviewBranchingWorkflow` - Complete review pass/fail scenario
- `TestMixedBranchingAndLinear` - AddBranch + AddTransition in same config
- `TestBranchIntrospection` - Verify introspection returns correct info

---

## 5. Testing Strategy

### Unit Tests

**File**: `cli/internal/sdks/project/branch_test.go` (new)

```go
// Test option functions
func TestBranchOn(t *testing.T)
func TestWhen(t *testing.T)

// Test auto-generation
func TestAddBranchGeneratesTransitions(t *testing.T)
func TestAddBranchGeneratesOnAdvance(t *testing.T)

// Test workflows
func TestAddBranchBinary(t *testing.T)  // 2 branches
func TestAddBranchNWay(t *testing.T)    // 3+ branches

// Test error cases
func TestDiscriminatorNoMatch(t *testing.T)
func TestAddBranchNoDiscriminator(t *testing.T)
func TestAddBranchNoBranches(t *testing.T)
```

**File**: `cli/internal/sdks/project/config_test.go` (existing, add tests)

```go
// Test introspection methods
func TestGetAvailableTransitions(t *testing.T)
func TestGetTransitionDescription(t *testing.T)
func TestGetTargetState(t *testing.T)
func TestGetGuardDescription(t *testing.T)
func TestIsBranchingState(t *testing.T)
```

**File**: `cli/internal/sdks/project/options_test.go` (existing, add test)

```go
func TestWithDescription(t *testing.T)
```

### Integration Tests

**File**: `cli/internal/sdks/project/integration_test.go` (existing, add tests)

```go
// Test complete review branching scenario
func TestReviewBranchingWorkflow(t *testing.T) {
    // Create config with AddBranch for ReviewActive
    // Build machine with test project
    // Set review assessment to "pass"
    // Fire advance event
    // Assert state changed to FinalizeChecks

    // Reset, set assessment to "fail"
    // Fire advance event
    // Assert state changed to ImplementationPlanning
}

// Test mixing branching and linear states
func TestMixedBranchingAndLinear(t *testing.T)

// Test branch introspection returns correct info
func TestBranchIntrospection(t *testing.T)
```

### Test Coverage Goals

- All public methods have unit tests
- All option functions have unit tests
- Binary and N-way branching have integration tests
- Error cases have explicit tests
- Introspection methods tested for both branching and non-branching states

---

## 6. Dependencies

**None** - This is the foundation work unit.

Work unit 003 (CLI Enhanced Advance Command) depends on the introspection methods implemented here, but this work unit has no dependencies on other work.

---

## 7. Acceptance Criteria

Reviewers will verify:

- [ ] Binary branches (2 paths) can be configured using `AddBranch()` with `BranchOn()` and `When()`
- [ ] N-way branches (3+ paths) can be configured using multiple `When()` clauses
- [ ] Discriminator function is called during advance and automatically determines which event fires
- [ ] Guards can be added to individual branch paths via `When()` with `WithGuard()` option
- [ ] Descriptions can be added to transitions via `WithDescription()` option
- [ ] Descriptions are stored in TransitionConfig and retrievable via `GetTransitionDescription()`
- [ ] `GetAvailableTransitions()` returns all transitions from a state with descriptions and target states
- [ ] `GetTargetState()` returns correct target state for a from-state/event pair
- [ ] `GetGuardDescription()` returns guard description if guard exists
- [ ] `IsBranchingState()` returns true for states with AddBranch configuration, false otherwise
- [ ] Error is returned when discriminator returns value with no matching `When()` clause
- [ ] Error message includes list of available discriminator values
- [ ] All unit tests pass (branch options, AddBranch logic, introspection)
- [ ] All integration tests pass (binary branching, N-way branching, mixed workflows)
- [ ] Code follows existing SDK patterns (builder chaining, option functions, closure binding)
- [ ] No breaking changes to existing SDK functionality

---

## 8. Code Examples

### Example 1: Basic AddBranch Usage (Binary Branch)

```go
// Define a binary branch for review outcomes
builder.AddBranch(
    sdkstate.State(ReviewActive),
    project.BranchOn(func(p *state.Project) string {
        // Discriminator examines project state
        phase := p.Phases["review"]
        for i := len(phase.Outputs) - 1; i >= 0; i-- {
            artifact := phase.Outputs[i]
            if artifact.Type == "review" && artifact.Approved {
                return artifact.Metadata["assessment"].(string) // "pass" or "fail"
            }
        }
        return "" // No approved review yet
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
        project.WithOnEntry(setupReworkAction),
    ),
)
```

This auto-generates:
1. Two `AddTransition()` calls (one per `When()` clause)
2. One `OnAdvance()` determiner that calls the discriminator and returns the matching event

### Example 2: Introspection Usage

```go
// Get all available transitions from a state
transitions := config.GetAvailableTransitions(sdkstate.State(ReviewActive))
for _, t := range transitions {
    fmt.Printf("Event: %s → %s (%s)\n", t.Event, t.To, t.Description)
}
// Output:
// Event: review_pass → finalize_checks (Review approved - proceed to finalization)
// Event: review_fail → implementation_planning (Review failed - return to planning for rework)

// Check if state has branches
if config.IsBranchingState(sdkstate.State(ReviewActive)) {
    fmt.Println("This is a branching state")
}

// Get target state for a transition
target := config.GetTargetState(
    sdkstate.State(ReviewActive),
    sdkstate.Event(EventReviewPass),
)
// Returns: finalize_checks
```

---

## 9. Implementation Notes

### Backward Compatibility

- All existing project types continue working unchanged
- `AddBranch()` is opt-in - coexists with `AddTransition()` and manual `OnAdvance()`
- No changes to state machine wrapper or underlying stateless library
- No breaking changes to ProjectTypeConfig public API

### Patterns to Follow

**Builder Pattern Consistency**:
- `AddBranch()` must return `*ProjectTypeConfigBuilder` for chaining
- Store configuration in builder, copy to config in `Build()`
- Apply option functions in order received

**Option Function Pattern**:
```go
type BranchOption func(*BranchConfig)

func BranchOn(discriminator func(*state.Project) string) BranchOption {
    return func(bc *BranchConfig) {
        bc.discriminator = discriminator
    }
}
```

**Closure Binding Pattern** (already handled by BuildMachine):
- Discriminators are `func(*state.Project) string` (template functions)
- BuildMachine binds them to project instance via closure
- No changes needed to BuildMachine - it already handles OnAdvance closures

### Edge Cases to Handle

1. **No discriminator provided**: Validate in AddBranch that BranchOn was called
2. **No When clauses provided**: Validate at least one branch path exists
3. **Discriminator returns unmapped value**: Error in OnAdvance with list of available values
4. **Duplicate When values**: Last one wins (consistent with option function pattern)
5. **AddBranch + OnAdvance on same state**: Conflict - AddBranch generates OnAdvance, so using both should error or last one wins (document this)

### Performance Considerations

- Auto-generation happens once at startup during Build()
- No runtime overhead compared to manual AddTransition + OnAdvance
- Introspection methods iterate transitions (O(n) where n = number of transitions)
- Branch map lookup is O(1) for IsBranchingState

### Testing Considerations

- Use minimal test projects (don't need full phase configurations)
- Mock discriminators with simple functions returning hardcoded values
- Test guards separately from discriminators (they serve different purposes)
- Verify both chainability (returns builder) and correctness (config contains data)

---

## 10. File Locations

### New Files

- `cli/internal/sdks/project/branch.go` - BranchConfig, BranchPath, option functions
- `cli/internal/sdks/project/branch_test.go` - Unit tests for branching

### Modified Files

- `cli/internal/sdks/project/builder.go` - Add AddBranch() method, branches field
- `cli/internal/sdks/project/config.go` - Add branches field, introspection methods
- `cli/internal/sdks/project/types.go` - Add description field to TransitionConfig
- `cli/internal/sdks/project/options.go` - Add WithDescription() option
- `cli/internal/sdks/project/integration_test.go` - Add integration tests
- `cli/internal/sdks/project/config_test.go` - Add introspection method tests
- `cli/internal/sdks/project/options_test.go` - Add WithDescription test

### No Changes Needed

- `cli/internal/sdks/project/machine.go` - Already handles closures correctly
- `cli/internal/sdks/state/` - State machine wrapper unchanged

---

## 11. References

- **Design Document**: `.sow/knowledge/designs/sdk-addbranch-api.md` - Complete API specification
- **Discovery Document**: `.sow/project/discovery/analysis.md` - Codebase analysis with file locations
- **ADR-003**: `.sow/knowledge/adrs/003-state-machine-branching.md` - Branching decision rationale
- **ADR-004**: `.sow/knowledge/adrs/004-introduce-project-sdk-architecture.md` - SDK architecture patterns
- **Reference Implementation**: `cli/internal/projects/standard/standard.go:118-241` - ReviewActive workaround

---

## 12. Success Metrics

This work unit succeeds when:

1. A developer can define state-determined branching declaratively without hidden logic
2. The CLI can introspect available transitions (enables work unit 003)
3. All tests pass including binary, N-way, and error cases
4. Code follows established SDK patterns
5. No breaking changes to existing functionality
6. Documentation in code comments is clear and complete

The ultimate validation will be refactoring the standard project's ReviewActive state to use AddBranch (happens in work unit 003 after CLI integration is complete).
