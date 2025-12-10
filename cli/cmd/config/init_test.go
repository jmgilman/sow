package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/libs/config"
	"github.com/jmgilman/sow/libs/schemas"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TestInitConfigAtPath_CreatesFile verifies that initConfigAtPath creates
// a config file with the expected template content.
func TestInitConfigAtPath_CreatesFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	err := initConfigAtPath(configPath)
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

// TestInitConfigAtPath_ErrorsOnExistingFile verifies that initConfigAtPath
// returns an error when the file already exists.
func TestInitConfigAtPath_ErrorsOnExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create directory and file first
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	// Attempt to init should fail
	err := initConfigAtPath(configPath)
	if err == nil {
		t.Fatal("expected error for existing file, got nil")
	}

	// Error message should be helpful
	errStr := err.Error()
	if !containsSubstring(errStr, "already exists") {
		t.Errorf("expected error to mention 'already exists', got: %s", errStr)
	}
	if !containsSubstring(errStr, "sow config edit") {
		t.Errorf("expected error to suggest 'sow config edit', got: %s", errStr)
	}
}

// TestInitConfigAtPath_CreatesParentDirectories verifies that initConfigAtPath
// creates parent directories if they don't exist.
func TestInitConfigAtPath_CreatesParentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	// Use nested path where parent doesn't exist
	configPath := filepath.Join(tempDir, "deeply", "nested", "sow", "config.yaml")

	err := initConfigAtPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to be created in nested path")
	}

	// Verify parent directory has correct permissions (0755)
	parentInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("failed to stat parent directory: %v", err)
	}
	// Check permissions (mask off file type bits)
	perm := parentInfo.Mode().Perm()
	if perm != 0755 {
		t.Errorf("expected parent directory permissions 0755, got %04o", perm)
	}
}

// TestInitConfigAtPath_FilePermissions verifies that the created file has
// correct permissions (0644).
func TestInitConfigAtPath_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	err := initConfigAtPath(configPath)
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

// TestConfigTemplate_ValidYAML verifies that the configTemplate constant
// is valid YAML that can be parsed.
func TestConfigTemplate_ValidYAML(t *testing.T) {
	// Parse template as YAML
	var userCfg schemas.UserConfig
	err := yaml.Unmarshal([]byte(configTemplate), &userCfg)
	if err != nil {
		t.Fatalf("template is not valid YAML: %v", err)
	}

	// Verify basic structure
	if userCfg.Agents == nil {
		t.Fatal("expected agents section in config")
	}
}

// TestConfigTemplate_PassesValidation verifies that the config template
// passes full validation when loaded.
func TestConfigTemplate_PassesValidation(t *testing.T) {
	// Parse template as YAML
	var userCfg schemas.UserConfig
	err := yaml.Unmarshal([]byte(configTemplate), &userCfg)
	if err != nil {
		t.Fatalf("template is not valid YAML: %v", err)
	}

	// Validate using config.ValidateUserConfig
	err = config.ValidateUserConfig(&userCfg)
	if err != nil {
		t.Fatalf("template failed validation: %v", err)
	}

	// Verify executors section
	if userCfg.Agents.Executors == nil {
		t.Fatal("expected executors in template")
	}
	claudeExec, ok := userCfg.Agents.Executors["claude-code"]
	if !ok {
		t.Fatal("expected 'claude-code' executor in template")
	}
	if claudeExec.Type != "claude" {
		t.Errorf("expected type 'claude', got %q", claudeExec.Type)
	}

	// Verify bindings section
	if userCfg.Agents.Bindings == nil {
		t.Fatal("expected bindings in template")
	}
	if userCfg.Agents.Bindings.Orchestrator == nil || *userCfg.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator binding 'claude-code'")
	}
	if userCfg.Agents.Bindings.Implementer == nil || *userCfg.Agents.Bindings.Implementer != "claude-code" {
		t.Error("expected implementer binding 'claude-code'")
	}
}

// TestConfigTemplate_ContainsDocumentation verifies that the template
// includes helpful documentation comments.
func TestConfigTemplate_ContainsDocumentation(t *testing.T) {
	// Verify template has key documentation elements
	expectedPhrases := []string{
		"# Sow Agent Configuration",
		"This file configures which AI CLI tools",
		"Configuration priority:",
		"Environment variables",
		"# Executor definitions",
		"# Bindings:",
		"yolo_mode",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(configTemplate, phrase) {
			t.Errorf("expected template to contain %q", phrase)
		}
	}
}

// TestNewInitCmd_Structure verifies the init command has correct structure.
func TestNewInitCmd_Structure(t *testing.T) {
	cmd := newInitCmd()

	if cmd.Use != "init" {
		t.Errorf("expected Use='init', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewInitCmd_LongDescription verifies the long description contains
// helpful information.
func TestNewInitCmd_LongDescription(t *testing.T) {
	cmd := newInitCmd()

	expectedPhrases := []string{
		"configuration file",
		"template",
		"sow config edit",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}

// TestRunInit_OutputsPath tests that runInit prints the path on success.
func TestRunInit_OutputsPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create a command with custom output buffer
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Call runInitWithPath directly (we'll create this helper)
	err := runInitWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !containsSubstring(output, "Created configuration") {
		t.Errorf("expected output to contain 'Created configuration', got: %s", output)
	}
	if !containsSubstring(output, configPath) {
		t.Errorf("expected output to contain path %q, got: %s", configPath, output)
	}
}
