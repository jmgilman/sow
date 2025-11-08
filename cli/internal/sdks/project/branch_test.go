package project

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projectschema "github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchOn(t *testing.T) {
	t.Run("sets discriminator function", func(t *testing.T) {
		bc := &BranchConfig{}
		discriminator := func(_ *state.Project) string {
			return "test_value"
		}

		BranchOn(discriminator)(bc)

		require.NotNil(t, bc.discriminator)
	})

	t.Run("discriminator is called correctly", func(t *testing.T) {
		bc := &BranchConfig{}
		discriminator := func(p *state.Project) string {
			// Examine project state
			if p.Name == "test_project" {
				return "pass"
			}
			return "fail"
		}

		BranchOn(discriminator)(bc)

		// Test discriminator can be called and examines project state
		project := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name: "test_project",
			},
		}
		result := bc.discriminator(project)
		assert.Equal(t, "pass", result)

		project2 := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name: "other_project",
			},
		}
		result2 := bc.discriminator(project2)
		assert.Equal(t, "fail", result2)
	})
}

//nolint:funlen // Test comprehensiveness requires length
func TestWhen(t *testing.T) {
	t.Run("creates branch path with value, event, and target state", func(t *testing.T) {
		bc := &BranchConfig{}

		When("pass",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
		)(bc)

		require.NotNil(t, bc.branches)
		require.Contains(t, bc.branches, "pass")

		path := bc.branches["pass"]
		require.NotNil(t, path)
		assert.Equal(t, "pass", path.value)
		assert.Equal(t, sdkstate.Event("test_event"), path.event)
		assert.Equal(t, sdkstate.State("test_state"), path.to)
	})

	t.Run("stores path in branches map", func(t *testing.T) {
		bc := &BranchConfig{}

		When("value1",
			sdkstate.Event("event1"),
			sdkstate.State("state1"),
		)(bc)

		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 1)
		assert.Contains(t, bc.branches, "value1")
	})

	t.Run("initializes branches map if nil", func(t *testing.T) {
		bc := &BranchConfig{
			branches: nil, // Explicitly nil
		}

		When("test",
			sdkstate.Event("event"),
			sdkstate.State("state"),
		)(bc)

		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 1)
	})

	t.Run("forwards WithDescription option", func(t *testing.T) {
		bc := &BranchConfig{}

		When("pass",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithDescription("Test description"),
		)(bc)

		path := bc.branches["pass"]
		require.NotNil(t, path)
		assert.Equal(t, "Test description", path.description)
	})

	t.Run("forwards WithGuard option", func(t *testing.T) {
		bc := &BranchConfig{}
		guardFunc := func(_ *state.Project) bool {
			return true
		}

		When("pass",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithGuard("test guard", guardFunc),
		)(bc)

		path := bc.branches["pass"]
		require.NotNil(t, path)
		assert.Equal(t, "test guard", path.guardTemplate.Description)
		require.NotNil(t, path.guardTemplate.Func)
		// Test that guard function works
		assert.True(t, path.guardTemplate.Func(&state.Project{}))
	})

	t.Run("forwards WithOnEntry option", func(t *testing.T) {
		bc := &BranchConfig{}
		entryAction := func(p *state.Project) error {
			p.Name = "modified"
			return nil
		}

		When("pass",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithOnEntry(entryAction),
		)(bc)

		path := bc.branches["pass"]
		require.NotNil(t, path)
		require.NotNil(t, path.onEntry)

		// Test that action works
		project := &state.Project{}
		err := path.onEntry(project)
		assert.NoError(t, err)
		assert.Equal(t, "modified", project.Name)
	})

	t.Run("forwards WithOnExit option", func(t *testing.T) {
		bc := &BranchConfig{}
		exitAction := func(p *state.Project) error {
			p.Name = "exited"
			return nil
		}

		When("pass",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithOnExit(exitAction),
		)(bc)

		path := bc.branches["pass"]
		require.NotNil(t, path)
		require.NotNil(t, path.onExit)

		// Test that action works
		project := &state.Project{}
		err := path.onExit(project)
		assert.NoError(t, err)
		assert.Equal(t, "exited", project.Name)
	})

	t.Run("forwards WithFailedPhase option", func(t *testing.T) {
		bc := &BranchConfig{}

		When("fail",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithFailedPhase("review"),
		)(bc)

		path := bc.branches["fail"]
		require.NotNil(t, path)
		assert.Equal(t, "review", path.failedPhase)
	})

	t.Run("multiple When clauses accumulate", func(t *testing.T) {
		bc := &BranchConfig{}

		When("pass",
			sdkstate.Event("event1"),
			sdkstate.State("state1"),
		)(bc)

		When("fail",
			sdkstate.Event("event2"),
			sdkstate.State("state2"),
		)(bc)

		When("pending",
			sdkstate.Event("event3"),
			sdkstate.State("state3"),
		)(bc)

		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 3)
		assert.Contains(t, bc.branches, "pass")
		assert.Contains(t, bc.branches, "fail")
		assert.Contains(t, bc.branches, "pending")
	})

	t.Run("duplicate value overwrites previous path", func(t *testing.T) {
		bc := &BranchConfig{}

		When("test",
			sdkstate.Event("event1"),
			sdkstate.State("state1"),
			WithDescription("First description"),
		)(bc)

		When("test",
			sdkstate.Event("event2"),
			sdkstate.State("state2"),
			WithDescription("Second description"),
		)(bc)

		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 1, "Should only have one path for duplicate value")

		path := bc.branches["test"]
		require.NotNil(t, path)
		// Last one wins
		assert.Equal(t, sdkstate.Event("event2"), path.event)
		assert.Equal(t, sdkstate.State("state2"), path.to)
		assert.Equal(t, "Second description", path.description)
	})
}

