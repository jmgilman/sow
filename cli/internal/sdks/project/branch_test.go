package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projectschema "github.com/jmgilman/sow/cli/schemas/project"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchOn(t *testing.T) {
	t.Run("sets discriminator function", func(t *testing.T) {
		bc := &BranchConfig{}
		discriminator := func(p *state.Project) string {
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
		guardFunc := func(p *state.Project) bool {
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

		discriminator := func(p *state.Project) string {
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

		BranchOn(func(p *state.Project) string {
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
