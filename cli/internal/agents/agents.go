// Package agents provides agent definitions for the sow multi-agent system.
//
// This package defines the Agent struct and provides standard agent definitions
// for common roles in software development: Implementer, Architect, Reviewer,
// Planner, Researcher, and Decomposer.
//
// Agents are lightweight configuration (data), not behavior. They are simple
// structs representing roles that can be used by the executor system to spawn
// workers with appropriate context and capabilities.
//
// Example usage:
//
//	// Access a standard agent
//	agent := agents.Implementer
//	fmt.Printf("Agent: %s\n", agent.Name)
//	fmt.Printf("Description: %s\n", agent.Description)
//
//	// Get all standard agents
//	for _, agent := range agents.StandardAgents() {
//	    fmt.Printf("%s: %s\n", agent.Name, agent.Description)
//	}
package agents

// Agent represents a role in the sow system.
//
// Agents are lightweight configuration structs that define what a particular
// role does and what capabilities it needs. They are used by the executor
// system to spawn workers with appropriate context.
//
// Agents do not contain behavior - they are pure data. The executor is
// responsible for interpreting the agent configuration and spawning the
// appropriate worker with the right tools and prompt.
type Agent struct {
	// Name is the agent identifier (e.g., "implementer", "architect").
	// This should be lowercase and match the prompt template filename.
	Name string

	// Description explains what this agent does.
	// This is a short, human-readable summary of the agent's purpose.
	Description string

	// Capabilities describes what the agent must be able to do (prose).
	// This documents the required tools and permissions for the agent.
	Capabilities string

	// PromptPath is the path to the embedded prompt template.
	// Relative to the templates/ directory.
	PromptPath string
}

// Standard agent definitions for the sow multi-agent system.
// These are the core roles used in software development workflows.
var (
	// Implementer is the agent responsible for code implementation using TDD.
	Implementer = &Agent{
		Name:         "implementer",
		Description:  "Code implementation using Test-Driven Development",
		Capabilities: "Must be able to read/write files, execute shell commands, search codebase",
		PromptPath:   "implementer.md",
	}

	// Architect is the agent responsible for system design and architecture decisions.
	Architect = &Agent{
		Name:         "architect",
		Description:  "System design and architecture decisions",
		Capabilities: "Must be able to read/write files, search codebase",
		PromptPath:   "architect.md",
	}

	// Reviewer is the agent responsible for code review and quality assessment.
	Reviewer = &Agent{
		Name:         "reviewer",
		Description:  "Code review and quality assessment",
		Capabilities: "Must be able to read files, search codebase, execute shell commands",
		PromptPath:   "reviewer.md",
	}

	// Planner is the agent responsible for researching codebase and creating task breakdowns.
	Planner = &Agent{
		Name:         "planner",
		Description:  "Research codebase and create comprehensive implementation task breakdown",
		Capabilities: "Must be able to read files, search codebase, write task descriptions",
		PromptPath:   "planner.md",
	}

	// Researcher is the agent responsible for focused, impartial research.
	Researcher = &Agent{
		Name:         "researcher",
		Description:  "Focused, impartial research with comprehensive source investigation and citation",
		Capabilities: "Must be able to read files, search codebase, access web resources",
		PromptPath:   "researcher.md",
	}

	// Decomposer is the agent responsible for decomposing complex features into work units.
	Decomposer = &Agent{
		Name:         "decomposer",
		Description:  "Specialized for decomposing complex features into project-sized, implementable work units",
		Capabilities: "Must be able to read files, search codebase, write specifications",
		PromptPath:   "decomposer.md",
	}
)

// StandardAgents returns all standard agents defined in this package.
//
// The returned slice contains pointers to the package-level agent definitions:
// Implementer, Architect, Reviewer, Planner, Researcher, and Decomposer.
//
// This function is useful for iterating over all available agents or for
// building registries and lookups.
func StandardAgents() []*Agent {
	return []*Agent{
		Implementer,
		Architect,
		Reviewer,
		Planner,
		Researcher,
		Decomposer,
	}
}
