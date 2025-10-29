package standard

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

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

// TestPhaseImplementsInterface verifies that all phase implementations
// properly implement the domain.Phase interface.
func TestPhaseImplementsInterface(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()

	state := &projects.StandardProjectState{
		Project: struct {
			Type         string    `json:"type"`
			Name         string    `json:"name"`
			Branch       string    `json:"branch"`
			Description  string    `json:"description"`
			Github_issue *int64    `json:"github_issue,omitempty"` //nolint:revive // matches JSON schema
			Created_at   time.Time `json:"created_at"`              //nolint:revive // matches JSON schema
			Updated_at   time.Time `json:"updated_at"`              //nolint:revive // matches JSON schema
		}{
			Type:        "standard",
			Name:        "test",
			Branch:      "test",
			Description: "test",
			Created_at:  now,
			Updated_at:  now,
		},
		Phases: struct {
			Planning       phasesSchema.Phase `json:"planning"`
			Implementation phasesSchema.Phase `json:"implementation"`
			Review         phasesSchema.Phase `json:"review"`
			Finalize       phasesSchema.Phase `json:"finalize"`
		}{
			Planning:       phasesSchema.Phase{Status: "completed", Created_at: now, Enabled: true},
			Implementation: phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
			Review:         phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
			Finalize:       phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
		},
	}

	proj := New(state, ctx)

	phases := []domain.Phase{
		NewPlanningPhase(&state.Phases.Planning, proj, ctx),
		NewImplementationPhase(&state.Phases.Implementation, proj, ctx),
		NewReviewPhase(&state.Phases.Review, proj, ctx),
		NewFinalizePhase(&state.Phases.Finalize, proj, ctx),
	}

	for _, phase := range phases {
		// Verify basic interface methods work
		if phase.Name() == "" {
			t.Errorf("Phase %T returned empty name", phase)
		}
		if phase.Status() == "" {
			t.Errorf("Phase %T returned empty status", phase)
		}

		// Verify all phases can list artifacts (even if empty)
		artifacts := phase.ListArtifacts()
		if artifacts == nil {
			t.Errorf("Phase %T returned nil artifacts", phase)
		}

		// Verify all phases can list tasks (even if empty)
		tasks := phase.ListTasks()
		if tasks == nil {
			t.Errorf("Phase %T returned nil tasks", phase)
		}
	}
}

// TestPlanningPhaseArtifacts verifies planning phase properly supports artifacts.
func TestPlanningPhaseArtifacts(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()
	var err error

	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts:  []phasesSchema.Artifact{},
		Tasks:      []phasesSchema.Task{},
	}

	state := &projects.StandardProjectState{}
	proj := New(state, ctx)

	phase := NewPlanningPhase(phaseState, proj, ctx)

	// Should support artifacts
	err = phase.AddArtifact("task-list.md", domain.WithMetadata(map[string]interface{}{"type": "task_list"}))
	if err != nil {
		t.Fatalf("Planning phase should support artifacts, got error: %v", err)
	}

	// Should NOT support tasks
	_, err = phase.AddTask("test task", domain.WithDescription("test"))
	if !errors.Is(err, project.ErrNotSupported) {
		t.Errorf("Planning phase should not support tasks, expected ErrNotSupported, got: %v", err)
	}
}

// TestImplementationPhaseTasks verifies implementation phase properly supports tasks.
func TestImplementationPhaseTasks(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()

	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts:  []phasesSchema.Artifact{},
		Tasks:      []phasesSchema.Task{},
	}

	state := &projects.StandardProjectState{}
	proj := New(state, ctx)

	phase := NewImplementationPhase(phaseState, proj, ctx)

	// Should support tasks
	task, err := phase.AddTask("test task", domain.WithDescription("test"))
	if err != nil {
		t.Fatalf("Implementation phase should support tasks, got error: %v", err)
	}
	if task == nil {
		t.Fatal("Implementation phase AddTask returned nil task")
	}

	// Should NOT support artifacts
	err = phase.AddArtifact("test.md", domain.WithMetadata(map[string]interface{}{"type": "doc"}))
	if !errors.Is(err, project.ErrNotSupported) {
		t.Errorf("Implementation phase should not support artifacts, expected ErrNotSupported, got: %v", err)
	}
}

