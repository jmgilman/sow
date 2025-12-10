package state

import (
	"errors"
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStructure_ValidState(t *testing.T) {
	now := time.Now()
	state := &project.ProjectState{
		Name:   "valid-project",
		Type:   "standard",
		Branch: "feat/test",
		Phases: map[string]project.PhaseState{},
		Statechart: project.StatechartState{
			Current_state: "PlanningActive",
			Updated_at:    now,
		},
		Created_at: now,
		Updated_at: now,
	}

	err := validateStructure(state)
	assert.NoError(t, err)
}

func TestValidateStructure_MissingRequiredFields(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		state *project.ProjectState
	}{
		{
			name: "empty name",
			state: &project.ProjectState{
				Name:   "",
				Type:   "standard",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "empty type",
			state: &project.ProjectState{
				Name:   "valid-name",
				Type:   "",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "empty branch",
			state: &project.ProjectState{
				Name:   "valid-name",
				Type:   "standard",
				Branch: "",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "empty current_state",
			state: &project.ProjectState{
				Name:   "valid-name",
				Type:   "standard",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStructure(tt.state)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidationFailed), "expected ErrValidationFailed")
		})
	}
}

func TestValidateStructure_InvalidFieldValues(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		state *project.ProjectState
	}{
		{
			name: "invalid name - uppercase",
			state: &project.ProjectState{
				Name:   "INVALID",
				Type:   "standard",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "invalid name - starts with hyphen",
			state: &project.ProjectState{
				Name:   "-invalid",
				Type:   "standard",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "invalid type - uppercase",
			state: &project.ProjectState{
				Name:   "valid-name",
				Type:   "STANDARD",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
		{
			name: "invalid type - contains hyphen",
			state: &project.ProjectState{
				Name:   "valid-name",
				Type:   "my-type",
				Branch: "feat/test",
				Statechart: project.StatechartState{
					Current_state: "Active",
					Updated_at:    now,
				},
				Created_at: now,
				Updated_at: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStructure(tt.state)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidationFailed), "expected ErrValidationFailed")
		})
	}
}

func TestValidateMetadata(t *testing.T) {
	t.Run("passes for valid metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"key":   "value",
			"count": 42,
		}
		schema := `{
			key: string
			count: int
		}`

		err := ValidateMetadata(metadata, schema)
		assert.NoError(t, err)
	})

	t.Run("skips validation if schema is empty", func(t *testing.T) {
		metadata := map[string]interface{}{
			"any": "value",
		}

		err := ValidateMetadata(metadata, "")
		assert.NoError(t, err)
	})

	t.Run("fails for invalid metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"key":   123, // should be string
			"count": "not-a-number",
		}
		schema := `{
			key: string
			count: int
		}`

		err := ValidateMetadata(metadata, schema)
		assert.Error(t, err)
	})

	t.Run("fails for missing required fields", func(t *testing.T) {
		metadata := map[string]interface{}{
			"key": "value",
			// missing "count"
		}
		schema := `{
			key: string
			count: int
		}`

		err := ValidateMetadata(metadata, schema)
		assert.Error(t, err)
	})
}

func TestValidateArtifactTypes(t *testing.T) {
	t.Run("passes for allowed types", func(t *testing.T) {
		now := time.Now()
		artifacts := []project.ArtifactState{
			{Type: "task_list", Path: "tasks.md", Created_at: now},
			{Type: "design_doc", Path: "design.md", Created_at: now},
		}
		allowedTypes := []string{"task_list", "design_doc", "adr"}

		err := ValidateArtifactTypes(artifacts, allowedTypes, "planning", "output")
		assert.NoError(t, err)
	})

	t.Run("skips validation if allowed list empty", func(t *testing.T) {
		now := time.Now()
		artifacts := []project.ArtifactState{
			{Type: "anything", Path: "file.md", Created_at: now},
		}
		allowedTypes := []string{}

		err := ValidateArtifactTypes(artifacts, allowedTypes, "planning", "output")
		assert.NoError(t, err)
	})

	t.Run("fails for disallowed types", func(t *testing.T) {
		now := time.Now()
		artifacts := []project.ArtifactState{
			{Type: "task_list", Path: "tasks.md", Created_at: now},
			{Type: "forbidden", Path: "bad.md", Created_at: now},
		}
		allowedTypes := []string{"task_list", "design_doc"}

		err := ValidateArtifactTypes(artifacts, allowedTypes, "planning", "output")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidArtifactType), "expected ErrInvalidArtifactType")
		assert.Contains(t, err.Error(), "forbidden")
		assert.Contains(t, err.Error(), "planning")
		assert.Contains(t, err.Error(), "output")
	})

	t.Run("passes for empty artifact list", func(t *testing.T) {
		artifacts := []project.ArtifactState{}
		allowedTypes := []string{"task_list", "design_doc"}

		err := ValidateArtifactTypes(artifacts, allowedTypes, "planning", "output")
		assert.NoError(t, err)
	})
}
