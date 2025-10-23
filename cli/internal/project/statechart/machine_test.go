package statechart

import (
	"os"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// testMachine creates a machine with prompts suppressed for cleaner test output.
func testMachine(state *schemas.ProjectState) *Machine {
	m := NewMachine(state)
	m.SuppressPrompts(true)
	return m
}

// TestProjectLifecycle demonstrates a complete project lifecycle through the state machine.
func TestProjectLifecycle(t *testing.T) {
	// Start with no project
	state := &schemas.ProjectState{
		Phases: struct {
			Discovery      phases.Phase `json:"discovery"`
			Design         phases.Phase `json:"design"`
			Implementation phases.Phase `json:"implementation"`
			Review         phases.Phase `json:"review"`
			Finalize       phases.Phase `json:"finalize"`
		}{
			Discovery: phases.Phase{
				Enabled: false,
				Status:  "skipped",
			},
			Design: phases.Phase{
				Enabled: false,
				Status:  "skipped",
			},
			Implementation: phases.Phase{
				Enabled: true,
				Status:  "pending",
			},
			Review: phases.Phase{
				Enabled:  true,
				Status:   "pending",
				Metadata: map[string]interface{}{"iteration": 1},
			},
			Finalize: phases.Phase{
				Enabled:  true,
				Status:   "pending",
				Metadata: map[string]interface{}{"project_deleted": false},
			},
		},
	}

	machine := testMachine(nil) // Start with NoProject

	// Verify initial state
	if machine.State() != NoProject {
		t.Errorf("Expected initial state NoProject, got %s", machine.State())
	}

	// Step 1: Initialize project
	if err := machine.Fire(EventProjectInit); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}
	if machine.State() != DiscoveryDecision {
		t.Errorf("Expected DiscoveryDecision after init, got %s", machine.State())
	}

	// Step 2: Skip discovery (implicit transition)
	if err := machine.Fire(EventSkipDiscovery); err != nil {
		t.Fatalf("Failed to skip discovery: %v", err)
	}
	if machine.State() != DesignDecision {
		t.Errorf("Expected DesignDecision after skipping discovery, got %s", machine.State())
	}

	// Step 3: Skip design (implicit transition)
	if err := machine.Fire(EventSkipDesign); err != nil {
		t.Fatalf("Failed to skip design: %v", err)
	}
	if machine.State() != ImplementationPlanning {
		t.Errorf("Expected ImplementationPlanning after skipping design, got %s", machine.State())
	}

	// Step 4: Create tasks and transition to executing
	// Update the machine's project state to have at least one task
	machine.projectState = state
	state.Phases.Implementation.Tasks = []phases.Task{
		{Id: "010", Name: "Create model", Status: "pending", Parallel: false},
	}

	// Approve tasks to transition to executing
	if state.Phases.Implementation.Metadata == nil {
		state.Phases.Implementation.Metadata = make(map[string]interface{})
	}
	state.Phases.Implementation.Metadata["tasks_approved"] = true
	if err := machine.Fire(EventTasksApproved); err != nil {
		t.Fatalf("Failed to transition to executing: %v", err)
	}
	if machine.State() != ImplementationExecuting {
		t.Errorf("Expected ImplementationExecuting after task approval, got %s", machine.State())
	}

	// Step 5: Complete all tasks and transition to review
	state.Phases.Implementation.Tasks[0].Status = "completed"

	if err := machine.Fire(EventAllTasksComplete); err != nil {
		t.Fatalf("Failed to transition to review: %v", err)
	}
	if machine.State() != ReviewActive {
		t.Errorf("Expected ReviewActive after tasks complete, got %s", machine.State())
	}

	// Step 6: Review passes
	// Add review report as artifact and approve it
	state.Phases.Review.Artifacts = []phases.Artifact{
		{
			Path:     "reports/001.md",
			Approved: true,
			Metadata: map[string]interface{}{"type": "review", "assessment": "pass"},
		},
	}
	if err := machine.Fire(EventReviewPass); err != nil {
		t.Fatalf("Failed to pass review: %v", err)
	}
	if machine.State() != FinalizeDocumentation {
		t.Errorf("Expected FinalizeDocumentation after review pass, got %s", machine.State())
	}

	// Step 7: Documentation assessed (simplified - just update status)
	state.Phases.Finalize.Status = "in_progress"

	if err := machine.Fire(EventDocumentationDone); err != nil {
		t.Fatalf("Failed to complete documentation: %v", err)
	}
	if machine.State() != FinalizeChecks {
		t.Errorf("Expected FinalizeChecks after documentation, got %s", machine.State())
	}

	// Step 8: Checks assessed (guards return true automatically)
	if err := machine.Fire(EventChecksDone); err != nil {
		t.Fatalf("Failed to complete checks: %v", err)
	}
	if machine.State() != FinalizeDelete {
		t.Errorf("Expected FinalizeDelete after checks, got %s", machine.State())
	}

	// Step 9: Delete project
	if state.Phases.Finalize.Metadata == nil {
		state.Phases.Finalize.Metadata = make(map[string]interface{})
	}
	state.Phases.Finalize.Metadata["project_deleted"] = true

	if err := machine.Fire(EventProjectDelete); err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}
	if machine.State() != NoProject {
		t.Errorf("Expected NoProject after deletion, got %s", machine.State())
	}
}

