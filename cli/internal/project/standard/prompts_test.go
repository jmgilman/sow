package standard

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockContext creates a mock sow.Context for testing.
func mockContext(t *testing.T) *sow.Context {
	t.Helper()
	tmpDir := t.TempDir()
	cmdCtx := context.Background()

	// Initialize git repo
	cmd := exec.CommandContext(cmdCtx, "git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(cmdCtx, "git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	ctx, err := sow.NewContext(tmpDir)
	require.NoError(t, err)
	return ctx
}

// mockProjectState creates a minimal project state for testing.
func mockProjectState() *schemas.ProjectState {
	state := &projects.StandardProjectState{}

	// Set project fields
	state.Project.Name = "test-project"
	state.Project.Branch = "feat/test"
	state.Project.Description = "Test project description"

	// Set statechart
	state.Statechart.Current_state = string(PlanningActive)

	// Set phases
	state.Phases.Planning.Enabled = true
	state.Phases.Planning.Status = "in_progress"

	state.Phases.Implementation.Enabled = true
	state.Phases.Implementation.Status = "not_started"

	state.Phases.Review.Enabled = true
	state.Phases.Review.Status = "not_started"

	state.Phases.Finalize.Enabled = true
	state.Phases.Finalize.Status = "not_started"

	return (*schemas.ProjectState)(state)
}

func TestNewStandardPromptGenerator(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)

	assert.NotNil(t, generator)
	assert.NotNil(t, generator.components)
}

func TestGeneratePrompt_Routing(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	tests := []struct {
		name    string
		state   statechart.State
		wantErr bool
	}{
		{"PlanningActive", PlanningActive, false},
		{"ImplementationPlanning", ImplementationPlanning, false},
		{"ImplementationExecuting", ImplementationExecuting, false},
		{"ReviewActive", ReviewActive, false},
		{"FinalizeDocumentation", FinalizeDocumentation, false},
		{"FinalizeChecks", FinalizeChecks, false},
		{"FinalizeDelete", FinalizeDelete, false},
		{"UnknownState", statechart.State("UnknownState"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := generator.GeneratePrompt(tt.state, projectState)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, prompt)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, prompt)
			}
		})
	}
}

func TestGeneratePlanningPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("WithoutArtifacts", func(t *testing.T) {
		prompt, err := generator.generatePlanningPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "test-project")
		assert.Contains(t, prompt, "feat/test")
		assert.Contains(t, prompt, "Test project description")
		assert.NotContains(t, prompt, "Planning Artifacts")
	})

	t.Run("WithArtifacts", func(t *testing.T) {
		approvedFalse := false
		approvedTrue := true
		projectState.Phases.Planning.Artifacts = []phases.Artifact{
			{
				Path:     "task-list.md",
				Approved: &approvedFalse,
			},
			{
				Path:     "context.md",
				Approved: &approvedTrue,
			},
		}

		prompt, err := generator.generatePlanningPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Planning Artifacts")
		assert.Contains(t, prompt, "task-list.md (pending)")
		assert.Contains(t, prompt, "context.md (approved)")
	})
}

func TestGenerateImplementationPlanningPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("WithPlanningArtifacts", func(t *testing.T) {
		approvedTrue := true
		projectState.Phases.Planning.Artifacts = []phases.Artifact{
			{Path: "task-list.md", Approved: &approvedTrue},
			{Path: "requirements.md", Approved: &approvedTrue},
		}

		prompt, err := generator.generateImplementationPlanningPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Planning Context")
		assert.Contains(t, prompt, "task-list.md")
		assert.Contains(t, prompt, "requirements.md")
	})

	t.Run("WithoutPlanningArtifacts", func(t *testing.T) {
		projectState.Phases.Planning.Artifacts = []phases.Artifact{}

		prompt, err := generator.generateImplementationPlanningPrompt(projectState)

		require.NoError(t, err)
		assert.NotContains(t, prompt, "Planning Context")
	})
}

func TestGenerateImplementationExecutingPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("WithNoTasks", func(t *testing.T) {
		projectState.Phases.Implementation.Tasks = []phases.Task{}

		prompt, err := generator.generateImplementationExecutingPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Tasks (0 total)")
	})

	t.Run("WithPendingTasks", func(t *testing.T) {
		projectState.Phases.Implementation.Tasks = []phases.Task{
			{Id: "task-001", Name: "Task 1", Status: "pending"},
			{Id: "task-002", Name: "Task 2", Status: "in_progress"},
		}

		prompt, err := generator.generateImplementationExecutingPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Tasks (2 total)")
		assert.Contains(t, prompt, "1 in progress")
		assert.Contains(t, prompt, "1 pending")
		// Should not include commits when no tasks are completed
		assert.NotContains(t, prompt, "Recent Commits")
	})

	t.Run("WithCompletedTasks", func(t *testing.T) {
		projectState.Phases.Implementation.Tasks = []phases.Task{
			{Id: "task-001", Name: "Task 1", Status: "completed"},
			{Id: "task-002", Name: "Task 2", Status: "pending"},
		}

		prompt, err := generator.generateImplementationExecutingPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "1 completed")
		// Should include commits section when tasks are completed
		assert.Contains(t, prompt, "Recent Commits")
	})
}

func TestGenerateReviewPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("FirstIteration", func(t *testing.T) {
		// No metadata means iteration 1
		projectState.Phases.Review.Metadata = nil

		prompt, err := generator.generateReviewPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Review Iteration: 1")
		assert.NotContains(t, prompt, "Previous Review")
	})

	t.Run("SecondIteration", func(t *testing.T) {
		projectState.Phases.Review.Metadata = map[string]interface{}{
			"iteration": 2,
		}

		prompt, err := generator.generateReviewPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Review Iteration: 2")
		assert.Contains(t, prompt, "Previous Review")
	})

	t.Run("WithPreviousReviewArtifact", func(t *testing.T) {
		approvedFalse := false
		projectState.Phases.Review.Metadata = map[string]interface{}{
			"iteration": 2,
		}
		projectState.Phases.Review.Artifacts = []phases.Artifact{
			{
				Path:     "review-1.md",
				Approved: &approvedFalse,
				Metadata: map[string]interface{}{
					"type":       "review",
					"iteration":  1,
					"assessment": "fail",
				},
			},
		}

		prompt, err := generator.generateReviewPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Assessment: fail")
		assert.Contains(t, prompt, "review-1.md")
	})

	t.Run("WithTaskSummary", func(t *testing.T) {
		projectState.Phases.Implementation.Tasks = []phases.Task{
			{Id: "task-001", Name: "Task 1", Status: "completed"},
			{Id: "task-002", Name: "Task 2", Status: "completed"},
		}

		prompt, err := generator.generateReviewPrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Tasks (2 total)")
		assert.Contains(t, prompt, "2 completed")
	})
}

func TestGenerateFinalizeDocumentationPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	prompt, err := generator.generateFinalizeDocumentationPrompt(projectState)

	require.NoError(t, err)
	assert.Contains(t, prompt, "test-project")
	assert.NotEmpty(t, prompt)
}

func TestGenerateFinalizeChecksPrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	prompt, err := generator.generateFinalizeChecksPrompt(projectState)

	require.NoError(t, err)
	assert.Contains(t, prompt, "test-project")
	assert.NotEmpty(t, prompt)
}

func TestGenerateFinalizeDeletePrompt(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("WithoutPR", func(t *testing.T) {
		projectState.Phases.Finalize.Metadata = nil

		prompt, err := generator.generateFinalizeDeletePrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "test-project")
		assert.NotContains(t, prompt, "Pull Request")
	})

	t.Run("WithPR", func(t *testing.T) {
		projectState.Phases.Finalize.Metadata = map[string]interface{}{
			"pr_url": "https://github.com/org/repo/pull/123",
		}

		prompt, err := generator.generateFinalizeDeletePrompt(projectState)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Pull Request")
		assert.Contains(t, prompt, "https://github.com/org/repo/pull/123")
	})
}

func TestAllPromptsContainProjectHeader(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	states := []statechart.State{
		PlanningActive,
		ImplementationPlanning,
		ImplementationExecuting,
		ReviewActive,
		FinalizeDocumentation,
		FinalizeChecks,
		FinalizeDelete,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			prompt, err := generator.GeneratePrompt(state, projectState)

			require.NoError(t, err)
			assert.Contains(t, prompt, "test-project", "prompt should contain project name")
			assert.Contains(t, prompt, "feat/test", "prompt should contain branch")
		})
	}
}

func TestPromptErrorHandling(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	t.Run("InvalidState", func(t *testing.T) {
		_, err := generator.GeneratePrompt(statechart.State("InvalidState"), projectState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown state")
	})

	t.Run("NilProjectState", func(t *testing.T) {
		// This should panic or handle gracefully
		// For now, we expect it to fail when accessing fields
		defer func() {
			if r := recover(); r != nil {
				t.Log("Recovered from panic as expected with nil project state")
			}
		}()

		_, _ = generator.GeneratePrompt(PlanningActive, nil)
	})
}

func TestPromptOutputFormat(t *testing.T) {
	ctx := mockContext(t)
	generator := NewStandardPromptGenerator(ctx)
	projectState := mockProjectState()

	states := []statechart.State{
		PlanningActive,
		ImplementationPlanning,
		ImplementationExecuting,
		ReviewActive,
		FinalizeDocumentation,
		FinalizeChecks,
		FinalizeDelete,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			prompt, err := generator.GeneratePrompt(state, projectState)

			require.NoError(t, err)

			// Check for markdown formatting
			assert.True(t, strings.Contains(prompt, "#"), "prompt should contain markdown headers")

			// Check it's not empty or just whitespace
			assert.NotEmpty(t, strings.TrimSpace(prompt))

			// Check reasonable length (at least 50 characters for any prompt)
			assert.Greater(t, len(prompt), 50, "prompt should have substantial content")
		})
	}
}
