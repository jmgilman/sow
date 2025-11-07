package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project"
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

			// Get current state
			currentState := state.State(project.Statechart.Current_state)

			// Auto-determination mode: no flags, no event argument
			return executeAutoTransition(project, currentState)
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

// executeAutoTransition performs automatic event determination and transition.
// This is the backward-compatible default mode when no event is specified.
//
// Returns error if:
// - DetermineEvent fails (terminal state, intent-based branching)
// - Transition fails (guard blocked, invalid event)
// - Save fails (I/O error)
func executeAutoTransition(
	proj *state.Project,
	currentState state.State,
) error {
	fmt.Printf("Current state: %s\n", currentState)

	// Build state machine for current state
	machine := proj.Machine()

	// Determine which event to fire from current state
	event, err := proj.Config().DetermineEvent(proj)
	if err != nil {
		// Enhanced error handling
		return enhanceAutoTransitionError(err, proj, currentState)
	}

	// Fire the event with automatic phase status updates
	if err := proj.Config().FireWithPhaseUpdates(machine, event, proj); err != nil {
		// Provide helpful error messages based on error type
		if strings.Contains(err.Error(), "cannot fire event") {
			return fmt.Errorf("transition blocked: %w\n\nCheck that all prerequisites for this state transition are met", err)
		}
		return fmt.Errorf("failed to advance: %w", err)
	}

	// Save updated state
	if err := proj.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Display new state
	newState := proj.Statechart.Current_state
	fmt.Printf("Advanced to: %s\n", newState)

	return nil
}

// enhanceAutoTransitionError provides helpful error messages when auto-determination fails.
// Distinguishes between terminal states and intent-based branching scenarios.
func enhanceAutoTransitionError(err error, proj *state.Project, currentState state.State) error {
	// Type assert to get access to introspection methods
	// The config is always *project.ProjectTypeConfig which has GetAvailableTransitions
	config, ok := proj.Config().(*project.ProjectTypeConfig)
	if !ok {
		// Fallback to basic error if type assertion fails (shouldn't happen in practice)
		return fmt.Errorf("cannot advance from state %s: %w", currentState, err)
	}

	// Check if this is a terminal state (no transitions configured)
	transitions := config.GetAvailableTransitions(currentState)
	if len(transitions) == 0 {
		return fmt.Errorf(
			"cannot advance from state %s: %w\n\nThis may be a terminal state",
			currentState,
			err,
		)
	}

	// Intent-based branching case (multiple transitions, no discriminator)
	if len(transitions) > 1 {
		// Extract event names
		events := make([]string, len(transitions))
		for i, t := range transitions {
			events[i] = string(t.Event)
		}

		return fmt.Errorf(
			"cannot advance from state %s: %w\n\n"+
				"Use 'sow advance --list' to see available transitions and select one explicitly.\n"+
				"Available events: %s",
			currentState,
			err,
			strings.Join(events, ", "),
		)
	}

	// Default error wrapping
	return fmt.Errorf("cannot advance from state %s: %w", currentState, err)
}
