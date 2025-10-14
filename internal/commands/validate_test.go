package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCommand_SingleFile_AutoDetect(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)

	// Write valid project state
	stateContent := `project:
  name: "test-project"
  branch: "test-branch"
  created_at: "2025-10-13T00:00:00Z"
  updated_at: "2025-10-13T00:00:00Z"
  description: "Test project"
  complexity:
    rating: 1
    metrics:
      estimated_files: 5
      cross_cutting: false
      new_dependencies: false
  active_phase: "implement"

phases:
  - name: "implement"
    status: in_progress
    created_at: "2025-10-13T00:00:00Z"
    completed_at: null
    tasks: []
`
	statePath := filepath.Join(projectDir, "state.yaml")
	os.WriteFile(statePath, []byte(stateContent), 0644)

	// Run validate command
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{statePath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") || !strings.Contains(output, statePath) {
		t.Errorf("expected success indicator for valid file, got: %s", output)
	}
}

func TestValidateCommand_SingleFile_Invalid(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)

	// Write invalid project state (missing required fields)
	stateContent := `project:
  name: "test-project"
  # missing required fields
`
	statePath := filepath.Join(projectDir, "state.yaml")
	os.WriteFile(statePath, []byte(stateContent), 0644)

	// Run validate command
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{statePath})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid file, got nil")
	}

	output := buf.String()
	if !strings.Contains(output, "✗") || !strings.Contains(output, statePath) {
		t.Errorf("expected error indicator for invalid file, got: %s", output)
	}
}

func TestValidateCommand_ExplicitType(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, "tasks", "040")

	os.MkdirAll(taskDir, 0755)

	// Write valid task state
	stateContent := `task:
  id: "040"
  name: "Test task"
  phase: "implement"
  status: "in_progress"
  created_at: "2025-10-13T00:00:00Z"
  started_at: "2025-10-13T00:00:00Z"
  updated_at: "2025-10-13T00:00:00Z"
  completed_at: null
  iteration: 1
  assigned_agent: "implementer"
  references: []
  feedback: []
  files_modified: []
`
	statePath := filepath.Join(taskDir, "state.yaml")
	os.WriteFile(statePath, []byte(stateContent), 0644)

	// Run validate command with explicit type
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--type", "task-state", statePath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("expected success indicator, got: %s", output)
	}
}

func TestValidateCommand_GlobPattern(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	phasesDir := filepath.Join(sowDir, "project", "phases", "implement", "tasks")

	// Create multiple task directories
	for _, taskID := range []string{"010", "020", "030"} {
		taskDir := filepath.Join(phasesDir, taskID)
		os.MkdirAll(taskDir, 0755)

		stateContent := `task:
  id: "` + taskID + `"
  name: "Test task"
  phase: "implement"
  status: "in_progress"
  created_at: "2025-10-13T00:00:00Z"
  started_at: "2025-10-13T00:00:00Z"
  updated_at: "2025-10-13T00:00:00Z"
  completed_at: null
  iteration: 1
  assigned_agent: "implementer"
  references: []
  feedback: []
  files_modified: []
`
		os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)
	}

	// Run validate command with glob pattern
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	pattern := filepath.Join(phasesDir, "*/state.yaml")
	cmd.SetArgs([]string{"--type", "task-state", pattern})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}

	output := buf.String()

	// Should validate all three files
	for _, taskID := range []string{"010", "020", "030"} {
		taskPath := filepath.Join(phasesDir, taskID, "state.yaml")
		if !strings.Contains(output, taskPath) {
			t.Errorf("expected output to include %s", taskPath)
		}
	}
}

func TestValidateCommand_MixedResults(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	phasesDir := filepath.Join(sowDir, "project", "phases", "implement", "tasks")

	// Create one valid and one invalid task
	validTaskDir := filepath.Join(phasesDir, "010")
	os.MkdirAll(validTaskDir, 0755)

	validContent := `task:
  id: "010"
  name: "Valid task"
  phase: "implement"
  status: "in_progress"
  created_at: "2025-10-13T00:00:00Z"
  started_at: "2025-10-13T00:00:00Z"
  updated_at: "2025-10-13T00:00:00Z"
  completed_at: null
  iteration: 1
  assigned_agent: "implementer"
  references: []
  feedback: []
  files_modified: []
`
	os.WriteFile(filepath.Join(validTaskDir, "state.yaml"), []byte(validContent), 0644)

	invalidTaskDir := filepath.Join(phasesDir, "020")
	os.MkdirAll(invalidTaskDir, 0755)

	invalidContent := `task:
  id: "020"
  # missing required fields
`
	os.WriteFile(filepath.Join(invalidTaskDir, "state.yaml"), []byte(invalidContent), 0644)

	// Run validate command with glob pattern
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	pattern := filepath.Join(phasesDir, "*/state.yaml")
	cmd.SetArgs([]string{"--type", "task-state", pattern})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for mixed results, got nil")
	}

	output := buf.String()

	// Should show both success and failure
	if !strings.Contains(output, "✓") {
		t.Error("expected success indicator for valid file")
	}
	if !strings.Contains(output, "✗") {
		t.Error("expected error indicator for invalid file")
	}
}

func TestValidateCommand_TypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "project state",
			path:     ".sow/project/state.yaml",
			expected: "project-state",
		},
		{
			name:     "task state",
			path:     ".sow/project/phases/implement/tasks/040/state.yaml",
			expected: "task-state",
		},
		{
			name:     "sink index",
			path:     ".sow/sinks/index.json",
			expected: "sink-index",
		},
		{
			name:     "repo index",
			path:     ".sow/repos/index.json",
			expected: "repo-index",
		},
		{
			name:     "version file",
			path:     ".sow/.version",
			expected: "sow-version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := detectFileType(tt.path)
			if detected != tt.expected {
				t.Errorf("expected type %s for path %s, got %s", tt.expected, tt.path, detected)
			}
		})
	}
}

func TestValidateCommand_NoFilesFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Run validate with pattern that matches nothing
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	pattern := filepath.Join(tmpDir, "nonexistent", "*.yaml")
	cmd.SetArgs([]string{pattern})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no files found, got nil")
	}
}

func TestValidateCommand_MissingType(t *testing.T) {
	// Create a file that can't be auto-detected
	tmpDir := t.TempDir()
	unknownFile := filepath.Join(tmpDir, "unknown.yaml")
	os.WriteFile(unknownFile, []byte("data: value"), 0644)

	// Run validate without type
	cmd := NewValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{unknownFile})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown file type, got nil")
	}
}
