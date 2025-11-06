package design

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// Helper to create a minimal project for testing.
func newTestProject() *state.Project {
	return &state.Project{
		ProjectState: projschema.ProjectState{
			Name:       "test-project",
			Type:       "design",
			Created_at: time.Now(),
			Updated_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}
}

// Helper to create a task with given status.
func newTask(id, status string) projschema.TaskState {
	return projschema.TaskState{
		Id:             id,
		Name:           "Test Task " + id,
		Phase:          "design",
		Status:         status,
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
		Iteration:      1,
		Assigned_agent: "implementer",
		Inputs:         []projschema.ArtifactState{},
		Outputs:        []projschema.ArtifactState{},
	}
}

// Helper to create a task with metadata.
//nolint:unparam // id parameter kept for consistency with test helper pattern
func newTaskWithMetadata(id, status string, metadata map[string]any) projschema.TaskState {
	task := newTask(id, status)
	task.Metadata = metadata
	return task
}

// Helper to create an artifact.
//nolint:unparam // approved parameter kept for consistency with test helper pattern
func newArtifact(path string, approved bool) projschema.ArtifactState {
	return projschema.ArtifactState{
		Type:       "document",
		Path:       path,
		Approved:   approved,
		Created_at: time.Now(),
	}
}

// Tests for allDocumentsApproved

func TestAllDocumentsApproved_MissingDesignPhase(t *testing.T) {
	p := newTestProject()
	// No design phase exists

	result := allDocumentsApproved(p)

	if result {
		t.Error("Expected false when design phase is missing, got true")
	}
}

func TestAllDocumentsApproved_NoTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if result {
		t.Error("Expected false when no tasks exist, got true")
	}
}

func TestAllDocumentsApproved_PendingTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if result {
		t.Error("Expected false when tasks are pending, got true")
	}
}

func TestAllDocumentsApproved_InProgressTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "in_progress"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if result {
		t.Error("Expected false when tasks are in_progress, got true")
	}
}

func TestAllDocumentsApproved_AllTasksAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "abandoned"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if result {
		t.Error("Expected false when all tasks are abandoned, got true")
	}
}

func TestAllDocumentsApproved_AllCompleted(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "completed"),
			newTask("020", "completed"),
			newTask("030", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if !result {
		t.Error("Expected true when all tasks are completed, got false")
	}
}

func TestAllDocumentsApproved_MixOfCompletedAndAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "completed"),
			newTask("020", "abandoned"),
			newTask("030", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allDocumentsApproved(p)

	if !result {
		t.Error("Expected true when mix includes completed tasks, got false")
	}
}

// Tests for allFinalizationTasksComplete

func TestAllFinalizationTasksComplete_MissingFinalizationPhase(t *testing.T) {
	p := newTestProject()
	// No finalization phase exists

	result := allFinalizationTasksComplete(p)

	if result {
		t.Error("Expected false when finalization phase is missing, got true")
	}
}

func TestAllFinalizationTasksComplete_NoTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := allFinalizationTasksComplete(p)

	if result {
		t.Error("Expected false when no tasks exist, got true")
	}
}

