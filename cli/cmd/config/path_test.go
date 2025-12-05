package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewPathCmd_Structure verifies the path command has correct structure.
func TestNewPathCmd_Structure(t *testing.T) {
	cmd := newPathCmd()

	if cmd.Use != "path" {
		t.Errorf("expected Use='path', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewPathCmd_HasExistsFlag verifies the --exists flag is registered.
func TestNewPathCmd_HasExistsFlag(t *testing.T) {
	cmd := newPathCmd()
	flag := cmd.Flags().Lookup("exists")
	if flag == nil {
		t.Error("expected --exists flag")
	}

	// Verify it's a bool flag
	if flag.Value.Type() != "bool" {
		t.Errorf("expected --exists to be bool, got %s", flag.Value.Type())
	}

	// Verify default is false
	if flag.DefValue != "false" {
		t.Errorf("expected --exists default to be 'false', got '%s'", flag.DefValue)
	}
}

// TestNewPathCmd_LongDescription verifies the long description contains key info.
func TestNewPathCmd_LongDescription(t *testing.T) {
	cmd := newPathCmd()

	expectedPhrases := []string{
		"path",
		"configuration file",
		"--exists",
	}

	for _, phrase := range expectedPhrases {
		if !containsSubstring(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain '%s'", phrase)
		}
	}
}

// TestRunPath_ShowsPath verifies that the command outputs the config path.
func TestRunPath_ShowsPath(t *testing.T) {
	cmd := newPathCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Verify output contains expected path components
	if !strings.Contains(output, "sow") || !strings.Contains(output, "config.yaml") {
		t.Errorf("unexpected output: %s", output)
	}

	// Verify output is an absolute path (starts with / on Unix)
	trimmedOutput := strings.TrimSpace(output)
	if !filepath.IsAbs(trimmedOutput) {
		t.Errorf("expected absolute path, got: %s", trimmedOutput)
	}
}

// TestRunPath_ExistsFlag_FileExists verifies --exists outputs "true" when file exists.
func TestRunPath_ExistsFlag_FileExists(t *testing.T) {
	// Create a temp directory with a config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the config file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test: content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Test with the helper that accepts a custom path
	cmd := newPathCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPathWithOptions(cmd, configPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "true" {
		t.Errorf("expected 'true', got '%s'", output)
	}
}

// TestRunPath_ExistsFlag_FileNotExists verifies --exists outputs "false" when file doesn't exist.
func TestRunPath_ExistsFlag_FileNotExists(t *testing.T) {
	// Create a path that doesn't exist
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent", "config.yaml")

	cmd := newPathCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPathWithOptions(cmd, configPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "false" {
		t.Errorf("expected 'false', got '%s'", output)
	}
}

// TestRunPath_ExistsFlag_NoError verifies --exists never returns an error.
func TestRunPath_ExistsFlag_NoError(t *testing.T) {
	testCases := []struct {
		name       string
		setupFile  bool
		wantOutput string
	}{
		{
			name:       "file exists",
			setupFile:  true,
			wantOutput: "true",
		},
		{
			name:       "file does not exist",
			setupFile:  false,
			wantOutput: "false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "sow", "config.yaml")

			if tc.setupFile {
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			cmd := newPathCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Should never return an error
			err := runPathWithOptions(cmd, configPath, true)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			output := strings.TrimSpace(buf.String())
			if output != tc.wantOutput {
				t.Errorf("expected '%s', got '%s'", tc.wantOutput, output)
			}
		})
	}
}

// TestRunPath_ShowsPath_WithCustomPath verifies runPathWithOptions shows the path correctly.
func TestRunPath_ShowsPath_WithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom", "config.yaml")

	cmd := newPathCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// checkExists = false means just print the path
	err := runPathWithOptions(cmd, configPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != configPath {
		t.Errorf("expected path '%s', got '%s'", configPath, output)
	}
}

// TestRunPath_OutputCleanForScripting verifies output is clean (no extra formatting).
func TestRunPath_OutputCleanForScripting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "sow", "config.yaml")

	// Create the file
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	cmd := newPathCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPathWithOptions(cmd, configPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should have exactly one line with newline
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 line, got %d", len(lines))
	}

	// Should not contain any special formatting
	if strings.Contains(output, "\t") {
		t.Error("output should not contain tabs")
	}

	// Value should be exactly "true"
	if strings.TrimSpace(output) != "true" {
		t.Errorf("expected 'true', got '%s'", strings.TrimSpace(output))
	}
}
