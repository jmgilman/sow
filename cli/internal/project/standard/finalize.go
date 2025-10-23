package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// FinalizePhase implements the finalize phase for standard projects.
type FinalizePhase struct {
	state   *phasesSchema.Phase
	project *StandardProject
	ctx     *sow.Context
}

// NewFinalizePhase creates a new finalize phase.
func NewFinalizePhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *FinalizePhase {
	return &FinalizePhase{
		state:   state,
		project: proj,
		ctx:     ctx,
	}
}

func (p *FinalizePhase) Name() string {
	return "finalize"
}

func (p *FinalizePhase) Status() string {
	return p.state.Status
}

func (p *FinalizePhase) Enabled() bool {
	return p.state.Enabled
}

func (p *FinalizePhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return project.ErrNotSupported
}

func (p *FinalizePhase) ApproveArtifact(path string) error {
	return project.ErrNotSupported
}

func (p *FinalizePhase) ListArtifacts() []*phasesSchema.Artifact {
	return []*phasesSchema.Artifact{}
}

func (p *FinalizePhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *FinalizePhase) GetTask(id string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *FinalizePhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

func (p *FinalizePhase) ApproveTasks() error {
	return project.ErrNotSupported
}

func (p *FinalizePhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

func (p *FinalizePhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

func (p *FinalizePhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	// Finalize is the last phase - no state transition needed
	return p.project.Save()
}

func (p *FinalizePhase) Skip() error {
	return project.ErrNotSupported // Finalize is required
}

func (p *FinalizePhase) Enable(opts ...domain.PhaseOption) error {
	return project.ErrNotSupported // Finalize is always enabled
}
