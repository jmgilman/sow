package standard

import (
	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/internal/phases/design"
	"github.com/jmgilman/sow/cli/internal/phases/discovery"
	"github.com/jmgilman/sow/cli/internal/phases/finalize"
	"github.com/jmgilman/sow/cli/internal/phases/implementation"
	"github.com/jmgilman/sow/cli/internal/phases/review"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/project/types"
	"github.com/jmgilman/sow/cli/schemas/projects"
	"github.com/qmuntal/stateless"
)

// init registers the StandardProject implementation with the types package.
// This avoids circular import issues.
func init() {
	types.RegisterStandardProject(func(state *projects.StandardProjectState) types.ProjectType {
		return New(state)
	})
}

// StandardProject implements the standard 5-phase project lifecycle.
//
// Phase sequence:
//  1. Discovery (optional) - Problem understanding and context gathering
//  2. Design (optional) - Architecture and approach design
//  3. Implementation (required) - Code implementation with tasks
//  4. Review (required) - Code review with possible iteration
//  5. Finalize (required) - Documentation, checks, and cleanup
//
// Special transition: Review can fail and loop back to Implementation.
type StandardProject struct {
	state *projects.StandardProjectState
}

// New creates a new StandardProject with the given state.
func New(state *projects.StandardProjectState) *StandardProject {
	return &StandardProject{state: state}
}

// Type returns the project type identifier.
func (p *StandardProject) Type() string {
	return "standard"
}

// BuildStateMachine constructs and configures the state machine for a standard project.
//
// Architecture:
//  1. Creates all 5 phases with their respective state data
//  2. Uses BuildPhaseChain to wire forward transitions automatically
//  3. Adds exceptional backward transition: Review → Implementation
//  4. Returns fully configured machine with prompts and guards
func (p *StandardProject) BuildStateMachine() *statechart.Machine {
	// Get current state from the project
	currentState := statechart.State(p.state.Statechart.Current_state)

	// Create underlying stateless machine
	sm := stateless.NewStateMachine(currentState)

	// Get project info for phases to use in templates
	projectInfo := phases.ProjectInfo{
		Name:        p.state.Project.Name,
		Description: p.state.Project.Description,
		Branch:      p.state.Project.Branch,
	}

	// Instantiate all 5 phases with their data from state
	phaseList := []phases.Phase{
		discovery.New(true, &p.state.Phases.Discovery, projectInfo),      // Optional
		design.New(true, &p.state.Phases.Design, projectInfo),            // Optional
		implementation.New(&p.state.Phases.Implementation, projectInfo),  // Required
		review.New(&p.state.Phases.Review, projectInfo),                  // Required
		finalize.New(&p.state.Phases.Finalize, projectInfo),              // Required
	}

	// Build forward chain: NoProject → Discovery → ... → Finalize → NoProject
	phaseMap := phases.BuildPhaseChain(sm, phaseList)

	// Add exceptional backward transition: Review fail → Implementation
	// This allows iterating on implementation based on review feedback
	implPhase := phaseMap["implementation"]
	reviewPhase := phaseMap["review"].(*review.ReviewPhase)

	sm.Configure(statechart.ReviewActive).
		Permit(statechart.EventReviewFail, implPhase.EntryState(), reviewPhase.LatestReviewFailedGuard)

	// Convert StandardProjectState to ProjectState (they're type aliases of each other)
	projectState := (*projects.ProjectState)(p.state)

	// Wrap in our Machine type that includes project state and persistence
	machine := statechart.NewMachineFromPhases(sm, projectState)

	return machine
}

// Phases returns metadata for all phases in the standard project lifecycle.
//
// This metadata is used by the CLI to:
//   - Validate commands (e.g., "does this phase support tasks?")
//   - Provide introspection (e.g., "what custom fields does this phase have?")
//   - Generate help text
func (p *StandardProject) Phases() map[string]phases.PhaseMetadata {
	// Create temporary phase instances just to get metadata
	// We pass nil for state data since metadata doesn't depend on state
	return map[string]phases.PhaseMetadata{
		"discovery":      discovery.New(true, nil, phases.ProjectInfo{}).Metadata(),
		"design":         design.New(true, nil, phases.ProjectInfo{}).Metadata(),
		"implementation": implementation.New(nil, phases.ProjectInfo{}).Metadata(),
		"review":         review.New(nil, phases.ProjectInfo{}).Metadata(),
		"finalize":       finalize.New(nil, phases.ProjectInfo{}).Metadata(),
	}
}
