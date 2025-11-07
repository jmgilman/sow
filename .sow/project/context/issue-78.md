# Issue #78: CLI Enhanced Advance Command and Standard Project Refactoring

**URL**: https://github.com/jmgilman/sow/issues/78
**State**: OPEN

## Description

# Work Unit 003: CLI Enhanced Advance Command and Standard Project Refactoring

**Status**: Ready for Implementation
**Estimated Duration**: 4-5 days
**Implementation Approach**: Test-Driven Development (TDD)

---

## 1. Behavioral Goals

This work unit has two distinct user stories addressing different personas:

### Story 1: Orchestrator Explicit Event Selection

**As an** orchestrator agent,
**I need** to explicitly select events and discover available transitions when using the `sow advance` command,
**So that** I can handle intent-based branching and make informed decisions about state progression.

**Success Criteria for Reviewers**:
1. Orchestrators can list all available transitions with `sow advance --list` showing descriptions and requirements
2. Orchestrators can explicitly fire events with `sow advance [event]` for intent-based branching
3. Orchestrators can validate transitions before executing with `sow advance --dry-run [event]`
4. Guard failures show helpful error messages with guard descriptions
5. Auto-determination still works for linear and state-determined branching (backward compatible)
6. Error messages suggest next actions (e.g., "use --list to see options")

### Story 2: Standard Project Branching Refactoring

**As a** project type maintainer,
**I need** to replace the ReviewActive workaround with the new AddBranch API,
**So that** the standard project demonstrates proper branching patterns and serves as a reference implementation.

**Success Criteria for Reviewers**:
1. ReviewActive state uses `AddBranch()` with `BranchOn()` discriminator and `When()` clauses
2. All transitions in standard project have human-readable descriptions
3. Existing standard projects continue to work (backward compatibility)
4. Review pass/fail workflow still functions correctly
5. Code is cleaner and more maintainable than the workaround

---

## 2. Existing Code Context

### Integration Overview

This work unit has two major components that build on work unit 002:

**CLI Enhancement**: The current advance command (`cli/cmd/advance.go`, 87 lines) only supports auto-determination via `DetermineEvent()`. We'll add four operation modes by accepting an optional `[event]` positional argument and adding `--list` and `--dry-run` flags. The implementation uses the introspection methods from work unit 002 (`GetAvailableTransitions`, `GetTransitionDescription`, `GetTargetState`, `GetGuardDescription`, `IsBranchingState`) to discover and display transition information. The state machine wrapper already provides `PermittedTriggers()` for guard-filtered event lists and `CanFire()` for validation.

**Standard Project Refactoring**: The standard project's ReviewActive state (lines 118-241 of `standard.go`) currently uses a workaround where two transitions share identical guards ("latest review approved") but the real branching logic lives in the OnAdvance discriminator (lines 207-241). This discriminator examines the review artifact's assessment metadata and returns either `EventReviewPass` or `EventReviewFail`. We'll replace this split logic with a single `AddBranch()` call that uses `BranchOn(getReviewAssessment)` as the discriminator and defines `When("pass")` and `When("fail")` branches with proper descriptions. The existing `latestReviewApproved` guard helper can be refactored or removed since AddBranch will handle the logic.

The CLI enhancement depends on SDK introspection methods being available. The standard project refactoring demonstrates that the entire SDK+CLI integration works end-to-end. Both components follow TDD principles with comprehensive test coverage at unit and integration levels.

### Key Files to Modify

**CLI Command Enhancement**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/cmd/advance.go:1-87` (entire file)
  - Current: Lines 42-82 implement simple auto-determination only
  - New: Split into mode-switching logic with 4 helper functions
  - Add: `[event]` positional arg (line 23: `Args: cobra.NoArgs` → `Args: cobra.MaximumNArgs(1)`)
  - Add: `--list` and `--dry-run` flags (lines ~85-86, after cmd definition)

**Standard Project Refactoring**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/standard.go:118-241`
  - Current ReviewActive transitions: Lines 118-165 (two transitions with identical guards)
  - Current ReviewActive OnAdvance: Lines 207-241 (discriminator logic)
  - Replace with: Single `AddBranch()` call using new API from work unit 002
  - Enhance: Add descriptions to ALL transitions in standard project (not just ReviewActive)

