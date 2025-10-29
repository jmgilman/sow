package standard

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// StatePromptID uniquely identifies a prompt template for standard project states.
type StatePromptID string

// State prompt IDs - one for each state in the standard project state machine.
const (
	PromptNoProject               StatePromptID = "no_project"
	PromptPlanningActive          StatePromptID = "planning_active"
	PromptImplementationPlanning  StatePromptID = "implementation_planning"
	PromptImplementationExecuting StatePromptID = "implementation_executing"
	PromptReviewActive            StatePromptID = "review_active"
	PromptFinalizeDocumentation   StatePromptID = "finalize_documentation"
	PromptFinalizeChecks          StatePromptID = "finalize_checks"
	PromptFinalizeDelete          StatePromptID = "finalize_delete"
)

// Embed standard project templates
//
//go:embed templates/*.md
var templatesFS embed.FS

// standardRegistry is the prompt registry for standard project state templates.
var standardRegistry *prompts.Registry[StatePromptID]

func init() {
	standardRegistry = prompts.NewRegistry[StatePromptID]()

	// Map state prompt IDs to template files
	statePrompts := map[StatePromptID]string{
		PromptNoProject:               "templates/no_project.md",
		PromptPlanningActive:          "templates/planning_active.md",
		PromptImplementationPlanning:  "templates/implementation_planning.md",
		PromptImplementationExecuting: "templates/implementation_executing.md",
		PromptReviewActive:            "templates/review_active.md",
		PromptFinalizeDocumentation:   "templates/finalize_documentation.md",
		PromptFinalizeChecks:          "templates/finalize_checks.md",
		PromptFinalizeDelete:          "templates/finalize_delete.md",
	}

	// Load and parse all templates from the embedded filesystem
	for id, path := range statePrompts {
		if err := standardRegistry.RegisterFromFS(templatesFS, id, path); err != nil {
			panic(fmt.Sprintf("failed to register standard project prompt %s: %v", id, err))
		}
	}
}

// stateToPromptID maps standard project states to their corresponding prompt template IDs.
func stateToPromptID(state statechart.State) StatePromptID {
	switch state {
	case statechart.NoProject:
		return PromptNoProject
	case PlanningActive:
		return PromptPlanningActive
	case ImplementationPlanning:
		return PromptImplementationPlanning
	case ImplementationExecuting:
		return PromptImplementationExecuting
	case ReviewActive:
		return PromptReviewActive
	case FinalizeDocumentation:
		return PromptFinalizeDocumentation
	case FinalizeChecks:
		return PromptFinalizeChecks
	case FinalizeDelete:
		return PromptFinalizeDelete
	default:
		// Return empty, caller will handle unknown state
		return ""
	}
}

// StandardPromptGenerator implements the PromptGenerator interface for standard projects.
// It generates contextual prompts for each state in the standard project state machine,
// combining reusable components with state-specific logic.
//
//nolint:revive // Name intentionally mirrors standard package pattern
type StandardPromptGenerator struct {
	components *statechart.PromptComponents
}

// NewStandardPromptGenerator creates a new StandardPromptGenerator with access to external systems.
func NewStandardPromptGenerator(ctx *sow.Context) *StandardPromptGenerator {
	return &StandardPromptGenerator{
		components: statechart.NewPromptComponents(ctx),
	}
}

// GeneratePrompt generates a contextual prompt for the given state.
// It routes to state-specific generation methods based on the current state.
func (g *StandardPromptGenerator) GeneratePrompt(
	state statechart.State,
	projectState *schemas.ProjectState,
) (string, error) {
	switch state {
	case statechart.NoProject:
		// NoProject state has no prompt - project is deleted
		return "", nil
	case PlanningActive:
		return g.generatePlanningPrompt(projectState)
	case ImplementationPlanning:
		return g.generateImplementationPlanningPrompt(projectState)
	case ImplementationExecuting:
		return g.generateImplementationExecutingPrompt(projectState)
	case ReviewActive:
		return g.generateReviewPrompt(projectState)
	case FinalizeDocumentation:
		return g.generateFinalizeDocumentationPrompt(projectState)
	case FinalizeChecks:
		return g.generateFinalizeChecksPrompt(projectState)
	case FinalizeDelete:
		return g.generateFinalizeDeletePrompt(projectState)
	default:
		return "", fmt.Errorf("unknown state: %s", state)
	}
}

