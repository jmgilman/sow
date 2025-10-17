// Package project provides business logic for project operations.
//
// This package handles project-specific concerns like initializing project state,
// validating branch names, and formatting project information for display.
package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
)

// Phase name constants.
const (
	PhaseDiscovery      = "discovery"
	PhaseDesign         = "design"
	PhaseImplementation = "implementation"
	PhaseReview         = "review"
	PhaseFinalize       = "finalize"
)

// validPhases maps phase names to their validity.
var validPhases = map[string]bool{
	PhaseDiscovery:      true,
	PhaseDesign:         true,
	PhaseImplementation: true,
	PhaseReview:         true,
	PhaseFinalize:       true,
}

// validDiscoveryTypes lists valid discovery type values.
var validDiscoveryTypes = map[string]bool{
	"bug":      true,
	"feature":  true,
	"docs":     true,
	"refactor": true,
	"general":  true,
}

// protectedBranches lists branch names that cannot have active projects.
// Projects must be on feature branches, never on main/master.
var protectedBranches = map[string]bool{
	"main":   true,
	"master": true,
}

// NewProjectState creates an initial ProjectState for a new project.
//
// All projects start with the same 5-phase structure:
//   - Discovery: disabled (enabled: false, status: skipped)
//   - Design: disabled (enabled: false, status: skipped)
//   - Implementation: enabled (enabled: true, status: pending)
//   - Review: enabled (enabled: true, status: pending, iteration: 1)
//   - Finalize: enabled (enabled: true, status: pending, project_deleted: false)
//
// The truth table in /project:new will later modify phase enablement based on
// user requirements. This creates the minimal valid starting state.
//
// Parameters:
//   - name: Project name (must be kebab-case, validated by CUE schema)
//   - branch: Git branch name (must not be main/master)
//   - description: Human-readable project description
//
// Returns:
//   - Fully initialized ProjectState ready to be written to disk
func NewProjectState(name, branch, description string) *schemas.ProjectState {
	now := time.Now()

	state := &schemas.ProjectState{}

	// Project metadata
	state.Project.Name = name
	state.Project.Branch = branch
	state.Project.Description = description
	state.Project.Created_at = now
	state.Project.Updated_at = now

	// Phase 1: Discovery (disabled by default)
	state.Phases.Discovery = schemas.DiscoveryPhase{
		Status:         "skipped",
		Created_at:     now,
		Started_at:     nil,
		Completed_at:   nil,
		Enabled:        false,
		Discovery_type: nil,
		Artifacts:      []schemas.Artifact{},
	}

	// Phase 2: Design (disabled by default)
	state.Phases.Design = schemas.DesignPhase{
		Status:         "skipped",
		Created_at:     now,
		Started_at:     nil,
		Completed_at:   nil,
		Enabled:        false,
		Architect_used: nil,
		Artifacts:      []schemas.Artifact{},
	}

	// Phase 3: Implementation (always enabled, required)
	state.Phases.Implementation = schemas.ImplementationPhase{
		Status:                  "pending",
		Created_at:              now,
		Started_at:              nil,
		Completed_at:            nil,
		Enabled:                 true,
		Planner_used:            nil,
		Tasks:                   []schemas.Task{},
		Pending_task_additions:  nil,
	}

	// Phase 4: Review (always enabled, required)
	state.Phases.Review = schemas.ReviewPhase{
		Status:       "pending",
		Created_at:   now,
		Started_at:   nil,
		Completed_at: nil,
		Enabled:      true,
		Iteration:    1,
		Reports:      []schemas.ReviewReport{},
	}

	// Phase 5: Finalize (always enabled, required)
	state.Phases.Finalize = schemas.FinalizePhase{
		Status:                "pending",
		Created_at:            now,
		Started_at:            nil,
		Completed_at:          nil,
		Enabled:               true,
		Documentation_updates: nil,
		Artifacts_moved:       nil,
		Project_deleted:       false,
		Pr_url:                nil,
	}

	return state
}

