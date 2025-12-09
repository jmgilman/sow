package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/schemas"
	"github.com/jmgilman/sow/libs/schemas/project"
	"gopkg.in/yaml.v3"

	// Register built-in project types for testing.
	_ "github.com/jmgilman/sow/cli/internal/projects/standard"
)

// TestNewSpawnCmd_Structure verifies the spawn command has correct structure.
func TestNewSpawnCmd_Structure(t *testing.T) {
	cmd := newSpawnCmd()

	if cmd.Use != "spawn [task-id]" {
		t.Errorf("expected Use='spawn [task-id]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewSpawnCmd_AcceptsOptionalArg verifies spawn accepts 0 or 1 args.
func TestNewSpawnCmd_AcceptsOptionalArg(t *testing.T) {
	cmd := newSpawnCmd()

	// The Args field should be set to accept 0 or 1 arg
	if cmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

// TestNewSpawnCmd_HasPhaseFlag verifies --phase flag exists.
func TestNewSpawnCmd_HasPhaseFlag(t *testing.T) {
	cmd := newSpawnCmd()

	phaseFlag := cmd.Flags().Lookup("phase")
	if phaseFlag == nil {
		t.Error("expected --phase flag to be defined")
	}
}

// TestNewSpawnCmd_HasAgentFlag verifies --agent flag exists.
func TestNewSpawnCmd_HasAgentFlag(t *testing.T) {
	cmd := newSpawnCmd()

	agentFlag := cmd.Flags().Lookup("agent")
	if agentFlag == nil {
		t.Error("expected --agent flag to be defined")
	}
}

// TestNewSpawnCmd_HasPromptFlag verifies --prompt flag exists.
func TestNewSpawnCmd_HasPromptFlag(t *testing.T) {
	cmd := newSpawnCmd()

	promptFlag := cmd.Flags().Lookup("prompt")
	if promptFlag == nil {
		t.Error("expected --prompt flag to be defined")
	}
}

// TestBuildTaskPrompt_Format tests the prompt builder function.
func TestBuildTaskPrompt_Format(t *testing.T) {
	prompt := buildTaskPrompt("010", "implementation")

	// Check that prompt contains task ID
	if !strings.Contains(prompt, "010") {
		t.Error("expected prompt to contain task ID '010'")
	}

	// Check that prompt contains phase
	if !strings.Contains(prompt, "implementation") {
		t.Error("expected prompt to contain phase 'implementation'")
	}

	// Check that prompt contains the task location path
	expectedPath := ".sow/project/phases/implementation/tasks/010/"
	if !strings.Contains(prompt, expectedPath) {
		t.Errorf("expected prompt to contain path '%s'", expectedPath)
	}

	// Check that prompt mentions state.yaml
	if !strings.Contains(prompt, "state.yaml") {
		t.Error("expected prompt to mention state.yaml")
	}

	// Check that prompt mentions description.md
	if !strings.Contains(prompt, "description.md") {
		t.Error("expected prompt to mention description.md")
	}

	// Check that prompt mentions feedback/
	if !strings.Contains(prompt, "feedback/") {
		t.Error("expected prompt to mention feedback/")
	}
}

// TestBuildTaskPrompt_DifferentTaskIDs verifies prompt works with different IDs.
func TestBuildTaskPrompt_DifferentTaskIDs(t *testing.T) {
	testCases := []struct {
		taskID   string
		phase    string
		wantPath string
	}{
		{"010", "implementation", ".sow/project/phases/implementation/tasks/010/"},
		{"020", "planning", ".sow/project/phases/planning/tasks/020/"},
		{"999", "review", ".sow/project/phases/review/tasks/999/"},
	}

	for _, tc := range testCases {
		t.Run(tc.taskID, func(t *testing.T) {
			prompt := buildTaskPrompt(tc.taskID, tc.phase)
			if !strings.Contains(prompt, tc.wantPath) {
				t.Errorf("expected prompt to contain path '%s', got:\n%s", tc.wantPath, prompt)
			}
		})
	}
}

// initGitRepo initializes a git repository in the given directory.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmdCtx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(cmdCtx, "git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	// Disable GPG signing for tests
	cmd = exec.CommandContext(cmdCtx, "git", "config", "commit.gpgsign", "false")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to disable gpgsign: %v", err)
	}

	// Create initial commit
	cmd = exec.CommandContext(cmdCtx, "git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}
}

// setupTestProject creates a temporary directory with a valid project state
// and returns a sow.Context, the temp dir path, and a cleanup function.
func setupTestProject(t *testing.T, tasks []project.TaskState) (*sow.Context, string, func()) {
	t.Helper()

	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "sow-spawn-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repository
	initGitRepo(t, tmpDir)

	// Create .sow directory
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.MkdirAll(filepath.Join(sowDir, "project"), 0755); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to create .sow/project: %v", err)
	}

	// Create project state
	now := time.Now()
	projectState := project.ProjectState{
		Name:        "test-project",
		Type:        "standard",
		Branch:      "feat/test",
		Description: "Test project for spawn tests",
		Created_at:  now,
		Updated_at:  now,
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status:     "in_progress",
				Enabled:    true,
				Created_at: now,
				Started_at: now,
				Tasks:      tasks,
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "ImplementationExecuting",
			Updated_at:    now,
		},
	}

	// Write state.yaml
	stateData, err := yaml.Marshal(projectState)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to marshal project state: %v", err)
	}

	statePath := filepath.Join(sowDir, "project", "state.yaml")
	if err := os.WriteFile(statePath, stateData, 0644); err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	// Create task directories with required files
	for _, task := range tasks {
		taskDir := filepath.Join(sowDir, "project", "phases", task.Phase, "tasks", task.Id)
		if err := os.MkdirAll(filepath.Join(taskDir, "feedback"), 0755); err != nil {
			_ = os.RemoveAll(tmpDir)
			t.Fatalf("failed to create task dir: %v", err)
		}
		// Create description.md
		descPath := filepath.Join(taskDir, "description.md")
		if err := os.WriteFile(descPath, []byte("# Test Task\n"), 0644); err != nil {
			_ = os.RemoveAll(tmpDir)
			t.Fatalf("failed to create description.md: %v", err)
		}
		// Create log.md
		logPath := filepath.Join(taskDir, "log.md")
		if err := os.WriteFile(logPath, []byte("# Task Log\n"), 0644); err != nil {
			_ = os.RemoveAll(tmpDir)
			t.Fatalf("failed to create log.md: %v", err)
		}
	}

	// Create sow.Context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to create sow context: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return ctx, tmpDir, cleanup
}

