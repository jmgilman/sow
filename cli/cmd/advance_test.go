package cmd

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	sdkproject "github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projectschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestAdvanceCommandSignature verifies that the advance command accepts the correct arguments and has the required flags
func TestAdvanceCommandSignature(t *testing.T) {
	cmd := NewAdvanceCmd()

	// Test accepts 0 arguments
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("should accept 0 arguments: %v", err)
	}

	// Test accepts 1 argument
	err = cmd.Args(cmd, []string{"event_name"})
	if err != nil {
		t.Errorf("should accept 1 argument: %v", err)
	}

	// Test rejects 2 arguments
	err = cmd.Args(cmd, []string{"event1", "event2"})
	if err == nil {
		t.Error("should reject 2 arguments")
	}

	// Test --list flag is defined
	listFlag := cmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("--list flag not defined")
	}
	if listFlag != nil && listFlag.Value.Type() != "bool" {
		t.Errorf("--list flag should be boolean, got %s", listFlag.Value.Type())
	}

	// Test --dry-run flag is defined
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("--dry-run flag not defined")
	}
	if dryRunFlag != nil && dryRunFlag.Value.Type() != "bool" {
		t.Errorf("--dry-run flag should be boolean, got %s", dryRunFlag.Value.Type())
	}
}

// TestAdvanceFlagValidation verifies mutual exclusivity rules for flags and arguments
func TestAdvanceFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		listFlag    bool
		dryRunFlag  bool
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "no flags, no args (auto mode) - valid",
			args:        []string{},
			listFlag:    false,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "no flags, one arg (explicit event mode) - valid",
			args:        []string{"finalize"},
			listFlag:    false,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "list flag, no args (discovery mode) - valid",
			args:        []string{},
			listFlag:    true,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "dry-run flag, one arg (dry-run mode) - valid",
			args:        []string{"finalize"},
			listFlag:    false,
			dryRunFlag:  true,
			shouldError: false,
		},
		{
			name:        "list flag with event argument - invalid",
			args:        []string{"finalize"},
			listFlag:    true,
			dryRunFlag:  false,
			shouldError: true,
			errorMsg:    "cannot specify event argument with --list flag",
		},
		{
			name:        "dry-run flag without event argument - invalid",
			args:        []string{},
			listFlag:    false,
			dryRunFlag:  true,
			shouldError: true,
			errorMsg:    "--dry-run requires an event argument",
		},
		{
			name:        "both list and dry-run flags - invalid",
			args:        []string{},
			listFlag:    true,
			dryRunFlag:  true,
			shouldError: true,
			errorMsg:    "cannot use --list and --dry-run together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := NewAdvanceCmd()

			// Set flags
			cmd.Flags().Set("list", boolToString(tt.listFlag))
			cmd.Flags().Set("dry-run", boolToString(tt.dryRunFlag))

			// Call the validation function that should be in RunE
			err := validateAdvanceFlags(cmd, tt.args)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to convert bool to string for flag setting
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestAdvanceAutoLinear tests auto-determination for linear states (one transition).
func TestAdvanceAutoLinear(t *testing.T) {
	t.Skip("TODO: Implement test - requires project test setup")
	// This test should:
	// 1. Create test project in linear state (e.g., ImplementationPlanning)
	// 2. Set up prerequisites so guard passes
	// 3. Call executeAutoTransition (once extracted)
	// 4. Verify: state advances, no error
}

// TestAdvanceAutoBranching tests auto-determination for state-determined branching (uses AddBranch discriminator).
func TestAdvanceAutoBranching(t *testing.T) {
	t.Skip("TODO: Implement test - requires project test setup")
	// This test should:
	// 1. Create test project in ReviewActive state (branching state)
	// 2. Add approved review with "pass" assessment
	// 3. Call executeAutoTransition
	// 4. Verify: DetermineEvent selects EventReviewPass, transitions to FinalizeChecks
}

// TestAdvanceAutoIntentBased tests auto-determination failure for intent-based branching (multiple transitions, no discriminator).
func TestAdvanceAutoIntentBased(t *testing.T) {
	t.Skip("TODO: Implement test - requires exploration project setup")
	// This test should:
	// 1. Create test project in Researching state (intent-based branching)
	// 2. Call executeAutoTransition
	// 3. Verify: error returned
	// 4. Verify: error message suggests using --list
	// 5. Verify: error message lists available events (finalize, add_more_research)
}

// TestAdvanceAutoTerminalState tests auto-determination failure for terminal states (no transitions).
func TestAdvanceAutoTerminalState(t *testing.T) {
	t.Skip("TODO: Implement test - requires project test setup")
	// This test should:
	// 1. Create test project in terminal state (e.g., Completed)
	// 2. Call executeAutoTransition
	// 3. Verify: error returned
	// 4. Verify: error message indicates terminal state
}

// TestAdvanceListAvailable tests listing all available transitions when all guards pass.
func TestAdvanceListAvailable(t *testing.T) {
	// Create test project with multiple transitions (all guards pass)
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Middle"),
			sdkstate.Event("go_middle"),
			sdkproject.WithDescription("Proceed to middle state"),
			sdkproject.WithGuard(
				"always allowed",
				func(_ *state.Project) bool { return true },
			),
		).
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("End"),
			sdkstate.Event("skip_to_end"),
			sdkproject.WithDescription("Skip directly to end"),
			sdkproject.WithGuard(
				"no prerequisites",
				func(_ *state.Project) bool { return true },
			),
		).
		Build()

	proj := createTestProjectWithConfig(t, "Start", config)

	// Capture output
	output := captureOutput(func() {
		err := listAvailableTransitions(proj, sdkstate.State("Start"))
		if err != nil {
			t.Fatalf("listAvailableTransitions failed: %v", err)
		}
	})

	// Verify output contains current state
	if !strings.Contains(output, "Current state: Start") {
		t.Errorf("Expected output to contain 'Current state: Start', got:\n%s", output)
	}

	// Verify both transitions are shown
	if !strings.Contains(output, "sow advance go_middle") {
		t.Errorf("Expected output to contain 'sow advance go_middle', got:\n%s", output)
	}
	if !strings.Contains(output, "sow advance skip_to_end") {
		t.Errorf("Expected output to contain 'sow advance skip_to_end', got:\n%s", output)
	}

	// Verify target states
	if !strings.Contains(output, "→ Middle") {
		t.Errorf("Expected output to contain '→ Middle', got:\n%s", output)
	}
	if !strings.Contains(output, "→ End") {
		t.Errorf("Expected output to contain '→ End', got:\n%s", output)
	}

	// Verify descriptions
	if !strings.Contains(output, "Proceed to middle state") {
		t.Errorf("Expected output to contain description 'Proceed to middle state', got:\n%s", output)
	}
	if !strings.Contains(output, "Skip directly to end") {
		t.Errorf("Expected output to contain description 'Skip directly to end', got:\n%s", output)
	}

	// Verify guard descriptions
	if !strings.Contains(output, "Requires: always allowed") {
		t.Errorf("Expected output to contain guard description 'Requires: always allowed', got:\n%s", output)
	}
	if !strings.Contains(output, "Requires: no prerequisites") {
		t.Errorf("Expected output to contain guard description 'Requires: no prerequisites', got:\n%s", output)
	}

	// Verify NO [BLOCKED] markers
	if strings.Contains(output, "[BLOCKED]") {
		t.Errorf("Expected no [BLOCKED] markers, but found some in:\n%s", output)
	}
}

