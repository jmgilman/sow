package project

import (
	"fmt"
	"sort"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TransitionInfo describes a single available transition from a state.
// Used by introspection methods to provide structured information about
// state machine configuration without exposing internal implementation details.
type TransitionInfo struct {
	Event       sdkstate.Event // Event that triggers this transition
	From        sdkstate.State // Source state
	To          sdkstate.State // Target state
	Description string         // Human-readable description (empty if not provided)
	GuardDesc   string         // Guard description (empty if no guard)
}

// PhaseConfig holds configuration for a single phase in a project type.
type PhaseConfig struct {
	// name is the phase identifier
	name string

	// startState is the state when the phase begins
	startState sdkstate.State

	// endState is the state when the phase ends
	endState sdkstate.State

	// allowedInputTypes are the artifact types allowed as inputs
	// Empty slice means all types are allowed
	allowedInputTypes []string

	// allowedOutputTypes are the artifact types allowed as outputs
	// Empty slice means all types are allowed
	allowedOutputTypes []string

	// supportsTasks indicates whether the phase can have tasks
	supportsTasks bool

	// metadataSchema is an embedded CUE schema for metadata validation
	metadataSchema string
}

// TransitionConfig holds configuration for a state machine transition.
type TransitionConfig struct {
	// From is the source state
	From sdkstate.State

	// To is the target state
	To sdkstate.State

	// Event is the event that triggers the transition
	Event sdkstate.Event

	// guardTemplate is a function template that becomes a bound guard
	guardTemplate GuardTemplate

	// onEntry is an action to execute when entering the target state
	onEntry Action

	// onExit is an action to execute when exiting the source state
	onExit Action

	// failedPhase optionally specifies a phase to mark as "failed" instead of "completed"
	// when exiting its end state on this transition. Used for error/failure paths.
	failedPhase string

	// description is a human-readable explanation of what this transition does.
	// Context-specific: same event from different states can have different meanings.
	description string
}

// ProjectTypeConfig holds the complete configuration for a project type.
type ProjectTypeConfig struct {
	// name is the project type identifier
	name string

	// phaseConfigs are the phase configurations indexed by phase name
	phaseConfigs map[string]*PhaseConfig

	// initialState is the starting state of the state machine
	initialState sdkstate.State

	// transitions are all state transitions for this project type
	transitions []TransitionConfig

	// onAdvance are event determiners mapped by state
	// These determine which event to use for the Advance() command
	onAdvance map[sdkstate.State]EventDeterminer

	// prompts are prompt generators mapped by state
	// These generate contextual prompts for users in each state
	prompts map[sdkstate.State]PromptGenerator

	// orchestratorPrompt generates project-type-specific orchestrator guidance
	// This explains how the project type works and how orchestrator coordinates work
	orchestratorPrompt PromptGenerator

	// initializer is called during Create() to initialize the project
	// with phases, metadata, and any type-specific initial state
	initializer state.Initializer

	// branches are branch configurations mapped by state
	// Stored for introspection and debugging
	branches map[sdkstate.State]*BranchConfig
}

// InitialState returns the configured initial state for this project type.
func (ptc *ProjectTypeConfig) InitialState() sdkstate.State {
	return ptc.initialState
}

// Initialize calls the configured initializer function if present.
// Returns nil if no initializer is configured.
func (ptc *ProjectTypeConfig) Initialize(project *state.Project, initialInputs map[string][]projschema.ArtifactState) error {
	if ptc.initializer == nil {
		return nil
	}
	return ptc.initializer(project, initialInputs)
}

// OrchestratorPrompt returns the orchestrator prompt for this project type.
// This explains how the project type works and how the orchestrator should coordinate work.
// Returns empty string if no orchestrator prompt is configured.
func (ptc *ProjectTypeConfig) OrchestratorPrompt(project *state.Project) string {
	if ptc.orchestratorPrompt == nil {
		return ""
	}
	return ptc.orchestratorPrompt(project)
}

// GetStatePrompt returns the prompt for a specific state.
// Returns empty string if no prompt is configured for the state.
func (ptc *ProjectTypeConfig) GetStatePrompt(state sdkstate.State, project *state.Project) string {
	gen, exists := ptc.prompts[state]
	if !exists {
		return ""
	}
	return gen(project)
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
	sort.Strings(phases)
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
func (ptc *ProjectTypeConfig) GetDefaultTaskPhase(currentState sdkstate.State) string {
	// Try to map current state to a phase
	for name, config := range ptc.phaseConfigs {
		if (config.startState == currentState || config.endState == currentState) && config.supportsTasks {
			return name
		}
	}

	// Fallback: return first task-supporting phase
	phases := ptc.GetTaskSupportingPhases()
	if len(phases) > 0 {
		return phases[0]
	}
	return ""
}

// Validate validates project state against project type configuration.
//
// Performs two-tier validation:
//  1. Artifact type validation - Checks inputs/outputs against allowed types
//  2. Metadata validation - Validates metadata against embedded CUE schemas
//
// Returns error describing first validation failure found.
func (ptc *ProjectTypeConfig) Validate(project *state.Project) error {
	// Validate each phase
	for phaseName, phaseConfig := range ptc.phaseConfigs {
		phase, exists := project.Phases[phaseName]
		if !exists {
			// Phase not in state - skip (may be optional/future phase)
			continue
		}

		// Validate artifact types using state package helpers
		if err := state.ValidateArtifactTypes(
			phase.Inputs,
			phaseConfig.allowedInputTypes,
			phaseName,
			"input",
		); err != nil {
			return fmt.Errorf("validating inputs: %w", err)
		}

		if err := state.ValidateArtifactTypes(
			phase.Outputs,
			phaseConfig.allowedOutputTypes,
			phaseName,
			"output",
		); err != nil {
			return fmt.Errorf("validating outputs: %w", err)
		}

		// Validate metadata against embedded schema (if schema provided)
		// Phases without schemas can have arbitrary metadata
		if phaseConfig.metadataSchema != "" {
			if err := state.ValidateMetadata(
				phase.Metadata,
				phaseConfig.metadataSchema,
			); err != nil {
				return fmt.Errorf("phase %s metadata: %w", phaseName, err)
			}
		}
	}

	return nil
}

// DetermineEvent determines which event to fire from the current state.
// Returns the event to fire, or an error if no determiner is configured.
func (ptc *ProjectTypeConfig) DetermineEvent(project *state.Project) (sdkstate.Event, error) {
	currentState := sdkstate.State(project.Statechart.Current_state)
	determiner, exists := ptc.onAdvance[currentState]
	if !exists {
		return "", fmt.Errorf("no event determiner configured for state %s", currentState)
	}
	return determiner(project)
}

// GetPhaseForState returns the phase name that contains the given state.
// Returns empty string if the state doesn't belong to any phase's start or end state.
// If multiple phases have the same state (which shouldn't happen in a well-designed
// project type), returns the first match in iteration order.
func (ptc *ProjectTypeConfig) GetPhaseForState(state sdkstate.State) string {
	for name, config := range ptc.phaseConfigs {
		if config.startState == state || config.endState == state {
			return name
		}
	}
	return ""
}

// IsPhaseStartState checks if the given state is the startState of the specified phase.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseStartState(phaseName string, state sdkstate.State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.startState == state
}

// IsPhaseEndState checks if the given state is the endState of the specified phase.
// Returns false if the phase doesn't exist or the state doesn't match.
func (ptc *ProjectTypeConfig) IsPhaseEndState(phaseName string, state sdkstate.State) bool {
	config, exists := ptc.phaseConfigs[phaseName]
	if !exists {
		return false
	}
	return config.endState == state
}

// GetTransition looks up a transition config by from state, event, and to state.
// Returns nil if no matching transition is found.
func (ptc *ProjectTypeConfig) GetTransition(from, to sdkstate.State, event sdkstate.Event) *TransitionConfig {
	for i := range ptc.transitions {
		tc := &ptc.transitions[i]
		if tc.From == from && tc.To == to && tc.Event == event {
			return tc
		}
	}
	return nil
}

// GetAvailableTransitions returns all configured transitions from a state.
//
// This returns the transitions defined in the project type configuration,
// not filtered by guards. To check if a transition is currently allowed,
// use machine.CanFire(event) or machine.PermittedTriggers().
//
// Transitions are returned in a deterministic order:
//   1. First, transitions from branches (if state is a branching state)
//   2. Then, direct AddTransition calls
//   Both sorted by event name for consistency
//
// Returns empty slice if no transitions are defined from the state.
//
// Example:
//   transitions := config.GetAvailableTransitions(sdkstate.State(ReviewActive))
//   for _, t := range transitions {
//       fmt.Printf("%s -> %s: %s\n", t.Event, t.To, t.Description)
//   }
//
func (ptc *ProjectTypeConfig) GetAvailableTransitions(from sdkstate.State) []TransitionInfo {
	var result []TransitionInfo

	// Check if this is a branching state
	branchConfig, isBranching := ptc.branches[from]
	if isBranching {
		// Add transitions from branch paths
		for _, path := range branchConfig.branches {
			result = append(result, TransitionInfo{
				Event:       path.event,
				From:        from,
				To:          path.to,
				Description: path.description,
				GuardDesc:   path.guardTemplate.Description,
			})
		}
	}

	// Add direct transitions (from AddTransition calls)
	// Skip transitions that are part of a branch configuration (they're already included above)
	for _, tc := range ptc.transitions {
		//nolint:nestif // Deduplication logic requires nesting
		if tc.From == from {
			// Skip if this transition is part of a branch
			if isBranching {
				// Check if this transition matches any branch path
				isBranchTransition := false
				for _, path := range branchConfig.branches {
					if tc.Event == path.event && tc.To == path.to {
						isBranchTransition = true
						break
					}
				}
				if isBranchTransition {
					continue
				}
			}

			result = append(result, TransitionInfo{
				Event:       tc.Event,
				From:        tc.From,
				To:          tc.To,
				Description: tc.description,
				GuardDesc:   tc.guardTemplate.Description,
			})
		}
	}

	// Sort by event name for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Event < result[j].Event
	})

	return result
}

