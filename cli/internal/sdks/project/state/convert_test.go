package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConvertArtifacts_AllFieldsMapped tests that all artifact fields are correctly
// mapped from CUE-generated ArtifactState to wrapper Artifact type.
func TestConvertArtifacts_AllFieldsMapped(t *testing.T) {
	now := time.Now()
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	stateArtifacts := []project.ArtifactState{
		{
			Type:       "design_doc",
			Path:       "phases/design/design.md",
			Approved:   true,
			Created_at: now,
			Metadata:   metadata,
		},
		{
			Type:       "task_list",
			Path:       "phases/planning/tasks.md",
			Approved:   false,
			Created_at: now.Add(-time.Hour),
			Metadata:   nil, // Test nil metadata
		},
	}

	coll := convertArtifacts(stateArtifacts)

	require.Len(t, coll, 2)

	// First artifact - all fields populated
	art1, err := coll.Get(0)
	require.NoError(t, err)
	assert.Equal(t, "design_doc", art1.Type)
	assert.Equal(t, "phases/design/design.md", art1.Path)
	assert.True(t, art1.Approved)
	assert.Equal(t, now, art1.Created_at)
	assert.Equal(t, metadata, art1.Metadata)

	// Second artifact - nil metadata
	art2, err := coll.Get(1)
	require.NoError(t, err)
	assert.Equal(t, "task_list", art2.Type)
	assert.Equal(t, "phases/planning/tasks.md", art2.Path)
	assert.False(t, art2.Approved)
	assert.Equal(t, now.Add(-time.Hour), art2.Created_at)
	assert.Nil(t, art2.Metadata)
}

// TestConvertArtifacts_EmptySlice tests that empty artifact slices are handled correctly.
func TestConvertArtifacts_EmptySlice(t *testing.T) {
	stateArtifacts := []project.ArtifactState{}

	coll := convertArtifacts(stateArtifacts)

	assert.Len(t, coll, 0)
	assert.NotNil(t, coll)
}

// TestConvertArtifactsToState_ReverseMapping tests that artifacts convert back
// to CUE-generated types correctly.
func TestConvertArtifactsToState_ReverseMapping(t *testing.T) {
	now := time.Now()
	metadata := map[string]interface{}{"key": "value"}

	artifacts := ArtifactCollection{
		{
			ArtifactState: project.ArtifactState{
				Type:       "review",
				Path:       "phases/review/report.md",
				Approved:   true,
				Created_at: now,
				Metadata:   metadata,
			},
		},
	}

	stateArtifacts := convertArtifactsToState(artifacts)

	require.Len(t, stateArtifacts, 1)
	assert.Equal(t, "review", stateArtifacts[0].Type)
	assert.Equal(t, "phases/review/report.md", stateArtifacts[0].Path)
	assert.True(t, stateArtifacts[0].Approved)
	assert.Equal(t, now, stateArtifacts[0].Created_at)
	assert.Equal(t, metadata, stateArtifacts[0].Metadata)
}

// TestConvertTasks_AllFieldsMapped tests that all task fields are correctly
// mapped from CUE-generated TaskState to wrapper Task type.
func TestConvertTasks_AllFieldsMapped(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	metadata := map[string]interface{}{
		"complexity": "high",
		"priority":   1,
	}

	stateTasks := []project.TaskState{
		{
			Id:             "010",
			Name:           "Implement feature X",
			Phase:          "implementation",
			Status:         "completed",
			Created_at:     now.Add(-2 * time.Hour),
			Started_at:     startedAt,
			Updated_at:     now,
			Completed_at:   completedAt,
			Iteration:      2,
			Assigned_agent: "implementer",
			Inputs: []project.ArtifactState{
				{
					Type:       "design_doc",
					Path:       "design.md",
					Approved:   true,
					Created_at: now.Add(-3 * time.Hour),
				},
			},
			Outputs: []project.ArtifactState{
				{
					Type:       "code",
					Path:       "feature.go",
					Approved:   false,
					Created_at: now,
				},
			},
			Metadata: metadata,
		},
	}

	coll := convertTasks(stateTasks)

	require.Len(t, coll, 1)

	task, err := coll.Get("010")
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, "010", task.Id)
	assert.Equal(t, "Implement feature X", task.Name)
	assert.Equal(t, "implementation", task.Phase)
	assert.Equal(t, "completed", task.Status)
	assert.Equal(t, now.Add(-2*time.Hour), task.Created_at)
	assert.Equal(t, startedAt, task.Started_at)
	assert.Equal(t, now, task.Updated_at)
	assert.Equal(t, completedAt, task.Completed_at)
	assert.Equal(t, int64(2), task.Iteration)
	assert.Equal(t, "implementer", task.Assigned_agent)
	assert.Equal(t, metadata, task.Metadata)

	// Verify nested artifacts
	assert.Len(t, task.Inputs, 1)
	assert.Equal(t, "design_doc", task.Inputs[0].Type)

	assert.Len(t, task.Outputs, 1)
	assert.Equal(t, "code", task.Outputs[0].Type)
}

