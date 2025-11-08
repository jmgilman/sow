package project

import (
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newWizardCmd creates the interactive wizard command.
func newWizardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Create or continue a project (interactive)",
		Long: `Launch the interactive project wizard to create or continue a project.

The wizard will guide you through:
- Creating a new project from an issue or branch
- Continuing an existing project
- Configuring project settings

Claude Code Flags:
  Use -- to pass additional flags to the Claude Code CLI:
    sow project wizard -- --model opus --verbose`,
		Args: cobra.NoArgs,
		RunE: runWizard,
	}

	return cmd
}

// runWizard executes the interactive wizard.
func runWizard(cmd *cobra.Command, args []string) error {
	mainCtx := cmdutil.GetContext(cmd.Context())

	if !mainCtx.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Extract Claude Code flags (everything after --)
	var claudeFlags []string
	if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
		allArgs := cmd.Flags().Args()
		if dashIndex < len(allArgs) {
			claudeFlags = allArgs[dashIndex:]
		}
	}

	wizard := NewWizard(mainCtx, claudeFlags)
	return wizard.Run()
}