**Test Files** (new):
- `cli/cmd/advance_test.go` - CLI unit tests (currently missing)

**Integration Tests** (existing, modify):
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/lifecycle_test.go`
  - Verify backward compatibility after refactoring

### Reference Implementations

**State Machine Methods** (already available):
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/state/machine.go:50-63`
  - `PermittedTriggers()` returns guard-filtered events - use for `--list`
  - `CanFire()` validates individual events - use for `--dry-run` and explicit events

**Guard Error Handling** (pattern to follow):
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/sdks/state/machine.go:65-83`
  - `guardDescriptions` map stores custom descriptions for better error messages
  - Pattern: Check error string, look up description, provide helpful context

**Current Discriminator Pattern** (to refactor):
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/standard.go:207-241`
  - Shows the exact logic AddBranch should auto-generate
  - Finds latest approved review, extracts assessment, returns corresponding event
  - This pattern becomes the `BranchOn()` discriminator function

---

## 3. Existing Documentation Context

### CLI Enhanced Advance Design Document

The complete CLI specification (`.sow/knowledge/designs/cli-enhanced-advance.md`) provides the implementation roadmap for the command enhancement:

**Section 5 (Command Specification)**: Defines the exact command signature (`sow advance [event] [flags]`), flag definitions (`--list` and `--dry-run`), and mutual exclusivity rules. The `[event]` argument is optional and validated against permitted events. The `--list` flag cannot combine with `[event]`, and `--dry-run` requires `[event]`.

**Section 6 (Behavior Modes)**: Specifies all four operation modes with exact output formats:
1. **Auto-determination** (backward compatible): `sow advance` with no args calls `DetermineEvent()` and fires that event
2. **Explicit event**: `sow advance [event]` validates and fires the specified event
3. **Discovery**: `sow advance --list` shows all available transitions with descriptions, target states, and guard requirements
4. **Dry-run**: `sow advance --dry-run [event]` validates without executing, showing what would happen

**Section 7 (Implementation Details)**: Provides exact function signatures and logic flow for all helper functions (`listAvailableTransitions`, `validateTransition`, `executeExplicitTransition`, `executeAutoTransition`). These functions demonstrate how to use SDK introspection methods and state machine methods together. The error handling patterns show how to provide helpful suggestions (e.g., "Use --list to see available transitions").

**Section 10 (Error Handling)**: Specifies error messages for all failure cases (invalid events, guard failures, terminal states, conflicting flags) with exact output formats that orchestrators can parse.

### SDK AddBranch Design Document

The SDK design (`.sow/knowledge/designs/sdk-addbranch-api.md`) informs the standard project refactoring:

**Section 7.1 (Binary Branch Example)**: Shows the exact pattern for refactoring ReviewActive - the discriminator function (lines 391-403) finds the latest approved review and extracts the assessment, the `When()` clauses (lines 404-431) define pass/fail branches with descriptions, guards, and actions. This is the template to follow.

**Section 2 (Problem Statement)**: Documents the current ReviewActive workaround issues (lines 48-73) - identical guards that are misleading, split logic between guards and determiner, lack of discoverability. Understanding these problems helps ensure the refactoring addresses them.

**Section 9 (Backward Compatibility)**: Emphasizes that AddBranch is opt-in and coexists with existing patterns. The standard project refactoring must maintain backward compatibility - existing projects already in ReviewActive state must continue working.

### Discovery Document

The discovery analysis (`.sow/project/discovery/analysis.md`) provides detailed code locations:

**Section 2 (Current Advance Command)**: Lines 48-104 document the existing implementation with exact line references, showing the basic structure (load project, determine event, fire event, save) that the new modes build upon.

