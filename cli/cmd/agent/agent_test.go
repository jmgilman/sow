package agent

import (
	"strings"
	"testing"
)

// TestNewAgentCmd_Structure verifies the parent command has correct structure.
func TestNewAgentCmd_Structure(t *testing.T) {
	cmd := NewAgentCmd()

	if cmd.Use != "agent" {
		t.Errorf("expected Use='agent', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewAgentCmd_ShortDescription verifies the short description is accurate.
func TestNewAgentCmd_ShortDescription(t *testing.T) {
	cmd := NewAgentCmd()

	if cmd.Short != "Manage AI agents" {
		t.Errorf("expected Short='Manage AI agents', got '%s'", cmd.Short)
	}
}

// TestNewAgentCmd_LongDescription verifies the long description contains
// helpful information about available subcommands.
func TestNewAgentCmd_LongDescription(t *testing.T) {
	cmd := NewAgentCmd()

	expectedPhrases := []string{
		"agent",
		"list",
		"spawn",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}

// TestNewAgentCmd_HasSubcommands verifies that subcommands are registered.
func TestNewAgentCmd_HasSubcommands(t *testing.T) {
	cmd := NewAgentCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("expected agent command to have subcommands")
	}

	// Check for expected subcommands
	expectedSubcmds := map[string]bool{
		"list":                    false,
		"spawn [task-id]":         false,
		"resume [task-id] <prompt>": false,
	}

	for _, sub := range subcommands {
		if _, expected := expectedSubcmds[sub.Use]; expected {
			expectedSubcmds[sub.Use] = true
		}
	}

	for subcmd, found := range expectedSubcmds {
		if !found {
			t.Errorf("expected '%s' subcommand to be registered", subcmd)
		}
	}
}
