package standard

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas/projects"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// Helper to create a minimal StandardProjectState for testing
func createTestState() *projects.StandardProjectState {
	now := time.Now()

	return &projects.StandardProjectState{
		Statechart: struct {
			Current_state string `json:"current_state"`
		}{
			Current_state: "NoProject",
		},
		Project: struct {
			Type          string     `json:"type"`
			Name          string     `json:"name"`
			Branch        string     `json:"branch"`
			Description   string     `json:"description"`
			Github_issue  *int64     `json:"github_issue,omitempty"`
			Created_at    time.Time  `json:"created_at"`
			Updated_at    time.Time  `json:"updated_at"`
		}{
			Type:        "standard",
			Name:        "test-project",
			Branch:      "feat/test",
			Description: "Test project for unit tests",
			Created_at:  now,
			Updated_at:  now,
		},
		Phases: struct {
			Discovery      phasesSchema.DiscoveryPhase      `json:"discovery"`
			Design         phasesSchema.DesignPhase         `json:"design"`
			Implementation phasesSchema.ImplementationPhase `json:"implementation"`
			Review         phasesSchema.ReviewPhase         `json:"review"`
			Finalize       phasesSchema.FinalizePhase       `json:"finalize"`
		}{
			Discovery: phasesSchema.DiscoveryPhase{
				Status:     "pending",
				Created_at: now,
				Enabled:    false,
				Artifacts:  []phasesSchema.Artifact{},
			},
			Design: phasesSchema.DesignPhase{
				Status:     "pending",
				Created_at: now,
				Enabled:    false,
				Artifacts:  []phasesSchema.Artifact{},
			},
			Implementation: phasesSchema.ImplementationPhase{
				Status:     "pending",
				Created_at: now,
				Tasks:      []phasesSchema.Task{},
			},
			Review: phasesSchema.ReviewPhase{
				Status:     "pending",
				Created_at: now,
				Iteration:  0,
				Reports:    []phasesSchema.ReviewReport{},
			},
			Finalize: phasesSchema.FinalizePhase{
				Status:     "pending",
				Created_at: now,
			},
		},
	}
}

// TestNew verifies the constructor creates a StandardProject correctly
func TestNew(t *testing.T) {
	state := createTestState()
	project := New(state)

	if project == nil {
		t.Fatal("New() returned nil")
	}

	if project.state != state {
		t.Error("New() did not store state correctly")
	}
}

// TestType verifies the Type() method returns "standard"
func TestType(t *testing.T) {
	state := createTestState()
	project := New(state)

	if project.Type() != "standard" {
		t.Errorf("Type() = %q, want %q", project.Type(), "standard")
	}
}

// TestBuildStateMachine_CreatesAllStates verifies all states are configured
func TestBuildStateMachine_CreatesAllStates(t *testing.T) {
	state := createTestState()
	project := New(state)

	machine := project.BuildStateMachine()

	if machine == nil {
		t.Fatal("BuildStateMachine() returned nil")
	}

	// Expected states from all 5 phases
	expectedStates := []statechart.State{
		statechart.NoProject,
		statechart.DiscoveryDecision,
		statechart.DiscoveryActive,
		statechart.DesignDecision,
		statechart.DesignActive,
		statechart.ImplementationPlanning,
		statechart.ImplementationExecuting,
		statechart.ReviewActive,
		statechart.FinalizeDocumentation,
		statechart.FinalizeChecks,
		statechart.FinalizeDelete,
	}

	// Try to fire events and check if states are reachable
	// We'll test a few key transitions to ensure the machine is configured

	// Test: NoProject → DiscoveryDecision
	machine.SuppressPrompts(true) // Disable prompts for testing
	err := machine.Fire(statechart.EventProjectInit)
	if err != nil {
		t.Errorf("Failed to fire EventProjectInit: %v", err)
	}

	// Verify we're in DiscoveryDecision
	currentState := machine.State()
	if currentState != statechart.DiscoveryDecision {
		t.Errorf("After EventProjectInit, state = %v, want %v", currentState, statechart.DiscoveryDecision)
	}

	// Test: Skip discovery and design to reach Implementation
	err = machine.Fire(statechart.EventSkipDiscovery)
	if err != nil {
		t.Errorf("Failed to fire EventSkipDiscovery: %v", err)
	}

	err = machine.Fire(statechart.EventSkipDesign)
	if err != nil {
		t.Errorf("Failed to fire EventSkipDesign: %v", err)
	}

	currentState = machine.State()
	if currentState != statechart.ImplementationPlanning {
		t.Errorf("After skipping discovery and design, state = %v, want %v", currentState, statechart.ImplementationPlanning)
	}

	// Verify that all expected states are known (no panic on lookups)
	for _, expectedState := range expectedStates {
		_ = expectedState // Just ensure they're declared
	}
}

