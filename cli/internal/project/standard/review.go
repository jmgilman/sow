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

func (p *ReviewPhase) Name() string {
	return "review"
}

func (p *ReviewPhase) Status() string {
	return p.state.Status
}

func (p *ReviewPhase) Enabled() bool {
	return p.state.Enabled
}

func (p *ReviewPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return p.artifacts.Add(path, opts...)
}

func (p *ReviewPhase) ApproveArtifact(path string) error {
	return p.artifacts.Approve(path)
}

func (p *ReviewPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

func (p *ReviewPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *ReviewPhase) GetTask(id string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *ReviewPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

func (p *ReviewPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

func (p *ReviewPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

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

func (p *ReviewPhase) Complete() error {
	// Update status and timestamps
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	// Fire state machine event
	if err := p.project.Machine().Fire(statechart.EventReviewPass); err != nil {
		return err
	}

	return p.project.Save()
}

func (p *ReviewPhase) Skip() error {
	return project.ErrNotSupported // Review is required
}

func (p *ReviewPhase) Enable(opts ...domain.PhaseOption) error {
	return project.ErrNotSupported // Review is always enabled
}

// Phase-specific guard (used by state machine)
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

// Helper for accessing metadata
func (p *ReviewPhase) Iteration() int {
	if p.state.Metadata == nil {
		return 1
	}
	if iter, ok := p.state.Metadata["iteration"].(int); ok {
		return iter
	}
	return 1
}