// generatePlanningPrompt generates the prompt for the PlanningActive state.
// This phase focuses on gathering context, confirming requirements, and creating a task list.
func (g *StandardPromptGenerator) generatePlanningPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Add git status
	gitStatus, err := g.components.GitStatus()
	if err != nil {
		// Log warning but continue (git status is nice-to-have)
		buf.WriteString("## Git Status\n\nUnavailable\n\n")
	} else {
		buf.WriteString(gitStatus)
		buf.WriteString("\n")
	}

	// Render static guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(PlanningActive),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptPlanningActive, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render planning template: %w", err)
	}
	buf.WriteString(guidance)

	// Show artifact status if artifacts exist
	if len(projectState.Phases.Planning.Artifacts) > 0 {
		buf.WriteString("\n## Planning Artifacts\n\n")
		for _, artifact := range projectState.Phases.Planning.Artifacts {
			status := "pending"
			if artifact.Approved {
				status = "approved"
			}
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", artifact.Path, status))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// generateImplementationPlanningPrompt generates the prompt for the ImplementationPlanning state.
// This phase focuses on breaking down the work into concrete implementation tasks.
func (g *StandardPromptGenerator) generateImplementationPlanningPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Include planning phase artifacts as context
	if len(projectState.Phases.Planning.Artifacts) > 0 {
		buf.WriteString("## Planning Context\n\n")
		buf.WriteString("The following artifacts were created during planning:\n\n")
		for _, artifact := range projectState.Phases.Planning.Artifacts {
			buf.WriteString(fmt.Sprintf("- %s\n", artifact.Path))
		}
		buf.WriteString("\n")
	}

	// Render implementation planning guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(ImplementationPlanning),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptImplementationPlanning, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render implementation planning template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// generateImplementationExecutingPrompt generates the prompt for the ImplementationExecuting state.
// This phase focuses on executing implementation tasks with status tracking.
func (g *StandardPromptGenerator) generateImplementationExecutingPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Add task summary
	buf.WriteString(g.components.TaskSummary(projectState.Phases.Implementation.Tasks))
	buf.WriteString("\n")

	// Conditionally include recent commits if any tasks have been completed
	hasCompleted := false
	for _, t := range projectState.Phases.Implementation.Tasks {
		if t.Status == "completed" {
			hasCompleted = true
			break
		}
	}

	if hasCompleted {
		commits, err := g.components.RecentCommits(5)
		if err != nil {
			// Log warning but continue (commits are nice-to-have)
			buf.WriteString("## Recent Commits\n\nUnavailable\n\n")
		} else {
			buf.WriteString(commits)
			buf.WriteString("\n")
		}
	}

	// Render execution guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(ImplementationExecuting),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptImplementationExecuting, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render implementation executing template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// generateReviewPrompt generates the prompt for the ReviewActive state.
// This phase focuses on reviewing the implementation and providing feedback.
func (g *StandardPromptGenerator) generateReviewPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Display review iteration from metadata
	iteration := 1
	if projectState.Phases.Review.Metadata != nil {
		if iter, ok := projectState.Phases.Review.Metadata["iteration"].(int); ok {
			iteration = iter
		}
	}
	buf.WriteString(fmt.Sprintf("## Review Iteration: %d\n\n", iteration))

	// Show previous review assessment if iteration > 1
	if iteration > 1 {
		buf.WriteString("### Previous Review\n\n")
		if prevReview := findPreviousReviewArtifact(projectState, iteration-1); prevReview != nil {
			assessment := extractReviewAssessment(prevReview)
			buf.WriteString(fmt.Sprintf("Assessment: %s\n", assessment))
			buf.WriteString(fmt.Sprintf("Report: %s\n\n", prevReview.Path))
		}
	}

	// Include task completion summary
	buf.WriteString(g.components.TaskSummary(projectState.Phases.Implementation.Tasks))
	buf.WriteString("\n")

	// Render review guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(ReviewActive),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptReviewActive, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render review template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// generateFinalizeDocumentationPrompt generates the prompt for the FinalizeDocumentation state.
// This phase focuses on updating documentation to reflect the changes made.
func (g *StandardPromptGenerator) generateFinalizeDocumentationPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Render finalize documentation guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(FinalizeDocumentation),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptFinalizeDocumentation, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render finalize documentation template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// generateFinalizeChecksPrompt generates the prompt for the FinalizeChecks state.
// This phase focuses on running final validation checks before completion.
func (g *StandardPromptGenerator) generateFinalizeChecksPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Render finalize checks guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(FinalizeChecks),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptFinalizeChecks, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render finalize checks template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// generateFinalizeDeletePrompt generates the prompt for the FinalizeDelete state.
// This phase focuses on cleaning up the project directory and creating a PR.
func (g *StandardPromptGenerator) generateFinalizeDeletePrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Add project header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Include PR URL if available
	if projectState.Phases.Finalize.Metadata != nil {
		if prURL, ok := projectState.Phases.Finalize.Metadata["pr_url"].(string); ok && prURL != "" {
			buf.WriteString(fmt.Sprintf("## Pull Request\n\n%s\n\n", prURL))
		}
	}

	// Render finalize delete guidance template using local registry
	templateCtx := &prompts.StatechartContext{
		State:        string(FinalizeDelete),
		ProjectState: projectState,
	}
	guidance, err := standardRegistry.Render(PromptFinalizeDelete, templateCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render finalize delete template: %w", err)
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// findPreviousReviewArtifact searches for a review artifact from a specific iteration.
func findPreviousReviewArtifact(projectState *schemas.ProjectState, targetIteration int) *phases.Artifact {
	for i := len(projectState.Phases.Review.Artifacts) - 1; i >= 0; i-- {
		artifact := &projectState.Phases.Review.Artifacts[i]
		if !isReviewArtifact(artifact) {
			continue
		}
		if iter, ok := artifact.Metadata["iteration"].(int); ok && iter == targetIteration {
			return artifact
		}
	}
	return nil
}

// isReviewArtifact checks if an artifact is a review artifact.
func isReviewArtifact(artifact *phases.Artifact) bool {
	if artifact.Metadata == nil {
		return false
	}
	artifactType, ok := artifact.Metadata["type"].(string)
	return ok && artifactType == "review"
}

// extractReviewAssessment extracts the assessment string from a review artifact.
func extractReviewAssessment(artifact *phases.Artifact) string {
	if artifact.Metadata == nil {
		return "unknown"
	}
	if assess, ok := artifact.Metadata["assessment"].(string); ok {
		return assess
	}
	return "unknown"
}
