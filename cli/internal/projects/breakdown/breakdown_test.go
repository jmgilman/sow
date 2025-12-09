package breakdown

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

// TestPackageRegistration verifies that the breakdown project type is registered
// with the global registry during package initialization.
func TestPackageRegistration(t *testing.T) {
	// The init() function should have registered "breakdown" type
	config, exists := state.Registry["breakdown"]
	assert.True(t, exists, "breakdown project type should be registered")
	assert.NotNil(t, config, "breakdown config should not be nil")
}

// TestNewBreakdownProjectConfig verifies that NewBreakdownProjectConfig
// returns a valid configuration with all required components.
func TestNewBreakdownProjectConfig(t *testing.T) {
	config := NewBreakdownProjectConfig()

	require.NotNil(t, config, "config should not be nil")
	assert.NotNil(t, config, "NewBreakdownProjectConfig should return non-nil config")
}

// TestConfigurePhases_ReturnsNonNilBuilder verifies that configurePhases
// returns a non-nil builder for chaining.
func TestConfigurePhases_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	result := configurePhases(builder)
	assert.NotNil(t, result, "configurePhases should return non-nil builder")
}

// TestConfigureTransitions_ReturnsNonNilBuilder verifies that configureTransitions
// returns a non-nil builder for chaining.
func TestConfigureTransitions_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	result := configureTransitions(builder)
	assert.NotNil(t, result, "configureTransitions should return non-nil builder")
}

// TestConfigureEventDeterminers_ReturnsNonNilBuilder verifies that configureEventDeterminers
// returns a non-nil builder for chaining.
func TestConfigureEventDeterminers_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	result := configureEventDeterminers(builder)
	assert.NotNil(t, result, "configureEventDeterminers should return non-nil builder")
}

// TestConfigurePrompts_ReturnsNonNilBuilder verifies that configurePrompts
// returns a non-nil builder for chaining.
func TestConfigurePrompts_ReturnsNonNilBuilder(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	result := configurePrompts(builder)
	assert.NotNil(t, result, "configurePrompts should return non-nil builder")
}

// TestConfigurePhases_SingleBreakdownPhase verifies that configurePhases
// sets up exactly one phase named "breakdown" with correct configuration.
func TestConfigurePhases_SingleBreakdownPhase(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	config := builder.Build()

	// Verify breakdown phase supports tasks
	assert.True(t, config.PhaseSupportsTasks("breakdown"), "breakdown phase should support tasks")

	// Verify only one phase exists
	phases := config.GetTaskSupportingPhases()
	assert.Len(t, phases, 1, "should have exactly one task-supporting phase")
	assert.Equal(t, "breakdown", phases[0], "task-supporting phase should be breakdown")
}

// TestConfigurePhases_NoFinalizationPhase verifies that breakdown project
// does not create a finalization phase (unlike design/exploration projects).
func TestConfigurePhases_NoFinalizationPhase(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	config := builder.Build()

	// Only one task-supporting phase should exist
	phases := config.GetTaskSupportingPhases()
	assert.Len(t, phases, 1, "should have exactly one phase")
	assert.False(t, config.PhaseSupportsTasks("finalization"), "should not have finalization phase")
}

// TestConfigurePhases_CorrectStartEndStates verifies that the breakdown
// phase has the correct start and end states.
func TestConfigurePhases_CorrectStartEndStates(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	config := builder.Build()

	// Verify breakdown phase is associated with Discovery state
	assert.True(t, config.IsPhaseStartState("breakdown", sdkstate.State(Discovery)), "Discovery should be start state")
	assert.True(t, config.IsPhaseEndState("breakdown", sdkstate.State(Publishing)), "Publishing should be end state")
}

// TestConfigurePhases_GetDefaultTaskPhase verifies that breakdown is returned
// as the default task phase when in Discovery or Active state.
func TestConfigurePhases_GetDefaultTaskPhase(t *testing.T) {
	builder := project.NewProjectTypeConfigBuilder("breakdown")
	builder = configurePhases(builder)
	config := builder.Build()

	// When in Discovery state, should return breakdown as default
	defaultPhase := config.GetDefaultTaskPhase(sdkstate.State(Discovery))
	assert.Equal(t, "breakdown", defaultPhase, "default task phase should be breakdown in Discovery state")

	// When in Active state, should return breakdown as default
	defaultPhase = config.GetDefaultTaskPhase(sdkstate.State(Active))
	assert.Equal(t, "breakdown", defaultPhase, "default task phase should be breakdown in Active state")
}

