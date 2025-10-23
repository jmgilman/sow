package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewEnableCmd creates the command to enable an optional phase.
//
// Usage:
//
//	sow agent enable <phase>
//
// Only discovery and design phases can be enabled (they are optional).
// Implementation, review, and finalize are always required and cannot be manually enabled.
func NewEnableCmd() *cobra.Command {
	var discoveryType string

	cmd := &cobra.Command{
		Use:   "enable <phase>",
		Short: "Enable an optional phase",
		Long: `Enable an optional phase (discovery or design).

Only discovery and design phases can be enabled as they are optional.
Implementation, review, and finalize phases are always required and cannot be manually enabled.

When enabling discovery, you must specify the type using --type flag.

Example:
  # Enable discovery phase
  sow agent enable discovery --type feature

  # Enable design phase
  sow agent enable design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Validate phase
			if phase != "discovery" && phase != "design" {
				return fmt.Errorf("only discovery and design phases can be enabled (got: %s)", phase)
			}

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Build options based on phase
			var opts []sow.PhaseOption

			if phase == "discovery" {
				if discoveryType == "" {
					return fmt.Errorf("--type flag is required when enabling discovery phase")
				}
				opts = append(opts, sow.WithDiscoveryType(discoveryType))
			}

			// Enable the phase
			if err := project.EnablePhase(phase, opts...); err != nil {
				return fmt.Errorf("failed to enable %s phase: %w", phase, err)
			}

			cmd.Printf("\n✓ Enabled %s phase\n", phase)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&discoveryType, "type", "", "Discovery type (bug, feature, docs, refactor, general) - required for discovery phase")

	return cmd
}
