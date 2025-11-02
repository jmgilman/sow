package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// PlanningPhase implements the planning phase for standard projects.
type PlanningPhase struct {
	state     *phasesSchema.Phase
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewPlanningPhase creates a new planning phase.
func NewPlanningPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *PlanningPhase {
	return &PlanningPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Name returns the name of the phase.
func (p *PlanningPhase) Name() string {
	return "planning"
}

// Status returns the current status of the phase.
func (p *PlanningPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *PlanningPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the planning phase.
func (p *PlanningPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the planning phase.
func (p *PlanningPhase) ApproveArtifact(path string) error {
	if err := p.artifacts.Approve(path); err != nil {
		return fmt.Errorf("failed to approve artifact: %w", err)
	}
	return nil
}

// ListArtifacts returns all artifacts in the planning phase.
func (p *PlanningPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the planning phase.
func (p *PlanningPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the planning phase.
func (p *PlanningPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in planning phase.
func (p *PlanningPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the planning phase.
func (p *PlanningPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

// Set sets a metadata field in the planning phase.
func (p *PlanningPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return err
	}
	return nil
}

// Get retrieves a metadata field from the planning phase.
func (p *PlanningPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete is not supported - use Advance() instead.
func (p *PlanningPhase) Complete() error {
	return project.ErrNotSupported
}

// Skip is not supported as planning phase is required.
func (p *PlanningPhase) Skip() error {
	return project.ErrNotSupported // Planning is required
}

// Enable is not supported as planning phase is always enabled.
func (p *PlanningPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Planning is always enabled
}

// Advance moves the planning phase to implementation by firing EventCompletePlanning.
func (p *PlanningPhase) Advance() error {
	// Planning phase has only one possible transition
	machine := p.project.Machine()
	if err := machine.Fire(EventCompletePlanning); err != nil {
		return fmt.Errorf("%w: %w", project.ErrCannotAdvance, err)
	}

	if err := p.project.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}