// TestDiscoveryPhase tests the discovery phase workflow.
func TestDiscoveryPhase(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Phases.Discovery = phases.Phase{
		Enabled: true,
		Status:  "pending",
		Artifacts: []phases.Artifact{
			{Path: "phases/discovery/notes.md", Approved: false},
		},
	}

	machine := testMachine(state)

	// Initialize and enter discovery
	_ = machine.Fire(EventProjectInit)
	_ = machine.Fire(EventEnableDiscovery)

	if machine.State() != DiscoveryActive {
		t.Errorf("Expected DiscoveryActive, got %s", machine.State())
	}

	// Try to complete discovery without approvals - should fail
	if err := machine.Fire(EventCompleteDiscovery); err == nil {
		t.Error("Expected error completing discovery with unapproved artifacts")
	}

	// Approve artifact
	state.Phases.Discovery.Artifacts[0].Approved = true

	// Now should succeed
	if err := machine.Fire(EventCompleteDiscovery); err != nil {
		t.Fatalf("Failed to complete discovery with approved artifacts: %v", err)
	}

	if machine.State() != DesignDecision {
		t.Errorf("Expected DesignDecision after discovery complete, got %s", machine.State())
	}
}

// TestReviewLoop tests the review fail → implementation loop.
func TestReviewLoop(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Phases.Implementation = phases.Phase{
		Enabled: true,
		Status:  "completed",
		Tasks: []phases.Task{
			{Id: "010", Name: "Fix bug", Status: "completed", Parallel: false},
		},
	}
	state.Phases.Review = phases.Phase{
		Enabled:  true,
		Status:   "pending",
		Metadata: map[string]interface{}{"iteration": 1},
	}

	machine := testMachine(state)

	// Fast-forward to review state (simulate getting there)
	_ = machine.Fire(EventProjectInit)
	_ = machine.Fire(EventSkipDiscovery)
	_ = machine.Fire(EventSkipDesign)
	machine.projectState = state
	if state.Phases.Implementation.Metadata == nil {
		state.Phases.Implementation.Metadata = make(map[string]interface{})
	}
	state.Phases.Implementation.Metadata["tasks_approved"] = true
	_ = machine.Fire(EventTasksApproved)
	_ = machine.Fire(EventAllTasksComplete)

	if machine.State() != ReviewActive {
		t.Fatalf("Expected ReviewActive, got %s", machine.State())
	}

	// Review fails - add review report as artifact and approve it
	state.Phases.Review.Artifacts = []phases.Artifact{
		{
			Path:     "reports/001.md",
			Approved: true,
			Metadata: map[string]interface{}{"type": "review", "assessment": "fail"},
		},
	}
	if err := machine.Fire(EventReviewFail); err != nil {
		t.Fatalf("Failed to loop back to implementation: %v", err)
	}

	if machine.State() != ImplementationPlanning {
		t.Errorf("Expected ImplementationPlanning after review fail, got %s", machine.State())
	}

	// Add new task (or proceed with existing tasks)
	// Since tasks already exist, guard will pass after approval
	if state.Phases.Implementation.Metadata == nil {
		state.Phases.Implementation.Metadata = make(map[string]interface{})
	}
	state.Phases.Implementation.Metadata["tasks_approved"] = true
	if err := machine.Fire(EventTasksApproved); err != nil {
		t.Fatalf("Failed to transition to executing: %v", err)
	}

	if machine.State() != ImplementationExecuting {
		t.Errorf("Expected ImplementationExecuting after task created, got %s", machine.State())
	}

	// Complete tasks again and transition back to review
	if state.Phases.Review.Metadata == nil {
		state.Phases.Review.Metadata = make(map[string]interface{})
	}
	state.Phases.Review.Metadata["iteration"] = 2
	if err := machine.Fire(EventAllTasksComplete); err != nil {
		t.Fatalf("Failed to return to review: %v", err)
	}

	if machine.State() != ReviewActive {
		t.Errorf("Expected ReviewActive on second iteration, got %s", machine.State())
	}
}

