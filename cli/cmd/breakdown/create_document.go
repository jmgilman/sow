package breakdown

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/breakdown"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCreateDocumentCmd creates the breakdown create-document command.
func NewCreateDocumentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-document <id>",
		Short: "Create a detailed markdown document for a work unit",
		Long: `Create a detailed markdown document template for a work unit.

This command creates a markdown file at .sow/breakdown/units/<id>.md
with a template structure for documenting the work unit in detail.

The orchestrator will typically expand the brief work unit description
into a comprehensive document that will become the GitHub issue body.

The work unit's status will be updated to "document_created" and the
document path will be recorded in the index.

Requirements:
  - Must be in a sow repository with an active breakdown session
  - Work unit must exist in the index

Examples:
  # Create document for a work unit
  sow breakdown create-document unit-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateDocument(cmd, args)
		},
	}

	return cmd
}

func runCreateDocument(cmd *cobra.Command, args []string) error {
	id := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get work unit to use for template
	unit, err := breakdown.GetWorkUnit(ctx, id)
	if err != nil {
		if errors.Is(err, breakdown.ErrNoBreakdown) {
			return fmt.Errorf("no active breakdown session")
		}
		if errors.Is(err, breakdown.ErrWorkUnitNotFound) {
			return fmt.Errorf("work unit %s not found in breakdown index", id)
		}
		return fmt.Errorf("failed to get work unit: %w", err)
	}

	// Create document path
	documentPath := fmt.Sprintf("units/%s.md", id)

	// Create document template
	template := fmt.Sprintf(`# %s

## Overview

%s

## Objectives

- [ ] TODO: List specific objectives

## Acceptance Criteria

- [ ] TODO: Define what "done" means for this work unit

## Technical Approach

TODO: Describe the technical approach and implementation details

## Dependencies

`, unit.Title, unit.Description)

	if len(unit.Depends_on) > 0 {
		template += "\nThis work unit depends on:\n"
		for _, dep := range unit.Depends_on {
			template += fmt.Sprintf("- %s\n", dep)
		}
	} else {
		template += "None\n"
	}

	template += `
## Testing Plan

TODO: Describe how this work will be tested

## Notes

TODO: Any additional notes or considerations
`

	// Write document to filesystem
	fs := ctx.FS()
	if fs == nil {
		return fmt.Errorf("sow not initialized")
	}

	fullPath := fmt.Sprintf("breakdown/%s", documentPath)
	if err := fs.WriteFile(fullPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	// Update index with document path
	if err := breakdown.SetWorkUnitDocumentPath(ctx, id, documentPath); err != nil {
		return fmt.Errorf("failed to update work unit: %w", err)
	}

	// Success
	absPath := breakdown.GetFilePath(ctx, documentPath)
	cmd.Printf("\n✓ Created document for work unit: %s\n", id)
	cmd.Printf("  Path: %s\n", absPath)
	cmd.Printf("  Status updated: proposed → document_created\n")

	return nil
}
