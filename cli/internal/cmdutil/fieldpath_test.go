package cmdutil

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ParseFieldPath - splits field paths on dots.
func TestParseFieldPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple field",
			path:     "approved",
			expected: []string{"approved"},
		},
		{
			name:     "metadata field",
			path:     "metadata.assessment",
			expected: []string{"metadata", "assessment"},
		},
		{
			name:     "nested metadata",
			path:     "metadata.foo.bar",
			expected: []string{"metadata", "foo", "bar"},
		},
		{
			name:     "deep nested metadata",
			path:     "metadata.a.b.c.d",
			expected: []string{"metadata", "a", "b", "c", "d"},
		},
		{
			name:     "empty string",
			path:     "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFieldPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test IsMetadataPath - determines if path routes to metadata.
func TestIsMetadataPath(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		expected bool
	}{
		{
			name:     "metadata path",
			segments: []string{"metadata", "assessment"},
			expected: true,
		},
		{
			name:     "nested metadata path",
			segments: []string{"metadata", "foo", "bar"},
			expected: true,
		},
		{
			name:     "direct field",
			segments: []string{"approved"},
			expected: false,
		},
		{
			name:     "direct field with dots",
			segments: []string{"created_at"},
			expected: false,
		},
		{
			name:     "empty segments",
			segments: []string{},
			expected: false,
		},
		{
			name:     "single metadata segment",
			segments: []string{"metadata"},
			expected: false, // Need at least 2 segments for valid metadata path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMetadataPath(tt.segments)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test ConvertValue - converts strings to appropriate types.
func TestConvertValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected interface{}
	}{
		{
			name:     "true boolean",
			value:    "true",
			expected: true,
		},
		{
			name:     "false boolean",
			value:    "false",
			expected: false,
		},
		{
			name:     "positive integer",
			value:    "123",
			expected: 123,
		},
		{
			name:     "negative integer",
			value:    "-456",
			expected: -456,
		},
		{
			name:     "zero",
			value:    "0",
			expected: 0,
		},
		{
			name:     "regular string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "string that looks like bool",
			value:    "True",
			expected: "True", // Case sensitive, not a bool
		},
		{
			name:     "string that looks like number",
			value:    "123abc",
			expected: "123abc", // Not a valid number
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test SetField - sets direct fields on Artifact.
func TestSetFieldArtifactDirect(t *testing.T) {
	t.Run("set approved to true", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "approved", "true")
		require.NoError(t, err)
		assert.True(t, artifact.Approved)
	})

	t.Run("set approved to false", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Approved: true,
			},
		}
		err := SetField(artifact, "approved", "false")
		require.NoError(t, err)
		assert.False(t, artifact.Approved)
	})

	t.Run("set type field", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "type", "review")
		require.NoError(t, err)
		assert.Equal(t, "review", artifact.Type)
	})

	t.Run("set path field", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "path", "phases/review/review.md")
		require.NoError(t, err)
		assert.Equal(t, "phases/review/review.md", artifact.Path)
	})

	t.Run("error on invalid field", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "nonexistent", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown field")
	})
}

// Test SetField - sets metadata fields on Artifact.
func TestSetFieldArtifactMetadata(t *testing.T) {
	t.Run("set simple metadata field", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata.assessment", "pass")
		require.NoError(t, err)
		assert.Equal(t, "pass", artifact.Metadata["assessment"])
	})

	t.Run("set nested metadata field", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata.foo.bar", "baz")
		require.NoError(t, err)

		// Check nested structure
		fooMap, ok := artifact.Metadata["foo"].(map[string]interface{})
		require.True(t, ok, "foo should be a map")
		assert.Equal(t, "baz", fooMap["bar"])
	})

	t.Run("set multiple metadata fields", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata.assessment", "pass")
		require.NoError(t, err)
		err = SetField(artifact, "metadata.reviewer", "alice")
		require.NoError(t, err)

		assert.Equal(t, "pass", artifact.Metadata["assessment"])
		assert.Equal(t, "alice", artifact.Metadata["reviewer"])
	})

	t.Run("metadata with type conversion", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata.score", "95")
		require.NoError(t, err)
		assert.Equal(t, 95, artifact.Metadata["score"])
	})
}

