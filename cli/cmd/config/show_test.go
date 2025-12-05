package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TestNewShowCmd_Structure verifies the show command has correct structure.
func TestNewShowCmd_Structure(t *testing.T) {
	cmd := newShowCmd()

	if cmd.Use != "show" {
		t.Errorf("expected Use='show', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewShowCmd_LongDescription verifies the long description contains key info.
func TestNewShowCmd_LongDescription(t *testing.T) {
	cmd := newShowCmd()

	expectedPhrases := []string{
		"effective",
		"configuration",
		"defaults",
		"Environment", // capitalized as in the actual Long description
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain '%s'", phrase)
		}
	}
}

// TestRunShow_NoConfigFile verifies that show works when no config file exists.
// Should display defaults with "(not found, using defaults)" in header.
func TestRunShow_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain header indicating file not found
	if !strings.Contains(output, "not found") {
		t.Errorf("expected output to contain 'not found', got: %s", output)
	}
	if !strings.Contains(output, "using defaults") {
		t.Errorf("expected output to contain 'using defaults', got: %s", output)
	}

	// Should still contain YAML config output
	if !strings.Contains(output, "agents:") {
		t.Errorf("expected output to contain 'agents:', got: %s", output)
	}
}

// TestRunShow_WithConfigFile verifies that show works when config file exists.
// Should display "(exists)" in header.
func TestRunShow_WithConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

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

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain header indicating file exists
	if !strings.Contains(output, "exists") {
		t.Errorf("expected output to contain 'exists', got: %s", output)
	}

	// Should not contain "not found"
	if strings.Contains(output, "not found") {
		t.Errorf("expected output NOT to contain 'not found', got: %s", output)
	}
}

// TestRunShow_EnvOverrides verifies that environment overrides are shown in header.
func TestRunShow_EnvOverrides(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain header indicating env overrides
	if !strings.Contains(output, "Environment overrides") {
		t.Errorf("expected output to contain 'Environment overrides', got: %s", output)
	}
	if !strings.Contains(output, "SOW_AGENTS_IMPLEMENTER") {
		t.Errorf("expected output to contain 'SOW_AGENTS_IMPLEMENTER', got: %s", output)
	}
}

// TestRunShow_MultipleEnvOverrides verifies that multiple env overrides are shown.
func TestRunShow_MultipleEnvOverrides(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
	t.Setenv("SOW_AGENTS_ARCHITECT", "windsurf")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain both env vars
	if !strings.Contains(output, "SOW_AGENTS_IMPLEMENTER") {
		t.Errorf("expected output to contain 'SOW_AGENTS_IMPLEMENTER', got: %s", output)
	}
	if !strings.Contains(output, "SOW_AGENTS_ARCHITECT") {
		t.Errorf("expected output to contain 'SOW_AGENTS_ARCHITECT', got: %s", output)
	}
}

// TestGetEnvOverrides_ReturnsSetVars verifies getEnvOverrides returns set variables.
func TestGetEnvOverrides_ReturnsSetVars(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")
	t.Setenv("SOW_AGENTS_ARCHITECT", "windsurf")

	overrides := getEnvOverrides()

	if len(overrides) != 2 {
		t.Errorf("expected 2 overrides, got %d", len(overrides))
	}

	// Check both are present
	hasImplementer := false
	hasArchitect := false
	for _, o := range overrides {
		if o == "SOW_AGENTS_IMPLEMENTER" {
			hasImplementer = true
		}
		if o == "SOW_AGENTS_ARCHITECT" {
			hasArchitect = true
		}
	}
	if !hasImplementer {
		t.Error("expected SOW_AGENTS_IMPLEMENTER in overrides")
	}
	if !hasArchitect {
		t.Error("expected SOW_AGENTS_ARCHITECT in overrides")
	}
}

// TestGetEnvOverrides_IgnoresEmpty verifies getEnvOverrides ignores empty values.
func TestGetEnvOverrides_IgnoresEmpty(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "")

	overrides := getEnvOverrides()

	// Empty strings should not be included
	for _, o := range overrides {
		if o == "SOW_AGENTS_IMPLEMENTER" {
			t.Error("empty env var should not be in overrides")
		}
	}
}

// TestGetEnvOverrides_NoVarsSet verifies getEnvOverrides returns empty slice when nothing set.
func TestGetEnvOverrides_NoVarsSet(t *testing.T) {
	// No env vars set (test isolation from t.Setenv in other tests)
	overrides := getEnvOverrides()

	// Should return empty (or nil which will be coerced to empty)
	if len(overrides) != 0 {
		t.Errorf("expected 0 overrides, got %d: %v", len(overrides), overrides)
	}
}

// TestGetEnvOverrides_AllVars verifies all supported env vars are detected.
func TestGetEnvOverrides_AllVars(t *testing.T) {
	t.Setenv("SOW_AGENTS_ORCHESTRATOR", "exec-1")
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "exec-2")
	t.Setenv("SOW_AGENTS_ARCHITECT", "exec-3")
	t.Setenv("SOW_AGENTS_REVIEWER", "exec-4")
	t.Setenv("SOW_AGENTS_PLANNER", "exec-5")
	t.Setenv("SOW_AGENTS_RESEARCHER", "exec-6")
	t.Setenv("SOW_AGENTS_DECOMPOSER", "exec-7")

	overrides := getEnvOverrides()

	if len(overrides) != 7 {
		t.Errorf("expected 7 overrides, got %d: %v", len(overrides), overrides)
	}
}

