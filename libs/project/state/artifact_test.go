package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/stretchr/testify/assert"
)

func TestArtifact_EmbeddedState(t *testing.T) {
	t.Run("embeds ArtifactState fields", func(t *testing.T) {
		now := time.Now()
		artifact := Artifact{
			ArtifactState: project.ArtifactState{
				Type:       "review",
				Path:       "/path/to/review.md",
				Approved:   true,
				Created_at: now,
				Metadata:   map[string]any{"assessment": "pass"},
			},
		}

		assert.Equal(t, "review", artifact.Type)
		assert.Equal(t, "/path/to/review.md", artifact.Path)
		assert.True(t, artifact.Approved)
		assert.Equal(t, now, artifact.Created_at)
		assert.Equal(t, "pass", artifact.Metadata["assessment"])
	})

	t.Run("handles nil metadata", func(t *testing.T) {
		artifact := Artifact{
			ArtifactState: project.ArtifactState{
				Type:     "design_doc",
				Path:     "/path/to/design.md",
				Metadata: nil,
			},
		}

		assert.Nil(t, artifact.Metadata)
	})
}