// TestInitializeBreakdownProject_CreatesBreakdownPhase verifies that
// initializeBreakdownProject creates the breakdown phase correctly.
func TestInitializeBreakdownProject_CreatesBreakdownPhase(t *testing.T) {
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)

	require.NoError(t, err)
	phase, exists := proj.Phases["breakdown"]
	assert.True(t, exists, "breakdown phase should exist")
	assert.Equal(t, "in_progress", phase.Status)
	assert.True(t, phase.Enabled)
	assert.Equal(t, now, phase.Created_at)
	assert.Equal(t, now, phase.Started_at, "started_at should be set since phase is in_progress")
	assert.NotNil(t, phase.Inputs)
	assert.NotNil(t, phase.Outputs)
	assert.NotNil(t, phase.Tasks)
	assert.NotNil(t, phase.Metadata)
}

// TestInitializeBreakdownProject_InProgressStatus verifies that breakdown
// phase starts in "in_progress" status (state machine starts in Discovery state).
func TestInitializeBreakdownProject_InProgressStatus(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.Equal(t, "in_progress", phase.Status, "breakdown phase should start in in_progress status")
}

// TestInitializeBreakdownProject_EnabledTrue verifies that breakdown
// phase starts enabled (not disabled).
func TestInitializeBreakdownProject_EnabledTrue(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.True(t, phase.Enabled, "breakdown phase should start enabled")
}

// TestInitializeBreakdownProject_NilInputs verifies that initialization
// handles nil initialInputs safely without panic.
func TestInitializeBreakdownProject_NilInputs(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.NotNil(t, phase.Inputs)
	assert.Empty(t, phase.Inputs)
}

// TestInitializeBreakdownProject_EmptyInputsMap verifies that initialization
// handles empty initialInputs map correctly.
func TestInitializeBreakdownProject_EmptyInputsMap(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, map[string][]projschema.ArtifactState{})
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.NotNil(t, phase.Inputs)
	assert.Empty(t, phase.Inputs)
}

// TestInitializeBreakdownProject_WithInputs verifies that initialization
// correctly extracts inputs for breakdown phase from initialInputs map.
func TestInitializeBreakdownProject_WithInputs(t *testing.T) {
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	inputs := map[string][]projschema.ArtifactState{
		"breakdown": {
			{
				Type:       "design",
				Path:       "designs/auth.md",
				Approved:   true,
				Created_at: now,
			},
			{
				Type:       "adr",
				Path:       "adrs/001-jwt-tokens.md",
				Approved:   true,
				Created_at: now,
			},
		},
	}

	err := initializeBreakdownProject(proj, inputs)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.Len(t, phase.Inputs, 2)
	assert.Equal(t, "designs/auth.md", phase.Inputs[0].Path)
	assert.Equal(t, "adrs/001-jwt-tokens.md", phase.Inputs[1].Path)
}

// TestInitializeBreakdownProject_EmptyOutputsAndTasks verifies that
// outputs and tasks start as empty slices.
func TestInitializeBreakdownProject_EmptyOutputsAndTasks(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.NotNil(t, phase.Outputs)
	assert.Empty(t, phase.Outputs)
	assert.NotNil(t, phase.Tasks)
	assert.Empty(t, phase.Tasks)
}

// TestInitializeBreakdownProject_EmptyMetadata verifies that
// metadata starts as an empty map.
func TestInitializeBreakdownProject_EmptyMetadata(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.NotNil(t, phase.Metadata)
	assert.Empty(t, phase.Metadata)
}

// TestInitializeBreakdownProject_NoFinalizationPhase verifies that
// initialization does NOT create a finalization phase.
func TestInitializeBreakdownProject_NoFinalizationPhase(t *testing.T) {
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: time.Now(),
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	assert.Len(t, proj.Phases, 1, "should have exactly one phase")
	assert.Contains(t, proj.Phases, "breakdown")
	assert.NotContains(t, proj.Phases, "finalization")
}

// TestInitializeBreakdownProject_UsesProjectCreatedAt verifies that
// phase Created_at uses the project's Created_at timestamp.
func TestInitializeBreakdownProject_UsesProjectCreatedAt(t *testing.T) {
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Type:       "breakdown",
			Created_at: now,
			Phases:     make(map[string]projschema.PhaseState),
		},
	}

	err := initializeBreakdownProject(proj, nil)
	require.NoError(t, err)

	phase := proj.Phases["breakdown"]
	assert.Equal(t, now, phase.Created_at)
}

