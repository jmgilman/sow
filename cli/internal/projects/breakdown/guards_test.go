package breakdown

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// Helper functions for test setup

func newTestProject() *state.Project {
	return &state.Project{
		ProjectState: projschema.ProjectState{
			Name:       "test-project",
			Type:       "breakdown",
			Created_at: time.Now(),
			Updated_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}
}

func newTask(id, status string) projschema.TaskState {
	return projschema.TaskState{
		Id:             id,
		Name:           "Test Task " + id,
		Phase:          "breakdown",
		Status:         status,
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
		Iteration:      1,
		Assigned_agent: "implementer",
		Inputs:         []projschema.ArtifactState{},
		Outputs:        []projschema.ArtifactState{},
	}
}

func newTaskWithMetadata(id, status string, metadata map[string]any) projschema.TaskState {
	task := newTask(id, status)
	task.Metadata = metadata
	return task
}

func newArtifact(path string) projschema.ArtifactState {
	return projschema.ArtifactState{
		Type:       "work_unit",
		Path:       path,
		Approved:   false,
		Created_at: time.Now(),
	}
}

// Tests for hasApprovedDiscoveryDocument

func TestHasApprovedDiscoveryDocument_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()
	// No breakdown phase exists

	result := hasApprovedDiscoveryDocument(p)

	if result {
		t.Error("Expected false when breakdown phase is missing, got true")
	}
}

func TestHasApprovedDiscoveryDocument_NoDiscoveryArtifact(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "discovery",
		Enabled:    true,
		Created_at: time.Now(),
		Outputs:    []projschema.ArtifactState{},
	}

	result := hasApprovedDiscoveryDocument(p)

	if result {
		t.Error("Expected false when no discovery artifact exists, got true")
	}
}

func TestHasApprovedDiscoveryDocument_UnapprovedDiscoveryArtifact(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "discovery",
		Enabled:    true,
		Created_at: time.Now(),
		Outputs: []projschema.ArtifactState{
			{
				Type:       "discovery",
				Path:       "project/discovery/analysis.md",
				Approved:   false,
				Created_at: time.Now(),
			},
		},
	}

	result := hasApprovedDiscoveryDocument(p)

	if result {
		t.Error("Expected false when discovery artifact is not approved, got true")
	}
}

func TestHasApprovedDiscoveryDocument_ApprovedDiscoveryArtifact(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "discovery",
		Enabled:    true,
		Created_at: time.Now(),
		Outputs: []projschema.ArtifactState{
			{
				Type:       "discovery",
				Path:       "project/discovery/analysis.md",
				Approved:   true,
				Created_at: time.Now(),
			},
		},
	}

	result := hasApprovedDiscoveryDocument(p)

	if !result {
		t.Error("Expected true when discovery artifact is approved, got false")
	}
}

func TestHasApprovedDiscoveryDocument_MultipleArtifactsOneApproved(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "discovery",
		Enabled:    true,
		Created_at: time.Now(),
		Outputs: []projschema.ArtifactState{
			{
				Type:       "work_unit_spec",
				Path:       "project/work-units/001.md",
				Approved:   false,
				Created_at: time.Now(),
			},
			{
				Type:       "discovery",
				Path:       "project/discovery/analysis.md",
				Approved:   true,
				Created_at: time.Now(),
			},
		},
	}

	result := hasApprovedDiscoveryDocument(p)

	if !result {
		t.Error("Expected true when at least one discovery artifact is approved, got false")
	}
}

// Tests for allWorkUnitsApproved

func TestAllWorkUnitsApproved_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()
	// No breakdown phase exists

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when breakdown phase is missing, got true")
	}
}

func TestAllWorkUnitsApproved_NoTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when no tasks exist, got true")
	}
}

func TestAllWorkUnitsApproved_PendingTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when tasks are pending, got true")
	}
}

func TestAllWorkUnitsApproved_InProgressTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "in_progress"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when tasks are in_progress, got true")
	}
}

func TestAllWorkUnitsApproved_NeedsReviewTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "needs_review"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when tasks are needs_review, got true")
	}
}

func TestAllWorkUnitsApproved_AllTasksAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "abandoned"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsApproved(p)

	if result {
		t.Error("Expected false when all tasks are abandoned, got true")
	}
}

func TestAllWorkUnitsApproved_AllCompleted(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
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

	result := allWorkUnitsApproved(p)

	if !result {
		t.Error("Expected true when all tasks are completed, got false")
	}
}

func TestAllWorkUnitsApproved_MixOfCompletedAndAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
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

	result := allWorkUnitsApproved(p)

	if !result {
		t.Error("Expected true when mix includes completed tasks, got false")
	}
}

// Tests for dependenciesValid

func TestDependenciesValid_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()
	// No breakdown phase exists

	result := dependenciesValid(p)

	if result {
		t.Error("Expected false when breakdown phase is missing, got true")
	}
}

func TestDependenciesValid_NoTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true when no tasks exist (vacuously valid), got false")
	}
}

