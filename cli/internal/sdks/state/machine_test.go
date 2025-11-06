package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMachine_InitialState verifies that a machine starts in the correct initial state.
func TestMachine_InitialState(t *testing.T) {
	initialState := State("InitialState")
	builder := NewBuilder(initialState, nil)
	machine := builder.Build()

	assert.Equal(t, initialState, machine.State())
}

// TestMachine_State_UnchangedByCanFire verifies that CanFire doesn't mutate state.
func TestMachine_State_UnchangedByCanFire(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo)
	machine := builder.Build()

	// Check that CanFire doesn't change state
	can, err := machine.CanFire(eventGo)
	require.NoError(t, err)
	assert.True(t, can)
	assert.Equal(t, stateA, machine.State(), "CanFire should not change state")
}

// TestMachine_Fire_UpdatesState verifies that Fire transitions to the correct state.
func TestMachine_Fire_UpdatesState(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventAdvance := Event("advance")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventAdvance)
	machine := builder.Build()

	// Fire event
	err := machine.Fire(eventAdvance)
	require.NoError(t, err)

	// Verify state changed
	assert.Equal(t, stateB, machine.State())
}

// TestMachine_Fire_InvalidEvent verifies that firing an invalid event returns an error.
func TestMachine_Fire_InvalidEvent(t *testing.T) {
	stateA := State("A")
	invalidEvent := Event("invalid")

	builder := NewBuilder(stateA, nil)
	machine := builder.Build()

	// Try to fire invalid event
	err := machine.Fire(invalidEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid", "Error should mention the event")

	// State should not change
	assert.Equal(t, stateA, machine.State())
}

// TestMachine_Fire_WithFailingGuard verifies that failing guards block transitions.
func TestMachine_Fire_WithFailingGuard(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	// Guard that always fails
	guardPassed := false
	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo, WithGuard(func() bool {
		return guardPassed
	}))
	machine := builder.Build()

	// Fire should fail due to guard
	err := machine.Fire(eventGo)
	assert.Error(t, err)

	// State should not change
	assert.Equal(t, stateA, machine.State())
}

// TestMachine_Fire_WithPassingGuard verifies that passing guards allow transitions.
func TestMachine_Fire_WithPassingGuard(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	// Guard that always passes
	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo, WithGuard(func() bool {
		return true
	}))
	machine := builder.Build()

	// Fire should succeed
	err := machine.Fire(eventGo)
	require.NoError(t, err)

	// State should change
	assert.Equal(t, stateB, machine.State())
}

// TestMachine_MultipleSequentialTransitions verifies that a machine can handle
// multiple transitions in sequence.
func TestMachine_MultipleSequentialTransitions(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	event1 := Event("to_b")
	event2 := Event("to_c")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, event1)
	builder.AddTransition(stateB, stateC, event2)
	machine := builder.Build()

	// Transition A -> B
	err := machine.Fire(event1)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())

	// Transition B -> C
	err = machine.Fire(event2)
	require.NoError(t, err)
	assert.Equal(t, stateC, machine.State())
}

// TestMachine_CanFire_WithNilGuard verifies that nil guards always allow transitions.
func TestMachine_CanFire_WithNilGuard(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo) // No guard = nil guard
	machine := builder.Build()

	can, err := machine.CanFire(eventGo)
	require.NoError(t, err)
	assert.True(t, can, "Nil guard should always allow transition")
}

// TestMachine_CanFire_WithPassingGuard verifies that CanFire returns true for passing guards.
func TestMachine_CanFire_WithPassingGuard(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo, WithGuard(func() bool {
		return true
	}))
	machine := builder.Build()

	can, err := machine.CanFire(eventGo)
	require.NoError(t, err)
	assert.True(t, can)
}

// TestMachine_CanFire_WithFailingGuard verifies that CanFire returns false for failing guards.
func TestMachine_CanFire_WithFailingGuard(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo, WithGuard(func() bool {
		return false
	}))
	machine := builder.Build()

	can, err := machine.CanFire(eventGo)
	require.NoError(t, err)
	assert.False(t, can)
}

