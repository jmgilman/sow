package project

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/spf13/cobra"
)

// newPhaseEnableCmd creates the command to enable an optional phase.
//
// Usage:
//   sow project phase enable <phase>
//   sow project phase enable discovery --type <type>
//
// Only discovery and design phases can be enabled.
// Implementation, review, and finalize are always enabled.
func newPhaseEnableCmd() *cobra.Command {
	var discoveryType string

	cmd := &cobra.Command{
		Use:   "enable <phase>",
		Short: "Enable an optional phase (discovery or design)",
		Long: `Enable an optional phase.

Only discovery and design phases can be enabled.
Implementation, review, and finalize phases are always enabled.

Discovery phase requires a type:
  - bug: Bug discovery and analysis
  - feature: Feature exploration and requirements
  - docs: Documentation research
  - refactor: Refactoring analysis
  - general: General investigation

Example:
  # Enable discovery phase for a feature
  sow project phase enable discovery --type feature

  # Enable design phase
  sow project phase enable design`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]

			// Only discovery and design can be manually enabled
			if phase != "discovery" && phase != "design" {
				return fmt.Errorf("only discovery and design phases can be manually enabled")
			}

			// Get Sow from context
			ctx := cmdutil.GetContext(cmd.Context())

			// Get project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow project init' first")
			}

			// Build options based on phase
			var opts []sow.PhaseOption
			if phase == "discovery" {
				if discoveryType == "" {
					return fmt.Errorf("discovery type required (use --type flag)")
				}
				opts = append(opts, sow.WithDiscoveryType(discoveryType))
			}

			// Enable phase (handles validation, state machine transitions, file creation)
			if err := project.EnablePhase(phase, opts...); err != nil {
				return err
			}

			cmd.Printf("\n✓ Enabled %s phase\n", phase)
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&discoveryType, "type", "", "Discovery type (required for discovery phase): bug, feature, docs, refactor, general")

	return cmd
}
