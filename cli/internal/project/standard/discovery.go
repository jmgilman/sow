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

// DiscoveryPhase implements the discovery phase for standard projects.
type DiscoveryPhase struct {
	state     *phasesSchema.Phase
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewDiscoveryPhase creates a new discovery phase.
func NewDiscoveryPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *DiscoveryPhase {
	return &DiscoveryPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Name returns the name of the phase.
func (p *DiscoveryPhase) Name() string {
	return "discovery"
}

// Status returns the current status of the phase.
func (p *DiscoveryPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *DiscoveryPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the discovery phase.
func (p *DiscoveryPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the discovery phase.
func (p *DiscoveryPhase) ApproveArtifact(path string) error {
	if err := p.artifacts.Approve(path); err != nil {
		return fmt.Errorf("failed to approve artifact: %w", err)
	}
	return nil
}

// ListArtifacts returns all artifacts in the discovery phase.
func (p *DiscoveryPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the discovery phase.
func (p *DiscoveryPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the discovery phase.
func (p *DiscoveryPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in discovery phase.
func (p *DiscoveryPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the discovery phase.
func (p *DiscoveryPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

// Set sets a metadata field in the discovery phase.
func (p *DiscoveryPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

// Get retrieves a metadata field from the discovery phase.
func (p *DiscoveryPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete marks the discovery phase as completed.
func (p *DiscoveryPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventCompleteDiscovery); err != nil {
		return fmt.Errorf("failed to fire complete discovery event: %w", err)
	}

	return p.project.Save()
}

// Skip marks the discovery phase as skipped.
func (p *DiscoveryPhase) Skip() error {
	p.state.Status = "skipped"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventSkipDiscovery); err != nil {
		return fmt.Errorf("failed to fire skip discovery event: %w", err)
	}

	return p.project.Save()
}

// Enable enables the discovery phase.
func (p *DiscoveryPhase) Enable(opts ...domain.PhaseOption) error {
	cfg := &domain.PhaseConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	p.state.Enabled = true
	p.state.Status = "in_progress"
	now := time.Now()
	p.state.Started_at = &now

	if cfg.Metadata != nil {
		if p.state.Metadata == nil {
			p.state.Metadata = make(map[string]interface{})
		}
		for k, v := range cfg.Metadata {
			p.state.Metadata[k] = v
		}
	}

	// Fire event first to transition the machine
	if err := p.project.Machine().Fire(statechart.EventEnableDiscovery); err != nil {
		return fmt.Errorf("failed to fire enable discovery event: %w", err)
	}

	// Then save the state (including the new machine state)
	return p.project.Save()
}