// TestConvertTasks_OptionalTimestampsNil tests that optional timestamps
// (StartedAt, CompletedAt) are handled correctly when not set.
func TestConvertTasks_OptionalTimestampsNil(t *testing.T) {
	now := time.Now()

	stateTasks := []project.TaskState{
		{
			Id:             "020",
			Name:           "Pending task",
			Phase:          "implementation",
			Status:         "pending",
			Created_at:     now,
			Updated_at:     now,
			Iteration:      1,
			Assigned_agent: "implementer",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		},
	}

	coll := convertTasks(stateTasks)

	task, err := coll.Get("020")
	require.NoError(t, err)

	// Zero time values should be preserved
	assert.True(t, task.Started_at.IsZero())
	assert.True(t, task.Completed_at.IsZero())
}

// TestConvertTasksToState_ReverseMapping tests that tasks convert back
// to CUE-generated types correctly.
func TestConvertTasksToState_ReverseMapping(t *testing.T) {
	now := time.Now()

	tasks := TaskCollection{
		{
			TaskState: project.TaskState{
				Id:             "030",
				Name:           "Test task",
				Phase:          "testing",
				Status:         "in_progress",
				Created_at:     now,
				Started_at:     now,
				Updated_at:     now,
				Iteration:      1,
				Assigned_agent: "tester",
				Inputs:         []project.ArtifactState{},
				Outputs:        []project.ArtifactState{},
			},
		},
	}

	stateTasks := convertTasksToState(tasks)

	require.Len(t, stateTasks, 1)
	assert.Equal(t, "030", stateTasks[0].Id)
	assert.Equal(t, "Test task", stateTasks[0].Name)
	assert.Equal(t, "testing", stateTasks[0].Phase)
	assert.Equal(t, "in_progress", stateTasks[0].Status)
}

// TestConvertPhases_AllFieldsMapped tests that all phase fields are correctly
// mapped from CUE-generated PhaseState to wrapper Phase type.
func TestConvertPhases_AllFieldsMapped(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now
	metadata := map[string]interface{}{
		"approval_required": true,
	}

	statePhases := map[string]project.PhaseState{
		"implementation": {
			Status:       "completed",
			Enabled:      true,
			Created_at:   now.Add(-2 * time.Hour),
			Started_at:   startedAt,
			Completed_at: completedAt,
			Metadata:     metadata,
			Inputs: []project.ArtifactState{
				{
					Type:       "task_list",
					Path:       "tasks.md",
					Approved:   true,
					Created_at: now.Add(-3 * time.Hour),
				},
			},
			Outputs: []project.ArtifactState{
				{
					Type:       "code",
					Path:       "implementation.go",
					Approved:   false,
					Created_at: now,
				},
			},
			Tasks: []project.TaskState{
				{
					Id:             "010",
					Name:           "Task 1",
					Phase:          "implementation",
					Status:         "completed",
					Created_at:     now.Add(-time.Hour),
					Updated_at:     now,
					Iteration:      1,
					Assigned_agent: "implementer",
					Inputs:         []project.ArtifactState{},
					Outputs:        []project.ArtifactState{},
				},
			},
		},
	}

	coll := convertPhases(statePhases)

	require.Len(t, coll, 1)

	phase, err := coll.Get("implementation")
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, "completed", phase.Status)
	assert.True(t, phase.Enabled)
	assert.Equal(t, now.Add(-2*time.Hour), phase.Created_at)
	assert.Equal(t, startedAt, phase.Started_at)
	assert.Equal(t, completedAt, phase.Completed_at)
	assert.Equal(t, metadata, phase.Metadata)

	// Verify nested collections
	assert.Len(t, phase.Inputs, 1)
	assert.Equal(t, "task_list", phase.Inputs[0].Type)

	assert.Len(t, phase.Outputs, 1)
	assert.Equal(t, "code", phase.Outputs[0].Type)

	assert.Len(t, phase.Tasks, 1)
	assert.Equal(t, "010", phase.Tasks[0].Id)
	assert.Equal(t, "Task 1", phase.Tasks[0].Name)
}

