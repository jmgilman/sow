package config

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestGetEditor_RespectsEnvVar verifies getEditor returns the EDITOR environment variable.
func TestGetEditor_RespectsEnvVar(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	editor := getEditor()
	if editor != "nano" {
		t.Errorf("expected 'nano', got '%s'", editor)
	}
}

// TestGetEditor_FallsBackToVi verifies getEditor returns "vi" when EDITOR is not set.
func TestGetEditor_FallsBackToVi(t *testing.T) {
	t.Setenv("EDITOR", "")
	editor := getEditor()
	if editor != "vi" {
		t.Errorf("expected 'vi', got '%s'", editor)
	}
}

// TestGetEditor_HandlesEmptyString verifies getEditor handles empty string.
func TestGetEditor_HandlesEmptyString(t *testing.T) {
	// Explicitly set to empty string
	t.Setenv("EDITOR", "")
	editor := getEditor()
	if editor != "vi" {
		t.Errorf("expected 'vi', got '%s'", editor)
	}
}

// TestRunEditWithPath_CreatesFileIfMissing verifies that runEditWithPath creates
// a config file with template content when file doesn't exist.
func TestRunEditWithPath_CreatesFileIfMissing(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create a command with custom output buffer and context
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Use true (shell builtin) as the editor - it does nothing and exits successfully
	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created")
	}

	// Verify content matches template
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if string(content) != configTemplate {
		t.Error("config file content does not match template")
	}
}

// TestRunEditWithPath_OutputsCreationMessage verifies that runEditWithPath
// prints a creation message when file is created.
func TestRunEditWithPath_OutputsCreationMessage(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !containsSubstring(output, "Created new configuration") {
		t.Errorf("expected output to contain 'Created new configuration', got: %s", output)
	}
	if !containsSubstring(output, configPath) {
		t.Errorf("expected output to contain path %q, got: %s", configPath, output)
	}
}

// TestRunEditWithPath_DoesNotOutputCreationMessageForExisting verifies that
// runEditWithPath does not print creation message when file already exists.
func TestRunEditWithPath_DoesNotOutputCreationMessageForExisting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create existing file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if containsSubstring(output, "Created") {
		t.Errorf("expected no creation message for existing file, got: %s", output)
	}
}

// TestRunEditWithPath_UsesExistingFile verifies that runEditWithPath opens
// existing file without overwriting it.
func TestRunEditWithPath_UsesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")
	existingContent := "# My custom config\nagents:\n  bindings:\n    orchestrator: custom"

	// Create existing file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file content was not modified
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if string(content) != existingContent {
		t.Errorf("expected existing content to be preserved, got: %s", string(content))
	}
}

// TestRunEditWithPath_CreatesParentDirectories verifies that runEditWithPath
// creates parent directories when they don't exist.
func TestRunEditWithPath_CreatesParentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "deeply", "nested", "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created in nested path")
	}
}

// TestRunEditWithPath_FilePermissions verifies that created file has 0644 permissions.
func TestRunEditWithPath_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("expected file permissions 0644, got %04o", perm)
	}
}

// TestRunEditWithPath_DirectoryPermissions verifies that created directory has 0755 permissions.
func TestRunEditWithPath_DirectoryPermissions(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parentInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("failed to stat parent directory: %v", err)
	}

	perm := parentInfo.Mode().Perm()
	if perm != 0755 {
		t.Errorf("expected parent directory permissions 0755, got %04o", perm)
	}
}

// TestRunEditWithPath_EditorError verifies that editor errors are propagated.
func TestRunEditWithPath_EditorError(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Use false (shell builtin) as the editor - it exits with error code 1
	err := runEditWithPath(cmd, configPath, "false")
	if err == nil {
		t.Fatal("expected error from editor, got nil")
	}

	if !containsSubstring(err.Error(), "editor failed") {
		t.Errorf("expected error to mention 'editor failed', got: %s", err.Error())
	}
}

// TestRunEditWithPath_NonExistentEditor verifies error handling for non-existent editor.
func TestRunEditWithPath_NonExistentEditor(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEditWithPath(cmd, configPath, "nonexistent-editor-12345")
	if err == nil {
		t.Fatal("expected error for non-existent editor, got nil")
	}

	// Should contain "editor failed"
	if !containsSubstring(err.Error(), "editor failed") {
		t.Errorf("expected error to mention 'editor failed', got: %s", err.Error())
	}
}

// TestNewEditCmd_Structure verifies the edit command has correct structure.
func TestNewEditCmd_Structure(t *testing.T) {
	cmd := newEditCmd()

	if cmd.Use != "edit" {
		t.Errorf("expected Use='edit', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewEditCmd_LongDescription verifies the long description contains helpful information.
func TestNewEditCmd_LongDescription(t *testing.T) {
	cmd := newEditCmd()

	expectedPhrases := []string{
		"$EDITOR",
		"vi",
		"configuration file",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}