// TestRunSpawn_RequiresTaskIDOrAgentFlag tests validation.
func TestRunSpawn_RequiresTaskIDOrAgentFlag(t *testing.T) {
	sowCtx, _, cleanup := setupTestProject(t, []project.TaskState{})
	defer cleanup()

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run without task-id or --agent
	err := runSpawn(cmd, []string{}, "", "", "")
	if err == nil {
		t.Fatal("expected error when neither task-id nor --agent provided")
	}
	if !strings.Contains(err.Error(), "must provide either") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

// TestRunSpawn_TaskHasUnknownAgent tests error when task has invalid assigned_agent.
func TestRunSpawn_TaskHasUnknownAgent(t *testing.T) {
	// Setup test project with a task that has an unknown agent
	tasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Test Task",
			Phase:          "implementation",
			Status:         "pending",
			Iteration:      1,
			Assigned_agent: "badagent", // Unknown agent
			Created_at:     time.Now(),
			Updated_at:     time.Now(),
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock registry with mock executor
	mockExec := &agents.MockExecutor{}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn - agent comes from task's assigned_agent field
	err := runSpawn(cmd, []string{"010"}, "", "", "")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}

	// Check error message mentions task and unknown agent
	if !strings.Contains(err.Error(), "010") {
		t.Errorf("expected error to contain task ID '010', got: %v", err)
	}
	if !strings.Contains(err.Error(), "badagent") {
		t.Errorf("expected error to contain 'badagent', got: %v", err)
	}

	// Check that available agents are listed
	expectedAgents := []string{"architect", "decomposer", "implementer", "planner", "researcher", "reviewer"}
	for _, agent := range expectedAgents {
		if !strings.Contains(err.Error(), agent) {
			t.Errorf("expected error to list available agent '%s', got: %v", agent, err)
		}
	}
}

