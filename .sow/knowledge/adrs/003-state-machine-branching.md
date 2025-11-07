# ADR: State Machine Branching Support

**Status**: Proposed
**Date**: 2025-11-06
**Deciders**: Core Team
**Context**: Adding first-class support for multi-way state transitions in sow project state machines

---

## Context and Problem Statement

Sow projects use state machines to model workflows through phases. Currently, the Project SDK assumes **linear state progression**: each state maps to one "advance" event via `OnAdvance()`, and transitions flow in a single direction.

However, real-world workflows often require **branching**: from a single state, multiple possible next states exist depending on conditions or decisions. Examples:

1. **Review outcomes**: After code review, either proceed to finalization (pass) or return to implementation (fail)
2. **Deployment targets**: After build completion, deploy to staging, production, or canary based on configuration
3. **Research continuation**: After research phase, either finalize (complete) or add more topics (incomplete)

**Current workaround** (from standard project's ReviewActive state):
```go
// Two transitions with misleading guards
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass,
    WithGuard("latest review approved", ...))
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail,
    WithGuard("latest review approved", ...))  // Same guard! Misleading.

// Real branching logic hidden in event determiner
OnAdvance(ReviewActive, func(p *Project) (Event, error) {
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass": return EventReviewPass, nil
    case "fail": return EventReviewFail, nil
    }
})
```

**Problems with this approach**:
- Guards are misleading (both say "review approved")
- Branching logic split between guards and event determiner
- Not discoverable (can't introspect branches)
- Doesn't scale to N-way branches
- No descriptions explaining what each branch does

## Decision Drivers

1. **Clarity**: Branching logic should be explicit, not hidden in event determiners
2. **Discoverability**: CLI and tooling should be able to introspect available branches
3. **Maintainability**: Branching configuration should be declarative, not imperative
4. **Scalability**: Support N-way branches (not just binary pass/fail)
5. **Backward Compatibility**: Existing code must continue working unchanged
6. **Flexibility**: Support both state-determined branching (discriminator-based) and intent-based branching (orchestrator choice)
7. **Self-Documentation**: Branch paths should include descriptions for orchestrator guidance

## Two Types of Branching

Through exploration and design work, we identified two distinct branching patterns that require different solutions:

### Type 1: State-Determined Branching

**Characteristic**: The decision can be discovered by examining project state.

**Examples**:
- Review assessment: already set in artifact metadata ("pass" or "fail")
- Test results: already in output artifacts ("passed" or "failed")
- Task status: checkable from task collection ("all_complete", "some_abandoned", "in_progress")

**Solution**: Provide declarative API that:
- Accepts a discriminator function that examines state
- Defines branch paths for each discriminator value
- Auto-generates transitions and event determiner

### Type 2: Intent-Based Branching

**Characteristic**: The decision requires external orchestrator/user judgment.

**Examples**:
- "Add more research" vs "Finalize research" (depends on orchestrator's assessment of completeness)
- "Deploy now" vs "Skip deployment" (depends on user choice)
- "Retry with changes" vs "Abandon feature" (depends on strategic decision)

**Solution**: Use existing `AddTransition()` for multiple paths + allow explicit event selection in CLI (see ADR: Explicit Event Selection).

## Considered Options

### Option 1: AddBranch API with Discriminator

Introduce declarative `AddBranch()` API for state-determined branching:

```go
AddBranch(ReviewActive,
    BranchOn(func(p *Project) string {
        // Examine state, return discriminator value
        return getReviewAssessment(p)  // "pass" or "fail"
    }),
    When("pass", EventReviewPass, FinalizeChecks,
        WithDescription("Review approved - proceed to finalization")),
    When("fail", EventReviewFail, ImplementationPlanning,
        WithDescription("Review failed - return to planning for rework")),
)
```

**Mechanism**:
- `BranchOn()` specifies discriminator function
- `When()` defines each branch path (value, event, target state, options)
- Builder auto-generates `AddTransition()` calls for each branch
- Builder auto-generates `OnAdvance()` determiner using discriminator

**Pros**:
- Explicit branching logic (discriminator function visible)
- Declarative (configuration, not imperative code)
- Discoverable (introspection methods can query branches)
- Self-documenting (descriptions co-located with branches)
- Scales to N-way branches
- DRY (discriminator written once, not duplicated in guards)
- Backward compatible (opt-in feature)

**Cons**:
- New API to learn
- Slight complexity increase
- Only works for state-determined branching (not intent-based)

### Option 2: Enhanced OnAdvance with Metadata

Allow `OnAdvance()` to return metadata about available branches:

```go
OnAdvance(ReviewActive, func(p *Project) (Event, BranchMetadata, error) {
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass":
        return EventReviewPass, BranchMetadata{
            Options: []BranchOption{
                {Event: EventReviewPass, To: FinalizeChecks, Desc: "Proceed"},
                {Event: EventReviewFail, To: ImplementationPlanning, Desc: "Rework"},
            },
        }, nil
    }
})
```

**Pros**:
- Builds on existing OnAdvance pattern
- Discoverable (metadata available)

**Cons**:
- Still imperative (switch logic in code)
- Branching logic still hidden in function
- Metadata and logic can drift apart
- More complex function signatures
- Harder to validate at build time

### Option 3: State Machine Definition Language

Introduce DSL or config files for state machine definition:

```yaml
states:
  ReviewActive:
    branches:
      discriminator: getReviewAssessment
      paths:
        pass:
          event: review_pass
          to: FinalizeChecks
          description: "Review approved"
        fail:
          event: review_fail
          to: ImplementationPlanning
          description: "Review failed"
```

**Pros**:
- Very declarative
- External to code (easier to visualize)
- Could generate documentation automatically

**Cons**:
- Major paradigm shift (current SDK is code-based)
- Loses type safety
- Harder to reference Go functions (discriminators, guards, actions)
- More tooling complexity
- Breaks existing patterns

### Option 4: Keep Current Workaround

Continue using `AddTransition()` + `OnAdvance()` workaround.

**Pros**:
- No new APIs
- Works today

**Cons**:
- Guards remain misleading
- Logic remains hidden
- Not discoverable
- Not self-documenting
- Doesn't scale well
- Violates design principles (clarity, discoverability)

## Decision Outcome

**Chosen option: Option 1 (AddBranch API with Discriminator)** for state-determined branching, combined with existing `AddTransition()` pattern for intent-based branching.

### Reasoning

1. **Addresses the two branching types appropriately**:
   - `AddBranch()` for state-determined branching (discriminator can examine state)
   - Multiple `AddTransition()` + explicit events for intent-based branching (orchestrator chooses)

2. **Declarative and explicit**: Branching logic visible in builder configuration

3. **Discoverable**: Introspection methods enable CLI `--list` to show available transitions

4. **Backward compatible**: Existing code unchanged, `AddBranch()` is opt-in

5. **Self-documenting**: Descriptions co-located with branch definitions

6. **Scales**: Supports N-way branches naturally

7. **Maintains consistency**: Uses existing SDK patterns (builder, options, transitions)

8. **Type-safe**: Discriminators and guards are Go functions with compile-time checking

### Why Not the Alternatives?

**Option 2 (Enhanced OnAdvance)**: Still imperative, logic hidden in functions, harder to validate.

**Option 3 (DSL)**: Too large a paradigm shift, loses type safety, increases tooling complexity.

**Option 4 (Status quo)**: Doesn't solve the problems, violates design principles.

## Implementation Details

### Core API

```go
// New builder method
func (b *ProjectTypeConfigBuilder) AddBranch(
    from State,
    opts ...BranchOption,
) *ProjectTypeConfigBuilder

// Configuration options
func BranchOn(discriminator func(*Project) string) BranchOption

func When(
    value string,           // Discriminator value to match
    event Event,            // Event to fire
    to State,              // Target state
    opts ...TransitionOption, // Guards, actions, descriptions
) BranchOption
```

### Auto-Generation

When `AddBranch()` is called:

1. **Generate transitions**: For each `When()` clause, call `AddTransition()` with appropriate options
2. **Generate event determiner**: Create `OnAdvance()` that calls discriminator and maps to events
3. **Store metadata**: Enable introspection via `GetAvailableTransitions()`, etc.

### Example: Review Branching

**Before** (workaround):
```go
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass, ...)
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail, ...)
OnAdvance(ReviewActive, func(p) { /* hidden switch logic */ })
```

**After** (with AddBranch):
```go
AddBranch(ReviewActive,
    BranchOn(func(p *Project) string {
        phase := p.Phases["review"]
        for i := len(phase.Outputs) - 1; i >= 0; i-- {
            if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
                return phase.Outputs[i].Metadata["assessment"].(string)
            }
        }
        return ""
    }),
    When("pass", EventReviewPass, FinalizeChecks,
        WithGuard("review passed", func(p) bool {
            return getReviewAssessment(p) == "pass"
        }),
        WithDescription("Review approved - proceed to finalization"),
    ),
    When("fail", EventReviewFail, ImplementationPlanning,
        WithGuard("review failed", func(p) bool {
            return getReviewAssessment(p) == "fail"
        }),
        WithFailedPhase("review"),
        WithDescription("Review failed - return to planning for rework"),
        WithOnEntry(setupRework),
    ),
)
```

### Example: N-Way Branching

Deployment target selection (hypothetical):

```go
AddBranch(DeploymentReady,
    BranchOn(func(p *Project) string {
        return p.PhaseMetadata("deployment", "target").(string)
    }),
    When("staging", EventDeployStaging, DeployingStaging,
        WithDescription("Deploy to staging environment")),
    When("production", EventDeployProduction, DeployingProduction,
        WithGuard("all tests passed", allTestsPassed),
        WithDescription("Deploy to production environment")),
    When("canary", EventDeployCanary, DeployingCanary,
        WithGuard("canary enabled", canaryEnabled),
        WithDescription("Deploy to canary environment")),
)
```

## Consequences

### Positive

- **Explicit branching**: Logic visible in configuration, not hidden in functions
- **Discoverable**: CLI can introspect branches via SDK methods
- **Self-documenting**: Descriptions explain what each branch does
- **Maintainable**: Declarative configuration easier to understand and modify
- **Scalable**: N-way branches natural, not limited to binary
- **Backward compatible**: Existing projects continue working
- **Flexible**: Two-pattern approach (state-determined + intent-based) covers all use cases
- **Type-safe**: Discriminators and guards are compiled Go code

### Negative

- **Learning curve**: New API to understand (BranchOn, When)
- **Code verbosity**: AddBranch calls are longer than OnAdvance
  - *Mitigation*: Clarity and maintainability worth the verbosity
- **Two patterns**: Developers must choose between AddBranch and multiple AddTransition
  - *Mitigation*: Clear documentation on when to use each

### Neutral

- **No performance impact**: Auto-generation happens at startup, runtime same as before
- **Additional SDK code**: New types and methods (~300-500 lines)
- **Documentation updates**: SDK docs, orchestrator prompts need updates

## Migration Path

### Phase 1: Add AddBranch API

- Implement `AddBranch()`, `BranchOn()`, `When()`
- Add introspection methods
- Write tests
- Document usage

### Phase 2: Update Standard Project

- Replace ReviewActive workaround with AddBranch
- Verify functionality unchanged

### Phase 3: Documentation

- Update SDK documentation
- Add examples and best practices
- Update orchestrator prompts

### Backward Compatibility

- All existing code continues working unchanged
- `AddBranch()` is opt-in
- Can mix AddBranch and traditional AddTransition in same project type
- No breaking changes to state machine behavior

## Validation

### Success Criteria

- [ ] `AddBranch()` auto-generates transitions and event determiners correctly
- [ ] Introspection methods return accurate branch information
- [ ] Standard project ReviewActive uses AddBranch
- [ ] N-way branching (3+ branches) works correctly
- [ ] Error handling (no matching discriminator value) provides helpful messages
- [ ] CLI `--list` shows branch descriptions
- [ ] Orchestrators can successfully navigate branching states
- [ ] No performance regression in state machine operations

### Testing Plan

- Unit tests for BranchOn, When, AddBranch
- Integration tests for auto-generated transitions
- Tests for N-way branching
- Tests for error cases (invalid discriminator values)
- Backward compatibility tests (existing projects still work)
- Performance benchmarks (Load/Save/Advance operations)

## Related Decisions

- **ADR: Explicit Event Selection in advance Command** - Covers intent-based branching via CLI
- **SDK Design: AddBranch API** - Technical design specification
- **CLI Design: Enhanced advance Command** - Discovery and validation features

## References

- Exploration document: `.sow/knowledge/explorations/advance-branching/add-branch.md`
- Stateless library: https://github.com/qmuntal/stateless
- Project SDK doc: `cli/internal/sdks/project/state/doc.go`
- Standard project workaround: `cli/internal/projects/standard/standard.go:ReviewActive`

## Open Questions

### 1. Should we support "Otherwise" branch (default case)?

**Question**: What if discriminator returns a value with no matching `When()` clause?

**Current**: Return error from DetermineEvent with list of available values

**Alternative**: Add `Otherwise()` option:
```go
AddBranch(State,
    BranchOn(discriminator),
    When("value1", Event1, State1),
    When("value2", Event2, State2),
    Otherwise(EventDefault, StateDefault),  // Catches all other values
)
```

**Decision**: Defer. Start with explicit error, add Otherwise if pattern emerges.

### 2. Should guards be auto-generated from discriminator?

**Question**: Should each `When()` branch automatically get a guard that checks discriminator value?

**Current**: Discriminator determines intent, optional guards provide additional validation

**Alternative**: Auto-generate guards:
```go
When("pass", EventPass, StateNext)
// Automatically adds: WithGuard("discriminator == 'pass'", func(p) { return discriminator(p) == "pass" })
```

**Decision**: No. Discriminator and guards serve different purposes:
- Discriminator: Determines intent (which branch to take)
- Guard: Validates preconditions (whether safe to proceed)

Separation of concerns is clearer.

---

## Revision History

- **2025-11-06**: Initial proposal
