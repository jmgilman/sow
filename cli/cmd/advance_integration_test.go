package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/projects/standard"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sow"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestCLIWithStandardProject tests all four CLI modes against a real standard project.
//
// This is an end-to-end integration test that:
// - Creates a real project on disk with git repo
// - Advances through the full lifecycle using all CLI modes
// - Verifies state persistence across command executions
func TestCLIWithStandardProject(t *testing.T) {
	t.Run("auto-advance through full lifecycle", func(t *testing.T) {
		// Setup: Create standard project in ImplementationPlanning state
		ctx := setupTestRepoWithProject(t, "standard", "test-project")

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// Verify initial state
		if proj.Statechart.Current_state != string(standard.ImplementationPlanning) {
			t.Fatalf("expected initial state %s, got %s", standard.ImplementationPlanning, proj.Statechart.Current_state)
		}

		// Advance through states using auto mode
		states := []struct {
			current      string
			expected     string
			setupFunc    func(*testing.T, *sow.Context)
			validateFunc func(*testing.T, *state.Project)
		}{
			{
				current:  string(standard.ImplementationPlanning),
				expected: string(standard.ImplementationDraftPRCreation),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Approve planning
					setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)
				},
			},
			{
				current:  string(standard.ImplementationDraftPRCreation),
				expected: string(standard.ImplementationExecuting),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Set draft PR created
					setPhaseMetadata(t, ctx, "implementation", "draft_pr_created", true)
				},
			},
			{
				current:  string(standard.ImplementationExecuting),
				expected: string(standard.ReviewActive),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Add completed tasks
					addCompletedTasks(t, ctx, "implementation")
				},
			},
			{
				current:  string(standard.ReviewActive),
				expected: string(standard.FinalizeChecks),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Add approved review with pass assessment
					addApprovedReviewArtifact(t, ctx, "pass")
				},
				validateFunc: func(t *testing.T, p *state.Project) {
					// Verify review phase completed
					reviewPhase := p.Phases["review"]
					if reviewPhase.Status != "completed" {
						t.Errorf("review phase should be completed, got %s", reviewPhase.Status)
					}
				},
			},
			{
				current:  string(standard.FinalizeChecks),
				expected: string(standard.FinalizePRReady),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// No prerequisites for EventChecksDone
				},
			},
			{
				current:  string(standard.FinalizePRReady),
				expected: string(standard.FinalizePRChecks),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Add approved PR body
					addApprovedPRBody(t, ctx)
				},
			},
			{
				current:  string(standard.FinalizePRChecks),
				expected: string(standard.FinalizeCleanup),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Set PR checks passed
					setPhaseMetadata(t, ctx, "finalize", "pr_checks_passed", true)
				},
			},
			{
				current:  string(standard.FinalizeCleanup),
				expected: string(standard.NoProject),
				setupFunc: func(t *testing.T, ctx *sow.Context) {
					// Set project deleted
					setPhaseMetadata(t, ctx, "finalize", "project_deleted", true)
				},
				validateFunc: func(t *testing.T, p *state.Project) {
					// Verify finalize phase completed
					finalizePhase := p.Phases["finalize"]
					if finalizePhase.Status != "completed" {
						t.Errorf("finalize phase should be completed, got %s", finalizePhase.Status)
					}
				},
			},
		}

		for i, step := range states {
			// Reload project
			proj, err = state.Load(ctx)
			if err != nil {
				t.Fatalf("step %d: failed to reload project: %v", i, err)
			}

			// Verify current state
			if proj.Statechart.Current_state != step.current {
				t.Fatalf("step %d: expected current state %s, got %s", i, step.current, proj.Statechart.Current_state)
			}

			// Setup prerequisites
			if step.setupFunc != nil {
				step.setupFunc(t, ctx)
			}

			// Reload project after setup to pick up metadata changes
			proj, err = state.Load(ctx)
			if err != nil {
				t.Fatalf("step %d: failed to reload project after setup: %v", i, err)
			}

			// Execute auto-advance
			err = executeAutoTransition(proj, state.State(step.current))
			if err != nil {
				t.Fatalf("step %d: auto-advance from %s failed: %v", i, step.current, err)
			}

			// Reload project to verify persistence
			proj, err = state.Load(ctx)
			if err != nil {
				t.Fatalf("step %d: failed to reload after advance: %v", i, err)
			}

			// Verify state advanced
			if proj.Statechart.Current_state != step.expected {
				t.Errorf("step %d: expected state %s, got %s", i, step.expected, proj.Statechart.Current_state)
			}

			// Run custom validation if provided
			if step.validateFunc != nil {
				step.validateFunc(t, proj)
			}
		}
	})

	t.Run("list mode shows all transitions", func(t *testing.T) {
		// Setup: Create standard project in ImplementationPlanning state
		ctx := setupTestRepoWithProject(t, "standard", "test-project")

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		// Approve planning so transition is permitted
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)

		// Reload after metadata change
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload project: %v", err)
		}

		// Capture output from list mode
		output := captureOutput(func() {
			err := listAvailableTransitions(proj, state.State(standard.ImplementationPlanning))
			if err != nil {
				t.Fatalf("listAvailableTransitions failed: %v", err)
			}
		})

		// Verify output contains transition command
		if !strings.Contains(output, "sow advance") {
			t.Error("list mode output missing transition commands")
		}

		// Verify output contains target state
		if !strings.Contains(output, string(standard.ImplementationDraftPRCreation)) {
			t.Error("list mode output missing target state")
		}

		// Verify output contains description
		if !strings.Contains(output, "→") {
			t.Error("list mode output missing target state indicator (→)")
		}
	})

	t.Run("dry-run validates without side effects", func(t *testing.T) {
		// Setup: Create standard project in ImplementationPlanning state
		ctx := setupTestRepoWithProject(t, "standard", "test-project")

		// Approve planning
		setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)

		// Load project
		proj, err := state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to load project: %v", err)
		}

		initialState := proj.Statechart.Current_state

		// Dry-run a valid transition
		machine := proj.Machine()
		err = validateTransition(
			ctx,
			proj,
			machine,
			state.State(standard.ImplementationPlanning),
			state.Event(standard.EventPlanningComplete),
		)

		// Should succeed (no error)
		if err != nil {
			t.Errorf("dry-run should succeed for valid transition, got error: %v", err)
		}

		// Reload project to verify no state change
		proj, err = state.Load(ctx)
		if err != nil {
			t.Fatalf("failed to reload project: %v", err)
		}

		// Verify no side effects - state unchanged
		if proj.Statechart.Current_state != initialState {
			t.Errorf("dry-run modified state: was %s, now %s", initialState, proj.Statechart.Current_state)
		}
	})

	t.Run("explicit event with ReviewActive branching", func(t *testing.T) {
		t.Run("pass path", func(t *testing.T) {
			// Setup: Create standard project in ReviewActive state
			ctx := setupTestRepoWithProject(t, "standard", "test-project")
			advanceToReviewActive(t, ctx)

			// Add approved review with pass assessment
			addApprovedReviewArtifact(t, ctx, "pass")

			// Load project
			proj, err := state.Load(ctx)
			if err != nil {
				t.Fatalf("failed to load project: %v", err)
			}

			machine := proj.Machine()

			// Execute explicit event (review_pass)
			err = executeExplicitTransition(
				ctx,
				proj,
				machine,
				state.State(standard.ReviewActive),
				state.Event(standard.EventReviewPass),
			)
			if err != nil {
				t.Fatalf("explicit review_pass failed: %v", err)
			}

			// Reload project
			proj, err = state.Load(ctx)
			if err != nil {
				t.Fatalf("failed to reload project: %v", err)
			}

			// Verify state advanced to FinalizeChecks
			if proj.Statechart.Current_state != string(standard.FinalizeChecks) {
				t.Errorf("review_pass did not advance to FinalizeChecks, got %s", proj.Statechart.Current_state)
			}
		})

		t.Run("fail path", func(t *testing.T) {
			// Setup: Create standard project in ReviewActive state
			ctx := setupTestRepoWithProject(t, "standard", "test-project")
			advanceToReviewActive(t, ctx)

			// Add approved review with fail assessment
			addApprovedReviewArtifact(t, ctx, "fail")

			// Load project
			proj, err := state.Load(ctx)
			if err != nil {
				t.Fatalf("failed to load project: %v", err)
			}

			machine := proj.Machine()

			// Execute explicit event (review_fail)
			err = executeExplicitTransition(
				ctx,
				proj,
				machine,
				state.State(standard.ReviewActive),
				state.Event(standard.EventReviewFail),
			)
			if err != nil {
				t.Fatalf("explicit review_fail failed: %v", err)
			}

			// Reload project
			proj, err = state.Load(ctx)
			if err != nil {
				t.Fatalf("failed to reload project: %v", err)
			}

			// Verify state returned to ImplementationPlanning
			if proj.Statechart.Current_state != string(standard.ImplementationPlanning) {
				t.Errorf("review_fail did not return to ImplementationPlanning, got %s", proj.Statechart.Current_state)
			}

			// Verify implementation iteration incremented
			implPhase := proj.Phases["implementation"]
			if implPhase.Iteration != 1 {
				t.Errorf("implementation iteration should be 1, got %d", implPhase.Iteration)
			}
		})
	})
}

