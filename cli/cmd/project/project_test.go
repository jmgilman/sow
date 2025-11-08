package project

import (
	"testing"
)

// TestProjectCmd_Structure verifies that the project command has the correct structure.
// after the wizard integration.
func TestProjectCmd_Structure(t *testing.T) {
	cmd := NewProjectCmd()

	// Verify basic properties - Use field now includes Claude flag syntax
	expectedUse := "project [-- <claude-flags>...]"
	if cmd.Use != expectedUse {
		t.Errorf("Expected Use to be '%s', got '%s'", expectedUse, cmd.Use)
	}

	if cmd.Short != "Create or continue a project (interactive)" {
		t.Errorf("Expected Short description about interactive wizard, got '%s'", cmd.Short)
	}

	// Verify wizard is the main command (has RunE)
	if cmd.RunE == nil {
		t.Error("Expected project command to have RunE (wizard handler), but it's nil")
	}

	// Verify Args is NOT set (nil) to allow pass-through flags after --
	if cmd.Args != nil {
		t.Error("Expected project command to have no Args validator (to allow -- pass-through), but it's set")
	}
}

// TestProjectCmd_HasCorrectSubcommands verifies that only set and delete subcommands exist.
// (new and continue should be removed).
func TestProjectCmd_HasCorrectSubcommands(t *testing.T) {
	cmd := NewProjectCmd()

	// Get all subcommands
	subcommands := cmd.Commands()

	// Helper function to check if any subcommand starts with a given prefix
	hasCommandWithPrefix := func(prefix string) bool {
		for _, subcmd := range subcommands {
			if len(subcmd.Use) >= len(prefix) && subcmd.Use[:len(prefix)] == prefix {
				return true
			}
		}
		return false
	}

	// Verify 'set' and 'delete' exist (check by prefix since they may have args in Use)
	expectedCommands := []string{"set", "delete"}
	for _, expected := range expectedCommands {
		if !hasCommandWithPrefix(expected) {
			t.Errorf("Expected subcommand starting with '%s' to exist, but it doesn't", expected)
		}
	}

	// Verify 'new', 'continue', and 'wizard' do NOT exist as subcommands
	removedCommands := []string{"new", "continue", "wizard"}
	for _, removed := range removedCommands {
		if hasCommandWithPrefix(removed) {
			t.Errorf("Expected subcommand starting with '%s' to be removed, but it still exists", removed)
		}
	}

	// Verify we have exactly 2 subcommands (set and delete)
	if len(subcommands) != 2 {
		t.Errorf("Expected exactly 2 subcommands (set and delete), got %d", len(subcommands))
		t.Log("Subcommands found:")
		for _, subcmd := range subcommands {
			t.Logf("  - %s", subcmd.Use)
		}
	}
}

// TestProjectCmd_LongDescription verifies the long description mentions the wizard.
// and provides examples.
func TestProjectCmd_LongDescription(t *testing.T) {
	cmd := NewProjectCmd()

	if cmd.Long == "" {
		t.Error("Expected Long description to be non-empty")
	}

	// Verify key content in Long description
	expectedPhrases := []string{
		"interactive",
		"wizard",
		"sow project",
	}

	for _, phrase := range expectedPhrases {
		if !contains(cmd.Long, phrase) {
			t.Errorf("Expected Long description to contain '%s', but it doesn't", phrase)
		}
	}
}

// TestProjectCmd_OldCommandsRemoved verifies that old subcommands are truly gone.
// by attempting to find them in the command tree.
func TestProjectCmd_OldCommandsRemoved(t *testing.T) {
	cmd := NewProjectCmd()

	// Try to find 'new' subcommand
	newCmd, _, err := cmd.Find([]string{"new"})
	if err == nil && newCmd != cmd {
		t.Error("Expected 'new' subcommand to not exist, but it was found")
	}

	// Try to find 'continue' subcommand
	continueCmd, _, err := cmd.Find([]string{"continue"})
	if err == nil && continueCmd != cmd {
		t.Error("Expected 'continue' subcommand to not exist, but it was found")
	}

	// Try to find 'wizard' subcommand
	wizardCmd, _, err := cmd.Find([]string{"wizard"})
	if err == nil && wizardCmd != cmd {
		t.Error("Expected 'wizard' subcommand to not exist, but it was found")
	}
}

// TestProjectCmd_KeptCommandsExist verifies that set and delete still work.
func TestProjectCmd_KeptCommandsExist(t *testing.T) {
	cmd := NewProjectCmd()

	// Try to find 'set' subcommand
	setCmd, _, err := cmd.Find([]string{"set"})
	if err != nil {
		t.Errorf("Expected 'set' subcommand to exist, but got error: %v", err)
	}
	if setCmd == cmd {
		t.Error("Expected to find 'set' subcommand, but got parent command")
	}

	// Try to find 'delete' subcommand
	deleteCmd, _, err := cmd.Find([]string{"delete"})
	if err != nil {
		t.Errorf("Expected 'delete' subcommand to exist, but got error: %v", err)
	}
	if deleteCmd == cmd {
		t.Error("Expected to find 'delete' subcommand, but got parent command")
	}
}

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
