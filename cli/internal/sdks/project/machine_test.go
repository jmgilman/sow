package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	stateMachine "github.com/jmgilman/sow/cli/internal/sdks/state"
	projectSchemas "github.com/jmgilman/sow/cli/schemas/project"
)

// TestBuildMachineCreatesInitializedMachine tests that BuildMachine creates a state machine
// initialized with the given initial state.
func TestBuildMachineCreatesInitializedMachine(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		Build()

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{},
	}

	initialState := stateMachine.State("test_state")
	machine := config.BuildMachine(proj, initialState)

	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}

	if machine.State() != initialState {
		t.Errorf("expected initial state %s, got %s", initialState, machine.State())
	}
}

// TestBuildMachineAddsTransitions tests that all transitions from the config
// are added to the machine.
func TestBuildMachineAddsTransitions(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("middle"),
			stateMachine.Event("advance"),
		).
		AddTransition(
			stateMachine.State("middle"),
			stateMachine.State("end"),
			stateMachine.Event("complete"),
		).
		Build()

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{},
	}

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	// Test first transition
	can, err := machine.CanFire(stateMachine.Event("advance"))
	if err != nil {
		t.Fatalf("CanFire(advance) failed: %v", err)
	}
	if !can {
		t.Error("expected transition 'advance' to be possible from 'start'")
	}

	// Fire first transition
	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire(advance) failed: %v", err)
	}

	if machine.State() != stateMachine.State("middle") {
		t.Errorf("expected state 'middle', got %s", machine.State())
	}

	// Test second transition
	can, err = machine.CanFire(stateMachine.Event("complete"))
	if err != nil {
		t.Fatalf("CanFire(complete) failed: %v", err)
	}
	if !can {
		t.Error("expected transition 'complete' to be possible from 'middle'")
	}
}

// TestBuildMachineGuardsAccessProjectState tests that guards are properly bound
// and can access live project state.
func TestBuildMachineGuardsAccessProjectState(t *testing.T) {
	// Create project with known state
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"test": {
					Outputs: []projectSchemas.ArtifactState{
						{Type: "result", Approved: true},
					},
				},
			},
		},
	}

	// Create config with guard that checks project state
	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			WithGuard(func(p *state.Project) bool {
				return p.PhaseOutputApproved("test", "result")
			}),
		).
		Build()

	// Build machine
	machine := config.BuildMachine(proj, stateMachine.State("start"))

	// Verify guard works (should pass)
	can, err := machine.CanFire(stateMachine.Event("advance"))
	if err != nil {
		t.Fatalf("CanFire failed: %v", err)
	}
	if !can {
		t.Error("expected transition to be allowed (guard should pass)")
	}

	// Change project state
	proj.Phases["test"].Outputs[0].Approved = false

	// Verify guard reflects new state (should fail)
	can, err = machine.CanFire(stateMachine.Event("advance"))
	if err != nil {
		t.Fatalf("CanFire failed: %v", err)
	}
	if can {
		t.Error("expected transition to be blocked (guard should fail)")
	}
}

// TestBuildMachineOnEntryActions tests that onEntry actions are properly bound
// and execute when entering a state.
func TestBuildMachineOnEntryActions(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"test": {Status: "pending"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			WithOnEntry(func(p *state.Project) error {
				phase := p.Phases["test"]
				phase.Status = "active"
				p.Phases["test"] = phase
				return nil
			}),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	if proj.Phases["test"].Status != "active" {
		t.Errorf("expected onEntry to set status=active, got %s", proj.Phases["test"].Status)
	}
}

// TestBuildMachineOnExitActions tests that onExit actions are properly bound
// and execute when leaving a state.
func TestBuildMachineOnExitActions(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"test": {Status: "active"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			WithOnExit(func(p *state.Project) error {
				phase := p.Phases["test"]
				phase.Status = "completed"
				p.Phases["test"] = phase
				return nil
			}),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	if proj.Phases["test"].Status != "completed" {
		t.Errorf("expected onExit to set status=completed, got %s", proj.Phases["test"].Status)
	}
}

// TestBuildMachineTransitionsWithoutGuards tests that transitions without guards
// are always allowed (unconditional transitions).
func TestBuildMachineTransitionsWithoutGuards(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			// No guard - should always be allowed
		).
		Build()

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{},
	}

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	can, err := machine.CanFire(stateMachine.Event("advance"))
	if err != nil {
		t.Fatalf("CanFire failed: %v", err)
	}
	if !can {
		t.Error("expected transition without guard to always be allowed")
	}

	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	if machine.State() != stateMachine.State("end") {
		t.Errorf("expected state 'end', got %s", machine.State())
	}
}

