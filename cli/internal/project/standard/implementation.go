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

// ImplementationPhase implements the implementation phase for standard projects.
type ImplementationPhase struct {
	state   *phasesSchema.Phase
	tasks   *project.TaskCollection
	project *StandardProject
	ctx     *sow.Context
}

// NewImplementationPhase creates a new implementation phase.
func NewImplementationPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *ImplementationPhase {
	return &ImplementationPhase{
		state:   state,
		tasks:   project.NewTaskCollection(state, proj, ctx),
		project: proj,
		ctx:     ctx,
	}
}

func (p *ImplementationPhase) Name() string {
	return "implementation"
}

func (p *ImplementationPhase) Status() string {
	return p.state.Status
}

func (p *ImplementationPhase) Enabled() bool {
	return p.state.Enabled
}

func (p *ImplementationPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return project.ErrNotSupported
}

func (p *ImplementationPhase) ApproveArtifact(path string) error {
	return project.ErrNotSupported
}

func (p *ImplementationPhase) ListArtifacts() []*phasesSchema.Artifact {
	return []*phasesSchema.Artifact{}
}

func (p *ImplementationPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	return p.tasks.Add(name, opts...)
}

func (p *ImplementationPhase) GetTask(id string) (*domain.Task, error) {
	return p.tasks.Get(id)
}

func (p *ImplementationPhase) ListTasks() []*domain.Task {
	return p.tasks.List()
}

func (p *ImplementationPhase) ApproveTasks() error {
	if err := p.tasks.Approve(); err != nil {
		return err
	}

	// Fire state machine event
	if err := p.project.Machine().Fire(statechart.EventTasksApproved); err != nil {
		return err
	}

	// Save after firing to persist the machine state transition
	return p.project.Save()
}

func (p *ImplementationPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

func (p *ImplementationPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

func (p *ImplementationPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventAllTasksComplete); err != nil {
		return err
	}

	return p.project.Save()
}

func (p *ImplementationPhase) Skip() error {
	return project.ErrNotSupported // Implementation is required
}

func (p *ImplementationPhase) Enable(opts ...domain.PhaseOption) error {
	return project.ErrNotSupported // Implementation is always enabled
}
