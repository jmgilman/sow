package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
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
	// Set tasks_approved in typed field (used by guards)
	tasksApproved := true
	p.project.state.Phases.Implementation.Tasks_approved = &tasksApproved

	if err := p.project.Save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// Set sets a metadata field in the implementation phase.
func (p *ImplementationPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return err
	}
	return nil
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

// Complete is not supported - use Advance() instead.
func (p *ImplementationPhase) Complete() error {
	return project.ErrNotSupported
}

// Skip is not supported as implementation phase is required.
func (p *ImplementationPhase) Skip() error {
	return project.ErrNotSupported // Implementation is required
}

// Enable is not supported as implementation phase is always enabled.
func (p *ImplementationPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Implementation is always enabled
}

// Advance progresses the implementation phase by examining the current state
// and firing the appropriate event.
func (p *ImplementationPhase) Advance() error {
	machine := p.project.Machine()
	currentState := machine.State()

	// Determine event based on state machine state
	var event statechart.Event
	switch currentState {
	case ImplementationPlanning:
		event = EventTasksApproved
	case ImplementationExecuting:
		event = EventAllTasksComplete
	default:
		return fmt.Errorf("%w: %s", project.ErrUnexpectedState, currentState)
	}

	if err := machine.Fire(event); err != nil {
		return fmt.Errorf("%w: cannot advance from %s: %w", project.ErrCannotAdvance, currentState, err)
	}

	if err := p.project.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}
