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
  - project: Project lifecycle management
  - task: Task management within implementation phase`,
	}

	// Add subcommands
	cmd.AddCommand(NewLogCmd())
	cmd.AddCommand(NewSessionInfoCmd())
	cmd.AddCommand(project.NewProjectCmd())
	cmd.AddCommand(task.NewTaskCmd())

	return cmd
}