func TestBranchConfigIntegration(t *testing.T) {
	t.Run("BranchOn and When work together", func(t *testing.T) {
		bc := &BranchConfig{}

		discriminator := func(_ *state.Project) string {
			return "test_value"
		}
		BranchOn(discriminator)(bc)

		When("test_value",
			sdkstate.Event("test_event"),
			sdkstate.State("test_state"),
			WithDescription("Test branch"),
		)(bc)

		require.NotNil(t, bc.discriminator)
		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 1)

		path := bc.branches["test_value"]
		require.NotNil(t, path)
		assert.Equal(t, "test_value", path.value)
		assert.Equal(t, sdkstate.Event("test_event"), path.event)
		assert.Equal(t, sdkstate.State("test_state"), path.to)
		assert.Equal(t, "Test branch", path.description)
	})

	t.Run("complete binary branch configuration", func(t *testing.T) {
		bc := &BranchConfig{}

		// Set discriminator
		BranchOn(func(p *state.Project) string {
			if p.Name == "approved" {
				return "pass"
			}
			return "fail"
		})(bc)

		// Define pass branch
		When("pass",
			sdkstate.Event("review_pass"),
			sdkstate.State("finalize"),
			WithDescription("Review approved - proceed to finalization"),
		)(bc)

		// Define fail branch
		When("fail",
			sdkstate.Event("review_fail"),
			sdkstate.State("planning"),
			WithDescription("Review failed - return to planning for rework"),
			WithFailedPhase("review"),
		)(bc)

		// Verify configuration
		require.NotNil(t, bc.discriminator)
		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 2)

		// Test discriminator
		passProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name: "approved",
			},
		}
		assert.Equal(t, "pass", bc.discriminator(passProject))

		failProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name: "rejected",
			},
		}
		assert.Equal(t, "fail", bc.discriminator(failProject))

		// Verify pass path
		passPath := bc.branches["pass"]
		require.NotNil(t, passPath)
		assert.Equal(t, sdkstate.Event("review_pass"), passPath.event)
		assert.Equal(t, sdkstate.State("finalize"), passPath.to)
		assert.Equal(t, "Review approved - proceed to finalization", passPath.description)
		assert.Empty(t, passPath.failedPhase)

		// Verify fail path
		failPath := bc.branches["fail"]
		require.NotNil(t, failPath)
		assert.Equal(t, sdkstate.Event("review_fail"), failPath.event)
		assert.Equal(t, sdkstate.State("planning"), failPath.to)
		assert.Equal(t, "Review failed - return to planning for rework", failPath.description)
		assert.Equal(t, "review", failPath.failedPhase)
	})

	t.Run("N-way branch configuration", func(t *testing.T) {
		bc := &BranchConfig{}

		BranchOn(func(_ *state.Project) string {
			return "staging" // Simplified for test
		})(bc)

		When("staging",
			sdkstate.Event("deploy_staging"),
			sdkstate.State("deploying_staging"),
			WithDescription("Deploy to staging environment"),
		)(bc)

		When("production",
			sdkstate.Event("deploy_production"),
			sdkstate.State("deploying_production"),
			WithDescription("Deploy to production"),
		)(bc)

		When("canary",
			sdkstate.Event("deploy_canary"),
			sdkstate.State("deploying_canary"),
			WithDescription("Deploy to canary environment"),
		)(bc)

		// Verify configuration
		require.NotNil(t, bc.discriminator)
		require.NotNil(t, bc.branches)
		assert.Len(t, bc.branches, 3)
		assert.Contains(t, bc.branches, "staging")
		assert.Contains(t, bc.branches, "production")
		assert.Contains(t, bc.branches, "canary")
	})
}

