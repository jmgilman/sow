package standard

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

// PlanningPhase implements the planning phase for standard projects.
type PlanningPhase struct {
	state     *phasesSchema.Phase
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewPlanningPhase creates a new planning phase.
func NewPlanningPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *PlanningPhase {
	return &PlanningPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Name returns the name of the phase.
func (p *PlanningPhase) Name() string {
	return "planning"
}

// Status returns the current status of the phase.
func (p *PlanningPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *PlanningPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the planning phase.
func (p *PlanningPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the planning phase.
func (p *PlanningPhase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
	if err := p.artifacts.Approve(path); err != nil {
		return nil, fmt.Errorf("failed to approve artifact: %w", err)
	}
	return domain.NoEvent(), nil
}

// ListArtifacts returns all artifacts in the planning phase.
func (p *PlanningPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the planning phase.
func (p *PlanningPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the planning phase.
func (p *PlanningPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in planning phase.
func (p *PlanningPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the planning phase.
func (p *PlanningPhase) ApproveTasks() (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
}

// Set sets a metadata field in the planning phase.
func (p *PlanningPhase) Set(field string, value interface{}) (*domain.PhaseOperationResult, error) {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return nil, err
	}
	return domain.NoEvent(), nil
}

// Get retrieves a metadata field from the planning phase.
func (p *PlanningPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete marks the planning phase as completed.
// This requires that the task list artifact has been approved.
func (p *PlanningPhase) Complete() (*domain.PhaseOperationResult, error) {
	// Verify that the task list artifact is approved
	taskListApproved := false
	for _, artifact := range p.state.Artifacts {
		if artifact.Type != nil && *artifact.Type == "task_list" && artifact.Approved {
			taskListApproved = true
			break
		}
	}

	if !taskListApproved {
		return nil, fmt.Errorf("cannot complete planning: task list artifact must be approved")
	}

	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// Return event to be fired by CLI
	return domain.WithEvent(EventCompletePlanning), nil
}

// Skip is not supported as planning phase is required.
func (p *PlanningPhase) Skip() error {
	return project.ErrNotSupported // Planning is required
}

// Enable is not supported as planning phase is always enabled.
func (p *PlanningPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Planning is always enabled
}
