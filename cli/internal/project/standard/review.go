package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// ReviewPhase implements the review phase for standard projects.
type ReviewPhase struct {
	state     *phasesSchema.Phase // Generic schema!
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewReviewPhase creates a new review phase.
func NewReviewPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *ReviewPhase {
	return &ReviewPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Implements Phase interface - delegate to helpers

// Name returns the name of the phase.
func (p *ReviewPhase) Name() string {
	return "review"
}

// Status returns the current status of the phase.
func (p *ReviewPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *ReviewPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the review phase.
func (p *ReviewPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the review phase.
func (p *ReviewPhase) ApproveArtifact(path string) error {
	if err := p.artifacts.Approve(path); err != nil {
		return fmt.Errorf("failed to approve artifact: %w", err)
	}
	return nil
}

// ListArtifacts returns all artifacts in the review phase.
func (p *ReviewPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the review phase.
func (p *ReviewPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the review phase.
func (p *ReviewPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in review phase.
func (p *ReviewPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the review phase.
func (p *ReviewPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

// Set sets a metadata field in the review phase.
func (p *ReviewPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

// Get retrieves a metadata field from the review phase.
func (p *ReviewPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete marks the review phase as completed.
func (p *ReviewPhase) Complete() error {
	// Update status and timestamps
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	// Fire state machine event
	if err := p.project.Machine().Fire(statechart.EventReviewPass); err != nil {
		return fmt.Errorf("failed to fire review pass event: %w", err)
	}

	return p.project.Save()
}

// Skip is not supported as review phase is required.
func (p *ReviewPhase) Skip() error {
	return project.ErrNotSupported // Review is required
}

// Enable is not supported as review phase is always enabled.
func (p *ReviewPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Review is always enabled
}

// AllReviewsApproved checks if all review artifacts have been approved.
func (p *ReviewPhase) AllReviewsApproved() bool {
	// Check for artifacts with type=review that aren't approved
	for _, artifact := range p.state.Artifacts {
		if artifactType, ok := artifact.Metadata["type"].(string); ok {
			if artifactType == "review" && !artifact.Approved {
				return false
			}
		}
	}
	return true
}

// Iteration returns the current iteration count from metadata.
func (p *ReviewPhase) Iteration() int {
	if p.state.Metadata == nil {
		return 1
	}
	if iter, ok := p.state.Metadata["iteration"].(int); ok {
		return iter
	}
	return 1
}