func TestAddBranchGeneratesTransitions(t *testing.T) {
	t.Run("creates transitions for each When clause", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string { return "value1" }),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
			When("value2", sdkstate.Event("Event2"), sdkstate.State("State2")),
		)

		// Should have generated 2 transitions
		assert.Len(t, builder.transitions, 2)

		// Verify first transition
		assert.Equal(t, sdkstate.State("BranchState"), builder.transitions[0].From)
		assert.Equal(t, sdkstate.State("State1"), builder.transitions[0].To)
		assert.Equal(t, sdkstate.Event("Event1"), builder.transitions[0].Event)

		// Verify second transition
		assert.Equal(t, sdkstate.State("BranchState"), builder.transitions[1].From)
		assert.Equal(t, sdkstate.State("State2"), builder.transitions[1].To)
		assert.Equal(t, sdkstate.Event("Event2"), builder.transitions[1].Event)
	})

	t.Run("forwards transition options correctly", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		guardFunc := func(_ *state.Project) bool { return true }
		onEntryFunc := func(_ *state.Project) error { return nil }
		onExitFunc := func(_ *state.Project) error { return nil }

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string { return "value1" }),
			When("value1",
				sdkstate.Event("Event1"),
				sdkstate.State("State1"),
				WithDescription("Test description"),
				WithGuard("test guard", guardFunc),
				WithOnEntry(onEntryFunc),
				WithOnExit(onExitFunc),
				WithFailedPhase("test_phase"),
			),
		)

		require.Len(t, builder.transitions, 1)
		tc := builder.transitions[0]

		// Verify options were forwarded
		assert.Equal(t, "Test description", tc.description)
		assert.Equal(t, "test guard", tc.guardTemplate.Description)
		assert.NotNil(t, tc.guardTemplate.Func)
		assert.NotNil(t, tc.onEntry)
		assert.NotNil(t, tc.onExit)
		assert.Equal(t, "test_phase", tc.failedPhase)
	})
}

func TestAddBranchGeneratesOnAdvance(t *testing.T) {
	t.Run("creates event determiner using discriminator", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		discriminator := func(_ *state.Project) string {
			return "test_value"
		}

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(discriminator),
			When("test_value", sdkstate.Event("TestEvent"), sdkstate.State("NextState")),
		)

		// Should have OnAdvance registered
		assert.Contains(t, builder.onAdvance, sdkstate.State("BranchState"))
	})

	t.Run("determiner returns correct event for discriminator value", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(p *state.Project) string {
				// Return value based on project state (use phase metadata)
				if phase, ok := p.Phases["test_phase"]; ok {
					if val, ok := phase.Metadata["branch_value"].(string); ok {
						return val
					}
				}
				return "default"
			}),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
			When("value2", sdkstate.Event("Event2"), sdkstate.State("State2")),
			When("default", sdkstate.Event("EventDefault"), sdkstate.State("StateDefault")),
		)

		determiner := builder.onAdvance[sdkstate.State("BranchState")]
		require.NotNil(t, determiner)

		// Test with value1
		project1 := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"test_phase": {
						Metadata: map[string]interface{}{
							"branch_value": "value1",
						},
					},
				},
			},
		}
		event1, err1 := determiner(project1)
		require.NoError(t, err1)
		assert.Equal(t, sdkstate.Event("Event1"), event1)

		// Test with value2
		project2 := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"test_phase": {
						Metadata: map[string]interface{}{
							"branch_value": "value2",
						},
					},
				},
			},
		}
		event2, err2 := determiner(project2)
		require.NoError(t, err2)
		assert.Equal(t, sdkstate.Event("Event2"), event2)

		// Test with default
		project3 := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"test_phase": {
						Metadata: map[string]interface{}{},
					},
				},
			},
		}
		event3, err3 := determiner(project3)
		require.NoError(t, err3)
		assert.Equal(t, sdkstate.Event("EventDefault"), event3)
	})

	t.Run("determiner returns error for unmapped value", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string {
				return "unmapped_value"
			}),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
			When("value2", sdkstate.Event("Event2"), sdkstate.State("State2")),
		)

		determiner := builder.onAdvance[sdkstate.State("BranchState")]
		require.NotNil(t, determiner)

		project := &state.Project{}
		_, err := determiner(project)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no branch defined for discriminator value \"unmapped_value\"")
		assert.Contains(t, err.Error(), "available values:")
		assert.Contains(t, err.Error(), "\"value1\"")
		assert.Contains(t, err.Error(), "\"value2\"")
	})
}

