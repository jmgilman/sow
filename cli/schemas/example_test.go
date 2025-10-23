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
	// Create a StandardProjectState instance using generated types
	now := time.Now()
	state := projects.StandardProjectState{
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
			Created_at:   now,
			Updated_at:   now,
		},
		Phases: struct {
			Discovery      phases.Phase `json:"discovery"`
			Design         phases.Phase `json:"design"`
			Implementation phases.Phase `json:"implementation"`
			Review         phases.Phase `json:"review"`
			Finalize       phases.Phase `json:"finalize"`
		}{
			// All phases now use the same generic Phase schema
			Implementation: phases.Phase{
				Status:     "in_progress",
				Created_at: now,
				Enabled:    true,
				Artifacts:  []phases.Artifact{}, // Generic artifacts field
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

// TestGenericPhaseWithMetadata demonstrates using the generic Phase schema
// with metadata for phase-specific data (like review assessments).
func TestGenericPhaseWithMetadata(t *testing.T) {
	now := time.Now()

	// Create a review phase with artifacts containing metadata
	reviewPhase := phases.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
		Tasks:      []phases.Task{}, // Review doesn't use tasks
		Artifacts: []phases.Artifact{
			{
				Path:       "project/phases/review/reports/001.md",
				Approved:   true,
				Created_at: now,
				Metadata: map[string]interface{}{
					"type":       "review",
					"assessment": "pass",
				},
			},
			{
				Path:       "project/phases/review/reports/002.md",
				Approved:   false,
				Created_at: now,
				Metadata: map[string]interface{}{
					"type":       "review",
					"assessment": "fail",
				},
			},
		},
		Metadata: map[string]interface{}{
			"iteration": 2,
		},
	}

	// Verify we can access phase data
	if len(reviewPhase.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(reviewPhase.Artifacts))
	}

	// Verify we can filter artifacts by metadata
	var reviewArtifacts []phases.Artifact
	for _, artifact := range reviewPhase.Artifacts {
		if artifactType, ok := artifact.Metadata["type"].(string); ok && artifactType == "review" {
			reviewArtifacts = append(reviewArtifacts, artifact)
		}
	}

	if len(reviewArtifacts) != 2 {
		t.Errorf("Expected 2 review artifacts, got %d", len(reviewArtifacts))
	}

	// Verify we can check assessment from metadata
	firstArtifact := reviewPhase.Artifacts[0]
	if assessment, ok := firstArtifact.Metadata["assessment"].(string); ok {
		if assessment != "pass" {
			t.Errorf("Expected assessment 'pass', got '%s'", assessment)
		}
	} else {
		t.Error("Expected assessment metadata to be present")
	}

	// Verify phase metadata
	if iteration, ok := reviewPhase.Metadata["iteration"].(int); ok {
		if iteration != 2 {
			t.Errorf("Expected iteration 2, got %d", iteration)
		}
	} else {
		t.Error("Expected iteration metadata to be present")
	}
}