// ========== Transition Configuration Tests ==========

// TestConfigureTransitions_SetsInitialState verifies that configureTransitions
// sets the initial state to Discovery.
func TestConfigureTransitions_SetsInitialState(t *testing.T) {
	config := NewBreakdownProjectConfig()

	assert.Equal(t, sdkstate.State(Discovery), config.InitialState(), "initial state should be Discovery")
}

// TestActiveToPublishing_GuardBlocksWhenTasksPending verifies that the guard
// blocks the Active → Publishing transition when work units are not approved.
func TestActiveToPublishing_GuardBlocksWhenTasksPending(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add a pending work unit task
	addWorkUnit(t, proj, "001", "Feature A", "pending")

	// Verify guard blocks transition
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.False(t, can, "guard should block with pending tasks")

	// Attempt to fire should fail
	err = machine.Fire(sdkstate.Event(EventBeginPublishing))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all work units approved and dependencies valid")
}

// TestActiveToPublishing_GuardAllowsWhenAllApproved verifies that the guard
// allows the transition when all work units are approved and dependencies valid.
func TestActiveToPublishing_GuardAllowsWhenAllApproved(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add completed work units
	addWorkUnit(t, proj, "001", "Feature A", "completed")
	addWorkUnit(t, proj, "002", "Feature B", "completed")

	// Verify guard allows transition
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.True(t, can, "guard should allow with all tasks completed")
}

// TestActiveToPublishing_OnEntryUpdatesPhaseStatus verifies that the OnEntry
// action updates the breakdown phase status to "publishing".
func TestActiveToPublishing_OnEntryUpdatesPhaseStatus(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add completed work unit
	addWorkUnit(t, proj, "001", "Feature A", "completed")

	// Fire transition
	err := machine.Fire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)

	// Verify state changed
	assert.Equal(t, sdkstate.State(Publishing), machine.State())

	// Verify phase status updated
	phase := proj.Phases["breakdown"]
	assert.Equal(t, "publishing", phase.Status)
}

// TestPublishingToCompleted_GuardBlocksWhenTasksUnpublished verifies that
// the guard blocks the Publishing → Completed transition when work units
// are not published.
func TestPublishingToCompleted_GuardBlocksWhenTasksUnpublished(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Setup: Get to Publishing state
	addWorkUnit(t, proj, "001", "Feature A", "completed")
	err := machine.Fire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)

	// Verify guard blocks transition (task not published)
	can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.False(t, can, "guard should block with unpublished tasks")

	// Attempt to fire should fail
	err = machine.Fire(sdkstate.Event(EventCompleteBreakdown))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all work units published")
}

// TestPublishingToCompleted_GuardAllowsWhenAllPublished verifies that the
// guard allows the transition when all work units are published.
func TestPublishingToCompleted_GuardAllowsWhenAllPublished(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Setup: Get to Publishing state
	addWorkUnit(t, proj, "001", "Feature A", "completed")
	err := machine.Fire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)

	// Mark task as published
	markWorkUnitPublished(t, proj, "001")

	// Verify guard allows transition
	can, err := machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.True(t, can, "guard should allow with all tasks published")
}

// ========== Event Determiner Tests ==========

// TestEventDeterminers_ActiveState verifies that the event determiner for
// Active state returns EventBeginPublishing.
func TestEventDeterminers_ActiveState(t *testing.T) {
	config := NewBreakdownProjectConfig()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Statechart: projschema.StatechartState{
				Current_state: string(Active),
			},
		},
	}

	// Use DetermineEvent to get the event for current state
	event, err := config.DetermineEvent(proj)
	require.NoError(t, err)
	assert.Equal(t, sdkstate.Event(EventBeginPublishing), event)
}

// TestEventDeterminers_PublishingState verifies that the event determiner for
// Publishing state returns EventCompleteBreakdown.
func TestEventDeterminers_PublishingState(t *testing.T) {
	config := NewBreakdownProjectConfig()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Statechart: projschema.StatechartState{
				Current_state: string(Publishing),
			},
		},
	}

	// Use DetermineEvent to get the event for current state
	event, err := config.DetermineEvent(proj)
	require.NoError(t, err)
	assert.Equal(t, sdkstate.Event(EventCompleteBreakdown), event)
}

// ========== Full Lifecycle Integration Tests ==========

