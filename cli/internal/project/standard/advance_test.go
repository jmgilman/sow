package standard

import (
	"errors"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// TestAdvance_AllPhasesReturnErrNotSupported verifies all standard project phases
// return ErrNotSupported for Advance() since they don't have internal states.
func TestAdvance_AllPhasesReturnErrNotSupported(t *testing.T) {
	ctx := setupTestRepo(t)
	now := time.Now()

	phaseState := &phasesSchema.Phase{
		Status:     "in_progress",
		Created_at: now,
		Enabled:    true,
	}

	state := &projects.StandardProjectState{}
	proj := New(state, ctx)

	tests := []struct {
		name  string
		phase domain.Phase
	}{
		{
			name:  "PlanningPhase",
			phase: NewPlanningPhase(phaseState, proj, ctx),
		},
		{
			name:  "ImplementationPhase",
			phase: NewImplementationPhase(phaseState, proj, ctx),
		},
		{
			name:  "ReviewPhase",
			phase: NewReviewPhase(phaseState, proj, ctx),
		},
		{
			name:  "FinalizePhase",
			phase: NewFinalizePhase(phaseState, proj, ctx),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.phase.Advance()

			// Should return ErrNotSupported
			if !errors.Is(err, project.ErrNotSupported) {
				t.Errorf("Expected ErrNotSupported, got: %v", err)
			}

			// Result should be nil
			if result != nil {
				t.Errorf("Expected nil result with ErrNotSupported, got: %+v", result)
			}
		})
	}
}
