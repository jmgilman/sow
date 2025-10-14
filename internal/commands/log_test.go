package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogCommand_AutoDetectTask(t *testing.T) {
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

	// Create initial log file
	logPath := filepath.Join(taskDir, "log.md")
	initialLog := "# Task Log\n\n**Worker Actions**\n\n---\n"
	os.WriteFile(logPath, []byte(initialLog), 0644)

	// Change to task directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Run log command
	cmd := NewLogCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--action", "created_file",
		"--result", "success",
		"--file", "src/main.go",
		"Created main file",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}

	// Read log file and verify entry was added
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	logStr := string(logContent)

	// Check for required fields
	if !strings.Contains(logStr, "## ") {
		t.Error("log entry missing timestamp header")
	}
	if !strings.Contains(logStr, "implementer-1") {
		t.Error("log entry missing agent ID")
	}
	if !strings.Contains(logStr, "**Action**: created_file") {
		t.Error("log entry missing action")
	}
	if !strings.Contains(logStr, "**Result**: success") {
		t.Error("log entry missing result")
	}
	if !strings.Contains(logStr, "**Files**:") {
		t.Error("log entry missing files section")
	}
	if !strings.Contains(logStr, "- src/main.go") {
		t.Error("log entry missing file")
	}
	if !strings.Contains(logStr, "**Details**: Created main file") {
		t.Error("log entry missing details")
	}
}

func TestLogCommand_ExplicitProject(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)

	// Create initial log file
	logPath := filepath.Join(projectDir, "log.md")
	initialLog := "# Project Log\n\n---\n"
	os.WriteFile(logPath, []byte(initialLog), 0644)

	// Change to some directory within project
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(projectDir)

	// Run log command with --project flag
	cmd := NewLogCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--project",
		"--action", "started_phase",
		"--result", "success",
		"Started implement phase",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}

	// Read log file and verify entry was added
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	logStr := string(logContent)

	// Check for required fields (project logs don't have agent ID)
	if !strings.Contains(logStr, "**Action**: started_phase") {
		t.Error("log entry missing action")
	}
	if !strings.Contains(logStr, "Started implement phase") {
		t.Error("log entry missing details")
	}
}

func TestLogCommand_MultipleFiles(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	taskDir := filepath.Join(sowDir, "project", "phases", "implement", "tasks", "040")

	os.MkdirAll(taskDir, 0755)

	// Write task state
	stateContent := `task:
  id: "040"
  phase: "implement"
  iteration: 2
  assigned_agent: "architect"
`
	os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)

	// Create initial log file
	logPath := filepath.Join(taskDir, "log.md")
	os.WriteFile(logPath, []byte("# Task Log\n\n---\n"), 0644)

	// Change to task directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Run log command with multiple files
	cmd := NewLogCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--action", "modified_file",
		"--result", "success",
		"--file", "src/main.go",
		"--file", "src/utils.go",
		"--file", "README.md",
		"Updated multiple files",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}

	// Read log file and verify all files listed
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	logStr := string(logContent)

	if !strings.Contains(logStr, "- src/main.go") {
		t.Error("log entry missing first file")
	}
	if !strings.Contains(logStr, "- src/utils.go") {
		t.Error("log entry missing second file")
	}
	if !strings.Contains(logStr, "- README.md") {
		t.Error("log entry missing third file")
	}
	if !strings.Contains(logStr, "architect-2") {
		t.Error("log entry has wrong agent ID")
	}
}

func TestLogCommand_NoFiles(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)

	// Create initial log file
	logPath := filepath.Join(projectDir, "log.md")
	os.WriteFile(logPath, []byte("# Project Log\n\n---\n"), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(projectDir)

	// Run log command without files
	cmd := NewLogCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--project",
		"--action", "phase_completed",
		"--result", "success",
		"Design phase complete",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}

	// Read log file and verify no files section
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}

	logStr := string(logContent)

	// Should not have Files section if no files provided
	if strings.Contains(logStr, "**Files**:") {
		t.Error("log entry should not have Files section when no files provided")
	}
}

func TestLogCommand_MissingRequired(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")

	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "log.md"), []byte("# Log\n"), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(projectDir)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing action",
			args: []string{"--project", "--result", "success", "Details"},
		},
		{
			name: "missing result",
			args: []string{"--project", "--action", "started", "Details"},
		},
		{
			name: "missing details",
			args: []string{"--project", "--action", "started", "--result", "success"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLogCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing required flag, got nil")
			}
		})
	}
}

func TestLogCommand_Performance(t *testing.T) {
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

	// Create initial log file
	logPath := filepath.Join(taskDir, "log.md")
	os.WriteFile(logPath, []byte("# Task Log\n\n---\n"), 0644)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	// Measure execution time
	start := time.Now()

	cmd := NewLogCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{
		"--action", "created_file",
		"--result", "success",
		"--file", "test.go",
		"Performance test",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}

	elapsed := time.Since(start)

	// Should complete in under 1 second (requirement)
	if elapsed > time.Second {
		t.Errorf("log command took %v, expected < 1s", elapsed)
	}

	// Ideally under 100ms
	if elapsed > 100*time.Millisecond {
		t.Logf("Warning: log command took %v, ideally should be < 100ms", elapsed)
	}
}
