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
	config, err := LoadUserConfigFromPath(filepath.Join(tempDir, "sow", "config.yaml"))
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

	config, err := LoadUserConfigFromPath(configPath)
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

	config, err := LoadUserConfigFromPath(configPath)
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
	defer func() { _ = os.Chmod(configPath, 0644) }() // Restore permissions for cleanup

	_, err := LoadUserConfigFromPath(configPath)
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

	config, err := LoadUserConfigFromPath(configPath)
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

// =============================================================================
// Environment Override Tests
// =============================================================================

// TestApplyEnvOverrides_SingleVar tests that a single environment variable
// overrides the corresponding binding.
func TestApplyEnvOverrides_SingleVar(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")

	config := getDefaultUserConfig()
	applyEnvOverrides(config)

	if config.Agents.Bindings.Implementer == nil {
		t.Fatal("expected implementer binding, got nil")
	}
	if *config.Agents.Bindings.Implementer != "cursor" {
		t.Errorf("expected implementer 'cursor', got %q", *config.Agents.Bindings.Implementer)
	}

	// Verify other bindings unchanged
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "claude-code" {
		t.Error("expected orchestrator to remain 'claude-code'")
	}
}

// TestApplyEnvOverrides_MultipleVars tests that multiple environment variables
// work together to override multiple bindings.
func TestApplyEnvOverrides_MultipleVars(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
	t.Setenv("SOW_AGENTS_ARCHITECT", "windsurf")
	t.Setenv("SOW_AGENTS_ORCHESTRATOR", "my-claude")

	config := getDefaultUserConfig()
	applyEnvOverrides(config)

	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "cursor" {
		t.Error("expected implementer 'cursor'")
	}
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "windsurf" {
		t.Error("expected architect 'windsurf'")
	}
	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "my-claude" {
		t.Error("expected orchestrator 'my-claude'")
	}

	// Verify others unchanged
	if config.Agents.Bindings.Reviewer == nil || *config.Agents.Bindings.Reviewer != "claude-code" {
		t.Error("expected reviewer to remain 'claude-code'")
	}
}

// TestApplyEnvOverrides_EmptyVar tests that empty environment variables are ignored.
func TestApplyEnvOverrides_EmptyVar(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "")

	config := getDefaultUserConfig()
	applyEnvOverrides(config)

	// Should remain default
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "claude-code" {
		t.Error("expected implementer to remain 'claude-code' when env var is empty")
	}
}

// TestApplyEnvOverrides_NilAgents tests that applyEnvOverrides handles nil config.Agents gracefully.
func TestApplyEnvOverrides_NilAgents(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")

	config := &schemas.UserConfig{}
	applyEnvOverrides(config)

	// Should create the agents and bindings structs
	if config.Agents == nil {
		t.Fatal("expected agents to be initialized")
	}
	if config.Agents.Bindings == nil {
		t.Fatal("expected bindings to be initialized")
	}
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "cursor" {
		t.Error("expected implementer 'cursor'")
	}
}

