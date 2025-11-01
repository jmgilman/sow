// Package agent provides commands primarily used by AI agents during orchestration.
package agent

import (
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
  - task: Task management

Project lifecycle:
  - init: Initialize a new project
  - delete: Delete the project directory
  - create-pr: Create a pull request

Phase management (work on implicit active phase):
  - complete: Complete the active phase
  - skip: Skip the active phase
  - enable: Enable a specific phase
  - status: Show current project status
  - info: Show phase information
  - advance: Advance to next state within current phase

Artifact management:
  - artifact: Manage artifacts (add, approve, list)

Utilities:
  - set: Set custom fields on active phase`,
	}

	// Add subcommands
	cmd.AddCommand(NewLogCmd())
	cmd.AddCommand(NewSessionInfoCmd())

	// Project lifecycle
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewCreatePRCmd())

	// Phase management (implicit active phase)
	cmd.AddCommand(NewCompleteCmd())
	cmd.AddCommand(NewSkipCmd())
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewInfoCmd())
	cmd.AddCommand(NewAdvanceCmd())

	// Artifact management
	cmd.AddCommand(NewArtifactCmd())

	// Task management
	cmd.AddCommand(task.NewTaskCmd())

	// Utilities
	cmd.AddCommand(NewSetCmd())

	return cmd
}
