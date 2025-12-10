package design

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/project"
	"github.com/jmgilman/sow/libs/project/state"
	
	projschema "github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageRegistration verifies that the design project type is registered
// with the global registry during package initialization.
func TestPackageRegistration(t *testing.T) {
	// The init() function should have registered "design" type
	config, exists := state.GetConfig("design")
	assert.True(t, exists, "design project type should be registered")
	assert.NotNil(t, config, "design config should not be nil")
}

// TestNewDesignProjectConfig verifies that NewDesignProjectConfig
// returns a valid configuration with all required components.
func TestNewDesignProjectConfig(t *testing.T) {
	config := NewDesignProjectConfig()

	require.NotNil(t, config, "config should not be nil")
	assert.NotNil(t, config, "NewDesignProjectConfig should return non-nil config")
}

// TestNewDesignProjectConfig_InitialState verifies that the configuration
// has the correct initial state set to Active.
func TestNewDesignProjectConfig_InitialState(t *testing.T) {
	config := NewDesignProjectConfig()

	// Verify initial state is Active
	assert.Equal(t, string(Active), config.InitialState(), "initial state should be Active")
}

// TestInitializeDesignProject_CreatesPhases verifies that the initializer
// creates both design and finalization phases with correct structure.
func TestInitializeDesignProject_CreatesPhases(t *testing.T) {
	// Create a minimal project for testing
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Call initializer with no initial inputs
	err := initializeDesignProject(project, nil)
	require.NoError(t, err, "initialization should not error")

	// Verify both phases were created
	assert.Len(t, project.Phases, 2, "should create exactly 2 phases")
	assert.Contains(t, project.Phases, "design", "should create design phase")
	assert.Contains(t, project.Phases, "finalization", "should create finalization phase")
}

// TestInitializeDesignProject_DesignPhaseActive verifies that the
// design phase starts in active status with enabled=true.
func TestInitializeDesignProject_DesignPhaseActive(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeDesignProject(project, nil)
	require.NoError(t, err)

	designPhase := project.Phases["design"]

	// Design phase should start active
	assert.Equal(t, "active", designPhase.Status, "design phase should have status='active'")
	assert.True(t, designPhase.Enabled, "design phase should be enabled")

	// Verify timestamps
	assert.Equal(t, now, designPhase.Created_at, "design phase should use project creation time")

	// Verify collections are initialized
	assert.NotNil(t, designPhase.Inputs, "design phase inputs should be initialized")
	assert.NotNil(t, designPhase.Outputs, "design phase outputs should be initialized")
	assert.NotNil(t, designPhase.Tasks, "design phase tasks should be initialized")
	assert.NotNil(t, designPhase.Metadata, "design phase metadata should be initialized")

	// Verify collections are empty
	assert.Empty(t, designPhase.Inputs, "design phase inputs should be empty")
	assert.Empty(t, designPhase.Outputs, "design phase outputs should be empty")
	assert.Empty(t, designPhase.Tasks, "design phase tasks should be empty")
}

// TestInitializeDesignProject_FinalizationPhasePending verifies that the
// finalization phase starts in pending status with enabled=false.
func TestInitializeDesignProject_FinalizationPhasePending(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeDesignProject(project, nil)
	require.NoError(t, err)

	finalizationPhase := project.Phases["finalization"]

	// Finalization phase should start pending
	assert.Equal(t, "pending", finalizationPhase.Status, "finalization phase should have status='pending'")
	assert.False(t, finalizationPhase.Enabled, "finalization phase should be disabled")

	// Verify timestamps
	assert.Equal(t, now, finalizationPhase.Created_at, "finalization phase should use project creation time")

	// Verify collections are initialized
	assert.NotNil(t, finalizationPhase.Inputs, "finalization phase inputs should be initialized")
	assert.NotNil(t, finalizationPhase.Outputs, "finalization phase outputs should be initialized")
	assert.NotNil(t, finalizationPhase.Tasks, "finalization phase tasks should be initialized")
	assert.NotNil(t, finalizationPhase.Metadata, "finalization phase metadata should be initialized")

	// Verify collections are empty
	assert.Empty(t, finalizationPhase.Inputs, "finalization phase inputs should be empty")
	assert.Empty(t, finalizationPhase.Outputs, "finalization phase outputs should be empty")
	assert.Empty(t, finalizationPhase.Tasks, "finalization phase tasks should be empty")
}

