# SDK Design: AddBranch API for State-Determined Branching

**Status**: Design
**Date**: 2025-11-06
**Context**: Adding declarative branching support to the Project SDK
**Related**: CLI Design (Enhanced advance Command), Implementation Guide

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Design Principles](#design-principles)
4. [Solution Overview](#solution-overview)
5. [API Specification](#api-specification)
6. [Internal Architecture](#internal-architecture)
7. [Usage Examples](#usage-examples)
8. [Integration with Existing SDK](#integration-with-existing-sdk)
9. [Backward Compatibility](#backward-compatibility)
10. [Implementation Plan](#implementation-plan)
11. [Testing Strategy](#testing-strategy)
12. [Open Questions](#open-questions)

---

## Executive Summary

This document specifies the `AddBranch()` API for the sow Project SDK, enabling declarative configuration of **state-determined branching** in project state machines.

**Key Concepts**:
- **State-Determined Branching**: Transitions where the next state can be determined by examining project state (e.g., review assessment = "pass" or "fail")
- **Declarative API**: `AddBranch()` with `BranchOn()` discriminator and `When()` clauses
- **Auto-Generation**: Generates transitions and event determiners automatically
- **Coexists with Intent-Based**: Multiple `AddTransition()` calls handle cases requiring explicit orchestrator choice

**Benefits**:
- Explicit branching logic (no hidden logic in OnAdvance)
- Discoverable transitions (introspection APIs)
- Self-documenting (descriptions co-located with branches)
- Scales to N-way branches (not limited to binary)

---

## Problem Statement

### Current Approach (Workaround)

The standard project's ReviewActive state uses branching but lacks first-class support:

```go
// Two transitions with identical guards (misleading)
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass,
    WithGuard("latest review approved", ...))
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail,
    WithGuard("latest review approved", ...))  // Same guard!

// Real branching logic hidden in event determiner
OnAdvance(ReviewActive, func(p *Project) (Event, error) {
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass": return EventReviewPass, nil
    case "fail": return EventReviewFail, nil
    }
})
```

**Issues**:
1. Guards are misleading (both say "review approved")
2. Branching logic split between guards and determiner
3. Not discoverable (can't introspect branches programmatically)
4. Doesn't scale to N-way branches
5. No descriptions explaining what each branch does

### Two Types of Branching

Through exploration, we identified two distinct patterns:

#### Type 1: State-Determined Branching
**Decision discoverable from project state**:
- Review assessment already in artifact metadata
- Test results already in outputs
- Validation status checkable from phase data

**Characteristic**: A function can examine the project and determine which way to go.

**Solution**: `AddBranch()` API (this design)

#### Type 2: Intent-Based Branching
**Decision requires orchestrator/user choice**:
- "Add more research" vs "Finalize research"
- "Skip this phase" vs "Do this phase"
- "Retry with changes" vs "Abandon"

**Characteristic**: Cannot be discovered from state - requires external input.

**Solution**: Multiple `AddTransition()` + explicit event in CLI (separate design)

---

## Design Principles

1. **Declarative over Imperative**: Configuration, not code
2. **Explicit over Implicit**: Branching logic visible in builder
3. **Discoverable**: Support introspection for CLI and tooling
4. **Composable**: Works alongside existing `AddTransition()`
5. **Backward Compatible**: Existing code unchanged, opt-in
6. **Self-Documenting**: Descriptions co-located with branches
7. **Type-Safe**: Compile-time validation where possible

---

## Solution Overview

### High-Level Approach

`AddBranch()` is a new builder method that:
1. Accepts a discriminator function (examines project state)
2. Defines branch paths with `When()` clauses
3. Auto-generates `AddTransition()` calls for each branch
4. Auto-generates `OnAdvance()` determiner using discriminator

### Conceptual Model

```
State: ReviewActive
  │
  ├─ BranchOn(getReviewAssessment)  ← Discriminator function
  │
  ├─ When("pass") → EventReviewPass → FinalizeChecks
  │    ├─ Description: "Review approved - proceed to finalization"
  │    └─ Guard: reviewAssessment == "pass"
  │
  └─ When("fail") → EventReviewFail → ImplementationPlanning
       ├─ Description: "Review failed - return to planning for rework"
       ├─ Guard: reviewAssessment == "fail"
       └─ OnEntry: setupRework()
```

### Why Both Event and Target State?

The underlying state machine library (stateless) requires events to trigger transitions. We specify both `event` and `to` (target state) in `When()` for:

- **Documentation**: Makes branching self-documenting
- **Validation**: Ensures event-to-state binding is correct
- **Introspection**: Allows querying target state without looking up transition tables
- **Clarity**: Explicit is better than implicit

---

## API Specification

### Core Types

```go
// BranchOption configures a branch
type BranchOption func(*BranchConfig)

// TransitionOption configures a transition (existing type, extended)
type TransitionOption func(*TransitionConfig)
```

### New Builder Method

```go
// AddBranch configures state-determined branching from a state
//
// Example:
//   AddBranch(ReviewActive,
//       BranchOn(getReviewAssessment),
//       When("pass", EventReviewPass, FinalizeChecks,
//           WithDescription("Review approved - proceed to finalization")),
//       When("fail", EventReviewFail, ImplementationPlanning,
//           WithDescription("Review failed - return to planning for rework")),
//   )
//
// This automatically generates:
// - Transitions for each When clause
// - Event determiner using the discriminator
//
func (b *ProjectTypeConfigBuilder) AddBranch(
    from State,
    opts ...BranchOption,
) *ProjectTypeConfigBuilder
```

### Branch Configuration Options

```go
// BranchOn specifies the discriminator function
//
// The discriminator examines project state and returns a string value
// that determines which When clause matches.
//
// Example:
//   BranchOn(func(p *state.Project) string {
//       // Get review assessment from latest approved review
//       phase := p.Phases["review"]
//       for i := len(phase.Outputs) - 1; i >= 0; i-- {
//           if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
//               return phase.Outputs[i].Metadata["assessment"].(string)
//           }
//       }
//       return ""  // No approved review yet
//   })
//
func BranchOn(discriminator func(*state.Project) string) BranchOption

// When defines a branch path based on discriminator value
//
// Parameters:
//   value - Discriminator value to match (e.g., "pass", "fail")
//   event - Event to fire when this branch is taken
//   to - Target state for this branch
//   opts - Standard transition options (guards, actions, descriptions)
//
// Example:
//   When("pass",
//       sdkstate.Event(EventReviewPass),
//       sdkstate.State(FinalizeChecks),
//       WithGuard("review passed", func(p) bool { return reviewAssessment(p) == "pass" }),
//       WithDescription("Review approved - proceed to finalization"),
//   )
//
func When(
    value string,
    event Event,
    to State,
    opts ...TransitionOption,
) BranchOption
```

### New Transition Option: WithDescription

```go
// WithDescription adds a human-readable description to a transition
//
// Descriptions are:
// - Context-specific (same event from different states can have different meanings)
// - Co-located with guards and actions
// - Used by CLI --list to show what each transition does
// - Visible to orchestrators for decision-making
//
// Example:
//   WithDescription("Review approved - proceed to finalization")
//
func WithDescription(description string) TransitionOption
```

### New Introspection Methods

```go
// TransitionInfo describes a single transition
type TransitionInfo struct {
    Event       Event
    From        State
    To          State
    Description string
    GuardDesc   string
}

// GetAvailableTransitions returns all configured transitions from a state
// Note: Returns configured transitions, not filtered by guards
// Use machine.PermittedTriggers() for guard-filtered list
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from State) []TransitionInfo

// GetTransitionDescription returns human-readable description
// Returns empty string if no description configured
func (ptc *ProjectTypeConfig) GetTransitionDescription(from State, event Event) string

// GetTargetState returns the target state for a transition
// Returns empty state if transition not found
func (ptc *ProjectTypeConfig) GetTargetState(from State, event Event) State

// GetGuardDescription returns the guard description for a transition
// Returns empty string if no guard or no description
func (ptc *ProjectTypeConfig) GetGuardDescription(from State, event Event) string

// IsBranchingState checks if a state has branches configured via AddBranch
func (ptc *ProjectTypeConfig) IsBranchingState(state State) bool
```

---

## Internal Architecture

### Data Structures

```go
// In config.go

// BranchConfig represents a state-determined branch point
type BranchConfig struct {
    from          State                                    // Source state
    discriminator func(*state.Project) string             // Returns branch value
    branches      map[string]BranchPath                   // value -> branch path
}

// BranchPath represents one possible branch destination
type BranchPath struct {
    value       string              // Discriminator value that triggers this path
    event       Event               // Event to fire
    to          State               // Target state
    description string              // Human-readable description

    // Standard transition configuration
    guardTemplate GuardTemplate     // Optional guard (in addition to discriminator)
    onEntry      Action             // OnEntry action
    onExit       Action             // OnExit action
    failedPhase  string             // Phase to mark as failed
}

// TransitionConfig extended with description
type TransitionConfig struct {
    From  sdkstate.State
    To    sdkstate.State
    Event sdkstate.Event

    guardTemplate GuardTemplate
    onEntry      Action
    onExit       Action
    failedPhase  string
    description  string  // NEW
}

// ProjectTypeConfig extended with branches
type ProjectTypeConfig struct {
    // ... existing fields ...

    branches map[State]*BranchConfig  // NEW: state -> branch config
}
```

### Auto-Generation Logic

When `AddBranch()` is called, the builder:

1. **Creates BranchConfig** from options
2. **Generates transitions** for each `When()` clause:
   ```go
   for _, path := range bc.branches {
       b.AddTransition(from, path.to, path.event,
           WithGuard(path.guardTemplate...),
           WithDescription(path.description),
           WithOnEntry(path.onEntry),
           // ... other options
       )
   }
   ```

3. **Generates event determiner**:
   ```go
   b.OnAdvance(from, func(p *state.Project) (Event, error) {
       value := bc.discriminator(p)
       path, exists := bc.branches[value]
       if !exists {
           return "", fmt.Errorf("no branch defined for discriminator value %q from state %s",
               value, from)
       }
       return path.event, nil
   })
   ```

### File Locations

**New files**:
- `cli/internal/sdks/project/branch.go` - Branch types and option functions

**Modified files**:
- `cli/internal/sdks/project/builder.go` - Add `AddBranch()` method
- `cli/internal/sdks/project/config.go` - Add `branches` field, introspection methods
- `cli/internal/sdks/project/types.go` - Add `BranchConfig`, `BranchPath`

---

## Usage Examples

### Example 1: Binary Branch (Review Pass/Fail)

```go
// Standard project's ReviewActive state
import (
    "github.com/yourusername/sow/cli/internal/sdks/project"
    sdkstate "github.com/yourusername/sow/cli/internal/sdks/state"
    "github.com/yourusername/sow/cli/internal/sdks/project/state"
)

builder.
    AddBranch(
        sdkstate.State(ReviewActive),
        project.BranchOn(func(p *state.Project) string {
            // Get review assessment from latest approved review artifact
            phase := p.Phases["review"]
            for i := len(phase.Outputs) - 1; i >= 0; i-- {
                artifact := phase.Outputs[i]
                if artifact.Type == "review" && artifact.Approved {
                    if assessment, ok := artifact.Metadata["assessment"].(string); ok {
                        return assessment  // "pass" or "fail"
                    }
                }
            }
            return ""  // No approved review yet
        }),
        project.When("pass",
            sdkstate.Event(EventReviewPass),
            sdkstate.State(FinalizeChecks),
            project.WithGuard("review passed", func(p *state.Project) bool {
                return getReviewAssessment(p) == "pass"
            }),
            project.WithDescription("Review approved - proceed to finalization"),
        ),
        project.When("fail",
            sdkstate.Event(EventReviewFail),
            sdkstate.State(ImplementationPlanning),
            project.WithGuard("review failed", func(p *state.Project) bool {
                return getReviewAssessment(p) == "fail"
            }),
            project.WithFailedPhase("review"),
            project.WithDescription("Review failed - return to planning for rework"),
            project.WithOnEntry(func(p *state.Project) error {
                // Increment implementation phase iteration
                impl := p.Phases["implementation"]
                impl.Iteration++

                // Add failed review as input to implementation phase
                reviewArtifact := getLatestReviewArtifact(p)
                impl.Inputs = append(impl.Inputs, reviewArtifact)

                return nil
            }),
        ),
    )
```

### Example 2: N-Way Branch (Deployment Targets)

```go
// Hypothetical: deployment target selection based on metadata
builder.
    AddBranch(
        sdkstate.State(DeploymentReady),
        project.BranchOn(func(p *state.Project) string {
            // Deployment target set during planning phase
            deploy := p.Phases["deployment"]
            if target, ok := deploy.Metadata["target"].(string); ok {
                return target  // "staging", "production", or "canary"
            }
            return "staging"  // Default
        }),
        project.When("staging",
            sdkstate.Event(EventDeployStaging),
            sdkstate.State(DeployingStaging),
            project.WithDescription("Deploy to staging environment for testing"),
        ),
        project.When("production",
            sdkstate.Event(EventDeployProduction),
            sdkstate.State(DeployingProduction),
            project.WithGuard("all tests passed", func(p *state.Project) bool {
                return allTestsPassed(p)
            }),
            project.WithDescription("Deploy to production environment (requires all tests passed)"),
        ),
        project.When("canary",
            sdkstate.Event(EventDeployCanary),
            sdkstate.State(DeployingCanary),
            project.WithGuard("canary feature flag enabled", func(p *state.Project) bool {
                return p.PhaseMetadataBool("deployment", "canary_enabled")
            }),
            project.WithDescription("Deploy to canary environment for gradual rollout"),
        ),
    )
```

### Example 3: Combining with Intent-Based Branching

```go
// Exploration project: some states auto-determine, others require explicit choice

builder.
    // State-determined: All tasks complete or abandoned (discoverable from state)
    AddBranch(
        sdkstate.State(Active),
        project.BranchOn(func(p *state.Project) string {
            if p.AllTasksComplete() {
                return "complete"
            }
            if p.AllTasksAbandoned() {
                return "abandoned"
            }
            return "in_progress"
        }),
        project.When("complete",
            sdkstate.Event(EventBeginSummarizing),
            sdkstate.State(Summarizing),
            project.WithDescription("All research tasks complete - begin summarization"),
        ),
        project.When("abandoned",
            sdkstate.Event(EventFinalize),
            sdkstate.State(Finalizing),
            project.WithDescription("All tasks abandoned - finalize without summary"),
        ),
        // No "in_progress" branch - will error if discriminator returns this
    ).

    // Intent-based: Orchestrator decides (NOT using AddBranch)
    AddTransition(
        sdkstate.State(Summarizing),
        sdkstate.State(Finalizing),
        sdkstate.Event(EventFinalize),
        project.WithDescription("Complete research and move to finalization phase"),
    ).
    AddTransition(
        sdkstate.State(Summarizing),
        sdkstate.State(Planning),
        sdkstate.Event(EventAddMoreResearch),
        project.WithDescription("Return to planning to add more research topics"),
    )
    // No OnAdvance for Summarizing - orchestrator must choose via: sow advance [event]
```

---

## Integration with Existing SDK

### Coexistence with AddTransition

`AddBranch()` and `AddTransition()` can coexist in the same project type:

```go
builder.
    // Linear state
    AddTransition(Planning, Active, EventStartResearch,
        WithDescription("Begin research phase")).
    OnAdvance(Planning, func(_ *state.Project) (Event, error) {
        return EventStartResearch, nil
    }).

    // State-determined branching
    AddBranch(Active,
        BranchOn(checkTaskStatus),
        When("complete", EventBeginSummarizing, Summarizing, ...),
        When("abandoned", EventFinalize, Finalizing, ...),
    ).

    // Intent-based branching (no OnAdvance)
    AddTransition(Summarizing, Finalizing, EventFinalize, ...).
    AddTransition(Summarizing, Planning, EventAddMoreResearch, ...)
```

### Relationship with OnAdvance

**AddBranch auto-generates OnAdvance**:
- Calls discriminator
- Returns matching event
- Errors if no matching branch

**Manual OnAdvance still works**:
- For simple linear states
- Can coexist with AddBranch on other states
- Do NOT use both AddBranch and OnAdvance on same state (will conflict)

### Relationship with Guards

**Branches can have guards**:
- Discriminator determines intent (which branch to take)
- Guard validates preconditions (whether branch is safe)
- Both must pass for transition to succeed

**Example**:
```go
When("production",
    EventDeployProduction,
    DeployingProduction,
    WithGuard("all tests passed", allTestsPassed),  // Guard checks precondition
    // Discriminator already determined intent is "production"
)
```

---

## Backward Compatibility

### Existing Code Unchanged

All existing project types continue working without modification:

```go
// Simple linear flow (no changes needed)
builder.
    AddTransition(Active, Summarizing, EventBeginSummarizing,
        WithGuard("all tasks resolved", allTasksResolved)).
    OnAdvance(Active, func(_ *state.Project) (Event, error) {
        return EventBeginSummarizing, nil
    })
```

### Opt-In Feature

`AddBranch()` is opt-in:
- Use for state-determined branching
- Continue using AddTransition + OnAdvance for linear states
- Use multiple AddTransition (no OnAdvance) for intent-based branching

### Migration Path

**Before** (ReviewActive workaround):
```go
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass, ...)
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail, ...)
OnAdvance(ReviewActive, func(p) { /* complex switch logic */ })
```

**After** (with AddBranch):
```go
AddBranch(ReviewActive,
    BranchOn(getReviewAssessment),
    When("pass", EventReviewPass, FinalizeChecks, ...),
    When("fail", EventReviewFail, ImplementationPlanning, ...),
)
```

**Benefits**: Clearer intent, less duplication, better discoverability.

---

## Implementation Plan

### Phase 1: Core Types and Options (1-2 days)

**Files**: `cli/internal/sdks/project/branch.go`

```go
// Implement:
type BranchConfig struct { ... }
type BranchPath struct { ... }
func BranchOn(discriminator) BranchOption
func When(value, event, to, opts...) BranchOption
```

### Phase 2: WithDescription for Transitions (0.5 day)

**Files**: `cli/internal/sdks/project/types.go`, `cli/internal/sdks/project/builder.go`

```go
// Add description field to TransitionConfig
// Implement WithDescription option
// Update AddTransition to store descriptions
```

### Phase 3: AddBranch Builder Method (1 day)

**Files**: `cli/internal/sdks/project/builder.go`

```go
// Implement AddBranch:
// 1. Accept BranchConfig
// 2. Generate AddTransition calls for each When clause
// 3. Generate OnAdvance determiner
// 4. Store BranchConfig in builder
```

### Phase 4: Introspection Methods (1 day)

**Files**: `cli/internal/sdks/project/config.go`

```go
// Implement:
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from State) []TransitionInfo
func (ptc *ProjectTypeConfig) GetTransitionDescription(from, event) string
func (ptc *ProjectTypeConfig) GetTargetState(from, event) State
func (ptc *ProjectTypeConfig) GetGuardDescription(from, event) string
func (ptc *ProjectTypeConfig) IsBranchingState(state) bool
```

### Phase 5: Update Standard Project (1 day)

**Files**: `cli/internal/projects/standard/standard.go`

```go
// Replace ReviewActive workaround with AddBranch
// Add descriptions to all transitions
```

### Phase 6: Testing (1-2 days)

**Files**: `cli/internal/sdks/project/branch_test.go`, `cli/internal/sdks/project/builder_test.go`

```go
// Unit tests for:
// - BranchOn/When option functions
// - AddBranch auto-generation
// - Introspection methods
// - N-way branching
// - Error cases (no matching branch)
```

**Total Estimated Time**: 5-7 days

---

## Testing Strategy

### Unit Tests

```go
// cli/internal/sdks/project/branch_test.go

func TestBranchOn(t *testing.T) {
    // Test discriminator configuration
}

func TestWhen(t *testing.T) {
    // Test branch path creation
    // Test transition option forwarding
}

func TestAddBranchGeneratesTransitions(t *testing.T) {
    // Verify AddTransition calls generated
    // Verify OnAdvance determiner created
}

func TestAddBranchNWay(t *testing.T) {
    // Test 3+ branch paths
}

func TestDiscriminatorNoMatch(t *testing.T) {
    // Test error when discriminator returns unknown value
}

func TestIntrospectionMethods(t *testing.T) {
    // Test GetAvailableTransitions
    // Test GetTransitionDescription
    // Test GetTargetState
    // Test IsBranchingState
}
```

### Integration Tests

```go
// cli/internal/sdks/project/integration_test.go

func TestReviewBranching(t *testing.T) {
    // Create standard project
    // Advance to ReviewActive
    // Set review assessment to "pass"
    // Advance (should go to FinalizeChecks)

    // Reset, set assessment to "fail"
    // Advance (should go to ImplementationPlanning)
}

func TestBranchingWithExploration(t *testing.T) {
    // Mix AddBranch and multiple AddTransition
    // Verify both work correctly
}
```

### CLI Integration Tests

See CLI design document for tests involving `sow advance --list` and explicit events.

---

## Open Questions

### 1. Error Handling: No Matching Branch

**Question**: What if discriminator returns a value with no matching `When` clause?

**Options**:
- A) Return error from DetermineEvent with helpful message (current design)
- B) Add `Otherwise` branch (default case)
- C) Panic (fail fast)

**Decision**: Option A - clear error message. Option B (Otherwise) can be added later if needed.

**Implementation**:
```go
b.OnAdvance(from, func(p *state.Project) (Event, error) {
    value := bc.discriminator(p)
    path, exists := bc.branches[value]
    if !exists {
        availableValues := make([]string, 0, len(bc.branches))
        for v := range bc.branches {
            availableValues = append(availableValues, v)
        }
        return "", fmt.Errorf(
            "no branch defined for discriminator value %q from state %s (available: %v)",
            value, from, availableValues)
    }
    return path.event, nil
})
```

### 2. Guard Interaction

**Question**: Should branch paths have automatic guards based on discriminator?

**Decision**: No automatic guards. Use `WithGuard()` option in `When()` clause.

**Rationale**:
- Discriminator determines intent (which branch)
- Guard validates preconditions (whether safe to proceed)
- Separation of concerns
- Guards provide additional safety

### 3. Naming

**Question**: Is "BranchOn" + "When" clear enough?

**Alternatives**: "Switch/Case", "Discriminate/On", "Route/To"

**Decision**: "BranchOn" + "When" - reads naturally and aligns with branching concept.

### 4. Transition Descriptions

**Question**: Should descriptions be required or optional?

**Decision**: Optional but strongly recommended. CLI will work without descriptions but be less helpful.

**Best Practice**: Always add descriptions for transitions, especially in branching states.

### 5. Multiple Discriminators

**Question**: Should we support multiple discriminators (AND/OR logic)?

**Example**:
```go
BranchOn(getReviewAssessment, getCIStatus)  // Both matter
```

**Decision**: Defer. Use single discriminator that combines logic:
```go
BranchOn(func(p) string {
    assessment := getReviewAssessment(p)
    ciStatus := getCIStatus(p)
    return fmt.Sprintf("%s-%s", assessment, ciStatus)
})
When("pass-success", ...)
When("pass-failure", ...)
When("fail-success", ...)
When("fail-failure", ...)
```

If this pattern becomes common, add `CombineDiscriminators()` helper.

---

## Conclusion

The `AddBranch()` API provides first-class support for state-determined branching in the Project SDK:

- **Declarative**: Configuration over code
- **Discoverable**: Introspection for CLI and tooling
- **Self-documenting**: Descriptions co-located with branches
- **Scalable**: Supports N-way branches
- **Compatible**: Coexists with existing patterns
- **Focused**: Solves state-determined branching, not intent-based

Next steps:
1. Review and approve this design
2. Implement in phases (5-7 days)
3. Update standard project to use AddBranch
4. Coordinate with CLI enhancements (see CLI design document)
