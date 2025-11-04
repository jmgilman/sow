package state

import "fmt"

// Registry maps project type names to their configurations.
// This is a global registry that is populated during initialization.
// Full implementation will be provided in Unit 3 (SDK Builder).
var Registry = make(map[string]*ProjectTypeConfig)

// Register adds a project type configuration to the global registry.
// Panics if a project type with the same name is already registered.
// This prevents accidental duplicate registrations which could cause
// non-deterministic behavior.
//
// Typical usage in project type packages:
//
//	func init() {
//	    Register("standard", NewStandardProjectConfig())
//	}
func Register(typeName string, config *ProjectTypeConfig) {
	if _, exists := Registry[typeName]; exists {
		panic(fmt.Sprintf("project type already registered: %s", typeName))
	}
	Registry[typeName] = config
}

// Get retrieves a project type configuration from the registry.
// Returns (config, true) if found, (nil, false) if not found.
//
// Used by Load() to attach project type config to loaded project:
//
//	config, exists := Registry[project.Type]
//	if !exists {
//	    return fmt.Errorf("unknown project type: %s", project.Type)
//	}
func Get(typeName string) (*ProjectTypeConfig, bool) {
	config, exists := Registry[typeName]
	return config, exists
}

// State represents a state machine state.
// This is a placeholder type that will be replaced with the actual
// state machine implementation in future tasks.
type State string

// String returns the string representation of the state.
func (s State) String() string {
	return string(s)
}

// GetEventDeterminer returns the configured event determiner for the given state.
// Returns nil if no determiner is configured for the state.
//
// Event determiners are configured via builder's OnAdvance() method:
//
//	.OnAdvance(ReviewActive, func(p *Project) (Event, error) {
//	    // Examine state and determine event
//	})
func (ptc *ProjectTypeConfig) GetEventDeterminer(state State) EventDeterminer {
	return ptc.onAdvance[state]
}

// BuildMachine builds a state machine for the project.
// This is a stub that will be implemented in Unit 3.
func (ptc *ProjectTypeConfig) BuildMachine(project *Project, initialState State) *Machine {
	// Build transitions map for O(1) lookup
	transitions := make(map[State]map[Event]*TransitionConfig)
	for i := range ptc.transitions {
		tc := &ptc.transitions[i]
		if transitions[tc.From] == nil {
			transitions[tc.From] = make(map[Event]*TransitionConfig)
		}
		transitions[tc.From][tc.Event] = tc
	}

	return &Machine{
		currentState: initialState,
		transitions:  transitions,
		project:      project,
	}
}

// Validate validates project state against project type configuration.
//
// Performs two-tier validation:
//  1. Artifact type validation - Checks inputs/outputs against allowed types
//  2. Metadata validation - Validates metadata against embedded CUE schemas
//
// Returns error describing first validation failure found.
func (ptc *ProjectTypeConfig) Validate(project *Project) error {
	// Validate each phase
	for phaseName, phaseConfig := range ptc.phaseConfigs {
		phase, exists := project.Phases[phaseName]
		if !exists {
			// Phase not in state - skip (may be optional/future phase)
			continue
		}

		// Validate artifact types
		if err := validateArtifactTypes(
			phase.Inputs,
			phaseConfig.allowedInputTypes,
			phaseName,
			"input",
		); err != nil {
			return err
		}

		if err := validateArtifactTypes(
			phase.Outputs,
			phaseConfig.allowedOutputTypes,
			phaseName,
			"output",
		); err != nil {
			return err
		}

		// Validate metadata against embedded schema
		if phaseConfig.metadataSchema != "" {
			if err := validateMetadata(
				phase.Metadata,
				phaseConfig.metadataSchema,
			); err != nil {
				return fmt.Errorf("phase %s metadata: %w", phaseName, err)
			}
		} else if len(phase.Metadata) > 0 {
			return fmt.Errorf("phase %s does not support metadata", phaseName)
		}
	}

	return nil
}

// GetTaskSupportingPhases returns the names of all phases that support tasks.
// Returns an empty slice if no phases support tasks.
// Phase names are returned in sorted order for deterministic behavior.
func (ptc *ProjectTypeConfig) GetTaskSupportingPhases() []string {
	var phases []string
	for name, config := range ptc.phaseConfigs {
		if config.supportsTasks {
			phases = append(phases, name)
		}
	}
	// Sort for deterministic ordering
	if len(phases) > 1 {
		for i := 0; i < len(phases)-1; i++ {
			for j := i + 1; j < len(phases); j++ {
				if phases[i] > phases[j] {
					phases[i], phases[j] = phases[j], phases[i]
				}
			}
		}
	}
	return phases
}

// PhaseSupportsTasks checks if a specific phase supports tasks.
// Returns false if the phase doesn't exist or doesn't support tasks.
func (ptc *ProjectTypeConfig) PhaseSupportsTasks(phaseName string) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.supportsTasks
}

// GetDefaultTaskPhase returns the default phase for task operations based on current state.
// Returns empty string if no phase supports tasks or state mapping is ambiguous.
//
// Logic:
//  1. Check if current state maps to a phase's start or end state
//  2. If that phase supports tasks, return it
//  3. Otherwise return first task-supporting phase alphabetically
func (ptc *ProjectTypeConfig) GetDefaultTaskPhase(currentState State) string {
	// Try to map current state to a phase
	// Note: state.PhaseConfig doesn't have startState/endState fields yet
	// So for now, just return the first task-supporting phase
	// This will be improved when phase state mappings are added

	// Fallback: return first task-supporting phase
	phases := ptc.GetTaskSupportingPhases()
	if len(phases) > 0 {
		return phases[0]
	}
	return ""
}