// TestReviewPhaseGuards verifies review phase metadata-based guards work correctly.
func TestReviewPhaseGuards(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()

	// Create phase with review artifacts
	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts: []phasesSchema.Artifact{
			{
				Path:       "review-001.md",
				Approved:   true,
				Created_at: now,
				Metadata: map[string]interface{}{
					"type":       "review",
					"assessment": "pass",
				},
			},
		},
		Tasks: []phasesSchema.Task{},
	}

	state := &projects.StandardProjectState{}
	proj := New(state, ctx)

	phase := NewReviewPhase(phaseState, proj, ctx)

	// Verify AllReviewsApproved checks metadata correctly
	if !phase.AllReviewsApproved() {
		t.Error("Expected AllReviewsApproved to return true with approved review artifact")
	}

	// Add unapproved review artifact
	_ = phase.AddArtifact("review-002.md", domain.WithMetadata(map[string]interface{}{
		"type":       "review",
		"assessment": "fail",
	}))

	if phase.AllReviewsApproved() {
		t.Error("Expected AllReviewsApproved to return false with unapproved review artifact")
	}
}

// TestPhaseMetadataOperations verifies Set/Get operations on phase metadata.
func TestPhaseMetadataOperations(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()
	var err error

	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Artifacts:  []phasesSchema.Artifact{},
		Tasks:      []phasesSchema.Task{},
		Metadata:   make(map[string]interface{}),
	}

	state := &projects.StandardProjectState{}
	proj := New(state, ctx)

	phase := NewReviewPhase(phaseState, proj, ctx)

	// Set metadata
	_, err = phase.Set("iteration", 2)
	if err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}

	// Get metadata
	val, err := phase.Get("iteration")
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if iteration, ok := val.(int); !ok || iteration != 2 {
		t.Errorf("Expected iteration 2, got %v", val)
	}

	// Get non-existent field should error
	_, err = phase.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent field")
	}
}

// TestStandardProjectInitialState verifies that StandardProject honors
// its declared initial state. This test validates the contract that
// Project.InitialState() returns the actual starting state.
func TestStandardProjectInitialState(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()

	// Create a fresh project state as would happen during initialization
	state := &projects.StandardProjectState{
		Statechart: struct {
			Current_state string `json:"current_state"` //nolint:revive // matches JSON schema
		}{
			Current_state: "PlanningActive", // StandardProject's initial state
		},
		Project: struct {
			Type         string    `json:"type"`
			Name         string    `json:"name"`
			Branch       string    `json:"branch"`
			Description  string    `json:"description"`
			Github_issue *int64    `json:"github_issue,omitempty"` //nolint:revive // matches JSON schema
			Created_at   time.Time `json:"created_at"`              //nolint:revive // matches JSON schema
			Updated_at   time.Time `json:"updated_at"`              //nolint:revive // matches JSON schema
		}{
			Type:        "standard",
			Name:        "test",
			Branch:      "test",
			Description: "test",
			Created_at:  now,
			Updated_at:  now,
		},
		Phases: struct {
			Planning       phasesSchema.Phase `json:"planning"`
			Implementation phasesSchema.Phase `json:"implementation"`
			Review         phasesSchema.Phase `json:"review"`
			Finalize       phasesSchema.Phase `json:"finalize"`
		}{
			Planning:       phasesSchema.Phase{Status: "in_progress", Created_at: now, Enabled: true},
			Implementation: phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
			Review:         phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
			Finalize:       phasesSchema.Phase{Status: "pending", Created_at: now, Enabled: true},
		},
	}

	// Create StandardProject
	proj := New(state, ctx)

	// Verify the contract: project's declared initial state matches actual state
	declaredInitial := proj.InitialState()
	actualState := proj.Machine().State()

	if actualState != declaredInitial {
		t.Errorf("StandardProject did not initialize to declared state: declared %s, actual %s",
			declaredInitial, actualState)
	}

	// Verify it's specifically PlanningActive for StandardProject
	if declaredInitial != "PlanningActive" {
		t.Errorf("StandardProject should declare PlanningActive as initial state, got %s", declaredInitial)
	}
}