// TestRunShow_OutputIsValidYAML verifies the YAML portion of output is parseable.
func TestRunShow_OutputIsValidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Extract YAML portion (skip comment lines)
	lines := strings.Split(output, "\n")
	var yamlLines []string
	foundYAML := false
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) == "" && !foundYAML {
			continue
		}
		foundYAML = true
		yamlLines = append(yamlLines, line)
	}

	yamlContent := strings.Join(yamlLines, "\n")

	// Parse as YAML
	var config schemas.UserConfig
	err = yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("output YAML is not valid: %v\nYAML content:\n%s", err, yamlContent)
	}

	// Verify basic structure
	if config.Agents == nil {
		t.Error("expected agents in parsed config")
	}
}

// TestRunShow_HeaderFormat verifies the header format is correct.
func TestRunShow_HeaderFormat(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// First line should be the main header comment
	lines := strings.Split(output, "\n")
	if !strings.HasPrefix(lines[0], "# ") {
		t.Errorf("expected first line to start with '# ', got: %s", lines[0])
	}

	// Should contain "Effective configuration"
	if !strings.Contains(lines[0], "Effective configuration") {
		t.Errorf("expected first line to contain 'Effective configuration', got: %s", lines[0])
	}

	// Second line should show config file status
	if !strings.HasPrefix(lines[1], "# Config file:") {
		t.Errorf("expected second line to start with '# Config file:', got: %s", lines[1])
	}
}

// TestRunShow_HeaderLinesAreComments verifies all header lines start with #.
func TestRunShow_HeaderLinesAreComments(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "cursor")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Count header lines (until we hit blank line or non-comment)
	headerCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			break
		}
		if !strings.HasPrefix(line, "#") {
			t.Errorf("header line should start with '#', got: %s", line)
		}
		headerCount++
	}

	// Should have at least 2 header lines (main header + config file status)
	if headerCount < 2 {
		t.Errorf("expected at least 2 header lines, got %d", headerCount)
	}
}

// TestRunShow_NoEnvOverrides_NoEnvLine verifies env line is not shown when no overrides.
func TestRunShow_NoEnvOverrides_NoEnvLine(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should NOT contain "Environment overrides" when no env vars are set
	if strings.Contains(output, "Environment overrides") {
		t.Errorf("expected output NOT to contain 'Environment overrides' when no env vars set")
	}
}

// TestRunShow_ShowsMergedConfig verifies the output shows merged config values.
func TestRunShow_ShowsMergedConfig(t *testing.T) {
	t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-executor")

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create config directory and file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `agents:
  executors:
    cursor:
      type: cursor
  bindings:
    architect: cursor
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Extract YAML portion
	lines := strings.Split(output, "\n")
	var yamlLines []string
	foundYAML := false
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) == "" && !foundYAML {
			continue
		}
		foundYAML = true
		yamlLines = append(yamlLines, line)
	}

	yamlContent := strings.Join(yamlLines, "\n")

	// Parse and verify merged values
	var config schemas.UserConfig
	err = yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("failed to parse output YAML: %v", err)
	}

	// Verify env override took precedence
	if config.Agents.Bindings.Implementer == nil || *config.Agents.Bindings.Implementer != "env-executor" {
		var got string
		if config.Agents.Bindings.Implementer != nil {
			got = *config.Agents.Bindings.Implementer
		}
		t.Errorf("expected implementer 'env-executor' (env), got %q", got)
	}

	// Verify file config preserved for architect
	if config.Agents.Bindings.Architect == nil || *config.Agents.Bindings.Architect != "cursor" {
		var got string
		if config.Agents.Bindings.Architect != nil {
			got = *config.Agents.Bindings.Architect
		}
		t.Errorf("expected architect 'cursor' (file), got %q", got)
	}

	// Verify default for non-specified binding
	if config.Agents.Bindings.Reviewer == nil || *config.Agents.Bindings.Reviewer != "claude-code" {
		var got string
		if config.Agents.Bindings.Reviewer != nil {
			got = *config.Agents.Bindings.Reviewer
		}
		t.Errorf("expected reviewer 'claude-code' (default), got %q", got)
	}
}

// TestRunShow_BlankLineBeforeYAML verifies there's a blank line between header and YAML.
func TestRunShow_BlankLineBeforeYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runShowWithPath(cmd, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Find the blank line between header and YAML
	foundBlank := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "" && i > 0 {
			// Check if previous line was a comment
			if strings.HasPrefix(lines[i-1], "#") {
				foundBlank = true
				// Check if next line is YAML (not a comment)
				if i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "#") {
					break
				}
			}
		}
	}

	if !foundBlank {
		t.Error("expected blank line between header and YAML content")
	}
}