// ValidateBranch checks if the given branch name is allowed for projects.
//
// Projects cannot be created on protected branches (main, master) to ensure
// clean separation between feature work and the main codebase. The .sow/project/
// directory is committed to feature branches but must be deleted before merge.
//
// Parameters:
//   - branch: Branch name to validate
//
// Returns:
//   - nil if branch is valid
//   - error if branch is protected
func ValidateBranch(branch string) error {
	if protectedBranches[branch] {
		return fmt.Errorf("cannot create project on protected branch '%s' - use a feature branch instead", branch)
	}
	return nil
}

// FormatStatus generates a human-readable status summary for a project.
//
// Output format:
//   Project: {name} (on {branch})
//   Description: {description}
//
//   Phases:
//     [enabled/disabled] Phase    Status
//     [x] Discovery               skipped
//     [x] Design                  skipped
//     [✓] Implementation          in_progress
//     [✓] Review                  pending
//     [✓] Finalize                pending
//
//   Tasks: 3 total (1 completed, 1 in_progress, 1 pending)
//
// Parameters:
//   - state: Project state to format
//
// Returns:
//   - Formatted string ready for display
func FormatStatus(state *schemas.ProjectState) string {
	var b strings.Builder

	// Project header
	fmt.Fprintf(&b, "Project: %s (on %s)\n", state.Project.Name, state.Project.Branch)
	fmt.Fprintf(&b, "Description: %s\n\n", state.Project.Description)

	// Phases table
	fmt.Fprintln(&b, "Phases:")
	fmt.Fprintln(&b, "  [enabled] Phase              Status")

	phases := []struct {
		name    string
		enabled bool
		status  string
	}{
		{"Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Status},
		{"Design", state.Phases.Design.Enabled, state.Phases.Design.Status},
		{"Implementation", state.Phases.Implementation.Enabled, state.Phases.Implementation.Status},
		{"Review", state.Phases.Review.Enabled, state.Phases.Review.Status},
		{"Finalize", state.Phases.Finalize.Enabled, state.Phases.Finalize.Status},
	}

	for _, p := range phases {
		enabledMark := " "
		if p.enabled {
			enabledMark = "✓"
		}
		// Pad phase name to 20 characters for alignment
		fmt.Fprintf(&b, "  [%s] %-20s %s\n", enabledMark, p.name, p.status)
	}

	// Task summary
	formatTaskSummary(&b, state.Phases.Implementation.Tasks)

	return b.String()
}

// formatTaskSummary formats the task summary section.
func formatTaskSummary(b *strings.Builder, tasks []schemas.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(b, "\nTasks: none yet")
		return
	}

	total := len(tasks)
	completed := 0
	inProgress := 0
	pending := 0
	abandoned := 0

	for _, task := range tasks {
		switch task.Status {
		case "completed":
			completed++
		case "in_progress":
			inProgress++
		case "pending":
			pending++
		case "abandoned":
			abandoned++
		}
	}

	fmt.Fprintf(b, "\nTasks: %d total", total)
	details := []string{}
	if completed > 0 {
		details = append(details, fmt.Sprintf("%d completed", completed))
	}
	if inProgress > 0 {
		details = append(details, fmt.Sprintf("%d in_progress", inProgress))
	}
	if pending > 0 {
		details = append(details, fmt.Sprintf("%d pending", pending))
	}
	if abandoned > 0 {
		details = append(details, fmt.Sprintf("%d abandoned", abandoned))
	}
	if len(details) > 0 {
		fmt.Fprintf(b, " (%s)", strings.Join(details, ", "))
	}
	fmt.Fprintln(b)
}

// ValidatePhase checks if the given phase name is valid.
//
// Parameters:
//   - phase: Phase name to validate
//
// Returns:
//   - nil if phase is valid
//   - error if phase name is invalid
func ValidatePhase(phase string) error {
	if !validPhases[phase] {
		return fmt.Errorf("invalid phase '%s': must be one of discovery, design, implementation, review, finalize", phase)
	}
	return nil
}

