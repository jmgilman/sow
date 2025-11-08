package cmd

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	"github.com/jmgilman/sow/cli/internal/sow"
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

			// Check for list mode
			listFlag, _ := cmd.Flags().GetBool("list")
			if listFlag {
				return listAvailableTransitions(project, currentState)
			}

			// Check for dry-run mode
			dryRunFlag, _ := cmd.Flags().GetBool("dry-run")
			if dryRunFlag {
				// Get event argument (validation ensures it exists)
				event := args[0]
				machine := project.Machine()
				return validateTransition(ctx, project, machine, currentState, sdkstate.Event(event))
			}

			// Get event argument if provided
			var event string
			if len(args) > 0 {
				event = args[0]
			}

			// Explicit event mode: event argument provided without flags
			if event != "" {
				machine := project.Machine()
				return executeExplicitTransition(ctx, project, machine, currentState, sdkstate.Event(event))
			}

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
// - Save fails (I/O error).
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

// listAvailableTransitions displays all available transitions from the current state.
// Shows both permitted and blocked transitions with guard status.
func listAvailableTransitions(
	proj *state.Project,
	currentState state.State,
) error {
	fmt.Printf("Current state: %s\n\n", currentState)

	// Type assert to get access to introspection methods
	config, ok := proj.Config().(*project.ProjectTypeConfig)
	if !ok {
		return fmt.Errorf("cannot list transitions: invalid project configuration")
	}

	// Get all configured transitions
	allTransitions := config.GetAvailableTransitions(currentState)
	if len(allTransitions) == 0 {
		fmt.Println("No transitions available from current state.")
		fmt.Println("This may be a terminal state.")
		return nil
	}

	// Get guard-filtered events (what can fire now)
	machine := proj.Machine()
	permittedEvents, err := machine.PermittedTriggers()
	if err != nil {
		return fmt.Errorf("failed to get permitted triggers: %w", err)
	}

	// Build set of permitted events for quick lookup
	permitted := make(map[state.Event]bool)
	for _, event := range permittedEvents {
		permitted[event] = true
	}

	// Display transitions
	fmt.Println("Available transitions:")

	if len(permittedEvents) == 0 {
		fmt.Println()
		fmt.Println("(All configured transitions are currently blocked by guard conditions)")
		fmt.Println()
	} else {
		fmt.Println()
	}

	for _, transition := range allTransitions {
		// Check if permitted
		blocked := !permitted[transition.Event]
		blockedMarker := ""
		if blocked {
			blockedMarker = "  [BLOCKED]"
		}

		// Display event and target
		fmt.Printf("  sow advance %s%s\n", transition.Event, blockedMarker)
		fmt.Printf("    → %s\n", transition.To)

		// Display description if present
		if transition.Description != "" {
			fmt.Printf("    %s\n", transition.Description)
		}

		// Display guard description if present
		if transition.GuardDesc != "" {
			fmt.Printf("    Requires: %s\n", transition.GuardDesc)
		}

		fmt.Println()
	}

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

// validateTransition validates whether a transition can be executed without executing it.
// This is the core implementation of dry-run mode.
//
// Returns nil if the transition is valid and would succeed.
// Returns an error if the transition is invalid or blocked.
//
// Side effects: NONE - this function never modifies project state.
func validateTransition(
	_ *sow.Context,
	proj *state.Project,
	machine *sdkstate.Machine,
	currentState sdkstate.State,
	event sdkstate.Event,
) error {
	fmt.Printf("Validating transition: %s -> %s\n\n", currentState, event)

	// Type assert to get access to introspection methods
	projectConfig, ok := proj.Config().(*project.ProjectTypeConfig)
	if !ok {
		return fmt.Errorf("cannot validate transition: invalid project configuration")
	}

	// Check if event is configured for this state
	targetState := projectConfig.GetTargetState(currentState, event)
	if targetState == "" {
		fmt.Printf("✗ Event '%s' is not configured for state %s\n\n", event, currentState)
		fmt.Println("Use 'sow advance --list' to see available transitions.")
		return fmt.Errorf("event not configured")
	}

	// Check if event can fire (guard passes)
	canFire, err := machine.CanFire(event)
	if err != nil {
		return fmt.Errorf("failed to validate transition: %w", err)
	}

	if !canFire {
		// Blocked by guard
		fmt.Println("✗ Transition blocked by guard condition")
		fmt.Println()

		guardDesc := projectConfig.GetGuardDescription(currentState, event)
		if guardDesc != "" {
			fmt.Printf("Guard description: %s\n", guardDesc)
		}
		fmt.Println("Current status: Guard not satisfied")
		fmt.Println()
		fmt.Println("Fix the guard condition, then try again.")

		return fmt.Errorf("transition blocked by guard")
	}

	// Valid - would succeed
	fmt.Println("✓ Transition is valid and can be executed")
	fmt.Println()
	fmt.Printf("Target state: %s\n", targetState)

	description := projectConfig.GetTransitionDescription(currentState, event)
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}

	fmt.Println()
	fmt.Printf("To execute: sow advance %s\n", event)

	return nil
}

