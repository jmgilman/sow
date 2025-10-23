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

func (p *DiscoveryPhase) Name() string {
	return "discovery"
}

func (p *DiscoveryPhase) Status() string {
	return p.state.Status
}

func (p *DiscoveryPhase) Enabled() bool {
	return p.state.Enabled
}

func (p *DiscoveryPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return p.artifacts.Add(path, opts...)
}

func (p *DiscoveryPhase) ApproveArtifact(path string) error {
	return p.artifacts.Approve(path)
}

func (p *DiscoveryPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

func (p *DiscoveryPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *DiscoveryPhase) GetTask(id string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *DiscoveryPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

func (p *DiscoveryPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

func (p *DiscoveryPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

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

func (p *DiscoveryPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventCompleteDiscovery); err != nil {
		return err
	}

	return p.project.Save()
}

func (p *DiscoveryPhase) Skip() error {
	p.state.Status = "skipped"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventSkipDiscovery); err != nil {
		return err
	}

	return p.project.Save()
}

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
		return err
	}

	// Then save the state (including the new machine state)
	return p.project.Save()
}
