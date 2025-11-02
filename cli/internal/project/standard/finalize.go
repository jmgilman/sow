// Package standard implements the standard project type for sow.
//
// The standard project follows a 4-phase lifecycle:
// Planning → Implementation → Review → Finalize
package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
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

// Name returns the name of the phase.
func (p *FinalizePhase) Name() string {
	return "finalize"
}

// Status returns the current status of the phase.
func (p *FinalizePhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *FinalizePhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact is not supported in the finalize phase.
func (p *FinalizePhase) AddArtifact(_ string, _ ...domain.ArtifactOption) error {
	return project.ErrNotSupported
}

// ApproveArtifact is not supported in the finalize phase.
func (p *FinalizePhase) ApproveArtifact(_ string) error {
	return project.ErrNotSupported
}

// ListArtifacts returns an empty list as artifacts are not supported in finalize phase.
func (p *FinalizePhase) ListArtifacts() []*phasesSchema.Artifact {
	return []*phasesSchema.Artifact{}
}

// AddTask is not supported in the finalize phase.
func (p *FinalizePhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the finalize phase.
func (p *FinalizePhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in finalize phase.
func (p *FinalizePhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the finalize phase.
func (p *FinalizePhase) ApproveTasks() error {
	return project.ErrNotSupported
}

// Set sets a metadata field in the finalize phase.
func (p *FinalizePhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return err
	}
	return nil
}

// Get retrieves a metadata field from the finalize phase.
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

// Complete is not supported - use Advance() instead.
func (p *FinalizePhase) Complete() error {
	return project.ErrNotSupported
}

// Skip is not supported as finalize phase is required.
func (p *FinalizePhase) Skip() error {
	return project.ErrNotSupported // Finalize is required
}

// Enable is not supported as finalize phase is always enabled.
func (p *FinalizePhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Finalize is always enabled
}

// Advance progresses the finalize phase to the next substate.
// Examines the current finalize substate and fires the appropriate event.
func (p *FinalizePhase) Advance() error {
	machine := p.project.Machine()
	currentState := machine.State()

	// Determine event based on finalize substate
	var event statechart.Event
	switch currentState {
	case FinalizeDocumentation:
		event = EventDocumentationDone
	case FinalizeChecks:
		event = EventChecksDone
	case FinalizeDelete:
		event = EventProjectDelete
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