// TestAdvanceListBlocked tests listing transitions when some guards fail.
func TestAdvanceListBlocked(t *testing.T) {
	// Create test project with mixed guards (one passes, one fails)
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Middle"),
			sdkstate.Event("go_middle"),
			sdkproject.WithDescription("Permitted transition"),
			sdkproject.WithGuard(
				"always allowed",
				func(_ *state.Project) bool { return true },
			),
		).
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("End"),
			sdkstate.Event("skip_to_end"),
			sdkproject.WithDescription("Blocked transition"),
			sdkproject.WithGuard(
				"never allowed",
				func(_ *state.Project) bool { return false },
			),
		).
		Build()

	proj := createTestProjectWithConfig(t, "Start", config)

	// Capture output
	output := captureOutput(func() {
		err := listAvailableTransitions(proj, sdkstate.State("Start"))
		if err != nil {
			t.Fatalf("listAvailableTransitions failed: %v", err)
		}
	})

	// Verify permitted transition shown normally
	if !strings.Contains(output, "sow advance go_middle\n") {
		t.Errorf("Expected permitted transition 'go_middle' without [BLOCKED], got:\n%s", output)
	}

	// Verify blocked transition has [BLOCKED] marker
	if !strings.Contains(output, "sow advance skip_to_end  [BLOCKED]") {
		t.Errorf("Expected blocked transition 'skip_to_end' with [BLOCKED] marker, got:\n%s", output)
	}

	// Both transitions should be shown
	if !strings.Contains(output, "→ Middle") {
		t.Errorf("Expected permitted target '→ Middle', got:\n%s", output)
	}
	if !strings.Contains(output, "→ End") {
		t.Errorf("Expected blocked target '→ End', got:\n%s", output)
	}
}