// Test SetField - sets fields on Task.
func TestSetFieldTask(t *testing.T) {
	t.Run("set task iteration", func(t *testing.T) {
		task := &state.Task{}
		err := SetField(task, "iteration", "2")
		require.NoError(t, err)
		assert.Equal(t, int64(2), task.Iteration)
	})

	t.Run("set task status", func(t *testing.T) {
		task := &state.Task{}
		err := SetField(task, "status", "in_progress")
		require.NoError(t, err)
		assert.Equal(t, "in_progress", task.Status)
	})

	t.Run("set task metadata", func(t *testing.T) {
		task := &state.Task{}
		err := SetField(task, "metadata.priority", "high")
		require.NoError(t, err)
		assert.Equal(t, "high", task.Metadata["priority"])
	})
}

// Test SetField - sets fields on Phase.
func TestSetFieldPhase(t *testing.T) {
	t.Run("set phase enabled", func(t *testing.T) {
		phase := &state.Phase{}
		err := SetField(phase, "enabled", "true")
		require.NoError(t, err)
		assert.True(t, phase.Enabled)
	})

	t.Run("set phase status", func(t *testing.T) {
		phase := &state.Phase{}
		err := SetField(phase, "status", "completed")
		require.NoError(t, err)
		assert.Equal(t, "completed", phase.Status)
	})

	t.Run("set phase metadata", func(t *testing.T) {
		phase := &state.Phase{}
		err := SetField(phase, "metadata.complexity", "high")
		require.NoError(t, err)
		assert.Equal(t, "high", phase.Metadata["complexity"])
	})
}

// Test GetField - retrieves direct fields.
func TestGetFieldDirect(t *testing.T) {
	t.Run("get artifact approved", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Approved: true,
			},
		}
		value, err := GetField(artifact, "approved")
		require.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("get artifact type", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Type: "review",
			},
		}
		value, err := GetField(artifact, "type")
		require.NoError(t, err)
		assert.Equal(t, "review", value)
	})

	t.Run("get task iteration", func(t *testing.T) {
		task := &state.Task{
			TaskState: project.TaskState{
				Iteration: 3,
			},
		}
		value, err := GetField(task, "iteration")
		require.NoError(t, err)
		assert.Equal(t, int64(3), value)
	})

	t.Run("error on invalid field", func(t *testing.T) {
		artifact := &state.Artifact{}
		_, err := GetField(artifact, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown field")
	})
}

