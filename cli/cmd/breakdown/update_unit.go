package breakdown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewUpdateUnitCmd creates the breakdown update-unit command.
func NewUpdateUnitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-unit <id>",
		Short: "Update a work unit's metadata",
		Long: `Update a work unit's title, description, or dependencies.

Only the provided flags will be updated; other fields remain unchanged.

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Work unit must exist in the index

Examples:
  # Update title only
  sow breakdown update-unit unit-001 --title "JWT token service (revised)"

  # Update description only
  sow breakdown update-unit unit-001 \
    --description "Implement JWT generation, validation, and refresh"

  # Update dependencies
  sow breakdown update-unit unit-003 --depends-on "unit-001,unit-002"

  # Update multiple fields
  sow breakdown update-unit unit-001 \
    --title "JWT token service" \
    --description "Complete JWT implementation" \
    --depends-on ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateUnit(cmd, args)
		},
	}

	// Flags
	cmd.Flags().String("title", "", "New title for the work unit")
	cmd.Flags().String("description", "", "New description for the work unit")
	cmd.Flags().StringSlice("depends-on", nil, "New dependency list (empty string to clear)")

	return cmd
}

func runUpdateUnit(cmd *cobra.Command, args []string) error {
	id := args[0]

	// Check which flags were provided
	titleFlag := cmd.Flags().Changed("title")
	descFlag := cmd.Flags().Changed("description")
	depsFlag := cmd.Flags().Changed("depends-on")

	if !titleFlag && !descFlag && !depsFlag {
		return fmt.Errorf("no updates provided - specify at least one of: --title, --description, --depends-on")
	}

	var title, description *string
	var dependsOn []string

	if titleFlag {
		t, _ := cmd.Flags().GetString("title")
		title = &t
	}

	if descFlag {
		d, _ := cmd.Flags().GetString("description")
		description = &d
	}

	if depsFlag {
		deps, _ := cmd.Flags().GetStringSlice("depends-on")
		// Trim dependency IDs
		for _, dep := range deps {
			trimmed := strings.TrimSpace(dep)
			if trimmed != "" {
				dependsOn = append(dependsOn, trimmed)
			}
		}
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Update work unit
	if err := breakdown.UpdateWorkUnit(ctx, id, title, description, dependsOn); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrWorkUnitNotFound) {
			return fmt.Errorf("work unit %s not found in breakdown index", id)
		}
		return fmt.Errorf("failed to update work unit: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Updated work unit: %s\n", id)

	return nil
}
