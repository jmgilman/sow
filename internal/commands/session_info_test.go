package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionInfoCommand_NoSowDirectory(t *testing.T) {
	// Create temporary directory without .sow
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Run session-info command
	cmd := NewSessionInfoCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("session-info command failed: %v", err)
	}

	// Parse JSON output
	var output map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if output["context"] != "none" {
		t.Errorf("expected context 'none', got %v", output["context"])
	}

	if output["cli_version"] == nil {
		t.Error("expected cli_version in output")
	}
}

func TestSessionInfoCommand_ProjectContext(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)

	// Write project state
	stateContent := `current_phase: "implement"
active_task: "040"
`
	os.WriteFile(filepath.Join(projectDir, "state.yaml"), []byte(stateContent), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(projectDir)

	// Run session-info command
	cmd := NewSessionInfoCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("session-info command failed: %v", err)
	}

	// Parse JSON output
	var output map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if output["context"] != "project" {
		t.Errorf("expected context 'project', got %v", output["context"])
	}

	if output["task_id"] != nil {
		t.Errorf("project context should not have task_id, got %v", output["task_id"])
	}

	if output["phase"] != nil {
		t.Errorf("project context should not have phase, got %v", output["phase"])
	}
}

func TestSessionInfoCommand_TaskContext(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	taskDir := filepath.Join(sowDir, "project", "phases", "implement", "tasks", "040")

	os.MkdirAll(taskDir, 0755)

	// Write task state
	stateContent := `task:
  id: "040"
  phase: "implement"
  iteration: 1
  assigned_agent: "implementer"
`
	os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Run session-info command
	cmd := NewSessionInfoCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("session-info command failed: %v", err)
	}

	// Parse JSON output
	var output map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if output["context"] != "task" {
		t.Errorf("expected context 'task', got %v", output["context"])
	}

	if output["task_id"] != "040" {
		t.Errorf("expected task_id '040', got %v", output["task_id"])
	}

	if output["phase"] != "implement" {
		t.Errorf("expected phase 'implement', got %v", output["phase"])
	}

	if output["cli_version"] == nil {
		t.Error("expected cli_version in output")
	}
}

func TestSessionInfoCommand_Performance(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	taskDir := filepath.Join(sowDir, "project", "phases", "implement", "tasks", "040")

	os.MkdirAll(taskDir, 0755)

	// Write task state
	stateContent := `task:
  id: "040"
  phase: "implement"
  iteration: 1
  assigned_agent: "implementer"
`
	os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Measure execution time
	start := time.Now()

	cmd := NewSessionInfoCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("session-info command failed: %v", err)
	}

	elapsed := time.Since(start)

	// Should complete in under 100ms (requirement)
	if elapsed > 100*time.Millisecond {
		t.Errorf("session-info command took %v, expected < 100ms", elapsed)
	}
}

func TestSessionInfoCommand_JSONFormat(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	taskDir := filepath.Join(sowDir, "project", "phases", "design", "tasks", "010")

	os.MkdirAll(taskDir, 0755)

	// Write task state
	stateContent := `task:
  id: "010"
  phase: "design"
  iteration: 3
  assigned_agent: "architect"
`
	os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Run session-info command
	cmd := NewSessionInfoCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("session-info command failed: %v", err)
	}

	// Verify it's valid JSON
	var output map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify required fields
	requiredFields := []string{"context", "cli_version"}
	for _, field := range requiredFields {
		if output[field] == nil {
			t.Errorf("missing required field: %s", field)
		}
	}

	// Task context should have task_id and phase
	if output["context"] == "task" {
		if output["task_id"] == nil {
			t.Error("task context missing task_id")
		}
		if output["phase"] == nil {
			t.Error("task context missing phase")
		}
	}
}