func TestAllFinalizationTasksComplete_PendingTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			{
				Id:             "010",
				Name:           "Finalization Task 1",
				Phase:          "finalization",
				Status:         "pending",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
			{
				Id:             "020",
				Name:           "Finalization Task 2",
				Phase:          "finalization",
				Status:         "completed",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allFinalizationTasksComplete(p)

	if result {
		t.Error("Expected false when tasks are pending, got true")
	}
}

func TestAllFinalizationTasksComplete_InProgressTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			{
				Id:             "010",
				Name:           "Finalization Task 1",
				Phase:          "finalization",
				Status:         "in_progress",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allFinalizationTasksComplete(p)

	if result {
		t.Error("Expected false when tasks are in_progress, got true")
	}
}

func TestAllFinalizationTasksComplete_AbandonedTasksNotAllowed(t *testing.T) {
	p := newTestProject()
	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			{
				Id:             "010",
				Name:           "Finalization Task 1",
				Phase:          "finalization",
				Status:         "completed",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
			{
				Id:             "020",
				Name:           "Finalization Task 2",
				Phase:          "finalization",
				Status:         "abandoned",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allFinalizationTasksComplete(p)

	if result {
		t.Error("Expected false when finalization tasks are abandoned (must be completed), got true")
	}
}

func TestAllFinalizationTasksComplete_AllCompleted(t *testing.T) {
	p := newTestProject()
	p.Phases["finalization"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			{
				Id:             "010",
				Name:           "Finalization Task 1",
				Phase:          "finalization",
				Status:         "completed",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
			{
				Id:             "020",
				Name:           "Finalization Task 2",
				Phase:          "finalization",
				Status:         "completed",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []projschema.ArtifactState{},
				Outputs:        []projschema.ArtifactState{},
			},
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allFinalizationTasksComplete(p)

	if !result {
		t.Error("Expected true when all tasks are completed, got false")
	}
}

// Tests for countUnresolvedTasks

func TestCountUnresolvedTasks_MissingPhase(t *testing.T) {
	p := newTestProject()

	count := countUnresolvedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when phase is missing, got %d", count)
	}
}

func TestCountUnresolvedTasks_AllResolved(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "completed"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnresolvedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when all tasks are resolved, got %d", count)
	}
}

func TestCountUnresolvedTasks_SomeUnresolved(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "completed"),
			newTask("030", "in_progress"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnresolvedTasks(p)

	if count != 2 {
		t.Errorf("Expected 2 unresolved tasks, got %d", count)
	}
}

// Tests for validateTaskForCompletion

func TestValidateTaskForCompletion_MissingDesignPhase(t *testing.T) {
	p := newTestProject()

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when design phase is missing, got nil")
	}
	if err != nil && err.Error() != "design phase not found" {
		t.Errorf("Expected 'design phase not found', got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_TaskNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "in_progress"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := validateTaskForCompletion(p, "020")

	if err == nil {
		t.Error("Expected error when task not found, got nil")
	}
	if err != nil && err.Error() != "task 020 not found" {
		t.Errorf("Expected 'task 020 not found', got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_NoMetadata(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "in_progress"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when task has no metadata, got nil")
	}
	if err != nil && err.Error() != "task 010 has no metadata - set artifact_path before completing" {
		t.Errorf("Expected metadata error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_MissingArtifactPath(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"document_type": "design",
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when artifact_path missing, got nil")
	}
	if err != nil && err.Error() != "task 010 has no artifact_path in metadata - link artifact to task before completing" {
		t.Errorf("Expected artifact_path error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_ArtifactPathNotString(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": 123, // Not a string
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when artifact_path is not a string, got nil")
	}
	if err != nil && err.Error() != "task 010 has no artifact_path in metadata - link artifact to task before completing" {
		t.Errorf("Expected artifact_path error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_ArtifactNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": "project/auth-design.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/other-design.md", false),
		},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when artifact not found, got nil")
	}
	if err != nil && err.Error() != "artifact not found at project/auth-design.md - add artifact before completing task" {
		t.Errorf("Expected artifact not found error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_Success(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": "project/auth-design.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/auth-design.md", false),
		},
	}

	err := validateTaskForCompletion(p, "010")

	if err != nil {
		t.Errorf("Expected nil error when validation passes, got %v", err)
	}
}

// Tests for autoApproveArtifact

func TestAutoApproveArtifact_MissingDesignPhase(t *testing.T) {
	p := newTestProject()

	err := autoApproveArtifact(p, "010")

	if err == nil {
		t.Error("Expected error when design phase is missing, got nil")
	}
	if err != nil && err.Error() != "design phase not found" {
		t.Errorf("Expected 'design phase not found', got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_TaskNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := autoApproveArtifact(p, "020")

	if err == nil {
		t.Error("Expected error when task not found, got nil")
	}
	if err != nil && err.Error() != "task 020 not found" {
		t.Errorf("Expected 'task 020 not found', got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_InvalidArtifactPath(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": 123, // Not a string
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := autoApproveArtifact(p, "010")

	if err == nil {
		t.Error("Expected error when artifact_path invalid, got nil")
	}
	if err != nil && err.Error() != "task 010 has invalid artifact_path in metadata" {
		t.Errorf("Expected invalid artifact_path error, got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_ArtifactNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/auth-design.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/other-design.md", false),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err == nil {
		t.Error("Expected error when artifact not found, got nil")
	}
	if err != nil && err.Error() != "artifact not found at project/auth-design.md" {
		t.Errorf("Expected artifact not found error, got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_Success(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/auth-design.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/auth-design.md", false),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err != nil {
		t.Errorf("Expected nil error when auto-approval succeeds, got %v", err)
	}

	// Verify artifact was approved
	phase := p.Phases["design"]
	if len(phase.Outputs) != 1 {
		t.Fatal("Expected 1 artifact in outputs")
	}
	if !phase.Outputs[0].Approved {
		t.Error("Expected artifact to be approved, but it wasn't")
	}
}

func TestAutoApproveArtifact_UpdatesProjectState(t *testing.T) {
	p := newTestProject()
	p.Phases["design"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/auth-design.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/auth-design.md", false),
			newArtifact("project/other-design.md", false),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify only the correct artifact was approved
	phase := p.Phases["design"]
	if len(phase.Outputs) != 2 {
		t.Fatal("Expected 2 artifacts in outputs")
	}

	approvedCount := 0
	for _, artifact := range phase.Outputs {
		switch artifact.Path {
		case "project/auth-design.md":
			if !artifact.Approved {
				t.Error("Expected auth-design.md to be approved")
			}
			approvedCount++
		case "project/other-design.md":
			if artifact.Approved {
				t.Error("Expected other-design.md to remain unapproved")
			}
		}
	}

	if approvedCount != 1 {
		t.Errorf("Expected exactly 1 approved artifact, got %d", approvedCount)
	}
}