func TestDependenciesValid_NoDependencies(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
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

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true when no dependencies exist, got false")
	}
}

func TestDependenciesValid_ValidLinearDependencies(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"dependencies": []interface{}{"010"},
			}),
			newTaskWithMetadata("030", "completed", map[string]any{
				"dependencies": []interface{}{"020"},
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true for valid linear dependencies (010 → 020 → 030), got false")
	}
}

func TestDependenciesValid_ValidComplexDAG(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{}),
			newTaskWithMetadata("020", "completed", map[string]any{}),
			newTaskWithMetadata("030", "completed", map[string]any{
				"dependencies": []interface{}{"010", "020"},
			}),
			newTaskWithMetadata("040", "completed", map[string]any{
				"dependencies": []interface{}{"030"},
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true for valid complex DAG, got false")
	}
}

func TestDependenciesValid_CyclicDependency(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"dependencies": []interface{}{"030"},
			}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"dependencies": []interface{}{"010"},
			}),
			newTaskWithMetadata("030", "completed", map[string]any{
				"dependencies": []interface{}{"020"},
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if result {
		t.Error("Expected false for cyclic dependency (010 → 030 → 020 → 010), got true")
	}
}

func TestDependenciesValid_SelfReferencingDependency(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"dependencies": []interface{}{"010"},
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if result {
		t.Error("Expected false for self-referencing dependency, got true")
	}
}

func TestDependenciesValid_InvalidDependencyReference(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"dependencies": []interface{}{"999"}, // Non-existent task
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if result {
		t.Error("Expected false for invalid dependency reference, got true")
	}
}

func TestDependenciesValid_OnlyConsidersCompletedTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{}),
			newTaskWithMetadata("020", "abandoned", map[string]any{
				"dependencies": []interface{}{"010"},
			}),
			newTaskWithMetadata("030", "pending", map[string]any{
				"dependencies": []interface{}{"010"},
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true - only completed tasks are validated, got false")
	}
}

func TestDependenciesValid_NilMetadata(t *testing.T) {
	p := newTestProject()
	task := newTask("010", "completed")
	task.Metadata = nil

	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			task,
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true when metadata is nil (no dependencies), got false")
	}
}

func TestDependenciesValid_NonArrayDependencies(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"dependencies": "not-an-array", // Invalid type
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true when dependencies is not an array (ignored), got false")
	}
}

func TestDependenciesValid_NonStringDependencyValues(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"dependencies": []interface{}{123, true, "010"}, // Mixed types
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := dependenciesValid(p)

	if !result {
		t.Error("Expected true - non-string values are skipped, valid string dependency exists, got false")
	}
}

// Tests for allWorkUnitsPublished

func TestAllWorkUnitsPublished_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()
	// No breakdown phase exists

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when breakdown phase is missing, got true")
	}
}

func TestAllWorkUnitsPublished_NoCompletedTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when no completed tasks exist, got true")
	}
}

func TestAllWorkUnitsPublished_CompletedTaskWithoutPublishedMetadata(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "completed"), // No metadata
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when completed task has no metadata, got true")
	}
}

func TestAllWorkUnitsPublished_CompletedTaskWithoutPublishedField(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"other_field": "value",
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when completed task missing 'published' field, got true")
	}
}

func TestAllWorkUnitsPublished_CompletedTaskPublishedFalse(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": false,
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when completed task has published=false, got true")
	}
}

func TestAllWorkUnitsPublished_CompletedTaskPublishedNotBoolean(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": "yes", // Not a boolean
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when published is not a boolean, got true")
	}
}

func TestAllWorkUnitsPublished_AllCompletedTasksPublished(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"published": true,
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if !result {
		t.Error("Expected true when all completed tasks are published, got false")
	}
}

func TestAllWorkUnitsPublished_IgnoresAbandonedTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTask("020", "abandoned"), // No published field, but should be ignored
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if !result {
		t.Error("Expected true - abandoned tasks should be ignored, got false")
	}
}

func TestAllWorkUnitsPublished_MixedPublishedStatuses(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"published": false,
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allWorkUnitsPublished(p)

	if result {
		t.Error("Expected false when some completed tasks are not published, got true")
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
	p.Phases["breakdown"] = projschema.PhaseState{
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
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "completed"),
			newTask("030", "in_progress"),
			newTask("040", "needs_review"),
			newTask("050", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnresolvedTasks(p)

	if count != 3 {
		t.Errorf("Expected 3 unresolved tasks (pending, in_progress, needs_review), got %d", count)
	}
}

// Tests for countUnpublishedTasks

func TestCountUnpublishedTasks_MissingPhase(t *testing.T) {
	p := newTestProject()

	count := countUnpublishedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when phase is missing, got %d", count)
	}
}

func TestCountUnpublishedTasks_NoCompletedTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnpublishedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when no completed tasks exist, got %d", count)
	}
}

func TestCountUnpublishedTasks_AllPublished(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"published": true,
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnpublishedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when all completed tasks are published, got %d", count)
	}
}