func TestAddBranchBinary(t *testing.T) {
	t.Run("binary branch workflow", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("ReviewState"),
			BranchOn(func(p *state.Project) string {
				// Return "pass" or "fail" based on phase metadata
				if phase, ok := p.Phases["review"]; ok {
					if val, ok := phase.Metadata["review_result"].(string); ok {
						return val
					}
				}
				return "pass"
			}),
			When("pass",
				sdkstate.Event("PassEvent"),
				sdkstate.State("PassState"),
				WithDescription("Test passed"),
			),
			When("fail",
				sdkstate.Event("FailEvent"),
				sdkstate.State("FailState"),
				WithDescription("Test failed"),
			),
		)

		config := builder.Build()

		// Verify transitions were created
		require.Len(t, config.transitions, 2)

		// Test with "pass" discriminator value
		passProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"review": {
						Metadata: map[string]interface{}{
							"review_result": "pass",
						},
					},
				},
			},
		}
		passEvent, err := config.onAdvance[sdkstate.State("ReviewState")](passProject)
		require.NoError(t, err)
		assert.Equal(t, sdkstate.Event("PassEvent"), passEvent)

		// Test with "fail" discriminator value
		failProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"review": {
						Metadata: map[string]interface{}{
							"review_result": "fail",
						},
					},
				},
			},
		}
		failEvent, err := config.onAdvance[sdkstate.State("ReviewState")](failProject)
		require.NoError(t, err)
		assert.Equal(t, sdkstate.Event("FailEvent"), failEvent)
	})
}

func TestAddBranchNWay(t *testing.T) {
	t.Run("N-way branch workflow (3+ branches)", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("DeployState"),
			BranchOn(func(p *state.Project) string {
				if phase, ok := p.Phases["deployment"]; ok {
					if target, ok := phase.Metadata["target"].(string); ok {
						return target
					}
				}
				return "staging"
			}),
			When("staging", sdkstate.Event("DeployStaging"), sdkstate.State("StagingState")),
			When("production", sdkstate.Event("DeployProd"), sdkstate.State("ProdState")),
			When("canary", sdkstate.Event("DeployCanary"), sdkstate.State("CanaryState")),
		)

		config := builder.Build()

		// Verify 3 transitions were created
		require.Len(t, config.transitions, 3)

		// Test staging path
		stagingProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"deployment": {
						Metadata: map[string]interface{}{
							"target": "staging",
						},
					},
				},
			},
		}
		stagingEvent, err := config.onAdvance[sdkstate.State("DeployState")](stagingProject)
		require.NoError(t, err)
		assert.Equal(t, sdkstate.Event("DeployStaging"), stagingEvent)

		// Test production path
		prodProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"deployment": {
						Metadata: map[string]interface{}{
							"target": "production",
						},
					},
				},
			},
		}
		prodEvent, err := config.onAdvance[sdkstate.State("DeployState")](prodProject)
		require.NoError(t, err)
		assert.Equal(t, sdkstate.Event("DeployProd"), prodEvent)

		// Test canary path
		canaryProject := &state.Project{
			ProjectState: projectschema.ProjectState{
				Phases: map[string]projectschema.PhaseState{
					"deployment": {
						Metadata: map[string]interface{}{
							"target": "canary",
						},
					},
				},
			},
		}
		canaryEvent, err := config.onAdvance[sdkstate.State("DeployState")](canaryProject)
		require.NoError(t, err)
		assert.Equal(t, sdkstate.Event("DeployCanary"), canaryEvent)
	})
}