// GetTransitionDescription returns the human-readable description for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the first matching description found.
//
// Returns empty string if:
//   - No transition exists for the from/event combination
//   - The transition exists but has no description
//
// Note: This searches by from-state and event, not by to-state, since that's
// how transitions are triggered. The same event from different states can have
// different descriptions (context-specific).
//
// Example:
//   desc := config.GetTransitionDescription(
//       sdkstate.State(ReviewActive),
//       sdkstate.Event(EventReviewPass))
//   // Returns: "Review approved - proceed to finalization"
//
func (ptc *ProjectTypeConfig) GetTransitionDescription(from sdkstate.State, event sdkstate.Event) string {
	// Check branch paths first
	if branchConfig, exists := ptc.branches[from]; exists {
		for _, path := range branchConfig.branches {
			if path.event == event {
				return path.description
			}
		}
	}

	// Check direct transitions
	for _, tc := range ptc.transitions {
		if tc.From == from && tc.Event == event {
			return tc.description
		}
	}

	return ""
}

// GetTargetState returns the target state for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the target state of the first matching transition.
//
// Returns empty State if no transition exists for the from/event combination.
//
// Example:
//   target := config.GetTargetState(
//       sdkstate.State(ReviewActive),
//       sdkstate.Event(EventReviewPass))
//   // Returns: sdkstate.State(FinalizeChecks)
//
func (ptc *ProjectTypeConfig) GetTargetState(from sdkstate.State, event sdkstate.Event) sdkstate.State {
	// Check branch paths first
	if branchConfig, exists := ptc.branches[from]; exists {
		for _, path := range branchConfig.branches {
			if path.event == event {
				return path.to
			}
		}
	}

	// Check direct transitions
	for _, tc := range ptc.transitions {
		if tc.From == from && tc.Event == event {
			return tc.To
		}
	}

	return ""
}

