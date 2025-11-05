package standard

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

//go:embed templates/*.md
var templatesFS embed.FS

// generateOrchestratorPrompt generates the orchestrator-level prompt for standard projects.
// This explains how the standard project type works and how to coordinate work through phases.
func generateOrchestratorPrompt(p *state.Project) string {
	// Render orchestrator guidance template
	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateImplementationPlanningPrompt generates the prompt for the ImplementationPlanning state.
// This phase focuses on gathering requirements and creating a high-level task list.
func generateImplementationPlanningPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Check if this is a rework iteration
	implPhase, exists := p.Phases["implementation"]
	if exists && implPhase.Iteration > 1 {
		// Show rework iteration number
		buf.WriteString(fmt.Sprintf("## 🔄 Rework Iteration: %d\n\n", implPhase.Iteration))

		// Show previous review failure context
		reviewPhase, rExists := p.Phases["review"]
		if rExists && reviewPhase.Status == "failed" {
			buf.WriteString("⚠️ **Previous review failed** - tasks must address identified issues.\n\n")
			addFailedReviewContext(&buf, &reviewPhase)
		}
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

	// Show iteration if this is rework
	if implPhase, exists := p.Phases["implementation"]; exists && implPhase.Iteration > 1 {
		buf.WriteString(fmt.Sprintf("## Implementation Iteration: %d\n\n", implPhase.Iteration))
	}

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
			buf.WriteString(fmt.Sprintf("Report: %s\n", prevReview.Path))

			// Show when implementation was marked as failed
			if implPhase, exists := p.Phases["implementation"]; exists && implPhase.Failed_at.Year() > 1 {
				buf.WriteString(fmt.Sprintf("Implementation marked failed: %s\n", implPhase.Failed_at.Format("2006-01-02 15:04:05")))
			}
			buf.WriteString("\n")
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

// generateFinalizePRCreationPrompt generates the prompt for the FinalizePRCreation state.
// This phase focuses on creating PR body document and getting approval to create PR.
func generateFinalizePRCreationPrompt(p *state.Project) string {
	var buf strings.Builder

	// Add project header
	buf.WriteString(fmt.Sprintf("# Project: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Show PR body artifact status if exists
	if finalizePhase, exists := p.Phases["finalize"]; exists && len(finalizePhase.Outputs) > 0 {
		buf.WriteString("## PR Body Artifact\n\n")
		for _, artifact := range finalizePhase.Outputs {
			if artifact.Type == "pr_body" {
				status := "pending approval"
				if artifact.Approved {
					status = "approved"
				}
				buf.WriteString(fmt.Sprintf("- %s (%s)\n\n", artifact.Path, status))
			}
		}
	}

	// Render finalize PR creation guidance template
	guidance, err := templates.Render(templatesFS, "templates/finalize_pr_creation.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// addFailedReviewContext adds failed review context to the buffer.
func addFailedReviewContext(buf *strings.Builder, reviewPhase *projschema.PhaseState) {
	// Find latest failed review
	for i := len(reviewPhase.Outputs) - 1; i >= 0; i-- {
		artifact := &reviewPhase.Outputs[i]
		if artifact.Type == "review" && artifact.Approved {
			assessment, ok := artifact.Metadata["assessment"].(string)
			if ok && assessment == "fail" {
				fmt.Fprintf(buf, "Review report: %s\n", artifact.Path)
				if reviewPhase.Failed_at.Year() > 1 {
					fmt.Fprintf(buf, "Failed at: %s\n", reviewPhase.Failed_at.Format("2006-01-02 15:04:05"))
				}
				buf.WriteString("\n")
				break
			}
		}
	}
}

// generateFinalizeCleanupPrompt generates the prompt for the FinalizeCleanup state.
// This phase focuses on cleaning up the project directory after PR creation.
func generateFinalizeCleanupPrompt(p *state.Project) string {
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

	// Render finalize cleanup guidance template
	guidance, err := templates.Render(templatesFS, "templates/finalize_cleanup.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// Helper functions

// taskSummary generates a summary of tasks with their status breakdown.
// Shows total, completed, in-progress, and pending task counts.
func taskSummary(tasks []projschema.TaskState) string {
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
func findPreviousReviewArtifact(p *state.Project, targetIteration int) *projschema.ArtifactState {
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
func isReviewArtifact(artifact *projschema.ArtifactState) bool {
	if artifact.Metadata == nil {
		return false
	}
	artifactType, ok := artifact.Metadata["type"].(string)
	return ok && artifactType == "review"
}

// extractReviewAssessment extracts the assessment string from a review artifact.
func extractReviewAssessment(artifact *projschema.ArtifactState) string {
	if artifact.Metadata == nil {
		return "unknown"
	}
	if assess, ok := artifact.Metadata["assessment"].(string); ok {
		return assess
	}
	return "unknown"
}