func TestAddBranchValidation(t *testing.T) {
	t.Run("panics if no discriminator provided", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("BranchState"),
				// No BranchOn call
				When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
			)
		})
	})

	t.Run("panics if no branch paths provided", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("BranchState"),
				BranchOn(func(_ *state.Project) string { return "test" }),
				// No When calls
			)
		})
	})
}

func TestAddBranchStoresBranchConfig(t *testing.T) {
	t.Run("stores branch config in builder", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string { return "value1" }),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
		)

		// Should have stored BranchConfig
		assert.Contains(t, builder.branches, sdkstate.State("BranchState"))
		assert.NotNil(t, builder.branches[sdkstate.State("BranchState")])
	})

	t.Run("branch config copied to built config", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string { return "value1" }),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
		)

		config := builder.Build()

		// Should have branches in config
		assert.Contains(t, config.branches, sdkstate.State("BranchState"))
		assert.NotNil(t, config.branches[sdkstate.State("BranchState")])
	})
}

func TestAddBranchChaining(t *testing.T) {
	t.Run("returns builder for method chaining", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		result := builder.AddBranch(
			sdkstate.State("BranchState"),
			BranchOn(func(_ *state.Project) string { return "value1" }),
			When("value1", sdkstate.Event("Event1"), sdkstate.State("State1")),
		)

		assert.Equal(t, builder, result, "AddBranch should return the same builder instance")
	})

	t.Run("can chain multiple calls", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("test").
			AddBranch(
				sdkstate.State("Branch1"),
				BranchOn(func(_ *state.Project) string { return "v1" }),
				When("v1", sdkstate.Event("E1"), sdkstate.State("S1")),
			).
			AddBranch(
				sdkstate.State("Branch2"),
				BranchOn(func(_ *state.Project) string { return "v2" }),
				When("v2", sdkstate.Event("E2"), sdkstate.State("S2")),
			).
			Build()

		// Should have transitions from both branches
		assert.Len(t, config.transitions, 2)
		assert.Len(t, config.branches, 2)
	})
}

// createMinimalTestProject creates a minimal project for testing.
func createMinimalTestProject(t *testing.T) *state.Project {
	t.Helper()

	return &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   "test",
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{},
			Statechart: projectschema.StatechartState{
				Current_state: "TestState",
				Updated_at:    time.Now(),
			},
		},
	}
}

func TestDiscriminatorNoMatch(t *testing.T) {
	t.Run("returns error when discriminator returns unmapped value", func(t *testing.T) {
		// Create config with AddBranch
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string {
				// Return value that has no When clause
				return "unmapped_value"
			}),
			When("expected_value",
				sdkstate.Event("TestEvent"),
				sdkstate.State("NextState"),
			),
		)

		config := builder.Build()
		proj := createMinimalTestProject(t)

		// Try to determine event - should fail
		event, err := config.DetermineEvent(proj)

		require.Error(t, err)
		assert.Empty(t, event)
		assert.Contains(t, err.Error(), "no branch defined")
		assert.Contains(t, err.Error(), "unmapped_value")
		assert.Contains(t, err.Error(), "expected_value") // Shows available values
	})

	t.Run("error message lists all available values", func(t *testing.T) {
		// Test with multiple branches to ensure all are listed
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string { return "invalid" }),
			When("value1", sdkstate.Event("E1"), sdkstate.State("S1")),
			When("value2", sdkstate.Event("E2"), sdkstate.State("S2")),
			When("value3", sdkstate.Event("E3"), sdkstate.State("S3")),
		)

		config := builder.Build()
		proj := createMinimalTestProject(t)

		_, err := config.DetermineEvent(proj)

		require.Error(t, err)
		// Should list all three values
		assert.Contains(t, err.Error(), "value1")
		assert.Contains(t, err.Error(), "value2")
		assert.Contains(t, err.Error(), "value3")
	})
}

