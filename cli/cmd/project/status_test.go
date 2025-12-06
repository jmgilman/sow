package project

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
	"github.com/spf13/cobra"
)

// TestNewStatusCmd_Structure verifies the status command has correct structure.
func TestNewStatusCmd_Structure(t *testing.T) {
	cmd := newStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("expected Use='status', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Short != "Show current project status" {
		t.Errorf("expected Short='Show current project status', got '%s'", cmd.Short)
	}
}

// TestNewStatusCmd_HasLongDescription verifies long description exists.
func TestNewStatusCmd_HasLongDescription(t *testing.T) {
	cmd := newStatusCmd()

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestRunStatus_NoProject_ReturnsError verifies error when no project exists.
func TestRunStatus_NoProject_ReturnsError(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runStatus(cmd, nil)

	if err == nil {
		t.Fatal("expected error when no project exists")
	}

	if !strings.Contains(err.Error(), "no active project") {
		t.Errorf("expected error to contain 'no active project', got: %v", err)
	}
}

// TestRunStatus_WithProject_ShowsProjectHeader verifies project header output.
func TestRunStatus_WithProject_ShowsProjectHeader(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	_, err := initializeProject(sowCtx, "feat/test-status", "Test status command", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify project header elements
	if !strings.Contains(output, "Project:") {
		t.Error("expected output to contain 'Project:'")
	}
	if !strings.Contains(output, "test-status") {
		t.Error("expected output to contain project name 'test-status'")
	}
	if !strings.Contains(output, "Branch:") {
		t.Error("expected output to contain 'Branch:'")
	}
	if !strings.Contains(output, "feat/test-status") {
		t.Error("expected output to contain branch 'feat/test-status'")
	}
	if !strings.Contains(output, "Type:") {
		t.Error("expected output to contain 'Type:'")
	}
	if !strings.Contains(output, "standard") {
		t.Error("expected output to contain type 'standard'")
	}
}

// TestRunStatus_WithProject_ShowsState verifies state is shown.
func TestRunStatus_WithProject_ShowsState(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	_, err := initializeProject(sowCtx, "feat/test-state", "Test state display", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "State:") {
		t.Error("expected output to contain 'State:'")
	}
}

// TestRunStatus_WithProject_ShowsPhases verifies phases section exists.
func TestRunStatus_WithProject_ShowsPhases(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	_, err := initializeProject(sowCtx, "feat/test-phases", "Test phases display", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Phases:") {
		t.Error("expected output to contain 'Phases:'")
	}
	// Standard project should have implementation phase
	if !strings.Contains(output, "implementation") {
		t.Error("expected output to contain 'implementation' phase")
	}
}

// TestRunStatus_WithProject_NoTasksHidesTasksSection verifies tasks section is hidden when no tasks exist.
func TestRunStatus_WithProject_NoTasksHidesTasksSection(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project (standard project starts with empty tasks)
	_, err := initializeProject(sowCtx, "feat/test-tasks", "Test tasks display", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should NOT have tasks section when no tasks exist
	if strings.Contains(output, "Tasks (") {
		t.Error("expected output to NOT contain 'Tasks' section when no tasks exist")
	}
}

// TestRunStatus_WithTasks_ShowsTasksSection verifies tasks section appears when tasks exist.
func TestRunStatus_WithTasks_ShowsTasksSection(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	proj, err := initializeProject(sowCtx, "feat/test-tasks", "Test tasks display", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Add a task to the implementation phase
	implPhase := proj.Phases["implementation"]
	implPhase.Tasks = []projschema.TaskState{
		{
			Id:             "010",
			Name:           "Test task",
			Phase:          "implementation",
			Status:         "pending",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     proj.Created_at,
			Updated_at:     proj.Created_at,
			Inputs:         []projschema.ArtifactState{},
			Outputs:        []projschema.ArtifactState{},
		},
	}
	proj.Phases["implementation"] = implPhase

	// Save the updated project
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should have tasks section
	if !strings.Contains(output, "Tasks (implementation):") {
		t.Error("expected output to contain 'Tasks (implementation):'")
	}
	if !strings.Contains(output, "010") {
		t.Error("expected output to contain task ID '010'")
	}
	if !strings.Contains(output, "Test task") {
		t.Error("expected output to contain task name 'Test task'")
	}
	if !strings.Contains(output, "pending") {
		t.Error("expected output to contain task status 'pending'")
	}
}

// TestRunStatus_OutputsToStdout verifies output goes to stdout not stderr.
func TestRunStatus_OutputsToStdout(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	_, err := initializeProject(sowCtx, "feat/test-stdout", "Test stdout", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetContext(cmdutil.WithContext(context.Background(), sowCtx))

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = runStatus(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected output to go to stdout")
	}

	if stderr.Len() != 0 {
		t.Errorf("expected no output to stderr, got: %s", stderr.String())
	}
}

// TestInferCurrentPhase verifies phase inference from state names.
func TestInferCurrentPhase(t *testing.T) {
	tests := []struct {
		name       string
		stateName  string
		wantPhase  string
	}{
		{
			name:      "implementation planning state",
			stateName: "ImplementationPlanning",
			wantPhase: "implementation",
		},
		{
			name:      "implementation executing state",
			stateName: "ImplementationExecuting",
			wantPhase: "implementation",
		},
		{
			name:      "review active state",
			stateName: "ReviewActive",
			wantPhase: "review",
		},
		{
			name:      "finalize checks state",
			stateName: "FinalizeChecks",
			wantPhase: "finalize",
		},
		{
			name:      "finalize cleanup state",
			stateName: "FinalizeCleanup",
			wantPhase: "finalize",
		},
		{
			name:      "unknown state defaults to implementation",
			stateName: "SomeUnknownState",
			wantPhase: "implementation",
		},
		{
			name:      "empty state defaults to implementation",
			stateName: "",
			wantPhase: "implementation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inferCurrentPhase(tc.stateName)
			if got != tc.wantPhase {
				t.Errorf("inferCurrentPhase(%q) = %q, want %q", tc.stateName, got, tc.wantPhase)
			}
		})
	}
}
