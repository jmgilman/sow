package config

import (
	"bytes"
	"testing"
)

// TestNewConfigCmd_Structure verifies the config command has the correct structure.
func TestNewConfigCmd_Structure(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify command properties
	if cmd.Use != "config" {
		t.Errorf("expected Use='config', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewConfigCmd_ShortDescription verifies the short description is correct.
func TestNewConfigCmd_ShortDescription(t *testing.T) {
	cmd := NewConfigCmd()

	expected := "Manage user configuration"
	if cmd.Short != expected {
		t.Errorf("expected Short='%s', got '%s'", expected, cmd.Short)
	}
}

// TestNewConfigCmd_LongDescription verifies the long description has key content.
func TestNewConfigCmd_LongDescription(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify key content in Long description
	expectedPhrases := []string{
		"configuration",
		"config.yaml",
		"defaults",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain '%s', but it doesn't", phrase)
		}
	}
}

// TestNewConfigCmd_NoErrorWhenRun verifies running without subcommand shows help (not error).
func TestNewConfigCmd_NoErrorWhenRun(t *testing.T) {
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Running without subcommand should show help (not error)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestNewConfigCmd_HelpOutput verifies help output contains expected content.
func TestNewConfigCmd_HelpOutput(t *testing.T) {
	cmd := NewConfigCmd()
	cmd.SetArgs([]string{"--help"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute with help flag
	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify help output contains subcommand descriptions
	expectedInHelp := []string{
		"init",
		"path",
		"show",
		"validate",
		"edit",
		"reset",
	}

	for _, phrase := range expectedInHelp {
		if !containsSubstring(output, phrase) {
			t.Errorf("expected help output to contain '%s', but it doesn't", phrase)
		}
	}
}

// TestNewConfigCmd_NoSubcommandsYet verifies no subcommands are added yet.
// Subcommands will be added in subsequent tasks.
func TestNewConfigCmd_NoSubcommandsYet(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify no subcommands exist yet (they'll be added in other tasks)
	subcommands := cmd.Commands()
	if len(subcommands) != 0 {
		t.Errorf("expected 0 subcommands (not implemented yet), got %d", len(subcommands))
		for _, sub := range subcommands {
			t.Logf("  found subcommand: %s", sub.Use)
		}
	}
}

// containsSubstring checks if a string contains a substring.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
