package commands

import (
	"bytes"
	"strings"
	"testing"
)

// TestRootCommandExists verifies the root command can be created
func TestRootCommandExists(t *testing.T) {
	rootCmd := NewRootCmd()

	if rootCmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}

	if rootCmd.Use != "sow" {
		t.Errorf("Root command Use = %q, want %q", rootCmd.Use, "sow")
	}
}

// TestRootCommandHasVersion verifies the root command has version info
func TestRootCommandHasVersion(t *testing.T) {
	rootCmd := NewRootCmd()

	if rootCmd.Version == "" {
		t.Error("Root command Version is empty")
	}
}

// TestRootCommandHasGlobalFlags verifies global flags exist
func TestRootCommandHasGlobalFlags(t *testing.T) {
	rootCmd := NewRootCmd()

	flags := []string{"quiet", "verbose", "no-color"}

	for _, flagName := range flags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Global flag %q not found", flagName)
		}
	}
}

// TestRootCommandHasSubcommands verifies subcommands are registered
func TestRootCommandHasSubcommands(t *testing.T) {
	rootCmd := NewRootCmd()

	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("Root command has no subcommands")
	}

	// Verify version command exists
	hasVersion := false
	for _, cmd := range commands {
		if cmd.Use == "version" {
			hasVersion = true
			break
		}
	}

	if !hasVersion {
		t.Error("Root command missing 'version' subcommand")
	}
}

// TestVersionCommandOutput verifies version command produces output
func TestVersionCommandOutput(t *testing.T) {
	rootCmd := NewRootCmd()

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set args to run version command
	rootCmd.SetArgs([]string{"version"})

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains "sow"
	if !strings.Contains(output, "sow") {
		t.Errorf("version output missing 'sow': %s", output)
	}
}

// TestVersionCommandShowsVersion verifies version command displays version
func TestVersionCommandShowsVersion(t *testing.T) {
	versionCmd := NewVersionCmd()

	if versionCmd == nil {
		t.Fatal("NewVersionCmd() returned nil")
	}

	if versionCmd.Use != "version" {
		t.Errorf("Version command Use = %q, want %q", versionCmd.Use, "version")
	}

	// Verify it has a Run function
	if versionCmd.Run == nil {
		t.Error("Version command has no Run function")
	}
}