// TestBreakdownLifecycle_FullWorkflow verifies the complete breakdown workflow
// from Discovery through Active, Publishing to Completed.
func TestBreakdownLifecycle_FullWorkflow(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)

	// Verify initial state is Discovery
	assert.Equal(t, sdkstate.State(Discovery), machine.State())
	phase := proj.Phases["breakdown"]
	assert.Equal(t, "in_progress", phase.Status)

	// Transition to Active
	transitionToActive(t, proj, machine, config)
	assert.Equal(t, sdkstate.State(Active), machine.State())
	phase = proj.Phases["breakdown"]
	assert.Equal(t, "active", phase.Status)

	// Add and complete work units
	addWorkUnit(t, proj, "001", "Feature A", "completed")
	addWorkUnit(t, proj, "002", "Feature B", "completed")

	// Transition to Publishing
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.True(t, can)

	err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventBeginPublishing), proj)
	require.NoError(t, err)
	assert.Equal(t, sdkstate.State(Publishing), machine.State())

	phase = proj.Phases["breakdown"]
	assert.Equal(t, "publishing", phase.Status)

	// Publish work units
	markWorkUnitPublished(t, proj, "001")
	markWorkUnitPublished(t, proj, "002")

	// Transition to Completed
	can, err = machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.True(t, can)

	err = config.FireWithPhaseUpdates(machine, sdkstate.Event(EventCompleteBreakdown), proj)
	require.NoError(t, err)
	assert.Equal(t, sdkstate.State(Completed), machine.State())

	// Verify phase completed
	phase = proj.Phases["breakdown"]
	assert.Equal(t, "completed", phase.Status)
	assert.False(t, phase.Completed_at.IsZero())
}

// TestBreakdownLifecycle_GuardsPreventInvalidTransitions verifies that
// guards properly block transitions when conditions are not met.
func TestBreakdownLifecycle_GuardsPreventInvalidTransitions(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Cannot transition to Publishing without work units
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.False(t, can, "should not allow transition without work units")

	// Add pending work unit
	addWorkUnit(t, proj, "001", "Feature A", "pending")

	// Cannot transition with pending work unit
	can, err = machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.False(t, can, "should not allow transition with pending work unit")

	// Complete work unit
	updateWorkUnitStatus(t, proj, "001", "completed")

	// Now can transition
	can, err = machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.True(t, can, "should allow transition with completed work unit")

	// Transition to Publishing
	err = machine.Fire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)

	// Cannot complete without publishing
	can, err = machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.False(t, can, "should not allow completion without publishing")

	// Mark as published
	markWorkUnitPublished(t, proj, "001")

	// Now can complete
	can, err = machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.True(t, can, "should allow completion after publishing")
}

// TestBreakdownLifecycle_WithAbandonedTasks verifies that abandoned tasks
// don't block transitions (as long as there's at least one completed task).
func TestBreakdownLifecycle_WithAbandonedTasks(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add mix of completed and abandoned tasks
	addWorkUnit(t, proj, "001", "Feature A", "completed")
	addWorkUnit(t, proj, "002", "Feature B", "abandoned")
	addWorkUnit(t, proj, "003", "Feature C", "completed")

	// Should be able to transition (abandoned tasks are ignored)
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.True(t, can, "should allow transition with abandoned tasks")

	err = machine.Fire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)

	// Publish only completed tasks
	markWorkUnitPublished(t, proj, "001")
	markWorkUnitPublished(t, proj, "003")

	// Should be able to complete
	can, err = machine.CanFire(sdkstate.Event(EventCompleteBreakdown))
	require.NoError(t, err)
	assert.True(t, can, "should allow completion with abandoned tasks")
}

// TestBreakdownLifecycle_WithDependencies verifies that the guard validates
// dependencies before allowing transition to Publishing.
func TestBreakdownLifecycle_WithDependencies(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add tasks with dependencies
	addWorkUnitWithDeps(t, proj, "001", "Foundation", "completed", []string{})
	addWorkUnitWithDeps(t, proj, "002", "Feature A", "completed", []string{"001"})
	addWorkUnitWithDeps(t, proj, "003", "Feature B", "completed", []string{"001"})

	// Valid DAG, should allow transition
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.True(t, can, "should allow transition with valid dependencies")
}

// TestBreakdownLifecycle_CyclicDependenciesBlock verifies that cyclic
// dependencies prevent transition to Publishing.
func TestBreakdownLifecycle_CyclicDependenciesBlock(t *testing.T) {
	proj, machine, config := setupBreakdownProject(t)
	transitionToActive(t, proj, machine, config)

	// Add tasks with cyclic dependency
	addWorkUnitWithDeps(t, proj, "001", "Feature A", "completed", []string{"002"})
	addWorkUnitWithDeps(t, proj, "002", "Feature B", "completed", []string{"001"})

	// Cyclic dependency should block transition
	can, err := machine.CanFire(sdkstate.Event(EventBeginPublishing))
	require.NoError(t, err)
	assert.False(t, can, "should block transition with cyclic dependencies")
}

