package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch Claude Code in orchestrator mode",
		Long: `Start Claude Code with sow orchestrator.

The orchestrator will greet you and present options based on your
repository state. You can then choose to continue existing work or
start something new.

This command runs 'claude /sow-greet', which automatically detects:
- Whether sow is initialized
- Whether an active project exists
- Current project phase and task status

Claude will greet you with context-aware information and let you
choose what to do next.

Examples:
  sow start    Launch orchestrator with greeting`,
		RunE: runStart,
	}

	return cmd
}

func runStart(cmd *cobra.Command, _ []string) error {
	s := cmdutil.SowFromContext(cmd.Context())

	// Validate sow initialized
	if !s.IsInitialized() {
		fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
		fmt.Fprintln(os.Stderr, "Run: sow init")
		return fmt.Errorf("not initialized")
	}

	// Check claude CLI available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Launch claude with /sow-greet slash command
	claudeCmd := exec.CommandContext(cmd.Context(), claudePath, "/sow-greet")
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = s.RepoRoot()

	// Run and wait for completion
	return claudeCmd.Run()
}