**Section 4 (Standard Project ReviewActive Workaround)**: Lines 245-361 provide complete analysis of the problematic pattern with line-by-line explanation of both transitions and the discriminator. This analysis shows exactly what needs to be refactored and why.

**Section 5 (State Machine Integration)**: Lines 365-494 document how the state machine wrapper works, including guard evaluation flow and closure binding patterns. Understanding this is critical for using `PermittedTriggers()` and `CanFire()` correctly.

---

## 4. Implementation Approach

This work unit has two major components implemented in sequence:

### Component 1: CLI Enhanced Advance Command (3-4 days)

Implement incrementally, testing each mode before adding the next:

#### Phase 1A: Command Infrastructure (0.5 days)
**TDD Approach**:
1. Write test: `TestAdvanceCommandSignature` - verify flags and args accepted
2. Modify NewAdvanceCmd():
   - Change `Args: cobra.NoArgs` to `Args: cobra.MaximumNArgs(1)`
   - Add flag variables and definitions for `--list` and `--dry-run`
3. Write test: `TestAdvanceFlagValidation` - verify mutual exclusivity
4. Implement flag validation logic in main RunE
5. Verify tests pass

#### Phase 1B: Auto-Determination Mode (Enhanced) (0.5 days)
**TDD Approach**:
1. Write tests:
   - `TestAdvanceAutoLinear` - linear state progression
   - `TestAdvanceAutoBranching` - state-determined branching
   - `TestAdvanceAutoIntentBased` - fails with helpful message
2. Extract current logic to `executeAutoTransition()` helper
3. Enhance error message when DetermineEvent fails (show available options)
4. Verify tests pass and backward compatibility maintained

#### Phase 1C: List Mode (1 day)
**TDD Approach**:
1. Write tests:
   - `TestAdvanceListAvailable` - shows all permitted transitions
   - `TestAdvanceListBlocked` - shows blocked transitions with guards
   - `TestAdvanceListTerminal` - terminal state message
2. Implement `listAvailableTransitions()` function:
   - Call `machine.PermittedTriggers()` for guard-filtered events
   - For each event, call introspection methods (GetTargetState, GetTransitionDescription, GetGuardDescription)
   - Format output per design spec
   - Handle edge case: all transitions blocked by guards
3. Verify tests pass

#### Phase 1D: Dry-Run Mode (0.5 days)
**TDD Approach**:
1. Write tests:
   - `TestAdvanceDryRunValid` - transition would succeed
   - `TestAdvanceDryRunBlocked` - guard failure
   - `TestAdvanceDryRunInvalid` - event not configured
   - `TestAdvanceDryRunNoEvent` - missing event argument
2. Implement `validateTransition()` function:
   - Check event configured via GetTargetState
   - Check event permitted via CanFire
   - Format output per design spec
3. Verify tests pass

#### Phase 1E: Explicit Event Mode (1 day)
**TDD Approach**:
1. Write tests:
   - `TestAdvanceExplicitSuccess` - fires event correctly
   - `TestAdvanceExplicitGuardFailure` - shows guard description
   - `TestAdvanceExplicitInvalidEvent` - helpful error
2. Implement `executeExplicitTransition()` function:
   - Validate event configured via GetTargetState
   - Fire event via FireWithPhaseUpdates
   - Enhanced error messages using GetGuardDescription
3. Verify tests pass

#### Phase 1F: Integration and Error Handling (0.5 days)
**TDD Approach**:
1. Write tests for all error cases from design doc section 10
2. Wire up mode switching in main RunE function
3. Add error message suggestions (e.g., "Use --list to see available transitions")
4. End-to-end manual testing with real projects
5. Verify all tests pass

### Component 2: Standard Project Refactoring (1 day)

After CLI is working, refactor standard project to demonstrate the full integration:

#### Phase 2A: Add Descriptions to All Transitions (0.25 days)
**TDD Approach**:
1. Write test: `TestStandardProjectDescriptions` - verify all transitions have descriptions
2. Add `WithDescription()` to all existing transitions in standard.go
3. Use clear, concise descriptions explaining what each transition does
4. Verify test passes