// TestRunSpawn_TaskNotFound tests error for non-existent task.
func TestRunSpawn_TaskNotFound(t *testing.T) {
	// Setup test project with no tasks
	tasks := []project.TaskState{}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Create mock registry with mock executor
	mockExec := &agents.MockExecutor{}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run with non-existent task
	err := runSpawn(cmd, []string{"999"}, "implementation", "", "")
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

// TestRunSpawn_GeneratesSessionID verifies session ID generated when empty.
func TestRunSpawn_GeneratesSessionID(t *testing.T) {
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
	sowCtx, tmpDir, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track if spawn was called
	var spawnedSessionID string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, _ string, sessionID string) error {
			spawnedSessionID = sessionID
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err := runSpawn(cmd, []string{"010"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session ID was generated (non-empty and looks like a UUID)
	if spawnedSessionID == "" {
		t.Error("expected session ID to be generated")
	}
	// UUIDs are 36 characters with dashes
	if len(spawnedSessionID) != 36 {
		t.Errorf("expected session ID to be UUID format (36 chars), got length %d: %s",
			len(spawnedSessionID), spawnedSessionID)
	}

	// Verify session ID was persisted to state file
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
	if phase.Tasks[0].Session_id != spawnedSessionID {
		t.Errorf("expected session ID '%s' to be persisted, got '%s'",
			spawnedSessionID, phase.Tasks[0].Session_id)
	}
}

// TestRunSpawn_PreservesExistingSessionID verifies existing session ID not overwritten.
func TestRunSpawn_PreservesExistingSessionID(t *testing.T) {
	// Setup test project with a task that already has a session ID
	existingSessionID := "existing-session-id-12345"
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

	// Track what session ID was passed to spawn
	var spawnedSessionID string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, _ string, sessionID string) error {
			spawnedSessionID = sessionID
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err := runSpawn(cmd, []string{"010"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify existing session ID was used
	if spawnedSessionID != existingSessionID {
		t.Errorf("expected existing session ID '%s', got '%s'",
			existingSessionID, spawnedSessionID)
	}
}

// TestRunSpawn_PersistsSessionBeforeSpawn verifies state saved before executor.Spawn called.
func TestRunSpawn_PersistsSessionBeforeSpawn(t *testing.T) {
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
	sowCtx, tmpDir, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track call order to verify save happens before spawn
	var callOrder []string

	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, _ string, sessionID string) error {
			// At the time spawn is called, check if state was already saved
			stateData, err := os.ReadFile(filepath.Join(tmpDir, ".sow", "project", "state.yaml"))
			if err != nil {
				t.Fatalf("state.yaml should exist when spawn is called: %v", err)
			}

			var savedState project.ProjectState
			if err := yaml.Unmarshal(stateData, &savedState); err != nil {
				t.Fatalf("failed to unmarshal state: %v", err)
			}

			phase := savedState.Phases["implementation"]
			if len(phase.Tasks) == 0 {
				t.Fatal("expected task to exist in saved state")
			}

			// Verify session ID was persisted BEFORE spawn was called
			switch phase.Tasks[0].Session_id {
			case "":
				callOrder = append(callOrder, "spawn_without_saved_session")
			case sessionID:
				callOrder = append(callOrder, "spawn_with_saved_session")
			default:
				callOrder = append(callOrder, "spawn_with_different_session")
			}

			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err := runSpawn(cmd, []string{"010"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the order: session should be saved before spawn
	if len(callOrder) != 1 || callOrder[0] != "spawn_with_saved_session" {
		t.Errorf("expected session to be saved before spawn, got: %v", callOrder)
	}
}

// TestRunSpawn_CallsExecutorSpawn verifies executor.Spawn called with correct args.
func TestRunSpawn_CallsExecutorSpawn(t *testing.T) {
	// Setup test project with a task
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
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track spawn call
	var spawnCalled bool
	var spawnedAgent *agents.Agent
	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, agent *agents.Agent, prompt string, _ string) error {
			spawnCalled = true
			spawnedAgent = agent
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err := runSpawn(cmd, []string{"010"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify spawn was called
	if !spawnCalled {
		t.Fatal("expected executor.Spawn to be called")
	}

	// Verify agent
	if spawnedAgent == nil {
		t.Fatal("expected agent to be passed to Spawn")
	}
	if spawnedAgent.Name != "implementer" {
		t.Errorf("expected agent name 'implementer', got '%s'", spawnedAgent.Name)
	}

	// Verify prompt contains expected content
	if !strings.Contains(spawnedPrompt, "010") {
		t.Error("expected prompt to contain task ID")
	}
	if !strings.Contains(spawnedPrompt, "implementation") {
		t.Error("expected prompt to contain phase")
	}
}

// TestRunSpawn_BuildsCorrectPrompt verifies prompt includes task location.
func TestRunSpawn_BuildsCorrectPrompt(t *testing.T) {
	// Setup test project with a task
	now := time.Now()
	tasks := []project.TaskState{
		{
			Id:             "020",
			Name:           "Another Task",
			Phase:          "implementation",
			Status:         "pending",
			Iteration:      1,
			Assigned_agent: "architect",
			Created_at:     now,
			Updated_at:     now,
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track spawn call
	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, prompt string, _ string) error {
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err := runSpawn(cmd, []string{"020"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify prompt contains task location
	expectedPath := ".sow/project/phases/implementation/tasks/020/"
	if !strings.Contains(spawnedPrompt, expectedPath) {
		t.Errorf("expected prompt to contain task location '%s', got:\n%s",
			expectedPath, spawnedPrompt)
	}
}

// TestRunSpawn_NotInitialized tests error when sow not initialized.
func TestRunSpawn_NotInitialized(t *testing.T) {
	// Create temp dir without .sow
	tmpDir, err := os.MkdirTemp("", "sow-spawn-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize git repo (required for sow.NewContext)
	initGitRepo(t, tmpDir)

	// Create sow.Context that's not initialized (no .sow directory)
	sowCtx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create sow context: %v", err)
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err = runSpawn(cmd, []string{"010"}, "", "", "")
	if err == nil {
		t.Fatal("expected error when sow not initialized")
	}

	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected error to mention 'not initialized', got: %v", err)
	}
}

// TestRunSpawn_NoProject tests error when no project exists.
func TestRunSpawn_NoProject(t *testing.T) {
	// Create temp dir with .sow but no project
	tmpDir, err := os.MkdirTemp("", "sow-spawn-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn
	err = runSpawn(cmd, []string{"010"}, "", "", "")
	if err == nil {
		t.Fatal("expected error when no project exists")
	}

	if !strings.Contains(err.Error(), "no active project") {
		t.Errorf("expected error to mention 'no active project', got: %v", err)
	}
}

// TestRunSpawn_WithPhaseFlag tests using explicit --phase flag.
func TestRunSpawn_WithPhaseFlag(t *testing.T) {
	// Setup test project with a task
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
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	// Track spawn call
	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, prompt string, _ string) error {
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	// Save original and restore after test
	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	// Create command with context
	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run spawn with explicit phase
	err := runSpawn(cmd, []string{"010"}, "implementation", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify prompt uses the specified phase
	if !strings.Contains(spawnedPrompt, "implementation") {
		t.Error("expected prompt to contain phase 'implementation'")
	}
}

// TestRunSpawn_TaskModeWithAgentOverride tests --agent override in task mode.
func TestRunSpawn_TaskModeWithAgentOverride(t *testing.T) {
	// Setup with implementer task
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
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	var spawnedAgent *agents.Agent
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, agent *agents.Agent, _ string, _ string) error {
			spawnedAgent = agent
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run with --agent override
	err := runSpawn(cmd, []string{"010"}, "implementation", "reviewer", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify override worked
	if spawnedAgent == nil || spawnedAgent.Name != "reviewer" {
		t.Error("expected reviewer agent (override), not implementer")
	}
}

// TestRunSpawn_TaskModeWithCustomPrompt tests --prompt flag in task mode.
func TestRunSpawn_TaskModeWithCustomPrompt(t *testing.T) {
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
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}
	sowCtx, _, cleanup := setupTestProject(t, tasks)
	defer cleanup()

	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, prompt string, _ string) error {
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	customPrompt := "Focus on error handling and edge cases"
	err := runSpawn(cmd, []string{"010"}, "implementation", "", customPrompt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify custom prompt was appended
	if !strings.Contains(spawnedPrompt, customPrompt) {
		t.Errorf("expected prompt to contain custom prompt '%s', got:\n%s", customPrompt, spawnedPrompt)
	}

	// Verify task prompt is still there
	if !strings.Contains(spawnedPrompt, "010") {
		t.Error("expected prompt to still contain task ID")
	}
}

// TestRunSpawn_TasklessMode tests spawning without task.
func TestRunSpawn_TasklessMode(t *testing.T) {
	sowCtx, tmpDir, cleanup := setupTestProject(t, []project.TaskState{})
	defer cleanup()

	var spawnedSessionID string
	var spawnedAgent *agents.Agent
	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, agent *agents.Agent, prompt string, sessionID string) error {
			spawnedSessionID = sessionID
			spawnedAgent = agent
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run taskless spawn
	err := runSpawn(cmd, []string{}, "", "planner", "Create a plan for auth feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify agent
	if spawnedAgent == nil || spawnedAgent.Name != "planner" {
		t.Error("expected planner agent to be spawned")
	}

	// Verify prompt
	if !strings.Contains(spawnedPrompt, "Create a plan for auth feature") {
		t.Error("expected prompt to contain custom prompt")
	}

	// Verify session ID generated
	if len(spawnedSessionID) != 36 {
		t.Errorf("expected UUID session ID, got: %s", spawnedSessionID)
	}

	// Verify persisted to state file
	stateData, _ := os.ReadFile(filepath.Join(tmpDir, ".sow", "project", "state.yaml"))
	var savedState project.ProjectState
	_ = yaml.Unmarshal(stateData, &savedState)

	if savedState.Agent_sessions == nil {
		t.Fatal("expected agent_sessions to be set")
	}
	if savedState.Agent_sessions["planner"] != spawnedSessionID {
		t.Errorf("expected session ID to be persisted")
	}
}

// TestRunSpawn_TasklessModeUnknownAgent tests error for unknown agent in taskless mode.
func TestRunSpawn_TasklessModeUnknownAgent(t *testing.T) {
	sowCtx, _, cleanup := setupTestProject(t, []project.TaskState{})
	defer cleanup()

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run taskless spawn with unknown agent
	err := runSpawn(cmd, []string{}, "", "badagent", "some prompt")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}

	if !strings.Contains(err.Error(), "badagent") {
		t.Errorf("expected error to contain 'badagent', got: %v", err)
	}
	if !strings.Contains(err.Error(), "Available agents") {
		t.Errorf("expected error to list available agents, got: %v", err)
	}
}

// TestRunSpawn_TasklessModeDefaultPrompt tests default prompt in taskless mode.
func TestRunSpawn_TasklessModeDefaultPrompt(t *testing.T) {
	sowCtx, _, cleanup := setupTestProject(t, []project.TaskState{})
	defer cleanup()

	var spawnedPrompt string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, prompt string, _ string) error {
			spawnedPrompt = prompt
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run taskless spawn without custom prompt
	err := runSpawn(cmd, []string{}, "", "planner", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify default prompt was used
	if spawnedPrompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !strings.Contains(spawnedPrompt, "spawned") {
		t.Errorf("expected default prompt about being spawned, got: %s", spawnedPrompt)
	}
}

// TestRunSpawn_TasklessModePreservesExistingSession tests reusing existing session.
func TestRunSpawn_TasklessModePreservesExistingSession(t *testing.T) {
	sowCtx, tmpDir, cleanup := setupTestProject(t, []project.TaskState{})
	defer cleanup()

	// Manually add agent_sessions to state
	existingSessionID := "existing-planner-session-uuid"
	stateData, _ := os.ReadFile(filepath.Join(tmpDir, ".sow", "project", "state.yaml"))
	var savedState project.ProjectState
	_ = yaml.Unmarshal(stateData, &savedState)
	savedState.Agent_sessions = map[string]string{"planner": existingSessionID}
	newData, _ := yaml.Marshal(savedState)
	_ = os.WriteFile(filepath.Join(tmpDir, ".sow", "project", "state.yaml"), newData, 0644)

	var spawnedSessionID string
	mockExec := &agents.MockExecutor{
		SpawnFunc: func(_ context.Context, _ *agents.Agent, _ string, sessionID string) error {
			spawnedSessionID = sessionID
			return nil
		},
	}
	mockRegistry := agents.NewExecutorRegistry()
	mockRegistry.RegisterNamed("claude-code", mockExec)

	originalLoadRegistry := loadExecutorRegistry
	defer func() { loadExecutorRegistry = originalLoadRegistry }()
	loadExecutorRegistry = func(_ *schemas.UserConfig, _ string) (*agents.ExecutorRegistry, error) {
		return mockRegistry, nil
	}

	cmd := newSpawnCmd()
	ctx := cmdutil.WithContext(context.Background(), sowCtx)
	cmd.SetContext(ctx)

	// Run taskless spawn - should reuse existing session
	err := runSpawn(cmd, []string{}, "", "planner", "Continue planning")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify existing session ID was used
	if spawnedSessionID != existingSessionID {
		t.Errorf("expected existing session ID '%s', got '%s'", existingSessionID, spawnedSessionID)
	}
}
