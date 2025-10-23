package schemas

import (
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// Re-export project types for backward compatibility
// This ensures existing code importing schemas.ProjectState continues to work
#ProjectState: projects.#ProjectState

#StandardProjectState: projects.#StandardProjectState

// Re-export common types
#Phase: phases.#Phase

#Artifact: phases.#Artifact

#Task: phases.#Task
