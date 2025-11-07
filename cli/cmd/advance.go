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
//	sow advance                    # Auto-determine next event
//	sow advance [event]            # Fire explicit event
//	sow advance --list             # List available transitions
//	sow advance --dry-run [event]  # Validate without executing
//
// This command examines the current phase and state, determines the appropriate
// transition event, validates prerequisites via guards, and advances the state
// machine if all conditions are met.
func NewAdvanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance [event]",
		Short: "Progress project to next state",
		Long: `Progress the project through its state machine.

The advance command:
1. Determines the next event based on current state (auto mode)
2. Or fires an explicitly specified event (explicit mode)
3. Evaluates guards to ensure transition is allowed
4. Fires the event if guards pass
5. Saves the updated state

Flags:
  --list     List available transitions without executing
  --dry-run  Validate transition without executing (requires event argument)

Guards may prevent transitions. Common guard failures:
- Planning → Implementation: task_list output not approved
- Implementation Planning → Executing: tasks not approved (metadata.tasks_approved)
- Implementation Executing → Review: not all tasks completed
- Review → Finalize: review not approved or assessment not set

Examples:
  sow advance                    # Auto-determine next event
  sow advance finalize           # Fire explicit event
  sow advance --list             # Show available transitions
  sow advance --dry-run finalize # Validate before executing`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate flags and arguments
			if err := validateAdvanceFlags(cmd, args); err != nil {
				return err
			}

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

			// Determine which event to fire from current state
			event, err := project.Config().DetermineEvent(project)
			if err != nil {
				return fmt.Errorf("cannot advance from state %s: %w\n\nThis may be a terminal state", currentState, err)
			}

			// Fire the event with automatic phase status updates
			// (evaluates guards, executes actions, transitions state, updates phase status)
			if err := project.Config().FireWithPhaseUpdates(project.Machine(), event, project); err != nil {
				// Provide helpful error messages based on error type
				if strings.Contains(err.Error(), "cannot fire event") {
					return fmt.Errorf("transition blocked: %w\n\nCheck that all prerequisites for this state transition are met", err)
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

	// Add flags
	cmd.Flags().Bool("list", false, "List available transitions without executing")
	cmd.Flags().Bool("dry-run", false, "Validate transition without executing")

	return cmd
}

// validateAdvanceFlags checks mutual exclusivity rules for flags and arguments.
// Returns an error if invalid flag/argument combinations are detected.
func validateAdvanceFlags(cmd *cobra.Command, args []string) error {
	listFlag, _ := cmd.Flags().GetBool("list")
	dryRunFlag, _ := cmd.Flags().GetBool("dry-run")

	// Get event argument if provided
	var event string
	if len(args) > 0 {
		event = args[0]
	}

	// Validate flag combinations (order matters - check conflicting flags first)
	if listFlag && dryRunFlag {
		return fmt.Errorf("cannot use --list and --dry-run together")
	}

	if listFlag && event != "" {
		return fmt.Errorf("cannot specify event argument with --list flag")
	}

	if dryRunFlag && event == "" {
		return fmt.Errorf("--dry-run requires an event argument")
	}

	return nil
}
