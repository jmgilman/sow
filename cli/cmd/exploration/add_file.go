package exploration

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/exploration"
	"github.com/spf13/cobra"
)

// NewAddFileCmd creates the exploration add-file command.
func NewAddFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-file <path>",
		Short: "Add a file to the exploration index",
		Long: `Add a file to the current exploration's index with description and tags.

The file path should be relative to .sow/exploration/. This registers the
file in the index for context management and discoverability.

Requirements:
  - Must be in a sow repository with an active exploration
  - File must not already be in the index

Example:
  sow exploration add-file research.md --description "OAuth vs JWT comparison" --tags "oauth,jwt,research"
  sow exploration add-file findings/auth.md --description "Key findings on auth approaches" --tags "findings,auth"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddFile(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Brief description of file contents (required)")
	cmd.Flags().StringSliceP("tags", "t", nil, "Comma-separated tags for discoverability")

	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runAddFile(cmd *cobra.Command, args []string) error {
	path := args[0]
	description, _ := cmd.Flags().GetString("description")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Trim tags
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Get context
	ctx := cmdutil.GetContext(cmd.Context())

	// Add file to index
	if err := exploration.AddFile(ctx, path, description, tags); err != nil {
		if errors.Is(err, exploration.ErrNoExploration) {
			return fmt.Errorf("no active exploration - run 'sow explore <topic>' first")
		}
		if errors.Is(err, exploration.ErrFileExists) {
			return fmt.Errorf("file %s already exists in exploration index", path)
		}
		return fmt.Errorf("failed to add file: %w", err)
	}

	// Success
	cmd.Printf("\n✓ Added file to exploration index: %s\n", path)
	cmd.Printf("\nFile Details:\n")
	cmd.Printf("  Path:        %s\n", path)
	cmd.Printf("  Description: %s\n", description)
	if len(tags) > 0 {
		cmd.Printf("  Tags:        %s\n", strings.Join(tags, ", "))
	} else {
		cmd.Printf("  Tags:        none\n")
	}

	return nil
}
