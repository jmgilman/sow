package agent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the command to show current project status.
//
// Usage:
//
//	sow agent status
//
// Shows the current active phase, status, and task summary in a concise format.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current project status",
		Long: `Show current project status including active phase and task summary.

Displays:
  - Project name and branch
  - Current active phase and its status
  - Task summary (if in implementation phase)
  - Next recommended actions

Example:
  sow agent status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Build output
			var output strings.Builder
			output.WriteString(fmt.Sprintf("\n━━━ Project: %s ━━━\n", proj.Name()))
			output.WriteString(fmt.Sprintf("Branch: %s\n", proj.Branch()))
			output.WriteString(fmt.Sprintf("Description: %s\n\n", proj.Description()))

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				output.WriteString("Status: All phases complete\n")
			} else {
				writePhaseStatus(&output, phase)
			}

			cmd.Print(output.String())
			return nil
		},
	}

	return cmd
}

// writePhaseStatus writes the current phase status and task summary to the output.
func writePhaseStatus(output *strings.Builder, phase domain.Phase) {
	fmt.Fprintf(output, "Active Phase: %s (%s)\n", phase.Name(), phase.Status())

	// Try to show task summary (only works for implementation phase)
	tasks := phase.ListTasks()
	if len(tasks) > 0 {
		writeTaskSummary(output, tasks)
	}

	// Next actions
	output.WriteString("\nNext Actions:\n")
	output.WriteString(getNextActions(phase.Name(), phase.Status()))
}

// writeTaskSummary writes the task summary to the output.
func writeTaskSummary(output *strings.Builder, tasks []*domain.Task) {
	completed := 0
	inProgress := 0
	pending := 0
	abandoned := 0

	for _, task := range tasks {
		taskState, err := task.State()
		if err != nil {
			continue // Skip tasks with errors
		}
		switch taskState.Task.Status {
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

	fmt.Fprintf(output, "\nTasks: %d total\n", len(tasks))
	if completed > 0 {
		fmt.Fprintf(output, "  ✓ %d completed\n", completed)
	}
	if inProgress > 0 {
		fmt.Fprintf(output, "  ⟳ %d in progress\n", inProgress)
	}
	if pending > 0 {
		fmt.Fprintf(output, "  ○ %d pending\n", pending)
	}
	if abandoned > 0 {
		fmt.Fprintf(output, "  ✗ %d abandoned\n", abandoned)
	}
}

// getNextActions returns suggested next actions based on the current phase and status.
func getNextActions(phase, status string) string {
	switch phase {
	case "discovery":
		return "  • sow agent artifact add <path> --metadata type=<type>\n  • sow agent complete\n"

	case "design":
		return "  • sow agent artifact add <path> --metadata type=design\n  • sow agent complete\n"

	case "implementation":
		if status == "pending" {
			return "  • sow agent task add <name> --description <desc>\n  • sow agent task approve\n"
		}
		return "  • sow agent task update <id> --status <status>\n  • sow agent complete\n"

	case "review":
		return "  • sow agent artifact add <path> --metadata type=review --metadata assessment=<pass|fail>\n  • sow agent complete\n"

	case "finalize":
		return "  • sow agent complete\n  • sow agent delete\n"

	default:
		return "  • Unknown phase\n"
	}
}
