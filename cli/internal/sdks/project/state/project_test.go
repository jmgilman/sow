package state

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// Wrapper Type Tests

func TestProject_EmbeddedFields(t *testing.T) {
	// Arrange
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Name:       "test-project",
			Type:       "standard",
			Branch:     "feat/test",
			Created_at: now,
			Updated_at: now,
		},
	}

	// Act & Assert - verify embedded fields are accessible
	if p.Name != "test-project" {
		t.Errorf("Expected Name 'test-project', got: %s", p.Name)
	}
	if p.Type != "standard" {
		t.Errorf("Expected Type 'standard', got: %s", p.Type)
	}
	if p.Branch != "feat/test" {
		t.Errorf("Expected Branch 'feat/test', got: %s", p.Branch)
	}
}

func TestProject_RuntimeFieldsNotSerialized(t *testing.T) {
	// Arrange
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Name:       "test-project",
			Type:       "standard",
			Branch:     "feat/test",
			Created_at: now,
			Updated_at: now,
		},
		config:  &ProjectTypeConfig{},
		machine: &Machine{},
	}

	// Act - serialize to JSON
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Assert - runtime fields should not be in JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, exists := result["config"]; exists {
		t.Error("config field should not be serialized")
	}
	if _, exists := result["machine"]; exists {
		t.Error("machine field should not be serialized")
	}

	// Verify expected fields are present
	if _, exists := result["name"]; !exists {
		t.Error("name field should be serialized")
	}
	if _, exists := result["type"]; !exists {
		t.Error("type field should be serialized")
	}
}

func TestPhase_EmbeddedFields(t *testing.T) {
	// Arrange
	now := time.Now()
	phase := Phase{
		PhaseState: project.PhaseState{
			Status:     "in_progress",
			Enabled:    true,
			Created_at: now,
		},
	}

	// Act & Assert - verify embedded fields are accessible
	if phase.Status != "in_progress" {
		t.Errorf("Expected Status 'in_progress', got: %s", phase.Status)
	}
	if !phase.Enabled {
		t.Error("Expected Enabled to be true")
	}
}

func TestArtifact_EmbeddedFields(t *testing.T) {
	// Arrange
	now := time.Now()
	artifact := Artifact{
		ArtifactState: project.ArtifactState{
			Type:       "design_doc",
			Path:       "design.md",
			Approved:   true,
			Created_at: now,
		},
	}

	// Act & Assert - verify embedded fields are accessible
	if artifact.Type != "design_doc" {
		t.Errorf("Expected Type 'design_doc', got: %s", artifact.Type)
	}
	if artifact.Path != "design.md" {
		t.Errorf("Expected Path 'design.md', got: %s", artifact.Path)
	}
	if !artifact.Approved {
		t.Error("Expected Approved to be true")
	}
}

func TestTask_EmbeddedFields(t *testing.T) {
	// Arrange
	now := time.Now()
	task := Task{
		TaskState: project.TaskState{
			Id:             "010",
			Name:           "Implement feature",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      1,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
		},
	}

	// Act & Assert - verify embedded fields are accessible
	if task.Id != "010" {
		t.Errorf("Expected Id '010', got: %s", task.Id)
	}
	if task.Name != "Implement feature" {
		t.Errorf("Expected Name 'Implement feature', got: %s", task.Name)
	}
	if task.Status != "in_progress" {
		t.Errorf("Expected Status 'in_progress', got: %s", task.Status)
	}
	if task.Iteration != 1 {
		t.Errorf("Expected Iteration 1, got: %d", task.Iteration)
	}
}

