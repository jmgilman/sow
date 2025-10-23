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
  - enable: Enable an optional phase
  - skip: Skip an optional phase
  - complete: Complete the active phase
  - status: Show current project status
  - info: Show phase information

Artifact management:
  - artifact: Manage artifacts (add, approve, list)

Review management:
  - review: Manage review phase (add, approve, increment)

Finalize management:
  - finalize: Manage finalize phase (complete, doc, move)

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
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewSkipCmd())
	cmd.AddCommand(NewCompleteCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewInfoCmd())

	// Artifact management
	cmd.AddCommand(NewArtifactCmd())

	// Review management
	cmd.AddCommand(NewReviewCmd())

	// Finalize management
	cmd.AddCommand(NewFinalizeCmd())

	// Task management
	cmd.AddCommand(task.NewTaskCmd())

	// Utilities
	cmd.AddCommand(NewSetCmd())

	return cmd
}
