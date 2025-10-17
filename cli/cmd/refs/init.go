package refs

import (
	"context"
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/spf13/cobra"
)

func newInitCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize refs after cloning",
		Long: `Initialize references by caching and symlinking.

Run this after cloning a repository to set up all configured refs.
Each ref is cached locally and symlinked into .sow/refs/.

Types that are not enabled on this system will be skipped with warnings.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRefsInit(cmd.Context(), cmd, sowFSFromContext)
		},
	}

	return cmd
}

//nolint:funlen // Command handlers have inherent complexity
func runRefsInit(ctx context.Context, cmd *cobra.Command, sowFSFromContext func(context.Context) sowfs.SowFS) error {
	// Get SowFS from context
	sfs := sowFSFromContext(ctx)
	if sfs == nil {
		return fmt.Errorf(".sow directory not found - run 'sow init' first")
	}

	refsFS := sfs.Refs()

	// Load all refs from both indexes
	var refsToInit []refWithSource

	// Load committed refs
	committedIndex, err := refsFS.CommittedIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load committed index: %w", err)
	}
	if committedIndex != nil {
		for _, ref := range committedIndex.Refs {
			refsToInit = append(refsToInit, refWithSource{Ref: ref, Source: "committed"})
		}
	}

	// Load local refs
	localIndex, err := refsFS.LocalIndex()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load local index: %w", err)
	}
	if localIndex != nil {
		for _, ref := range localIndex.Refs {
			refsToInit = append(refsToInit, refWithSource{Ref: ref, Source: "local"})
		}
	}

	if len(refsToInit) == 0 {
		cmd.Println("No refs configured")
		return nil
	}

	// Check which types are enabled
	cmd.Println("Checking available reference types...")
	enabledTypes, err := refs.EnabledTypes(ctx)
	if err != nil {
		return fmt.Errorf("failed to check enabled types: %w", err)
	}

	enabledTypeMap := make(map[string]bool)
	for _, t := range enabledTypes {
		enabledTypeMap[t.Name()] = true
		cmd.Printf("✓ %s type enabled\n", t.Name())
	}

	disabledTypes, err := refs.DisabledTypes(ctx)
	if err != nil {
		return fmt.Errorf("failed to check disabled types: %w", err)
	}
	for _, t := range disabledTypes {
		cmd.Printf("✗ %s type disabled\n", t.Name())
	}

	cmd.Println("\nInitializing refs...")

	// Create manager
	manager, err := refs.NewManager(sfs.SowDir())
	if err != nil {
		return fmt.Errorf("failed to create refs manager: %w", err)
	}

	// Install each ref
	installed := 0
	skipped := 0
	var skippedRefs []string

	for _, rws := range refsToInit {
		ref := rws.Ref

		// Infer type
		typeName, err := refs.InferTypeFromURL(ref.Source)
		if err != nil {
			cmd.Printf("⚠ Skipped %s: failed to infer type\n", ref.Id)
			skipped++
			skippedRefs = append(skippedRefs, fmt.Sprintf("%s (failed to infer type)", ref.Id))
			continue
		}

		// Check if type is enabled
		if !enabledTypeMap[typeName] {
			cmd.Printf("⚠ Skipped %s: type %s not enabled\n", ref.Id, typeName)
			skipped++
			skippedRefs = append(skippedRefs, fmt.Sprintf("%s (type %s not enabled)", ref.Id, typeName))
			continue
		}

		// Install the ref
		workspacePath, err := manager.Install(ctx, &ref)
		if err != nil {
			cmd.Printf("✗ Failed to install %s: %v\n", ref.Id, err)
			continue
		}

		cmd.Printf("✓ Installed %s → %s\n", ref.Id, workspacePath)
		installed++
	}

	// Summary
	cmd.Printf("\n✓ Installed %d ref(s)", installed)
	if skipped > 0 {
		cmd.Printf(", skipped %d", skipped)
	}
	cmd.Println()

	if skipped > 0 {
		cmd.Println("\nSkipped refs:")
		for _, skippedRef := range skippedRefs {
			cmd.Printf("  - %s\n", skippedRef)
		}
	}

	return nil
}
