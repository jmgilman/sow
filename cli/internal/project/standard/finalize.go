// Package standard implements the standard project type for sow.
//
// The standard project follows a 4-phase lifecycle:
// Planning → Implementation → Review → Finalize
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
func (p *FinalizePhase) ApproveArtifact(_ string) (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
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
func (p *FinalizePhase) ApproveTasks() (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
}

// Set sets a metadata field in the finalize phase.
func (p *FinalizePhase) Set(field string, value interface{}) (*domain.PhaseOperationResult, error) {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return nil, err
	}
	return domain.NoEvent(), nil
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

// Complete handles completion of the current finalize sub-state.
// Finalize has 3 sub-states that must be completed in sequence:
// FinalizeDocumentation → FinalizeChecks → FinalizeDelete.
func (p *FinalizePhase) Complete() (*domain.PhaseOperationResult, error) {
	// Get current state to determine which event to fire
	machine := p.project.Machine()
	currentState := machine.State()

	// Handle completion based on current state
	switch currentState {
	case FinalizeDocumentation:
		// Documentation work complete - transition to checks
		if err := p.project.Save(); err != nil {
			return nil, err
		}
		return domain.WithEvent(EventDocumentationDone), nil

	case FinalizeChecks:
		// Checks complete - transition to delete
		if err := p.project.Save(); err != nil {
			return nil, err
		}
		return domain.WithEvent(EventChecksDone), nil

	case FinalizeDelete:
		// Delete is the final step - mark phase as completed
		p.state.Status = "completed"
		now := time.Now()
		p.state.Completed_at = &now
		if err := p.project.Save(); err != nil {
			return nil, err
		}
		// Note: EventProjectDelete must be fired separately via `sow agent delete`
		return domain.NoEvent(), nil

	default:
		return nil, fmt.Errorf("unexpected state: %s", currentState)
	}
}

// Skip is not supported as finalize phase is required.
func (p *FinalizePhase) Skip() error {
	return project.ErrNotSupported // Finalize is required
}

// Enable is not supported as finalize phase is always enabled.
func (p *FinalizePhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Finalize is always enabled
}

// Advance is not supported as finalize phase has no internal states.
func (p *FinalizePhase) Advance() (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
}
