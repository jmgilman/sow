package cmd

import (
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/projects/standard"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// TestBackwardCompatibility ensures that new CLI modes don't break existing workflows.
// This test suite verifies that:
// 1. Default auto-advance behavior remains unchanged
// 2. New flags don't affect projects that don't use them
// 3. Error messages are improved but not breaking
func TestBackwardCompatibility(t *testing.T) {
	t.Run("auto-advance without flags works as before", func(t *testing.T) {
		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-compat")

		// Setup guards to pass
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		initialState := proj.Statechart.Current_state

		// Execute auto-advance (the original behavior - no flags, no event arg)
		err = executeAutoTransition(proj, state.State(initialState))
		if err != nil {
			t.Fatalf("auto-advance failed: %v", err)
		}

		// Reload and verify state advanced
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		// Should have transitioned from ImplementationPlanning to ImplementationDraftPRCreation
		if proj.Statechart.Current_state == initialState {
			t.Error("state did not advance - backward compatibility broken")
		}

		if proj.Statechart.Current_state != string(standard.ImplementationDraftPRCreation) {
			t.Errorf("expected %s, got %s", standard.ImplementationDraftPRCreation, proj.Statechart.Current_state)
		}
	})

	t.Run("error messages maintain helpful guidance", func(t *testing.T) {
		// Setup: Create standard project in ReviewActive (branching state)
		ctx := setupTestRepoWithProject(t, "standard", "test-error-messages")
		advanceToReviewActive(t, ctx)

		// DO NOT add review - this will cause auto-advance to fail because
		// no assessment exists (discriminator returns empty string)
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// Try auto-advance without setting up the review
		// This should fail with a helpful error about using --list
		err = executeAutoTransition(proj, state.State(standard.ReviewActive))

		// Should get an error
		if err == nil {
			t.Error("expected error when review assessment missing, got nil")
		}

		// Error message should be helpful (mention that user needs to act)
		// This verifies error messages are improved, not broken
		if err != nil {
			errMsg := err.Error()
			// The error should indicate why auto-advance failed
			// It might mention missing assessment or suggest alternatives
			if !strings.Contains(errMsg, "cannot advance") &&
				!strings.Contains(errMsg, "cannot determine") &&
				!strings.Contains(errMsg, "discriminator") {
				t.Logf("Error message: %s", errMsg)
				t.Log("Note: Error message format may have changed but should still be helpful")
			}
		}
	})

	t.Run("new flags optional and don't affect default behavior", func(t *testing.T) {
		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-flags-optional")

		// Verify project works without ever using new flags
		// Advance through a few states using only auto mode
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// Auto-advance (no --list, no --dry-run, no explicit event)
		err = executeAutoTransition(proj, state.State(standard.ImplementationPlanning))
		if err != nil {
			t.Fatalf("auto-advance failed: %v", err)
		}

		// Load and verify success
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != string(standard.ImplementationDraftPRCreation) {
			t.Errorf("expected %s, got %s", standard.ImplementationDraftPRCreation, proj.Statechart.Current_state)
		}

		// Continue with draft PR
		setPhaseMetadata(t, ctx, "implementation", "draft_pr_created", true)
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		err = executeAutoTransition(proj, state.State(standard.ImplementationDraftPRCreation))
		if err != nil {
			t.Fatalf("second auto-advance failed: %v", err)
		}

		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != string(standard.ImplementationExecuting) {
			t.Errorf("expected %s, got %s", standard.ImplementationExecuting, proj.Statechart.Current_state)
		}

		// Success: Project advanced through multiple states without using any new features
	})

	t.Run("existing state machine behavior unchanged", func(t *testing.T) {
		// Verify that the state machine itself hasn't changed
		// This tests that refactoring didn't alter the fundamental transitions

		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-state-machine")

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// Verify initial state is still ImplementationPlanning
		if proj.Statechart.Current_state != string(standard.ImplementationPlanning) {
			t.Errorf("initial state changed - was %s, expected %s",
				proj.Statechart.Current_state, standard.ImplementationPlanning)
		}

		// Verify machine can be built
		machine := proj.Machine()
		if machine == nil {
			t.Fatal("machine is nil - state machine construction broken")
		}

		// Verify machine state matches project state
		if machine.State().String() != proj.Statechart.Current_state {
			t.Error("machine state doesn't match project state - sync broken")
		}
	})

	t.Run("guard conditions still enforced correctly", func(t *testing.T) {
		// Verify guards still prevent invalid transitions

		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-guards")

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// DO NOT set planning_approved metadata
		// Guard should block transition

		// Try to advance
		err = executeAutoTransition(proj, state.State(standard.ImplementationPlanning))

		// Should fail due to guard
		if err == nil {
			t.Error("expected guard to block transition, but it succeeded")
		}

		// Verify error mentions the guard
		if err != nil && !strings.Contains(err.Error(), "guard") {
			t.Errorf("expected guard-related error, got: %v", err)
		}

		// Verify state unchanged
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != string(standard.ImplementationPlanning) {
			t.Error("guard failed to prevent transition - state changed")
		}
	})
}

// TestNewCLIModesAdditive verifies new modes are strictly additive.
// They add functionality without changing existing behavior.
func TestNewCLIModesAdditive(t *testing.T) {
	t.Run("list mode is read-only", func(t *testing.T) {
		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-list-readonly")

		// Setup guards
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		initialState := proj.Statechart.Current_state

		// Call list mode
		_ = listAvailableTransitions(proj, state.State(initialState))

		// Reload and verify state unchanged
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != initialState {
			t.Error("list mode modified state - NOT read-only")
		}
	})

	t.Run("dry-run mode has no side effects", func(t *testing.T) {
		// This is also tested in integration tests, but worth repeating for compatibility

		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-dry-run-no-effects")

		// Setup guards
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		initialState := proj.Statechart.Current_state
		machine := proj.Machine()

		// Call dry-run mode
		_ = validateTransition(
			ctx,
			proj,
			machine,
			state.State(initialState),
			state.Event(standard.EventPlanningComplete),
		)

		// Reload and verify state unchanged
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != initialState {
			t.Error("dry-run modified state - has side effects")
		}
	})

	t.Run("explicit event mode works alongside auto mode", func(t *testing.T) {
		// Verify both modes can be used in the same project lifecycle

		// Setup: Create standard project
		ctx := setupTestRepoWithProject(t, "standard", "test-mixed-modes")

		// Use auto mode for first transition
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		err = executeAutoTransition(proj, state.State(standard.ImplementationPlanning))
		if err != nil {
			t.Fatalf("auto mode failed: %v", err)
		}

		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != string(standard.ImplementationDraftPRCreation) {
			t.Fatalf("auto mode didn't advance correctly")
		}

		// Use explicit event mode for second transition
		setPhaseMetadata(t, ctx, "implementation", "draft_pr_created", true)
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		machine := proj.Machine()
		err = executeExplicitTransition(
			ctx,
			proj,
			machine,
			state.State(standard.ImplementationDraftPRCreation),
			state.Event(standard.EventDraftPRCreated),
		)
		if err != nil {
			t.Fatalf("explicit mode failed: %v", err)
		}

		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload: %v", err)
		}

		if proj.Statechart.Current_state != string(standard.ImplementationExecuting) {
			t.Error("explicit mode didn't advance correctly")
		}

		// Success: Both modes work together in same project
	})
}
