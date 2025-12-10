package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMemoryBackend tests the NewMemoryBackend constructor.
func TestNewMemoryBackend(t *testing.T) {
	t.Run("creates empty backend where Exists returns false", func(t *testing.T) {
		backend := NewMemoryBackend()

		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("State returns nil for empty backend", func(t *testing.T) {
		backend := NewMemoryBackend()

		assert.Nil(t, backend.State())
	})
}

// TestNewMemoryBackendWithState tests the NewMemoryBackendWithState constructor.
func TestNewMemoryBackendWithState(t *testing.T) {
	t.Run("stores provided state", func(t *testing.T) {
		initialState := &project.ProjectState{
			Name:   "test-project",
			Type:   "standard",
			Branch: "feat/test",
		}

		backend := NewMemoryBackendWithState(initialState)

		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.True(t, exists)

		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "test-project", loaded.Name)
	})

	t.Run("handles nil state", func(t *testing.T) {
		backend := NewMemoryBackendWithState(nil)

		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("stores deep copy of initial state", func(t *testing.T) {
		initialState := &project.ProjectState{
			Name:   "original",
			Type:   "standard",
			Branch: "feat/test",
			Phases: map[string]project.PhaseState{
				"planning": {Status: "pending"},
			},
		}

		backend := NewMemoryBackendWithState(initialState)

		// Modify the original state
		initialState.Name = "modified"
		initialState.Phases["planning"] = project.PhaseState{Status: "completed"}

		// Backend should have the original values
		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "original", loaded.Name)
		assert.Equal(t, "pending", loaded.Phases["planning"].Status)
	})
}

// TestMemoryBackend_Load tests the Load method.
func TestMemoryBackend_Load(t *testing.T) {
	t.Run("returns ErrNotFound when empty", func(t *testing.T) {
		backend := NewMemoryBackend()

		_, err := backend.Load(context.Background())
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("returns state when populated", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name:        "test-project",
			Type:        "standard",
			Branch:      "feat/test",
			Description: "A test project",
		})

		state, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "test-project", state.Name)
		assert.Equal(t, "standard", state.Type)
		assert.Equal(t, "feat/test", state.Branch)
		assert.Equal(t, "A test project", state.Description)
	})

	t.Run("returns deep copy - modifying returned state does not affect backend", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name:   "original",
			Type:   "standard",
			Branch: "feat/test",
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:  "pending",
					Enabled: true,
					Tasks: []project.TaskState{
						{Id: "010", Name: "Task 1", Phase: "planning", Status: "pending"},
					},
				},
			},
			Agent_sessions: map[string]string{
				"planner": "session-123",
			},
		})

		// Load state and modify it
		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		loaded.Name = "modified"
		loaded.Phases["planning"] = project.PhaseState{Status: "completed"}
		loaded.Agent_sessions["planner"] = "modified-session"

		// Load again and verify original values are preserved
		reloaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "original", reloaded.Name)
		assert.Equal(t, "pending", reloaded.Phases["planning"].Status)
		assert.Equal(t, "session-123", reloaded.Agent_sessions["planner"])
	})
}

