package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/stretchr/testify/assert"
)

// Test states and events for branch tests.
const (
	branchTestStateReview   State = "ReviewActive"
	branchTestStateFinalize State = "FinalizeChecks"
	branchTestStateImpl     State = "ImplementationPlanning"

	branchTestEventPass Event = "ReviewPass"
	branchTestEventFail Event = "ReviewFail"
)

func TestBranchPath_Fields(t *testing.T) {
	t.Parallel()

	t.Run("stores branch path values correctly", func(t *testing.T) {
		t.Parallel()

		path := &BranchPath{
			value:       "pass",
			event:       branchTestEventPass,
			to:          branchTestStateFinalize,
			description: "Review approved - proceed to finalization",
			failedPhase: "",
		}

		assert.Equal(t, "pass", path.value)
		assert.Equal(t, branchTestEventPass, path.event)
		assert.Equal(t, branchTestStateFinalize, path.to)
		assert.Equal(t, "Review approved - proceed to finalization", path.description)
		assert.Empty(t, path.failedPhase)
	})

	t.Run("stores failed phase for failure path", func(t *testing.T) {
		t.Parallel()

		path := &BranchPath{
			value:       "fail",
			event:       branchTestEventFail,
			to:          branchTestStateImpl,
			description: "Review failed - return to implementation",
			failedPhase: "review",
		}

		assert.Equal(t, "fail", path.value)
		assert.Equal(t, branchTestEventFail, path.event)
		assert.Equal(t, branchTestStateImpl, path.to)
		assert.Equal(t, "Review failed - return to implementation", path.description)
		assert.Equal(t, "review", path.failedPhase)
	})

	t.Run("stores guard template", func(t *testing.T) {
		t.Parallel()

		guardFunc := func(*state.Project) bool { return true }
		path := &BranchPath{
			value: "pass",
			event: branchTestEventPass,
			to:    branchTestStateFinalize,
			guardTemplate: GuardTemplate{
				Description: "review passed",
				Func:        guardFunc,
			},
		}

		assert.Equal(t, "review passed", path.guardTemplate.Description)
		assert.NotNil(t, path.guardTemplate.Func)
	})

	t.Run("stores actions", func(t *testing.T) {
		t.Parallel()

		onEntry := func(*state.Project) error { return nil }
		onExit := func(*state.Project) error { return nil }
		path := &BranchPath{
			value:   "pass",
			event:   branchTestEventPass,
			to:      branchTestStateFinalize,
			onEntry: onEntry,
			onExit:  onExit,
		}

		assert.NotNil(t, path.onEntry)
		assert.NotNil(t, path.onExit)
	})
}

func TestBranchConfig_Structure(t *testing.T) {
	t.Parallel()

	t.Run("stores branch config correctly", func(t *testing.T) {
		t.Parallel()

		discriminator := func(*state.Project) string { return "pass" }
		bc := &BranchConfig{
			from:          branchTestStateReview,
			discriminator: discriminator,
			branches: map[string]*BranchPath{
				"pass": {
					value: "pass",
					event: branchTestEventPass,
					to:    branchTestStateFinalize,
				},
				"fail": {
					value: "fail",
					event: branchTestEventFail,
					to:    branchTestStateImpl,
				},
			},
		}

		assert.Equal(t, branchTestStateReview, bc.from)
		assert.NotNil(t, bc.discriminator)
		assert.Len(t, bc.branches, 2)
		assert.NotNil(t, bc.branches["pass"])
		assert.NotNil(t, bc.branches["fail"])
	})

	t.Run("discriminator returns expected value", func(t *testing.T) {
		t.Parallel()

		discriminator := func(*state.Project) string { return "pass" }
		bc := &BranchConfig{
			from:          branchTestStateReview,
			discriminator: discriminator,
		}

		result := bc.discriminator(nil)
		assert.Equal(t, "pass", result)
	})
}
