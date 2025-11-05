package state

import (
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/jmgilman/sow/cli/schemas/project"
)

// mockProjectTypeConfig is a mock implementation of ProjectTypeConfig for testing.
type mockProjectTypeConfig struct {
	initialState State
}

func newMockProjectTypeConfig() *mockProjectTypeConfig {
	return &mockProjectTypeConfig{
		initialState: State("PlanningActive"),
	}
}

func (m *mockProjectTypeConfig) InitialState() State {
	return m.initialState
}

func (m *mockProjectTypeConfig) Initialize(_ *Project, _ map[string][]project.ArtifactState) error {
	// Mock initialization - do nothing
	return nil
}

func (m *mockProjectTypeConfig) BuildMachine(_ *Project, _ State) *sdkstate.Machine {
	// For tests, we return nil since most tests don't need the actual machine
	// Tests that need a real machine should use the full SDK builder
	return nil
}

func (m *mockProjectTypeConfig) Validate(_ *Project) error {
	// Mock validation - always pass
	return nil
}

func (m *mockProjectTypeConfig) GetTaskSupportingPhases() []string {
	return []string{"implementation"}
}

func (m *mockProjectTypeConfig) PhaseSupportsTasks(phaseName string) bool {
	return phaseName == "implementation"
}

func (m *mockProjectTypeConfig) GetDefaultTaskPhase(_ State) string {
	return "implementation"
}

func (m *mockProjectTypeConfig) OrchestratorPrompt(_ *Project) string {
	return ""
}

func (m *mockProjectTypeConfig) GetStatePrompt(_ State, _ *Project) string {
	return ""
}

func (m *mockProjectTypeConfig) DetermineEvent(_ *Project) (Event, error) {
	// Mock event determination - return empty event
	return "", nil
}
