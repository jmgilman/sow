package sow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
)

// TestGetUserConfigPath tests that GetUserConfigPath returns the correct path
// using os.UserConfigDir() for cross-platform compatibility.
func TestGetUserConfigPath(t *testing.T) {
	path, err := GetUserConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify path ends with expected components
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected path to end with config.yaml, got %s", path)
	}

	dir := filepath.Dir(path)
	if filepath.Base(dir) != "sow" {
		t.Errorf("expected parent directory to be 'sow', got %s", filepath.Base(dir))
	}

	// Verify it uses the system config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("failed to get user config dir: %v", err)
	}
	expectedPath := filepath.Join(configDir, "sow", "config.yaml")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}
}

// TestLoadUserConfig_MissingFile tests that LoadUserConfig returns defaults
// without error when the config file doesn't exist.
func TestLoadUserConfig_MissingFile(t *testing.T) {
	// Create a temp directory as a mock config dir
	tempDir := t.TempDir()

	// Use a loader with custom path for testing
	config, err := loadUserConfigFromPath(filepath.Join(tempDir, "sow", "config.yaml"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	// Verify defaults are returned
	if config == nil {
		t.Fatal("expected config, got nil")
	}

	if config.Agents == nil {
		t.Fatal("expected agents config, got nil")
	}

	// Verify default executor exists
	if config.Agents.Executors == nil {
		t.Fatal("expected executors, got nil")
	}
	executor, ok := config.Agents.Executors["claude-code"]
	if !ok {
		t.Error("expected 'claude-code' executor in defaults")
	}
	if executor.Type != "claude" {
		t.Errorf("expected executor type 'claude', got %q", executor.Type)
	}

	// Verify default bindings
	if config.Agents.Bindings == nil {
		t.Fatal("expected bindings, got nil")
	}
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator binding to be 'claude-code'")
	}
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "claude-code" {
		t.Error("expected implementer binding to be 'claude-code'")
	}
}

// TestLoadUserConfig_ValidYAML tests that LoadUserConfig correctly parses
// a valid YAML configuration file.
func TestLoadUserConfig_ValidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write valid YAML config
	configYAML := `agents:
  executors:
    cursor:
      type: cursor
      settings:
        yolo_mode: true
  bindings:
    implementer: cursor
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := loadUserConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cursor executor was parsed
	if config.Agents == nil || config.Agents.Executors == nil {
		t.Fatal("expected agents.executors to be populated")
	}

	cursorExec, ok := config.Agents.Executors["cursor"]
	if !ok {
		t.Fatal("expected 'cursor' executor")
	}
	if cursorExec.Type != "cursor" {
		t.Errorf("expected cursor type, got %q", cursorExec.Type)
	}
	if cursorExec.Settings == nil || cursorExec.Settings.Yolo_mode == nil || !*cursorExec.Settings.Yolo_mode {
		t.Error("expected yolo_mode to be true")
	}

	// Verify implementer binding
	if config.Agents.Bindings == nil || config.Agents.Bindings.Implementer == nil {
		t.Fatal("expected implementer binding")
	}
	if *config.Agents.Bindings.Implementer != "cursor" {
		t.Errorf("expected implementer binding 'cursor', got %q", *config.Agents.Bindings.Implementer)
	}

	// Verify defaults were applied for missing bindings
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator to default to 'claude-code'")
	}
}

// TestLoadUserConfig_InvalidYAML tests that LoadUserConfig returns an error
// for invalid YAML syntax.
func TestLoadUserConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
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

	config, err := loadUserConfigFromPath(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if config != nil {
		t.Error("expected nil config for invalid YAML")
	}
}

// TestApplyUserConfigDefaults_NilAgents tests that applyUserConfigDefaults
// sets all defaults when agents is nil.
func TestApplyUserConfigDefaults_NilAgents(t *testing.T) {
	config := &schemas.UserConfig{}

	applyUserConfigDefaults(config)

	if config.Agents == nil {
		t.Fatal("expected agents to be set")
	}
	if config.Agents.Executors == nil {
		t.Fatal("expected executors to be set")
	}
	if config.Agents.Bindings == nil {
		t.Fatal("expected bindings to be set")
	}

	// Check all bindings default to claude-code
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator binding 'claude-code'")
	}
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "claude-code" {
		t.Error("expected implementer binding 'claude-code'")
	}
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "claude-code" {
		t.Error("expected architect binding 'claude-code'")
	}
	if config.Agents.Bindings.Reviewer == nil || *config.Agents.Bindings.Reviewer != "claude-code" {
		t.Error("expected reviewer binding 'claude-code'")
	}
	if config.Agents.Bindings.Planner == nil || *config.Agents.Bindings.Planner != "claude-code" {
		t.Error("expected planner binding 'claude-code'")
	}
	if config.Agents.Bindings.Researcher == nil || *config.Agents.Bindings.Researcher != "claude-code" {
		t.Error("expected researcher binding 'claude-code'")
	}
	if config.Agents.Bindings.Decomposer == nil || *config.Agents.Bindings.Decomposer != "claude-code" {
		t.Error("expected decomposer binding 'claude-code'")
	}
}

// TestApplyUserConfigDefaults_PartialBindings tests that applyUserConfigDefaults
// fills in missing bindings while preserving existing ones.
func TestApplyUserConfigDefaults_PartialBindings(t *testing.T) {
	cursor := "cursor"
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
			Bindings: &struct {
				Orchestrator *string `json:"orchestrator,omitempty"`
				Implementer  *string `json:"implementer,omitempty"`
				Architect    *string `json:"architect,omitempty"`
				Reviewer     *string `json:"reviewer,omitempty"`
				Planner      *string `json:"planner,omitempty"`
				Researcher   *string `json:"researcher,omitempty"`
				Decomposer   *string `json:"decomposer,omitempty"`
			}{
				Implementer: &cursor,
			},
		},
	}

	applyUserConfigDefaults(config)

	// Verify implementer preserved
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "cursor" {
		t.Error("expected implementer to remain 'cursor'")
	}

	// Verify others defaulted
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator to default to 'claude-code'")
	}
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "claude-code" {
		t.Error("expected architect to default to 'claude-code'")
	}
}

// TestApplyUserConfigDefaults_ExistingExecutors tests that applyUserConfigDefaults
// preserves user-defined executors while adding the default claude-code executor.
func TestApplyUserConfigDefaults_ExistingExecutors(t *testing.T) {
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
				"cursor": {
					Type: "cursor",
				},
			},
		},
	}

	applyUserConfigDefaults(config)

	// Verify cursor preserved
	cursorExec, ok := config.Agents.Executors["cursor"]
	if !ok {
		t.Error("expected cursor executor to be preserved")
	}
	if cursorExec.Type != "cursor" {
		t.Errorf("expected cursor type, got %q", cursorExec.Type)
	}

	// Verify claude-code added
	claudeExec, ok := config.Agents.Executors["claude-code"]
	if !ok {
		t.Error("expected claude-code executor to be added")
	}
	if claudeExec.Type != "claude" {
		t.Errorf("expected claude type, got %q", claudeExec.Type)
	}
}

// TestGetDefaultUserConfig tests that getDefaultUserConfig returns a config
// with all bindings pointing to "claude-code".
func TestGetDefaultUserConfig(t *testing.T) {
	config := getDefaultUserConfig()

	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.Agents == nil {
		t.Fatal("expected agents, got nil")
	}

	// Verify claude-code executor
	if config.Agents.Executors == nil {
		t.Fatal("expected executors, got nil")
	}
	executor, ok := config.Agents.Executors["claude-code"]
	if !ok {
		t.Fatal("expected 'claude-code' executor")
	}
	if executor.Type != "claude" {
		t.Errorf("expected type 'claude', got %q", executor.Type)
	}
	if executor.Settings == nil || executor.Settings.Yolo_mode == nil || *executor.Settings.Yolo_mode {
		// yolo_mode should be false by default
		t.Error("expected yolo_mode to be false")
	}

	// Verify all bindings
	bindings := config.Agents.Bindings
	if bindings == nil {
		t.Fatal("expected bindings, got nil")
	}

	checkBinding := func(name string, value *string) {
		t.Helper()
		if value == nil {
			t.Errorf("expected %s binding, got nil", name)
			return
		}
		if *value != "claude-code" {
			t.Errorf("expected %s binding 'claude-code', got %q", name, *value)
		}
	}

	checkBinding("orchestrator", bindings.Orchestrator)
	checkBinding("implementer", bindings.Implementer)
	checkBinding("architect", bindings.Architect)
	checkBinding("reviewer", bindings.Reviewer)
	checkBinding("planner", bindings.Planner)
	checkBinding("researcher", bindings.Researcher)
	checkBinding("decomposer", bindings.Decomposer)
}

// TestLoadUserConfig_PermissionDenied tests that LoadUserConfig returns
// an error when the config file exists but cannot be read.
func TestLoadUserConfig_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with no read permissions
	if err := os.WriteFile(configPath, []byte("agents: {}"), 0000); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	defer os.Chmod(configPath, 0644) // Restore permissions for cleanup

	_, err := loadUserConfigFromPath(configPath)
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}
}

// TestLoadUserConfig_EmptyFile tests that an empty config file returns defaults.
func TestLoadUserConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write empty config file
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := loadUserConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have defaults applied
	if config.Agents == nil {
		t.Fatal("expected agents with defaults")
	}
	if config.Agents.Bindings == nil || config.Agents.Bindings.Orchestrator == nil {
		t.Fatal("expected default bindings")
	}
	if *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Errorf("expected default binding 'claude-code', got %q", *config.Agents.Bindings.Orchestrator)
	}
}
