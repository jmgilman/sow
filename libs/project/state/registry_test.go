package state

import (
	"sort"
	"testing"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/qmuntal/stateless"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registryTestConfig implements ProjectTypeConfig for registry testing.
// This is separate from mockProjectTypeConfig in project_test.go.
type registryTestConfig struct {
	name string
}

func (r *registryTestConfig) Name() string {
	return r.name
}

func (r *registryTestConfig) InitialState() string {
	return "MockInitial"
}

func (r *registryTestConfig) Initialize(_ *Project, _ map[string][]project.ArtifactState) error {
	return nil
}

func (r *registryTestConfig) Validate(_ *Project) error {
	return nil
}

func (r *registryTestConfig) BuildMachine(_ *Project, initialState string) *stateless.StateMachine {
	return stateless.NewStateMachine(initialState)
}

func (r *registryTestConfig) GetPhaseForState(_ string) string {
	return ""
}

func (r *registryTestConfig) IsPhaseStartState(_ string, _ string) bool {
	return false
}

func (r *registryTestConfig) OrchestratorPrompt(_ *Project) string {
	return ""
}

func (r *registryTestConfig) GetStatePrompt(_ string, _ *Project) string {
	return ""
}

func (r *registryTestConfig) PhaseSupportsTasks(_ string) bool {
	return false
}

func (r *registryTestConfig) GetTaskSupportingPhases() []string {
	return nil
}

func (r *registryTestConfig) GetDefaultTaskPhase(_ string) string {
	return ""
}

func TestRegister(t *testing.T) {
	t.Run("adds config to registry", func(t *testing.T) {
		ClearRegistry()

		config := &registryTestConfig{name: "test-type"}
		Register("test-type", config)

		got, exists := GetConfig("test-type")
		require.True(t, exists)
		assert.Equal(t, "test-type", got.Name())
	})

	t.Run("panics on duplicate registration", func(t *testing.T) {
		ClearRegistry()

		config := &registryTestConfig{name: "duplicate-type"}
		Register("duplicate-type", config)

		assert.Panics(t, func() {
			Register("duplicate-type", config)
		}, "expected panic on duplicate registration")
	})
}

func TestGetConfig(t *testing.T) {
	t.Run("returns registered config", func(t *testing.T) {
		ClearRegistry()

		config := &registryTestConfig{name: "existing-type"}
		Register("existing-type", config)

		got, exists := GetConfig("existing-type")
		require.True(t, exists)
		assert.Equal(t, "existing-type", got.Name())
	})

	t.Run("returns false for unknown type", func(t *testing.T) {
		ClearRegistry()

		got, exists := GetConfig("unknown-type")
		assert.False(t, exists)
		assert.Nil(t, got)
	})
}

func TestRegisteredTypes(t *testing.T) {
	t.Run("returns empty slice when no types registered", func(t *testing.T) {
		ClearRegistry()

		got := RegisteredTypes()
		assert.Empty(t, got)
	})

	t.Run("returns all registered type names", func(t *testing.T) {
		ClearRegistry()

		Register("alpha", &registryTestConfig{name: "alpha"})
		Register("beta", &registryTestConfig{name: "beta"})
		Register("gamma", &registryTestConfig{name: "gamma"})

		got := RegisteredTypes()
		sort.Strings(got) // Sort for deterministic comparison
		assert.Equal(t, []string{"alpha", "beta", "gamma"}, got)
	})
}

func TestClearRegistry(t *testing.T) {
	t.Run("removes all registered configurations", func(t *testing.T) {
		ClearRegistry()

		Register("to-be-cleared", &registryTestConfig{name: "to-be-cleared"})
		_, exists := GetConfig("to-be-cleared")
		require.True(t, exists)

		ClearRegistry()

		_, exists = GetConfig("to-be-cleared")
		assert.False(t, exists)
	})
}