// TestMemoryBackend_Save tests the Save method.
func TestMemoryBackend_Save(t *testing.T) {
	t.Run("stores state that can be retrieved via Load", func(t *testing.T) {
		backend := NewMemoryBackend()
		ctx := context.Background()

		state := &project.ProjectState{
			Name:   "saved-project",
			Type:   "standard",
			Branch: "feat/save",
		}

		err := backend.Save(ctx, state)
		require.NoError(t, err)

		loaded, err := backend.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, "saved-project", loaded.Name)
	})

	t.Run("creates deep copy - modifying original after save does not affect backend", func(t *testing.T) {
		backend := NewMemoryBackend()
		ctx := context.Background()

		state := &project.ProjectState{
			Name:   "original",
			Type:   "standard",
			Branch: "feat/test",
			Phases: map[string]project.PhaseState{
				"planning": {
					Status:  "pending",
					Enabled: true,
					Inputs: []project.ArtifactState{
						{Type: "doc", Path: "design.md", Approved: false},
					},
				},
			},
		}

		err := backend.Save(ctx, state)
		require.NoError(t, err)

		// Modify the original state after save
		state.Name = "modified"
		state.Phases["planning"] = project.PhaseState{Status: "completed"}

		// Backend should have the original values
		loaded, err := backend.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, "original", loaded.Name)
		assert.Equal(t, "pending", loaded.Phases["planning"].Status)
	})

	t.Run("replaces existing state", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name:   "old-project",
			Type:   "standard",
			Branch: "feat/old",
		})
		ctx := context.Background()

		newState := &project.ProjectState{
			Name:   "new-project",
			Type:   "exploration",
			Branch: "explore/new",
		}

		err := backend.Save(ctx, newState)
		require.NoError(t, err)

		loaded, err := backend.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, "new-project", loaded.Name)
		assert.Equal(t, "exploration", loaded.Type)
	})

	t.Run("handles nil state", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name: "existing",
		})
		ctx := context.Background()

		err := backend.Save(ctx, nil)
		require.NoError(t, err)

		exists, err := backend.Exists(ctx)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestMemoryBackend_Exists tests the Exists method.
func TestMemoryBackend_Exists(t *testing.T) {
	t.Run("returns false when empty", func(t *testing.T) {
		backend := NewMemoryBackend()

		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns true after Save", func(t *testing.T) {
		backend := NewMemoryBackend()
		ctx := context.Background()

		err := backend.Save(ctx, &project.ProjectState{Name: "test"})
		require.NoError(t, err)

		exists, err := backend.Exists(ctx)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false after Delete", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})
		ctx := context.Background()

		err := backend.Delete(ctx)
		require.NoError(t, err)

		exists, err := backend.Exists(ctx)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestMemoryBackend_Delete tests the Delete method.
func TestMemoryBackend_Delete(t *testing.T) {
	t.Run("clears state - subsequent Exists returns false", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})
		ctx := context.Background()

		err := backend.Delete(ctx)
		require.NoError(t, err)

		exists, err := backend.Exists(ctx)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete on empty backend succeeds", func(t *testing.T) {
		backend := NewMemoryBackend()

		err := backend.Delete(context.Background())
		require.NoError(t, err)
	})

	t.Run("subsequent Load returns ErrNotFound after Delete", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})
		ctx := context.Background()

		err := backend.Delete(ctx)
		require.NoError(t, err)

		_, err = backend.Load(ctx)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

// TestMemoryBackend_State tests the State helper method.
func TestMemoryBackend_State(t *testing.T) {
	t.Run("returns nil for empty backend", func(t *testing.T) {
		backend := NewMemoryBackend()

		assert.Nil(t, backend.State())
	})

	t.Run("returns raw state for populated backend", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name: "test-project",
		})

		state := backend.State()
		require.NotNil(t, state)
		assert.Equal(t, "test-project", state.Name)
	})

	t.Run("returns same pointer for repeated calls", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})

		state1 := backend.State()
		state2 := backend.State()

		assert.Same(t, state1, state2)
	})
}

