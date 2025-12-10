package project

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
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
	if err := proj.Save(context.Background()); err != nil {
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

// TestCountTasksByStatus verifies progress calculation for a phase.
func TestCountTasksByStatus(t *testing.T) {
	tests := []struct {
		name           string
		tasks          []projschema.TaskState
		wantCompleted  int
		wantTotal      int
	}{
		{
			name:          "empty tasks",
			tasks:         []projschema.TaskState{},
			wantCompleted: 0,
			wantTotal:     0,
		},
		{
			name: "all pending",
			tasks: []projschema.TaskState{
				{Id: "001", Status: "pending"},
				{Id: "002", Status: "pending"},
			},
			wantCompleted: 0,
			wantTotal:     2,
		},
		{
			name: "all completed",
			tasks: []projschema.TaskState{
				{Id: "001", Status: "completed"},
				{Id: "002", Status: "completed"},
			},
			wantCompleted: 2,
			wantTotal:     2,
		},
		{
			name: "mixed statuses",
			tasks: []projschema.TaskState{
				{Id: "001", Status: "completed"},
				{Id: "002", Status: "in_progress"},
				{Id: "003", Status: "pending"},
				{Id: "004", Status: "needs_review"},
				{Id: "005", Status: "completed"},
			},
			wantCompleted: 2,
			wantTotal:     5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			phase := projschema.PhaseState{Tasks: tc.tasks}
			completed, total := countTasksByStatus(phase)
			if completed != tc.wantCompleted {
				t.Errorf("completed = %d, want %d", completed, tc.wantCompleted)
			}
			if total != tc.wantTotal {
				t.Errorf("total = %d, want %d", total, tc.wantTotal)
			}
		})
	}
}

// TestFormatPhaseLine verifies phase line formatting.
func TestFormatPhaseLine(t *testing.T) {
	tests := []struct {
		name      string
		phaseName string
		status    string
		completed int
		total     int
		wantParts []string
	}{
		{
			name:      "basic phase line",
			phaseName: "implementation",
			status:    "in_progress",
			completed: 2,
			total:     5,
			wantParts: []string{"implementation", "in_progress", "2/5 tasks completed"},
		},
		{
			name:      "zero tasks",
			phaseName: "review",
			status:    "pending",
			completed: 0,
			total:     0,
			wantParts: []string{"review", "pending", "0/0 tasks completed"},
		},
		{
			name:      "all completed",
			phaseName: "finalize",
			status:    "completed",
			completed: 3,
			total:     3,
			wantParts: []string{"finalize", "completed", "3/3 tasks completed"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPhaseLine(tc.phaseName, tc.status, tc.completed, tc.total)
			for _, part := range tc.wantParts {
				if !strings.Contains(result, part) {
					t.Errorf("formatPhaseLine() = %q, missing %q", result, part)
				}
			}
			// Verify it ends with newline
			if !strings.HasSuffix(result, "\n") {
				t.Errorf("formatPhaseLine() should end with newline")
			}
		})
	}
}

// TestFormatPhaseLine_Alignment verifies consistent alignment across phase names.
func TestFormatPhaseLine_Alignment(t *testing.T) {
	// Format lines for all three phases
	impl := formatPhaseLine("implementation", "in_progress", 2, 5)
	review := formatPhaseLine("review", "pending", 0, 0)
	finalize := formatPhaseLine("finalize", "completed", 1, 1)

	// All lines should have the status at the same position
	// This checks that shorter phase names are padded
	implBracketPos := strings.Index(impl, "[")
	reviewBracketPos := strings.Index(review, "[")
	finalizeBracketPos := strings.Index(finalize, "[")

	if implBracketPos != reviewBracketPos || reviewBracketPos != finalizeBracketPos {
		t.Errorf("Status brackets not aligned: impl=%d, review=%d, finalize=%d",
			implBracketPos, reviewBracketPos, finalizeBracketPos)
	}
}

// TestFormatTaskLine verifies task line formatting.
func TestFormatTaskLine(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		id        string
		taskName  string
		wantParts []string
	}{
		{
			name:      "pending task",
			status:    "pending",
			id:        "010",
			taskName:  "Implement feature",
			wantParts: []string{"[pending", "010", "Implement feature"},
		},
		{
			name:      "needs_review status (max width)",
			status:    "needs_review",
			id:        "001",
			taskName:  "Test task",
			wantParts: []string{"[needs_review]", "001", "Test task"},
		},
		{
			name:      "in_progress status",
			status:    "in_progress",
			id:        "999",
			taskName:  "Working on it",
			wantParts: []string{"[in_progress", "999", "Working on it"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatTaskLine(tc.status, tc.id, tc.taskName)
			for _, part := range tc.wantParts {
				if !strings.Contains(result, part) {
					t.Errorf("formatTaskLine() = %q, missing %q", result, part)
				}
			}
			// Verify it ends with newline
			if !strings.HasSuffix(result, "\n") {
				t.Errorf("formatTaskLine() should end with newline")
			}
		})
	}
}