#### Phase 2B: Refactor ReviewActive with AddBranch (0.5 days)
**TDD Approach**:
1. Write test: `TestReviewActiveBranchingRefactored` - verify both pass/fail paths work
2. Replace lines 118-165 (two transitions) with single AddBranch:
   - Use `BranchOn(getReviewAssessment)` discriminator
   - Define `When("pass")` and `When("fail")` branches
   - Include descriptions, guards, and OnEntry actions
3. Remove lines 207-241 (OnAdvance discriminator - no longer needed)
4. Refactor or remove `latestReviewApproved` guard helper if no longer used
5. Verify test passes

#### Phase 2C: Backward Compatibility Testing (0.25 days)
**TDD Approach**:
1. Write test: `TestExistingProjectContinuesWorking` - load old project state, verify advance still works
2. Run all existing standard project tests
3. Manual testing with real standard projects
4. Verify no regressions

---

## 5. Testing Strategy

### Unit Tests (CLI Command)

**File**: `cli/cmd/advance_test.go` (new)

```go
// Infrastructure tests
func TestAdvanceCommandSignature(t *testing.T)  // Flags and args
func TestAdvanceFlagValidation(t *testing.T)    // Mutual exclusivity

// Auto-determination mode
func TestAdvanceAutoLinear(t *testing.T)        // Linear progression
func TestAdvanceAutoBranching(t *testing.T)     // State-determined branching
func TestAdvanceAutoIntentBased(t *testing.T)   // Shows helpful error

// List mode
func TestAdvanceListAvailable(t *testing.T)     // Shows permitted transitions
func TestAdvanceListBlocked(t *testing.T)       // Shows blocked transitions
func TestAdvanceListTerminal(t *testing.T)      // Terminal state message

// Dry-run mode
func TestAdvanceDryRunValid(t *testing.T)       // Would succeed
func TestAdvanceDryRunBlocked(t *testing.T)     // Guard failure
func TestAdvanceDryRunInvalid(t *testing.T)     // Event not configured
func TestAdvanceDryRunNoEvent(t *testing.T)     // Missing argument

// Explicit event mode
func TestAdvanceExplicitSuccess(t *testing.T)         // Fires successfully
func TestAdvanceExplicitGuardFailure(t *testing.T)    // Shows guard description
func TestAdvanceExplicitInvalidEvent(t *testing.T)    // Helpful error
```

### Integration Tests (Standard Project)

**File**: `cli/internal/projects/standard/lifecycle_test.go` (existing, add tests)

```go
// Description tests
func TestStandardProjectDescriptions(t *testing.T)  // All transitions have descriptions

// Refactoring validation
func TestReviewActiveBranchingRefactored(t *testing.T)  // Pass/fail paths work
func TestExistingProjectContinuesWorking(t *testing.T)  // Backward compatibility

// End-to-end workflow
func TestStandardProjectFullLifecycleWithCLI(t *testing.T)  // Use new advance modes
```

### End-to-End CLI Tests

Test all modes against real project types:

```go
// Test auto-advance
func TestCLIAutoAdvanceLinear(t *testing.T)      // Exploration project
func TestCLIAutoAdvanceBranching(t *testing.T)   // Standard project ReviewActive

// Test explicit events
func TestCLIExplicitEvent(t *testing.T)          // Intent-based branching

// Test discovery
func TestCLIListTransitions(t *testing.T)        // Shows options

// Test validation
func TestCLIDryRunTransition(t *testing.T)       // Pre-flight check
```

### Test Coverage Goals

- All four advance modes have unit tests
- All error cases have explicit tests
- Standard project refactoring has integration tests
- Backward compatibility verified with existing projects
- End-to-end workflows tested against multiple project types

---

## 6. Dependencies

**Depends on Work Unit 002** - Requires the following from the SDK branching support work unit:

