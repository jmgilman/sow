package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectContext_NoSowDirectory(t *testing.T) {
	// Create temporary directory without .sow
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	ctx, err := DetectContext()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ctx.Type != ContextTypeNone {
		t.Errorf("expected context type %s, got %s", ContextTypeNone, ctx.Type)
	}
}

func TestDetectContext_ProjectLevel(t *testing.T) {
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

	// Change to project directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(projectDir)

	ctx, err := DetectContext()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ctx.Type != ContextTypeProject {
		t.Errorf("expected context type %s, got %s", ContextTypeProject, ctx.Type)
	}

	// Compare resolved paths to handle symlinks
	resolvedExpected, _ := filepath.EvalSymlinks(sowDir)
	resolvedGot, _ := filepath.EvalSymlinks(ctx.SowRoot)
	if resolvedGot != resolvedExpected {
		t.Errorf("expected sow root %s, got %s", resolvedExpected, resolvedGot)
	}
}

func TestDetectContext_TaskLevel(t *testing.T) {
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

	// Change to task directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(taskDir)

	ctx, err := DetectContext()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ctx.Type != ContextTypeTask {
		t.Errorf("expected context type %s, got %s", ContextTypeTask, ctx.Type)
	}

	if ctx.TaskID != "040" {
		t.Errorf("expected task ID 040, got %s", ctx.TaskID)
	}

	if ctx.Phase != "implement" {
		t.Errorf("expected phase implement, got %s", ctx.Phase)
	}

	if ctx.Iteration != 1 {
		t.Errorf("expected iteration 1, got %d", ctx.Iteration)
	}

	if ctx.AgentRole != "implementer" {
		t.Errorf("expected agent role implementer, got %s", ctx.AgentRole)
	}
}

func TestDetectContext_FromSubdirectory(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")
	taskDir := filepath.Join(sowDir, "project", "phases", "design", "tasks", "010")
	subDir := filepath.Join(taskDir, "feedback", "subdir")

	os.MkdirAll(subDir, 0755)

	// Write task state
	stateContent := `task:
  id: "010"
  phase: "design"
  iteration: 2
  assigned_agent: "architect"
`
	os.WriteFile(filepath.Join(taskDir, "state.yaml"), []byte(stateContent), 0644)

	// Change to subdirectory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(subDir)

	ctx, err := DetectContext()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ctx.Type != ContextTypeTask {
		t.Errorf("expected context type %s, got %s", ContextTypeTask, ctx.Type)
	}

	if ctx.TaskID != "010" {
		t.Errorf("expected task ID 010, got %s", ctx.TaskID)
	}
}

func TestGetAgentID(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		iteration int
		want      string
	}{
		{"implementer iteration 1", "implementer", 1, "implementer-1"},
		{"architect iteration 2", "architect", 2, "architect-2"},
		{"orchestrator iteration 5", "orchestrator", 5, "orchestrator-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{
				Type:      ContextTypeTask,
				AgentRole: tt.role,
				Iteration: tt.iteration,
			}

			got := ctx.GetAgentID()
			if got != tt.want {
				t.Errorf("expected agent ID %s, got %s", tt.want, got)
			}
		})
	}
}

func TestGetLogPath_Task(t *testing.T) {
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")

	ctx := &Context{
		Type:    ContextTypeTask,
		SowRoot: sowDir,
		Phase:   "implement",
		TaskID:  "040",
	}

	expected := filepath.Join(sowDir, "project", "phases", "implement", "tasks", "040", "log.md")
	got := ctx.GetLogPath()

	if got != expected {
		t.Errorf("expected log path %s, got %s", expected, got)
	}
}

func TestGetLogPath_Project(t *testing.T) {
	tmpDir := t.TempDir()
	sowDir := filepath.Join(tmpDir, ".sow")

	ctx := &Context{
		Type:    ContextTypeProject,
		SowRoot: sowDir,
	}

	expected := filepath.Join(sowDir, "project", "log.md")
	got := ctx.GetLogPath()

	if got != expected {
		t.Errorf("expected log path %s, got %s", expected, got)
	}
}
