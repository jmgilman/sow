package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
	"gopkg.in/yaml.v3"

	// Register built-in project types for testing.
	_ "github.com/jmgilman/sow/cli/internal/projects/standard"
)

// TestNewResumeCmd_Structure verifies the resume command has correct structure.
func TestNewResumeCmd_Structure(t *testing.T) {
	cmd := newResumeCmd()

	if cmd.Use != "resume <task-id> <prompt>" {
		t.Errorf("expected Use='resume <task-id> <prompt>', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewResumeCmd_RequiresExactlyTwoArgs verifies resume requires task ID and prompt.
func TestNewResumeCmd_RequiresExactlyTwoArgs(t *testing.T) {
	cmd := newResumeCmd()

	// The Args field should be set to require exactly 2 args
	if cmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

// TestNewResumeCmd_HasPhaseFlag verifies --phase flag exists.
func TestNewResumeCmd_HasPhaseFlag(t *testing.T) {
	cmd := newResumeCmd()

	phaseFlag := cmd.Flags().Lookup("phase")
	if phaseFlag == nil {
		t.Error("expected --phase flag to be defined")
	}
}

// TestRunResume_TaskNotFound tests error for non-existent task.
func TestRunResume_TaskNotFound(t *testing.T) {
	// Setup test project with no tasks
	tasks := []project.TaskState{}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock executor that supports resumption
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run with non-existent task
	err := runResume(cmd, []string{"999", "feedback prompt"}, "implementation")
	if err == nil {
		t.Fatal("expected error for task not found")
	}

	// Check error message
	if !strings.Contains(err.Error(), "999") {
		t.Errorf("expected error to contain task ID '999', got: %v", err)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %v", err)
	}
}

// TestRunResume_NoSessionID tests error when task.Session_id is empty.
func TestRunResume_NoSessionID(t *testing.T) {
	// Setup test project with a task that has no session ID
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "pending",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     "", // Empty session ID
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock executor that supports resumption
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume on task with no session
	err := runResume(cmd, []string{"010", "feedback prompt"}, "implementation")
	if err == nil {
		t.Fatal("expected error for missing session")
	}

	// Check error message
	if !strings.Contains(err.Error(), "no session found") {
		t.Errorf("expected 'no session found' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "spawn first") {
		t.Errorf("expected error to mention 'spawn first', got: %v", err)
	}
}

// TestRunResume_ExecutorNoResumption tests error when executor doesn't support resume.
func TestRunResume_ExecutorNoResumption(t *testing.T) {
	// Setup test project with a task that has a session ID
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     "existing-session-123",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock executor that does NOT support resumption
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return false },
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err := runResume(cmd, []string{"010", "feedback prompt"}, "implementation")
	if err == nil {
		t.Fatal("expected error for no resumption support")
	}

	// Check error message
	if !strings.Contains(err.Error(), "does not support session resumption") {
		t.Errorf("expected 'does not support session resumption' error, got: %v", err)
	}
}

// TestRunResume_CallsExecutorResume verifies executor.Resume called with correct args.
func TestRunResume_CallsExecutorResume(t *testing.T) {
	// Setup test project with a task that has a session ID
	existingSessionID := "session-abc-123"
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     existingSessionID,
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track resume call
	var resumeCalled bool
	var resumedSessionID string
	var resumedPrompt string
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
		ResumeFunc: func(ctx context.Context, sessionID string, prompt string) error {
			resumeCalled = true
			resumedSessionID = sessionID
			resumedPrompt = prompt
			return nil
		},
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err := runResume(cmd, []string{"010", "feedback prompt"}, "implementation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify resume was called
	if !resumeCalled {
		t.Fatal("expected executor.Resume to be called")
	}

	// Verify session ID
	if resumedSessionID != existingSessionID {
		t.Errorf("expected session ID '%s', got '%s'", existingSessionID, resumedSessionID)
	}

	// Verify prompt was passed
	if resumedPrompt != "feedback prompt" {
		t.Errorf("expected prompt 'feedback prompt', got '%s'", resumedPrompt)
	}
}

// TestRunResume_PassesCorrectPrompt verifies user's prompt is passed correctly to executor.
func TestRunResume_PassesCorrectPrompt(t *testing.T) {
	// Setup test project with a task that has a session ID
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "020",
			Name:           "Another Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      2,
			Assigned_agent: "architect",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     "session-xyz",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track resume prompt
	var capturedPrompt string
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
		ResumeFunc: func(ctx context.Context, sessionID string, prompt string) error {
			capturedPrompt = prompt
			return nil
		},
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Test with a multi-word prompt
	multiWordPrompt := "Use RS256 algorithm for JWT signing. Please add error handling."
	err := runResume(cmd, []string{"020", multiWordPrompt}, "implementation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify exact prompt was passed
	if capturedPrompt != multiWordPrompt {
		t.Errorf("expected prompt '%s', got '%s'", multiWordPrompt, capturedPrompt)
	}
}

// TestRunResume_NotInitialized tests error when sow not initialized.
func TestRunResume_NotInitialized(t *testing.T) {
	// Create temp dir without .sow
	tmpDir, err := os.MkdirTemp("", "sow-resume-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo (required for sow.NewContext)
	initGitRepo(t, tmpDir)

	// Create sow.Context that's not initialized (no .sow directory)
	sowCtx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create sow context: %v", err)
	}

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err = runResume(cmd, []string{"010", "feedback"}, "")
	if err == nil {
		t.Fatal("expected error when sow not initialized")
	}

	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected error to mention 'not initialized', got: %v", err)
	}
}

// TestRunResume_NoProject tests error when no project exists.
func TestRunResume_NoProject(t *testing.T) {
	// Create temp dir with .sow but no project
	tmpDir, err := os.MkdirTemp("", "sow-resume-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo (required for sow.NewContext)
	initGitRepo(t, tmpDir)

	// Create .sow directory but no project
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.MkdirAll(sowDir, 0755); err != nil {
		t.Fatalf("failed to create .sow: %v", err)
	}

	// Create sow.Context
	sowCtx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create sow context: %v", err)
	}

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err = runResume(cmd, []string{"010", "feedback"}, "")
	if err == nil {
		t.Fatal("expected error when no project exists")
	}

	if !strings.Contains(err.Error(), "no active project") {
		t.Errorf("expected error to mention 'no active project', got: %v", err)
	}
}

// TestRunResume_WithPhaseFlag tests using explicit --phase flag.
func TestRunResume_WithPhaseFlag(t *testing.T) {
	// Setup test project with a task that has a session ID
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     "session-with-phase",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track resume call
	var resumeCalled bool
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
		ResumeFunc: func(ctx context.Context, sessionID string, prompt string) error {
			resumeCalled = true
			return nil
		},
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume with explicit phase
	err := runResume(cmd, []string{"010", "feedback"}, "implementation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify resume was called
	if !resumeCalled {
		t.Fatal("expected executor.Resume to be called")
	}
}

// TestRunResume_ReturnsExecutorError verifies executor errors are propagated.
func TestRunResume_ReturnsExecutorError(t *testing.T) {
	// Setup test project with a task that has a session ID
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     "session-with-error",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock executor that returns an error from Resume
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
		ResumeFunc: func(ctx context.Context, sessionID string, prompt string) error {
			return fmt.Errorf("executor resume error")
		},
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err := runResume(cmd, []string{"010", "feedback"}, "implementation")
	if err == nil {
		t.Fatal("expected error from executor to be propagated")
	}

	if !strings.Contains(err.Error(), "resume failed") {
		t.Errorf("expected 'resume failed' error, got: %v", err)
	}
}

// TestRunResume_DoesNotModifySessionID verifies that session ID is not modified.
func TestRunResume_DoesNotModifySessionID(t *testing.T) {
	// Setup test project with a task that has a session ID
	existingSessionID := "existing-session-do-not-change"
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Session_id:     existingSessionID,
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, tmpDir, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock executor that supports resumption
	mockExec := &agents.MockExecutor{
		SupportsResumptionFunc: func() bool { return true },
		ResumeFunc: func(ctx context.Context, sessionID string, prompt string) error {
			return nil
		},
	}

	// Save original and restore after test
	originalNewExecutor := newExecutor
	defer func() { newExecutor = originalNewExecutor }()
	newExecutor = func() agents.Executor { return mockExec }

	// Create command with context
	cmd := newResumeCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run resume
	err := runResume(cmd, []string{"010", "feedback"}, "implementation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read state file and verify session ID was not changed
	stateData, err := os.ReadFile(filepath.Join(tmpDir, ".sow", "project", "state.yaml"))
	if err != nil {
		t.Fatalf("failed to read state.yaml: %v", err)
	}

	var savedState project.ProjectState
	if err := yaml.Unmarshal(stateData, &savedState); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}

	phase := savedState.Phases["implementation"]
	if len(phase.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(phase.Tasks))
	}
	if phase.Tasks[0].Session_id != existingSessionID {
		t.Errorf("expected session ID to remain '%s', got '%s'",
			existingSessionID, phase.Tasks[0].Session_id)
	}
}