**Required API Methods**:
- `AddBranch()` builder method for refactoring standard project
- `WithDescription()` transition option for adding descriptions
- All introspection methods for CLI modes:
  - `GetAvailableTransitions(from State) []TransitionInfo`
  - `GetTransitionDescription(from State, event Event) string`
  - `GetTargetState(from State, event Event) State`
  - `GetGuardDescription(from State, event Event) string`
  - `IsBranchingState(state State) bool`

**Why the dependency**: The CLI enhancement cannot list or introspect transitions without these methods. The standard project refactoring cannot use AddBranch until it exists.

---

## 7. Acceptance Criteria

Reviewers will verify:

### CLI Enhancement Criteria

- [ ] `sow advance` (no args) works for linear states (auto-determines and advances)
- [ ] `sow advance` (no args) works for state-determined branching (uses discriminator)
- [ ] `sow advance` (no args) shows helpful error for intent-based branching (lists options)
- [ ] `sow advance [event]` fires explicit events successfully when valid
- [ ] `sow advance [event]` shows guard description when guard fails
- [ ] `sow advance [event]` shows helpful error when event invalid
- [ ] `sow advance --list` shows all permitted transitions with descriptions and requirements
- [ ] `sow advance --list` shows blocked transitions when all guards fail
- [ ] `sow advance --list` shows terminal state message when no transitions configured
- [ ] `sow advance --dry-run [event]` validates successfully when event permitted
- [ ] `sow advance --dry-run [event]` shows guard description when guard blocks
- [ ] `sow advance --dry-run [event]` errors when event argument missing
- [ ] Error messages suggest helpful next steps (e.g., "Use --list to see available transitions")
- [ ] Backward compatibility: existing advance workflows unchanged

### Standard Project Refactoring Criteria

- [ ] ReviewActive uses `AddBranch()` with `BranchOn()` discriminator
- [ ] ReviewActive defines `When("pass")` and `When("fail")` branches
- [ ] Both branches have descriptions explaining what they do
- [ ] Pass branch transitions to FinalizeChecks
- [ ] Fail branch transitions to ImplementationPlanning with rework setup
- [ ] All transitions in standard project have descriptions
- [ ] Review pass workflow functions correctly (approval → finalization)
- [ ] Review fail workflow functions correctly (rejection → rework)
- [ ] Existing standard projects continue to work (no breaking changes)
- [ ] Code is cleaner than the workaround (no split logic)

### General Criteria

- [ ] All unit tests pass (CLI modes, standard project refactoring)
- [ ] All integration tests pass (end-to-end workflows)
- [ ] Code follows existing patterns and conventions
- [ ] No breaking changes to CLI or standard project behavior
- [ ] Documentation in code comments is clear

---

## 8. Code Examples

### Example 1: Main Mode Switching Logic

```go
func runAdvance(cmd *cobra.Command, args []string) error {
    // Get flags
    listFlag, _ := cmd.Flags().GetBool("list")
    dryRunFlag, _ := cmd.Flags().GetBool("dry-run")

    // Load project
    ctx := cmdutil.GetContext(cmd.Context())
    project, err := state.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    machine := project.Machine()
    currentState := sdkstate.State(project.Statechart.Current_state)

    // Mode 1: List available transitions
    if listFlag {
        if len(args) > 0 {
            return fmt.Errorf("--list cannot be combined with event argument")
        }
        return listAvailableTransitions(project, machine, currentState)
    }

    // Get event from argument if provided
    var event sdkstate.Event
    if len(args) > 0 {
        event = sdkstate.Event(args[0])
    }

    // Mode 2: Dry-run validation
    if dryRunFlag {
        if event == "" {
            return fmt.Errorf("--dry-run requires an event argument")
        }
        return validateTransition(project, machine, currentState, event)
    }

    // Mode 3: Explicit event execution
    if event != "" {
        return executeExplicitTransition(project, machine, currentState, event)
    }

    // Mode 4: Auto-determination (backward compatible)
    return executeAutoTransition(project, machine, currentState)
}
```

