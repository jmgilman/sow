package agent

import (
	"sort"

	"github.com/jmgilman/sow/cli/internal/agents"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available agents",
		Long: `List all available agents in the sow multi-agent system.

Agents are roles that can be spawned by the orchestrator to perform
specialized tasks. Each agent has specific capabilities and prompts
tailored to its role.

The list shows each agent's name and description.`,
		RunE: runList,
	}
}

func runList(cmd *cobra.Command, _ []string) error {
	registry := agents.NewAgentRegistry()
	agentList := registry.List()

	// Sort agents alphabetically by name
	sort.Slice(agentList, func(i, j int) bool {
		return agentList[i].Name < agentList[j].Name
	})

	cmd.Println("Available agents:")
	for _, agent := range agentList {
		cmd.Printf("  %-14s%s\n", agent.Name, agent.Description)
	}

	return nil
}