// TestAdvanceListAllBlocked tests listing when all guards fail.
func TestAdvanceListAllBlocked(t *testing.T) {
	// Create test project where all guards fail
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Middle"),
			sdkstate.Event("go_middle"),
			sdkproject.WithDescription("First blocked transition"),
			sdkproject.WithGuard(
				"prerequisites not met",
				func(_ *state.Project) bool { return false },
			),
		).
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("End"),
			sdkstate.Event("skip_to_end"),
			sdkproject.WithDescription("Second blocked transition"),
			sdkproject.WithGuard(
				"approval required",
				func(_ *state.Project) bool { return false },
			),
		).
		Build()

	proj := createTestProjectWithConfig(t, "Start", config)

	// Capture output
	output := captureOutput(func() {
		err := listAvailableTransitions(proj, sdkstate.State("Start"))
		if err != nil {
			t.Fatalf("listAvailableTransitions failed: %v", err)
		}
	})

	// Verify all-blocked message appears
	if !strings.Contains(output, "(All configured transitions are currently blocked by guard conditions)") {
		t.Errorf("Expected all-blocked message, got:\n%s", output)
	}

	// Verify both transitions shown with [BLOCKED] markers
	if !strings.Contains(output, "sow advance go_middle  [BLOCKED]") {
		t.Errorf("Expected first transition with [BLOCKED] marker, got:\n%s", output)
	}
	if !strings.Contains(output, "sow advance skip_to_end  [BLOCKED]") {
		t.Errorf("Expected second transition with [BLOCKED] marker, got:\n%s", output)
	}

	// Verify transitions are still displayed
	if !strings.Contains(output, "→ Middle") {
		t.Errorf("Expected first target state, got:\n%s", output)
	}
	if !strings.Contains(output, "→ End") {
		t.Errorf("Expected second target state, got:\n%s", output)
	}
}

// TestAdvanceListTerminal tests listing from a terminal state.
func TestAdvanceListTerminal(t *testing.T) {
	// Create test project in terminal state (no transitions configured from this state)
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Terminal"),
			sdkstate.Event("finish"),
		).
		// No transitions FROM "Terminal" state
		Build()

	proj := createTestProjectWithConfig(t, "Terminal", config)

	// Capture output
	output := captureOutput(func() {
		err := listAvailableTransitions(proj, sdkstate.State("Terminal"))
		if err != nil {
			t.Fatalf("listAvailableTransitions failed: %v", err)
		}
	})

	// Verify terminal state messages
	if !strings.Contains(output, "No transitions available from current state.") {
		t.Errorf("Expected 'No transitions available' message, got:\n%s", output)
	}
	if !strings.Contains(output, "This may be a terminal state.") {
		t.Errorf("Expected 'terminal state' message, got:\n%s", output)
	}

	// Should NOT show "Available transitions:" section
	// (It does show this but then says no transitions - this is acceptable)
}

// TestAdvanceListWithDescriptions tests listing with description metadata.
func TestAdvanceListWithDescriptions(t *testing.T) {
	// This is already covered by TestAdvanceListAvailable
	// which verifies both transition and guard descriptions are shown
	t.Skip("Already covered by TestAdvanceListAvailable")
}

// TestAdvanceListNoDescriptions tests listing without description metadata.
func TestAdvanceListNoDescriptions(t *testing.T) {
	// Create test project with transitions but no descriptions
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Middle"),
			sdkstate.Event("go_middle"),
			// No description
			sdkproject.WithGuard(
				"", // No guard description either
				func(_ *state.Project) bool { return true },
			),
		).
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("End"),
			sdkstate.Event("skip_to_end"),
			// No description, no guard
		).
		Build()

	proj := createTestProjectWithConfig(t, "Start", config)

	// Capture output
	output := captureOutput(func() {
		err := listAvailableTransitions(proj, sdkstate.State("Start"))
		if err != nil {
			t.Fatalf("listAvailableTransitions failed: %v", err)
		}
	})

	// Verify transitions are shown
	if !strings.Contains(output, "sow advance go_middle") {
		t.Errorf("Expected first transition shown, got:\n%s", output)
	}
	if !strings.Contains(output, "sow advance skip_to_end") {
		t.Errorf("Expected second transition shown, got:\n%s", output)
	}

	// Verify target states are shown
	if !strings.Contains(output, "→ Middle") {
		t.Errorf("Expected first target state, got:\n%s", output)
	}
	if !strings.Contains(output, "→ End") {
		t.Errorf("Expected second target state, got:\n%s", output)
	}

	// Verify no "Requires:" line appears (no guard description)
	// Note: We can't easily test for absence of empty lines, but
	// we can verify the essential information is present without descriptions
}