### Example 2: Using Introspection in List Mode

```go
func listAvailableTransitions(
    project *state.Project,
    machine *sdkstate.Machine,
    currentState sdkstate.State,
) error {
    fmt.Printf("Current state: %s\n\n", currentState)

    // Get guard-filtered events from state machine
    events, err := machine.PermittedTriggers()
    if err != nil {
        return fmt.Errorf("failed to get permitted triggers: %w", err)
    }

    if len(events) == 0 {
        // Check if transitions configured but all blocked
        allTransitions := project.Config().GetAvailableTransitions(currentState)
        if len(allTransitions) == 0 {
            fmt.Println("No transitions available from current state.")
            fmt.Println("This may be a terminal state.")
            return nil
        }

        // Show blocked transitions
        fmt.Println("Available transitions:")
        fmt.Println("\n(All configured transitions are currently blocked by guard conditions)\n")
        for _, t := range allTransitions {
            fmt.Printf("  sow advance %s  [BLOCKED]\n", t.Event)
            fmt.Printf("    → %s\n", t.To)
            if t.Description != "" {
                fmt.Printf("    %s\n", t.Description)
            }
            if t.GuardDesc != "" {
                fmt.Printf("    Requires: %s\n", t.GuardDesc)
            }
            fmt.Println()
        }
        return nil
    }

    // Show permitted transitions
    fmt.Println("Available transitions:")
    for _, event := range events {
        targetState := project.Config().GetTargetState(currentState, event)
        description := project.Config().GetTransitionDescription(currentState, event)
        guardDesc := project.Config().GetGuardDescription(currentState, event)

        fmt.Printf("\n  sow advance %s\n", event)
        fmt.Printf("    → %s\n", targetState)
        if description != "" {
            fmt.Printf("    %s\n", description)
        }
        if guardDesc != "" {
            fmt.Printf("    Requires: %s\n", guardDesc)
        }
    }
    fmt.Println()
    return nil
}
```

### Example 3: Standard Project ReviewActive Refactoring

```go
// BEFORE: Workaround with split logic
AddTransition(ReviewActive, FinalizeChecks, EventReviewPass,
    WithGuard("latest review approved", latestReviewApproved))
AddTransition(ReviewActive, ImplementationPlanning, EventReviewFail,
    WithGuard("latest review approved", latestReviewApproved))  // Same guard!

OnAdvance(ReviewActive, func(p *state.Project) (sdkstate.Event, error) {
    // Real logic hidden here
    assessment := getReviewAssessment(p)
    switch assessment {
    case "pass": return EventReviewPass, nil
    case "fail": return EventReviewFail, nil
    }
})

// AFTER: Declarative branching with AddBranch
AddBranch(
    sdkstate.State(ReviewActive),
    project.BranchOn(func(p *state.Project) string {
        // Discriminator examines project state
        phase := p.Phases["review"]
        for i := len(phase.Outputs) - 1; i >= 0; i-- {
            artifact := phase.Outputs[i]
            if artifact.Type == "review" && artifact.Approved {
                if assessment, ok := artifact.Metadata["assessment"].(string); ok {
                    return assessment  // "pass" or "fail"
                }
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
        project.WithOnEntry(/* rework setup action */),
    ),
)
```

---

## 9. Implementation Notes

### CLI Enhancement Notes

**Backward Compatibility**:
- No arguments = existing behavior (auto-determination)
- New flags and event argument are opt-in
- No breaking changes to command signature
- Error messages enhanced but not breaking

**Error Message Design**:
- Always provide context (current state, what went wrong)
- Suggest next actions (e.g., "Use --list to see available transitions")
- Include guard descriptions when guards fail
- Format for both human and orchestrator readability