// EnablePhase enables a phase in the project state.
//
// Only discovery and design phases can be enabled (impl/review/finalize are always enabled).
// When enabling discovery, a discovery_type must be provided.
//
// Parameters:
//   - state: Project state to modify
//   - phase: Phase name to enable
//   - discoveryType: Discovery type (required only if phase is discovery)
//
// Returns:
//   - nil on success
//   - error if phase can't be enabled or validation fails
func EnablePhase(state *schemas.ProjectState, phase string, discoveryType *string) error {
	now := time.Now()

	switch phase {
	case PhaseDiscovery:
		// Check if already enabled
		if state.Phases.Discovery.Enabled {
			return fmt.Errorf("discovery phase is already enabled")
		}

		// Validate discovery type is provided
		if discoveryType == nil || *discoveryType == "" {
			return fmt.Errorf("discovery_type is required when enabling discovery phase")
		}
		if !validDiscoveryTypes[*discoveryType] {
			return fmt.Errorf("invalid discovery_type '%s': must be one of bug, feature, docs, refactor, general", *discoveryType)
		}

		// Enable discovery phase
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "pending"
		state.Phases.Discovery.Discovery_type = discoveryType
		state.Project.Updated_at = now

	case PhaseDesign:
		// Check if already enabled
		if state.Phases.Design.Enabled {
			return fmt.Errorf("design phase is already enabled")
		}

		// Enable design phase
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "pending"
		state.Project.Updated_at = now

	case PhaseImplementation, PhaseReview, PhaseFinalize:
		return fmt.Errorf("phase '%s' is always enabled and cannot be manually enabled", phase)

	default:
		return fmt.Errorf("invalid phase '%s'", phase)
	}

	return nil
}

