package standard

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// ReviewPhase implements the review phase for standard projects.
type ReviewPhase struct {
	state     *phasesSchema.Phase // Generic schema!
	artifacts *project.ArtifactCollection
	project   *StandardProject
	ctx       *sow.Context
}

// NewReviewPhase creates a new review phase.
func NewReviewPhase(state *phasesSchema.Phase, proj *StandardProject, ctx *sow.Context) *ReviewPhase {
	return &ReviewPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Implements Phase interface - delegate to helpers

// Name returns the name of the phase.
func (p *ReviewPhase) Name() string {
	return "review"
}

// Status returns the current status of the phase.
func (p *ReviewPhase) Status() string {
	return p.state.Status
}

// Enabled returns whether the phase is enabled.
func (p *ReviewPhase) Enabled() bool {
	return p.state.Enabled
}

// AddArtifact adds an artifact to the review phase.
func (p *ReviewPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	if err := p.artifacts.Add(path, opts...); err != nil {
		return fmt.Errorf("failed to add artifact: %w", err)
	}
	return nil
}

// ApproveArtifact approves an artifact in the review phase.
func (p *ReviewPhase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
	if err := p.artifacts.Approve(path); err != nil {
		return nil, fmt.Errorf("failed to approve artifact: %w", err)
	}
	return domain.NoEvent(), nil
}

// ListArtifacts returns all artifacts in the review phase.
func (p *ReviewPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// AddTask is not supported in the review phase.
func (p *ReviewPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// GetTask is not supported in the review phase.
func (p *ReviewPhase) GetTask(_ string) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

// ListTasks returns an empty list as tasks are not supported in review phase.
func (p *ReviewPhase) ListTasks() []*domain.Task {
	return []*domain.Task{}
}

// ApproveTasks is not supported in the review phase.
func (p *ReviewPhase) ApproveTasks() (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
}

// Set sets a metadata field in the review phase.
func (p *ReviewPhase) Set(field string, value interface{}) (*domain.PhaseOperationResult, error) {
	if p.state.Metadata == nil {
		p.state.Metadata = make(map[string]interface{})
	}
	p.state.Metadata[field] = value
	if err := p.project.Save(); err != nil {
		return nil, err
	}
	return domain.NoEvent(), nil
}

// Get retrieves a metadata field from the review phase.
func (p *ReviewPhase) Get(field string) (interface{}, error) {
	if p.state.Metadata == nil {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	val, ok := p.state.Metadata[field]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}
	return val, nil
}

// Complete marks the review phase as completed.
// The event fired depends on the assessment of the latest approved review artifact:
// - "pass" fires EventReviewPass (transitions to FinalizeDocumentation)
// - "fail" fires EventReviewFail (transitions back to ImplementationPlanning).
func (p *ReviewPhase) Complete() (*domain.PhaseOperationResult, error) {
	// Find the latest approved review artifact
	var latestReview *phasesSchema.Artifact
	for i := len(p.state.Artifacts) - 1; i >= 0; i-- {
		artifact := &p.state.Artifacts[i]
		if artifact.Type != nil && *artifact.Type == "review" && artifact.Approved != nil && *artifact.Approved {
			latestReview = artifact
			break
		}
	}

	if latestReview == nil {
		return nil, fmt.Errorf("cannot complete review: no approved review artifact found")
	}

	// Extract assessment from typed field
	if latestReview.Assessment == nil {
		return nil, fmt.Errorf("cannot complete review: review artifact missing assessment")
	}
	assessment := *latestReview.Assessment

	// Update status and timestamps
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// Return appropriate event based on assessment
	switch assessment {
	case "pass":
		return domain.WithEvent(EventReviewPass), nil
	case "fail":
		return domain.WithEvent(EventReviewFail), nil
	default:
		return nil, fmt.Errorf("invalid assessment value: %s (must be 'pass' or 'fail')", assessment)
	}
}

// Skip is not supported as review phase is required.
func (p *ReviewPhase) Skip() error {
	return project.ErrNotSupported // Review is required
}

// Enable is not supported as review phase is always enabled.
func (p *ReviewPhase) Enable(_ ...domain.PhaseOption) error {
	return project.ErrNotSupported // Review is always enabled
}

// Advance is not supported as review phase has no internal states.
func (p *ReviewPhase) Advance() (*domain.PhaseOperationResult, error) {
	return nil, project.ErrNotSupported
}

// AllReviewsApproved checks if all review artifacts have been approved.
func (p *ReviewPhase) AllReviewsApproved() bool {
	// Check for artifacts with type=review that aren't approved
	for _, artifact := range p.state.Artifacts {
		if artifact.Type != nil && *artifact.Type == "review" && (artifact.Approved == nil || !*artifact.Approved) {
			return false
		}
	}
	return true
}

// Iteration returns the current iteration count from typed field.
func (p *ReviewPhase) Iteration() int {
	// Get iteration from typed field in review phase
	reviewPhase := (*projects.ReviewPhase)(unsafe.Pointer(p.state))
	if reviewPhase.Iteration != nil {
		return *reviewPhase.Iteration
	}
	return 1
}
