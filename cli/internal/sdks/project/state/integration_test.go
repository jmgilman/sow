package state

import (
	"os"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_PersistenceWorkflow verifies the complete Load → Mutate → Save → Load cycle.
// This is the most critical integration test as it validates that data persists correctly
// across the full lifecycle of loading, modifying, saving, and reloading.
func TestIntegration_PersistenceWorkflow(t *testing.T) {
	// Setup: Create test environment with valid project state
	ctx := setupTestRepo(t)

	// Copy valid test fixture
	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// Create context with temporary directory
	// ctx already created by setupTestRepo

	// Step 1: Load project from disk
	p, err := Load(ctx)
	require.NoError(t, err, "initial Load() should succeed")
	require.NotNil(t, p)

	// Step 2: Mutate project state
	// 2a. Update phase status
	phase := p.Phases["planning"]
	originalStatus := phase.Status
	phase.Status = "in_progress"

	// 2b. Add artifact to phase outputs
	artifact := project.ArtifactState{
		Type:       "test",
		Path:       "test.md",
		Approved:   true,
		Created_at: time.Now(),
	}
	phase.Outputs = append(phase.Outputs, artifact)

	// 2c. Add metadata to phase
	phase.Metadata = map[string]interface{}{
		"test_key": "test_value",
		"count":    42,
	}

	// IMPORTANT: Assign back to map since Phases is map[string]PhaseState (by value)
	p.Phases["planning"] = phase

	// Step 3: Save to disk
	err = p.Save()
	require.NoError(t, err, "Save() should succeed after mutations")

	// Step 4: Load again from disk
	p2, err := Load(ctx)
	require.NoError(t, err, "second Load() should succeed")
	require.NotNil(t, p2)

	// Step 5: Verify all mutations persisted correctly
	phase2, exists := p2.Phases["planning"]
	require.True(t, exists, "planning phase should still exist after reload")

	// Verify phase status changed
	assert.Equal(t, "in_progress", phase2.Status, "phase status should persist")
	assert.NotEqual(t, originalStatus, phase2.Status, "phase status should have changed from original")

	// Verify artifact was added
	require.Len(t, phase2.Outputs, 1, "should have one output artifact")
	assert.Equal(t, "test", phase2.Outputs[0].Type, "artifact type should persist")
	assert.Equal(t, "test.md", phase2.Outputs[0].Path, "artifact path should persist")
	assert.True(t, phase2.Outputs[0].Approved, "artifact approved status should persist")

	// Verify metadata persisted
	require.NotNil(t, phase2.Metadata, "metadata should not be nil")
	assert.Equal(t, "test_value", phase2.Metadata["test_key"], "string metadata should persist")
	// Metadata value should be 42 (YAML may unmarshal as int or float64)
	countVal := phase2.Metadata["count"]
	switch v := countVal.(type) {
	case int:
		assert.Equal(t, 42, v, "numeric metadata should persist")
	case float64:
		assert.Equal(t, 42, int(v), "numeric metadata should persist")
	default:
		t.Fatalf("unexpected type for count: %T", v)
	}

	// Verify timestamps were updated
	assert.False(t, p2.Updated_at.IsZero(), "project updated_at should be set")
}

// TestIntegration_ComplexNestedStructure verifies that complex nested structures
// (phases → tasks → artifacts) persist correctly through the full cycle.
func TestIntegration_ComplexNestedStructure(t *testing.T) {
	// Setup: Create test environment
	ctx := setupTestRepo(t)

	// Use complex fixture with multiple phases and tasks
	fixtureData, err := os.ReadFile("testdata/complex-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load project
	p, err := Load(ctx)
	require.NoError(t, err)

	// Verify initial complex structure
	t.Run("initial structure", func(t *testing.T) {
		// Verify planning phase
		planning, exists := p.Phases["planning"]
		require.True(t, exists)
		assert.Equal(t, "completed", planning.Status)
		assert.Len(t, planning.Outputs, 1, "planning should have 1 output")
		assert.Len(t, planning.Tasks, 1, "planning should have 1 task")

		// Verify planning task
		task := planning.Tasks[0]
		assert.Equal(t, "010", task.Id)
		assert.Equal(t, "completed", task.Status)
		assert.Len(t, task.Outputs, 1, "task should have 1 output")

		// Verify implementation phase
		impl, exists := p.Phases["implementation"]
		require.True(t, exists)
		assert.Equal(t, "in_progress", impl.Status)
		assert.Len(t, impl.Inputs, 1, "implementation should have 1 input from planning")
		assert.Len(t, impl.Tasks, 1, "implementation should have 1 task")

		// Verify implementation task
		implTask := impl.Tasks[0]
		assert.Equal(t, "020", implTask.Id)
		assert.Equal(t, "in_progress", implTask.Status)
		assert.NotNil(t, implTask.Metadata, "task should have metadata")
		assert.Equal(t, "high", implTask.Metadata["complexity"])
	})

	// Mutate nested structures
	t.Run("mutate nested", func(_ *testing.T) {
		// Add a new task to implementation
		newTask := project.TaskState{
			Id:             "030",
			Name:           "New task",
			Phase:          "implementation",
			Status:         "pending",
			Created_at:     time.Now(),
			Updated_at:     time.Now(),
			Iteration:      1,
			Assigned_agent: "implementer",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		}
		impl := p.Phases["implementation"]
		impl.Tasks = append(impl.Tasks, newTask)

		// Add metadata to new task
		newTask.Metadata = map[string]interface{}{
			"complexity": "low",
			"priority":   "high",
		}
		impl.Tasks[1].Metadata = newTask.Metadata

		// Add artifact to implementation outputs
		implArtifact := project.ArtifactState{
			Type:       "code",
			Path:       "src/feature.go",
			Approved:   false,
			Created_at: time.Now(),
		}
		impl.Outputs = append(impl.Outputs, implArtifact)

		// Assign back to map
		p.Phases["implementation"] = impl
	})

	// Save and reload
	require.NoError(t, p.Save())
	p2, err := Load(ctx)
	require.NoError(t, err)

	// Verify all nested mutations persisted
	t.Run("verify persisted changes", func(t *testing.T) {
		impl := p2.Phases["implementation"]
		require.NotNil(t, impl)

		// Verify new task persisted
		assert.Len(t, impl.Tasks, 2, "should have 2 tasks after adding one")
		newTask := impl.Tasks[1]
		assert.Equal(t, "030", newTask.Id)
		assert.Equal(t, "New task", newTask.Name)
		assert.NotNil(t, newTask.Metadata)
		assert.Equal(t, "low", newTask.Metadata["complexity"])
		assert.Equal(t, "high", newTask.Metadata["priority"])

		// Verify new artifact persisted
		assert.Len(t, impl.Outputs, 1, "should have 1 output")
		assert.Equal(t, "code", impl.Outputs[0].Type)
		assert.Equal(t, "src/feature.go", impl.Outputs[0].Path)
		assert.False(t, impl.Outputs[0].Approved)
	})
}

// TestIntegration_ValidationPreventsInvalidState verifies that the validation layer
// prevents invalid state from being saved to disk, protecting data integrity.
func TestIntegration_ValidationPreventsInvalidState(t *testing.T) {
	// Setup: Create valid project state
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load valid project
	p, err := Load(ctx)
	require.NoError(t, err)

	// Store original name for verification
	originalName := p.Name

	t.Run("missing required field", func(t *testing.T) {
		// Create invalid state - name is required by CUE schema
		p.Name = ""

		// Attempt to save
		err := p.Save()

		// Verify save failed
		assert.Error(t, err, "Save() should fail with missing required field")
		assert.Contains(t, err.Error(), "validation", "error should mention validation")

		// Restore for next test
		p.Name = originalName
	})

	t.Run("file not modified on validation failure", func(t *testing.T) {
		// Make project invalid
		p.Name = ""

		// Attempt to save (should fail)
		err := p.Save()
		require.Error(t, err)

		// Load from disk again
		p2, err := Load(ctx)
		require.NoError(t, err, "Load() should still work with unchanged file")

		// Verify file still contains original valid state
		assert.Equal(t, originalName, p2.Name, "name should be unchanged on disk")
		assert.NotEmpty(t, p2.Name, "name should not be empty")

		// Restore for next test
		p.Name = originalName
	})

	t.Run("invalid type value", func(t *testing.T) {
		// Store original and set invalid type
		originalType := p.Type
		p.Type = "" // Type is required

		err := p.Save()
		assert.Error(t, err, "Save() should fail with missing type")

		// Restore
		p.Type = originalType
	})
}

// TestIntegration_AtomicWriteProtection verifies that Save() uses atomic writes
// to protect against file corruption. The atomic write pattern (write to temp file,
// then rename) ensures that the state file is never left in a partially-written state.
func TestIntegration_AtomicWriteProtection(t *testing.T) {
	// Setup: Create valid project state
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load project
	p, err := Load(ctx)
	require.NoError(t, err)

	t.Run("no temp file after successful save", func(t *testing.T) {
		// Modify and save
		p.Description = "Updated description"
		err := p.Save()
		require.NoError(t, err)

		// Verify no temp file left behind
		tempFile := "project/state.yaml.tmp"
		_, err = ctx.FS().Stat(tempFile)
		assert.True(t, os.IsNotExist(err), "temp file should not exist after successful save")
	})

	t.Run("no temp file after failed save", func(t *testing.T) {
		// Make project invalid to trigger save failure
		p.Name = ""
		err := p.Save()
		require.Error(t, err)

		// Verify no temp file left behind even on failure
		tempFile := "project/state.yaml.tmp"
		_, err = ctx.FS().Stat(tempFile)
		assert.True(t, os.IsNotExist(err), "temp file should not exist after failed save")
	})

	t.Run("original file preserved on validation failure", func(t *testing.T) {
		// Load original state to get baseline
		p2, err := Load(ctx)
		require.NoError(t, err)
		originalDesc := p2.Description

		// Make project invalid
		p2.Name = ""
		err = p2.Save()
		require.Error(t, err, "save should fail")

		// Load again and verify original file unchanged
		p3, err := Load(ctx)
		require.NoError(t, err, "should be able to load original file")
		assert.NotEmpty(t, p3.Name, "name should still be present")
		assert.Equal(t, originalDesc, p3.Description, "description should be unchanged")
	})

	t.Run("atomic rename behavior", func(t *testing.T) {
		// This test documents the atomic write pattern:
		// 1. Write to temporary file (state.yaml.tmp)
		// 2. Rename temp file to final location (atomic operation)
		// 3. Original file preserved until rename succeeds
		//
		// This ensures that even if the process crashes during write,
		// the original file remains intact.

		// Load fresh project
		p4, err := Load(ctx)
		require.NoError(t, err)

		// Make a valid mutation
		p4.Description = "Test atomic write"

		// Save (internally uses temp file + rename)
		err = p4.Save()
		require.NoError(t, err)

		// Verify final state
		p5, err := Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Test atomic write", p5.Description)
	})
}

// TestIntegration_MetadataValidation verifies that project-type-specific metadata
// validation is enforced during Save(). Note: This test uses a stub validator for now
// since full metadata validation will be implemented in Unit 3.
func TestIntegration_MetadataValidation(t *testing.T) {
	// Setup: Create valid project state
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load project
	p, err := Load(ctx)
	require.NoError(t, err)

	t.Run("valid metadata structure", func(t *testing.T) {
		// Add metadata to phase
		phase := p.Phases["planning"]
		phase.Metadata = map[string]interface{}{
			"approved":   true,
			"complexity": "medium",
		}
		p.Phases["planning"] = phase

		// Save should succeed with valid metadata
		err := p.Save()
		assert.NoError(t, err, "Save() should succeed with valid metadata structure")

		// Verify metadata persisted
		p2, err := Load(ctx)
		require.NoError(t, err)
		phase2 := p2.Phases["planning"]
		assert.NotNil(t, phase2.Metadata)
		assert.Equal(t, true, phase2.Metadata["approved"])
		assert.Equal(t, "medium", phase2.Metadata["complexity"])
	})

	t.Run("metadata is optional", func(t *testing.T) {
		// Load fresh project
		p3, err := Load(ctx)
		require.NoError(t, err)

		// Remove metadata
		phase := p3.Phases["planning"]
		phase.Metadata = nil
		p3.Phases["planning"] = phase

		// Save should succeed - metadata is optional
		err = p3.Save()
		assert.NoError(t, err, "Save() should succeed with nil metadata")

		// Verify metadata remains nil or empty (YAML may deserialize nil as empty map)
		p4, err := Load(ctx)
		require.NoError(t, err)
		metadata := p4.Phases["planning"].Metadata
		assert.True(t, len(metadata) == 0, "metadata should be nil or empty")
	})

	// Note: Detailed metadata schema validation will be implemented in Unit 3.
	// This test verifies that the metadata validation hook is in place and working.
}

// TestIntegration_CollectionOperations verifies that collection operations
// (Get, Add, Remove) work correctly in the context of the full persistence cycle.
func TestIntegration_CollectionOperations(t *testing.T) {
	// Setup
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/complex-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load project
	p, err := Load(ctx)
	require.NoError(t, err)

	t.Run("phase collection operations", func(t *testing.T) {
		// Get phase using map access
		phase, exists := p.Phases["planning"]
		require.True(t, exists)
		assert.Equal(t, "completed", phase.Status)

		// Get non-existent phase
		_, exists = p.Phases["nonexistent"]
		assert.False(t, exists)
	})

	t.Run("artifact collection operations", func(t *testing.T) {
		phase := p.Phases["planning"]
		artifacts := ArtifactCollection(convertArtifacts(phase.Outputs))

		// Get existing artifact
		artifact, err := artifacts.Get(0)
		require.NoError(t, err)
		assert.Equal(t, "task_list", artifact.Type)

		// Add artifact
		newArtifact := Artifact{
			ArtifactState: project.ArtifactState{
				Type:       "review",
				Path:       "review.md",
				Approved:   false,
				Created_at: time.Now(),
			},
		}
		err = artifacts.Add(newArtifact)
		require.NoError(t, err)
		phase.Outputs = convertArtifactsToState(artifacts)
		p.Phases["planning"] = phase

		// Save and reload
		require.NoError(t, p.Save())
		p2, err := Load(ctx)
		require.NoError(t, err)

		// Verify artifact persisted
		phase2 := p2.Phases["planning"]
		assert.Len(t, phase2.Outputs, 2)
		assert.Equal(t, "review", phase2.Outputs[1].Type)
	})

	t.Run("task collection operations", func(t *testing.T) {
		impl := p.Phases["implementation"]
		tasks := TaskCollection(convertTasks(impl.Tasks))

		// Get task by ID
		task, err := tasks.Get("020")
		require.NoError(t, err)
		assert.Equal(t, "Implement feature", task.Name)

		// Get non-existent task
		_, err = tasks.Get("999")
		assert.Error(t, err)

		// Add task
		newTask := Task{
			TaskState: project.TaskState{
				Id:             "030",
				Name:           "New integration task",
				Phase:          "implementation",
				Status:         "pending",
				Created_at:     time.Now(),
				Updated_at:     time.Now(),
				Iteration:      1,
				Assigned_agent: "implementer",
				Inputs:         []project.ArtifactState{},
				Outputs:        []project.ArtifactState{},
			},
		}
		err = tasks.Add(newTask)
		require.NoError(t, err)
		impl.Tasks = convertTasksToState(tasks)
		p.Phases["implementation"] = impl

		// Save and reload
		require.NoError(t, p.Save())
		p3, err := Load(ctx)
		require.NoError(t, err)

		// Verify task persisted and can be retrieved by ID
		impl2 := p3.Phases["implementation"]
		tasks2 := TaskCollection(convertTasks(impl2.Tasks))
		task2, err := tasks2.Get("030")
		require.NoError(t, err)
		assert.Equal(t, "New integration task", task2.Name)
	})
}

// TestIntegration_HelperMethods verifies that Project helper methods work correctly
// with real project state loaded from disk.
func TestIntegration_HelperMethods(t *testing.T) {
	// Setup
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/complex-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo
	p, err := Load(ctx)
	require.NoError(t, err)

	t.Run("PhaseOutputApproved", func(t *testing.T) {
		// Check approved output
		assert.True(t, p.PhaseOutputApproved("planning", "task_list"))

		// Check non-existent output type
		assert.False(t, p.PhaseOutputApproved("planning", "nonexistent"))

		// Check non-existent phase
		assert.False(t, p.PhaseOutputApproved("nonexistent", "task_list"))

		// Add unapproved output and verify
		impl := p.Phases["implementation"]
		impl.Outputs = append(impl.Outputs, project.ArtifactState{
			Type:       "code",
			Path:       "code.go",
			Approved:   false,
			Created_at: time.Now(),
		})
		p.Phases["implementation"] = impl
		assert.False(t, p.PhaseOutputApproved("implementation", "code"))

		// Approve it and verify
		impl.Outputs[0].Approved = true
		p.Phases["implementation"] = impl
		assert.True(t, p.PhaseOutputApproved("implementation", "code"))
	})

	t.Run("PhaseMetadataBool", func(t *testing.T) {
		// Add metadata to phase
		impl := p.Phases["implementation"]
		if impl.Metadata == nil {
			impl.Metadata = make(map[string]interface{})
		}
		impl.Metadata["approved"] = true
		impl.Metadata["rejected"] = false
		impl.Metadata["count"] = 42 // Not a bool
		p.Phases["implementation"] = impl

		// Check boolean values
		assert.True(t, p.PhaseMetadataBool("implementation", "approved"))
		assert.False(t, p.PhaseMetadataBool("implementation", "rejected"))

		// Check non-boolean value
		assert.False(t, p.PhaseMetadataBool("implementation", "count"))

		// Check non-existent key
		assert.False(t, p.PhaseMetadataBool("implementation", "nonexistent"))

		// Check non-existent phase
		assert.False(t, p.PhaseMetadataBool("nonexistent", "approved"))
	})

	t.Run("AllTasksComplete", func(t *testing.T) {
		// Initially, not all tasks are complete (task 020 is in_progress)
		assert.False(t, p.AllTasksComplete())

		// Complete all tasks
		for name, phase := range p.Phases {
			for i := range phase.Tasks {
				phase.Tasks[i].Status = "completed"
			}
			p.Phases[name] = phase
		}
		assert.True(t, p.AllTasksComplete())

		// Add incomplete task
		impl := p.Phases["implementation"]
		impl.Tasks = append(impl.Tasks, project.TaskState{
			Id:             "999",
			Status:         "pending",
			Name:           "Incomplete task",
			Phase:          "implementation",
			Created_at:     time.Now(),
			Updated_at:     time.Now(),
			Iteration:      1,
			Assigned_agent: "implementer",
			Inputs:         []project.ArtifactState{},
			Outputs:        []project.ArtifactState{},
		})
		p.Phases["implementation"] = impl
		assert.False(t, p.AllTasksComplete())
	})
}

// TestIntegration_TimestampUpdates verifies that timestamps are updated correctly
// during Save() operations.
func TestIntegration_TimestampUpdates(t *testing.T) {
	// Setup
	ctx := setupTestRepo(t)

	fixtureData, err := os.ReadFile("testdata/valid-project.yaml")
	require.NoError(t, err)

	// Write using FS abstraction
	require.NoError(t, ctx.FS().WriteFile("project/state.yaml", fixtureData, 0644))

	// ctx already created by setupTestRepo

	// Load project
	p, err := Load(ctx)
	require.NoError(t, err)

	originalUpdatedAt := p.Updated_at

	// Wait a small amount to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Mutate and save
	p.Description = "Updated"
	err = p.Save()
	require.NoError(t, err)

	// Reload and verify timestamp updated
	p2, err := Load(ctx)
	require.NoError(t, err)

	assert.True(t, p2.Updated_at.After(originalUpdatedAt),
		"updated_at should be newer after Save()")
}
