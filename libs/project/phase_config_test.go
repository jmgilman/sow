package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhaseConfig_Name(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		phaseName string
		want      string
	}{
		{
			name:      "returns phase name",
			phaseName: "planning",
			want:      "planning",
		},
		{
			name:      "returns empty name",
			phaseName: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{name: tt.phaseName}

			got := pc.Name()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_StartState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		startState State
		want       State
	}{
		{
			name:       "returns start state",
			startState: State("PlanningActive"),
			want:       State("PlanningActive"),
		},
		{
			name:       "returns empty state",
			startState: State(""),
			want:       State(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{startState: tt.startState}

			got := pc.StartState()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_EndState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endState State
		want     State
	}{
		{
			name:     "returns end state",
			endState: State("PlanningComplete"),
			want:     State("PlanningComplete"),
		},
		{
			name:     "returns empty state",
			endState: State(""),
			want:     State(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{endState: tt.endState}

			got := pc.EndState()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_AllowedInputTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inputTypes []string
		want       []string
	}{
		{
			name:       "returns allowed input types",
			inputTypes: []string{"task_list", "design_doc"},
			want:       []string{"task_list", "design_doc"},
		},
		{
			name:       "returns empty slice when none set",
			inputTypes: nil,
			want:       nil,
		},
		{
			name:       "returns empty slice",
			inputTypes: []string{},
			want:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{allowedInputTypes: tt.inputTypes}

			got := pc.AllowedInputTypes()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_AllowedOutputTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		outputTypes []string
		want        []string
	}{
		{
			name:        "returns allowed output types",
			outputTypes: []string{"code", "tests"},
			want:        []string{"code", "tests"},
		},
		{
			name:        "returns empty slice when none set",
			outputTypes: nil,
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{allowedOutputTypes: tt.outputTypes}

			got := pc.AllowedOutputTypes()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_SupportsTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		supportsTasks bool
		want          bool
	}{
		{
			name:          "returns true when tasks supported",
			supportsTasks: true,
			want:          true,
		},
		{
			name:          "returns false when tasks not supported",
			supportsTasks: false,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{supportsTasks: tt.supportsTasks}

			got := pc.SupportsTasks()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseConfig_MetadataSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
		want   string
	}{
		{
			name:   "returns metadata schema",
			schema: `{ field: string }`,
			want:   `{ field: string }`,
		},
		{
			name:   "returns empty schema",
			schema: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := &PhaseConfig{metadataSchema: tt.schema}

			got := pc.MetadataSchema()

			assert.Equal(t, tt.want, got)
		})
	}
}
