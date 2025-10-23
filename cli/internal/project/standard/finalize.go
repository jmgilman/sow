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
	return p.project.Save()
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

// Complete marks the finalize phase as completed.
func (p *FinalizePhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	// Finalize is the last phase - no state transition needed
	return p.project.Save()
}

// Skip is not supported as finalize phase is required.
func (p *FinalizePhase) Skip() error {
	return project.ErrNotSupported // Finalize is required
}

// Enable is not supported as finalize phase is always enabled.
func (p *FinalizePhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Finalize is always enabled
}
