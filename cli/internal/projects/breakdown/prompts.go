package breakdown

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

// generateOrchestratorPrompt generates the orchestrator-level prompt for breakdown projects.
// This explains how the breakdown project type works and how to coordinate work through phases.
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
// Focus: Decompose work, specify units, review all work units.
// Returns a formatted prompt combining dynamic project state with static guidance.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Breakdown: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Breakdown\n\n")

	// Breakdown phase info
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return "Error: breakdown phase not found"
	}

	// Show inputs if any
	if len(phase.Inputs) > 0 {
		buf.WriteString("### Being Broken Down\n\n")
		buf.WriteString("Sources being decomposed:\n\n")
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

	// Work units
	buf.WriteString("### Work Units\n\n")

	//nolint:nestif // Complex but readable prompt generation logic
	if len(phase.Tasks) == 0 {
		buf.WriteString("No work units identified yet.\n\n")
		buf.WriteString("**Important**: Create at least one work unit task before adding specifications.\n\n")
		buf.WriteString("**Next steps**: Identify work units to create\n\n")
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

		buf.WriteString(fmt.Sprintf("Total: %d work units\n", len(phase.Tasks)))
		buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
		buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
		buf.WriteString(fmt.Sprintf("- Needs Review: %d\n", needsReview))
		buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List work units with status icons
		for _, task := range phase.Tasks {
			statusIcon := getStatusIcon(task.Status)
			buf.WriteString(fmt.Sprintf("%s %s - %s (%s)\n", statusIcon, task.Id, task.Name, task.Status))

			// Show dependencies if any
			if task.Metadata != nil {
				if depsRaw, ok := task.Metadata["dependencies"]; ok {
					if deps, ok := depsRaw.([]interface{}); ok && len(deps) > 0 {
						depStrs := make([]string, 0, len(deps))
						for _, d := range deps {
							if depStr, ok := d.(string); ok {
								depStrs = append(depStrs, depStr)
							}
						}
						if len(depStrs) > 0 {
							buf.WriteString(fmt.Sprintf("    Depends on: %v\n", depStrs))
						}
					}
				}

				// Show artifact path if linked
				if artifactPath, ok := task.Metadata["artifact_path"].(string); ok && artifactPath != "" {
					buf.WriteString(fmt.Sprintf("    Spec: %s\n", artifactPath))
				}
			}
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allWorkUnitsApproved(p) && dependenciesValid(p) {
			buf.WriteString("✓ All work units approved and dependencies validated!\n\n")
			buf.WriteString("Ready to publish GitHub issues. Run: `sow project advance`\n\n")
		} else {
			if !allWorkUnitsApproved(p) {
				unresolvedCount := countUnresolvedTasks(p)
				buf.WriteString(fmt.Sprintf("**Next steps**: Continue breakdown work (%d work units remaining)\n\n", unresolvedCount))
			} else {
				buf.WriteString("**Next steps**: Dependency validation failed - check for cycles or invalid references\n\n")
			}
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

// generatePublishingPrompt generates the prompt for the Publishing state.
// Focus: Create GitHub issues for work units in dependency order.
// Returns a formatted prompt combining dynamic publishing status with static guidance.
func generatePublishingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Breakdown: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Publishing\n\n")
	buf.WriteString("All work units approved. Creating GitHub issues in dependency order.\n\n")

	// Breakdown phase info
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return "Error: breakdown phase not found"
	}

	// Publishing status
	buf.WriteString("### Publishing Status\n\n")

	// Collect completed tasks
	completed := []projschema.TaskState{}
	for _, task := range phase.Tasks {
		if task.Status == "completed" {
			completed = append(completed, task)
		}
	}

	// Count published vs unpublished
	published := 0
	unpublished := 0
	for _, task := range completed {
		if task.Metadata != nil {
			if pub, ok := task.Metadata["published"].(bool); ok && pub {
				published++
			} else {
				unpublished++
			}
		} else {
			unpublished++
		}
	}

	buf.WriteString(fmt.Sprintf("Total work units: %d\n", len(completed)))
	buf.WriteString(fmt.Sprintf("Published: %d\n", published))
	buf.WriteString(fmt.Sprintf("Unpublished: %d\n\n", unpublished))

	// List publishing status for each work unit
	for _, task := range completed {
		isPublished := false
		var issueURL string

		if task.Metadata != nil {
			if pub, ok := task.Metadata["published"].(bool); ok && pub {
				isPublished = true
			}
			if url, ok := task.Metadata["github_issue_url"].(string); ok {
				issueURL = url
			}
		}

		status := "[ ] Pending"
		if isPublished {
			status = fmt.Sprintf("[✓] Published: %s", issueURL)
		}

		buf.WriteString(fmt.Sprintf("%s %s - %s\n", status, task.Id, task.Name))
	}
	buf.WriteString("\n")

	// Advancement readiness
	if allWorkUnitsPublished(p) {
		buf.WriteString("✓ All work units published!\n\n")
		buf.WriteString("Breakdown complete. Run: `sow project advance`\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("**Next steps**: Publish remaining %d work units\n\n", unpublished))
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/publishing.md", p)
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