func TestCountUnpublishedTasks_SomeUnpublished(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTaskWithMetadata("020", "completed", map[string]any{
				"published": false,
			}),
			newTask("030", "completed"), // No metadata
			newTaskWithMetadata("040", "completed", map[string]any{
				// No published field
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnpublishedTasks(p)

	if count != 3 {
		t.Errorf("Expected 3 unpublished tasks, got %d", count)
	}
}

func TestCountUnpublishedTasks_IgnoresNonCompletedTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"published": true,
			}),
			newTask("020", "pending"),   // Should not be counted
			newTask("030", "abandoned"), // Should not be counted
		},
		Outputs: []projschema.ArtifactState{},
	}

	count := countUnpublishedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 - only completed tasks should be counted, got %d", count)
	}
}

// Tests for validateTaskForCompletion

func TestValidateTaskForCompletion_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when breakdown phase is missing, got nil")
	}
	if err != nil && err.Error() != "breakdown phase not found" {
		t.Errorf("Expected 'breakdown phase not found', got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_TaskNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
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
	task := newTask("010", "in_progress")
	task.Metadata = nil

	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			task,
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

func TestValidateTaskForCompletion_EmptyMetadata(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{}),
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
	p.Phases["breakdown"] = projschema.PhaseState{
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

func TestValidateTaskForCompletion_ArtifactPathEmptyString(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": "",
			}),
		},
		Outputs: []projschema.ArtifactState{},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when artifact_path is empty string, got nil")
	}
	if err != nil && err.Error() != "task 010 has no artifact_path in metadata - link artifact to task before completing" {
		t.Errorf("Expected artifact_path error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_ArtifactNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": "project/work-unit-010.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/other-unit.md"),
		},
	}

	err := validateTaskForCompletion(p, "010")

	if err == nil {
		t.Error("Expected error when artifact not found, got nil")
	}
	if err != nil && err.Error() != "artifact not found at project/work-unit-010.md - add artifact before completing task" {
		t.Errorf("Expected artifact not found error, got '%s'", err.Error())
	}
}

func TestValidateTaskForCompletion_Success(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "in_progress", map[string]any{
				"artifact_path": "project/work-unit-010.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/work-unit-010.md"),
		},
	}

	err := validateTaskForCompletion(p, "010")

	if err != nil {
		t.Errorf("Expected nil error when validation passes, got %v", err)
	}
}

// Tests for autoApproveArtifact

func TestAutoApproveArtifact_MissingBreakdownPhase(t *testing.T) {
	p := newTestProject()

	err := autoApproveArtifact(p, "010")

	if err == nil {
		t.Error("Expected error when breakdown phase is missing, got nil")
	}
	if err != nil && err.Error() != "breakdown phase not found" {
		t.Errorf("Expected 'breakdown phase not found', got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_TaskNotFound(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
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
	p.Phases["breakdown"] = projschema.PhaseState{
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
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/work-unit-010.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/other-unit.md"),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err == nil {
		t.Error("Expected error when artifact not found, got nil")
	}
	if err != nil && err.Error() != "artifact not found at project/work-unit-010.md" {
		t.Errorf("Expected artifact not found error, got '%s'", err.Error())
	}
}

func TestAutoApproveArtifact_Success(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/work-unit-010.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/work-unit-010.md"),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err != nil {
		t.Errorf("Expected nil error when auto-approval succeeds, got %v", err)
	}

	// Verify artifact was approved
	phase := p.Phases["breakdown"]
	if len(phase.Outputs) != 1 {
		t.Fatal("Expected 1 artifact in outputs")
	}
	if !phase.Outputs[0].Approved {
		t.Error("Expected artifact to be approved, but it wasn't")
	}
}

func TestAutoApproveArtifact_UpdatesProjectState(t *testing.T) {
	p := newTestProject()
	p.Phases["breakdown"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTaskWithMetadata("010", "completed", map[string]any{
				"artifact_path": "project/work-unit-010.md",
			}),
		},
		Outputs: []projschema.ArtifactState{
			newArtifact("project/work-unit-010.md"),
			newArtifact("project/work-unit-020.md"),
		},
	}

	err := autoApproveArtifact(p, "010")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify only the correct artifact was approved
	phase := p.Phases["breakdown"]
	if len(phase.Outputs) != 2 {
		t.Fatal("Expected 2 artifacts in outputs")
	}

	approvedCount := 0
	for _, artifact := range phase.Outputs {
		switch artifact.Path {
		case "project/work-unit-010.md":
			if !artifact.Approved {
				t.Error("Expected work-unit-010.md to be approved")
			}
			approvedCount++
		case "project/work-unit-020.md":
			if artifact.Approved {
				t.Error("Expected work-unit-020.md to remain unapproved")
			}
		}
	}

	if approvedCount != 1 {
		t.Errorf("Expected exactly 1 approved artifact, got %d", approvedCount)
	}
}
