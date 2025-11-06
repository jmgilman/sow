package design

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
)

//go:embed templates/*.md
var templatesFS embed.FS

// configurePrompts registers all prompt generators with the project type builder.
// Prompts provide contextual guidance for the orchestrator and state-specific instructions.
// Returns the builder to enable method chaining.
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithOrchestratorPrompt(generateOrchestratorPrompt).
		WithPrompt(state.State(Active), generateActivePrompt).
		WithPrompt(state.State(Finalizing), generateFinalizingPrompt)
}

// generateOrchestratorPrompt generates the orchestrator-level prompt for design projects.
// This explains how the design project type works and how to coordinate work through phases.
// Returns a formatted prompt string. If rendering fails, returns an error message.
func generateOrchestratorPrompt(p *state.Project) string {
	if p == nil {
		return "Error: nil project provided to orchestrator prompt generator"
	}

	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Plan design documents, draft documents, request reviews, and approve.
// Returns a formatted prompt combining dynamic project state with static guidance.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Design: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Design\n\n")

	// Design phase info
	phase, exists := p.Phases["design"]
	if !exists {
		return "Error: design phase not found"
	}

	// Show inputs if any
	if len(phase.Inputs) > 0 {
		buf.WriteString("### Design Inputs\n\n")
		buf.WriteString("Sources informing this design:\n\n")
		for _, input := range phase.Inputs {
			buf.WriteString(fmt.Sprintf("- %s\n", input.Path))
			if input.Metadata != nil {
				if desc, ok := input.Metadata["description"].(string); ok && desc != "" {
					buf.WriteString(fmt.Sprintf("  %s\n", desc))
				}
			}
		}
		buf.WriteString("\n")
	}

	// Document tasks
	buf.WriteString("### Design Documents\n\n")

	//nolint:nestif // Complex but readable prompt generation logic
	if len(phase.Tasks) == 0 {
		buf.WriteString("No documents planned yet.\n\n")
		buf.WriteString("**Important**: Create at least one document task before adding artifacts.\n\n")
		buf.WriteString("**Next steps**: Plan document tasks\n\n")
	} else {
		// Count task statuses
		pending := 0
		inProgress := 0
		needsReview := 0
		completed := 0
		abandoned := 0

		for _, task := range phase.Tasks {
			switch task.Status {
			case "pending":
				pending++
			case "in_progress":
				inProgress++
			case "needs_review":
				needsReview++
			case "completed":
				completed++
			case "abandoned":
				abandoned++
			}
		}

		buf.WriteString(fmt.Sprintf("Total: %d documents\n", len(phase.Tasks)))
		buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
		buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
		buf.WriteString(fmt.Sprintf("- Needs Review: %d\n", needsReview))
		buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List documents with status icons
		for _, task := range phase.Tasks {
			statusIcon := getStatusIcon(task.Status)
			buf.WriteString(fmt.Sprintf("%s %s - %s (%s)\n", statusIcon, task.Id, task.Name, task.Status))

			// Show artifact if linked
			if task.Metadata != nil {
				if artifactPath, ok := task.Metadata["artifact_path"].(string); ok && artifactPath != "" {
					buf.WriteString(fmt.Sprintf("    Artifact: %s\n", artifactPath))
				}
				if docType, ok := task.Metadata["document_type"].(string); ok && docType != "" {
					buf.WriteString(fmt.Sprintf("    Type: %s\n", docType))
				}
			}
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allDocumentsApproved(p) {
			buf.WriteString("✓ All documents approved!\n\n")
			buf.WriteString("Ready to finalize. Run: `sow project advance`\n\n")
		} else {
			unresolvedCount := countUnresolvedTasks(p)
			buf.WriteString(fmt.Sprintf("**Next steps**: Continue design work (%d documents remaining)\n\n", unresolvedCount))
		}
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/active.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// generateFinalizingPrompt generates the prompt for the Finalizing state.
// Focus: Move documents to targets, create PR, cleanup.
// Returns a formatted prompt combining dynamic finalization status with static guidance.
func generateFinalizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Design: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Finalizing\n\n")
	buf.WriteString("All documents approved. Finalizing design by moving artifacts, creating PR, and cleaning up.\n\n")

	// Finalization tasks
	phase, exists := p.Phases["finalization"]
	if !exists {
		return "Error: finalization phase not found"
	}

	buf.WriteString("### Finalization Tasks\n\n")
	for _, task := range phase.Tasks {
		status := "[ ]"
		if task.Status == "completed" {
			status = "[✓]"
		}
		buf.WriteString(fmt.Sprintf("%s %s\n", status, task.Name))
	}
	buf.WriteString("\n")

	// Advancement readiness
	if allFinalizationTasksComplete(p) {
		buf.WriteString("✓ All finalization tasks complete!\n\n")
		buf.WriteString("Ready to complete design. Run: `sow project advance`\n\n")
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/finalizing.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// getStatusIcon returns the appropriate icon for a task status.
// Uses consistent icons across all prompts for visual clarity.
func getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "[✓]"
	case "abandoned":
		return "[✗]"
	case "needs_review":
		return "[?]"
	case "in_progress":
		return "[~]"
	case "pending":
		return "[ ]"
	default:
		return "[ ]"
	}
}