// GetGuardDescription returns the guard description for a transition.
//
// Searches both branch paths and direct transitions for the specified from-state
// and event combination. Returns the guard description if a guard exists.
//
// Returns empty string if:
//   - No transition exists for the from/event combination
//   - The transition exists but has no guard
//   - The guard exists but has no description
//
// Example:
//   desc := config.GetGuardDescription(
//       sdkstate.State(ImplementationExecuting),
//       sdkstate.Event(EventAllTasksComplete))
//   // Returns: "all tasks complete"
//
func (ptc *ProjectTypeConfig) GetGuardDescription(from sdkstate.State, event sdkstate.Event) string {
	// Check branch paths first
	if branchConfig, exists := ptc.branches[from]; exists {
		for _, path := range branchConfig.branches {
			if path.event == event {
				return path.guardTemplate.Description
			}
		}
	}

	// Check direct transitions
	for _, tc := range ptc.transitions {
		if tc.From == from && tc.Event == event {
			return tc.guardTemplate.Description
		}
	}

	return ""
}

// IsBranchingState checks if a state has branches configured via AddBranch.
//
// Returns true if the state was configured with AddBranch (state-determined branching).
// Returns false if the state:
//   - Has no transitions
//   - Has only direct transitions (via AddTransition)
//   - Has multiple transitions but no AddBranch configuration
//
// This distinction is useful for UI/CLI to:
//   - Show different help text for branching vs non-branching states
//   - Indicate that transition choice is automatic vs manual
//   - Highlight states where discriminator logic determines the path
//
// Example:
//   if config.IsBranchingState(sdkstate.State(ReviewActive)) {
//       fmt.Println("This is a branching state - the system will automatically")
//       fmt.Println("determine which transition to take based on project state")
//   }
//
func (ptc *ProjectTypeConfig) IsBranchingState(state sdkstate.State) bool {
	_, exists := ptc.branches[state]
	return exists
}
