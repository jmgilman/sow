package cmd

import (
	"fmt"
	"os"
	"os/exec"

	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// launchClaudeCode launches Claude Code with the given prompt and optional additional flags.
func launchClaudeCode(cmd *cobra.Command, ctx *sow.Context, prompt string, claudeFlags []string) error {
	claude := sowexec.NewLocal("claude")
	if !claude.Exists() {
		fmt.Fprintln(os.Stderr, "Error: Claude Code CLI not found")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.com/download")
		return fmt.Errorf("claude not found")
	}

	// Build command args: prompt first, then any additional flags
	args := []string{prompt}
	args = append(args, claudeFlags...)

	claudeCmd := exec.CommandContext(cmd.Context(), claude.Command(), args...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr
	claudeCmd.Dir = ctx.RepoRoot()

	return claudeCmd.Run()
}
