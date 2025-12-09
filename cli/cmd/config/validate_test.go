package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmgilman/sow/libs/schemas"
	"github.com/spf13/cobra"
)

// TestNewValidateCmd_Structure verifies the validate command has correct structure.
func TestNewValidateCmd_Structure(t *testing.T) {
	cmd := newValidateCmd()

	if cmd.Use != "validate" {
		t.Errorf("expected Use='validate', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewValidateCmd_LongDescription verifies the long description contains key info.
func TestNewValidateCmd_LongDescription(t *testing.T) {
	cmd := newValidateCmd()

	expectedPhrases := []string{
		"YAML",
		"syntax",
		"Executor",
		"Bindings",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain '%s'", phrase)
		}
	}
}

// TestRunValidate_NoConfigFile verifies that validate reports missing file without error.
func TestRunValidate_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should report "No configuration file found"
	if !strings.Contains(output, "No configuration file found") {
		t.Errorf("expected output to contain 'No configuration file found', got: %s", output)
	}

	// Should suggest running config init
	if !strings.Contains(output, "sow config init") {
		t.Errorf("expected output to suggest 'sow config init', got: %s", output)
	}
}

// TestRunValidate_ValidConfig verifies that validate reports success for valid config.
func TestRunValidate_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	validConfig := `agents:
  executors:
    claude-code:
      type: claude
      settings:
        yolo_mode: false
  bindings:
    orchestrator: claude-code
    implementer: claude-code
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain "OK" messages for each check
	if !strings.Contains(output, "OK YAML syntax valid") {
		t.Errorf("expected 'OK YAML syntax valid' in output, got: %s", output)
	}
	if !strings.Contains(output, "OK Schema valid") {
		t.Errorf("expected 'OK Schema valid' in output, got: %s", output)
	}
	if !strings.Contains(output, "OK Executor types valid") {
		t.Errorf("expected 'OK Executor types valid' in output, got: %s", output)
	}
	if !strings.Contains(output, "OK Bindings reference defined executors") {
		t.Errorf("expected 'OK Bindings reference defined executors' in output, got: %s", output)
	}

	// Should end with success message
	if !strings.Contains(output, "Configuration is valid") {
		t.Errorf("expected 'Configuration is valid' in output, got: %s", output)
	}
}

// TestRunValidate_InvalidYAML verifies that YAML syntax errors are caught.
func TestRunValidate_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write invalid YAML
	invalidYAML := `agents:
  bindings:
    implementer: [invalid: yaml
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	output := buf.String()

	// Should contain "X" for syntax error
	if !strings.Contains(output, "X YAML syntax error") {
		t.Errorf("expected 'X YAML syntax error' in output, got: %s", output)
	}

	// Error should mention "validation failed"
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected error to contain 'validation failed', got: %v", err)
	}
}

// TestRunValidate_InvalidExecutorType verifies that invalid executor types are caught.
func TestRunValidate_InvalidExecutorType(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config with invalid executor type
	invalidConfig := `agents:
  executors:
    my-copilot:
      type: copilot
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err == nil {
		t.Fatal("expected error for invalid executor type, got nil")
	}

	output := buf.String()

	// Should have passed YAML check first
	if !strings.Contains(output, "OK YAML syntax valid") {
		t.Errorf("expected 'OK YAML syntax valid' before failure, got: %s", output)
	}

	// Should contain "X" for validation error
	if !strings.Contains(output, "X Validation error") {
		t.Errorf("expected 'X Validation error' in output, got: %s", output)
	}

	// Should mention the invalid type
	if !strings.Contains(output, "copilot") {
		t.Errorf("expected error to mention 'copilot', got: %s", output)
	}
}

// TestRunValidate_UndefinedExecutorBinding verifies bindings to undefined executors are caught.
func TestRunValidate_UndefinedExecutorBinding(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config with binding to undefined executor
	invalidConfig := `agents:
  bindings:
    implementer: nonexistent-executor
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err == nil {
		t.Fatal("expected error for binding to undefined executor, got nil")
	}

	output := buf.String()

	// Should have passed YAML check first
	if !strings.Contains(output, "OK YAML syntax valid") {
		t.Errorf("expected 'OK YAML syntax valid' before failure, got: %s", output)
	}

	// Should contain "X" for validation error
	if !strings.Contains(output, "X Validation error") {
		t.Errorf("expected 'X Validation error' in output, got: %s", output)
	}

	// Should mention the undefined executor
	if !strings.Contains(output, "nonexistent-executor") {
		t.Errorf("expected error to mention 'nonexistent-executor', got: %s", output)
	}
}

