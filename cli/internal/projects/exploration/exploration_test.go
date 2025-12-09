package exploration

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageRegistration verifies that the exploration project type is registered
// with the global registry during package initialization.
func TestPackageRegistration(t *testing.T) {
	// The init() function should have registered "exploration" type
	config, exists := state.Registry["exploration"]
	assert.True(t, exists, "exploration project type should be registered")
	assert.NotNil(t, config, "exploration config should not be nil")
}

// TestNewExplorationProjectConfig verifies that NewExplorationProjectConfig
// returns a valid configuration with all required components.
func TestNewExplorationProjectConfig(t *testing.T) {
	config := NewExplorationProjectConfig()

	require.NotNil(t, config, "config should not be nil")

	// Verify the configuration is properly built
	// Note: We can't directly test internal fields without exposing them,
	// but we can verify the config is not nil and properly constructed
	assert.NotNil(t, config, "NewExplorationProjectConfig should return non-nil config")
}

// TestInitializeExplorationProject_CreatesPhases verifies that the initializer
// creates both exploration and finalization phases with correct structure.
func TestInitializeExplorationProject_CreatesPhases(t *testing.T) {
	// Create a minimal project for testing
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Call initializer with no initial inputs
	err := initializeExplorationProject(project, nil)
	require.NoError(t, err, "initialization should not error")

	// Verify both phases were created
	assert.Len(t, project.Phases, 2, "should create exactly 2 phases")
	assert.Contains(t, project.Phases, "exploration", "should create exploration phase")
	assert.Contains(t, project.Phases, "finalization", "should create finalization phase")
}

// TestInitializeExplorationProject_ExplorationPhaseActive verifies that the
// exploration phase starts in active status with enabled=true.
func TestInitializeExplorationProject_ExplorationPhaseActive(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeExplorationProject(project, nil)
	require.NoError(t, err)

	explorationPhase := project.Phases["exploration"]

	// Exploration phase should start active
	assert.Equal(t, "active", explorationPhase.Status, "exploration phase should have status='active'")
	assert.True(t, explorationPhase.Enabled, "exploration phase should be enabled")

	// Verify timestamps
	assert.Equal(t, now, explorationPhase.Created_at, "exploration phase should use project creation time")

	// Verify collections are initialized
	assert.NotNil(t, explorationPhase.Inputs, "exploration phase inputs should be initialized")
	assert.NotNil(t, explorationPhase.Outputs, "exploration phase outputs should be initialized")
	assert.NotNil(t, explorationPhase.Tasks, "exploration phase tasks should be initialized")
	assert.NotNil(t, explorationPhase.Metadata, "exploration phase metadata should be initialized")

	// Verify collections are empty
	assert.Empty(t, explorationPhase.Inputs, "exploration phase inputs should be empty")
	assert.Empty(t, explorationPhase.Outputs, "exploration phase outputs should be empty")
	assert.Empty(t, explorationPhase.Tasks, "exploration phase tasks should be empty")
}

// TestInitializeExplorationProject_FinalizationPhasePending verifies that the
// finalization phase starts in pending status with enabled=false.
func TestInitializeExplorationProject_FinalizationPhasePending(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeExplorationProject(project, nil)
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

// TestInitializeExplorationProject_WithInitialInputs verifies that the initializer
// properly handles initial inputs provided during project creation.
func TestInitializeExplorationProject_WithInitialInputs(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Create test initial inputs
	explorationInputs := []projschema.ArtifactState{
		{
			Type: "question",
			Path: "/test/question.md",
		},
	}

	finalizationInputs := []projschema.ArtifactState{
		{
			Type: "report",
			Path: "/test/report.md",
		},
	}

	initialInputs := map[string][]projschema.ArtifactState{
		"exploration":  explorationInputs,
		"finalization": finalizationInputs,
	}

	// Call initializer with initial inputs
	err := initializeExplorationProject(project, initialInputs)
	require.NoError(t, err)

	// Verify exploration phase received its inputs
	explorationPhase := project.Phases["exploration"]
	assert.Len(t, explorationPhase.Inputs, 1, "exploration phase should have 1 input")
	assert.Equal(t, "question", explorationPhase.Inputs[0].Type)

	// Verify finalization phase received its inputs
	finalizationPhase := project.Phases["finalization"]
	assert.Len(t, finalizationPhase.Inputs, 1, "finalization phase should have 1 input")
	assert.Equal(t, "report", finalizationPhase.Inputs[0].Type)
}

// TestInitializeExplorationProject_NoInitialInputs verifies that the initializer
// works correctly when initialInputs is nil.
func TestInitializeExplorationProject_NoInitialInputs(t *testing.T) {
	now := time.Now()
	project := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Call initializer with nil inputs
	err := initializeExplorationProject(project, nil)
	require.NoError(t, err)

	// Both phases should have empty inputs
	explorationPhase := project.Phases["exploration"]
	assert.Empty(t, explorationPhase.Inputs, "exploration phase inputs should be empty")

	finalizationPhase := project.Phases["finalization"]
	assert.Empty(t, finalizationPhase.Inputs, "finalization phase inputs should be empty")
}

// TestConfigurePhases_ReturnsNonNilBuilder verifies that configurePhases
// returns a non-nil builder for chaining.
func TestConfigurePhases_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("exploration")
	result := configurePhases(builder)
	assert.NotNil(t, result, "configurePhases should return non-nil builder")
}

// TestConfigurePhases_ExplorationPhaseTaskSupport verifies that the
// exploration phase supports tasks.
func TestConfigurePhases_ExplorationPhaseTaskSupport(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Exploration phase should support tasks
	assert.True(t, config.PhaseSupportsTasks("exploration"),
		"exploration phase should support tasks")
}