// TestMachine_CanFire_InvalidEvent verifies that CanFire returns false for invalid events.
func TestMachine_CanFire_InvalidEvent(t *testing.T) {
	stateA := State("A")
	invalidEvent := Event("invalid")

	builder := NewBuilder(stateA, nil)
	machine := builder.Build()

	can, err := machine.CanFire(invalidEvent)
	// The stateless library returns false and no error for invalid events
	assert.NoError(t, err)
	assert.False(t, can)
}

// TestMachine_Cycles verifies that machines can handle cyclic transitions (A -> B -> A).
func TestMachine_Cycles(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventToB := Event("to_b")
	eventToA := Event("to_a")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventToB)
	builder.AddTransition(stateB, stateA, eventToA)
	machine := builder.Build()

	// Cycle A -> B -> A
	err := machine.Fire(eventToB)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())

	err = machine.Fire(eventToA)
	require.NoError(t, err)
	assert.Equal(t, stateA, machine.State())

	// Cycle again
	err = machine.Fire(eventToB)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())
}

// TestMachine_Branching verifies that machines handle branching based on guards.
func TestMachine_Branching(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	eventBranch := Event("branch")

	// Shared guard condition
	goToB := true

	builder := NewBuilder(stateA, nil)
	// A -> B if goToB is true
	builder.AddTransition(stateA, stateB, eventBranch, WithGuard(func() bool {
		return goToB
	}))
	// A -> C if goToB is false
	builder.AddTransition(stateA, stateC, eventBranch, WithGuard(func() bool {
		return !goToB
	}))
	machine := builder.Build()

	// First branch: A -> B
	err := machine.Fire(eventBranch)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())

	// Reset to A for second branch
	builder2 := NewBuilder(stateA, nil)
	builder2.AddTransition(stateA, stateB, eventBranch, WithGuard(func() bool {
		return goToB
	}))
	builder2.AddTransition(stateA, stateC, eventBranch, WithGuard(func() bool {
		return !goToB
	}))
	machine2 := builder2.Build()

	goToB = false
	err = machine2.Fire(eventBranch)
	require.NoError(t, err)
	assert.Equal(t, stateC, machine2.State())
}

// TestMachine_DiamondPattern verifies that machines handle diamond patterns (A -> B -> D, A -> C -> D).
func TestMachine_DiamondPattern(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	stateD := State("D")
	eventToB := Event("to_b")
	eventToC := Event("to_c")
	eventToD := Event("to_d")

	// Path 1: A -> B -> D
	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventToB)
	builder.AddTransition(stateA, stateC, eventToC)
	builder.AddTransition(stateB, stateD, eventToD)
	builder.AddTransition(stateC, stateD, eventToD)
	machine := builder.Build()

	err := machine.Fire(eventToB)
	require.NoError(t, err)
	assert.Equal(t, stateB, machine.State())

	err = machine.Fire(eventToD)
	require.NoError(t, err)
	assert.Equal(t, stateD, machine.State())

	// Path 2: A -> C -> D (need a new machine starting from A)
	builder2 := NewBuilder(stateA, nil)
	builder2.AddTransition(stateA, stateB, eventToB)
	builder2.AddTransition(stateA, stateC, eventToC)
	builder2.AddTransition(stateB, stateD, eventToD)
	builder2.AddTransition(stateC, stateD, eventToD)
	machine2 := builder2.Build()

	err = machine2.Fire(eventToC)
	require.NoError(t, err)
	assert.Equal(t, stateC, machine2.State())

	err = machine2.Fire(eventToD)
	require.NoError(t, err)
	assert.Equal(t, stateD, machine2.State())
}

// TestMachine_PermittedTriggers verifies that PermittedTriggers returns available events.
func TestMachine_PermittedTriggers(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	eventToB := Event("to_b")
	eventToC := Event("to_c")

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventToB)
	builder.AddTransition(stateA, stateC, eventToC)
	machine := builder.Build()

	triggers, err := machine.PermittedTriggers()
	require.NoError(t, err)
	assert.Len(t, triggers, 2)
	assert.Contains(t, triggers, eventToB)
	assert.Contains(t, triggers, eventToC)
}

