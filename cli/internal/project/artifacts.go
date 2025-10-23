// Package project provides the core project management functionality for sow.
package project

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// ArtifactCollection provides artifact operations on a generic Phase.
// This helper eliminates duplication across phases that support artifacts.
type ArtifactCollection struct {
	state   *phases.Phase
	project domain.Project
}

// NewArtifactCollection creates a new artifact collection.
func NewArtifactCollection(state *phases.Phase, proj domain.Project) *ArtifactCollection {
	return &ArtifactCollection{
		state:   state,
		project: proj,
	}
}

// Add adds a new artifact to the phase.
func (ac *ArtifactCollection) Add(path string, opts ...domain.ArtifactOption) error {
	cfg := &domain.ArtifactConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	artifact := phases.Artifact{
		Path:       path,
		Approved:   false,
		Created_at: time.Now(),
		Metadata:   cfg.Metadata,
	}

	ac.state.Artifacts = append(ac.state.Artifacts, artifact)
	if err := ac.project.Save(); err != nil {
		return fmt.Errorf("failed to save project after adding artifact: %w", err)
	}
	return nil
}

// Approve marks an artifact as approved.
func (ac *ArtifactCollection) Approve(path string) error {
	for i := range ac.state.Artifacts {
		if ac.state.Artifacts[i].Path == path {
			ac.state.Artifacts[i].Approved = true
			if err := ac.project.Save(); err != nil {
				return fmt.Errorf("failed to save project after approving artifact: %w", err)
			}
			return nil
		}
	}
	return ErrArtifactNotFound
}

// List returns all artifacts in the phase.
func (ac *ArtifactCollection) List() []*phases.Artifact {
	result := make([]*phases.Artifact, len(ac.state.Artifacts))
	for i := range ac.state.Artifacts {
		result[i] = &ac.state.Artifacts[i]
	}
	return result
}

// AllApproved checks if all artifacts are approved.
func (ac *ArtifactCollection) AllApproved() bool {
	for _, a := range ac.state.Artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}

// HasArtifacts checks if the phase has any artifacts.
func (ac *ArtifactCollection) HasArtifacts() bool {
	return len(ac.state.Artifacts) > 0
}
