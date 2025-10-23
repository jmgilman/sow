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

// Name returns the name of the phase.
func (p *ImplementationPhase) Name() string {
	return "implementation"
}

// Status returns the current status of the phase.
func (p *ImplementationPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *ImplementationPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact is not supported in the implementation phase.
func (p *ImplementationPhase) AddArtifact(_ string, _ ...domain.ArtifactOption) error {
	return project.ErrNotSupported
}

// ApproveArtifact is not supported in the implementation phase.
func (p *ImplementationPhase) ApproveArtifact(_ string) error {
	return project.ErrNotSupported
}

// ListArtifacts returns an empty list as artifacts are not supported in implementation phase.
func (p *ImplementationPhase) ListArtifacts() []*phasesSchema.Artifact {
	return []*phasesSchema.Artifact{}
}

// AddTask adds a task to the implementation phase.
func (p *ImplementationPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	task, err := p.tasks.Add(name, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to add task: %w", err)
	}
	return task, nil
}

// GetTask retrieves a task by ID from the implementation phase.
func (p *ImplementationPhase) GetTask(id string) (*domain.Task, error) {
	task, err := p.tasks.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks returns all tasks in the implementation phase.
func (p *ImplementationPhase) ListTasks() []*domain.Task {
	return p.tasks.List()
}

// ApproveTasks approves all tasks in the implementation phase for execution.
func (p *ImplementationPhase) ApproveTasks() error {
	if err := p.tasks.Approve(); err != nil {
		return fmt.Errorf("failed to approve tasks: %w", err)
	}

	// Fire state machine event
	if err := p.project.Machine().Fire(statechart.EventTasksApproved); err != nil {
		return fmt.Errorf("failed to fire tasks approved event: %w", err)
	}

	// Save after firing to persist the machine state transition
	return p.project.Save()
}

// Set sets a metadata field in the implementation phase.
func (p *ImplementationPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

// Get retrieves a metadata field from the implementation phase.
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

// Complete marks the implementation phase as completed.
func (p *ImplementationPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventAllTasksComplete); err != nil {
		return fmt.Errorf("failed to fire all tasks complete event: %w", err)
	}

	return p.project.Save()
}

// Skip is not supported as implementation phase is required.
func (p *ImplementationPhase) Skip() error {
	return project.ErrNotSupported // Implementation is required
}

// Enable is not supported as implementation phase is always enabled.
func (p *ImplementationPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Implementation is always enabled
}
