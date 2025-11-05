package state

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// Phase wraps the CUE-generated PhaseState.
// This is a pure data wrapper with no additional runtime fields.
type Phase struct {
	project.PhaseState
}

// IncrementPhaseIteration increments the iteration counter for a phase.
// If the iteration is not set, initializes it to 1.
// This is typically called during onEntry actions when re-entering a phase
// after a failure in a downstream phase.
//
// Example usage in transition actions:
//
//	project.WithOnEntry(func(p *state.Project) error {
//	    return state.IncrementPhaseIteration(p, "implementation")
//	})
func IncrementPhaseIteration(p *Project, phaseName string) error {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase %s not found", phaseName)
	}

	// Get current iteration (0 if not set)
	currentIter := phase.Iteration

	// Increment
	phase.Iteration = currentIter + 1
	p.Phases[phaseName] = phase
	return nil
}

// MarkPhaseFailed sets a phase's status to "failed" and records the failed_at timestamp.
// This is typically called during onExit actions when transitioning away from a failed phase.
//
// Example usage in transition actions:
//
//	project.WithOnExit(func(p *state.Project) error {
//	    return state.MarkPhaseFailed(p, "review")
//	})
func MarkPhaseFailed(p *Project, phaseName string) error {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return fmt.Errorf("phase %s not found", phaseName)
	}

	now := time.Now()
	phase.Status = "failed"
	phase.Failed_at = now
	p.Phases[phaseName] = phase
	return nil
}

// AddPhaseInputFromOutput adds an artifact from one phase's outputs as an input to another phase.
// This is useful for creating data flow between phases (e.g., failed review → implementation input).
//
// Example usage:
//
//	err := state.AddPhaseInputFromOutput(p, "review", "implementation", "review", func(a *project.ArtifactState) bool {
//	    assessment, ok := a.Metadata["assessment"].(string)
//	    return ok && assessment == "fail"
//	})
func AddPhaseInputFromOutput(
	p *Project,
	sourcePhaseName string,
	targetPhaseName string,
	artifactType string,
	filter func(*project.ArtifactState) bool,
) error {
	sourcePhase, exists := p.Phases[sourcePhaseName]
	if !exists {
		return fmt.Errorf("source phase %s not found", sourcePhaseName)
	}

	targetPhase, exists := p.Phases[targetPhaseName]
	if !exists {
		return fmt.Errorf("target phase %s not found", targetPhaseName)
	}

	// Find matching artifact in source outputs (search backwards for latest)
	var matchingArtifact *project.ArtifactState
	for i := len(sourcePhase.Outputs) - 1; i >= 0; i-- {
		artifact := &sourcePhase.Outputs[i]
		if artifact.Type == artifactType && (filter == nil || filter(artifact)) {
			matchingArtifact = artifact
			break
		}
	}

	if matchingArtifact == nil {
		return fmt.Errorf("no matching artifact of type %s found in %s outputs", artifactType, sourcePhaseName)
	}

	// Add as input to target phase
	targetPhase.Inputs = append(targetPhase.Inputs, *matchingArtifact)
	p.Phases[targetPhaseName] = targetPhase
	return nil
}