**Integration with SDK**:
- Use `machine.PermittedTriggers()` for guard-filtered events (don't re-evaluate guards)
- Use introspection methods for descriptions and metadata
- Call `FireWithPhaseUpdates()` for consistency with existing behavior
- Save project state after successful transitions

**Testing Approach**:
- Mock or use minimal test projects for unit tests
- Test each mode independently before integration
- Verify flag validation and mutual exclusivity
- End-to-end tests against real project types

### Standard Project Refactoring Notes

**Migration Strategy**:
- Add descriptions first (separate commit, easier to review)
- Then refactor ReviewActive (single focused change)
- Test backward compatibility thoroughly
- Document what changed and why

**Discriminator Function**:
- Can extract discriminator logic to separate helper function for clarity
- Or inline in BranchOn (as shown in example) for co-location
- Must handle missing data gracefully (return empty string)

**OnEntry Actions**:
- Fail branch needs rework setup (increment iteration, add failed review as input)
- Preserve existing action logic from current OnAdvance/OnEntry

**Guard Considerations**:
- Current guards check "review approved" (binary)
- AddBranch discriminator checks "pass" vs "fail" (ternary: pass/fail/missing)
- Guards can still validate preconditions if needed
- Consider whether guards are still necessary after discriminator

### General Notes

**Code Organization**:
- CLI: Main RunE delegates to 4 helper functions (one per mode)
- Each helper is 30-50 lines, focused on single responsibility
- Standard project: Keep phase definitions together, transitions together

**Performance**:
- Introspection methods iterate transitions (O(n) where n = transitions from state)
- For standard project: ~10 transitions total, negligible performance impact
- No caching needed

**Documentation**:
- Add comments explaining each mode in RunE
- Document helper functions with usage examples
- Update standard project comments to explain AddBranch usage

**Error Handling**:
- Distinguish between user errors (invalid event) and system errors (failed to save)
- Provide actionable error messages
- Use fmt.Errorf with %w for error wrapping

---

## 10. File Locations

### Files to Create

- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/cmd/advance_test.go` - CLI unit tests

### Files to Modify

**CLI Enhancement**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/cmd/advance.go:1-87`
  - Add flags (lines ~85-86)
  - Change Args validation (line 41)
  - Replace RunE function (lines 42-82) with mode switching logic
  - Add 4 helper functions (~150 lines total)

**Standard Project Refactoring**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/standard.go:118-241`
  - Add descriptions to all transitions throughout file
  - Replace ReviewActive transitions (lines 118-165) with AddBranch
  - Remove ReviewActive OnAdvance (lines 207-241)
  - May refactor guards.go if helpers no longer needed

**Testing**:
- `/Users/josh/code/sow/.sow/worktrees/breakdown/branching/cli/internal/projects/standard/lifecycle_test.go`
  - Add tests for refactored ReviewActive
  - Add backward compatibility tests

---

## 11. References

- **CLI Design Document**: `.sow/knowledge/designs/cli-enhanced-advance.md` - Complete command specification with all modes, flags, and output formats
- **SDK Design Document**: `.sow/knowledge/designs/sdk-addbranch-api.md` - AddBranch API for standard project refactoring
- **Discovery Document**: `.sow/project/discovery/analysis.md` - Codebase analysis with file locations and current implementation details
- **Work Unit 002 Spec**: `.sow/project/work-units/002-sdk-branching-support.md` - Foundation work unit providing required APIs
- **Current Advance Command**: `cli/cmd/advance.go:1-87` - Starting point for enhancement
- **Current ReviewActive**: `cli/internal/projects/standard/standard.go:118-241` - Workaround to refactor

---

## 12. Success Metrics

This work unit succeeds when:

1. Orchestrators can discover available transitions without trial-and-error
2. Orchestrators can explicitly select events for intent-based branching
3. Orchestrators can validate transitions before executing (safety)
4. The standard project demonstrates proper branching patterns (no workarounds)
5. All existing workflows continue to function (backward compatibility)
6. Code is cleaner and more maintainable than before
7. Tests comprehensively cover all modes and edge cases

The ultimate validation is that orchestrator agents can now handle both state-determined and intent-based branching through the enhanced CLI, using the standard project as a reference implementation of best practices.