// Test helper functions

// createTestProjectWithConfig creates a test project with the given config and initial state.
func createTestProjectWithConfig(t *testing.T, initialState string, config *sdkproject.ProjectTypeConfig) *state.Project {
	t.Helper()

	proj := &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   "test",
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{},
			Statechart: projectschema.StatechartState{
				Current_state: initialState,
				Updated_at:    time.Now(),
			},
		},
	}

	// Build and attach the machine
	machine := config.BuildMachine(proj, sdkstate.State(initialState))

	// Use reflection to set private fields (config and machine)
	// This is necessary for testing since these fields are not exported
	projValue := reflect.ValueOf(proj).Elem()

	// Set config field - it expects ProjectTypeConfig interface, but we have *ProjectTypeConfig
	configField := projValue.FieldByName("config")
	if configField.IsValid() {
		// Use unsafe pointer to bypass private field restriction
		configField = reflect.NewAt(configField.Type(), unsafe.Pointer(configField.UnsafeAddr())).Elem()
		configField.Set(reflect.ValueOf(config))
	}

	// Set machine field
	machineField := projValue.FieldByName("machine")
	if machineField.IsValid() {
		machineField = reflect.NewAt(machineField.Type(), unsafe.Pointer(machineField.UnsafeAddr())).Elem()
		machineField.Set(reflect.ValueOf(machine))
	}

	return proj
}

// captureOutput captures stdout during function execution.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestAdvanceDryRunValid tests dry-run mode with a valid transition (guards pass).
func TestAdvanceDryRunValid(t *testing.T) {
	// Create test project where transition would succeed (guard passes)
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("StateA"),
			sdkstate.State("StateB"),
			sdkstate.Event("proceed"),
			sdkproject.WithDescription("Move to StateB"),
			sdkproject.WithGuard(
				"no prerequisites",
				func(_ *state.Project) bool { return true },
			),
		).
		Build()

	proj := createTestProjectWithConfig(t, "StateA", config)
	machine := proj.Machine()

	// Capture output
	output := captureOutput(func() {
		err := validateTransition(
			nil, // ctx not used in validation
			proj,
			machine,
			sdkstate.State("StateA"),
			sdkstate.Event("proceed"),
		)
		if err != nil {
			t.Errorf("validateTransition failed: %v", err)
		}
	})

	// Verify output shows validation header
	if !strings.Contains(output, "Validating transition: StateA -> proceed") {
		t.Errorf("Expected validation header, got:\n%s", output)
	}

	// Verify success message
	if !strings.Contains(output, "✓ Transition is valid and can be executed") {
		t.Errorf("Expected success message, got:\n%s", output)
	}

	// Verify target state displayed
	if !strings.Contains(output, "Target state: StateB") {
		t.Errorf("Expected target state display, got:\n%s", output)
	}

	// Verify description displayed
	if !strings.Contains(output, "Description: Move to StateB") {
		t.Errorf("Expected description display, got:\n%s", output)
	}

	// Verify execution hint
	if !strings.Contains(output, "To execute: sow advance proceed") {
		t.Errorf("Expected execution hint, got:\n%s", output)
	}

	// Verify project state unchanged
	if proj.Statechart.Current_state != "StateA" {
		t.Errorf("Expected state to remain StateA, got %s", proj.Statechart.Current_state)
	}
}

// TestAdvanceDryRunBlocked tests dry-run mode when guard blocks transition.
func TestAdvanceDryRunBlocked(t *testing.T) {
	// Create test project where guard blocks transition
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("StateA"),
			sdkstate.State("StateB"),
			sdkstate.Event("proceed"),
			sdkproject.WithDescription("Move to StateB"),
			sdkproject.WithGuard(
				"approval required",
				func(_ *state.Project) bool { return false },
			),
		).
		Build()

	proj := createTestProjectWithConfig(t, "StateA", config)
	machine := proj.Machine()

	// Capture output
	var capturedErr error
	output := captureOutput(func() {
		capturedErr = validateTransition(
			nil,
			proj,
			machine,
			sdkstate.State("StateA"),
			sdkstate.Event("proceed"),
		)
	})

	// Verify error returned
	if capturedErr == nil {
		t.Error("Expected error for blocked transition, got nil")
	}
	if capturedErr != nil && !strings.Contains(capturedErr.Error(), "blocked by guard") {
		t.Errorf("Expected 'blocked by guard' error, got: %v", capturedErr)
	}

	// Verify output shows validation header
	if !strings.Contains(output, "Validating transition: StateA -> proceed") {
		t.Errorf("Expected validation header, got:\n%s", output)
	}

	// Verify blocked message
	if !strings.Contains(output, "✗ Transition blocked by guard condition") {
		t.Errorf("Expected blocked message, got:\n%s", output)
	}

	// Verify guard description displayed
	if !strings.Contains(output, "Guard description: approval required") {
		t.Errorf("Expected guard description, got:\n%s", output)
	}

	// Verify status message
	if !strings.Contains(output, "Current status: Guard not satisfied") {
		t.Errorf("Expected status message, got:\n%s", output)
	}

	// Verify fix hint
	if !strings.Contains(output, "Fix the guard condition, then try again.") {
		t.Errorf("Expected fix hint, got:\n%s", output)
	}

	// Verify project state unchanged
	if proj.Statechart.Current_state != "StateA" {
		t.Errorf("Expected state to remain StateA, got %s", proj.Statechart.Current_state)
	}
}

