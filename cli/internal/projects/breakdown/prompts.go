package breakdown

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
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

// generateDiscoveryPrompt generates the prompt for the Discovery state.
// Focus: Gather codebase/design context to inform work unit identification.
// Returns a formatted prompt with guidance on creating discovery documents.
func generateDiscoveryPrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Breakdown: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")
	buf.WriteString("## Current State: Discovery\n\n")

	// Breakdown phase info
	phase, exists := p.Phases["breakdown"]
	if !exists {
		return "Error: breakdown phase not found"
	}

	// Show inputs if any
	writeDiscoveryInputs(&buf, phase)

	// Discovery status and readiness
	writeDiscoveryStatus(&buf, p, phase)

	// Render guidance from template
	guidance, err := templates.Render(templatesFS, "templates/discovery.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// writeDiscoveryInputs writes the input materials section if inputs exist.
func writeDiscoveryInputs(buf *strings.Builder, phase projschema.PhaseState) {
	if len(phase.Inputs) == 0 {
		return
	}

	buf.WriteString("### Input Materials\n\n")
	buf.WriteString("Sources to analyze for breakdown:\n\n")
	for _, input := range phase.Inputs {
		fmt.Fprintf(buf, "- %s (%s)\n", input.Path, input.Type)
		if input.Metadata != nil {
			if desc, ok := input.Metadata["description"].(string); ok && desc != "" {
				fmt.Fprintf(buf, "  %s\n", desc)
			}
		}
	}
	buf.WriteString("\n")
}

// writeDiscoveryStatus writes the discovery status and advancement readiness sections.
func writeDiscoveryStatus(buf *strings.Builder, p *state.Project, phase projschema.PhaseState) {
	buf.WriteString("### Discovery Status\n\n")

	hasDiscovery := false
	for _, artifact := range phase.Outputs {
		if artifact.Type == "discovery" {
			hasDiscovery = true
			status := "[ ] Pending approval"
			if artifact.Approved {
				status = "[✓] Approved"
			}
			fmt.Fprintf(buf, "%s Discovery document: %s\n", status, artifact.Path)
		}
	}

	if !hasDiscovery {
		buf.WriteString("No discovery document created yet.\n\n")
	} else {
		buf.WriteString("\n")
	}

	// Advancement readiness
	if hasApprovedDiscoveryDocument(p) {
		buf.WriteString("✓ Discovery document approved!\n\n")
		buf.WriteString("Ready to begin work unit identification. Run: `sow project advance`\n\n")
	} else {
		buf.WriteString("**Next steps**: Create and approve discovery document\n\n")
	}
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Identify work units, spawn decomposer per unit, review specifications.
// Returns a formatted prompt combining dynamic project state with static guidance.
//
//nolint:funlen // Complex but readable prompt generation logic
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
		buf.WriteString("**Let's start by reviewing the discovery document together.**\n\n")
		buf.WriteString("I suggest we:\n")
		buf.WriteString("1. Review the approved discovery document to understand existing code\n")
		buf.WriteString("2. Identify project-sized work units (4-5 days each minimum)\n")
		buf.WriteString("3. Discuss and refine the decomposition until you're satisfied\n")
		buf.WriteString("4. Create tasks for approved work units\n")
		buf.WriteString("5. Prepare context and spawn decomposer agents to write specifications\n\n")
		buf.WriteString("**Remember**: Each work unit should include tests via TDD. No separate testing work units.\n\n")
		buf.WriteString("Ready to review the discovery and propose a breakdown?\n\n")
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
			buf.WriteString("We're ready to move to the Publishing phase to create GitHub issues.\n\n")
			buf.WriteString("Should I advance to Publishing? This will let us create GitHub issues for these work units.\n\n")
		} else {
			if !allWorkUnitsApproved(p) {
				unresolvedCount := countUnresolvedTasks(p)
				buf.WriteString(fmt.Sprintf("**Status**: %d work units still need attention.\n\n", unresolvedCount))
				buf.WriteString("Which work unit would you like to focus on next?\n\n")
			} else {
				buf.WriteString("**Issue**: Dependency validation failed - there may be cycles or invalid references in the dependency graph.\n\n")
				buf.WriteString("Should I review the dependencies and identify the issue?\n\n")
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
	buf.WriteString("All work units are approved and ready to be published as GitHub issues.\n\n")
	buf.WriteString("**Let's review the publishing plan together before proceeding.**\n\n")

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
		buf.WriteString("✓ All work units published successfully!\n\n")
		buf.WriteString("The breakdown is complete. All GitHub issues have been created with the 'sow' label.\n\n")
		buf.WriteString("Should I mark the project as completed?\n\n")
	} else {
		if unpublished == len(completed) {
			// None published yet
			buf.WriteString(fmt.Sprintf("Ready to publish %d work units as GitHub issues.\n\n", unpublished))
			buf.WriteString("Should I review the publishing plan with you before proceeding?\n\n")
		} else {
			// Some published, some remaining
			buf.WriteString(fmt.Sprintf("**Progress**: %d of %d work units published. %d remaining.\n\n", published, len(completed), unpublished))
			buf.WriteString("Should I continue publishing the remaining work units?\n\n")
		}
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
