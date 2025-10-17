package statechart

// ProjectState is a minimal representation of project state for guard evaluation.
// This will be replaced with the actual project state schema from internal/schemas.
type ProjectState struct {
	Statechart Meta   `yaml:"_statechart"`
	Phases     Phases `yaml:"phases"`
}

// Meta tracks the current state machine position.
type Meta struct {
	CurrentState State `yaml:"current_state"`
}

// Phases represents the 5 phases and their state.
type Phases struct {
	Discovery      PhaseState
	Design         PhaseState
	Implementation ImplementationPhase
	Review         ReviewPhase
	Finalize       FinalizePhase
}

// PhaseState represents the common state for discovery and design phases.
type PhaseState struct {
	Enabled   bool
	Artifacts []Artifact
}

// Artifact represents a phase artifact (document, ADR, etc.)
type Artifact struct {
	Path     string
	Approved bool
}

// ImplementationPhase represents the implementation phase state.
type ImplementationPhase struct {
	Enabled bool
	Tasks   []Task
}

// Task represents an implementation task.
type Task struct {
	ID     string
	Name   string
	Status string // pending, in_progress, completed, abandoned
}

// ReviewPhase represents the review phase state.
type ReviewPhase struct {
	Enabled   bool
	Iteration int
	Reports   []ReviewReport
}

// ReviewReport represents a review report with assessment.
type ReviewReport struct {
	ID         string
	Assessment string // pass, fail
}

// FinalizePhase represents the finalize phase state.
type FinalizePhase struct {
	Enabled               bool
	DocumentationAssessed bool
	ChecksAssessed        bool
	ProjectDeleted        bool
}

// Guards for conditional transitions

// ArtifactsApproved checks if all artifacts for a phase are approved (or no artifacts exist).
func ArtifactsApproved(phase PhaseState) bool {
	if len(phase.Artifacts) == 0 {
		return true
	}
	for _, a := range phase.Artifacts {
		if !a.Approved {
			return false
		}
	}
	return true
}

// HasAtLeastOneTask checks if at least one task has been created.
func HasAtLeastOneTask(state *ProjectState) bool {
	return len(state.Phases.Implementation.Tasks) >= 1
}

// AllTasksComplete checks if all tasks are completed or abandoned.
func AllTasksComplete(state *ProjectState) bool {
	tasks := state.Phases.Implementation.Tasks
	if len(tasks) == 0 {
		return false
	}
	for _, t := range tasks {
		if t.Status != "completed" && t.Status != "abandoned" {
			return false
		}
	}
	return true
}

// DocumentationAssessed checks if documentation has been assessed and handled.
func DocumentationAssessed(state *ProjectState) bool {
	return state.Phases.Finalize.DocumentationAssessed
}

// ChecksAssessed checks if final checks have been assessed and handled.
func ChecksAssessed(state *ProjectState) bool {
	return state.Phases.Finalize.ChecksAssessed
}

// ProjectDeleted checks if the project folder has been deleted.
func ProjectDeleted(state *ProjectState) bool {
	return state.Phases.Finalize.ProjectDeleted
}
