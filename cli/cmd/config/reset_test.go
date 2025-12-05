package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestResetConfigAtPath_NoFile verifies that resetConfigAtPath prints
// "No configuration file to reset" when file doesn't exist.
func TestResetConfigAtPath_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, true) // force=true to skip prompt
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !containsSubstring(output, "No configuration file to reset") {
		t.Errorf("expected output to contain 'No configuration file to reset', got: %s", output)
	}
}

// TestResetConfigAtPath_WithForce verifies that resetConfigAtPath removes
// the file and creates a backup when --force is used.
func TestResetConfigAtPath_WithForce(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")
	backupPath := configPath + ".backup"
	originalContent := "agents:\n  bindings:\n    orchestrator: custom"

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, true) // force=true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file is removed
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("expected config file to be removed")
	}

	// Verify backup was created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("expected backup file to be created")
	}

	// Verify backup contains original content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}
	if string(backupContent) != originalContent {
		t.Errorf("expected backup content '%s', got '%s'", originalContent, string(backupContent))
	}
}

// TestResetConfigAtPath_ConfirmYes verifies that resetConfigAtPath removes
// the file when user confirms with "y".
func TestResetConfigAtPath_ConfirmYes(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")
	originalContent := "test content"

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("y\n"))
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, false) // force=false
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file is removed
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("expected config file to be removed")
	}

	// Verify output mentions removal and backup
	output := buf.String()
	if !containsSubstring(output, "Configuration removed") {
		t.Errorf("expected output to contain 'Configuration removed', got: %s", output)
	}
}

// TestResetConfigAtPath_ConfirmNo verifies that resetConfigAtPath does NOT
// remove the file when user declines with "n".
func TestResetConfigAtPath_ConfirmNo(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")
	originalContent := "test content"

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("n\n"))
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, false) // force=false
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file is NOT removed
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to still exist")
	}

	// Verify "Cancelled" message
	output := buf.String()
	if !containsSubstring(output, "Cancelled") {
		t.Errorf("expected output to contain 'Cancelled', got: %s", output)
	}
}

// TestResetConfigAtPath_CreatesBackup verifies that resetConfigAtPath creates
// a backup with the original content.
func TestResetConfigAtPath_CreatesBackup(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")
	backupPath := configPath + ".backup"
	originalContent := "# My custom config\nagents:\n  bindings:\n    orchestrator: custom"

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, true) // force=true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify backup was created with correct content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}
	if string(backupContent) != originalContent {
		t.Errorf("expected backup content '%s', got '%s'", originalContent, string(backupContent))
	}

	// Verify output mentions backup path
	output := buf.String()
	if !containsSubstring(output, backupPath) {
		t.Errorf("expected output to contain backup path '%s', got: %s", backupPath, output)
	}
}

// TestResetConfigAtPath_AcceptsYes verifies that "yes" (full word) is accepted.
func TestResetConfigAtPath_AcceptsYes(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("yes\n"))
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, false) // force=false
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("expected config file to be removed with 'yes' confirmation")
	}
}

// TestResetConfigAtPath_CaseInsensitive verifies confirmation is case insensitive.
func TestResetConfigAtPath_CaseInsensitive(t *testing.T) {
	testCases := []string{"Y", "YES", "Yes", "yEs"}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "sow", "config.yaml")

			// Create the config file
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}
			if err := os.WriteFile(configPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create config file: %v", err)
			}

			cmd := &cobra.Command{}
			cmd.SetIn(strings.NewReader(input + "\n"))
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := resetConfigAtPath(cmd, configPath, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify file was removed
			if _, err := os.Stat(configPath); !os.IsNotExist(err) {
				t.Errorf("expected config file to be removed with '%s' confirmation", input)
			}
		})
	}
}

// TestResetConfigAtPath_OutputsDefaultMessage verifies that "Using built-in defaults"
// is printed after reset.
func TestResetConfigAtPath_OutputsDefaultMessage(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !containsSubstring(output, "Using built-in defaults") {
		t.Errorf("expected output to contain 'Using built-in defaults', got: %s", output)
	}
}

// TestResetConfigAtPath_PromptsPath verifies that the prompt includes the config path.
func TestResetConfigAtPath_PromptsPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("n\n"))
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !containsSubstring(output, "This will remove") {
		t.Errorf("expected prompt to contain 'This will remove', got: %s", output)
	}
	if !containsSubstring(output, configPath) {
		t.Errorf("expected prompt to contain config path '%s', got: %s", configPath, output)
	}
}

// TestResetConfigAtPath_EmptyInput verifies that empty input (just Enter) cancels.
func TestResetConfigAtPath_EmptyInput(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("\n")) // Just Enter
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := resetConfigAtPath(cmd, configPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was NOT removed (default is cancel)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to still exist with empty input")
	}

	// Verify "Cancelled" message
	output := buf.String()
	if !containsSubstring(output, "Cancelled") {
		t.Errorf("expected output to contain 'Cancelled', got: %s", output)
	}
}

// TestNewResetCmd_HasForceFlag verifies the reset command has --force/-f flag.
func TestNewResetCmd_HasForceFlag(t *testing.T) {
	cmd := newResetCmd()

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("expected --force flag")
	}
	if flag.Shorthand != "f" {
		t.Errorf("expected -f shorthand, got '%s'", flag.Shorthand)
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got '%s'", flag.DefValue)
	}
}

// TestNewResetCmd_Structure verifies the reset command has correct structure.
func TestNewResetCmd_Structure(t *testing.T) {
	cmd := newResetCmd()

	if cmd.Use != "reset" {
		t.Errorf("expected Use='reset', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewResetCmd_LongDescription verifies the long description contains helpful info.
func TestNewResetCmd_LongDescription(t *testing.T) {
	cmd := newResetCmd()

	expectedPhrases := []string{
		"configuration file",
		"backup",
		"--force",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}