// TestConfigurePhases_FinalizationPhaseNoTaskSupport verifies that the
// finalization phase does not support tasks.
func TestConfigurePhases_FinalizationPhaseNoTaskSupport(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Finalization phase should NOT support tasks
	assert.False(t, config.PhaseSupportsTasks("finalization"),
		"finalization phase should not support tasks")
}

// TestConfigurePhases_GetTaskSupportingPhases verifies that only the
// exploration phase is returned as task-supporting.
func TestConfigurePhases_GetTaskSupportingPhases(t *testing.T) {
	config := NewExplorationProjectConfig()

	phases := config.GetTaskSupportingPhases()

	// Should have exactly one task-supporting phase
	assert.Len(t, phases, 1, "should have exactly one task-supporting phase")
	assert.Contains(t, phases, "exploration", "task-supporting phases should include exploration")
}

// TestConfigurePhases_GetDefaultTaskPhase verifies that the default
// task phase is correctly determined based on state.
func TestConfigurePhases_GetDefaultTaskPhase(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Active state should map to exploration phase
	phase := config.GetDefaultTaskPhase(sdkstate.State(Active))
	assert.Equal(t, "exploration", phase, "Active state should default to exploration phase")

	// Summarizing state should map to exploration phase
	phase = config.GetDefaultTaskPhase(sdkstate.State(Summarizing))
	assert.Equal(t, "exploration", phase, "Summarizing state should default to exploration phase")

	// Finalizing state should fallback to exploration (finalization doesn't support tasks)
	phase = config.GetDefaultTaskPhase(sdkstate.State(Finalizing))
	assert.Equal(t, "exploration", phase, "Finalizing state should fallback to exploration phase")
}

// TestConfigurePhases_ExplorationOutputTypes verifies that the exploration
// phase allows "summary" and "findings" output types.
func TestConfigurePhases_ExplorationOutputTypes(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Create test project with exploration phase
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test "summary" output type is allowed
	summaryArtifact := projschema.ArtifactState{
		Type:     "summary",
		Path:     "/test/summary.md",
		Approved: true,
	}
	phase := proj.Phases["exploration"]
	phase.Outputs = []projschema.ArtifactState{summaryArtifact}
	proj.Phases["exploration"] = phase

	err = config.Validate(proj)
	assert.NoError(t, err, "exploration phase should allow 'summary' output type")

	// Test "findings" output type is allowed
	findingsArtifact := projschema.ArtifactState{
		Type:     "findings",
		Path:     "/test/findings.md",
		Approved: true,
	}
	phase.Outputs = []projschema.ArtifactState{findingsArtifact}
	proj.Phases["exploration"] = phase

	err = config.Validate(proj)
	assert.NoError(t, err, "exploration phase should allow 'findings' output type")

	// Test both output types together
	phase.Outputs = []projschema.ArtifactState{summaryArtifact, findingsArtifact}
	proj.Phases["exploration"] = phase

	err = config.Validate(proj)
	assert.NoError(t, err, "exploration phase should allow both 'summary' and 'findings' outputs")
}

// TestConfigurePhases_FinalizationOutputTypes verifies that the finalization
// phase allows "pr" output type.
func TestConfigurePhases_FinalizationOutputTypes(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Create test project with finalization phase
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
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

// TestConfigurePhases_ExplorationRejectsInvalidOutputs verifies that the
// exploration phase rejects output types not in the allowed list.
func TestConfigurePhases_ExplorationRejectsInvalidOutputs(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Create test project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test that "pr" output type is rejected in exploration phase
	prArtifact := projschema.ArtifactState{
		Type:     "pr",
		Path:     "/test/pr.md",
		Approved: true,
	}
	phase := proj.Phases["exploration"]
	phase.Outputs = []projschema.ArtifactState{prArtifact}
	proj.Phases["exploration"] = phase

	err = config.Validate(proj)
	assert.Error(t, err, "exploration phase should reject 'pr' output type")
}

// TestConfigurePhases_FinalizationRejectsInvalidOutputs verifies that the
// finalization phase rejects output types not in the allowed list.
func TestConfigurePhases_FinalizationRejectsInvalidOutputs(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Create test project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "exploration",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	// Initialize phases
	err := config.Initialize(proj, nil)
	require.NoError(t, err)

	// Test that "summary" output type is rejected in finalization phase
	summaryArtifact := projschema.ArtifactState{
		Type:     "summary",
		Path:     "/test/summary.md",
		Approved: true,
	}
	phase := proj.Phases["finalization"]
	phase.Outputs = []projschema.ArtifactState{summaryArtifact}
	proj.Phases["finalization"] = phase

	err = config.Validate(proj)
	assert.Error(t, err, "finalization phase should reject 'summary' output type")
}

// TestConfigureTransitions_ReturnsNonNilBuilder verifies that configureTransitions
// returns a non-nil builder for chaining.
func TestConfigureTransitions_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("exploration")
	result := configureTransitions(builder)
	assert.NotNil(t, result, "configureTransitions should return non-nil builder")
}

// TestConfigureEventDeterminers_ReturnsNonNilBuilder verifies that configureEventDeterminers
// returns a non-nil builder for chaining.
func TestConfigureEventDeterminers_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("exploration")
	result := configureEventDeterminers(builder)
	assert.NotNil(t, result, "configureEventDeterminers should return non-nil builder")
}

// TestConfigureTransitions_InitialState verifies that initial state is set to Active.
func TestConfigureTransitions_InitialState(t *testing.T) {
	config := NewExplorationProjectConfig()

	// Verify initial state is Active
	assert.Equal(t, sdkstate.State(Active), config.InitialState(), "initial state should be Active")
}
