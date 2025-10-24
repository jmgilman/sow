package exploration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewUpdateFileCmd creates the exploration update-file command.
func NewUpdateFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-file <path>",
		Short: "Update a file's metadata in the exploration index",
		Long: `Update a file's description and/or tags in the exploration index.

The file must already exist in the index. At least one of --description or
--tags must be provided.

Requirements:
  - Must be in a sow repository with an active exploration
  - File must exist in the index

Example:
  sow exploration update-file research.md --description "Updated OAuth comparison"
  sow exploration update-file findings/auth.md --tags "findings,auth,jwt"
  sow exploration update-file notes.md --description "New findings" --tags "notes,research"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateFile(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "New description for the file")
	cmd.Flags().StringSliceP("tags", "t", nil, "New comma-separated tags for the file")

	return cmd
}

func runUpdateFile(cmd *cobra.Command, args []string) error {
	path := args[0]
	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Check that at least one flag is provided
	if description == "" && len(tags) == 0 {
		return fmt.Errorf("must provide at least one of --description or --tags")
	}

	// Trim tags
	if len(tags) > 0 {
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Update file in index
	if err := exploration.UpdateFile(ctx, path, description, tags); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		if errors.Is(err, exploration.ErrFileNotFound) {
			return fmt.Errorf("file %s not found in exploration index", path)
		}
		return fmt.Errorf("failed to update file: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Updated file in exploration index: %s\n", path)

	return nil
}
