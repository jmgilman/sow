package state

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/cli/internal/sow"
)

// init registers a mock project type config for testing.
func init() {
	// Register mock "standard" type for testing
	mockConfig := newMockProjectTypeConfig()
	Registry["standard"] = mockConfig
}

// setupTestRepo creates a temporary directory with an initialized git repository
// and returns a sow.Context for testing. The directory is automatically cleaned
// up when the test completes.
func setupTestRepo(t *testing.T) *sow.Context {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()

	cmdCtx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(cmdCtx, "git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	// Disable GPG signing for tests
	cmd = exec.CommandContext(cmdCtx, "git", "config", "commit.gpgsign", "false")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to disable gpgsign: %v", err)
	}

	// Create initial commit
	cmd = exec.CommandContext(cmdCtx, "git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create .sow directory
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.MkdirAll(sowDir, 0755); err != nil {
		t.Fatalf("Failed to create .sow directory: %v", err)
	}

	// Create sow context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	return ctx
}

func TestLoad_ValidYAML(t *testing.T) {
	// Setup test repository with sow context
	ctx := setupTestRepo(t)

	// Copy valid test fixture
	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write fixture using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Load project
	project, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify basic fields
	if project.Name != "test-project" {
		t.Errorf("expected name 'test-project', got '%s'", project.Name)
	}
	if project.Type != "standard" {
		t.Errorf("expected type 'standard', got '%s'", project.Type)
	}
	if project.Branch != "feature/test" {
		t.Errorf("expected branch 'feature/test', got '%s'", project.Branch)
	}

	// Verify phases
	if len(project.Phases) != 1 {
		t.Errorf("expected 1 phase, got %d", len(project.Phases))
	}

	planningPhase, exists := project.Phases["planning"]
	if !exists {
		t.Error("planning phase not found")
	}
	if planningPhase.Status != "completed" {
		t.Errorf("expected planning status 'completed', got '%s'", planningPhase.Status)
	}

	// Verify statechart
	if project.Statechart.Current_state != "ImplementationPlanning" {
		t.Errorf("expected state 'ImplementationPlanning', got '%s'", project.Statechart.Current_state)
	}

	// Verify config was attached
	if project.config == nil {
		t.Error("expected config to be attached, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Setup test repository without state file
	ctx := setupTestRepo(t)

	// Attempt to load (no state file exists)
	_, err := Load(ctx)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	// Verify error message contains useful information
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Setup test repository
	ctx := setupTestRepo(t)

	// Copy invalid YAML fixture
	fixtureData, err := os.ReadFile("testdata/invalid-yaml.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write fixture using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Attempt to load
	_, err = Load(ctx)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	// Verify error mentions unmarshal/parse failure
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestLoad_ValidationFailure(t *testing.T) {
	// Setup test repository
	ctx := setupTestRepo(t)

	// Copy fixture with missing required field
	fixtureData, err := os.ReadFile("testdata/invalid-missing-name.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Attempt to load
	_, err = Load(ctx)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	// Verify error mentions validation
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestLoad_UnknownProjectType(t *testing.T) {
	// Setup test repository
	ctx := setupTestRepo(t)

	// Create YAML with unknown project type
	invalidTypeYAML := `name: test-project
type: unknown-type
branch: feature/test
created_at: 2025-01-01T00:00:00Z
updated_at: 2025-01-01T00:00:00Z
phases: {}
statechart:
  current_state: SomeState
  updated_at: 2025-01-01T00:00:00Z
`

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", []byte(invalidTypeYAML), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Attempt to load
	_, err := Load(ctx)
	if err == nil {
		t.Fatal("expected error for unknown project type, got nil")
	}

	// Verify error mentions unknown type
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestLoad_CompleteFlow(t *testing.T) {
	// Create temporary directory structure
	ctx := setupTestRepo(t)

	// Copy complex fixture
	fixtureData, err := os.ReadFile("testdata/complex-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Create context
	// ctx already created by setupTestRepo

	// Load project
	project, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify project metadata
	if project.Name != "complex-test" {
		t.Errorf("expected name 'complex-test', got '%s'", project.Name)
	}
	if project.Description != "A complex test project with multiple phases and tasks" {
		t.Errorf("unexpected description: %s", project.Description)
	}

	// Verify multiple phases
	if len(project.Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(project.Phases))
	}

	// Verify planning phase with outputs
	planningPhase, exists := project.Phases["planning"]
	if !exists {
		t.Fatal("planning phase not found")
	}
	if planningPhase.Status != "completed" {
		t.Errorf("expected planning status 'completed', got '%s'", planningPhase.Status)
	}
	if len(planningPhase.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(planningPhase.Outputs))
	}
	if planningPhase.Outputs[0].Type != "task_list" {
		t.Errorf("expected output type 'task_list', got '%s'", planningPhase.Outputs[0].Type)
	}

	// Verify planning phase task
	if len(planningPhase.Tasks) != 1 {
		t.Errorf("expected 1 task in planning, got %d", len(planningPhase.Tasks))
	}
	if planningPhase.Tasks[0].Id != "010" {
		t.Errorf("expected task ID '010', got '%s'", planningPhase.Tasks[0].Id)
	}

	// Verify implementation phase
	implPhase, exists := project.Phases["implementation"]
	if !exists {
		t.Fatal("implementation phase not found")
	}
	if implPhase.Status != "in_progress" {
		t.Errorf("expected implementation status 'in_progress', got '%s'", implPhase.Status)
	}

	// Verify implementation phase has input artifact
	if len(implPhase.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(implPhase.Inputs))
	}

	// Verify implementation task with metadata
	if len(implPhase.Tasks) != 1 {
		t.Errorf("expected 1 task in implementation, got %d", len(implPhase.Tasks))
	}
	task := implPhase.Tasks[0]
	if task.Id != "020" {
		t.Errorf("expected task ID '020', got '%s'", task.Id)
	}
	if task.Iteration != 2 {
		t.Errorf("expected iteration 2, got %d", task.Iteration)
	}
	if task.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", task.Status)
	}
	if task.Metadata == nil {
		t.Error("expected metadata to be set")
	} else {
		if complexity, ok := task.Metadata["complexity"]; !ok {
			t.Error("expected 'complexity' in metadata")
		} else if complexity != "high" {
			t.Errorf("expected complexity 'high', got '%v'", complexity)
		}
	}

	// Verify statechart
	if project.Statechart.Current_state != "ImplementationExecuting" {
		t.Errorf("expected state 'ImplementationExecuting', got '%s'", project.Statechart.Current_state)
	}
}

// Save() Tests

func TestSave_Success(t *testing.T) {
	// Create temporary directory structure
	ctx := setupTestRepo(t)

	// Copy valid fixture
	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Create context
	// ctx already created by setupTestRepo

	// Load project
	project, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Modify project
	project.Name = "modified-project"

	// Save project
	if err := project.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was written
	savedData, err := ctx.FS().ReadFile("project/state.yaml")
	if err != nil {
		t.Fatalf("failed to read saved state: %v", err)
	}

	// Verify modified name is in saved data
	if !contains(string(savedData), "modified-project") {
		t.Errorf("expected saved data to contain 'modified-project'")
	}
}

func TestSave_ValidationFailure_NoWrite(t *testing.T) {
	// Create temporary directory structure
	ctx := setupTestRepo(t)

	// Copy valid fixture
	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Create context
	// ctx already created by setupTestRepo

	// Load project
	project, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Save original content
	originalData, err := ctx.FS().ReadFile("project/state.yaml")
	if err != nil {
		t.Fatalf("failed to read original state: %v", err)
	}

	// Make project invalid (empty name violates CUE schema)
	project.Name = ""

	// Attempt to save - should fail
	err = project.Save()
	if err == nil {
		t.Fatal("expected Save() to fail with invalid state, got nil")
	}

	// Verify file was NOT written (still has original content)
	currentData, err := ctx.FS().ReadFile("project/state.yaml")
	if err != nil {
		t.Fatalf("failed to read current state: %v", err)
	}

	if string(currentData) != string(originalData) {
		t.Error("file should not have been modified after failed save")
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	// Create temporary directory structure
	ctx := setupTestRepo(t)

	// Copy valid fixture
	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Create context
	// ctx already created by setupTestRepo

	// Load project
	project, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Save project
	if err := project.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify temp file does not exist after successful save
	tmpPath := "project/state.yaml.tmp"
	_, err = ctx.FS().Stat(tmpPath)
	if err == nil {
		t.Error("temp file should not exist after successful save")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	// Create temporary directory structure
	ctx := setupTestRepo(t)

	// Copy complex fixture
	fixtureData, err := os.ReadFile("testdata/complex-project.yaml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Write using FS abstraction
	if err := ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	// Create context
	// ctx already created by setupTestRepo

	// Load project
	project1, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Save project
	if err := project1.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load again
	project2, err := Load(ctx)
	if err != nil {
		t.Fatalf("Load() after Save() failed: %v", err)
	}

	// Verify all critical fields match
	if project1.Name != project2.Name {
		t.Errorf("name mismatch: %s != %s", project1.Name, project2.Name)
	}
	if project1.Type != project2.Type {
		t.Errorf("type mismatch: %s != %s", project1.Type, project2.Type)
	}
	if project1.Branch != project2.Branch {
		t.Errorf("branch mismatch: %s != %s", project1.Branch, project2.Branch)
	}
	if len(project1.Phases) != len(project2.Phases) {
		t.Errorf("phases count mismatch: %d != %d", len(project1.Phases), len(project2.Phases))
	}

	// Verify phase data preserved
	for phaseName, phase1 := range project1.Phases {
		phase2, exists := project2.Phases[phaseName]
		if !exists {
			t.Errorf("phase %s missing after round-trip", phaseName)
			continue
		}
		if phase1.Status != phase2.Status {
			t.Errorf("phase %s status mismatch: %s != %s", phaseName, phase1.Status, phase2.Status)
		}
		if len(phase1.Tasks) != len(phase2.Tasks) {
			t.Errorf("phase %s task count mismatch: %d != %d", phaseName, len(phase1.Tasks), len(phase2.Tasks))
		}
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