// TestApplyEnvOverrides_AllVars tests that all supported environment variables work.
func TestApplyEnvOverrides_AllVars(t *testing.T) {
	envVars := map[string]string{
		"SOW_AGENTS_ORCHESTRATOR": "exec-1",
		"SOW_AGENTS_IMPLEMENTER":  "exec-2",
		"SOW_AGENTS_ARCHITECT":    "exec-3",
		"SOW_AGENTS_REVIEWER":     "exec-4",
		"SOW_AGENTS_PLANNER":      "exec-5",
		"SOW_AGENTS_RESEARCHER":   "exec-6",
		"SOW_AGENTS_DECOMPOSER":   "exec-7",
	}
	for k, v := range envVars {
		t.Setenv(k, v)
	}

	config := getDefaultUserConfig()
	applyEnvOverrides(config)

	if config.Agents.Bindings.Orchestrator == nil || *config.Agents.Bindings.Orchestrator != "exec-1" {
		t.Error("expected orchestrator 'exec-1'")
	}
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "exec-2" {
		t.Error("expected implementer 'exec-2'")
	}
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "exec-3" {
		t.Error("expected architect 'exec-3'")
	}
	if config.Agents.Bindings.Reviewer == nil || *config.Agents.Bindings.Reviewer != "exec-4" {
		t.Error("expected reviewer 'exec-4'")
	}
	if config.Agents.Bindings.Planner == nil || *config.Agents.Bindings.Planner != "exec-5" {
		t.Error("expected planner 'exec-5'")
	}
	if config.Agents.Bindings.Researcher == nil || *config.Agents.Bindings.Researcher != "exec-6" {
		t.Error("expected researcher 'exec-6'")
	}
	if config.Agents.Bindings.Decomposer == nil || *config.Agents.Bindings.Decomposer != "exec-7" {
		t.Error("expected decomposer 'exec-7'")
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

// TestValidateUserConfig_ValidConfig tests that a valid config passes validation.
func TestValidateUserConfig_ValidConfig(t *testing.T) {
	config := getDefaultUserConfig()

	err := ValidateUserConfig(config)
	if err != nil {
		t.Fatalf("expected valid config to pass validation, got: %v", err)
	}
}

// TestValidateUserConfig_InvalidExecutorType tests that an invalid executor type is caught.
func TestValidateUserConfig_InvalidExecutorType(t *testing.T) {
	config := getDefaultUserConfig()
	config.Agents.Executors["bad-executor"] = struct {
		Type     string `json:"type"`
		Settings *struct {
			Yolo_mode *bool   `json:"yolo_mode,omitempty"`
			Model     *string `json:"model,omitempty"`
		} `json:"settings,omitempty"`
		Custom_args []string `json:"custom_args,omitempty"`
	}{
		Type: "copilot", // Invalid type
	}

	err := ValidateUserConfig(config)
	if err == nil {
		t.Fatal("expected error for invalid executor type")
	}

	// Verify error message contains useful info
	errStr := err.Error()
	if !contains(errStr, "copilot") {
		t.Errorf("expected error to mention 'copilot', got: %s", errStr)
	}
	if !contains(errStr, "bad-executor") {
		t.Errorf("expected error to mention 'bad-executor', got: %s", errStr)
	}
}

// TestValidateUserConfig_BindingUndefinedExecutor tests that bindings to undefined executors are caught.
func TestValidateUserConfig_BindingUndefinedExecutor(t *testing.T) {
	nonexistent := "nonexistent"
	config := getDefaultUserConfig()
	config.Agents.Bindings.Implementer = &nonexistent

	err := ValidateUserConfig(config)
	if err == nil {
		t.Fatal("expected error for binding to undefined executor")
	}

	// Verify error message contains useful info
	errStr := err.Error()
	if !contains(errStr, "nonexistent") {
		t.Errorf("expected error to mention 'nonexistent', got: %s", errStr)
	}
	if !contains(errStr, "implementer") {
		t.Errorf("expected error to mention 'implementer', got: %s", errStr)
	}
}

// TestValidateUserConfig_BindingDefaultExecutor tests that binding to "claude-code" is always valid,
// even if not explicitly defined in executors.
func TestValidateUserConfig_BindingDefaultExecutor(t *testing.T) {
	claudeCode := "claude-code"
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
			// No executors defined
			Executors: map[string]struct {
				Type     string `json:"type"`
				Settings *struct {
					Yolo_mode *bool   `json:"yolo_mode,omitempty"`
					Model     *string `json:"model,omitempty"`
				} `json:"settings,omitempty"`
				Custom_args []string `json:"custom_args,omitempty"`
			}{},
			Bindings: &struct {
				Orchestrator *string `json:"orchestrator,omitempty"`
				Implementer  *string `json:"implementer,omitempty"`
				Architect    *string `json:"architect,omitempty"`
				Reviewer     *string `json:"reviewer,omitempty"`
				Planner      *string `json:"planner,omitempty"`
				Researcher   *string `json:"researcher,omitempty"`
				Decomposer   *string `json:"decomposer,omitempty"`
			}{
				// Binding to claude-code which is not in executors
				Implementer: &claudeCode,
			},
		},
	}

	err := ValidateUserConfig(config)
	if err != nil {
		t.Fatalf("expected claude-code binding to be valid even if not defined, got: %v", err)
	}
}

// TestValidateUserConfig_EmptyConfig tests that an empty config is valid.
func TestValidateUserConfig_EmptyConfig(t *testing.T) {
	config := &schemas.UserConfig{}

	err := ValidateUserConfig(config)
	if err != nil {
		t.Fatalf("expected empty config to be valid, got: %v", err)
	}
}

