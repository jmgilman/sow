package statechart

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
)

// PromptComponents provides reusable prompt building blocks that can be composed
// into complete prompts. Components can access external systems (git, GitHub)
// via sow.Context and render templates via the prompts package.
type PromptComponents struct {
	ctx *sow.Context
}

// NewPromptComponents creates a new PromptComponents instance with access
// to external systems via the provided context.
func NewPromptComponents(ctx *sow.Context) *PromptComponents {
	return &PromptComponents{
		ctx: ctx,
	}
}

// ProjectHeader generates a header section with project metadata.
// This typically includes project name, description, and branch.
//
// Example output:
//
//	# Project: add-auth
//	Branch: feat/auth-system
//	Description: Implement JWT authentication
func (c *PromptComponents) ProjectHeader(projectState *schemas.ProjectState) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("# Project: %s\n", projectState.Project.Name))
	buf.WriteString(fmt.Sprintf("Branch: %s\n", projectState.Project.Branch))
	if projectState.Project.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", projectState.Project.Description))
	}

	return buf.String()
}

// GitStatus generates a section showing the current git status.
// Returns the output formatted for display showing uncommitted changes.
// Returns an error if git operations fail.
//
// Example output:
//
//	## Git Status
//	Working tree is clean.
//
// Or:
//
//	## Git Status
//	Has uncommitted changes.
func (c *PromptComponents) GitStatus() (string, error) {
	git := c.ctx.Git()

	// Check if there are uncommitted changes
	hasChanges, err := git.HasUncommittedChanges()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("## Git Status\n\n")

	if hasChanges {
		buf.WriteString("Has uncommitted changes.\n")
	} else {
		buf.WriteString("Working tree is clean.\n")
	}

	return buf.String(), nil
}

// RecentCommits generates a section showing recent commit history.
// This is a placeholder that will be implemented when commit log functionality is needed.
// Returns an error if git operations fail.
//
// Example output:
//
//	## Recent Commits
//	(Recent commits will be shown here)
func (c *PromptComponents) RecentCommits(_ int) (string, error) {
	var buf strings.Builder
	buf.WriteString("## Recent Commits\n\n")
	buf.WriteString("(Recent commits will be shown here)\n")
	return buf.String(), nil
}

// TaskSummary generates a summary of tasks with their status breakdown.
// Shows total, completed, in-progress, and pending task counts.
//
// Example output:
//
//	## Tasks (3 total)
//	- 1 completed
//	- 1 in progress
//	- 1 pending
func (c *PromptComponents) TaskSummary(tasks []phases.Task) string {
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

// OpenIssues generates a section listing open GitHub issues.
// Returns an error if GitHub operations fail or gh CLI is unavailable.
//
// Example output:
//
//	## Open Issues
//	#42 Add JWT authentication
//	#43 Update documentation
func (c *PromptComponents) OpenIssues() (string, error) {
	gh := c.ctx.GitHub()

	// Check if GitHub CLI is available
	if err := gh.CheckInstalled(); err != nil {
		return "", fmt.Errorf("github CLI not available: %w", err)
	}

	// List open issues with "sow" label
	issues, err := gh.ListIssues("sow", "open")
	if err != nil {
		return "", fmt.Errorf("failed to list issues: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("## Open Issues\n\n")

	// Limit to first 10 issues
	displayCount := len(issues)
	if displayCount > 10 {
		displayCount = 10
	}

	if len(issues) == 0 {
		buf.WriteString("No open issues.\n")
	} else {
		for i := 0; i < displayCount; i++ {
			issue := issues[i]
			buf.WriteString(fmt.Sprintf("#%d %s\n", issue.Number, issue.Title))
		}
	}

	return buf.String(), nil
}

// RenderTemplate renders a prompt template with the given context.
// This delegates to the prompts package for template rendering.
// Returns an error if the template ID is unknown or rendering fails.
//
// Example:
//
//	ctx := &prompts.StatechartContext{
//	    State:        "PlanningActive",
//	    ProjectState: projectState,
//	}
//	content, err := components.RenderTemplate(prompts.PromptPlanningActive, ctx)
func (c *PromptComponents) RenderTemplate(
	templateID prompts.PromptID,
	ctx prompts.Context,
) (string, error) {
	output, err := prompts.Render(templateID, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateID, err)
	}
	return output, nil
}
