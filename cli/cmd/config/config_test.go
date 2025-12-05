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

// TestNewConfigCmd_HasInitSubcommand verifies init subcommand is registered.
func TestNewConfigCmd_HasInitSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify init subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'init' subcommand to be registered")
	}
}

// TestNewConfigCmd_HasPathSubcommand verifies path subcommand is registered.
func TestNewConfigCmd_HasPathSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify path subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "path" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'path' subcommand to be registered")
	}
}

// TestNewConfigCmd_HasShowSubcommand verifies show subcommand is registered.
func TestNewConfigCmd_HasShowSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify show subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "show" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'show' subcommand to be registered")
	}
}

// TestNewConfigCmd_HasValidateSubcommand verifies validate subcommand is registered.
func TestNewConfigCmd_HasValidateSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify validate subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "validate" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'validate' subcommand to be registered")
	}
}

// TestNewConfigCmd_HasEditSubcommand verifies edit subcommand is registered.
func TestNewConfigCmd_HasEditSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify edit subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "edit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'edit' subcommand to be registered")
	}
}

// TestNewConfigCmd_HasResetSubcommand verifies reset subcommand is registered.
func TestNewConfigCmd_HasResetSubcommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify reset subcommand exists
	subcommands := cmd.Commands()
	found := false
	for _, sub := range subcommands {
		if sub.Use == "reset" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'reset' subcommand to be registered")
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
