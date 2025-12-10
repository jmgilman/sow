# Task Log

## Implementation Summary

Migrated project configuration and builder from `cli/internal/sdks/project/` to `libs/project/`.

### Files Created

1. **config.go** - ProjectTypeConfig with:
   - Phase management (GetPhaseForState, IsPhaseStartState, IsPhaseEndState)
   - Transition management (GetTransition, GetAvailableTransitions)
   - Task support (GetTaskSupportingPhases, PhaseSupportsTasks, GetDefaultTaskPhase)
   - State machine building (BuildProjectMachine)
   - Fire with phase updates (FireWithPhaseUpdates)
   - Branching support (IsBranchingState, DetermineEvent)

2. **project_builder.go** - ProjectTypeConfigBuilder with fluent API:
   - SetInitialState, WithPhase, AddTransition
   - OnAdvance, AddBranch, WithPrompt, WithOrchestratorPrompt
   - WithInitializer, Build

3. **phase_config.go** - PhaseConfig and PhaseOpt:
   - WithStartState, WithEndState, WithTaskSupport
   - WithInputs, WithOutputs, WithMetadataSchema

4. **transition_config.go** - TransitionConfig and ProjectTransitionOption:
   - WithProjectGuard, WithProjectOnEntry, WithProjectOnExit
   - WithProjectFailedPhase, WithDescription

5. **branch.go** - BranchConfig, BranchPath, BranchOption:
   - BranchOn, When with transition options

6. **config_options.go** - Additional option types

7. **errors.go** - Error types:
   - ErrNoDeterminer, ErrBranchNotFound
   - ErrTransitionFailed, ErrPhaseStatusUpdate

### Verification

- All tests pass with race detector enabled
- golangci-lint passes with 0 issues
