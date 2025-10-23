// Package types defines project type abstractions and registry.
//
//nolint:revive // "types" package name is intentional and follows Go conventions
package types

import (
	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// ProjectType represents a type of project with its own phase composition and lifecycle.
//
// Different project types can compose phases in different orders, add custom transitions,
// and provide different metadata. This interface allows the project package to work
// with any project type without knowing the specific implementation.
//
// Example project types:
//   - StandardProject: Discovery → Design → Implementation → Review → Finalize
//   - DesignProject: Research → Exploration → Decomposition → Cleanup
//   - SpikeProject: Investigation → Analysis → Recommendation
type ProjectType interface {
	// BuildStateMachine constructs and configures the state machine for this project type.
	// The machine is fully configured with all states, transitions, guards, and entry actions.
	//
	// This method:
	// 1. Creates phases with data from the project state
	// 2. Uses BuildPhaseChain to wire phases together
	// 3. Adds any exceptional transitions (e.g., backward loops)
	// 4. Returns a fully configured machine
	BuildStateMachine() *statechart.Machine

	// Phases returns metadata for all phases in this project type.
	// Used by the CLI for validation and introspection.
	//
	// Map keys are phase names (e.g., "discovery", "implementation").
	Phases() map[string]phases.PhaseMetadata

	// Type returns the project type identifier (e.g., "standard", "design").
	// This matches the project.type field in the state schema.
	Type() string
}

// DetectProjectType returns the appropriate ProjectType implementation based on the state.
//
// This function examines the project.type field in the state and returns the corresponding
// ProjectType implementation. If the type is unknown or missing, it defaults to StandardProject.
//
// Example usage:
//
//	state := loadStateFromDisk()
//	projectType := types.DetectProjectType(state)
//	machine := projectType.BuildStateMachine()
func DetectProjectType(state *schemas.ProjectState) (ProjectType, error) {
	// Convert from schemas.ProjectState -> projects.ProjectState -> projects.StandardProjectState
	// These are distinct types (not aliases) due to how the CUE code generator works
	projectsState := (*projects.ProjectState)(state)
	standardState := (*projects.StandardProjectState)(projectsState)

	// State migration: default empty type to "standard" for backward compatibility
	// Existing projects created before the composable phases architecture won't have
	// a type field set, so we default them to "standard"
	if standardState.Project.Type == "" {
		standardState.Project.Type = "standard"
	}

	// For MVP, we only have StandardProject
	// Future: switch on state.Project.Type to support multiple types
	// Example:
	//   switch standardState.Project.Type {
	//   case "standard":
	//       return NewStandardProject(standardState), nil
	//   case "design":
	//       return NewDesignProject(designState), nil
	//   default:
	//       return nil, fmt.Errorf("unknown project type: %s", standardState.Project.Type)
	//   }

	return NewStandardProject(standardState), nil
}

// NewStandardProject creates a StandardProject type.
// This is exported for direct use when the type is known.
func NewStandardProject(state *projects.StandardProjectState) ProjectType {
	// Import here to avoid circular dependency
	// The actual implementation is in the standard package
	return newStandardProjectImpl(state)
}

// newStandardProjectImpl is implemented in the standard package.
// This indirection avoids circular imports while keeping the API clean.
// It's set by init() in the standard package.
var newStandardProjectImpl func(state *projects.StandardProjectState) ProjectType

// RegisterStandardProject registers the StandardProject implementation.
// This is called by init() in the standard package to avoid circular imports.
func RegisterStandardProject(fn func(state *projects.StandardProjectState) ProjectType) {
	newStandardProjectImpl = fn
}