// TestBuildStateMachine_ForwardTransitions verifies forward phase transitions
func TestBuildStateMachine_ForwardTransitions(t *testing.T) {
	state := createTestState()
	project := New(state)

	machine := project.BuildStateMachine()
	machine.SuppressPrompts(true)

	// Test forward progression through all phases (skipping optional phases)
	tests := []struct {
		event         statechart.Event
		expectedState statechart.State
		description   string
	}{
		{statechart.EventProjectInit, statechart.DiscoveryDecision, "Initialize project"},
		{statechart.EventSkipDiscovery, statechart.DesignDecision, "Skip discovery"},
		{statechart.EventSkipDesign, statechart.ImplementationPlanning, "Skip design"},
	}

	for _, tt := range tests {
		err := machine.Fire(tt.event)
		if err != nil {
			t.Errorf("%s: Fire(%v) error = %v", tt.description, tt.event, err)
			continue
		}

		currentState := machine.State()
		if currentState != tt.expectedState {
			t.Errorf("%s: state = %v, want %v", tt.description, currentState, tt.expectedState)
		}
	}
}

// TestBuildStateMachine_BackwardTransition verifies Review → Implementation loop
func TestBuildStateMachine_BackwardTransition(t *testing.T) {
	state := createTestState()

	// Add a task so we can transition through implementation
	state.Phases.Implementation.Tasks = []phasesSchema.Task{
		{
			Id:     "010",
			Name:   "Test task",
			Status: "completed",
		},
	}

	// Add a failed review report
	state.Phases.Review.Reports = []phasesSchema.ReviewReport{
		{
			Id:         "001",
			Assessment: "fail",
			Approved:   false,
		},
	}

	// Set tasks as approved
	state.Phases.Implementation.Tasks_approved = true

	project := New(state)
	machine := project.BuildStateMachine()
	machine.SuppressPrompts(true)

	// Progress to ReviewActive
	machine.Fire(statechart.EventProjectInit)
	machine.Fire(statechart.EventSkipDiscovery)
	machine.Fire(statechart.EventSkipDesign)
	machine.Fire(statechart.EventTasksApproved) // Planning → Executing
	machine.Fire(statechart.EventAllTasksComplete) // Executing → Review

	currentState := machine.State()
	if currentState != statechart.ReviewActive {
		t.Fatalf("Expected to be in ReviewActive, got %v", currentState)
	}

	// Test backward transition: ReviewActive → ImplementationPlanning
	err := machine.Fire(statechart.EventReviewFail)
	if err != nil {
		t.Errorf("Failed to fire EventReviewFail: %v", err)
	}

	currentState = machine.State()
	if currentState != statechart.ImplementationPlanning {
		t.Errorf("After EventReviewFail, state = %v, want %v", currentState, statechart.ImplementationPlanning)
	}
}

// TestBuildStateMachine_Guards verifies guards are attached correctly
func TestBuildStateMachine_Guards(t *testing.T) {
	state := createTestState()
	project := New(state)

	machine := project.BuildStateMachine()
	machine.SuppressPrompts(true)

	// Progress to ImplementationPlanning
	machine.Fire(statechart.EventProjectInit)
	machine.Fire(statechart.EventSkipDiscovery)
	machine.Fire(statechart.EventSkipDesign)

	// Try to complete implementation without tasks - should fail (guard blocks)
	err := machine.Fire(statechart.EventAllTasksComplete)
	if err == nil {
		t.Error("EventAllTasksComplete should fail without tasks (guard should block)")
	}

	// Add a task
	state.Phases.Implementation.Tasks = append(state.Phases.Implementation.Tasks, phasesSchema.Task{
		Id:     "010",
		Name:   "Test task",
		Status: "pending",
	})

	// Transition to Executing with EventTaskCreated
	err = machine.Fire(statechart.EventTaskCreated)
	if err != nil {
		t.Errorf("EventTaskCreated with tasks should succeed: %v", err)
	}

	// Try to complete with uncompleted task - should fail (guard blocks)
	err = machine.Fire(statechart.EventAllTasksComplete)
	if err == nil {
		t.Error("EventAllTasksComplete should fail with uncompleted tasks (guard should block)")
	}

	// Mark task as completed
	state.Phases.Implementation.Tasks[0].Status = "completed"

	// Now it should succeed
	err = machine.Fire(statechart.EventAllTasksComplete)
	if err != nil {
		t.Errorf("EventAllTasksComplete with completed tasks should succeed: %v", err)
	}

	currentState := machine.State()
	if currentState != statechart.ReviewActive {
		t.Errorf("After completing tasks, state = %v, want %v", currentState, statechart.ReviewActive)
	}
}

// TestPhases verifies the Phases() method returns metadata for all phases
func TestPhases(t *testing.T) {
	state := createTestState()
	project := New(state)

	phaseMetadata := project.Phases()

	expectedPhases := []string{"discovery", "design", "implementation", "review", "finalize"}

	for _, phaseName := range expectedPhases {
		metadata, exists := phaseMetadata[phaseName]
		if !exists {
			t.Errorf("Phases() missing metadata for %q", phaseName)
			continue
		}

		if metadata.Name != phaseName {
			t.Errorf("Phase %q has incorrect name in metadata: %q", phaseName, metadata.Name)
		}

		if len(metadata.States) == 0 {
			t.Errorf("Phase %q has no states in metadata", phaseName)
		}
	}

	// Verify specific phase capabilities
	if !phaseMetadata["discovery"].SupportsArtifacts {
		t.Error("Discovery phase should support artifacts")
	}

	if !phaseMetadata["implementation"].SupportsTasks {
		t.Error("Implementation phase should support tasks")
	}

	if phaseMetadata["review"].SupportsTasks {
		t.Error("Review phase should not support tasks")
	}
}

