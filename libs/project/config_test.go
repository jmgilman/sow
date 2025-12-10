package project

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test states and events for config tests.
const (
	configTestStatePlanningActive State = "PlanningActive"
	configTestStateImplPlanning   State = "ImplementationPlanning"
	configTestStateImplExecuting  State = "ImplementationExecuting"
	configTestStateReviewActive   State = "ReviewActive"

	configTestEventAdvancePlanning Event = "AdvancePlanning"
	configTestEventStartImpl       Event = "StartImplementation"
)

func TestProjectTypeConfig_Name(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configName string
		want       string
	}{
		{
			name:       "returns config name",
			configName: "standard",
			want:       "standard",
		},
		{
			name:       "returns empty name",
			configName: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{name: tt.configName}

			got := ptc.Name()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_InitialState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		initialState State
		want         State
	}{
		{
			name:         "returns initial state",
			initialState: configTestStatePlanningActive,
			want:         configTestStatePlanningActive,
		},
		{
			name:         "returns NoProject state",
			initialState: NoProject,
			want:         NoProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{initialState: tt.initialState}

			got := ptc.InitialState()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_Phases(t *testing.T) {
	t.Parallel()

	t.Run("returns phase configs map", func(t *testing.T) {
		t.Parallel()

		phases := map[string]*PhaseConfig{
			"planning": {
				name:       "planning",
				startState: configTestStatePlanningActive,
				endState:   configTestStatePlanningActive,
			},
			"implementation": {
				name:       "implementation",
				startState: configTestStateImplPlanning,
				endState:   configTestStateImplExecuting,
			},
		}

		ptc := &ProjectTypeConfig{phaseConfigs: phases}

		got := ptc.Phases()

		assert.Equal(t, phases, got)
		assert.Len(t, got, 2)
	})

	t.Run("returns nil when no phases configured", func(t *testing.T) {
		t.Parallel()

		ptc := &ProjectTypeConfig{}

		got := ptc.Phases()

		assert.Nil(t, got)
	})
}

func TestProjectTypeConfig_GetPhaseForState(t *testing.T) {
	t.Parallel()

	phases := map[string]*PhaseConfig{
		"planning": {
			name:       "planning",
			startState: configTestStatePlanningActive,
			endState:   configTestStatePlanningActive,
		},
		"implementation": {
			name:       "implementation",
			startState: configTestStateImplPlanning,
			endState:   configTestStateImplExecuting,
		},
	}

	tests := []struct {
		name  string
		state State
		want  string
	}{
		{
			name:  "returns phase for start state",
			state: configTestStatePlanningActive,
			want:  "planning",
		},
		{
			name:  "returns phase for end state",
			state: configTestStateImplExecuting,
			want:  "implementation",
		},
		{
			name:  "returns empty for unknown state",
			state: State("Unknown"),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{phaseConfigs: phases}

			got := ptc.GetPhaseForState(tt.state)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_IsPhaseStartState(t *testing.T) {
	t.Parallel()

	phases := map[string]*PhaseConfig{
		"planning": {
			name:       "planning",
			startState: configTestStatePlanningActive,
			endState:   configTestStatePlanningActive,
		},
		"implementation": {
			name:       "implementation",
			startState: configTestStateImplPlanning,
			endState:   configTestStateImplExecuting,
		},
	}

	tests := []struct {
		name      string
		phaseName string
		state     State
		want      bool
	}{
		{
			name:      "returns true for start state",
			phaseName: "planning",
			state:     configTestStatePlanningActive,
			want:      true,
		},
		{
			name:      "returns false for end state",
			phaseName: "implementation",
			state:     configTestStateImplExecuting,
			want:      false,
		},
		{
			name:      "returns false for unknown phase",
			phaseName: "unknown",
			state:     configTestStatePlanningActive,
			want:      false,
		},
		{
			name:      "returns false for wrong state",
			phaseName: "planning",
			state:     configTestStateImplPlanning,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{phaseConfigs: phases}

			got := ptc.IsPhaseStartState(tt.phaseName, tt.state)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_IsPhaseEndState(t *testing.T) {
	t.Parallel()

	phases := map[string]*PhaseConfig{
		"planning": {
			name:       "planning",
			startState: configTestStatePlanningActive,
			endState:   configTestStatePlanningActive,
		},
		"implementation": {
			name:       "implementation",
			startState: configTestStateImplPlanning,
			endState:   configTestStateImplExecuting,
		},
	}

	tests := []struct {
		name      string
		phaseName string
		state     State
		want      bool
	}{
		{
			name:      "returns true for end state",
			phaseName: "implementation",
			state:     configTestStateImplExecuting,
			want:      true,
		},
		{
			name:      "returns true when start equals end",
			phaseName: "planning",
			state:     configTestStatePlanningActive,
			want:      true,
		},
		{
			name:      "returns false for start state",
			phaseName: "implementation",
			state:     configTestStateImplPlanning,
			want:      false,
		},
		{
			name:      "returns false for unknown phase",
			phaseName: "unknown",
			state:     configTestStateImplExecuting,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{phaseConfigs: phases}

			got := ptc.IsPhaseEndState(tt.phaseName, tt.state)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_GetTransition(t *testing.T) {
	t.Parallel()

	transitions := []TransitionConfig{
		{
			From:        configTestStatePlanningActive,
			To:          configTestStateImplPlanning,
			Event:       configTestEventAdvancePlanning,
			description: "Complete planning",
		},
		{
			From:        configTestStateImplPlanning,
			To:          configTestStateImplExecuting,
			Event:       configTestEventStartImpl,
			description: "Start implementation",
		},
	}

	tests := []struct {
		name  string
		from  State
		to    State
		event Event
		want  *TransitionConfig
	}{
		{
			name:  "returns matching transition",
			from:  configTestStatePlanningActive,
			to:    configTestStateImplPlanning,
			event: configTestEventAdvancePlanning,
			want:  &transitions[0],
		},
		{
			name:  "returns nil for no match",
			from:  configTestStatePlanningActive,
			to:    configTestStateImplPlanning,
			event: configTestEventStartImpl, // Wrong event
			want:  nil,
		},
		{
			name:  "returns nil for unknown from state",
			from:  State("Unknown"),
			to:    configTestStateImplPlanning,
			event: configTestEventAdvancePlanning,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{transitions: transitions}

			got := ptc.GetTransition(tt.from, tt.to, tt.event)

			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.From, got.From)
				assert.Equal(t, tt.want.To, got.To)
				assert.Equal(t, tt.want.Event, got.Event)
			}
		})
	}
}

func TestProjectTypeConfig_Initialize(t *testing.T) {
	t.Parallel()

	t.Run("calls initializer when set", func(t *testing.T) {
		t.Parallel()

		called := false
		var receivedInputs map[string][]project.ArtifactState

		initializer := func(_ *state.Project, inputs map[string][]project.ArtifactState) error {
			called = true
			receivedInputs = inputs
			return nil
		}

		ptc := &ProjectTypeConfig{initializer: initializer}
		proj := &state.Project{}
		inputs := map[string][]project.ArtifactState{
			"test": {{Type: "test"}},
		}

		err := ptc.Initialize(proj, inputs)

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, inputs, receivedInputs)
	})

	t.Run("returns nil when no initializer", func(t *testing.T) {
		t.Parallel()

		ptc := &ProjectTypeConfig{}
		proj := &state.Project{}

		err := ptc.Initialize(proj, nil)

		assert.NoError(t, err)
	})
}

func TestProjectTypeConfig_GetStatePrompt(t *testing.T) {
	t.Parallel()

	prompts := map[State]PromptGenerator{
		configTestStatePlanningActive: func(*state.Project) string {
			return "Create your plan"
		},
		configTestStateImplPlanning: func(p *state.Project) string {
			return "Planning implementation for: " + p.Name
		},
	}

	tests := []struct {
		name        string
		state       State
		projectName string
		want        string
	}{
		{
			name:        "returns prompt for configured state",
			state:       configTestStatePlanningActive,
			projectName: "",
			want:        "Create your plan",
		},
		{
			name:        "returns prompt with project context",
			state:       configTestStateImplPlanning,
			projectName: "Test Project",
			want:        "Planning implementation for: Test Project",
		},
		{
			name:        "returns empty for unconfigured state",
			state:       configTestStateImplExecuting,
			projectName: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{prompts: prompts}
			proj := &state.Project{}
			proj.Name = tt.projectName

			got := ptc.GetStatePrompt(tt.state, proj)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_GetTaskSupportingPhases(t *testing.T) {
	t.Parallel()

	t.Run("returns phases that support tasks", func(t *testing.T) {
		t.Parallel()

		phases := map[string]*PhaseConfig{
			"planning": {
				name:          "planning",
				supportsTasks: false,
			},
			"implementation": {
				name:          "implementation",
				supportsTasks: true,
			},
			"review": {
				name:          "review",
				supportsTasks: true,
			},
		}

		ptc := &ProjectTypeConfig{phaseConfigs: phases}

		got := ptc.GetTaskSupportingPhases()

		assert.Len(t, got, 2)
		assert.Contains(t, got, "implementation")
		assert.Contains(t, got, "review")
	})

	t.Run("returns empty when no phases support tasks", func(t *testing.T) {
		t.Parallel()

		phases := map[string]*PhaseConfig{
			"planning": {
				name:          "planning",
				supportsTasks: false,
			},
		}

		ptc := &ProjectTypeConfig{phaseConfigs: phases}

		got := ptc.GetTaskSupportingPhases()

		assert.Empty(t, got)
	})

	t.Run("returns sorted phase names", func(t *testing.T) {
		t.Parallel()

		phases := map[string]*PhaseConfig{
			"zebra": {
				name:          "zebra",
				supportsTasks: true,
			},
			"alpha": {
				name:          "alpha",
				supportsTasks: true,
			},
		}

		ptc := &ProjectTypeConfig{phaseConfigs: phases}

		got := ptc.GetTaskSupportingPhases()

		require.Len(t, got, 2)
		assert.Equal(t, "alpha", got[0])
		assert.Equal(t, "zebra", got[1])
	})
}

func TestProjectTypeConfig_PhaseSupportsTasks(t *testing.T) {
	t.Parallel()

	phases := map[string]*PhaseConfig{
		"planning": {
			name:          "planning",
			supportsTasks: false,
		},
		"implementation": {
			name:          "implementation",
			supportsTasks: true,
		},
	}

	tests := []struct {
		name      string
		phaseName string
		want      bool
	}{
		{
			name:      "returns true when phase supports tasks",
			phaseName: "implementation",
			want:      true,
		},
		{
			name:      "returns false when phase does not support tasks",
			phaseName: "planning",
			want:      false,
		},
		{
			name:      "returns false for unknown phase",
			phaseName: "unknown",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{phaseConfigs: phases}

			got := ptc.PhaseSupportsTasks(tt.phaseName)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectTypeConfig_IsBranchingState(t *testing.T) {
	t.Parallel()

	branches := map[State]*BranchConfig{
		configTestStateReviewActive: {
			from: configTestStateReviewActive,
		},
	}

	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{
			name:  "returns true for branching state",
			state: configTestStateReviewActive,
			want:  true,
		},
		{
			name:  "returns false for non-branching state",
			state: configTestStatePlanningActive,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ptc := &ProjectTypeConfig{branches: branches}

			got := ptc.IsBranchingState(tt.state)

			assert.Equal(t, tt.want, got)
		})
	}
}
