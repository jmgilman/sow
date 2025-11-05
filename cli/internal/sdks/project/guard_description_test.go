package project

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestGuardDescriptionInError verifies that custom guard descriptions
// appear in error messages when guards fail.
func TestGuardDescriptionInError(t *testing.T) {
	config := NewProjectTypeConfigBuilder("test").
		SetInitialState(sdkstate.State("start")).
		AddTransition(
			sdkstate.State("start"),
			sdkstate.State("end"),
			sdkstate.Event("go"),
			WithGuard("latest review approved", func(_ *state.Project) bool {
				return false // Always fail
			}),
		).
		Build()

	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Name: "test",
			Statechart: projschema.StatechartState{
				Current_state: "start",
			},
			Phases: make(map[string]projschema.PhaseState),
		},
	}

	machine := config.BuildMachine(proj, sdkstate.State("start"))

	// Try to fire the event - should fail with our custom description
	err := config.FireWithPhaseUpdates(machine, sdkstate.Event("go"), proj)
	
	if err == nil {
		t.Fatal("expected error when guard fails, got nil")
	}

	// Print the error message to see what it looks like
	fmt.Printf("\nError message: %v\n\n", err)

	// Check that the error message contains our custom description
	errMsg := err.Error()
	if !strings.Contains(errMsg, "latest review approved") {
		t.Errorf("expected error message to contain 'latest review approved', got: %v", errMsg)
	}

	// Should not contain generic func names like "func1"
	if strings.Contains(errMsg, "func1") {
		t.Errorf("error message should not contain generic func names like 'func1', got: %v", errMsg)
	}
}
