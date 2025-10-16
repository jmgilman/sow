//nolint:revive // Generated types use snake_case to match CUE schemas
package schemas_test

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
)

// Example demonstrating usage of generated types.
func TestGeneratedTypes(t *testing.T) {
	// Create a ProjectState instance using generated types
	state := schemas.ProjectState{
		Project: struct {
			Name        string    `json:"name"`
			Branch      string    `json:"branch"`
			Description string    `json:"description"`
			Created_at  time.Time `json:"created_at"`
			Updated_at  time.Time `json:"updated_at"`
		}{
			Name:        "my-feature",
			Branch:      "feat/my-feature",
			Description: "Implement new feature",
			Created_at:  time.Now(),
			Updated_at:  time.Now(),
		},
		Phases: struct {
			Discovery      schemas.DiscoveryPhase      `json:"discovery"`
			Design         schemas.DesignPhase         `json:"design"`
			Implementation schemas.ImplementationPhase `json:"implementation"`
			Review         schemas.ReviewPhase         `json:"review"`
			Finalize       schemas.FinalizePhase       `json:"finalize"`
		}{
			Implementation: schemas.ImplementationPhase{
				Status:     "in_progress",
				Created_at: time.Now(),
				Enabled:    true,
				Tasks: []schemas.Task{
					{
						Id:       "010",
						Name:     "Implement core logic",
						Status:   "completed",
						Parallel: false,
					},
					{
						Id:       "020",
						Name:     "Add tests",
						Status:   "in_progress",
						Parallel: false,
					},
				},
			},
		},
	}

	// Verify we can access the data
	if state.Project.Name != "my-feature" {
		t.Errorf("Expected name 'my-feature', got '%s'", state.Project.Name)
	}

	if len(state.Phases.Implementation.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(state.Phases.Implementation.Tasks))
	}

	// Test TaskState
	taskState := schemas.TaskState{
		Task: struct {
			Id             string              `json:"id"`
			Name           string              `json:"name"`
			Phase          string              `json:"phase"`
			Status         string              `json:"status"`
			Created_at     time.Time           `json:"created_at"`
			Started_at     any                 `json:"started_at"`
			Updated_at     time.Time           `json:"updated_at"`
			Completed_at   any                 `json:"completed_at"`
			Iteration      int64               `json:"iteration"`
			References     []string            `json:"references"`
			Feedback       []schemas.Feedback  `json:"feedback"`
			Files_modified []string            `json:"files_modified"`
		}{
			Id:         "010",
			Name:       "Test task",
			Phase:      "implementation",
			Status:     "in_progress",
			Created_at: time.Now(),
			Updated_at: time.Now(),
			Iteration:  1,
			References: []string{".sow/knowledge/overview.md"},
		},
	}

	if taskState.Task.Iteration != 1 {
		t.Errorf("Expected iteration 1, got %d", taskState.Task.Iteration)
	}

	// Test RefsCommittedIndex
	index := schemas.RefsCommittedIndex{
		Version: "1.0.0",
		Refs: []schemas.Ref{
			{
				Id:          "style-guide",
				Source:      "git+https://github.com/org/style-guide",
				Semantic:    "knowledge",
				Link:        "go-style",
				Tags:        []string{"go", "style"},
				Description: "Go style guide",
				Summary:     "Comprehensive Go coding standards and best practices for the team.",
				Config: schemas.RefConfig{
					Branch: "main",
					Path:   "docs/go",
				},
			},
		},
	}

	if len(index.Refs) != 1 {
		t.Errorf("Expected 1 ref, got %d", len(index.Refs))
	}
}
