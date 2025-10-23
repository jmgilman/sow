//nolint:revive // Generated types use snake_case to match CUE schemas
package schemas_test

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// Example demonstrating usage of generated types.
func TestGeneratedTypes(t *testing.T) {
	// Create a ProjectState instance using generated types
	state := projects.ProjectState{
		Project: struct {
			Type         string    `json:"type"`
			Name         string    `json:"name"`
			Branch       string    `json:"branch"`
			Description  string    `json:"description"`
			Github_issue *int64    `json:"github_issue,omitempty"`
			Created_at   time.Time `json:"created_at"`
			Updated_at   time.Time `json:"updated_at"`
		}{
			Type:         "standard",
			Name:         "my-feature",
			Branch:       "feat/my-feature",
			Description:  "Implement new feature",
			Github_issue: nil,
			Created_at:   time.Now(),
			Updated_at:   time.Now(),
		},
		Phases: struct {
			Discovery      phases.DiscoveryPhase      `json:"discovery"`
			Design         phases.DesignPhase         `json:"design"`
			Implementation phases.ImplementationPhase `json:"implementation"`
			Review         phases.ReviewPhase         `json:"review"`
			Finalize       phases.FinalizePhase       `json:"finalize"`
		}{
			Implementation: phases.ImplementationPhase{
				Status:     "in_progress",
				Created_at: time.Now(),
				Enabled:    true,
				Tasks: []phases.Task{
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
			Started_at     *time.Time          `json:"started_at,omitempty"`
			Updated_at     time.Time           `json:"updated_at"`
			Completed_at   *time.Time          `json:"completed_at,omitempty"`
			Iteration      int64               `json:"iteration"`
			Assigned_agent string              `json:"assigned_agent"`
			References     []string            `json:"references"`
			Feedback       []schemas.Feedback  `json:"feedback"`
			Files_modified []string            `json:"files_modified"`
		}{
			Id:             "010",
			Name:           "Test task",
			Phase:          "implementation",
			Status:         "in_progress",
			Created_at:     time.Now(),
			Updated_at:     time.Now(),
			Iteration:      1,
			Assigned_agent: "implementer",
			References:     []string{".sow/knowledge/overview.md"},
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