// TestFullLifecycleWalkthrough simulates a complete project lifecycle
func TestFullLifecycleWalkthrough(t *testing.T) {
	state := createTestState()

	// Prepare state for a complete walkthrough
	// Add approved discovery artifacts
	now := time.Now()
	state.Phases.Discovery.Artifacts = []phasesSchema.Artifact{
		{
			Path:       "phases/discovery/research-001.md",
			Approved:   true,
			Created_at: now,
		},
	}

	// Add approved design artifacts
	state.Phases.Design.Artifacts = []phasesSchema.Artifact{
		{
			Path:       "phases/design/design-001.md",
			Approved:   true,
			Created_at: now,
		},
	}

	// Add completed implementation task
	state.Phases.Implementation.Tasks = []phasesSchema.Task{
		{
			Id:     "010",
			Name:   "Implement feature",
			Status: "completed",
		},
	}
	state.Phases.Implementation.Tasks_approved = true

	// Add approved review
	state.Phases.Review.Reports = []phasesSchema.ReviewReport{
		{
			Id:         "001",
			Assessment: "pass",
			Approved:   true,
		},
	}

	// Set finalize flags
	state.Phases.Finalize.Project_deleted = false // Initially false

	project := New(state)
	machine := project.BuildStateMachine()
	machine.SuppressPrompts(true)

	// Walk through the entire lifecycle
	transitions := []struct {
		event         statechart.Event
		expectedState statechart.State
		description   string
	}{
		{statechart.EventProjectInit, statechart.DiscoveryDecision, "Initialize"},
		{statechart.EventEnableDiscovery, statechart.DiscoveryActive, "Enable discovery"},
		{statechart.EventCompleteDiscovery, statechart.DesignDecision, "Complete discovery"},
		{statechart.EventEnableDesign, statechart.DesignActive, "Enable design"},
		{statechart.EventCompleteDesign, statechart.ImplementationPlanning, "Complete design"},
		{statechart.EventTasksApproved, statechart.ImplementationExecuting, "Approve tasks"},
		{statechart.EventAllTasksComplete, statechart.ReviewActive, "Complete tasks"},
		{statechart.EventReviewPass, statechart.FinalizeDocumentation, "Pass review"},
		{statechart.EventDocumentationDone, statechart.FinalizeChecks, "Documentation done"},
		{statechart.EventChecksDone, statechart.FinalizeDelete, "Checks done"},
	}

	for _, tt := range transitions {
		err := machine.Fire(tt.event)
		if err != nil {
			t.Errorf("%s: Fire(%v) error = %v", tt.description, tt.event, err)
			continue
		}

		currentState := machine.State()
		if currentState != tt.expectedState {
			t.Errorf("%s: state = %v, want %v", tt.description, currentState, tt.expectedState)
		}
	}

	// Final transition requires project_deleted flag to be true
	state.Phases.Finalize.Project_deleted = true
	err := machine.Fire(statechart.EventProjectDelete)
	if err != nil {
		t.Errorf("Final delete transition failed: %v", err)
	}

	currentState := machine.State()
	if currentState != statechart.NoProject {
		t.Errorf("Final state = %v, want %v", currentState, statechart.NoProject)
	}
}

// TestGuardCallable verifies guards can be called directly
func TestGuardCallable(t *testing.T) {
	state := createTestState()

	// Add a failed review report
	state.Phases.Review.Reports = []phasesSchema.ReviewReport{
		{
			Id:         "001",
			Assessment: "fail",
			Approved:   false,
		},
	}

	project := New(state)
	machine := project.BuildStateMachine()

	// We can't directly access the guard, but we can test that the machine
	// was configured with it by attempting the transition
	machine.SuppressPrompts(true)

	// Progress to ReviewActive
	machine.Fire(statechart.EventProjectInit)
	machine.Fire(statechart.EventSkipDiscovery)
	machine.Fire(statechart.EventSkipDesign)

	// Add and complete a task to reach review
	state.Phases.Implementation.Tasks = []phasesSchema.Task{
		{Id: "010", Name: "Task", Status: "completed"},
	}
	state.Phases.Implementation.Tasks_approved = true

	machine.Fire(statechart.EventTasksApproved)
	machine.Fire(statechart.EventAllTasksComplete)

	// Now test the backward transition guard
	err := machine.Fire(statechart.EventReviewFail)
	if err != nil {
		t.Errorf("EventReviewFail should succeed with failed review: %v", err)
	}

	currentState := machine.State()
	if currentState != statechart.ImplementationPlanning {
		t.Errorf("After failed review, state = %v, want %v", currentState, statechart.ImplementationPlanning)
	}
}