// CompletePhase marks a phase as completed.
//
// Validates that the phase can be completed (all requirements met) before
// updating the state.
//
// Parameters:
//   - state: Project state to modify
//   - phase: Phase name to complete
//
// Returns:
//   - nil on success
//   - error if phase can't be completed or validation fails
func CompletePhase(state *schemas.ProjectState, phase string) error {
	// First validate the phase can be completed
	if err := ValidatePhaseCompletion(state, phase); err != nil {
		return err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// Update the appropriate phase
	switch phase {
	case PhaseDiscovery:
		state.Phases.Discovery.Status = "completed"
		state.Phases.Discovery.Completed_at = nowStr
		if state.Phases.Discovery.Started_at == nil {
			state.Phases.Discovery.Started_at = nowStr
		}

	case PhaseDesign:
		state.Phases.Design.Status = "completed"
		state.Phases.Design.Completed_at = nowStr
		if state.Phases.Design.Started_at == nil {
			state.Phases.Design.Started_at = nowStr
		}

	case PhaseImplementation:
		state.Phases.Implementation.Status = "completed"
		state.Phases.Implementation.Completed_at = nowStr
		if state.Phases.Implementation.Started_at == nil {
			state.Phases.Implementation.Started_at = nowStr
		}

	case PhaseReview:
		state.Phases.Review.Status = "completed"
		state.Phases.Review.Completed_at = nowStr
		if state.Phases.Review.Started_at == nil {
			state.Phases.Review.Started_at = nowStr
		}

	case PhaseFinalize:
		state.Phases.Finalize.Status = "completed"
		state.Phases.Finalize.Completed_at = nowStr
		if state.Phases.Finalize.Started_at == nil {
			state.Phases.Finalize.Started_at = nowStr
		}

	default:
		return fmt.Errorf("invalid phase '%s'", phase)
	}

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// ValidatePhaseCompletion checks if a phase meets all requirements for completion.
//
// Each phase has different completion requirements:
//   - Discovery: All artifacts must be approved
//   - Design: All artifacts must be approved
//   - Implementation: All tasks must be completed or abandoned
//   - Review: Latest report must have assessment "pass"
//   - Finalize: project_deleted must be true
//
// Parameters:
//   - state: Project state to validate
//   - phase: Phase name to check
//
// Returns:
//   - nil if phase can be completed
//   - error describing what requirements are not met
func ValidatePhaseCompletion(state *schemas.ProjectState, phase string) error {
	switch phase {
	case PhaseDiscovery:
		if !state.Phases.Discovery.Enabled {
			return fmt.Errorf("discovery phase is not enabled")
		}
		if state.Phases.Discovery.Status == "completed" {
			return fmt.Errorf("discovery phase is already completed")
		}

		// Check all artifacts are approved
		for _, artifact := range state.Phases.Discovery.Artifacts {
			if !artifact.Approved {
				return fmt.Errorf("artifact '%s' is not approved", artifact.Path)
			}
		}

	case PhaseDesign:
		if !state.Phases.Design.Enabled {
			return fmt.Errorf("design phase is not enabled")
		}
		if state.Phases.Design.Status == "completed" {
			return fmt.Errorf("design phase is already completed")
		}

		// Check all artifacts are approved
		for _, artifact := range state.Phases.Design.Artifacts {
			if !artifact.Approved {
				return fmt.Errorf("artifact '%s' is not approved", artifact.Path)
			}
		}

	case PhaseImplementation:
		if state.Phases.Implementation.Status == "completed" {
			return fmt.Errorf("implementation phase is already completed")
		}

		// Check all tasks are completed or abandoned
		for _, task := range state.Phases.Implementation.Tasks {
			if task.Status != "completed" && task.Status != "abandoned" {
				return fmt.Errorf("task '%s' (%s) is not completed or abandoned (status: %s)", task.Id, task.Name, task.Status)
			}
		}

	case PhaseReview:
		if state.Phases.Review.Status == "completed" {
			return fmt.Errorf("review phase is already completed")
		}

		// Check latest report has pass assessment
		if len(state.Phases.Review.Reports) == 0 {
			return fmt.Errorf("no review reports exist - at least one review report is required")
		}

		latestReport := state.Phases.Review.Reports[len(state.Phases.Review.Reports)-1]
		if latestReport.Assessment != "pass" {
			return fmt.Errorf("latest review report assessment is '%s' (must be 'pass' to complete)", latestReport.Assessment)
		}

	case PhaseFinalize:
		if state.Phases.Finalize.Status == "completed" {
			return fmt.Errorf("finalize phase is already completed")
		}

		// Check project_deleted is true
		if !state.Phases.Finalize.Project_deleted {
			return fmt.Errorf("project must be deleted before completing finalize phase (project_deleted must be true)")
		}

	default:
		return fmt.Errorf("invalid phase '%s'", phase)
	}

	return nil
}

// FormatPhaseStatus generates a human-readable phase status table.
//
// This is similar to FormatStatus but focuses only on phases without
// the project metadata and task summary.
//
// Output format:
//   Phases:
//     [enabled] Phase              Status
//     [ ] Discovery                skipped
//     [ ] Design                   skipped
//     [✓] Implementation           in_progress
//     [✓] Review                   pending
//     [✓] Finalize                 pending
//
// Parameters:
//   - state: Project state to format
//
// Returns:
//   - Formatted string ready for display
func FormatPhaseStatus(state *schemas.ProjectState) string {
	var b strings.Builder

	// Phases table
	fmt.Fprintln(&b, "Phases:")
	fmt.Fprintln(&b, "  [enabled] Phase              Status")

	phases := []struct {
		name    string
		enabled bool
		status  string
	}{
		{"Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Status},
		{"Design", state.Phases.Design.Enabled, state.Phases.Design.Status},
		{"Implementation", state.Phases.Implementation.Enabled, state.Phases.Implementation.Status},
		{"Review", state.Phases.Review.Enabled, state.Phases.Review.Status},
		{"Finalize", state.Phases.Finalize.Enabled, state.Phases.Finalize.Status},
	}

	for _, p := range phases {
		enabledMark := " "
		if p.enabled {
			enabledMark = "✓"
		}
		// Pad phase name to 20 characters for alignment
		fmt.Fprintf(&b, "  [%s] %-20s %s\n", enabledMark, p.name, p.status)
	}

	return b.String()
}
