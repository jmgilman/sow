package project

import (
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// newWizardCmd creates the interactive wizard command.
//
//nolint:unused // Foundation code - wizard now invoked directly via RunE
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
	// When -- is used, all args after it are passed through in args parameter
	var claudeFlags []string
	if cmd.ArgsLenAtDash() >= 0 {
		// Everything in args after the dash index is a Claude flag
		claudeFlags = args[cmd.ArgsLenAtDash():]
	}

	wizard := NewWizard(cmd, mainCtx, claudeFlags)
	return wizard.Run()
}