func TestArtifact_Serialization(t *testing.T) {
	// Arrange
	now := time.Now()
	artifact := Artifact{
		ArtifactState: project.ArtifactState{
			Type:       "design_doc",
			Path:       "design.md",
			Approved:   true,
			Created_at: now,
			Metadata: map[string]any{
				"category": "architecture",
			},
		},
	}

	// Act - serialize and deserialize
	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var restored Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Assert - verify round-trip
	if restored.Type != artifact.Type {
		t.Errorf("Expected Type '%s', got: %s", artifact.Type, restored.Type)
	}
	if restored.Path != artifact.Path {
		t.Errorf("Expected Path '%s', got: %s", artifact.Path, restored.Path)
	}
	if restored.Approved != artifact.Approved {
		t.Errorf("Expected Approved %v, got: %v", artifact.Approved, restored.Approved)
	}
}

func TestTask_Serialization(t *testing.T) {
	// Arrange
	now := time.Now()
	task := Task{
		TaskState: project.TaskState{
			Id:             "010",
			Name:           "Implement feature",
			Phase:          "implementation",
			Status:         "in_progress",
			Iteration:      2,
			Assigned_agent: "implementer",
			Created_at:     now,
			Updated_at:     now,
			Metadata: map[string]any{
				"complexity": "high",
			},
		},
	}

	// Act - serialize and deserialize
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var restored Task
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Assert - verify round-trip
	if restored.Id != task.Id {
		t.Errorf("Expected Id '%s', got: %s", task.Id, restored.Id)
	}
	if restored.Name != task.Name {
		t.Errorf("Expected Name '%s', got: %s", task.Name, restored.Name)
	}
	if restored.Iteration != task.Iteration {
		t.Errorf("Expected Iteration %d, got: %d", task.Iteration, restored.Iteration)
	}
}

// Helper Methods Tests

// PhaseOutputApproved Tests

func TestPhaseOutputApproved_Found(t *testing.T) {
	// Arrange - Phase with approved "task_list" artifact
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:     "completed",
					Enabled:    true,
					Created_at: now,
					Outputs: []project.ArtifactState{
						{
							Type:       "task_list",
							Path:       "planning/tasks.md",
							Approved:   true,
							Created_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseOutputApproved("planning", "task_list")

	// Assert
	if !result {
		t.Error("Expected PhaseOutputApproved to return true for approved artifact")
	}
}

func TestPhaseOutputApproved_NotApproved(t *testing.T) {
	// Arrange - Phase with unapproved artifact
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Outputs: []project.ArtifactState{
						{
							Type:       "task_list",
							Path:       "planning/tasks.md",
							Approved:   false, // Not approved
							Created_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseOutputApproved("planning", "task_list")

	// Assert
	if result {
		t.Error("Expected PhaseOutputApproved to return false for unapproved artifact")
	}
}

func TestPhaseOutputApproved_PhaseNotFound(t *testing.T) {
	// Arrange - Non-existent phase name
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{},
		},
	}

	// Act
	result := p.PhaseOutputApproved("nonexistent", "task_list")

	// Assert - graceful handling
	if result {
		t.Error("Expected PhaseOutputApproved to return false for non-existent phase")
	}
}

func TestPhaseOutputApproved_ArtifactNotFound(t *testing.T) {
	// Arrange - Phase without artifact of requested type
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Outputs: []project.ArtifactState{
						{
							Type:       "design_doc",
							Path:       "planning/design.md",
							Approved:   true,
							Created_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseOutputApproved("planning", "task_list")

	// Assert
	if result {
		t.Error("Expected PhaseOutputApproved to return false for missing artifact type")
	}
}

// PhaseMetadataBool Tests

func TestPhaseMetadataBool_Found(t *testing.T) {
	// Arrange - Phase with metadata["tasks_approved"] = true
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Metadata: map[string]any{
						"tasks_approved": true,
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseMetadataBool("implementation", "tasks_approved")

	// Assert
	if !result {
		t.Error("Expected PhaseMetadataBool to return true for true boolean metadata")
	}
}

func TestPhaseMetadataBool_KeyNotFound(t *testing.T) {
	// Arrange - Metadata exists but key missing
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Metadata: map[string]any{
						"other_key": true,
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseMetadataBool("implementation", "tasks_approved")

	// Assert
	if result {
		t.Error("Expected PhaseMetadataBool to return false for missing key")
	}
}

func TestPhaseMetadataBool_NilMetadata(t *testing.T) {
	// Arrange - Phase with nil metadata map
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Metadata:   nil, // Nil metadata
				},
			},
		},
	}

	// Act
	result := p.PhaseMetadataBool("implementation", "tasks_approved")

	// Assert - no panic
	if result {
		t.Error("Expected PhaseMetadataBool to return false for nil metadata")
	}
}

func TestPhaseMetadataBool_WrongType(t *testing.T) {
	// Arrange - metadata["key"] = "string" (not bool)
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Metadata: map[string]any{
						"tasks_approved": "not a boolean",
					},
				},
			},
		},
	}

	// Act
	result := p.PhaseMetadataBool("implementation", "tasks_approved")

	// Assert
	if result {
		t.Error("Expected PhaseMetadataBool to return false for wrong type")
	}
}

