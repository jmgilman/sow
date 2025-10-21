package refs

import (
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/refs"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize refs after cloning",
		Long: `Initialize references by caching and symlinking.

Run this after cloning a repository to set up all configured refs.
Each ref is cached locally and symlinked into .sow/refs/.

Types that are not enabled on this system will be skipped with warnings.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRefsInit(cmd)
		},
	}

	return cmd
}

func runRefsInit(cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Get context
	sowCtx := cmdutil.GetContext(ctx)

	// Create refs manager
	mgr := refs.NewManager(sowCtx)

	// Initialize all refs
	if err := mgr.InitRefs(ctx); err != nil {
		return err
	}

	cmd.Println("✓ Refs initialized successfully")
	return nil
}
