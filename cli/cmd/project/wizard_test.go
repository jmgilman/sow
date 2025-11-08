package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWizard_InitializesCorrectly tests that NewWizard creates a wizard with correct initial state.
func TestNewWizard_InitializesCorrectly(t *testing.T) {
	// Create a mock context (nil is acceptable for initialization test)
	var ctx *sow.Context
	claudeFlags := []string{"--model", "opus"}

	wizard := NewWizard(ctx, claudeFlags)

	// Verify wizard starts in StateEntry
	assert.Equal(t, StateEntry, wizard.state, "Wizard should start in StateEntry")

	// Verify choices map is initialized (not nil)
	assert.NotNil(t, wizard.choices, "Choices map should be initialized")
	assert.Empty(t, wizard.choices, "Choices map should be empty initially")

	// Verify context is stored
	assert.Equal(t, ctx, wizard.ctx, "Context should be stored")

	// Verify flags are stored
	assert.Equal(t, claudeFlags, wizard.claudeFlags, "Claude flags should be stored")
}

// TestHandleState_UnknownStateReturnsError tests that handleState returns an error for unknown states.
func TestHandleState_UnknownStateReturnsError(t *testing.T) {
	wizard := NewWizard(nil, nil)

	// Set wizard to an unknown/invalid state
	wizard.state = WizardState("invalid_state")

	err := wizard.handleState()

	require.Error(t, err, "handleState should return error for unknown state")
	assert.Contains(t, err.Error(), "unknown state", "Error should mention unknown state")
	assert.Contains(t, err.Error(), "invalid_state", "Error should include the state name")
}

// TestHandleState_DispatchesToCorrectHandler tests that handleState dispatches to the right handler.
func TestHandleState_DispatchesToCorrectHandler(t *testing.T) {
	tests := []struct {
		name  string
		state WizardState
	}{
		{"Entry state", StateEntry},
		{"CreateSource state", StateCreateSource},
		{"IssueSelect state", StateIssueSelect},
		{"TypeSelect state", StateTypeSelect},
		{"NameEntry state", StateNameEntry},
		{"PromptEntry state", StatePromptEntry},
		{"ProjectSelect state", StateProjectSelect},
		{"ContinuePrompt state", StateContinuePrompt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := NewWizard(nil, nil)
			wizard.state = tt.state

			// For stub handlers, we expect them to transition to StateComplete
			// For interactive handlers (Entry, CreateSource), they need user input (we can't fully test here)
			// Just verify no error for valid states
			err := wizard.handleState()

			// Entry and CreateSource states require interactive input, skip for now
			// Other states (stubs) should complete successfully
			if tt.state == StateEntry || tt.state == StateCreateSource {
				// These require interactive input, skip for now
				t.Skip("Interactive state requires user input")
			} else {
				assert.NoError(t, err, "Valid state should not error")
				assert.Equal(t, StateComplete, wizard.state, "Stub handlers should transition to StateComplete")
			}
		})
	}
}

// TestWizardRun_LoopsUntilTerminalState tests that Run() loops until terminal state.
func TestWizardRun_LoopsUntilTerminalState(t *testing.T) {
	t.Run("exits on StateComplete", func(t *testing.T) {
		wizard := NewWizard(nil, nil)
		// Start in a stub state that will transition to complete
		wizard.state = StateIssueSelect

		err := wizard.Run()

		assert.NoError(t, err, "Run should complete without error")
		assert.Equal(t, StateComplete, wizard.state, "Should end in StateComplete")
	})

	t.Run("exits on StateCancelled", func(t *testing.T) {
		wizard := NewWizard(nil, nil)
		wizard.state = StateCancelled

		err := wizard.Run()

		assert.NoError(t, err, "Run should return nil for cancellation")
		assert.Equal(t, StateCancelled, wizard.state, "Should stay in StateCancelled")
	})
}

// TestStateTransitions_StubHandlers tests that stub handlers transition to StateComplete.
func TestStateTransitions_StubHandlers(t *testing.T) {
	stubs := []WizardState{
		// Note: StateCreateSource is now implemented, so it's not a stub anymore
		StateIssueSelect,
		StateTypeSelect,
		StateNameEntry,
		StatePromptEntry,
		StateProjectSelect,
		StateContinuePrompt,
	}

	for _, state := range stubs {
		t.Run(string(state), func(t *testing.T) {
			wizard := NewWizard(nil, nil)
			wizard.state = state

			err := wizard.handleState()

			assert.NoError(t, err, "Stub handler should not error")
			assert.Equal(t, StateComplete, wizard.state, "Stub should transition to StateComplete")
		})
	}
}

// Note: Testing entry screen with actual user input requires integration tests
// or mocking the huh library, which is complex. We'll verify the state machine
// structure is correct and do manual testing for the interactive parts.
