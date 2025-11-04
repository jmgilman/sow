package cmdutil

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/schemas/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test GetArtifactByIndex - retrieves artifact from a collection by index.
func TestGetArtifactByIndex(t *testing.T) {
	artifacts := []state.Artifact{
		{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		{ArtifactState: project.ArtifactState{Type: "design", Path: "design1.md"}},
		{ArtifactState: project.ArtifactState{Type: "task", Path: "task1.md"}},
	}

	t.Run("valid index 0", func(t *testing.T) {
		artifact, err := GetArtifactByIndex(artifacts, 0)
		require.NoError(t, err)
		assert.Equal(t, "review", artifact.Type)
	})

	t.Run("valid index 1", func(t *testing.T) {
		artifact, err := GetArtifactByIndex(artifacts, 1)
		require.NoError(t, err)
		assert.Equal(t, "design", artifact.Type)
	})

	t.Run("valid index 2", func(t *testing.T) {
		artifact, err := GetArtifactByIndex(artifacts, 2)
		require.NoError(t, err)
		assert.Equal(t, "task", artifact.Type)
	})

	t.Run("negative index", func(t *testing.T) {
		_, err := GetArtifactByIndex(artifacts, -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("index too large", func(t *testing.T) {
		_, err := GetArtifactByIndex(artifacts, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("empty collection", func(t *testing.T) {
		_, err := GetArtifactByIndex([]state.Artifact{}, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})
}

// Test UpdateArtifactByIndex - updates artifact in a collection by index.
func TestUpdateArtifactByIndex(t *testing.T) {
	t.Run("update artifact in collection", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: false}},
			{ArtifactState: project.ArtifactState{Type: "design", Path: "design1.md", Approved: false}},
		}

		updatedArtifact := state.Artifact{
			ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: true},
		}

		err := UpdateArtifactByIndex(&artifacts, 0, updatedArtifact)
		require.NoError(t, err)
		assert.True(t, artifacts[0].Approved)
		assert.False(t, artifacts[1].Approved) // Second artifact unchanged
	})

	t.Run("error on invalid index", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		}

		updatedArtifact := state.Artifact{
			ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: true},
		}

		err := UpdateArtifactByIndex(&artifacts, 5, updatedArtifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("error on negative index", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		}

		updatedArtifact := state.Artifact{
			ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: true},
		}

		err := UpdateArtifactByIndex(&artifacts, -1, updatedArtifact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})
}

// Test SetArtifactField - sets a field on an artifact in a collection.
func TestSetArtifactField(t *testing.T) {
	t.Run("set approved field", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: false}},
		}

		err := SetArtifactField(&artifacts, 0, "approved", "true")
		require.NoError(t, err)
		assert.True(t, artifacts[0].Approved)
	})

	t.Run("set metadata field", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		}

		err := SetArtifactField(&artifacts, 0, "metadata.assessment", "pass")
		require.NoError(t, err)
		assert.Equal(t, "pass", artifacts[0].Metadata["assessment"])
	})

	t.Run("error on invalid index", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		}

		err := SetArtifactField(&artifacts, 10, "approved", "true")
		assert.Error(t, err)
	})
}

// Test GetArtifactField - gets a field from an artifact in a collection.
func TestGetArtifactField(t *testing.T) {
	t.Run("get approved field", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md", Approved: true}},
		}

		value, err := GetArtifactField(artifacts, 0, "approved")
		require.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("get metadata field", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{
				Type: "review",
				Path: "review1.md",
				Metadata: map[string]any{
					"assessment": "pass",
				},
			}},
		}

		value, err := GetArtifactField(artifacts, 0, "metadata.assessment")
		require.NoError(t, err)
		assert.Equal(t, "pass", value)
	})

	t.Run("error on invalid index", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review1.md"}},
		}

		_, err := GetArtifactField(artifacts, 10, "approved")
		assert.Error(t, err)
	})
}

// Test FormatArtifactList - formats artifacts for display.
func TestFormatArtifactList(t *testing.T) {
	t.Run("format single artifact", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "phases/review/review.md", Approved: true}},
		}

		output := FormatArtifactList(artifacts)
		assert.Contains(t, output, "0")
		assert.Contains(t, output, "review")
		assert.Contains(t, output, "phases/review/review.md")
		assert.Contains(t, output, "true")
	})

	t.Run("format multiple artifacts", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{Type: "review", Path: "review.md", Approved: true}},
			{ArtifactState: project.ArtifactState{Type: "design", Path: "design.md", Approved: false}},
			{ArtifactState: project.ArtifactState{Type: "task", Path: "task.md", Approved: false}},
		}

		output := FormatArtifactList(artifacts)
		assert.Contains(t, output, "0")
		assert.Contains(t, output, "1")
		assert.Contains(t, output, "2")
		assert.Contains(t, output, "review")
		assert.Contains(t, output, "design")
		assert.Contains(t, output, "task")
	})

	t.Run("empty list", func(t *testing.T) {
		output := FormatArtifactList([]state.Artifact{})
		assert.Contains(t, output, "No artifacts")
	})

	t.Run("format with metadata", func(t *testing.T) {
		artifacts := []state.Artifact{
			{ArtifactState: project.ArtifactState{
				Type: "review",
				Path: "review.md",
				Metadata: map[string]any{
					"assessment": "pass",
					"reviewer":   "alice",
				},
			}},
		}

		output := FormatArtifactList(artifacts)
		assert.Contains(t, output, "review")
		// Metadata should be shown when present
		assert.Contains(t, output, "assessment")
	})
}

// Test FormatArtifact - formats single artifact for display.
func TestFormatArtifact(t *testing.T) {
	t.Run("format basic artifact", func(t *testing.T) {
		artifact := state.Artifact{
			ArtifactState: project.ArtifactState{
				Type:     "review",
				Path:     "phases/review/review.md",
				Approved: true,
			},
		}

		output := FormatArtifact(artifact)
		assert.Contains(t, output, "review")
		assert.Contains(t, output, "phases/review/review.md")
		assert.Contains(t, output, "true")
	})

	t.Run("format artifact with metadata", func(t *testing.T) {
		artifact := state.Artifact{
			ArtifactState: project.ArtifactState{
				Type: "review",
				Path: "review.md",
				Metadata: map[string]any{
					"assessment": "pass",
				},
			},
		}

		output := FormatArtifact(artifact)
		assert.Contains(t, output, "review")
		assert.Contains(t, output, "assessment")
		assert.Contains(t, output, "pass")
	})
}

// Test IndexInRange - validates index is within bounds.
func TestIndexInRange(t *testing.T) {
	t.Run("valid index", func(t *testing.T) {
		err := IndexInRange(5, 0)
		assert.NoError(t, err)
	})

	t.Run("valid index at boundary", func(t *testing.T) {
		err := IndexInRange(5, 4)
		assert.NoError(t, err)
	})

	t.Run("invalid negative index", func(t *testing.T) {
		err := IndexInRange(5, -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("invalid index too large", func(t *testing.T) {
		err := IndexInRange(5, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("zero length collection", func(t *testing.T) {
		err := IndexInRange(0, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index out of range")
	})
}
