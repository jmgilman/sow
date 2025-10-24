package breakdown

import (
	"errors"
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewPublishCmd creates the breakdown publish command.
func NewPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish [unit-id]",
		Short: "Publish work unit(s) as GitHub issues",
		Long: `Publish approved work unit(s) as GitHub issues with the 'sow' label.

If a unit ID is provided, only that unit will be published.
If no ID is provided, all approved units will be published.

For each work unit:
1. Reads the detailed markdown document (if exists)
2. Creates a GitHub issue with the title and document body
3. Adds the 'sow' label automatically
4. Updates the index with the issue URL and number
5. Marks the unit as "published"

Requirements:
  - Must be in a sow repository with an active breakdown session
  - GitHub CLI (gh) must be installed and authenticated
  - Work unit(s) must be approved
  - Work unit(s) must not already be published

Examples:
  # Publish a specific work unit
  sow breakdown publish unit-001

  # Publish all approved work units
  sow breakdown publish`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var unitID string
			if len(args) > 0 {
				unitID = args[0]
			}
			return runPublish(cmd, unitID)
		},
	}

	return cmd
}

func runPublish(cmd *cobra.Command, unitID string) error {
	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Create GitHub client
	ghExec := exec.NewLocal("gh")
	gh := sow.NewGitHub(ghExec)

	// Get work units to publish
	units, err := getUnitsToPublish(ctx, unitID)
	if err != nil {
		return err
	}

	if len(units) == 0 {
		cmd.Println("No approved work units to publish")
		return nil
	}

	// Publish each unit
	cmd.Printf("\nPublishing %d work unit(s)...\n\n", len(units))

	for _, entry := range units {
		if err := publishWorkUnit(cmd, ctx, gh, entry.id, entry.unit); err != nil {
			return err
		}
	}

	cmd.Printf("\n✓ Successfully published %d work unit(s)\n", len(units))

	return nil
}

type workUnitEntry struct {
	id   string
	unit *schemas.BreakdownWorkUnit
}

func getUnitsToPublish(ctx *sow.Context, unitID string) ([]workUnitEntry, error) {
	if unitID != "" {
		return getSingleUnit(ctx, unitID)
	}
	return getAllUnpublishedUnits(ctx)
}

func getSingleUnit(ctx *sow.Context, unitID string) ([]workUnitEntry, error) {
	unit, err := breakdown.GetWorkUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return nil, fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrWorkUnitNotFound) {
			return nil, fmt.Errorf("work unit %s not found in breakdown index", unitID)
		}
		return nil, fmt.Errorf("failed to get work unit: %w", err)
	}

	return []workUnitEntry{{id: unitID, unit: unit}}, nil
}

func getAllUnpublishedUnits(ctx *sow.Context) ([]workUnitEntry, error) {
	unpublished, err := breakdown.GetUnpublishedUnits(ctx)
	if err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return nil, fmt.Errorf("no active breakdown session")
		}
		return nil, fmt.Errorf("failed to get unpublished units: %w", err)
	}

	var units []workUnitEntry
	for _, unit := range unpublished {
		unitCopy := unit
		units = append(units, workUnitEntry{id: unit.Id, unit: &unitCopy})
	}

	return units, nil
}

func publishWorkUnit(cmd *cobra.Command, ctx *sow.Context, gh *sow.GitHub, id string, unit *schemas.BreakdownWorkUnit) error {
	// Check if already published
	if unit.Status == "published" {
		return fmt.Errorf("work unit %s is already published (issue #%d)", id, unit.Github_issue_number)
	}

	// Check if approved
	if unit.Status != "approved" {
		return fmt.Errorf("work unit %s is not approved (current status: %s)", id, unit.Status)
	}

	// Read document body if exists
	var body string
	if unit.Document_path != "" {
		fs := ctx.FS()
		fullPath := fmt.Sprintf("breakdown/%s", unit.Document_path)
		content, err := fs.ReadFile(fullPath)
		if err != nil {
			// If document doesn't exist, use description as body
			cmd.Printf("  Warning: Could not read document at %s, using description as body\n", unit.Document_path)
			body = unit.Description
		} else {
			body = string(content)
		}
	} else {
		// No document, use description as body
		body = unit.Description
	}

	// Create GitHub issue with 'sow' label
	cmd.Printf("Creating issue for %s: %s...\n", id, unit.Title)

	issue, err := gh.CreateIssue(unit.Title, body, []string{"sow"})
	if err != nil {
		return fmt.Errorf("failed to create GitHub issue for %s: %w", id, err)
	}

	// Update index with issue information
	if err := breakdown.PublishWorkUnit(ctx, id, issue.URL, int64(issue.Number)); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: Issue created (#%d) but failed to update index: %v\n", issue.Number, err)
		return err
	}

	cmd.Printf("  ✓ Created issue #%d: %s\n", issue.Number, issue.URL)

	return nil
}