// Helper functions

// setupTestRepoWithProject creates a temp directory with git repo and sow project.
func setupTestRepoWithProject(t *testing.T, projectType, projectName string) *sow.Context {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()
	cmdCtx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(cmdCtx, "git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.email: %v", err)
	}

	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.name: %v", err)
	}

	// Disable GPG signing
	cmd = exec.CommandContext(cmdCtx, "git", "config", "commit.gpgsign", "false")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to disable gpgsign: %v", err)
	}

	// Create initial commit
	cmd = exec.CommandContext(cmdCtx, "git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	// Create .sow directory structure
	sowDir := filepath.Join(tmpDir, ".sow")
	projectDir := filepath.Join(sowDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create .sow/project directory: %v", err)
	}

	// Create sow context
	ctx, err := sow.NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Create project using Create (which calls Initialize)
	proj, err := state.Create(
		ctx,
		"test-branch",
		"Test project for integration tests",
		nil, // No initial inputs
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Save project to disk
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	return ctx
}

// setPhaseMetadata sets a metadata key on a phase and saves the project.
func setPhaseMetadata(t *testing.T, ctx *sow.Context, phaseName, key string, value interface{}) {
	t.Helper()

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get phase
	phase, exists := proj.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	// Set metadata
	if phase.Metadata == nil {
		phase.Metadata = make(map[string]interface{})
	}
	phase.Metadata[key] = value

	// Update project
	proj.Phases[phaseName] = phase

	// Save project
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project after setting metadata: %v", err)
	}
}