// ========== Test Helper Functions ==========

// setupBreakdownProject creates a test breakdown project and state machine.
func setupBreakdownProject(t *testing.T) (*state.Project, *sdkstate.Machine, *project.ProjectTypeConfig) {
	t.Helper()

	// Create project
	now := time.Now()
	proj := &state.Project{
		ProjectState: projschema.ProjectState{
			Name:        "test-breakdown",
			Branch:      "breakdown/test",
			Type:        "breakdown",
			Description: "Test breakdown project",
			Created_at:  now,
			Updated_at:  now,
			Phases:      make(map[string]projschema.PhaseState),
			Statechart: projschema.StatechartState{
				Current_state: string(Discovery),
				Updated_at:    now,
			},
		},
	}

	// Initialize phases using the config's initializer
	config := NewBreakdownProjectConfig()
	err := initializeBreakdownProject(proj, nil)
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Build state machine
	machine := config.BuildMachine(proj, sdkstate.State(Discovery))
	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}

	return proj, machine, config
}

// transitionToActive transitions the project from Discovery to Active state.
// Adds an approved discovery artifact and fires EventBeginActive.
// This is used by tests that need to start in Active state.
func transitionToActive(t *testing.T, proj *state.Project, machine *sdkstate.Machine, _ *project.ProjectTypeConfig) {
	t.Helper()

	// Add approved discovery artifact
	phase, exists := proj.Phases["breakdown"]
	if !exists {
		t.Fatal("breakdown phase not found")
	}

	artifact := projschema.ArtifactState{
		Type:       "discovery",
		Path:       "project/discovery/analysis.md",
		Created_at: time.Now(),
		Approved:   true,
	}

	phase.Outputs = append(phase.Outputs, artifact)
	proj.Phases["breakdown"] = phase

	// Fire EventBeginActive to transition to Active
	err := machine.Fire(sdkstate.Event(EventBeginActive))
	if err != nil {
		t.Fatalf("failed to transition to Active: %v", err)
	}

	// Verify we're in Active state
	if machine.State() != sdkstate.State(Active) {
		t.Fatalf("expected Active state, got %s", machine.State())
	}
}

// addWorkUnit adds a work unit task to the breakdown phase.
func addWorkUnit(t *testing.T, p *state.Project, id, name, status string) {
	t.Helper()
	addWorkUnitWithDeps(t, p, id, name, status, []string{})
}

// addWorkUnitWithDeps adds a work unit task with dependencies to the breakdown phase.
func addWorkUnitWithDeps(t *testing.T, p *state.Project, id, name, status string, deps []string) {
	t.Helper()

	phase := p.Phases["breakdown"]

	metadata := make(map[string]interface{})
	if len(deps) > 0 {
		// Convert []string to []interface{} for metadata
		depsInterface := make([]interface{}, len(deps))
		for i, dep := range deps {
			depsInterface[i] = dep
		}
		metadata["dependencies"] = depsInterface
	}

	task := projschema.TaskState{
		Id:         id,
		Name:       name,
		Status:     status,
		Created_at: time.Now(),
		Metadata:   metadata,
	}

	phase.Tasks = append(phase.Tasks, task)
	p.Phases["breakdown"] = phase
}

// updateWorkUnitStatus updates the status of a work unit task.
func updateWorkUnitStatus(t *testing.T, p *state.Project, id, status string) {
	t.Helper()

	phase := p.Phases["breakdown"]
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == id {
			phase.Tasks[i].Status = status
			p.Phases["breakdown"] = phase
			return
		}
	}
	t.Fatalf("task %s not found", id)
}

// markWorkUnitPublished marks a work unit task as published.
func markWorkUnitPublished(t *testing.T, p *state.Project, id string) {
	t.Helper()

	phase := p.Phases["breakdown"]
	for i := range phase.Tasks {
		if phase.Tasks[i].Id == id {
			if phase.Tasks[i].Metadata == nil {
				phase.Tasks[i].Metadata = make(map[string]interface{})
			}
			phase.Tasks[i].Metadata["published"] = true
			p.Phases["breakdown"] = phase
			return
		}
	}
	t.Fatalf("task %s not found", id)
}