// TestInitializeDesignProject_WithInitialInputs verifies that the initializer
// properly handles initial inputs provided during project creation.
func TestInitializeDesignProject_WithInitialInputs(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Create test initial inputs
	designInputs := []projschema.ArtifactState{
		{
			Type: "requirement",
			Path: "/test/requirement.md",
		},
	}

	finalizationInputs := []projschema.ArtifactState{
		{
			Type: "summary",
			Path: "/test/summary.md",
		},
	}

	initialInputs := map[string][]projschema.ArtifactState{
		"design":       designInputs,
		"finalization": finalizationInputs,
	}

	// Call initializer with initial inputs
	err := initializeDesignProject(project, initialInputs)
	require.NoError(t, err)

	// Verify design phase received its inputs
	designPhase := project.Phases["design"]
	assert.Len(t, designPhase.Inputs, 1, "design phase should have 1 input")
	assert.Equal(t, "requirement", designPhase.Inputs[0].Type)

	// Verify finalization phase received its inputs
	finalizationPhase := project.Phases["finalization"]
	assert.Len(t, finalizationPhase.Inputs, 1, "finalization phase should have 1 input")
	assert.Equal(t, "summary", finalizationPhase.Inputs[0].Type)
}

// TestInitializeDesignProject_NoInitialInputs verifies that the initializer
// works correctly when initialInputs is nil.
func TestInitializeDesignProject_NoInitialInputs(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Call initializer with nil inputs
	err := initializeDesignProject(project, nil)
	require.NoError(t, err)

	// Both phases should have empty inputs
	designPhase := project.Phases["design"]
	assert.Empty(t, designPhase.Inputs, "design phase inputs should be empty")

	finalizationPhase := project.Phases["finalization"]
	assert.Empty(t, finalizationPhase.Inputs, "finalization phase inputs should be empty")
}

// TestConfigurePhases_ReturnsNonNilBuilder verifies that configurePhases
// returns a non-nil builder for chaining.
func TestConfigurePhases_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("design")
	result := configurePhases(builder)
	assert.NotNil(t, result, "configurePhases should return non-nil builder")
}

// TestConfigurePhases_DesignPhaseTaskSupport verifies that the
// design phase supports tasks.
func TestConfigurePhases_DesignPhaseTaskSupport(t *testing.T) {
	config := NewDesignProjectConfig()

	// Design phase should support tasks
	assert.True(t, config.PhaseSupportsTasks("design"),
		"design phase should support tasks")
}

// TestConfigurePhases_FinalizationPhaseTaskSupport verifies that the
// finalization phase supports tasks.
func TestConfigurePhases_FinalizationPhaseTaskSupport(t *testing.T) {
	config := NewDesignProjectConfig()

	// Finalization phase should support tasks
	assert.True(t, config.PhaseSupportsTasks("finalization"),
		"finalization phase should support tasks")
}

// TestConfigurePhases_GetTaskSupportingPhases verifies that both the
// design and finalization phases are returned as task-supporting.
func TestConfigurePhases_GetTaskSupportingPhases(t *testing.T) {
	config := NewDesignProjectConfig()

	phases := config.GetTaskSupportingPhases()

	// Should have exactly two task-supporting phases
	assert.Len(t, phases, 2, "should have exactly two task-supporting phases")
	assert.Contains(t, phases, "design", "task-supporting phases should include design")
	assert.Contains(t, phases, "finalization", "task-supporting phases should include finalization")
}

// TestConfigurePhases_GetDefaultTaskPhase verifies that the default
// task phase is correctly determined based on state.
func TestConfigurePhases_GetDefaultTaskPhase(t *testing.T) {
	config := NewDesignProjectConfig()

	// Active state should map to design phase
	phase := config.GetDefaultTaskPhase(string(Active))
	assert.Equal(t, "design", phase, "Active state should default to design phase")

	// Finalizing state should map to finalization phase
	phase = config.GetDefaultTaskPhase(string(Finalizing))
	assert.Equal(t, "finalization", phase, "Finalizing state should default to finalization phase")
}

