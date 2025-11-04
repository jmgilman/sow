package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/spf13/cobra"
)

// NewAdvanceCmd creates the top-level command to advance project state.
//
// Usage:
//
//	sow advance
//
// This command examines the current phase and state, determines the appropriate
// transition event, validates prerequisites via guards, and advances the state
// machine if all conditions are met.
func NewAdvanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance",
		Short: "Progress project to next state",
		Long: `Progress the project through its state machine.

The advance command:
1. Determines the next event based on current state
2. Evaluates guards to ensure transition is allowed
3. Fires the event if guards pass
4. Saves the updated state

Guards may prevent transitions. Common guard failures:
- Planning → Implementation: task_list output not approved
- Implementation Planning → Executing: tasks not approved (metadata.tasks_approved)
- Implementation Executing → Review: not all tasks completed
- Review → Finalize: review not approved or assessment not set

Example:
  sow advance`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project using SDK
			project, err := state.Load(ctx)
			if err != nil {
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current state for display
			currentState := project.Statechart.Current_state
			fmt.Printf("Current state: %s\n", currentState)

			// Advance (calls OnAdvance determiner, evaluates guards, fires event)
			if err := project.Advance(); err != nil {
				// Provide helpful error messages based on error type
				if strings.Contains(err.Error(), "cannot fire event") {
					return fmt.Errorf("transition blocked: %w\n\nCheck that all prerequisites for this state transition are met", err)
				}
				if strings.Contains(err.Error(), "no event determiner") {
					return fmt.Errorf("cannot advance from state %s: no transition configured\n\nThis may be a terminal state", currentState)
				}
				return fmt.Errorf("failed to advance: %w", err)
			}

			// Save updated state
			if err := project.Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			// Display new state
			newState := project.Statechart.Current_state
			fmt.Printf("Advanced to: %s\n", newState)

			return nil
		},
	}

	return cmd
}