// addCompletedTasks adds some completed tasks to a phase.
func addCompletedTasks(t *testing.T, ctx *sow.Context, phaseName string) {
	t.Helper()

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get phase
	phase, exists := proj.Phases[phaseName]
	if !exists {
		t.Fatalf("phase %s not found", phaseName)
	}

	// Add completed tasks
	task1 := projschema.TaskState{
		Id:             "010",
		Name:           "Test Task 1",
		Phase:          phaseName,
		Status:         "completed",
		Iteration:      1,
		Assigned_agent: "implementer",
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
		Completed_at:   time.Now(),
		Inputs: []projschema.ArtifactState{
			{
				Type:       "task_description",
				Path:       ".sow/project/phases/implementation/tasks/010/description.md",
				Created_at: time.Now(),
				Approved:   true,
			},
		},
	}

	task2 := projschema.TaskState{
		Id:             "020",
		Name:           "Test Task 2",
		Phase:          phaseName,
		Status:         "completed",
		Iteration:      1,
		Assigned_agent: "implementer",
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
		Completed_at:   time.Now(),
		Inputs: []projschema.ArtifactState{
			{
				Type:       "task_description",
				Path:       ".sow/project/phases/implementation/tasks/020/description.md",
				Created_at: time.Now(),
				Approved:   true,
			},
		},
	}

	phase.Tasks = []projschema.TaskState{task1, task2}

	// Update project
	proj.Phases[phaseName] = phase

	// Save project
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project after adding tasks: %v", err)
	}
}

// addApprovedReviewArtifact adds an approved review output with assessment metadata.
func addApprovedReviewArtifact(t *testing.T, ctx *sow.Context, assessment string) {
	t.Helper()

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get review phase
	phase, exists := proj.Phases["review"]
	if !exists {
		t.Fatal("review phase not found")
	}

	// Add approved review artifact
	artifact := projschema.ArtifactState{
		Type:       "review",
		Path:       ".sow/project/phases/review/review.md",
		Created_at: time.Now(),
		Approved:   true,
		Metadata: map[string]interface{}{
			"assessment": assessment,
		},
	}

	phase.Outputs = append(phase.Outputs, artifact)

	// Update project
	proj.Phases["review"] = phase

	// Save project
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project after adding review: %v", err)
	}
}

// addApprovedPRBody adds an approved pr_body output artifact.
func addApprovedPRBody(t *testing.T, ctx *sow.Context) {
	t.Helper()

	// Load project
	proj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}

	// Get finalize phase
	phase, exists := proj.Phases["finalize"]
	if !exists {
		t.Fatal("finalize phase not found")
	}

	// Add approved PR body artifact
	artifact := projschema.ArtifactState{
		Type:       "pr_body",
		Path:       ".sow/project/phases/finalize/pr_body.md",
		Created_at: time.Now(),
		Approved:   true,
	}

	phase.Outputs = append(phase.Outputs, artifact)

	// Update project
	proj.Phases["finalize"] = phase

	// Save project
	if err := proj.Save(); err != nil {
		t.Fatalf("failed to save project after adding PR body: %v", err)
	}
}

// advanceToReviewActive advances a newly created project to ReviewActive state.
func advanceToReviewActive(t *testing.T, ctx *sow.Context) {
	t.Helper()

	// Planning complete
	setPhaseMetadata(t, ctx, "implementation", "planning_approved", true)
	proj, err := state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}
	if err := executeAutoTransition(proj, state.State(standard.ImplementationPlanning)); err != nil {
		t.Fatalf("failed to advance from ImplementationPlanning: %v", err)
	}

	// Draft PR created
	setPhaseMetadata(t, ctx, "implementation", "draft_pr_created", true)
	proj, err = state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}
	if err := executeAutoTransition(proj, state.State(standard.ImplementationDraftPRCreation)); err != nil {
		t.Fatalf("failed to advance from ImplementationDraftPRCreation: %v", err)
	}

	// All tasks complete
	addCompletedTasks(t, ctx, "implementation")
	proj, err = state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}
	if err := executeAutoTransition(proj, state.State(standard.ImplementationExecuting)); err != nil {
		t.Fatalf("failed to advance from ImplementationExecuting: %v", err)
	}

	// Verify we're in ReviewActive
	proj, err = state.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load project: %v", err)
	}
	if proj.Statechart.Current_state != string(standard.ReviewActive) {
		t.Fatalf("expected ReviewActive state, got %s", proj.Statechart.Current_state)
	}
}
