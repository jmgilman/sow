// Package standard implements the standard project type with a 5-phase workflow.
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

// DesignPhase implements the design phase for standard projects.
type DesignPhase struct {
	state     *phasesSchema.Phase
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewDesignPhase creates a new design phase.
func NewDesignPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *DesignPhase {
	return &DesignPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Name returns the name of the phase.
func (p *DesignPhase) Name() string {
	return "design"
}

// Status returns the current status of the phase.
func (p *DesignPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *DesignPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the design phase.
func (p *DesignPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the design phase.
func (p *DesignPhase) ApproveArtifact(path string) error {
	if err := p.artifacts.Approve(path); err != nil {
		return fmt.Errorf("failed to approve artifact: %w", err)
	}
	return nil
}

// ListArtifacts returns all artifacts in the design phase.
func (p *DesignPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the design phase.
func (p *DesignPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the design phase.
func (p *DesignPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in design phase.
func (p *DesignPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the design phase.
func (p *DesignPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

// Set sets a metadata field in the design phase.
func (p *DesignPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

// Get retrieves a metadata field from the design phase.
func (p *DesignPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete marks the design phase as completed.
func (p *DesignPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventCompleteDesign); err != nil {
		return fmt.Errorf("failed to fire complete design event: %w", err)
	}

	return p.project.Save()
}

// Skip marks the design phase as skipped.
func (p *DesignPhase) Skip() error {
	p.state.Status = "skipped"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventSkipDesign); err != nil {
		return fmt.Errorf("failed to fire skip design event: %w", err)
	}

	return p.project.Save()
}

// Enable enables the design phase.
func (p *DesignPhase) Enable(opts ...domain.PhaseOption) error {
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

	if err := p.project.Machine().Fire(statechart.EventEnableDesign); err != nil {
		return fmt.Errorf("failed to fire enable design event: %w", err)
	}

	return p.project.Save()
}