// TestValidateUserConfig_AllValidTypes tests all valid executor types.
func TestValidateUserConfig_AllValidTypes(t *testing.T) {
	config := getDefaultUserConfig()
	config.Agents.Executors["claude-exec"] = struct {
		Type     string `json:"type"`
		Settings *struct {
			Yolo_mode *bool   `json:"yolo_mode,omitempty"`
			Model     *string `json:"model,omitempty"`
		} `json:"settings,omitempty"`
		Custom_args []string `json:"custom_args,omitempty"`
	}{Type: "claude"}
	config.Agents.Executors["cursor-exec"] = struct {
		Type     string `json:"type"`
		Settings *struct {
			Yolo_mode *bool   `json:"yolo_mode,omitempty"`
			Model     *string `json:"model,omitempty"`
		} `json:"settings,omitempty"`
		Custom_args []string `json:"custom_args,omitempty"`
	}{Type: "cursor"}
	config.Agents.Executors["windsurf-exec"] = struct {
		Type     string `json:"type"`
		Settings *struct {
			Yolo_mode *bool   `json:"yolo_mode,omitempty"`
			Model     *string `json:"model,omitempty"`
		} `json:"settings,omitempty"`
		Custom_args []string `json:"custom_args,omitempty"`
	}{Type: "windsurf"}

	err := ValidateUserConfig(config)
	if err != nil {
		t.Fatalf("expected all valid types to pass, got: %v", err)
	}
}

// =============================================================================
// Integration Tests - LoadUserConfig with full pipeline
// =============================================================================

// TestLoadUserConfig_EnvOverridesFile tests that environment variables take
// precedence over file configuration.
func TestLoadUserConfig_EnvOverridesFile(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-executor")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with implementer set to cursor
	configYAML := `agents:
  executors:
    cursor:
      type: cursor
  bindings:
    implementer: cursor
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := LoadUserConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env var should override file config
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "env-executor" {
		t.Errorf("expected env override 'env-executor', got %q", *config.Agents.Bindings.Implementer)
	}
}

// TestLoadUserConfig_EnvOverridesNoFile tests that environment variables work
// even when config file doesn't exist.
func TestLoadUserConfig_EnvOverridesNoFile(t *testing.T) {
	t.Setenv("SOW_AGENTS_ARCHITECT", "env-architect")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	config, err := LoadUserConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env var should be applied on top of defaults
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "env-architect" {
		t.Errorf("expected env override 'env-architect', got %q", *config.Agents.Bindings.Architect)
	}

	// Other bindings should be default
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "claude-code" {
		t.Error("expected implementer to be default 'claude-code'")
	}
}

// TestLoadUserConfig_InvalidConfig tests that LoadUserConfig returns validation errors.
func TestLoadUserConfig_InvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with invalid executor type
	configYAML := `agents:
  executors:
    bad:
      type: invalid-type
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadUserConfigFromPath(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid executor type")
	}

	// Error should mention the path and the issue
	errStr := err.Error()
	if !contains(errStr, "invalid-type") {
		t.Errorf("expected error to mention 'invalid-type', got: %s", errStr)
	}
}

// TestLoadUserConfig_InvalidBinding tests that invalid bindings are caught.
func TestLoadUserConfig_InvalidBinding(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with binding to undefined executor
	configYAML := `agents:
  bindings:
    implementer: nonexistent
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := LoadUserConfigFromPath(configPath)
	if err == nil {
		t.Fatal("expected validation error for binding to undefined executor")
	}

	errStr := err.Error()
	if !contains(errStr, "nonexistent") {
		t.Errorf("expected error to mention 'nonexistent', got: %s", errStr)
	}
}

// TestLoadUserConfig_PriorityOrder tests the full priority order:
// env vars > file config > defaults.
func TestLoadUserConfig_PriorityOrder(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-value")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write config file with some values
	configYAML := `agents:
  executors:
    cursor:
      type: cursor
  bindings:
    implementer: cursor
    architect: cursor
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := LoadUserConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Implementer: env var > file
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "env-value" {
		t.Errorf("expected implementer 'env-value' (env), got %q", *config.Agents.Bindings.Implementer)
	}

	// Architect: file > default
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "cursor" {
		t.Errorf("expected architect 'cursor' (file), got %q", *config.Agents.Bindings.Architect)
	}

	// Reviewer: default (not in file, not in env)
	if config.Agents.Bindings.Reviewer == nil || *config.Agents.Bindings.Reviewer != "claude-code" {
		t.Errorf("expected reviewer 'claude-code' (default), got %q", *config.Agents.Bindings.Reviewer)
	}
}

// contains is a helper function for checking if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
