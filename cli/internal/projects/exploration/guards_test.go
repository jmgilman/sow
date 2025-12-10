package exploration

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// Helper to create a minimal project for testing.
func newTestProject() *state.Project {
	return &state.Project{
		ProjectState: projschema.ProjectState{
			Name:       "test-project",
			Type:       "exploration",
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
		Phase:          "exploration",
		Status:         status,
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
		Iteration:      1,
		Assigned_agent: "implementer",
		Inputs:         []projschema.ArtifactState{},
		Outputs:        []projschema.ArtifactState{},
	}
}

// Helper to create an artifact with given type and approval status.
func newArtifact(artifactType string, approved bool) projschema.ArtifactState {
	return projschema.ArtifactState{
		Type:       artifactType,
		Path:       "test/" + artifactType + ".md",
		Approved:   approved,
		Created_at: time.Now(),
	}
}

// Tests for allTasksResolved

func TestAllTasksResolved_MissingExplorationPhase(t *testing.T) {
	p := newTestProject()
	// No exploration phase exists

	result := allTasksResolved(p)

	if result {
		t.Error("Expected false when exploration phase is missing, got true")
	}
}

func TestAllTasksResolved_NoTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := allTasksResolved(p)

	if result {
		t.Error("Expected false when no tasks exist, got true")
	}
}

func TestAllTasksResolved_PendingTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "pending"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allTasksResolved(p)

	if result {
		t.Error("Expected false when tasks are pending, got true")
	}
}

func TestAllTasksResolved_InProgressTasks(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "in_progress"),
			newTask("020", "completed"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allTasksResolved(p)

	if result {
		t.Error("Expected false when tasks are in_progress, got true")
	}
}

func TestAllTasksResolved_AllCompleted(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
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

	result := allTasksResolved(p)

	if !result {
		t.Error("Expected true when all tasks are completed, got false")
	}
}

func TestAllTasksResolved_CompletedAndAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
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

	result := allTasksResolved(p)

	if !result {
		t.Error("Expected true when all tasks are completed or abandoned, got false")
	}
}

func TestAllTasksResolved_AllAbandoned(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "active",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks: []projschema.TaskState{
			newTask("010", "abandoned"),
			newTask("020", "abandoned"),
		},
		Outputs: []projschema.ArtifactState{},
	}

	result := allTasksResolved(p)

	if !result {
		t.Error("Expected true when all tasks are abandoned, got false")
	}
}

// Tests for allSummariesApproved

func TestAllSummariesApproved_MissingExplorationPhase(t *testing.T) {
	p := newTestProject()
	// No exploration phase exists

	result := allSummariesApproved(p)

	if result {
		t.Error("Expected false when exploration phase is missing, got true")
	}
}

func TestAllSummariesApproved_NoSummaryArtifacts(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs:    []projschema.ArtifactState{},
	}

	result := allSummariesApproved(p)

	if result {
		t.Error("Expected false when no summary artifacts exist, got true")
	}
}

func TestAllSummariesApproved_OnlyNonSummaryArtifacts(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("findings", true),
			newArtifact("notes", true),
		},
	}

	result := allSummariesApproved(p)

	if result {
		t.Error("Expected false when no summary artifacts exist (only other types), got true")
	}
}

func TestAllSummariesApproved_SummariesNotApproved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", false),
			newArtifact("summary", true),
		},
	}

	result := allSummariesApproved(p)

	if result {
		t.Error("Expected false when some summaries are not approved, got true")
	}
}

func TestAllSummariesApproved_AllSummariesApproved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", true),
			newArtifact("summary", true),
		},
	}

	result := allSummariesApproved(p)

	if !result {
		t.Error("Expected true when all summaries are approved, got false")
	}
}

func TestAllSummariesApproved_FiltersNonSummaryArtifacts(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", true),
			newArtifact("findings", false), // Not approved but not a summary
			newArtifact("summary", true),
		},
	}

	result := allSummariesApproved(p)

	if !result {
		t.Error("Expected true when all summary artifacts are approved (ignoring non-summary types), got false")
	}
}

func TestAllSummariesApproved_OneSummaryApproved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", true),
		},
	}

	result := allSummariesApproved(p)

	if !result {
		t.Error("Expected true when single summary is approved, got false")
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

// Tests for helper functions (if implemented)

func TestCountUnresolvedTasks_MissingPhase(t *testing.T) {
	p := newTestProject()

	count := countUnresolvedTasks(p)

	if count != 0 {
		t.Errorf("Expected 0 when phase is missing, got %d", count)
	}
}

func TestCountUnresolvedTasks_AllResolved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
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
	p.Phases["exploration"] = projschema.PhaseState{
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

func TestCountUnapprovedSummaries_MissingPhase(t *testing.T) {
	p := newTestProject()

	count := countUnapprovedSummaries(p)

	if count != 0 {
		t.Errorf("Expected 0 when phase is missing, got %d", count)
	}
}

func TestCountUnapprovedSummaries_AllApproved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", true),
			newArtifact("summary", true),
		},
	}

	count := countUnapprovedSummaries(p)

	if count != 0 {
		t.Errorf("Expected 0 when all summaries are approved, got %d", count)
	}
}

func TestCountUnapprovedSummaries_SomeUnapproved(t *testing.T) {
	p := newTestProject()
	p.Phases["exploration"] = projschema.PhaseState{
		Status:     "summarizing",
		Enabled:    true,
		Created_at: time.Now(),
		Tasks:      []projschema.TaskState{},
		Outputs: []projschema.ArtifactState{
			newArtifact("summary", false),
			newArtifact("summary", true),
			newArtifact("summary", false),
			newArtifact("findings", false), // Should not count
		},
	}

	count := countUnapprovedSummaries(p)

	if count != 2 {
		t.Errorf("Expected 2 unapproved summaries, got %d", count)
	}
}