func TestAddBranchNoDiscriminator(t *testing.T) {
	t.Run("panics when BranchOn not provided", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("TestState"),
				// Missing BranchOn!
				When("value", sdkstate.Event("E"), sdkstate.State("S")),
			)
		})
	})

	t.Run("panic message explains how to fix", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		defer func() {
			r := recover()
			require.NotNil(t, r)
			msg := fmt.Sprintf("%v", r)
			assert.Contains(t, msg, "no discriminator provided")
			assert.Contains(t, msg, "BranchOn()")
		}()

		builder.AddBranch(
			sdkstate.State("TestState"),
			When("value", sdkstate.Event("E"), sdkstate.State("S")),
		)
	})
}

func TestAddBranchNoBranches(t *testing.T) {
	t.Run("panics when no When clauses provided", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("TestState"),
				BranchOn(func(_ *state.Project) string { return "test" }),
				// Missing When!
			)
		})
	})

	t.Run("panic message explains how to fix", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		defer func() {
			r := recover()
			require.NotNil(t, r)
			msg := fmt.Sprintf("%v", r)
			assert.Contains(t, msg, "no branch paths provided")
			assert.Contains(t, msg, "When()")
		}()

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string { return "test" }),
		)
	})
}

func TestAddBranchConflictWithOnAdvance(t *testing.T) {
	t.Run("panics when state already has OnAdvance", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		// Add OnAdvance first
		builder.OnAdvance(
			sdkstate.State("TestState"),
			func(_ *state.Project) (sdkstate.Event, error) {
				return sdkstate.Event("TestEvent"), nil
			},
		)

		// Try to add AddBranch for same state
		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("TestState"),
				BranchOn(func(_ *state.Project) string { return "test" }),
				When("test", sdkstate.Event("E"), sdkstate.State("S")),
			)
		})
	})

	t.Run("panic message explains conflict", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		builder.OnAdvance(
			sdkstate.State("TestState"),
			func(_ *state.Project) (sdkstate.Event, error) {
				return sdkstate.Event("TestEvent"), nil
			},
		)

		defer func() {
			r := recover()
			require.NotNil(t, r)
			msg := fmt.Sprintf("%v", r)
			assert.Contains(t, msg, "already has OnAdvance")
			assert.Contains(t, msg, "cannot use both")
		}()

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string { return "test" }),
			When("test", sdkstate.Event("E"), sdkstate.State("S")),
		)
	})
}

func TestAddBranchEmptyDiscriminatorValue(t *testing.T) {
	t.Run("panics when When uses empty string as value", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		assert.Panics(t, func() {
			builder.AddBranch(
				sdkstate.State("TestState"),
				BranchOn(func(_ *state.Project) string { return "" }),
				When("", sdkstate.Event("E"), sdkstate.State("S")),
			)
		})
	})

	t.Run("panic message explains issue", func(t *testing.T) {
		builder := NewProjectTypeConfigBuilder("test")

		defer func() {
			r := recover()
			require.NotNil(t, r)
			msg := fmt.Sprintf("%v", r)
			assert.Contains(t, msg, "empty string")
			assert.Contains(t, msg, "not allowed")
		}()

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string { return "" }),
			When("", sdkstate.Event("E"), sdkstate.State("S")),
		)
	})
}

func TestDiscriminatorReturnsEmptyString(t *testing.T) {
	t.Run("returns helpful error when discriminator returns empty string", func(t *testing.T) {
		// This is different from validation - discriminator might legitimately
		// return empty string at runtime (e.g., no data available yet)
		builder := NewProjectTypeConfigBuilder("test")

		builder.AddBranch(
			sdkstate.State("TestState"),
			BranchOn(func(_ *state.Project) string {
				// Return empty string (e.g., data not ready)
				return ""
			}),
			When("ready", sdkstate.Event("E"), sdkstate.State("S")),
		)

		config := builder.Build()
		proj := createMinimalTestProject(t)

		_, err := config.DetermineEvent(proj)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no branch defined")
		// Empty string should be quoted for clarity
		assert.Contains(t, err.Error(), `""`)
	})
}
