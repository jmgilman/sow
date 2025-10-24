package exploration

import (
	"errors"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewRemoveFileCmd creates the exploration remove-file command.
func NewRemoveFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-file <path>",
		Short: "Remove a file from the exploration index",
		Long: `Remove a file from the exploration index.

This removes the file's metadata from the index but does not delete the
actual file from the filesystem.

Requirements:
  - Must be in a sow repository with an active exploration
  - File must exist in the index

Example:
  sow exploration remove-file old-research.md
  sow exploration remove-file notes/discarded.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemoveFile(cmd, args)
		},
	}

	return cmd
}

func runRemoveFile(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Remove file from index
	if err := exploration.RemoveFile(ctx, path); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		if errors.Is(err, exploration.ErrFileNotFound) {
			return fmt.Errorf("file %s not found in exploration index", path)
		}
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Removed file from exploration index: %s\n", path)
	cmd.Printf("\nNote: The file still exists on disk, only removed from index.\n")

	return nil
}
