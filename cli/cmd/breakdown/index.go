package breakdown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewIndexCmd creates the breakdown index command.
func NewIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Display the current breakdown index",
		Long: `Display the current breakdown session's input sources and work units.

This command shows:
- Breakdown session metadata (topic, branch, status)
- All registered input sources
- All work units with their status and GitHub links (if published)

Requirements:
  - Must be in a sow repository with an active breakdown session

Examples:
  # View current breakdown index
  sow breakdown index`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIndex(cmd)
		},
	}

	return cmd
}

func runIndex(cmd *cobra.Command) error {
	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Load index
	index, err := breakdown.LoadIndex(ctx)
	if err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		return fmt.Errorf("failed to load breakdown index: %w", err)
	}

	// Display breakdown metadata
	cmd.Printf("\n=== Breakdown Session ===\n")
	cmd.Printf("Topic:  %s\n", index.Breakdown.Topic)
	cmd.Printf("Branch: %s\n", index.Breakdown.Branch)
	cmd.Printf("Status: %s\n", index.Breakdown.Status)
	cmd.Printf("Created: %s\n", index.Breakdown.Created_at.Format("2006-01-02 15:04:05"))

	// Display inputs
	cmd.Printf("\n=== Input Sources (%d) ===\n", len(index.Inputs))
	if len(index.Inputs) == 0 {
		cmd.Println("  No inputs registered yet")
	} else {
		for i, input := range index.Inputs {
			cmd.Printf("\n%d. [%s] %s\n", i+1, input.Type, input.Path)
			cmd.Printf("   %s\n", input.Description)
			if len(input.Tags) > 0 {
				cmd.Printf("   Tags: %s\n", strings.Join(input.Tags, ", "))
			}
		}
	}

	// Display work units
	displayWorkUnits(cmd, index.Work_units)

	cmd.Println()

	return nil
}

func displayWorkUnits(cmd *cobra.Command, workUnits []schemas.BreakdownWorkUnit) {
	cmd.Printf("\n=== Work Units (%d) ===\n", len(workUnits))
	if len(workUnits) == 0 {
		cmd.Println("  No work units created yet")
		return
	}

	// Count by status
	statusCounts := make(map[string]int)
	for _, unit := range workUnits {
		statusCounts[unit.Status]++
	}

	cmd.Printf("  Status breakdown:\n")
	for status, count := range statusCounts {
		cmd.Printf("    - %s: %d\n", status, count)
	}

	cmd.Println()
	for i, unit := range workUnits {
		displayWorkUnit(cmd, i+1, unit)
		if i < len(workUnits)-1 {
			cmd.Println()
		}
	}
}

func displayWorkUnit(cmd *cobra.Command, index int, unit schemas.BreakdownWorkUnit) {
	cmd.Printf("%d. [%s] %s - %s\n", index, unit.Status, unit.Id, unit.Title)
	cmd.Printf("   %s\n", unit.Description)

	if len(unit.Depends_on) > 0 {
		cmd.Printf("   Depends on: %s\n", strings.Join(unit.Depends_on, ", "))
	}

	if unit.Document_path != "" {
		cmd.Printf("   Document: %s\n", unit.Document_path)
	}

	if unit.Status == "published" {
		cmd.Printf("   GitHub: #%d - %s\n", unit.Github_issue_number, unit.Github_issue_url)
	}
}
