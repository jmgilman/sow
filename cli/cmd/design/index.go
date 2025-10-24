package design

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/design"
	"github.com/spf13/cobra"
)

// NewIndexCmd creates the design index command.
func NewIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Display the design index",
		Long: `Display the current design session's index, including all inputs and outputs.

Shows:
- Design session metadata (topic, branch, status)
- All registered input sources
- All planned output documents with target locations

Example:
  sow design index`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load index
			index, err := design.LoadIndex(ctx)
			if err != nil {
				if errors.Is(err, design.ErrNoDesign) {
					return fmt.Errorf("no active design session - run 'sow design <topic>' first")
				}
				return fmt.Errorf("failed to load design index: %w", err)
			}

			// Display index
			cmd.Printf("\nDesign Session: %s\n", index.Design.Topic)
			cmd.Printf("Branch:         %s\n", index.Design.Branch)
			cmd.Printf("Status:         %s\n", index.Design.Status)
			cmd.Printf("Created:        %s\n", index.Design.Created_at.Format("2006-01-02 15:04:05"))

			// Display inputs
			cmd.Printf("\nInputs (%d):\n", len(index.Inputs))
			if len(index.Inputs) == 0 {
				cmd.Printf("  No inputs registered yet.\n")
			} else {
				for _, input := range index.Inputs {
					cmd.Printf("  [%s] %s\n", input.Type, input.Path)
					cmd.Printf("    %s\n", input.Description)
					if len(input.Tags) > 0 {
						cmd.Printf("    Tags: %s\n", strings.Join(input.Tags, ", "))
					}
					cmd.Printf("\n")
				}
			}

			// Display outputs
			cmd.Printf("Outputs (%d):\n", len(index.Outputs))
			if len(index.Outputs) == 0 {
				cmd.Printf("  No outputs planned yet.\n")
			} else {
				for _, output := range index.Outputs {
					typeStr := ""
					if output.Type != "" {
						typeStr = fmt.Sprintf("[%s] ", output.Type)
					}
					cmd.Printf("  %s%s → %s\n", typeStr, output.Path, output.Target_location)
					cmd.Printf("    %s\n", output.Description)
					if len(output.Tags) > 0 {
						cmd.Printf("    Tags: %s\n", strings.Join(output.Tags, ", "))
					}
					cmd.Printf("\n")
				}
			}

			return nil
		},
	}

	return cmd
}
