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

func (p *DesignPhase) Name() string {
	return "design"
}

func (p *DesignPhase) Status() string {
	return p.state.Status
}

func (p *DesignPhase) Enabled() bool {
	return p.state.Enabled
}

func (p *DesignPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return p.artifacts.Add(path, opts...)
}

func (p *DesignPhase) ApproveArtifact(path string) error {
	return p.artifacts.Approve(path)
}

func (p *DesignPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

func (p *DesignPhase) AddTask(name string, opts ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *DesignPhase) GetTask(id string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *DesignPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

func (p *DesignPhase) ApproveTasks() error {
	return project.ErrNotSupported
}

func (p *DesignPhase) Set(field string, value interface{}) error {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	return p.project.Save()
}

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

func (p *DesignPhase) Complete() error {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventCompleteDesign); err != nil {
		return err
	}

	return p.project.Save()
}

func (p *DesignPhase) Skip() error {
	p.state.Status = "skipped"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Machine().Fire(statechart.EventSkipDesign); err != nil {
		return err
	}

	return p.project.Save()
}

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
		return err
	}

	return p.project.Save()
}
