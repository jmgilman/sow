// Package agent implements commands for managing AI agents.
package agent

import (
	"github.com/spf13/cobra"
)

// NewAgentCmd creates the agent command with all subcommands.
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage AI agents",
		Long: `Manage AI agents for the sow multi-agent system.

Agents are roles (implementer, architect, reviewer, etc.) that can be
spawned by the orchestrator to perform specialized tasks. Each agent
has specific capabilities and prompts tailored to its role.

Commands:
  list      List available agents
  spawn     Spawn an agent to execute a task
  resume    Resume a paused agent session`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSpawnCmd())
	cmd.AddCommand(newResumeCmd())

	return cmd
}
