package agent

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewSetCmd creates the command to set a custom field on the active phase.
//
// Usage:
//
//	sow agent set <field> <value>
//
// Sets a field in the phase metadata. Works for all phases and project types.
func NewSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <field> <value>",
		Short: "Set a field on the active phase",
		Long: `Set a field in the active phase's metadata.

This command stores the field in the phase's generic metadata map.
The value is automatically parsed as:
  - Integer if it's a valid number
  - Boolean if it's "true" or "false"
  - String otherwise

Examples:
  # Set discovery type
  sow agent set discovery_type feature

  # Set review iteration
  sow agent set iteration 2

  # Set boolean flag
  sow agent set tasks_approved true

  # Set string value
  sow agent set pr_url https://github.com/org/repo/pull/123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			field := args[0]
			valueStr := args[1]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found")
			}

			// Try to parse value as int, bool, or keep as string
			var parsedValue interface{} = valueStr
			if intVal, err := strconv.Atoi(valueStr); err == nil {
				parsedValue = intVal
			} else if boolVal, err := strconv.ParseBool(valueStr); err == nil {
				parsedValue = boolVal
			}

			// Set field via Phase interface
			result, err := phase.Set(field, parsedValue)
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("field %s not supported by phase %s", field, phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to set field: %w", err)
			}

			// Fire event if phase returned one
			if result.Event != "" {
				machine := proj.Machine()
				if err := machine.Fire(result.Event); err != nil {
					return fmt.Errorf("failed to fire event %s: %w", result.Event, err)
				}
				// Save after transition
				if err := proj.Save(); err != nil {
					return fmt.Errorf("failed to save project state: %w", err)
				}
			}

			cmd.Printf("\n✓ Set %s = %v on %s phase\n", field, parsedValue, phase.Name())
			return nil
		},
	}

	return cmd
}