// TestConfigurePhases_DesignOutputTypes verifies that the design
// phase allows design document output types.
func TestConfigurePhases_DesignOutputTypes(t *testing.T) {
	config := NewDesignProjectConfig()

	// Create test project with design phase
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test allowed output types
	allowedTypes := []string{"design", "adr", "architecture", "diagram", "spec"}
	for _, outputType := range allowedTypes {
		artifact := projschema.ArtifactState{
			Type:     outputType,
			Path:     "/test/" + outputType + ".md",
			Approved: true,
		}
		phase := proj.Phases["design"]
		phase.Outputs = []projschema.ArtifactState{artifact}
		proj.Phases["design"] = phase

		err = config.Validate(proj)
		assert.NoError(t, err, "design phase should allow '%s' output type", outputType)
	}
}

// TestConfigurePhases_FinalizationOutputTypes verifies that the finalization
// phase allows "pr" output type.
func TestConfigurePhases_FinalizationOutputTypes(t *testing.T) {
	config := NewDesignProjectConfig()

	// Create test project with finalization phase
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test "pr" output type is allowed
	prArtifact := projschema.ArtifactState{
		Type:     "pr",
		Path:     "/test/pr.md",
		Approved: true,
	}
	phase := proj.Phases["finalization"]
	phase.Outputs = []projschema.ArtifactState{prArtifact}
	proj.Phases["finalization"] = phase

	err = config.Validate(proj)
	assert.NoError(t, err, "finalization phase should allow 'pr' output type")
}

// TestConfigurePhases_DesignRejectsInvalidOutputs verifies that the
// design phase rejects output types not in the allowed list.
func TestConfigurePhases_DesignRejectsInvalidOutputs(t *testing.T) {
	config := NewDesignProjectConfig()

	// Create test project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test that "pr" output type is rejected in design phase
	prArtifact := projschema.ArtifactState{
		Type:     "pr",
		Path:     "/test/pr.md",
		Approved: true,
	}
	phase := proj.Phases["design"]
	phase.Outputs = []projschema.ArtifactState{prArtifact}
	proj.Phases["design"] = phase

	err = config.Validate(proj)
	assert.Error(t, err, "design phase should reject 'pr' output type")
}

// TestConfigurePhases_FinalizationRejectsInvalidOutputs verifies that the
// finalization phase rejects output types not in the allowed list.
func TestConfigurePhases_FinalizationRejectsInvalidOutputs(t *testing.T) {
	config := NewDesignProjectConfig()

	// Create test project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test that "design" output type is rejected in finalization phase
	designArtifact := projschema.ArtifactState{
		Type:     "design",
		Path:     "/test/design.md",
		Approved: true,
	}
	phase := proj.Phases["finalization"]
	phase.Outputs = []projschema.ArtifactState{designArtifact}
	proj.Phases["finalization"] = phase

	err = config.Validate(proj)
	assert.Error(t, err, "finalization phase should reject 'design' output type")
}

// TestConfigureTransitions_ReturnsNonNilBuilder verifies that configureTransitions
// returns a non-nil builder for chaining.
func TestConfigureTransitions_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("design")
	result := configureTransitions(builder)
	assert.NotNil(t, result, "configureTransitions should return non-nil builder")
}

// TestConfigureEventDeterminers_ReturnsNonNilBuilder verifies that configureEventDeterminers
// returns a non-nil builder for chaining.
func TestConfigureEventDeterminers_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("design")
	result := configureEventDeterminers(builder)
	assert.NotNil(t, result, "configureEventDeterminers should return non-nil builder")
}

// TestConfigureTransitions_InitialState verifies that initial state is set to Active.
func TestConfigureTransitions_InitialState(t *testing.T) {
	config := NewDesignProjectConfig()

	// Verify initial state is Active
	assert.Equal(t, string(Active), config.InitialState(), "initial state should be Active")
}

