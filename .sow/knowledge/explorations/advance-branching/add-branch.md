# Design Proposal: Branching State Transitions in Project SDK

**Status:** Proposal (Refined)
**Date:** 2025-11-06
**Updated:** 2025-11-06
**Context:** Supporting multi-way state transitions and backward movement in project state machines

## Problem Statement

The current project SDK assumes linear state progression:
- One state maps to one "advance" event via `OnAdvance()`
- The `sow advance` command is necessarily linear
- Branching exists (standard project's ReviewActive) but uses workarounds
- No first-class support for N-way branches or "backward" transitions

**Current workaround (standard project):**
```go
// Two transitions with identical guards
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass,
    WithGuard("latest review approved", ...))
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail,
    WithGuard("latest review approved", ...))  // Same guard!

// Event determiner does the real branching logic
OnAdvance(ReviewActive, func(p *Project) (Event, error) {
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass": return EventReviewPass, nil
    case "fail": return EventReviewFail, nil
    }
})
```

**Issues:**
- Guards are misleading (both say "review approved")
- Branching logic split between guards and determiner
- Not discoverable (can't introspect available branches)
- Doesn't scale to N-way branches

## Two Types of Branching

Through design discussion, we identified two distinct branching patterns:

### Type 1: State-Determined Branching

**Decision is discoverable from project state:**
- Review assessment: "pass" or "fail" (in artifact metadata)
- Test results: "passed" or "failed" (in output artifacts)
- Validation results: "valid" or "invalid" (checkable from phase data)

**Characteristic:** A function can examine the project and determine which way to go.

**Solution:** `AddBranch()` API (this proposal)

### Type 2: Intent-Based Branching

**Decision is about orchestrator/user intent:**
- "I want to add more research" vs "I'm done researching"
- "Skip this phase" vs "Do this phase"
- "Retry with changes" vs "Abandon"

**Characteristic:** The decision cannot be discovered from state - it must be made externally.

**Solution:** Multiple `AddTransition()` + explicit event selection in CLI

## Proposed Solution: Declarative Branch API

Add optional `AddBranch()` API for **state-determined branching** that coexists with existing `AddTransition()` + `OnAdvance()`.

### When to Use AddBranch

**Use `AddBranch` when:**
- The branching decision can be determined by examining project state
- You can write a discriminator function that returns which branch to take
- The decision is already "made" before calling advance (encoded in artifacts, metadata, etc.)

**Use multiple `AddTransition` when:**
- The branching decision requires external input (orchestrator/user choice)
- The decision cannot be discovered from project state
- The orchestrator needs to explicitly choose between valid options

### API Design

```go
// New builder method
func (b *ProjectTypeConfigBuilder) AddBranch(
    from State,
    opts ...BranchOption,
) *ProjectTypeConfigBuilder

// Branch configuration options
type BranchOption func(*BranchConfig)

// BranchOn specifies the discriminator function
// Returns a string value that determines which When clause matches
func BranchOn(discriminator func(*state.Project) string) BranchOption

// When defines a branch path based on discriminator value
func When(
    value string,           // Discriminator value to match
    event Event,            // Event to fire (for logging/semantics)
    to State,              // Target state (for documentation/validation)
    opts ...TransitionOption, // Standard transition options (guards, actions, descriptions)
) BranchOption
```

**Note on Redundancy:** Both `event` and `to` (target state) are specified in `When()`. This is intentional:
- The event-to-state binding is created during transition registration
- Specifying both provides documentation and validation
- Allows introspection without looking up transition tables
- Makes the branching explicit and self-documenting

### Example Usage

```go
// Standard project's ReviewActive branching (improved)
builder.
    AddBranch(
        state.State(ReviewActive),
        BranchOn(func(p *state.Project) string {
            // Get review assessment from latest approved review
            phase := p.Phases["review"]
            for i := len(phase.Outputs) - 1; i >= 0; i-- {
                if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
                    assessment, _ := phase.Outputs[i].Metadata["assessment"].(string)
                    return assessment  // Returns "pass" or "fail"
                }
            }
            return ""  // No approved review yet
        }),
        When("pass",
            sdkstate.Event(EventReviewPass),
            sdkstate.State(FinalizeChecks),
            project.WithGuard("review passed", func(p *state.Project) bool {
                return reviewAssessment(p) == "pass"
            }),
            project.WithDescription("Review approved - proceed to finalization"),
        ),
        When("fail",
            sdkstate.Event(EventReviewFail),
            sdkstate.State(ImplementationPlanning),
            project.WithGuard("review failed", func(p *state.Project) bool {
                return reviewAssessment(p) == "fail"
            }),
            project.WithFailedPhase("review"),
            project.WithDescription("Review failed - return to planning for rework"),
            project.WithOnEntry(func(p *state.Project) error {
                // Rework logic: increment iteration, add review as input
                return setupRework(p)
            }),
        ),
    )
```

### N-Way Branching Example

```go
// Hypothetical: deployment target selection
builder.
    AddBranch(
        state.State(DeploymentReady),
        BranchOn(func(p *state.Project) string {
            return p.PhaseMetadata("deployment", "target").(string)
        }),
        When("staging",
            sdkstate.Event(EventDeployStaging),
            sdkstate.State(DeployingStaging),
            project.WithDescription("Deploy to staging environment"),
        ),
        When("production",
            sdkstate.Event(EventDeployProduction),
            sdkstate.State(DeployingProduction),
            project.WithGuard("all tests passed", allTestsPassed),
            project.WithDescription("Deploy to production environment"),
        ),
        When("canary",
            sdkstate.Event(EventDeployCanary),
            sdkstate.State(DeployingCanary),
            project.WithGuard("feature flag enabled", canaryEnabled),
            project.WithDescription("Deploy to canary environment"),
        ),
    )
```

### Intent-Based Branching Example (Using AddTransition)

```go
// Exploration project: orchestrator decides whether to continue or add more research
builder.
    AddTransition(
        Researching,
        Finalizing,
        EventFinalize,
        project.WithGuard("all tasks complete", allTasksComplete),
        project.WithDescription("Complete research and move to finalization phase"),
    ).
    AddTransition(
        Researching,
        Planning,
        EventAddMoreResearch,
        project.WithDescription("Return to planning to add more research topics"),
    )

// Orchestrator explicitly chooses via CLI:
// sow advance finalize           (if tasks are done)
// sow advance add_more_research  (if more research needed)
```

### Backward Compatibility

**Existing code continues to work unchanged:**

```go
// Simple linear flow (no changes needed)
builder.
    AddTransition(Active, Summarizing, EventBeginSummarizing,
        project.WithGuard("all tasks resolved", allTasksResolved),
    ).
    OnAdvance(Active, func(_ *state.Project) (sdkstate.Event, error) {
        return EventBeginSummarizing, nil
    })
```

**AddBranch is opt-in:**
- Use `AddBranch` for state-determined branching
- Use `AddTransition` + `OnAdvance` for simple linear states
- Use multiple `AddTransition` for intent-based branching
- All can coexist in the same project type

## New Feature: Transition Descriptions

Add `WithDescription()` option for transitions to improve CLI discoverability:

```go
// New option
func WithDescription(description string) TransitionOption

// Usage
AddTransition(
    Researching,
    Finalizing,
    EventFinalize,
    WithGuard("all tasks complete", allTasksComplete),
    WithDescription("Complete research and move to finalization phase"),
)
```

**Why transition descriptions?**
- Context-specific: same event from different states can have different meanings
- Co-located with guards and actions
- Used by CLI `--list` to show what each transition does
- Helps orchestrators understand available options

## Implementation Strategy

### 1. Add Description to TransitionConfig

```go
type TransitionConfig struct {
    From sdkstate.State
    To   sdkstate.State
    Event sdkstate.Event

    guardTemplate GuardTemplate
    onEntry      Action
    onExit       Action
    failedPhase  string
    description  string  // NEW
}
```

### 2. Add WithDescription Option

```go
func WithDescription(description string) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.description = description
    }
}
```

### 3. Internal Representation

```go
// In ProjectTypeConfig
type BranchConfig struct {
    from          State
    discriminator func(*state.Project) string
    branches      map[string]BranchPath
}

type BranchPath struct {
    value       string  // Discriminator value
    event       Event
    to          State
    description string
    // Standard transition config
    guardTemplate GuardTemplate
    onEntry      Action
    onExit       Action
    failedPhase  string
}

// In ProjectTypeConfig
type ProjectTypeConfig struct {
    // ... existing fields ...
    branches map[State]*BranchConfig
}
```

### 4. Builder Implementation

```go
func (b *ProjectTypeConfigBuilder) AddBranch(
    from State,
    opts ...BranchOption,
) *ProjectTypeConfigBuilder {
    bc := &BranchConfig{
        from:     from,
        branches: make(map[string]BranchPath),
    }

    // Apply options
    for _, opt := range opts {
        opt(bc)
    }

    // Register branch config
    b.branches[from] = bc

    // Generate transitions for each branch path
    for _, path := range bc.branches {
        // Create transition
        transitionOpts := []TransitionOption{}
        if path.guardTemplate.Func != nil {
            transitionOpts = append(transitionOpts,
                WithGuard(path.guardTemplate.Description, path.guardTemplate.Func))
        }
        if path.onEntry != nil {
            transitionOpts = append(transitionOpts, WithOnEntry(path.onEntry))
        }
        if path.onExit != nil {
            transitionOpts = append(transitionOpts, WithOnExit(path.onExit))
        }
        if path.failedPhase != "" {
            transitionOpts = append(transitionOpts, WithFailedPhase(path.failedPhase))
        }
        if path.description != "" {
            transitionOpts = append(transitionOpts, WithDescription(path.description))
        }

        b.AddTransition(from, path.to, path.event, transitionOpts...)
    }

    // Generate event determiner from discriminator
    b.OnAdvance(from, func(p *state.Project) (Event, error) {
        value := bc.discriminator(p)
        path, exists := bc.branches[value]
        if !exists {
            return "", fmt.Errorf("no branch defined for discriminator value %q from state %s",
                value, from)
        }
        return path.event, nil
    })

    return b
}
```

### 5. New ProjectTypeConfig Methods

```go
// GetAvailableTransitions returns all configured transitions from a state
// Uses stateless.PermittedTriggers() to filter by guard status
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from State) []TransitionInfo

// GetTransitionDescription returns human-readable description
func (ptc *ProjectTypeConfig) GetTransitionDescription(from State, event Event) string

// GetTargetState returns the target state for a transition
func (ptc *ProjectTypeConfig) GetTargetState(from State, event Event) State

// IsBranchingState checks if a state has branches configured
func (ptc *ProjectTypeConfig) IsBranchingState(state State) bool
```

### 6. Option Implementations

```go
func BranchOn(discriminator func(*state.Project) string) BranchOption {
    return func(bc *BranchConfig) {
        bc.discriminator = discriminator
    }
}

func When(
    value string,
    event Event,
    to State,
    opts ...TransitionOption,
) BranchOption {
    return func(bc *BranchConfig) {
        // Create transition config
        tc := &TransitionConfig{
            From:  bc.from,
            To:    to,
            Event: event,
        }

        // Apply transition options
        for _, opt := range opts {
            opt(tc)
        }

        // Create branch path
        path := BranchPath{
            value:         value,
            event:         event,
            to:           to,
            description:   tc.description,
            guardTemplate: tc.guardTemplate,
            onEntry:      tc.onEntry,
            onExit:       tc.onExit,
            failedPhase:  tc.failedPhase,
        }

        bc.branches[value] = path
    }
}
```

## Benefits

1. **Explicit branching** - State-determined branches are first-class citizens
2. **Discoverable** - Can introspect branches programmatically
3. **Self-documenting** - Branch paths include descriptions
4. **Type-safe** - Discriminator returns string, branches match on string
5. **Flexible** - Supports N-way branches (not just binary)
6. **Backward compatible** - Existing code unchanged, opt-in feature
7. **Natural for "backward"** - Just another branch destination
8. **DRY** - Discriminator logic written once, not duplicated in guards
9. **Appropriate scope** - Only for state-determined branching; intent-based uses explicit events

## CLI Integration

The `sow advance` command leverages AddBranch via:

1. **Auto-determination:** Discriminator called, appropriate event fired
2. **Discovery:** `--list` shows available transitions with descriptions
3. **Explicit override:** Can still fire specific event if needed

See `DESIGN_ADVANCE_CLI.md` for full CLI details.

## Migration Path

### Before (standard project ReviewActive)
```go
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass, ...)
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail, ...)
OnAdvance(ReviewActive, func(p) { /* complex switch logic */ })
```

### After (with AddBranch)
```go
AddBranch(ReviewActive,
    BranchOn(getReviewAssessment),
    When("pass", EventReviewPass, FinalizeChecks,
        WithDescription("Review approved - proceed to finalization"),
        ...),
    When("fail", EventReviewFail, ImplementationPlanning,
        WithDescription("Review failed - return to planning for rework"),
        ...),
)
```

**Result:** Clearer intent, less duplication, better discoverability, self-documenting.

## Open Questions

1. **Error handling:** What if discriminator returns a value with no matching `When` clause?
   - **Decision:** Return error from DetermineEvent with helpful message
   - Future: Could add `Otherwise` branch (default case)

2. **Guard interaction:** Should branch paths have automatic guards?
   - **Decision:** `When` accepts standard `WithGuard()` options
   - Guards provide safety - discriminator determines intent, guard validates

3. **Naming:** Is "BranchOn" + "When" clear enough?
   - **Decision:** Yes - reads naturally: "branch on assessment: when pass..., when fail..."
   - Alternative names considered: "Switch/Case", "Discriminate/On"

4. **CLI integration:** How does `--list` display branches?
   - **Decision:** Show event names with descriptions and target states
   - Leverages `PermittedTriggers()` from stateless to filter by guard status

## Next Steps

1. ✅ Define two branching types (state-determined vs intent-based)
2. ✅ Design AddBranch API for state-determined branching
3. ✅ Add WithDescription() for transition discoverability
4. ✅ Clarify when to use AddBranch vs multiple AddTransition
5. Implement `BranchConfig` structs and internal representation
6. Add `AddBranch` and option functions to builder
7. Generate transitions and event determiner from branch config
8. Add introspection methods to ProjectTypeConfig
9. Update standard project to use AddBranch
10. Coordinate with CLI changes (see DESIGN_ADVANCE_CLI.md)
11. Add tests for N-way branching
12. Document in project SDK README