// TestBuildMachineTransitionsWithoutActions tests that transitions without actions
// work correctly (actions are optional).
func TestBuildMachineTransitionsWithoutActions(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			// No onEntry or onExit actions
		).
		Build()

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{},
	}

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire failed for transition without actions: %v", err)
	}

	if machine.State() != stateMachine.State("end") {
		t.Errorf("expected state 'end', got %s", machine.State())
	}
}

// TestBuildMachineGuardBlocksTransition tests that when a guard returns false,
// the transition is blocked.
func TestBuildMachineGuardBlocksTransition(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			WithGuard(func(_ *state.Project) bool {
				return false // Always block
			}),
		).
		Build()

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{},
	}

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	can, err := machine.CanFire(stateMachine.Event("advance"))
	if err != nil {
		t.Fatalf("CanFire failed: %v", err)
	}
	if can {
		t.Error("expected guard to block transition")
	}

	// Attempting to fire should fail
	err = machine.Fire(stateMachine.Event("advance"))
	if err == nil {
		t.Error("expected Fire to fail when guard blocks transition")
	}
}

// TestBuildMachineCombinedGuardAndActions tests a transition with guard, onEntry,
// and onExit actions all working together.
func TestBuildMachineCombinedGuardAndActions(t *testing.T) {
	var exitCalled, entryCalled bool

	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"test": {
					Outputs: []projectSchemas.ArtifactState{
						{Type: "result", Approved: true},
					},
				},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		AddTransition(
			stateMachine.State("start"),
			stateMachine.State("end"),
			stateMachine.Event("advance"),
			WithGuard(func(p *state.Project) bool {
				return p.PhaseOutputApproved("test", "result")
			}),
			WithOnExit(func(_ *state.Project) error {
				exitCalled = true
				return nil
			}),
			WithOnEntry(func(_ *state.Project) error {
				entryCalled = true
				return nil
			}),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("start"))

	if err := machine.Fire(stateMachine.Event("advance")); err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	if !exitCalled {
		t.Error("onExit was not called")
	}

	if !entryCalled {
		t.Error("onEntry was not called")
	}

	if machine.State() != stateMachine.State("end") {
		t.Errorf("expected state 'end', got %s", machine.State())
	}
}