// TestMemoryBackend_ConcurrentAccess tests thread safety.
func TestMemoryBackend_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent reads do not block each other", func(t *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{
			Name: "concurrent-test",
		})
		ctx := context.Background()

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		// Spawn multiple concurrent readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := backend.Load(ctx)
				if err != nil {
					errChan <- err
				}
			}()
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("unexpected error during concurrent read: %v", err)
		}
	})

	t.Run("concurrent Save and Load operations complete without data races", func(_ *testing.T) {
		backend := NewMemoryBackend()
		ctx := context.Background()

		var wg sync.WaitGroup

		// Concurrent writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				state := &project.ProjectState{
					Name: fmt.Sprintf("project-%d", n),
				}
				_ = backend.Save(ctx, state)
			}(i)
		}

		// Concurrent readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = backend.Load(ctx)
				_, _ = backend.Exists(ctx)
			}()
		}

		wg.Wait()
		// If we get here without race detector errors, test passes
	})

	t.Run("concurrent Exists operations are safe", func(_ *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})
		ctx := context.Background()

		var wg sync.WaitGroup

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = backend.Exists(ctx)
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent Delete operations are safe", func(_ *testing.T) {
		backend := NewMemoryBackendWithState(&project.ProjectState{Name: "test"})
		ctx := context.Background()

		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = backend.Delete(ctx)
			}()
		}

		wg.Wait()
	})
}