// buildComplexWorkflowMachine creates a state machine with 12 states and 20+ transitions
// for testing complex workflows. It returns the machine and all guard variables.
func buildComplexWorkflowMachine(
	guards *complexWorkflowGuards,
) *Machine {
	// Define states (12 states)
	stateNoProject := State("NoProject")
	statePlanningActive := State("PlanningActive")
	statePlanningReview := State("PlanningReview")
	stateDesignActive := State("DesignActive")
	stateDesignReview := State("DesignReview")
	stateImplementationPlanning := State("ImplementationPlanning")
	stateImplementationActive := State("ImplementationActive")
	stateImplementationReview := State("ImplementationReview")
	stateTestingActive := State("TestingActive")
	stateTestingReview := State("TestingReview")
	stateDeploymentActive := State("DeploymentActive")
	stateDone := State("Done")

	// Define events
	eventInit := Event("init")
	eventCompletePlanning := Event("complete_planning")
	eventPlanningApproved := Event("planning_approved")
	eventPlanningRejected := Event("planning_rejected")
	eventCompleteDesign := Event("complete_design")
	eventDesignApproved := Event("design_approved")
	eventDesignRejected := Event("design_rejected")
	eventStartImplementation := Event("start_implementation")
	eventCompleteImplementation := Event("complete_implementation")
	eventImplementationApproved := Event("implementation_approved")
	eventImplementationRejected := Event("implementation_rejected")
	eventCompleteTesting := Event("complete_testing")
	eventTestingApproved := Event("testing_approved")
	eventTestingRejected := Event("testing_rejected")
	eventDeploymentComplete := Event("deployment_complete")
	eventSkipDesign := Event("skip_design")
	eventSkipTesting := Event("skip_testing")
	eventAbort := Event("abort")

	builder := NewBuilder(stateNoProject, nil)

	// Build complex state machine (20+ transitions)
	builder.AddTransition(stateNoProject, statePlanningActive, eventInit)
	builder.AddTransition(statePlanningActive, statePlanningReview, eventCompletePlanning, WithGuard(func() bool {
		return guards.planningComplete
	}))
	builder.AddTransition(statePlanningReview, stateDesignActive, eventPlanningApproved, WithGuard(func() bool {
		return guards.reviewApproved && !guards.skipDesignPhase
	}))
	builder.AddTransition(statePlanningReview, statePlanningActive, eventPlanningRejected, WithGuard(func() bool {
		return !guards.reviewApproved
	}))
	builder.AddTransition(statePlanningReview, stateImplementationPlanning, eventSkipDesign, WithGuard(func() bool {
		return guards.reviewApproved && guards.skipDesignPhase
	}))

	builder.AddTransition(stateDesignActive, stateDesignReview, eventCompleteDesign, WithGuard(func() bool {
		return guards.designComplete
	}))
	builder.AddTransition(stateDesignReview, stateImplementationPlanning, eventDesignApproved, WithGuard(func() bool {
		return guards.reviewApproved
	}))
	builder.AddTransition(stateDesignReview, stateDesignActive, eventDesignRejected, WithGuard(func() bool {
		return !guards.reviewApproved
	}))

	builder.AddTransition(stateImplementationPlanning, stateImplementationActive, eventStartImplementation)
	builder.AddTransition(stateImplementationActive, stateImplementationReview, eventCompleteImplementation, WithGuard(func() bool {
		return guards.implementationComplete
	}))
	builder.AddTransition(stateImplementationReview, stateTestingActive, eventImplementationApproved, WithGuard(func() bool {
		return guards.reviewApproved && !guards.skipTestingPhase
	}))
	builder.AddTransition(stateImplementationReview, stateImplementationActive, eventImplementationRejected, WithGuard(func() bool {
		return !guards.reviewApproved
	}))
	builder.AddTransition(stateImplementationReview, stateDeploymentActive, eventSkipTesting, WithGuard(func() bool {
		return guards.reviewApproved && guards.skipTestingPhase
	}))

	builder.AddTransition(stateTestingActive, stateTestingReview, eventCompleteTesting, WithGuard(func() bool {
		return guards.testingComplete
	}))
	builder.AddTransition(stateTestingReview, stateDeploymentActive, eventTestingApproved, WithGuard(func() bool {
		return guards.reviewApproved
	}))
	builder.AddTransition(stateTestingReview, stateTestingActive, eventTestingRejected, WithGuard(func() bool {
		return !guards.reviewApproved
	}))

	builder.AddTransition(stateDeploymentActive, stateDone, eventDeploymentComplete)

	// Abort transitions from any active state
	builder.AddTransition(statePlanningActive, stateNoProject, eventAbort)
	builder.AddTransition(stateDesignActive, stateNoProject, eventAbort)
	builder.AddTransition(stateImplementationActive, stateNoProject, eventAbort)
	builder.AddTransition(stateTestingActive, stateNoProject, eventAbort)

	return builder.Build()
}