// TestConvertPhases_NilMetadata tests that nil metadata maps are handled gracefully.
func TestConvertPhases_NilMetadata(t *testing.T) {
	now := time.Now()

	statePhases := map[string]project.PhaseState{
		"design": {
			Status:     "pending",
			Enabled:    true,
			Created_at: now,
			Metadata:   nil, // Nil metadata
			Inputs:     []project.ArtifactState{},
			Outputs:    []project.ArtifactState{},
			Tasks:      []project.TaskState{},
		},
	}

	coll := convertPhases(statePhases)

	phase, err := coll.Get("design")
	require.NoError(t, err)
	assert.Nil(t, phase.Metadata)
}

// TestConvertPhases_EmptyCollections tests that empty artifact and task collections
// are handled correctly.
func TestConvertPhases_EmptyCollections(t *testing.T) {
	now := time.Now()

	statePhases := map[string]project.PhaseState{
		"review": {
			Status:     "pending",
			Enabled:    true,
			Created_at: now,
			Inputs:     []project.ArtifactState{},
			Outputs:    []project.ArtifactState{},
			Tasks:      []project.TaskState{},
		},
	}

	coll := convertPhases(statePhases)

	phase, err := coll.Get("review")
	require.NoError(t, err)
	assert.NotNil(t, phase.Inputs)
	assert.Len(t, phase.Inputs, 0)
	assert.NotNil(t, phase.Outputs)
	assert.Len(t, phase.Outputs, 0)
	assert.NotNil(t, phase.Tasks)
	assert.Len(t, phase.Tasks, 0)
}

// TestConvertPhasesToState_ReverseMapping tests that phases convert back
// to CUE-generated types correctly.
func TestConvertPhasesToState_ReverseMapping(t *testing.T) {
	now := time.Now()

	phases := PhaseCollection{
		"planning": &Phase{
			PhaseState: project.PhaseState{
				Status:       "in_progress",
				Enabled:      true,
				Created_at:   now,
				Started_at:   now,
				Completed_at: time.Time{},
				Inputs:       []project.ArtifactState{},
				Outputs:      []project.ArtifactState{},
				Tasks:        []project.TaskState{},
			},
		},
	}

	statePhases := convertPhasesToState(phases)

	require.Len(t, statePhases, 1)
	phase, exists := statePhases["planning"]
	require.True(t, exists)
	assert.Equal(t, "in_progress", phase.Status)
	assert.True(t, phase.Enabled)
}

