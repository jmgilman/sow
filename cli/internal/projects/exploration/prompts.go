package exploration

import (
	"embed"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/internal/sdks/project/templates"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

//go:embed templates/*.md
var templatesFS embed.FS

// configurePrompts registers all prompt generators with the project type builder.
// Prompts provide contextual guidance for the orchestrator and state-specific instructions.
func configurePrompts(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithOrchestratorPrompt(generateOrchestratorPrompt).
		WithPrompt(state.State(Active), generateActivePrompt).
		WithPrompt(state.State(Summarizing), generateSummarizingPrompt).
		WithPrompt(state.State(Finalizing), generateFinalizingPrompt)
}

// generateOrchestratorPrompt generates the orchestrator-level prompt for exploration projects.
// This explains how the exploration project type works and how to coordinate work through phases.
func generateOrchestratorPrompt(p *state.Project) string {
	prompt, err := templates.Render(templatesFS, "templates/orchestrator.md", p)
	if err != nil {
		return fmt.Sprintf("Error rendering orchestrator prompt: %v", err)
	}
	return prompt
}

// generateActivePrompt generates the prompt for the Active state.
// Focus: Identify research topics, investigate, document findings.
func generateActivePrompt(p *state.Project) string {
	var buf strings.Builder

	// Project header
	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", p.Branch))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", p.Description))
	}
	buf.WriteString("\n")

	// Current state
	buf.WriteString("## Current State: Active Research\n\n")

	// Research topics summary
	phase, exists := p.Phases["exploration"]
	if !exists {
		return "Error: exploration phase not found"
	}

	if len(phase.Tasks) == 0 {
		buf.WriteString("No research topics identified yet.\n\n")
		buf.WriteString("**Next steps**: Create research topics to investigate.\n\n")
	} else {
		// Count task statuses
		pending := 0
		inProgress := 0
		completed := 0
		abandoned := 0

		for _, task := range phase.Tasks {
			switch task.Status {
			case "pending":
				pending++
			case "in_progress":
				inProgress++
			case "completed":
				completed++
			case "abandoned":
				abandoned++
			}
		}

		buf.WriteString(fmt.Sprintf("### Research Topics (%d total)\n\n", len(phase.Tasks)))
		buf.WriteString(fmt.Sprintf("- Pending: %d\n", pending))
		buf.WriteString(fmt.Sprintf("- In Progress: %d\n", inProgress))
		buf.WriteString(fmt.Sprintf("- Completed: %d\n", completed))
		buf.WriteString(fmt.Sprintf("- Abandoned: %d\n\n", abandoned))

		// List topics
		for _, task := range phase.Tasks {
			buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", task.Id, task.Name, task.Status))
		}
		buf.WriteString("\n")

		// Advancement readiness
		if allTasksResolved(p) {
			buf.WriteString("✓ All research topics resolved!\n\n")
			buf.WriteString("Ready to create summary. Run: `sow project advance`\n\n")
		} else {
			unresolvedCount := countUnresolvedTasks(p)
			buf.WriteString(fmt.Sprintf("**Next steps**: Continue research (%d topics remaining)\n\n", unresolvedCount))
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

// generateSummarizingPrompt generates the prompt for the Summarizing state.
// Focus: Synthesize findings into comprehensive summary document(s).
func generateSummarizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Summarizing Findings\n\n")
	buf.WriteString("All research topics are resolved. Create comprehensive summary document(s) synthesizing findings.\n\n")

	// Research completed summary
	phase, exists := p.Phases["exploration"]
	if !exists {
		return "Error: exploration phase not found"
	}

	completed := []projschema.TaskState{}
	abandoned := []projschema.TaskState{}

	for _, task := range phase.Tasks {
		switch task.Status {
		case "completed":
			completed = append(completed, task)
		case "abandoned":
			abandoned = append(abandoned, task)
		}
	}

	buf.WriteString(fmt.Sprintf("### Completed Topics: %d\n\n", len(completed)))
	for _, task := range completed {
		buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
	}
	buf.WriteString("\n")

	if len(abandoned) > 0 {
		buf.WriteString(fmt.Sprintf("### Abandoned Topics: %d\n\n", len(abandoned)))
		for _, task := range abandoned {
			buf.WriteString(fmt.Sprintf("- %s\n", task.Name))
		}
		buf.WriteString("\n")
	}

	// Summary artifacts status
	summaries := []projschema.ArtifactState{}
	for _, artifact := range phase.Outputs {
		if artifact.Type == "summary" {
			summaries = append(summaries, artifact)
		}
	}

	buf.WriteString("### Summary Artifacts\n\n")
	writeSummaryArtifactsSection(&buf, summaries, p)

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/summarizing.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}

// writeSummaryArtifactsSection writes the summary artifacts section to the buffer.
func writeSummaryArtifactsSection(buf *strings.Builder, summaries []projschema.ArtifactState, p *state.Project) {
	if len(summaries) == 0 {
		buf.WriteString("No summaries created yet.\n\n")
		buf.WriteString("**Next steps**: Create summary document(s) synthesizing findings.\n\n")
		return
	}

	approvedCount := 0
	for _, s := range summaries {
		if s.Approved {
			approvedCount++
		}
	}

	fmt.Fprintf(buf, "Total: %d | Approved: %d\n\n", len(summaries), approvedCount)

	for _, s := range summaries {
		status := "Pending approval"
		if s.Approved {
			status = "✓ Approved"
		}
		fmt.Fprintf(buf, "- %s (%s)\n", s.Path, status)
	}
	buf.WriteString("\n")

	// Advancement readiness
	if allSummariesApproved(p) {
		buf.WriteString("✓ All summaries approved!\n\n")
		buf.WriteString("Ready to finalize. Run: `sow project advance`\n\n")
	} else {
		unapprovedCount := countUnapprovedSummaries(p)
		fmt.Fprintf(buf, "**Next steps**: Review and approve %d summary document(s)\n\n", unapprovedCount)
	}
}

// generateFinalizingPrompt generates the prompt for the Finalizing state.
// Focus: Move artifacts to permanent location, create PR, cleanup.
func generateFinalizingPrompt(p *state.Project) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Exploration: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n\n", p.Branch))

	buf.WriteString("## Current State: Finalizing\n\n")
	buf.WriteString("Summary approved. Finalizing exploration by moving artifacts, creating PR, and cleaning up.\n\n")

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
		buf.WriteString("Ready to complete exploration. Run: `sow project advance`\n\n")
	}

	// Render additional guidance from template
	guidance, err := templates.Render(templatesFS, "templates/finalizing.md", p)
	if err != nil {
		return buf.String() + fmt.Sprintf("\nError rendering template: %v", err)
	}
	buf.WriteString(guidance)

	return buf.String()
}
