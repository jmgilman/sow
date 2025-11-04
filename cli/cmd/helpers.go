package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	sowexec "github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
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

// resolveTaskPhase determines which phase to use for task operations.
// It follows this priority:
//  1. Explicit --phase flag value (if provided)
//  2. Smart default based on project state
//  3. Error if no phase supports tasks
//
// Returns the resolved phase name or an error with helpful guidance.
func resolveTaskPhase(project *state.Project, explicitPhase string) (string, error) {
	// Get the project type config to query task support
	config := project.Config()

	// Case 1: Explicit phase provided via --phase flag
	if explicitPhase != "" {
		// Validate that the phase supports tasks
		if !config.PhaseSupportsTasks(explicitPhase) {
			supportedPhases := config.GetTaskSupportingPhases()
			if len(supportedPhases) == 0 {
				return "", fmt.Errorf("phase %s does not support tasks (no phases support tasks in this project type)", explicitPhase)
			}
			return "", fmt.Errorf("phase %s does not support tasks (supported phases: %s)",
				explicitPhase, strings.Join(supportedPhases, ", "))
		}
		return explicitPhase, nil
	}

	// Case 2: Smart default based on current project state
	currentState := state.State(project.Statechart.Current_state)
	defaultPhase := config.GetDefaultTaskPhase(currentState)

	if defaultPhase == "" {
		// No smart default found - provide helpful error
		supportedPhases := config.GetTaskSupportingPhases()
		if len(supportedPhases) == 0 {
			return "", fmt.Errorf("no phases in this project type support tasks")
		}
		return "", fmt.Errorf("could not determine default task phase for state %s\nSupported phases: %s\nSpecify phase with --phase flag",
			currentState, strings.Join(supportedPhases, ", "))
	}

	return defaultPhase, nil
}
