package standard

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	"github.com/jmgilman/sow/cli/schemas/project"
)

//go:embed templates/*.md
var templatesFS embed.FS

// generatePlanningPrompt generates the prompt for the PlanningActive state.
// This phase focuses on gathering context, confirming requirements, and creating a task list.
func generatePlanningPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Render template guidance
	guidance, err := templates.Render(templatesFS, "templates/planning_active.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	// Show artifacts if any
	phase, exists := p.Phases["planning"]
	if exists && len(phase.Outputs) > 0 {
		buf.WriteString("\n## Planning Artifacts\n\n")
		for _, artifact := range phase.Outputs {
			status := "pending"
			if artifact.Approved {
				status = "approved"
			}
			buf.WriteString(fmt.Sprintf("- %s (%s)\n", artifact.Path, status))
		}
	}

	return buf.String()
}

// generateImplementationPlanningPrompt generates the prompt for the ImplementationPlanning state.
// This phase focuses on breaking down the work into concrete implementation tasks.
func generateImplementationPlanningPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Include planning phase artifacts as context
	if planningPhase, exists := p.Phases["planning"]; exists && len(planningPhase.Outputs) > 0 {
		buf.WriteString("## Planning Context\n\n")
		buf.WriteString("The following artifacts were created during planning:\n\n")
		for _, artifact := range planningPhase.Outputs {
			buf.WriteString(fmt.Sprintf("- %s\n", artifact.Path))
		}
		buf.WriteString("\n")
	}

	// Render implementation planning guidance template
	guidance, err := templates.Render(templatesFS, "templates/implementation_planning.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateImplementationExecutingPrompt generates the prompt for the ImplementationExecuting state.
// This phase focuses on executing implementation tasks with status tracking.
func generateImplementationExecutingPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Add task summary
	if implPhase, exists := p.Phases["implementation"]; exists && len(implPhase.Tasks) > 0 {
		buf.WriteString(taskSummary(implPhase.Tasks))
		buf.WriteString("\n")
	}

	// Render execution guidance template
	guidance, err := templates.Render(templatesFS, "templates/implementation_executing.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateReviewPrompt generates the prompt for the ReviewActive state.
// This phase focuses on reviewing the implementation and providing feedback.
func generateReviewPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Display review iteration from metadata
	iteration := 1
	if reviewPhase, exists := p.Phases["review"]; exists && reviewPhase.Metadata != nil {
		if iter, ok := reviewPhase.Metadata["iteration"].(int); ok {
			iteration = iter
		} else if iter64, ok := reviewPhase.Metadata["iteration"].(int64); ok {
			iteration = int(iter64)
		}
	}
	buf.WriteString(fmt.Sprintf("## Review Iteration: %d\n\n", iteration))

	// Show previous review assessment if iteration > 1
	if iteration > 1 {
		buf.WriteString("### Previous Review\n\n")
		if prevReview := findPreviousReviewArtifact(p, iteration-1); prevReview != nil {
			assessment := extractReviewAssessment(prevReview)
			buf.WriteString(fmt.Sprintf("Assessment: %s\n", assessment))
			buf.WriteString(fmt.Sprintf("Report: %s\n\n", prevReview.Path))
		}
	}

	// Include task completion summary
	if implPhase, exists := p.Phases["implementation"]; exists && len(implPhase.Tasks) > 0 {
		buf.WriteString(taskSummary(implPhase.Tasks))
		buf.WriteString("\n")
	}

	// Render review guidance template
	guidance, err := templates.Render(templatesFS, "templates/review_active.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizeDocumentationPrompt generates the prompt for the FinalizeDocumentation state.
// This phase focuses on updating documentation to reflect the changes made.
func generateFinalizeDocumentationPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Render finalize documentation guidance template
	guidance, err := templates.Render(templatesFS, "templates/finalize_documentation.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizeChecksPrompt generates the prompt for the FinalizeChecks state.
// This phase focuses on running final validation checks before completion.
func generateFinalizeChecksPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Render finalize checks guidance template
	guidance, err := templates.Render(templatesFS, "templates/finalize_checks.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizeDeletePrompt generates the prompt for the FinalizeDelete state.
// This phase focuses on cleaning up the project directory and creating a PR.
func generateFinalizeDeletePrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Include PR URL if available
	if finalizePhase, exists := p.Phases["finalize"]; exists && finalizePhase.Metadata != nil {
		if prURL, ok := finalizePhase.Metadata["pr_url"].(string); ok && prURL != "" {
			buf.WriteString(fmt.Sprintf("## Pull Request\n\n%s\n\n", prURL))
		}
	}

	// Render finalize delete guidance template
	guidance, err := templates.Render(templatesFS, "templates/finalize_delete.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// Helper functions

// taskSummary generates a summary of tasks with their status breakdown.
// Shows total, completed, in-progress, and pending task counts.
func taskSummary(tasks []project.TaskState) string {
	var buf strings.Builder

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

	buf.WriteString(fmt.Sprintf("## Tasks (%d total)\n\n", total))
	if completed > 0 {
		buf.WriteString(fmt.Sprintf("- %d completed\n", completed))
	}
	if inProgress > 0 {
		buf.WriteString(fmt.Sprintf("- %d in progress\n", inProgress))
	}
	if pending > 0 {
		buf.WriteString(fmt.Sprintf("- %d pending\n", pending))
	}
	if abandoned > 0 {
		buf.WriteString(fmt.Sprintf("- %d abandoned\n", abandoned))
	}

	return buf.String()
}

// findPreviousReviewArtifact searches for a review artifact from a specific iteration.
func findPreviousReviewArtifact(p *state.Project, targetIteration int) *project.ArtifactState {
	reviewPhase, exists := p.Phases["review"]
	if !exists {
		return nil
	}

	// Search backwards through artifacts for matching iteration
	for i := len(reviewPhase.Outputs) - 1; i >= 0; i-- {
		artifact := &reviewPhase.Outputs[i]
		if !isReviewArtifact(artifact) {
			continue
		}
		if iter, ok := artifact.Metadata["iteration"].(int); ok && iter == targetIteration {
			return artifact
		}
		if iter64, ok := artifact.Metadata["iteration"].(int64); ok && int(iter64) == targetIteration {
			return artifact
		}
	}
	return nil
}

// isReviewArtifact checks if an artifact is a review artifact.
func isReviewArtifact(artifact *project.ArtifactState) bool {
	if artifact.Metadata == nil {
		return false
	}
	artifactType, ok := artifact.Metadata["type"].(string)
	return ok && artifactType == "review"
}

// extractReviewAssessment extracts the assessment string from a review artifact.
func extractReviewAssessment(artifact *project.ArtifactState) string {
	if artifact.Metadata == nil {
		return "unknown"
	}
	if assess, ok := artifact.Metadata["assessment"].(string); ok {
		return assess
	}
	return "unknown"
}
