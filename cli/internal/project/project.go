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

// ============================================================================
// Artifact Management
// ============================================================================

// AddArtifact adds an artifact to a phase (discovery or design).
//
// Parameters:
//   - state: Project state to modify
//   - phase: Phase name (discovery or design)
//   - path: Artifact path (relative to .sow/project/)
//   - approved: Whether artifact is pre-approved
//
// Returns:
//   - nil on success
//   - error if validation fails or artifact already exists
func AddArtifact(state *schemas.ProjectState, phase, path string, approved bool) error {
	now := time.Now()

	// Validate phase
	if phase != PhaseDiscovery && phase != PhaseDesign {
		return fmt.Errorf("artifacts can only be added to discovery or design phases, got: %s", phase)
	}

	// Get the appropriate artifact list
	var artifacts *[]schemas.Artifact
	var enabled bool

	if phase == PhaseDiscovery {
		artifacts = &state.Phases.Discovery.Artifacts
		enabled = state.Phases.Discovery.Enabled
	} else {
		artifacts = &state.Phases.Design.Artifacts
		enabled = state.Phases.Design.Enabled
	}

	// Check phase is enabled
	if !enabled {
		return fmt.Errorf("%s phase is not enabled", phase)
	}

	// Check if artifact already exists
	for _, artifact := range *artifacts {
		if artifact.Path == path {
			return fmt.Errorf("artifact '%s' already exists in %s phase", path, phase)
		}
	}

	// Create new artifact
	newArtifact := schemas.Artifact{
		Path:       path,
		Approved:   approved,
		Created_at: now,
	}

	// Append to list
	*artifacts = append(*artifacts, newArtifact)

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// ApproveArtifact marks an artifact as approved in a phase.
//
// Parameters:
//   - state: Project state to modify
//   - phase: Phase name (discovery or design)
//   - path: Artifact path to approve
//
// Returns:
//   - nil on success
//   - error if artifact not found or already approved
func ApproveArtifact(state *schemas.ProjectState, phase, path string) error {
	now := time.Now()

	// Validate phase
	if phase != PhaseDiscovery && phase != PhaseDesign {
		return fmt.Errorf("artifacts can only exist in discovery or design phases, got: %s", phase)
	}

	// Get the appropriate artifact list
	var artifacts *[]schemas.Artifact
	var enabled bool

	if phase == PhaseDiscovery {
		artifacts = &state.Phases.Discovery.Artifacts
		enabled = state.Phases.Discovery.Enabled
	} else {
		artifacts = &state.Phases.Design.Artifacts
		enabled = state.Phases.Design.Enabled
	}

	// Check phase is enabled
	if !enabled {
		return fmt.Errorf("%s phase is not enabled", phase)
	}

	// Find and approve the artifact
	found := false
	for i := range *artifacts {
		if (*artifacts)[i].Path == path {
			if (*artifacts)[i].Approved {
				return fmt.Errorf("artifact '%s' is already approved", path)
			}
			(*artifacts)[i].Approved = true
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("artifact '%s' not found in %s phase", path, phase)
	}

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// FormatArtifactList generates a human-readable artifact list.
//
// Parameters:
//   - state: Project state to format
//   - phase: Phase to show artifacts for (empty string = all phases)
//
// Returns:
//   - Formatted string ready for display
func FormatArtifactList(state *schemas.ProjectState, phase string) string {
	var b strings.Builder

	if phase == "" || phase == PhaseDiscovery {
		formatPhaseArtifacts(&b, "Discovery", state.Phases.Discovery.Enabled, state.Phases.Discovery.Artifacts)
	}

	if phase == "" || phase == PhaseDesign {
		if phase == "" && b.Len() > 0 {
			fmt.Fprintln(&b)
		}
		formatPhaseArtifacts(&b, "Design", state.Phases.Design.Enabled, state.Phases.Design.Artifacts)
	}

	if b.Len() == 0 {
		return "No artifacts\n"
	}

	return b.String()
}

// formatPhaseArtifacts formats artifacts for a single phase.
func formatPhaseArtifacts(b *strings.Builder, phaseName string, enabled bool, artifacts []schemas.Artifact) {
	if !enabled {
		fmt.Fprintf(b, "%s Phase (disabled)\n", phaseName)
		return
	}

	fmt.Fprintf(b, "%s Phase:\n", phaseName)

	if len(artifacts) == 0 {
		fmt.Fprintln(b, "  No artifacts")
		return
	}

	fmt.Fprintln(b, "  [approved] Path")
	for _, artifact := range artifacts {
		approvedMark := " "
		if artifact.Approved {
			approvedMark = "✓"
		}
		fmt.Fprintf(b, "  [%s] %s\n", approvedMark, artifact.Path)
	}
}

// ============================================================================
// Review Management
// ============================================================================

// IncrementReviewIteration increments the review iteration counter.
//
// Parameters:
//   - state: Project state to modify
//
// Returns:
//   - nil on success
func IncrementReviewIteration(state *schemas.ProjectState) error {
	now := time.Now()

	state.Phases.Review.Iteration++
	state.Project.Updated_at = now

	return nil
}

// AddReviewReport adds a review report to the review phase.
//
// Parameters:
//   - state: Project state to modify
//   - path: Report path (relative to .sow/project/phases/review/)
//   - assessment: Assessment result ("pass" or "fail")
//
// Returns:
//   - nil on success
//   - error if validation fails
func AddReviewReport(state *schemas.ProjectState, path, assessment string) error {
	now := time.Now()

	// Validate assessment
	if assessment != "pass" && assessment != "fail" {
		return fmt.Errorf("invalid assessment '%s': must be 'pass' or 'fail'", assessment)
	}

	// Generate next report ID
	reportID := nextReviewReportID(state)

	// Create new report
	newReport := schemas.ReviewReport{
		Id:         reportID,
		Path:       path,
		Created_at: now,
		Assessment: assessment,
	}

	// Append to reports list
	state.Phases.Review.Reports = append(state.Phases.Review.Reports, newReport)

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// nextReviewReportID calculates the next review report ID (001, 002, 003...).
func nextReviewReportID(state *schemas.ProjectState) string {
	count := len(state.Phases.Review.Reports)
	return fmt.Sprintf("%03d", count+1)
}

// ============================================================================
// Finalize Management
// ============================================================================

// AddDocumentationUpdate adds a documentation file to the finalize phase tracking.
//
// Parameters:
//   - state: Project state to modify
//   - path: Path to documentation file (relative to repo root)
//
// Returns:
//   - nil on success
//   - error if path already tracked
func AddDocumentationUpdate(state *schemas.ProjectState, path string) error {
	now := time.Now()

	// Get existing list or create empty slice
	docs := extractDocumentationUpdates(state.Phases.Finalize.Documentation_updates)

	// Check if already tracked
	for _, doc := range docs {
		if doc == path {
			return fmt.Errorf("documentation file '%s' is already tracked", path)
		}
	}

	// Append to list
	docs = append(docs, path)
	state.Phases.Finalize.Documentation_updates = docs

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// extractDocumentationUpdates converts the any type to []string.
func extractDocumentationUpdates(value any) []string {
	if value == nil {
		return []string{}
	}

	// Type assert from any to []interface{} then convert to []string
	if docsList, ok := value.([]interface{}); ok {
		docs := make([]string, len(docsList))
		for i, doc := range docsList {
			if docStr, ok := doc.(string); ok {
				docs[i] = docStr
			}
		}
		return docs
	}

	// Already the correct type
	if docsList, ok := value.([]string); ok {
		return docsList
	}

	return []string{}
}

// MovedArtifact represents an artifact moved from project to knowledge.
type MovedArtifact struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// AddMovedArtifact records an artifact that was moved from project to knowledge.
//
// Parameters:
//   - state: Project state to modify
//   - from: Source path (relative to .sow/project/)
//   - to: Destination path (relative to .sow/)
//
// Returns:
//   - nil on success
//   - error if validation fails
func AddMovedArtifact(state *schemas.ProjectState, from, to string) error {
	now := time.Now()

	// Validate destination is under .sow/knowledge/
	if !strings.HasPrefix(to, "knowledge/") {
		return fmt.Errorf("destination '%s' must be under knowledge/ directory", to)
	}

	// Get existing list or create empty slice
	artifacts := extractMovedArtifacts(state.Phases.Finalize.Artifacts_moved)

	// Create moved artifact record
	movedArtifact := MovedArtifact{
		From: from,
		To:   to,
	}

	// Append to list
	artifacts = append(artifacts, movedArtifact)
	state.Phases.Finalize.Artifacts_moved = artifacts

	// Update project timestamp
	state.Project.Updated_at = now

	return nil
}

// extractMovedArtifacts converts the any type to []MovedArtifact.
func extractMovedArtifacts(value any) []MovedArtifact {
	if value == nil {
		return []MovedArtifact{}
	}

	// Already the correct type
	if artifactsList, ok := value.([]MovedArtifact); ok {
		return artifactsList
	}

	// Type assert from any to []interface{} then convert to []MovedArtifact
	if artifactsList, ok := value.([]interface{}); ok {
		return convertToMovedArtifacts(artifactsList)
	}

	return []MovedArtifact{}
}

// convertToMovedArtifacts converts []interface{} to []MovedArtifact.
func convertToMovedArtifacts(list []interface{}) []MovedArtifact {
	artifacts := make([]MovedArtifact, len(list))
	for i, artifact := range list {
		if artifactMap, ok := artifact.(map[string]interface{}); ok {
			if from, ok := artifactMap["from"].(string); ok {
				artifacts[i].From = from
			}
			if to, ok := artifactMap["to"].(string); ok {
				artifacts[i].To = to
			}
		}
	}
	return artifacts
}