// Test GetField - retrieves metadata fields.
func TestGetFieldMetadata(t *testing.T) {
	t.Run("get simple metadata", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"assessment": "pass",
				},
			},
		}
		value, err := GetField(artifact, "metadata.assessment")
		require.NoError(t, err)
		assert.Equal(t, "pass", value)
	})

	t.Run("get nested metadata", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"foo": map[string]interface{}{
						"bar": "baz",
					},
				},
			},
		}
		value, err := GetField(artifact, "metadata.foo.bar")
		require.NoError(t, err)
		assert.Equal(t, "baz", value)
	})

	t.Run("error on missing metadata key", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{},
			},
		}
		_, err := GetField(artifact, "metadata.missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// Test IsKnownField - validates field names against known types.
func TestIsKnownField(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		fieldName string
		expected  bool
	}{
		{
			name:      "artifact known field - type",
			typeName:  "Artifact",
			fieldName: "type",
			expected:  true,
		},
		{
			name:      "artifact known field - approved",
			typeName:  "Artifact",
			fieldName: "approved",
			expected:  true,
		},
		{
			name:      "artifact unknown field",
			typeName:  "Artifact",
			fieldName: "unknown",
			expected:  false,
		},
		{
			name:      "task known field - iteration",
			typeName:  "Task",
			fieldName: "iteration",
			expected:  true,
		},
		{
			name:      "phase known field - status",
			typeName:  "Phase",
			fieldName: "status",
			expected:  true,
		},
		{
			name:      "project known field - name",
			typeName:  "Project",
			fieldName: "name",
			expected:  true,
		},
		{
			name:      "unknown type",
			typeName:  "UnknownType",
			fieldName: "field",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKnownField(tt.typeName, tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test GetTypeName - returns type name of state objects.
func TestGetTypeName(t *testing.T) {
	t.Run("artifact pointer", func(t *testing.T) {
		artifact := &state.Artifact{}
		result := GetTypeName(artifact)
		assert.Equal(t, "Artifact", result)
	})

	t.Run("artifact value", func(t *testing.T) {
		artifact := state.Artifact{}
		result := GetTypeName(artifact)
		assert.Equal(t, "Artifact", result)
	})

	t.Run("task pointer", func(t *testing.T) {
		task := &state.Task{}
		result := GetTypeName(task)
		assert.Equal(t, "Task", result)
	})

	t.Run("phase pointer", func(t *testing.T) {
		phase := &state.Phase{}
		result := GetTypeName(phase)
		assert.Equal(t, "Phase", result)
	})
}

// Test edge cases and error handling.
func TestFieldPathEdgeCases(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var artifact *state.Artifact
		err := SetField(artifact, "approved", "true")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("invalid target type", func(t *testing.T) {
		notAStruct := "string value"
		err := SetField(&notAStruct, "field", "value")
		assert.Error(t, err)
	})

	t.Run("empty field path", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("metadata with only one segment", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid metadata path")
	})

	t.Run("type conversion errors", func(t *testing.T) {
		// Try to set a bool field with invalid string
		artifact := &state.Artifact{}
		err := SetField(artifact, "approved", "not-a-bool")
		// Should error because can't convert to bool
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("nested metadata path conflict", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"foo": "string value", // Not a map
				},
			},
		}
		err := SetField(artifact, "metadata.foo.bar", "baz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a map")
	})

	t.Run("get field on nil metadata", func(t *testing.T) {
		artifact := &state.Artifact{}
		_, err := GetField(artifact, "metadata.missing")
		assert.Error(t, err)
	})

	t.Run("capitalizeFirst edge cases", func(t *testing.T) {
		// Test the capitalizeFirst helper indirectly
		artifact := &state.Artifact{}
		// Field name with underscore
		err := SetField(artifact, "created_at", "2024-01-01T00:00:00Z")
		// Should fail because created_at is a time.Time, not string
		assert.Error(t, err)
	})
}

// Test ProjectState field setting (to cover all state types).
func TestSetFieldProject(t *testing.T) {
	// Note: We can't import ProjectState here without circular dependency
	// So these tests are just placeholders demonstrating the pattern
	// The actual implementation works with any struct that has the right fields

	t.Run("set metadata on various types", func(t *testing.T) {
		// Test that metadata setting works consistently across types
		task := &state.Task{}
		err := SetField(task, "metadata.key", "value")
		require.NoError(t, err)
		assert.Equal(t, "value", task.Metadata["key"])

		phase := &state.Phase{}
		err = SetField(phase, "metadata.key", "value")
		require.NoError(t, err)
		assert.Equal(t, "value", phase.Metadata["key"])

		artifact := &state.Artifact{}
		err = SetField(artifact, "metadata.key", "value")
		require.NoError(t, err)
		assert.Equal(t, "value", artifact.Metadata["key"])
	})
}

// Test additional field types for complete coverage.
func TestSetFieldIntegerTypes(t *testing.T) {
	t.Run("set int64 field with negative value", func(t *testing.T) {
		task := &state.Task{}
		// Iteration is int64
		err := SetField(task, "iteration", "-5")
		require.NoError(t, err)
		assert.Equal(t, int64(-5), task.Iteration)
	})

	t.Run("set int64 field with zero", func(t *testing.T) {
		task := &state.Task{}
		err := SetField(task, "iteration", "0")
		require.NoError(t, err)
		assert.Equal(t, int64(0), task.Iteration)
	})

	t.Run("cannot set int field with non-numeric", func(t *testing.T) {
		task := &state.Task{}
		err := SetField(task, "iteration", "not-a-number")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

// Test GetField edge cases.
func TestGetFieldEdgeCases(t *testing.T) {
	t.Run("get field from nil pointer", func(t *testing.T) {
		var artifact *state.Artifact
		_, err := GetField(artifact, "approved")
		assert.Error(t, err)
		// GetField checks for nil at the top level
	})

	t.Run("get field with empty path", func(t *testing.T) {
		artifact := &state.Artifact{}
		_, err := GetField(artifact, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("get nested metadata that doesn't exist", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"foo": map[string]interface{}{
						"bar": "value",
					},
				},
			},
		}
		_, err := GetField(artifact, "metadata.foo.nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get metadata when intermediate value is not a map", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"foo": "string value",
				},
			},
		}
		_, err := GetField(artifact, "metadata.foo.bar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected map")
	})
}

// Test deeply nested metadata paths.
func TestDeepNestedMetadata(t *testing.T) {
	t.Run("create deeply nested metadata", func(t *testing.T) {
		artifact := &state.Artifact{}
		err := SetField(artifact, "metadata.level1.level2.level3.level4", "deep-value")
		require.NoError(t, err)

		// Verify the nested structure
		level1, ok := artifact.Metadata["level1"].(map[string]interface{})
		require.True(t, ok)
		level2, ok := level1["level2"].(map[string]interface{})
		require.True(t, ok)
		level3, ok := level2["level3"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "deep-value", level3["level4"])
	})

	t.Run("retrieve deeply nested metadata", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "nested-value",
						},
					},
				},
			},
		}

		value, err := GetField(artifact, "metadata.level1.level2.level3")
		require.NoError(t, err)
		assert.Equal(t, "nested-value", value)
	})

	t.Run("add to existing nested metadata", func(t *testing.T) {
		artifact := &state.Artifact{
			ArtifactState: project.ArtifactState{
				Metadata: map[string]any{
					"existing": map[string]interface{}{
						"key1": "value1",
					},
				},
			},
		}

		// Add a new key to existing nested map
		err := SetField(artifact, "metadata.existing.key2", "value2")
		require.NoError(t, err)

		existingMap, ok := artifact.Metadata["existing"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value1", existingMap["key1"])
		assert.Equal(t, "value2", existingMap["key2"])
	})
}

// Test all known field types.
func TestKnownFieldsComprehensive(t *testing.T) {
	t.Run("all artifact known fields", func(t *testing.T) {
		fields := []string{"type", "path", "approved", "created_at"}
		for _, field := range fields {
			assert.True(t, IsKnownField("Artifact", field), "field %s should be known", field)
		}
	})

	t.Run("all phase known fields", func(t *testing.T) {
		fields := []string{"status", "enabled", "created_at", "started_at", "completed_at"}
		for _, field := range fields {
			assert.True(t, IsKnownField("Phase", field), "field %s should be known", field)
		}
	})

	t.Run("all task known fields", func(t *testing.T) {
		fields := []string{"id", "name", "phase", "status", "iteration", "assigned_agent", "created_at", "started_at", "updated_at", "completed_at"}
		for _, field := range fields {
			assert.True(t, IsKnownField("Task", field), "field %s should be known", field)
		}
	})

	t.Run("all project known fields", func(t *testing.T) {
		fields := []string{"name", "type", "branch", "description", "created_at", "updated_at"}
		for _, field := range fields {
			assert.True(t, IsKnownField("Project", field), "field %s should be known", field)
		}
	})
}