// TestGuardPreventsInvalidTransition tests that guards properly block transitions.
func TestGuardPreventsInvalidTransition(t *testing.T) {
	state := &schemas.ProjectState{}
	state.Phases.Implementation = phases.Phase{
		Enabled: true,
		Status:  "pending",
		Tasks:   []phases.Task{}, // No tasks
	}

	machine := testMachine(state)

	// Fast-forward to implementation planning
	_ = machine.Fire(EventProjectInit)
	_ = machine.Fire(EventSkipDiscovery)
	_ = machine.Fire(EventSkipDesign)

	if machine.State() != ImplementationPlanning {
		t.Fatalf("Expected ImplementationPlanning, got %s", machine.State())
	}

	// Try to transition without any tasks - should fail
	machine.projectState = state
	if err := machine.Fire(EventTaskCreated); err == nil {
		t.Error("Expected error transitioning without tasks, but succeeded")
	}

	// Should still be in planning state
	if machine.State() != ImplementationPlanning {
		t.Errorf("State should not have changed, got %s", machine.State())
	}
}

// TestPermittedTriggers verifies which events are valid in each state.
func TestPermittedTriggers(t *testing.T) {
	machine := testMachine(nil)

	// NoProject should only permit ProjectInit
	triggers, err := machine.PermittedTriggers()
	if err != nil {
		t.Fatalf("Failed to get permitted triggers: %v", err)
	}
	if len(triggers) != 1 || triggers[0] != EventProjectInit {
		t.Errorf("NoProject should only permit EventProjectInit, got %v", triggers)
	}

	// Move to DiscoveryDecision
	_ = machine.Fire(EventProjectInit)

	triggers, err = machine.PermittedTriggers()
	if err != nil {
		t.Fatalf("Failed to get permitted triggers: %v", err)
	}
	hasEnableDiscovery := false
	hasSkipDiscovery := false
	for _, trigger := range triggers {
		if trigger == EventEnableDiscovery {
			hasEnableDiscovery = true
		}
		if trigger == EventSkipDiscovery {
			hasSkipDiscovery = true
		}
	}

	if !hasEnableDiscovery || !hasSkipDiscovery {
		t.Errorf("DiscoveryDecision should permit both enable and skip, got %v", triggers)
	}
}

// TestPersistence tests saving and loading state from disk.
func TestPersistence(t *testing.T) {
	// Clean up any existing state file
	defer func() {
		_ = os.Remove(stateFilePath)
		_ = os.RemoveAll(".sow")
	}()

	// Create a machine and advance through some states
	machine := testMachine(nil)
	state := &schemas.ProjectState{}
	state.Phases.Implementation = phases.Phase{
		Enabled: true,
		Status:  "pending",
		Tasks: []phases.Task{
			{Id: "010", Name: "Test task", Status: "pending", Parallel: false},
		},
	}
	machine.projectState = state

	// Advance to ImplementationExecuting
	_ = machine.Fire(EventProjectInit)
	_ = machine.Fire(EventSkipDiscovery)
	_ = machine.Fire(EventSkipDesign)
	_ = machine.Fire(EventTaskCreated)

	if machine.State() != ImplementationExecuting {
		t.Fatalf("Expected ImplementationExecuting, got %s", machine.State())
	}

	// Save state
	if err := machine.Save(); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load state back
	loadedMachine, err := LoadFS(nil)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	loadedMachine.SuppressPrompts(true)

	// Verify state was preserved
	if loadedMachine.State() != ImplementationExecuting {
		t.Errorf("Expected loaded state to be ImplementationExecuting, got %s", loadedMachine.State())
	}

	// Verify project state was preserved
	if loadedMachine.projectState == nil {
		t.Fatal("Expected project state to be preserved")
	}

	if len(loadedMachine.projectState.Phases.Implementation.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedMachine.projectState.Phases.Implementation.Tasks))
	}

	if loadedMachine.projectState.Phases.Implementation.Tasks[0].Name != "Test task" {
		t.Errorf("Expected task name 'Test task', got '%s'",
			loadedMachine.projectState.Phases.Implementation.Tasks[0].Name)
	}
}

// TestLoadNoProject tests loading when no project exists.
func TestLoadNoProject(t *testing.T) {
	// Ensure no state file exists
	_ = os.Remove(stateFilePath)
	defer func() {
		_ = os.RemoveAll(".sow")
	}()

	// Load should succeed and return NoProject state
	machine, err := LoadFS(nil)
	if err != nil {
		t.Fatalf("Failed to load (expected NoProject): %v", err)
	}
	machine.SuppressPrompts(true)

	if machine.State() != NoProject {
		t.Errorf("Expected NoProject state when no file exists, got %s", machine.State())
	}

	if machine.projectState != nil {
		t.Error("Expected nil project state when no file exists")
	}
}