// AllTasksComplete Tests

func TestAllTasksComplete_AllCompleted(t *testing.T) {
	// Arrange - All tasks have status="completed"
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Tasks: []project.TaskState{
						{
							Id:         "010",
							Name:       "Task 1",
							Phase:      "implementation",
							Status:     "completed",
							Created_at: now,
							Updated_at: now,
						},
						{
							Id:         "020",
							Name:       "Task 2",
							Phase:      "implementation",
							Status:     "completed",
							Created_at: now,
							Updated_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.AllTasksComplete()

	// Assert
	if !result {
		t.Error("Expected AllTasksComplete to return true when all tasks are completed")
	}
}

func TestAllTasksComplete_SomePending(t *testing.T) {
	// Arrange - At least one task status="pending"
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Tasks: []project.TaskState{
						{
							Id:         "010",
							Name:       "Task 1",
							Phase:      "implementation",
							Status:     "completed",
							Created_at: now,
							Updated_at: now,
						},
						{
							Id:         "020",
							Name:       "Task 2",
							Phase:      "implementation",
							Status:     "pending", // Not completed
							Created_at: now,
							Updated_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.AllTasksComplete()

	// Assert
	if result {
		t.Error("Expected AllTasksComplete to return false when some tasks are pending")
	}
}

func TestAllTasksComplete_NoTasks(t *testing.T) {
	// Arrange - No tasks in any phase
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Tasks:      []project.TaskState{}, // Empty tasks
				},
			},
		},
	}

	// Act
	result := p.AllTasksComplete()

	// Assert - vacuous truth
	if !result {
		t.Error("Expected AllTasksComplete to return true when no tasks exist")
	}
}

func TestAllTasksComplete_MultiplePhases(t *testing.T) {
	// Arrange - Tasks spread across multiple phases
	now := time.Now()
	p := Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:     "completed",
					Enabled:    true,
					Created_at: now,
					Tasks: []project.TaskState{
						{
							Id:         "010",
							Name:       "Plan task",
							Phase:      "planning",
							Status:     "completed",
							Created_at: now,
							Updated_at: now,
						},
					},
				},
				"implementation": {
					Status:     "in_progress",
					Enabled:    true,
					Created_at: now,
					Tasks: []project.TaskState{
						{
							Id:         "020",
							Name:       "Impl task 1",
							Phase:      "implementation",
							Status:     "completed",
							Created_at: now,
							Updated_at: now,
						},
						{
							Id:         "030",
							Name:       "Impl task 2",
							Phase:      "implementation",
							Status:     "in_progress", // Not completed
							Created_at: now,
							Updated_at: now,
						},
					},
				},
			},
		},
	}

	// Act
	result := p.AllTasksComplete()

	// Assert - checks ALL phases correctly
	if result {
		t.Error("Expected AllTasksComplete to check all phases and return false")
	}
}
