package cmd

import (
	"github.com/spf13/cobra"
)

// NewSessionInfoCmd creates the session-info command
func NewSessionInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session-info",
		Short: "Display session context information",
		Long: `Display current session context information.

Shows:
  - Repository root path
  - Current context (task or project)
  - Task details (if in task context)
  - Project details (if project exists)
  - CLI version
  - Schema version

Output is JSON for easy consumption by agents and tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := FilesystemFromContext(cmd.Context())
			_ = fs // TODO: implement

			cmd.Println(`{
  "context": "project",
  "repo_root": "/path/to/repo",
  "cli_version": "dev"
}`)
			return nil
		},
	}

	return cmd
}