// TestTransitions_ActiveToFinalizing_GuardBlocksWhenDocumentsNotApproved tests that
// the Active -> Finalizing transition is blocked when documents are not all approved.
func TestTransitions_ActiveToFinalizing_GuardBlocksWhenDocumentsNotApproved(t *testing.T) {
	config := NewDesignProjectConfig()

	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize project
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Build state machine
	machine := config.BuildProjectMachine(proj, project.State(Active))

	// Add a task that is not completed
	phase := proj.Phases["design"]
	phase.Tasks = []projschema.TaskState{
		{
			Id:     "task-1",
			Name:   "Design Document",
			Status: "in_progress",
		},
	}
	proj.Phases["design"] = phase

	// Should not be able to advance
	canFire := machine.CanFire(project.Event(EventCompleteDesign))
	require.NoError(t, err)
	assert.False(t, canFire, "should not be able to fire EventCompleteDesign when documents not approved")
}

// TestTransitions_ActiveToFinalizing_GuardAllowsWhenDocumentsApproved tests that
// the Active -> Finalizing transition is allowed when all documents are approved.
func TestTransitions_ActiveToFinalizing_GuardAllowsWhenDocumentsApproved(t *testing.T) {
	config := NewDesignProjectConfig()

	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize project
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Build state machine
	machine := config.BuildProjectMachine(proj, project.State(Active))

	// Add a completed task
	phase := proj.Phases["design"]
	phase.Tasks = []projschema.TaskState{
		{
			Id:     "task-1",
			Name:   "Design Document",
			Status: "completed",
		},
	}
	proj.Phases["design"] = phase

	// Should be able to advance
	canFire := machine.CanFire(project.Event(EventCompleteDesign))
	require.NoError(t, err)
	assert.True(t, canFire, "should be able to fire EventCompleteDesign when documents approved")
}

// TestTransitions_FinalizingToCompleted_GuardBlocksWhenTasksIncomplete tests that
// the Finalizing -> Completed transition is blocked when finalization tasks are incomplete.
func TestTransitions_FinalizingToCompleted_GuardBlocksWhenTasksIncomplete(t *testing.T) {
	config := NewDesignProjectConfig()

	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize project
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Build state machine and transition to Finalizing
	machine := config.BuildProjectMachine(proj, project.State(Active))

	// First transition to Finalizing
	phase := proj.Phases["design"]
	phase.Tasks = []projschema.TaskState{
		{Id: "task-1", Name: "Design Document", Status: "completed"},
	}
	proj.Phases["design"] = phase

	err = machine.Fire(project.Event(EventCompleteDesign))
	require.NoError(t, err)

	// Add incomplete finalization task
	finalizationPhase := proj.Phases["finalization"]
	finalizationPhase.Tasks = []projschema.TaskState{
		{
			Id:     "finalization-1",
			Name:   "Move documents",
			Status: "in_progress",
		},
	}
	proj.Phases["finalization"] = finalizationPhase

	// Should not be able to complete
	canFire := machine.CanFire(project.Event(EventCompleteFinalization))
	require.NoError(t, err)
	assert.False(t, canFire, "should not be able to fire EventCompleteFinalization when tasks incomplete")
}

// TestTransitions_FinalizingToCompleted_GuardAllowsWhenTasksComplete tests that
// the Finalizing -> Completed transition is allowed when all finalization tasks are complete.
func TestTransitions_FinalizingToCompleted_GuardAllowsWhenTasksComplete(t *testing.T) {
	config := NewDesignProjectConfig()

	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "design",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize project
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Build state machine and transition to Finalizing
	machine := config.BuildProjectMachine(proj, project.State(Active))

	// First transition to Finalizing
	phase := proj.Phases["design"]
	phase.Tasks = []projschema.TaskState{
		{Id: "task-1", Name: "Design Document", Status: "completed"},
	}
	proj.Phases["design"] = phase

	err = machine.Fire(project.Event(EventCompleteDesign))
	require.NoError(t, err)

	// Add completed finalization task
	finalizationPhase := proj.Phases["finalization"]
	finalizationPhase.Tasks = []projschema.TaskState{
		{
			Id:     "finalization-1",
			Name:   "Move documents",
			Status: "completed",
		},
	}
	proj.Phases["finalization"] = finalizationPhase

	// Should be able to complete
	canFire := machine.CanFire(project.Event(EventCompleteFinalization))
	require.NoError(t, err)
	assert.True(t, canFire, "should be able to fire EventCompleteFinalization when tasks complete")
}
