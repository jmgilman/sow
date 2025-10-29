package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
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
func (p *ImplementationPhase) ApproveArtifact(_ string) (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
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
func (p *ImplementationPhase) ApproveTasks() (*domain.PhaseOperationResult, error) {
	if err := p.tasks.Approve(); err != nil {
		return nil, fmt.Errorf("failed to approve tasks: %w", err)
	}

	// Save before returning event
	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// Return event to be fired by CLI
	return domain.WithEvent(EventTasksApproved), nil
}

// Set sets a metadata field in the implementation phase.
func (p *ImplementationPhase) Set(field string, value interface{}) (*domain.PhaseOperationResult, error) {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return nil, err
	}
	return domain.NoEvent(), nil
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
func (p *ImplementationPhase) Complete() (*domain.PhaseOperationResult, error) {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// Return event to be fired by CLI
	return domain.WithEvent(EventAllTasksComplete), nil
}

// Skip is not supported as implementation phase is required.
func (p *ImplementationPhase) Skip() error {
	return project.ErrNotSupported // Implementation is required
}

// Enable is not supported as implementation phase is always enabled.
func (p *ImplementationPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Implementation is always enabled
}
