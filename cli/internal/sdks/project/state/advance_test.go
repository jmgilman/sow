package state

import (
	"fmt"
	"testing"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// TestGetEventDeterminerReturnsConfiguredDeterminer verifies that GetEventDeterminer returns the configured determiner for a state.
func TestGetEventDeterminerReturnsConfiguredDeterminer(t *testing.T) {
	determiner := func(_ *Project) (Event, error) {
		return Event("test_event"), nil
	}

	config := &ProjectTypeConfig{
		onAdvance: map[State]EventDeterminer{
			State("test_state"): determiner,
		},
	}

	result := config.GetEventDeterminer(State("test_state"))

	if result == nil {
		t.Fatal("expected determiner to be returned")
	}

	// Verify the determiner works
	event, err := result(&Project{})
	if err != nil {
		t.Fatalf("determiner returned error: %v", err)
	}
	if event != Event("test_event") {
		t.Errorf("expected event='test_event', got %s", event)
	}
}

// TestGetEventDeterminerReturnsNilForUnconfiguredState verifies that GetEventDeterminer returns nil for states without determiners.
func TestGetEventDeterminerReturnsNilForUnconfiguredState(t *testing.T) {
	config := &ProjectTypeConfig{
		onAdvance: map[State]EventDeterminer{
			State("configured"): func(_ *Project) (Event, error) {
				return Event("event"), nil
			},
		},
	}

	result := config.GetEventDeterminer(State("unconfigured"))

	if result != nil {
		t.Error("expected nil for unconfigured state")
	}
}

// TestGetEventDeterminerMultipleStates verifies that different states can have different determiners.
func TestGetEventDeterminerMultipleStates(t *testing.T) {
	config := &ProjectTypeConfig{
		onAdvance: map[State]EventDeterminer{
			State("state1"): func(_ *Project) (Event, error) {
				return Event("event1"), nil
			},
			State("state2"): func(_ *Project) (Event, error) {
				return Event("event2"), nil
			},
		},
	}

	det1 := config.GetEventDeterminer(State("state1"))
	det2 := config.GetEventDeterminer(State("state2"))

	if det1 == nil || det2 == nil {
		t.Fatal("expected both determiners to be returned")
	}

	event1, _ := det1(&Project{})
	event2, _ := det2(&Project{})

	if event1 != Event("event1") {
		t.Errorf("expected state1 to return event1, got %s", event1)
	}
	if event2 != Event("event2") {
		t.Errorf("expected state2 to return event2, got %s", event2)
	}
}

// TestAdvanceCallsEventDeterminer verifies that Advance() calls the event determiner for current state.
func TestAdvanceCallsEventDeterminer(t *testing.T) {
	called := false
	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("end"),
				Event: Event("advance"),
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				called = true
				return Event("advance"), nil
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	_ = project.Advance()

	if !called {
		t.Error("expected event determiner to be called")
	}
}

// TestAdvanceFiresEvent verifies that Advance() fires the determined event.
func TestAdvanceFiresEvent(t *testing.T) {
	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("end"),
				Event: Event("advance"),
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err != nil {
		t.Fatalf("Advance() failed: %v", err)
	}

	if project.machine.State() != State("end") {
		t.Errorf("expected state=end, got %s", project.machine.State())
	}
}

// TestAdvanceTransitionsToNewState verifies that Advance() transitions to new state.
func TestAdvanceTransitionsToNewState(t *testing.T) {
	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("middle"),
				Event: Event("advance"),
			},
			{
				From:  State("middle"),
				To:    State("end"),
				Event: Event("finish"),
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err != nil {
		t.Fatalf("Advance() failed: %v", err)
	}

	if project.machine.State() != State("middle") {
		t.Errorf("expected state=middle, got %s", project.machine.State())
	}
}

// TestAdvanceErrorNoDeterminer verifies that Advance() returns error if no determiner configured.
func TestAdvanceErrorNoDeterminer(t *testing.T) {
	config := &ProjectTypeConfig{
		onAdvance: map[State]EventDeterminer{
			// No determiner for "start" state
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err == nil {
		t.Error("expected error when no determiner configured")
	}

	if err.Error() != "no event determiner for state: start" {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

// TestAdvanceDeterminerError verifies that Advance() returns error if determiner fails.
func TestAdvanceDeterminerError(t *testing.T) {
	config := &ProjectTypeConfig{
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return "", fmt.Errorf("cannot determine event")
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err == nil {
		t.Error("expected error when determiner fails")
	}

	// Check error is wrapped properly
	if err.Error() != "failed to determine event: cannot determine event" {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

// TestAdvanceGuardBlocks verifies that Advance() returns error if guard prevents transition.
func TestAdvanceGuardBlocks(t *testing.T) {
	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("end"),
				Event: Event("advance"),
				guardTemplate: func(_ *Project) bool {
					return false // Always block
				},
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err == nil {
		t.Error("expected error when guard blocks transition")
	}

	// State should not have changed
	if project.machine.State() != State("start") {
		t.Errorf("expected state to remain 'start', got %s", project.machine.State())
	}
}

// TestAdvanceGuardAllows verifies that Advance() succeeds when guard allows transition.
func TestAdvanceGuardAllows(t *testing.T) {
	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("end"),
				Event: Event("advance"),
				guardTemplate: func(_ *Project) bool {
					return true // Always allow
				},
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project := &Project{config: config}
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err != nil {
		t.Fatalf("Advance() should succeed when guard allows: %v", err)
	}

	if project.machine.State() != State("end") {
		t.Errorf("expected state=end, got %s", project.machine.State())
	}
}

// TestAdvanceFullFlow is an integration test of the complete advance flow.
func TestAdvanceFullFlow(t *testing.T) {
	// Setup: Create project with config
	project := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"test": {
					Outputs: []project.ArtifactState{
						{Type: "result", Approved: true},
					},
				},
			},
		},
	}

	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("middle"),
				Event: Event("advance"),
				guardTemplate: func(p *Project) bool {
					return p.PhaseOutputApproved("test", "result")
				},
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project.config = config
	project.machine = config.BuildMachine(project, State("start"))

	// Execute: Call Advance()
	err := project.Advance()

	// Verify: No error, state changed
	if err != nil {
		t.Fatalf("Advance() failed: %v", err)
	}
	if project.machine.State() != State("middle") {
		t.Errorf("expected state=middle, got %s", project.machine.State())
	}
}

// TestAdvanceFullFlowWithActions verifies that Advance() executes OnEntry and OnExit actions.
func TestAdvanceFullFlowWithActions(t *testing.T) {
	onExitCalled := false
	onEntryCalled := false

	project := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"test": {Status: "pending"},
			},
		},
	}

	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("start"),
				To:    State("end"),
				Event: Event("advance"),
				onExit: func(_ *Project) error {
					onExitCalled = true
					return nil
				},
				onEntry: func(p *Project) error {
					onEntryCalled = true
					phase := p.Phases["test"]
					phase.Status = "active"
					p.Phases["test"] = phase
					return nil
				},
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("start"): func(_ *Project) (Event, error) {
				return Event("advance"), nil
			},
		},
	}

	project.config = config
	project.machine = config.BuildMachine(project, State("start"))

	err := project.Advance()

	if err != nil {
		t.Fatalf("Advance() failed: %v", err)
	}

	if !onExitCalled {
		t.Error("expected onExit to be called")
	}
	if !onEntryCalled {
		t.Error("expected onEntry to be called")
	}
	if project.Phases["test"].Status != "active" {
		t.Errorf("expected onEntry to set status=active, got %s", project.Phases["test"].Status)
	}
}

// TestAdvanceDeterminerAccessesProjectState verifies that determiner can access and use project state.
func TestAdvanceDeterminerAccessesProjectState(t *testing.T) {
	project := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"review": {
					Metadata: map[string]interface{}{
						"assessment": "pass",
					},
				},
			},
		},
	}

	config := &ProjectTypeConfig{
		transitions: []TransitionConfig{
			{
				From:  State("review"),
				To:    State("passed"),
				Event: Event("pass"),
			},
			{
				From:  State("review"),
				To:    State("failed"),
				Event: Event("fail"),
			},
		},
		onAdvance: map[State]EventDeterminer{
			State("review"): func(p *Project) (Event, error) {
				phase := p.Phases["review"]
				assessment := phase.Metadata["assessment"]
				if assessment == "pass" {
					return Event("pass"), nil
				}
				return Event("fail"), nil
			},
		},
	}

	project.config = config
	project.machine = config.BuildMachine(project, State("review"))

	err := project.Advance()

	if err != nil {
		t.Fatalf("Advance() failed: %v", err)
	}

	if project.machine.State() != State("passed") {
		t.Errorf("expected state=passed based on assessment, got %s", project.machine.State())
	}
}