// complexWorkflowGuards holds guard condition variables for complex workflow tests.
type complexWorkflowGuards struct {
	planningComplete       bool
	designComplete         bool
	implementationComplete bool
	testingComplete        bool
	reviewApproved         bool
	skipDesignPhase        bool
	skipTestingPhase       bool
}

// TestMachine_ComplexWorkflow is a comprehensive test with 10+ states and 20+ transitions,
// simulating a real-world project lifecycle.
//
//nolint:funlen // Test intentionally long to demonstrate complete workflow sequence
func TestMachine_ComplexWorkflow(t *testing.T) {
	guards := &complexWorkflowGuards{}
	machine := buildComplexWorkflowMachine(guards)

	// Define state and event constants for test assertions
	stateNoProject := State("NoProject")
	statePlanningActive := State("PlanningActive")
	statePlanningReview := State("PlanningReview")
	stateDesignActive := State("DesignActive")
	stateDesignReview := State("DesignReview")
	stateImplementationPlanning := State("ImplementationPlanning")
	stateImplementationActive := State("ImplementationActive")
	stateImplementationReview := State("ImplementationReview")
	stateTestingActive := State("TestingActive")
	stateTestingReview := State("TestingReview")
	stateDeploymentActive := State("DeploymentActive")
	stateDone := State("Done")

	eventInit := Event("init")
	eventCompletePlanning := Event("complete_planning")
	eventPlanningApproved := Event("planning_approved")
	eventPlanningRejected := Event("planning_rejected")
	eventCompleteDesign := Event("complete_design")
	eventDesignApproved := Event("design_approved")
	eventStartImplementation := Event("start_implementation")
	eventCompleteImplementation := Event("complete_implementation")
	eventImplementationApproved := Event("implementation_approved")
	eventCompleteTesting := Event("complete_testing")
	eventTestingApproved := Event("testing_approved")
	eventDeploymentComplete := Event("deployment_complete")

	// Test workflow path: NoProject -> Planning -> Design -> Implementation -> Testing -> Deployment -> Done

	// 1. Initialize project
	assert.Equal(t, stateNoProject, machine.State())
	err := machine.Fire(eventInit)
	require.NoError(t, err)
	assert.Equal(t, statePlanningActive, machine.State())

	// 2. Try to complete planning without finishing (should fail)
	err = machine.Fire(eventCompletePlanning)
	assert.Error(t, err, "Should fail when planning not complete")
	assert.Equal(t, statePlanningActive, machine.State())

	// 3. Complete planning and move to review
	guards.planningComplete = true
	err = machine.Fire(eventCompletePlanning)
	require.NoError(t, err)
	assert.Equal(t, statePlanningReview, machine.State())

	// 4. Reject planning (go back to planning active)
	guards.reviewApproved = false
	err = machine.Fire(eventPlanningRejected)
	require.NoError(t, err)
	assert.Equal(t, statePlanningActive, machine.State())

	// 5. Complete planning again and approve
	err = machine.Fire(eventCompletePlanning)
	require.NoError(t, err)
	guards.reviewApproved = true
	err = machine.Fire(eventPlanningApproved)
	require.NoError(t, err)
	assert.Equal(t, stateDesignActive, machine.State())

	// 6. Complete design
	guards.designComplete = true
	err = machine.Fire(eventCompleteDesign)
	require.NoError(t, err)
	assert.Equal(t, stateDesignReview, machine.State())

	// 7. Approve design
	guards.reviewApproved = true
	err = machine.Fire(eventDesignApproved)
	require.NoError(t, err)
	assert.Equal(t, stateImplementationPlanning, machine.State())

	// 8. Start implementation
	err = machine.Fire(eventStartImplementation)
	require.NoError(t, err)
	assert.Equal(t, stateImplementationActive, machine.State())

	// 9. Complete implementation
	guards.implementationComplete = true
	err = machine.Fire(eventCompleteImplementation)
	require.NoError(t, err)
	assert.Equal(t, stateImplementationReview, machine.State())

	// 10. Approve implementation
	guards.reviewApproved = true
	err = machine.Fire(eventImplementationApproved)
	require.NoError(t, err)
	assert.Equal(t, stateTestingActive, machine.State())

	// 11. Complete testing
	guards.testingComplete = true
	err = machine.Fire(eventCompleteTesting)
	require.NoError(t, err)
	assert.Equal(t, stateTestingReview, machine.State())

	// 12. Approve testing
	guards.reviewApproved = true
	err = machine.Fire(eventTestingApproved)
	require.NoError(t, err)
	assert.Equal(t, stateDeploymentActive, machine.State())

	// 13. Complete deployment
	err = machine.Fire(eventDeploymentComplete)
	require.NoError(t, err)
	assert.Equal(t, stateDone, machine.State())

	// Verify we can't fire any more events from Done state
	can, _ := machine.CanFire(eventInit)
	assert.False(t, can, "Should not be able to fire events from Done state")
}