// TestMemoryBackend_DeepCopy_PhasesMap tests that phases map is deep copied.
func TestMemoryBackend_DeepCopy_PhasesMap(t *testing.T) {
	original := &project.ProjectState{
		Name: "test",
		Phases: map[string]project.PhaseState{
			"planning": {
				Status:    "pending",
				Enabled:   true,
				Iteration: 1,
			},
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Phases["planning"] = project.PhaseState{Status: "completed"}
	original.Phases["new-phase"] = project.PhaseState{Status: "pending"}

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	// Backend should have original values
	assert.Equal(t, "pending", loaded.Phases["planning"].Status)
	_, hasNewPhase := loaded.Phases["new-phase"]
	assert.False(t, hasNewPhase)
}

// TestMemoryBackend_DeepCopy_Tasks tests that tasks slice is deep copied.
func TestMemoryBackend_DeepCopy_Tasks(t *testing.T) {
	now := time.Now()
	original := &project.ProjectState{
		Name: "test",
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status: "in_progress",
				Tasks: []project.TaskState{
					{
						Id:         "010",
						Name:       "Original Task",
						Phase:      "implementation",
						Status:     "pending",
						Created_at: now,
						Updated_at: now,
						Iteration:  1,
					},
				},
			},
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Phases["implementation"].Tasks[0].Name = "Modified Task"
	original.Phases["implementation"].Tasks[0].Status = "completed"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	// Backend should have original values
	tasks := loaded.Phases["implementation"].Tasks
	require.Len(t, tasks, 1)
	assert.Equal(t, "Original Task", tasks[0].Name)
	assert.Equal(t, "pending", tasks[0].Status)
}

// TestMemoryBackend_DeepCopy_Artifacts tests that artifacts slices are deep copied.
func TestMemoryBackend_DeepCopy_Artifacts(t *testing.T) {
	now := time.Now()
	original := &project.ProjectState{
		Name: "test",
		Phases: map[string]project.PhaseState{
			"planning": {
				Status: "in_progress",
				Inputs: []project.ArtifactState{
					{Type: "doc", Path: "input.md", Approved: false, Created_at: now},
				},
				Outputs: []project.ArtifactState{
					{Type: "plan", Path: "output.md", Approved: true, Created_at: now},
				},
			},
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Phases["planning"].Inputs[0].Path = "modified-input.md"
	original.Phases["planning"].Outputs[0].Path = "modified-output.md"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	phase := loaded.Phases["planning"]
	require.Len(t, phase.Inputs, 1)
	require.Len(t, phase.Outputs, 1)
	assert.Equal(t, "input.md", phase.Inputs[0].Path)
	assert.Equal(t, "output.md", phase.Outputs[0].Path)
}

// TestMemoryBackend_DeepCopy_TaskArtifacts tests that task inputs and outputs are deep copied.
func TestMemoryBackend_DeepCopy_TaskArtifacts(t *testing.T) {
	now := time.Now()
	original := &project.ProjectState{
		Name: "test",
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status: "in_progress",
				Tasks: []project.TaskState{
					{
						Id:        "010",
						Name:      "Task",
						Phase:     "implementation",
						Status:    "pending",
						Iteration: 1,
						Inputs: []project.ArtifactState{
							{Type: "reference", Path: "ref.md", Created_at: now},
						},
						Outputs: []project.ArtifactState{
							{Type: "code", Path: "code.go", Created_at: now},
						},
					},
				},
			},
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Phases["implementation"].Tasks[0].Inputs[0].Path = "modified-ref.md"
	original.Phases["implementation"].Tasks[0].Outputs[0].Path = "modified-code.go"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	task := loaded.Phases["implementation"].Tasks[0]
	require.Len(t, task.Inputs, 1)
	require.Len(t, task.Outputs, 1)
	assert.Equal(t, "ref.md", task.Inputs[0].Path)
	assert.Equal(t, "code.go", task.Outputs[0].Path)
}

// TestMemoryBackend_DeepCopy_Metadata tests that metadata maps are copied.
func TestMemoryBackend_DeepCopy_Metadata(t *testing.T) {
	original := &project.ProjectState{
		Name: "test",
		Phases: map[string]project.PhaseState{
			"planning": {
				Status: "pending",
				Metadata: map[string]any{
					"complexity": "high",
					"count":      42,
				},
			},
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Phases["planning"].Metadata["complexity"] = "low"
	original.Phases["planning"].Metadata["new-key"] = "new-value"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	metadata := loaded.Phases["planning"].Metadata
	assert.Equal(t, "high", metadata["complexity"])
	_, hasNewKey := metadata["new-key"]
	assert.False(t, hasNewKey)
}

// TestMemoryBackend_DeepCopy_AgentSessions tests that agent_sessions map is deep copied.
func TestMemoryBackend_DeepCopy_AgentSessions(t *testing.T) {
	original := &project.ProjectState{
		Name: "test",
		Agent_sessions: map[string]string{
			"planner":  "session-123",
			"reviewer": "session-456",
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Agent_sessions["planner"] = "modified-session"
	original.Agent_sessions["new-agent"] = "new-session"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "session-123", loaded.Agent_sessions["planner"])
	_, hasNewAgent := loaded.Agent_sessions["new-agent"]
	assert.False(t, hasNewAgent)
}

// TestMemoryBackend_DeepCopy_Statechart tests that statechart is copied.
func TestMemoryBackend_DeepCopy_Statechart(t *testing.T) {
	now := time.Now()
	original := &project.ProjectState{
		Name: "test",
		Statechart: project.StatechartState{
			Current_state: "PlanningActive",
			Updated_at:    now,
		},
	}

	backend := NewMemoryBackendWithState(original)

	// Modify original
	original.Statechart.Current_state = "ImplementationActive"

	loaded, err := backend.Load(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "PlanningActive", loaded.Statechart.Current_state)
}

// TestMemoryBackend_FullWorkflow tests a complete workflow.
func TestMemoryBackend_FullWorkflow(t *testing.T) {
	backend := NewMemoryBackendWithState(&project.ProjectState{
		Name:   "workflow-test",
		Type:   "standard",
		Branch: "feat/workflow",
		Phases: map[string]project.PhaseState{
			"planning": {Status: "in_progress"},
		},
	})
	ctx := context.Background()

	// Load and verify initial state
	state, err := backend.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, "workflow-test", state.Name)
	assert.Equal(t, "in_progress", state.Phases["planning"].Status)

	// Modify and save
	state.Phases["planning"] = project.PhaseState{Status: "completed"}
	state.Phases["implementation"] = project.PhaseState{Status: "in_progress"}
	err = backend.Save(ctx, state)
	require.NoError(t, err)

	// Load and verify updated state
	updated, err := backend.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, "completed", updated.Phases["planning"].Status)
	assert.Equal(t, "in_progress", updated.Phases["implementation"].Status)

	// Delete
	err = backend.Delete(ctx)
	require.NoError(t, err)

	// Verify deleted
	exists, err := backend.Exists(ctx)
	require.NoError(t, err)
	assert.False(t, exists)
}
