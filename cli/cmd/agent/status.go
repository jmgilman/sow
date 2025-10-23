package agent

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
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

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Get state
			state := project.State()

			// Determine active phase
			activePhase, phaseStatus := projectpkg.DetermineActivePhase(state)

			// Build output
			var output strings.Builder
			output.WriteString(fmt.Sprintf("\n━━━ Project: %s ━━━\n", state.Project.Name))
			output.WriteString(fmt.Sprintf("Branch: %s\n", state.Project.Branch))
			output.WriteString(fmt.Sprintf("Description: %s\n\n", state.Project.Description))

			// Active phase
			if activePhase == "unknown" { //nolint:nestif // Complex logic required for comprehensive status output
				output.WriteString("Status: All phases complete\n")
			} else {
				output.WriteString(fmt.Sprintf("Active Phase: %s (%s)\n", activePhase, phaseStatus))

				// If implementation, show task summary
				if activePhase == "implementation" {
					tasks := project.ListTasks()
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

					output.WriteString(fmt.Sprintf("\nTasks: %d total\n", len(tasks)))
					if completed > 0 {
						output.WriteString(fmt.Sprintf("  ✓ %d completed\n", completed))
					}
					if inProgress > 0 {
						output.WriteString(fmt.Sprintf("  ⟳ %d in progress\n", inProgress))
					}
					if pending > 0 {
						output.WriteString(fmt.Sprintf("  ○ %d pending\n", pending))
					}
					if abandoned > 0 {
						output.WriteString(fmt.Sprintf("  ✗ %d abandoned\n", abandoned))
					}
				}

				// Next actions
				output.WriteString("\nNext Actions:\n")
				output.WriteString(getNextActions(activePhase, phaseStatus))
			}

			cmd.Print(output.String())
			return nil
		},
	}

	return cmd
}

// getNextActions returns suggested next actions based on the current phase and status.
func getNextActions(phase, status string) string {
	switch phase {
	case "discovery":
		if status == "pending" {
			return "  • sow agent enable discovery --type <type>\n  • sow agent skip discovery\n"
		}
		return "  • sow agent artifact add <path>\n  • sow agent complete\n"

	case "design":
		if status == "pending" {
			return "  • sow agent enable design\n  • sow agent skip design\n"
		}
		return "  • sow agent artifact add <path>\n  • sow agent complete\n"

	case "implementation":
		if status == "pending" {
			return "  • sow agent task add <name>\n"
		}
		return "  • sow agent task update <id> --status <status>\n  • sow agent complete\n"

	case "review":
		return "  • Add review reports\n  • sow agent complete\n"

	case "finalize":
		return "  • sow agent complete\n  • sow agent project delete\n"

	default:
		return "  • Unknown phase\n"
	}
}