// TestFormatTaskLine_StatusAlignment verifies consistent status width.
func TestFormatTaskLine_StatusAlignment(t *testing.T) {
	// Format lines with different status lengths
	pending := formatTaskLine("pending", "001", "Task A")
	needsReview := formatTaskLine("needs_review", "002", "Task B")
	completed := formatTaskLine("completed", "003", "Task C")

	// Find position of task ID in each line - should be aligned
	pendingIDPos := strings.Index(pending, "001")
	needsReviewIDPos := strings.Index(needsReview, "002")
	completedIDPos := strings.Index(completed, "003")

	if pendingIDPos != needsReviewIDPos || needsReviewIDPos != completedIDPos {
		t.Errorf("Task IDs not aligned: pending=%d, needs_review=%d, completed=%d",
			pendingIDPos, needsReviewIDPos, completedIDPos)
	}
}

// TestRunStatus_PhaseOrderCorrect verifies phases appear in lifecycle order.
func TestRunStatus_PhaseOrderCorrect(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	proj, err := initializeProject(sowCtx, "feat/test-order", "Test phase order", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Add all three phases with tasks
	now := proj.Created_at
	proj.Phases["implementation"] = projschema.PhaseState{
		Status:     "in_progress",
		Enabled:    true,
		Created_at: now,
		Inputs:     []projschema.ArtifactState{},
		Outputs:    []projschema.ArtifactState{},
		Tasks: []projschema.TaskState{
			{Id: "010", Name: "Impl task", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
		},
	}
	proj.Phases["review"] = projschema.PhaseState{
		Status:     "pending",
		Enabled:    true,
		Created_at: now,
		Inputs:     []projschema.ArtifactState{},
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
	}
	proj.Phases["finalize"] = projschema.PhaseState{
		Status:     "pending",
		Enabled:    true,
		Created_at: now,
		Inputs:     []projschema.ArtifactState{},
		Outputs:    []projschema.ArtifactState{},
		Tasks:      []projschema.TaskState{},
	}

	if err := proj.Save(context.Background()); err != nil {
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

	// Verify order: implementation before review before finalize
	implPos := strings.Index(output, "implementation")
	reviewPos := strings.Index(output, "review")
	finalizePos := strings.Index(output, "finalize")

	if implPos == -1 || reviewPos == -1 || finalizePos == -1 {
		t.Fatalf("missing phase in output: impl=%d, review=%d, finalize=%d", implPos, reviewPos, finalizePos)
	}

	if implPos >= reviewPos {
		t.Errorf("implementation should appear before review: impl=%d, review=%d", implPos, reviewPos)
	}
	if reviewPos >= finalizePos {
		t.Errorf("review should appear before finalize: review=%d, finalize=%d", reviewPos, finalizePos)
	}
}

// TestRunStatus_TaskCountAccurate verifies task progress is calculated correctly.
func TestRunStatus_TaskCountAccurate(t *testing.T) {
	sowCtx, _ := setupTestContext(t)

	// Initialize a project
	proj, err := initializeProject(sowCtx, "feat/test-count", "Test task count", nil, nil)
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Add tasks with mixed statuses
	now := proj.Created_at
	implPhase := proj.Phases["implementation"]
	implPhase.Tasks = []projschema.TaskState{
		{Id: "010", Name: "Task 1", Phase: "implementation", Status: "completed", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
		{Id: "020", Name: "Task 2", Phase: "implementation", Status: "completed", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
		{Id: "030", Name: "Task 3", Phase: "implementation", Status: "in_progress", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
		{Id: "040", Name: "Task 4", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
		{Id: "050", Name: "Task 5", Phase: "implementation", Status: "pending", Iteration: 1, Assigned_agent: "implementer", Created_at: now, Updated_at: now, Inputs: []projschema.ArtifactState{}, Outputs: []projschema.ArtifactState{}},
	}
	proj.Phases["implementation"] = implPhase

	if err := proj.Save(context.Background()); err != nil {
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

	// Should show 2/5 tasks completed
	if !strings.Contains(output, "2/5 tasks completed") {
		t.Errorf("expected '2/5 tasks completed' in output, got:\n%s", output)
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
