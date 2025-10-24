package breakdown

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewAddUnitCmd creates the breakdown add-unit command.
func NewAddUnitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-unit",
		Short: "Add a work unit to the breakdown index",
		Long: `Add a proposed work unit to the current breakdown's index.

A work unit represents a logical piece of work that will become a GitHub issue
for a sow project. Each unit has:
- A unique ID (e.g., unit-001)
- A title (will become the GitHub issue title)
- A description (brief summary)
- Optional dependencies on other units

The orchestrator will help expand these units into detailed markdown documents
before publishing them as GitHub issues.

Requirements:
  - Must be in a sow repository with an active breakdown session
  - ID must follow format: unit-NNN (e.g., unit-001, unit-042)
  - ID must be unique within the breakdown

Examples:
  # Add a simple work unit
  sow breakdown add-unit \
    --id unit-001 \
    --title "JWT token service" \
    --description "Implement JWT token generation and validation"

  # Add a work unit with dependencies
  sow breakdown add-unit \
    --id unit-002 \
    --title "User authentication endpoint" \
    --description "Create /auth/login endpoint using JWT service" \
    --depends-on "unit-001"

  # Add a work unit with multiple dependencies
  sow breakdown add-unit \
    --id unit-003 \
    --title "Integration tests" \
    --description "End-to-end tests for authentication flow" \
    --depends-on "unit-001,unit-002"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAddUnit(cmd)
		},
	}

	// Flags
	cmd.Flags().String("id", "", "Unique work unit ID (required, format: unit-NNN)")
	cmd.Flags().String("title", "", "Title for the work unit (required)")
	cmd.Flags().String("description", "", "Brief description of the work unit (required)")
	cmd.Flags().StringSlice("depends-on", nil, "Comma-separated list of work unit IDs this depends on")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runAddUnit(cmd *cobra.Command) error {
	id, _ := cmd.Flags().GetString("id")
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	dependsOn, _ := cmd.Flags().GetStringSlice("depends-on")

	// Validate ID format
	idPattern := regexp.MustCompile(`^unit-[0-9]{3}$`)
	if !idPattern.MatchString(id) {
		return fmt.Errorf("invalid work unit ID: %s (must be format: unit-NNN, e.g., unit-001)", id)
	}

	// Trim dependency IDs
	var cleanedDeps []string
	for _, dep := range dependsOn {
		trimmed := strings.TrimSpace(dep)
		if trimmed != "" {
			cleanedDeps = append(cleanedDeps, trimmed)
		}
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add work unit to index
	if err := breakdown.AddWorkUnit(ctx, id, title, description, cleanedDeps); err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session - run 'sow breakdown <topic>' first")
		}
		if errors.Is(err, breakdown.ErrWorkUnitExists) {
			return fmt.Errorf("work unit %s already exists in breakdown index", id)
		}
		return fmt.Errorf("failed to add work unit: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added work unit to breakdown index: %s\n", id)
	cmd.Printf("\nWork Unit Details:\n")
	cmd.Printf("  ID:          %s\n", id)
	cmd.Printf("  Title:       %s\n", title)
	cmd.Printf("  Description: %s\n", description)
	cmd.Printf("  Status:      proposed\n")
	if len(cleanedDeps) > 0 {
		cmd.Printf("  Depends on:  %s\n", strings.Join(cleanedDeps, ", "))
	} else {
		cmd.Printf("  Depends on:  none\n")
	}

	return nil
}