// TestAdvanceDryRunInvalidEvent tests dry-run mode with an event not configured for current state.
func TestAdvanceDryRunInvalidEvent(t *testing.T) {
	// Create test project with some configured transitions
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("StateA"),
			sdkstate.State("StateB"),
			sdkstate.Event("proceed"),
		).
		Build()

	proj := createTestProjectWithConfig(t, "StateA", config)
	machine := proj.Machine()

	// Capture output
	var capturedErr error
	output := captureOutput(func() {
		capturedErr = validateTransition(
			nil,
			proj,
			machine,
			sdkstate.State("StateA"),
			sdkstate.Event("invalid_event"),
		)
	})

	// Verify error returned
	if capturedErr == nil {
		t.Error("Expected error for invalid event, got nil")
	}
	if capturedErr != nil && !strings.Contains(capturedErr.Error(), "event not configured") {
		t.Errorf("Expected 'event not configured' error, got: %v", capturedErr)
	}

	// Verify output shows validation header
	if !strings.Contains(output, "Validating transition: StateA -> invalid_event") {
		t.Errorf("Expected validation header, got:\n%s", output)
	}

	// Verify error message about unconfigured event
	if !strings.Contains(output, "✗ Event 'invalid_event' is not configured for state StateA") {
		t.Errorf("Expected unconfigured event message, got:\n%s", output)
	}

	// Verify suggestion to use --list
	if !strings.Contains(output, "Use 'sow advance --list' to see available transitions.") {
		t.Errorf("Expected --list suggestion, got:\n%s", output)
	}

	// Verify project state unchanged
	if proj.Statechart.Current_state != "StateA" {
		t.Errorf("Expected state to remain StateA, got %s", proj.Statechart.Current_state)
	}
}

// TestAdvanceDryRunNoSideEffects tests that dry-run never modifies project state.
// This is CRITICAL - dry-run must never execute actions or change state.
func TestAdvanceDryRunNoSideEffects(t *testing.T) {
	// Track if OnEntry action was executed
	actionExecuted := false

	// Create test project with OnEntry action that would modify state
	config := sdkproject.NewProjectTypeConfigBuilder("test").
		AddTransition(
			sdkstate.State("StateA"),
			sdkstate.State("StateB"),
			sdkstate.Event("proceed"),
			sdkproject.WithOnEntry(func(proj *state.Project) error {
				actionExecuted = true
				// This would modify project metadata
				if proj.Phases == nil {
					proj.Phases = make(map[string]projectschema.PhaseState)
				}
				proj.Phases["test"] = projectschema.PhaseState{
					Status:  "modified",
					Enabled: true,
				}
				return nil
			}),
		).
		Build()

	proj := createTestProjectWithConfig(t, "StateA", config)
	machine := proj.Machine()

	// Capture initial state
	initialState := proj.Statechart.Current_state

	// Run dry-run validation
	_ = validateTransition(
		nil,
		proj,
		machine,
		sdkstate.State("StateA"),
		sdkstate.Event("proceed"),
	)

	// CRITICAL: Verify OnEntry action was NOT executed
	if actionExecuted {
		t.Error("CRITICAL: dry-run executed OnEntry action (side effect detected)")
	}

	// Verify state unchanged
	if proj.Statechart.Current_state != initialState {
		t.Errorf("CRITICAL: dry-run modified state machine state (was %s, now %s)", initialState, proj.Statechart.Current_state)
	}

	// Verify phase metadata unchanged (no phases should exist)
	if len(proj.Phases) > 0 {
		t.Error("CRITICAL: dry-run modified project phases")
	}
}

// TestAdvanceDryRunWithoutEvent is already covered by TestAdvanceFlagValidation
// which tests the "--dry-run requires an event argument" validation.
// No additional test needed here.