// TestConversion_RoundTrip_Simple tests that a simple project state can be
// converted to wrapper types and back without data loss.
func TestConversion_RoundTrip_Simple(t *testing.T) {
	now := time.Now()

	original := project.ProjectState{
		Name:        "test-project",
		Type:        "standard",
		Branch:      "feat/test",
		Description: "Test project",
		Created_at:  now,
		Updated_at:  now,
		Phases: map[string]project.PhaseState{
			"planning": {
				Status:       "completed",
				Enabled:      true,
				Created_at:   now,
				Started_at:   now,
				Completed_at: now,
				Inputs:       []project.ArtifactState{},
				Outputs:      []project.ArtifactState{},
				Tasks:        []project.TaskState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "Planning",
			Updated_at:    now,
		},
	}

	// Convert to wrapper types
	proj := &Project{
		ProjectState: original,
	}

	// Convert back to CUE types
	converted := project.ProjectState{
		Name:        proj.Name,
		Type:        proj.Type,
		Branch:      proj.Branch,
		Description: proj.Description,
		Created_at:  proj.Created_at,
		Updated_at:  proj.Updated_at,
		Phases:      convertPhasesToState(convertPhases(proj.Phases)),
		Statechart:  proj.Statechart,
	}

	// Verify all fields match
	assert.Equal(t, original.Name, converted.Name)
	assert.Equal(t, original.Type, converted.Type)
	assert.Equal(t, original.Branch, converted.Branch)
	assert.Equal(t, original.Description, converted.Description)
	assert.Equal(t, original.Created_at, converted.Created_at)
	assert.Equal(t, original.Updated_at, converted.Updated_at)
	assert.Equal(t, original.Statechart.Current_state, converted.Statechart.Current_state)
	assert.Equal(t, original.Statechart.Updated_at, converted.Statechart.Updated_at)
	assert.Len(t, converted.Phases, 1)
}

// TestConversion_RoundTrip_Complex tests that a complex project state with
// all optional fields populated can be round-tripped without data loss.
func TestConversion_RoundTrip_Complex(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-2 * time.Hour)
	completedAt := now.Add(-time.Hour)

	original := project.ProjectState{
		Name:        "complex-project",
		Type:        "standard",
		Branch:      "feat/complex",
		Description: "Complex test project",
		Created_at:  now.Add(-3 * time.Hour),
		Updated_at:  now,
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status:       "completed",
				Enabled:      true,
				Created_at:   now.Add(-3 * time.Hour),
				Started_at:   startedAt,
				Completed_at: completedAt,
				Metadata: map[string]interface{}{
					"complexity": "high",
					"team_size":  5,
				},
				Inputs: []project.ArtifactState{
					{
						Type:       "design_doc",
						Path:       "design.md",
						Approved:   true,
						Created_at: now.Add(-4 * time.Hour),
						Metadata: map[string]interface{}{
							"author": "john",
						},
					},
				},
				Outputs: []project.ArtifactState{
					{
						Type:       "code",
						Path:       "impl.go",
						Approved:   false,
						Created_at: now,
					},
				},
				Tasks: []project.TaskState{
					{
						Id:             "010",
						Name:           "Implement core",
						Phase:          "implementation",
						Status:         "completed",
						Created_at:     startedAt,
						Started_at:     startedAt,
						Updated_at:     completedAt,
						Completed_at:   completedAt,
						Iteration:      2,
						Assigned_agent: "implementer",
						Inputs: []project.ArtifactState{
							{
								Type:       "spec",
								Path:       "spec.md",
								Approved:   true,
								Created_at: now.Add(-5 * time.Hour),
							},
						},
						Outputs: []project.ArtifactState{
							{
								Type:       "test",
								Path:       "test.go",
								Approved:   true,
								Created_at: completedAt,
							},
						},
						Metadata: map[string]interface{}{
							"priority": 1,
						},
					},
				},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "ImplementationComplete",
			Updated_at:    completedAt,
		},
	}

	// Convert to wrapper types
	phases := convertPhases(original.Phases)

	// Convert back to CUE types
	convertedPhases := convertPhasesToState(phases)

	// Verify phase
	phase := convertedPhases["implementation"]
	assert.Equal(t, "completed", phase.Status)
	assert.True(t, phase.Enabled)
	assert.Equal(t, original.Phases["implementation"].Metadata, phase.Metadata)

	// Verify inputs
	require.Len(t, phase.Inputs, 1)
	assert.Equal(t, "design_doc", phase.Inputs[0].Type)
	assert.Equal(t, "design.md", phase.Inputs[0].Path)
	assert.Equal(t, original.Phases["implementation"].Inputs[0].Metadata, phase.Inputs[0].Metadata)

	// Verify outputs
	require.Len(t, phase.Outputs, 1)
	assert.Equal(t, "code", phase.Outputs[0].Type)

	// Verify tasks
	require.Len(t, phase.Tasks, 1)
	task := phase.Tasks[0]
	assert.Equal(t, "010", task.Id)
	assert.Equal(t, "Implement core", task.Name)
	assert.Equal(t, int64(2), task.Iteration)
	assert.Equal(t, original.Phases["implementation"].Tasks[0].Metadata, task.Metadata)

	// Verify task inputs/outputs
	require.Len(t, task.Inputs, 1)
	assert.Equal(t, "spec", task.Inputs[0].Type)
	require.Len(t, task.Outputs, 1)
	assert.Equal(t, "test", task.Outputs[0].Type)
}

// TestConversion_RoundTrip_NilAndEmpty tests that nil and empty collections
// are preserved during round-trip conversion.
func TestConversion_RoundTrip_NilAndEmpty(t *testing.T) {
	now := time.Now()

	original := project.ProjectState{
		Name:       "minimal-project",
		Type:       "standard",
		Branch:     "feat/minimal",
		Created_at: now,
		Updated_at: now,
		Phases: map[string]project.PhaseState{
			"design": {
				Status:     "pending",
				Enabled:    true,
				Created_at: now,
				Metadata:   nil, // nil metadata
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
				Tasks:      []project.TaskState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "Design",
			Updated_at:    now,
		},
	}

	// Convert to wrapper types and back
	phases := convertPhases(original.Phases)
	converted := convertPhasesToState(phases)

	// Verify phase
	phase := converted["design"]
	assert.Nil(t, phase.Metadata)
	assert.NotNil(t, phase.Inputs)
	assert.Len(t, phase.Inputs, 0)
	assert.NotNil(t, phase.Outputs)
	assert.Len(t, phase.Outputs, 0)
	assert.NotNil(t, phase.Tasks)
	assert.Len(t, phase.Tasks, 0)
}
