package agent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/loader"
	"github.com/spf13/cobra"
)

// NewArtifactAddCmd creates the command to add an artifact to the active phase.
//
// Usage:
//
//	sow agent artifact add <path> [--metadata key=value ...]
//
// Adds an artifact to the currently active phase. The phase must support artifacts.
func NewArtifactAddCmd() *cobra.Command {
	var metadataFlags []string

	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Add an artifact to the active phase",
		Long: `Add an artifact to the currently active phase.

The active phase must support artifacts (discovery, design, and review phases do).
The path should be relative to the .sow/ directory.

Artifacts track important files created during a phase, such as:
  - Research documents during discovery
  - Architecture diagrams during design
  - Review reports during review

Use --metadata to attach phase-specific metadata to the artifact.

Examples:
  # Add artifact (needs approval later)
  sow agent artifact add project/phases/discovery/research/001-jwt.md

  # Add review artifact with assessment
  sow agent artifact add project/phases/review/reports/001.md \
    --metadata type=review \
    --metadata assessment=pass

  # Add design artifact with category
  sow agent artifact add project/phases/design/architecture.md \
    --metadata type=design \
    --metadata category=architecture`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project via loader to get interface
			proj, err := loader.Load(ctx)
			if err != nil {
				if errors.Is(err, project.ErrNoProject) {
					return fmt.Errorf("no active project - run 'sow agent init' first")
				}
				return fmt.Errorf("failed to load project: %w", err)
			}

			// Get current phase
			phase := proj.CurrentPhase()
			if phase == nil {
				return fmt.Errorf("no active phase found")
			}

			// Parse metadata
			metadata := make(map[string]interface{})
			for _, pair := range metadataFlags {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid metadata format: %s (use key=value)", pair)
				}
				metadata[parts[0]] = parts[1]
			}

			// Add artifact via Phase interface
			err = phase.AddArtifact(path, domain.WithMetadata(metadata))
			if errors.Is(err, project.ErrNotSupported) {
				return fmt.Errorf("phase %s does not support artifacts", phase.Name())
			}
			if err != nil {
				return fmt.Errorf("failed to add artifact: %w", err)
			}

			cmd.Printf("\n✓ Added artifact to %s phase: %s\n", phase.Name(), path)
			if len(metadata) > 0 {
				cmd.Printf("  Metadata: %v\n", metadata)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringArrayVar(&metadataFlags, "metadata", nil,
		"Metadata key=value pairs (can specify multiple times)")

	return cmd
}
