package templates

import (
	"embed"
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

//go:embed testdata/*.md
var testTemplatesFS embed.FS

// TestRender_BasicTemplate tests basic template rendering with project fields.
func TestRender_BasicTemplate(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:        "test-project",
			Type:        "standard",
			Branch:      "feat/test",
			Description: "Test description",
			Created_at:  time.Now(),
			Updated_at:  time.Now(),
			Phases:      make(map[string]projschema.PhaseState),
			Statechart: projschema.StatechartState{
				Current_state: "PlanningActive",
				Updated_at:    time.Now(),
			},
		},
	}

	output, err := Render(testTemplatesFS, "testdata/basic.md", project)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify output contains project fields
	if !strings.Contains(output, "test-project") {
		t.Errorf("Output missing project name: %s", output)
	}
	if !strings.Contains(output, "feat/test") {
		t.Errorf("Output missing branch: %s", output)
	}
	if !strings.Contains(output, "Test description") {
		t.Errorf("Output missing description: %s", output)
	}
}

// TestRender_PhaseHelper tests the phase lookup helper function.
func TestRender_PhaseHelper(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test-project",
			Type:   "standard",
			Branch: "feat/test",
			Phases: map[string]projschema.PhaseState{
				"planning": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: time.Now(),
					Outputs: []projschema.ArtifactState{
						{
							Type:       "task_list",
							Path:       "/path/to/task-list.md",
							Approved:   true,
							Created_at: time.Now(),
						},
					},
				},
			},
		},
	}

	output, err := Render(testTemplatesFS, "testdata/phase_helper.md", project)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify phase data is accessible
	if !strings.Contains(output, "Planning phase exists") {
		t.Errorf("Output missing phase check: %s", output)
	}
}

// TestRender_HasApprovedOutput tests the hasApprovedOutput helper.
func TestRender_HasApprovedOutput(t *testing.T) {
	tests := []struct {
		name           string
		project        *state.Project
		expectApproved bool
	}{
		{
			name: "approved output exists",
			project: &state.Project{
				ProjectState: projschema.ProjectState{
					Name:   "test",
					Type:   "standard",
					Branch: "test",
					Phases: map[string]projschema.PhaseState{
						"planning": {
							Outputs: []projschema.ArtifactState{
								{
									Type:     "task_list",
									Approved: true,
								},
							},
						},
					},
				},
			},
			expectApproved: true,
		},
		{
			name: "output not approved",
			project: &state.Project{
				ProjectState: projschema.ProjectState{
					Name:   "test",
					Type:   "standard",
					Branch: "test",
					Phases: map[string]projschema.PhaseState{
						"planning": {
							Outputs: []projschema.ArtifactState{
								{
									Type:     "task_list",
									Approved: false,
								},
							},
						},
					},
				},
			},
			expectApproved: false,
		},
		{
			name: "wrong output type",
			project: &state.Project{
				ProjectState: projschema.ProjectState{
					Name:   "test",
					Type:   "standard",
					Branch: "test",
					Phases: map[string]projschema.PhaseState{
						"planning": {
							Outputs: []projschema.ArtifactState{
								{
									Type:     "other",
									Approved: true,
								},
							},
						},
					},
				},
			},
			expectApproved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Render(testTemplatesFS, "testdata/approved_output.md", tt.project)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			containsApproved := strings.Contains(output, "Task list is approved")
			containsNotApproved := strings.Contains(output, "Task list not approved")

			if tt.expectApproved && !containsApproved {
				t.Errorf("Expected approved message, got: %s", output)
			}
			if !tt.expectApproved && !containsNotApproved {
				t.Errorf("Expected not approved message, got: %s", output)
			}
		})
	}
}

// TestRender_CountTasksByStatus tests the countTasksByStatus helper.
func TestRender_CountTasksByStatus(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test",
			Type:   "standard",
			Branch: "test",
			Phases: map[string]projschema.PhaseState{
				"implementation": {
					Tasks: []projschema.TaskState{
						{Id: "001", Status: "completed", Created_at: time.Now(), Updated_at: time.Now()},
						{Id: "002", Status: "completed", Created_at: time.Now(), Updated_at: time.Now()},
						{Id: "003", Status: "in_progress", Created_at: time.Now(), Updated_at: time.Now()},
						{Id: "004", Status: "pending", Created_at: time.Now(), Updated_at: time.Now()},
					},
				},
			},
		},
	}

	output, err := Render(testTemplatesFS, "testdata/task_counts.md", project)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify counts
	if !strings.Contains(output, "Completed: 2") {
		t.Errorf("Expected 2 completed tasks, got: %s", output)
	}
	if !strings.Contains(output, "In Progress: 1") {
		t.Errorf("Expected 1 in-progress task, got: %s", output)
	}
	if !strings.Contains(output, "Pending: 1") {
		t.Errorf("Expected 1 pending task, got: %s", output)
	}
}

// TestRender_JoinHelper tests basic template rendering (join helper tested in integration).
func TestRender_JoinHelper(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test",
			Type:   "standard",
			Branch: "test",
		},
	}

	output, err := Render(testTemplatesFS, "testdata/join.md", project)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Testing join") {
		t.Errorf("Expected test output, got: %s", output)
	}
}

// TestRender_TemplateNotFound tests error handling for missing templates.
func TestRender_TemplateNotFound(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test",
			Type:   "standard",
			Branch: "test",
		},
	}

	_, err := Render(testTemplatesFS, "testdata/nonexistent.md", project)
	if err == nil {
		t.Fatal("Expected error for nonexistent template")
	}

	if !strings.Contains(err.Error(), "failed to read template") {
		t.Errorf("Expected read error, got: %v", err)
	}
}

// TestRender_InvalidTemplate tests error handling for invalid template syntax.
func TestRender_InvalidTemplate(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test",
			Type:   "standard",
			Branch: "test",
		},
	}

	_, err := Render(testTemplatesFS, "testdata/invalid.md", project)
	if err == nil {
		t.Fatal("Expected error for invalid template")
	}

	if !strings.Contains(err.Error(), "failed to parse template") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

// TestRender_PhaseMetadata tests the phaseMetadata helper.
func TestRender_PhaseMetadata(t *testing.T) {
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:   "test",
			Type:   "standard",
			Branch: "test",
			Phases: map[string]projschema.PhaseState{
				"implementation": {
					Metadata: map[string]interface{}{
						"tasks_approved": true,
						"complexity":     "high",
					},
				},
			},
		},
	}

	output, err := Render(testTemplatesFS, "testdata/metadata.md", project)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Tasks approved: true") {
		t.Errorf("Expected tasks_approved metadata, got: %s", output)
	}
	if !strings.Contains(output, "Complexity: high") {
		t.Errorf("Expected complexity metadata, got: %s", output)
	}
}