// TestBuildMachineAutomaticPhaseInProgress tests that entering a phase's start state
// automatically marks the phase as "in_progress" (if currently "pending").
func TestBuildMachineAutomaticPhaseInProgress(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"implementation": {Status: "pending"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		WithPhase("implementation",
			WithStartState(stateMachine.State("planning")),
			WithEndState(stateMachine.State("executing")),
		).
		AddTransition(
			stateMachine.State("init"),
			stateMachine.State("planning"),
			stateMachine.Event("start"),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("init"))

	// Verify phase starts as pending
	if proj.Phases["implementation"].Status != "pending" {
		t.Errorf("expected initial status=pending, got %s", proj.Phases["implementation"].Status)
	}

	// Fire transition with phase updates to enter phase's start state
	if err := config.FireWithPhaseUpdates(machine, stateMachine.Event("start"), proj); err != nil {
		t.Fatalf("FireWithPhaseUpdates failed: %v", err)
	}

	// Verify phase was automatically marked in_progress
	if proj.Phases["implementation"].Status != "in_progress" {
		t.Errorf("expected automatic status=in_progress, got %s", proj.Phases["implementation"].Status)
	}
}

// TestBuildMachineAutomaticPhaseCompleted tests that exiting a phase's end state
// automatically marks the phase as "completed" with a timestamp.
func TestBuildMachineAutomaticPhaseCompleted(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"implementation": {Status: "in_progress"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		WithPhase("implementation",
			WithStartState(stateMachine.State("planning")),
			WithEndState(stateMachine.State("executing")),
		).
		AddTransition(
			stateMachine.State("executing"),
			stateMachine.State("review"),
			stateMachine.Event("complete"),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("executing"))

	// Fire transition with phase updates to exit phase's end state
	if err := config.FireWithPhaseUpdates(machine, stateMachine.Event("complete"), proj); err != nil {
		t.Fatalf("FireWithPhaseUpdates failed: %v", err)
	}

	// Verify phase was automatically marked completed
	if proj.Phases["implementation"].Status != "completed" {
		t.Errorf("expected automatic status=completed, got %s", proj.Phases["implementation"].Status)
	}

	// Verify completed_at timestamp was set
	if proj.Phases["implementation"].Completed_at.IsZero() {
		t.Error("expected completed_at timestamp to be set")
	}
}

// TestBuildMachineAutomaticPhaseActionsBeforeUserActions tests that automatic phase
// status actions run after machine Fire but before user accesses the phase status.
func TestBuildMachineAutomaticPhaseActionsBeforeUserActions(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"implementation": {Status: "pending"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		WithPhase("implementation",
			WithStartState(stateMachine.State("planning")),
			WithEndState(stateMachine.State("executing")),
		).
		AddTransition(
			stateMachine.State("init"),
			stateMachine.State("planning"),
			stateMachine.Event("start"),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("init"))

	// Status should be pending before transition
	if proj.Phases["implementation"].Status != "pending" {
		t.Errorf("expected initial status=pending, got %s", proj.Phases["implementation"].Status)
	}

	if err := config.FireWithPhaseUpdates(machine, stateMachine.Event("start"), proj); err != nil {
		t.Fatalf("FireWithPhaseUpdates failed: %v", err)
	}

	// Status should be in_progress after transition (automatic update happened)
	if proj.Phases["implementation"].Status != "in_progress" {
		t.Errorf("expected automatic update to set status=in_progress, got %s", proj.Phases["implementation"].Status)
	}
}

// TestBuildMachinePhaseInProgressOnlyWhenPending tests that the automatic in_progress
// action only updates the status if currently "pending" (preserves manual overrides).
func TestBuildMachinePhaseInProgressOnlyWhenPending(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"implementation": {Status: "completed"}, // Already completed
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		WithPhase("implementation",
			WithStartState(stateMachine.State("planning")),
			WithEndState(stateMachine.State("executing")),
		).
		AddTransition(
			stateMachine.State("init"),
			stateMachine.State("planning"),
			stateMachine.Event("start"),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("init"))

	// Fire transition to enter phase's start state
	if err := machine.Fire(stateMachine.Event("start")); err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	// Verify status remains "completed" (not overwritten by automatic action)
	if proj.Phases["implementation"].Status != "completed" {
		t.Errorf("expected status to remain completed, got %s", proj.Phases["implementation"].Status)
	}
}

// TestBuildMachineMultiplePhaseTransitions tests that phase statuses are correctly
// updated across multiple phase transitions.
func TestBuildMachineMultiplePhaseTransitions(t *testing.T) {
	proj := &state.Project{
		ProjectState: projectSchemas.ProjectState{
			Phases: map[string]projectSchemas.PhaseState{
				"implementation": {Status: "pending"},
				"review":         {Status: "pending"},
			},
		},
	}

	config := NewProjectTypeConfigBuilder("test").
		WithPhase("implementation",
			WithStartState(stateMachine.State("planning")),
			WithEndState(stateMachine.State("executing")),
		).
		WithPhase("review",
			WithStartState(stateMachine.State("reviewing")),
			WithEndState(stateMachine.State("approved")),
		).
		AddTransition(
			stateMachine.State("init"),
			stateMachine.State("planning"),
			stateMachine.Event("start_impl"),
		).
		AddTransition(
			stateMachine.State("executing"),
			stateMachine.State("reviewing"),
			stateMachine.Event("start_review"),
		).
		Build()

	machine := config.BuildMachine(proj, stateMachine.State("init"))

	// Start implementation phase
	if err := config.FireWithPhaseUpdates(machine, stateMachine.Event("start_impl"), proj); err != nil {
		t.Fatalf("FireWithPhaseUpdates(start_impl) failed: %v", err)
	}

	if proj.Phases["implementation"].Status != "in_progress" {
		t.Errorf("expected implementation status=in_progress, got %s", proj.Phases["implementation"].Status)
	}

	if proj.Phases["review"].Status != "pending" {
		t.Errorf("expected review status=pending, got %s", proj.Phases["review"].Status)
	}

	// Transition to executing (still in implementation phase)
	machine = config.BuildMachine(proj, stateMachine.State("executing"))

	// Start review phase (exits implementation end state, enters review start state)
	if err := config.FireWithPhaseUpdates(machine, stateMachine.Event("start_review"), proj); err != nil {
		t.Fatalf("FireWithPhaseUpdates(start_review) failed: %v", err)
	}

	// Verify implementation completed
	if proj.Phases["implementation"].Status != "completed" {
		t.Errorf("expected implementation status=completed, got %s", proj.Phases["implementation"].Status)
	}

	// Verify review started
	if proj.Phases["review"].Status != "in_progress" {
		t.Errorf("expected review status=in_progress, got %s", proj.Phases["review"].Status)
	}
}
