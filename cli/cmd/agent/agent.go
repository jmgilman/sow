// Package agent provides commands primarily used by AI agents during orchestration.
package agent

import (
	"github.com/jmgilman/sow/cli/cmd/agent/project"
	"github.com/jmgilman/sow/cli/cmd/agent/task"
	"github.com/spf13/cobra"
)

// NewAgentCmd creates the root agent command.
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Commands for AI agent orchestration",
		Long: `Commands primarily used by AI agents during project orchestration.

These commands are invoked by the orchestrator and worker agents during
project execution. While humans can use them for debugging or manual
intervention, they are designed for agent workflows.

Agent commands include:
  - log: Fast structured logging for agent actions
  - session-info: Context detection for agents
  - project: Project lifecycle management (legacy)
  - task: Task management

New simplified commands (work on implicit active phase):
  - enable: Enable an optional phase
  - skip: Skip an optional phase
  - complete: Complete the active phase
  - status: Show current project status
  - info: Show phase information
  - artifact: Manage artifacts
  - set: Set custom fields on active phase`,
	}

	// Add subcommands
	cmd.AddCommand(NewLogCmd())
	cmd.AddCommand(NewSessionInfoCmd())

	// New simplified commands
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewSkipCmd())
	cmd.AddCommand(NewCompleteCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewInfoCmd())
	cmd.AddCommand(NewArtifactCmd())
	cmd.AddCommand(NewSetCmd())

	// Legacy commands (to be deprecated)
	cmd.AddCommand(project.NewProjectCmd())
	cmd.AddCommand(task.NewTaskCmd())

	return cmd
}
