package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/schemas"
)

// DetermineActivePhase returns the current active phase name and its status.
// It checks phases in order (discovery → design → implementation → review → finalize)
// and returns the first phase that is not completed or skipped.
//
// For optional phases (discovery, design):
// - If enabled and not completed/skipped: returns phase with current status
// - If not enabled and status is "pending": returns phase with "pending" (decision state)
// - If not enabled and status is "skipped": skips to next phase
//
// Returns ("unknown", "unknown") if no active phase is found (e.g., all phases completed).
func DetermineActivePhase(state *schemas.ProjectState) (phaseName string, phaseStatus string) {
	// Check phases in order

	// Discovery (optional)
	// If skipped or completed, continue to next phase.
	// Otherwise, either in pending decision state or active (enabled).
	if state.Phases.Discovery.Status != "skipped" && state.Phases.Discovery.Status != "completed" {
		return "discovery", state.Phases.Discovery.Status
	}

	// Design (optional)
	// If skipped or completed, continue to next phase.
	// Otherwise, either in pending decision state or active (enabled).
	if state.Phases.Design.Status != "skipped" && state.Phases.Design.Status != "completed" {
		return "design", state.Phases.Design.Status
	}

	// Implementation (required)
	if state.Phases.Implementation.Status != "completed" && state.Phases.Implementation.Status != "skipped" {
		return "implementation", state.Phases.Implementation.Status
	}

	// Review (required)
	if state.Phases.Review.Status != "completed" && state.Phases.Review.Status != "skipped" {
		return "review", state.Phases.Review.Status
	}

	// Finalize (required)
	if state.Phases.Finalize.Status != "completed" && state.Phases.Finalize.Status != "skipped" {
		return "finalize", state.Phases.Finalize.Status
	}

	return "unknown", "unknown"
}

// GetPhaseMetadata returns the metadata for a specific phase in the project.
// This includes information about supported operations (tasks, artifacts) and custom fields.
//
// For now, this is hardcoded for StandardProject. In the future, this could be made
// dynamic based on the project type.
//
// Returns an error if the phase name is invalid.
func GetPhaseMetadata(phaseName string) (*phases.PhaseMetadata, error) {
	// Hardcoded metadata for standard project phases
	// In the future, this could be retrieved from the project type dynamically
	allPhases := map[string]phases.PhaseMetadata{
		"discovery": {
			Name:              "discovery",
			SupportsTasks:     false,
			SupportsArtifacts: true,
			CustomFields: []phases.FieldDef{
				{Name: "discovery_type", Type: phases.StringField, Description: "Type of discovery work"},
			},
		},
		"design": {
			Name:              "design",
			SupportsTasks:     false,
			SupportsArtifacts: true,
			CustomFields: []phases.FieldDef{
				{Name: "architect_used", Type: phases.BoolField, Description: "Whether architect agent was used"},
			},
		},
		"implementation": {
			Name:              "implementation",
			SupportsTasks:     true,
			SupportsArtifacts: false,
			CustomFields: []phases.FieldDef{
				{Name: "planner_used", Type: phases.BoolField, Description: "Whether planner agent was used"},
				{Name: "tasks_approved", Type: phases.BoolField, Description: "Whether task plan is approved"},
			},
		},
		"review": {
			Name:              "review",
			SupportsTasks:     false,
			SupportsArtifacts: false,
			CustomFields: []phases.FieldDef{
				{Name: "iteration", Type: phases.IntField, Description: "Review iteration number"},
			},
		},
		"finalize": {
			Name:              "finalize",
			SupportsTasks:     false,
			SupportsArtifacts: false,
			CustomFields: []phases.FieldDef{
				{Name: "project_deleted", Type: phases.BoolField, Description: "Whether project has been deleted"},
				{Name: "pr_url", Type: phases.StringField, Description: "Pull request URL"},
			},
		},
	}

	// Look up the requested phase
	metadata, exists := allPhases[phaseName]
	if !exists {
		return nil, fmt.Errorf("unknown phase: %s", phaseName)
	}

	return &metadata, nil
}