// TestMachine_ComplexWorkflow_AlternatePath tests the complex workflow with design and testing skipped.
func TestMachine_ComplexWorkflow_AlternatePath(t *testing.T) {
	// Define states
	stateNoProject := State("NoProject")
	statePlanningActive := State("PlanningActive")
	statePlanningReview := State("PlanningReview")
	stateImplementationPlanning := State("ImplementationPlanning")
	stateImplementationActive := State("ImplementationActive")
	stateImplementationReview := State("ImplementationReview")
	stateDeploymentActive := State("DeploymentActive")
	stateDone := State("Done")

	// Define events
	eventInit := Event("init")
	eventCompletePlanning := Event("complete_planning")
	eventSkipDesign := Event("skip_design")
	eventStartImplementation := Event("start_implementation")
	eventCompleteImplementation := Event("complete_implementation")
	eventSkipTesting := Event("skip_testing")
	eventDeploymentComplete := Event("deployment_complete")

	// Guards
	planningComplete := false
	implementationComplete := false
	reviewApproved := false
	skipDesignPhase := false
	skipTestingPhase := false

	builder := NewBuilder(stateNoProject, nil)
	builder.AddTransition(stateNoProject, statePlanningActive, eventInit)
	builder.AddTransition(statePlanningActive, statePlanningReview, eventCompletePlanning, WithGuard(func() bool {
		return planningComplete
	}))
	builder.AddTransition(statePlanningReview, stateImplementationPlanning, eventSkipDesign, WithGuard(func() bool {
		return reviewApproved && skipDesignPhase
	}))
	builder.AddTransition(stateImplementationPlanning, stateImplementationActive, eventStartImplementation)
	builder.AddTransition(stateImplementationActive, stateImplementationReview, eventCompleteImplementation, WithGuard(func() bool {
		return implementationComplete
	}))
	builder.AddTransition(stateImplementationReview, stateDeploymentActive, eventSkipTesting, WithGuard(func() bool {
		return reviewApproved && skipTestingPhase
	}))
	builder.AddTransition(stateDeploymentActive, stateDone, eventDeploymentComplete)

	machine := builder.Build()

	// Test path: NoProject -> Planning -> Implementation (skip design) -> Deployment (skip testing) -> Done

	err := machine.Fire(eventInit)
	require.NoError(t, err)

	planningComplete = true
	err = machine.Fire(eventCompletePlanning)
	require.NoError(t, err)

	reviewApproved = true
	skipDesignPhase = true
	err = machine.Fire(eventSkipDesign)
	require.NoError(t, err)
	assert.Equal(t, stateImplementationPlanning, machine.State())

	err = machine.Fire(eventStartImplementation)
	require.NoError(t, err)

	implementationComplete = true
	err = machine.Fire(eventCompleteImplementation)
	require.NoError(t, err)

	reviewApproved = true
	skipTestingPhase = true
	err = machine.Fire(eventSkipTesting)
	require.NoError(t, err)
	assert.Equal(t, stateDeploymentActive, machine.State())

	err = machine.Fire(eventDeploymentComplete)
	require.NoError(t, err)
	assert.Equal(t, stateDone, machine.State())
}