// executeExplicitTransition validates and executes an explicit event transition.
// This is the core implementation of explicit event mode (sow advance [event]).
//
// The orchestrator specifies exactly which event to fire, which is used for:
// - Intent-based branching (multiple valid options, orchestrator decides)
// - Explicit state control (override auto-determination)
// - Debugging and testing (fire specific events)
//
// Returns error if:
// - Event not configured for current state
// - Guard conditions not met
// - Transition execution fails
// - Save fails
//
// Side effects:
// - Fires event and transitions state machine
// - Updates phase status based on transition
// - Saves project state to disk.
func executeExplicitTransition(
	ctx *sow.Context,
	proj *state.Project,
	machine *sdkstate.Machine,
	currentState sdkstate.State,
	event sdkstate.Event,
) error {
	fmt.Printf("Current state: %s\n", currentState)

	// Type assert to get access to introspection methods
	config, ok := proj.Config().(*project.ProjectTypeConfig)
	if !ok {
		return fmt.Errorf("cannot execute transition: invalid project configuration")
	}

	// Validate event is configured for this state
	targetState := config.GetTargetState(currentState, event)
	if targetState == "" {
		fmt.Printf("\nError: Event '%s' is not configured for state %s\n\n", event, currentState)
		fmt.Println("Use 'sow advance --list' to see available transitions.")
		return fmt.Errorf("event not configured")
	}

	// Fire the event (this validates guards and executes transition)
	err := config.FireWithPhaseUpdates(machine, event, proj)
	if err != nil {
		// Enhanced error handling for guard failures
		return enhanceTransitionError(err, currentState, event, targetState, config)
	}

	// Sync machine state to project state (Save() does this, but we need it before Save())
	// This is required because the machine's internal state has changed
	proj.Statechart.Current_state = machine.State().String()

	// Save updated state to disk
	// Note: In production, ctx is always set via Load()
	// In unit tests, proj may not have ctx, so we skip Save()
	if ctx != nil {
		if err := proj.Save(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	// Display new state
	newState := proj.Statechart.Current_state
	fmt.Printf("Advanced to: %s\n", newState)

	return nil
}

// enhanceTransitionError enhances error messages for transition failures.
// Detects guard failures and provides helpful context and next steps.
func enhanceTransitionError(
	err error,
	currentState sdkstate.State,
	event sdkstate.Event,
	targetState sdkstate.State,
	config *project.ProjectTypeConfig,
) error {
	// Check if this is a guard failure
	if strings.Contains(err.Error(), "guard condition is not met") {
		guardDesc := config.GetGuardDescription(currentState, event)

		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("Transition blocked: %s\n\n", guardDesc))
		msg.WriteString(fmt.Sprintf("Current state: %s\n", currentState))
		msg.WriteString(fmt.Sprintf("Event: %s\n", event))
		msg.WriteString(fmt.Sprintf("Target state: %s\n\n", targetState))

		if guardDesc != "" {
			msg.WriteString(fmt.Sprintf("The guard condition is not satisfied. %s\n\n", guardDesc))
		}

		msg.WriteString(fmt.Sprintf("Use 'sow advance --dry-run %s' to validate prerequisites.", event))

		return fmt.Errorf("%s", msg.String())
	}

	// Other error types - use default wrapping
	return fmt.Errorf("failed to advance: %w", err)
}