// TestCheckExecutorBinaries_MissingBinary verifies warnings for missing binaries.
func TestCheckExecutorBinaries_MissingBinary(t *testing.T) {
	// Create a config with an executor whose binary likely doesn't exist
	// Using "windsurf" as it's unlikely to be installed on CI
	config := &schemas.UserConfig{
		Agents: &struct {
			Executors map[string]struct {
				Type     string `json:"type"`
				Settings *struct {
					Yolo_mode *bool   `json:"yolo_mode,omitempty"`
					Model     *string `json:"model,omitempty"`
				} `json:"settings,omitempty"`
				Custom_args []string `json:"custom_args,omitempty"`
			} `json:"executors,omitempty"`
			Bindings *struct {
				Orchestrator *string `json:"orchestrator,omitempty"`
				Implementer  *string `json:"implementer,omitempty"`
				Architect    *string `json:"architect,omitempty"`
				Reviewer     *string `json:"reviewer,omitempty"`
				Planner      *string `json:"planner,omitempty"`
				Researcher   *string `json:"researcher,omitempty"`
				Decomposer   *string `json:"decomposer,omitempty"`
			} `json:"bindings,omitempty"`
		}{
			Executors: map[string]struct {
				Type     string `json:"type"`
				Settings *struct {
					Yolo_mode *bool   `json:"yolo_mode,omitempty"`
					Model     *string `json:"model,omitempty"`
				} `json:"settings,omitempty"`
				Custom_args []string `json:"custom_args,omitempty"`
			}{
				"windsurf-exec": {
					Type: "windsurf",
				},
			},
		},
	}

	warnings := checkExecutorBinariesFromConfig(config)

	// Should have at least one warning (windsurf binary not found)
	hasWindsurfWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "windsurf") {
			hasWindsurfWarning = true
			break
		}
	}

	if !hasWindsurfWarning {
		t.Errorf("expected warning about windsurf binary, got: %v", warnings)
	}
}

// TestCheckExecutorBinaries_NilConfig verifies no panic with nil config.
func TestCheckExecutorBinaries_NilConfig(t *testing.T) {
	warnings := checkExecutorBinariesFromConfig(nil)

	if len(warnings) != 0 {
		t.Errorf("expected empty warnings for nil config, got: %v", warnings)
	}
}

// TestCheckExecutorBinaries_NilAgents verifies no panic with nil agents.
func TestCheckExecutorBinaries_NilAgents(t *testing.T) {
	config := &schemas.UserConfig{Agents: nil}
	warnings := checkExecutorBinariesFromConfig(config)

	if len(warnings) != 0 {
		t.Errorf("expected empty warnings for nil agents, got: %v", warnings)
	}
}

// TestRunValidate_ValidWithWarnings verifies warnings don't cause failure.
func TestRunValidate_ValidWithWarnings(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write valid config with windsurf executor (binary unlikely to exist)
	validConfig := `agents:
  executors:
    windsurf-exec:
      type: windsurf
  bindings:
    implementer: windsurf-exec
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Should not return error even with warnings
	err := runValidateWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error (warnings should not cause failure): %v", err)
	}

	output := buf.String()

	// Should contain all OK messages
	if !strings.Contains(output, "OK YAML syntax valid") {
		t.Errorf("expected 'OK YAML syntax valid' in output, got: %s", output)
	}

	// Should contain warning about windsurf binary
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected 'WARN' in output, got: %s", output)
	}

	// Should indicate valid with warnings
	if !strings.Contains(output, "valid with") && !strings.Contains(output, "warning") {
		t.Errorf("expected output to mention warnings, got: %s", output)
	}
}

// TestRunValidate_OutputsConfigPath verifies the path is shown in output.
func TestRunValidate_OutputsConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	validConfig := `agents:
  executors:
    claude-code:
      type: claude
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain the config path
	if !strings.Contains(output, configPath) {
		t.Errorf("expected output to contain config path %q, got: %s", configPath, output)
	}
}

// TestRunValidate_EmptyConfig verifies empty config file passes validation.
func TestRunValidate_EmptyConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write empty config file (just a comment)
	emptyConfig := "# Empty config\n"
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runValidateWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error for empty config: %v", err)
	}

	output := buf.String()

	// Should pass validation
	if !strings.Contains(output, "Configuration is valid") {
		t.Errorf("expected 'Configuration is valid' in output, got: %s", output)
	}
}

