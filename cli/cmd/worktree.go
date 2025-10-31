package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/jmgilman/go/git"
	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewWorktreeCmd creates the worktree command.
func NewWorktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktrees",
		Long: `Manage git worktrees used for concurrent sow sessions.

Worktrees enable running multiple sow sessions simultaneously:
  - Exploration sessions on different topics
  - Project work on separate branches
  - Design sessions for different features

Each worktree has isolated state while sharing committed knowledge.`,
	}

	// Add subcommands
	cmd.AddCommand(newWorktreeListCmd())
	cmd.AddCommand(newWorktreeRemoveCmd())
	cmd.AddCommand(newWorktreePruneCmd())

	return cmd
}

// newWorktreeListCmd creates the worktree list subcommand.
func newWorktreeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees with session information",
		Long:  `Display all active worktrees showing path, branch, session type, and status.`,
		RunE:  runWorktreeList,
	}
}

func runWorktreeList(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Get all worktrees using git wrapper
	worktrees, err := ctx.Git().Repository().ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Filter to only .sow/worktrees/ (exclude main repo)
	sowWorktrees := []WorktreeInfo{}
	for _, wt := range worktrees {
		// Check if this is a .sow/worktrees/* worktree
		if !isSowWorktree(wt.Path(), ctx.RepoRoot()) {
			continue
		}

		// Detect session type
		sessionType := detectSessionType(wt.Path())

		// Get branch name
		branchName := extractBranchName(wt.Path(), ctx.RepoRoot())

		sowWorktrees = append(sowWorktrees, WorktreeInfo{
			Path:        wt.Path(),
			Branch:      branchName,
			SessionType: sessionType,
			Status:      "active", // Could check for locked, etc.
		})
	}

	// Display in table format
	if len(sowWorktrees) == 0 {
		fmt.Println("No sow worktrees found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PATH\tBRANCH\tSESSION TYPE\tSTATUS"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for _, wt := range sowWorktrees {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			wt.Path, wt.Branch, wt.SessionType, wt.Status); err != nil {
			return fmt.Errorf("failed to write worktree info: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	return nil
}

// WorktreeInfo holds information about a worktree.
type WorktreeInfo struct {
	Path        string
	Branch      string
	SessionType string
	Status      string
}

// isSowWorktree checks if a path is a .sow/worktrees/* worktree.
func isSowWorktree(path, repoRoot string) bool {
	worktreesDir := filepath.Join(repoRoot, ".sow", "worktrees")
	rel, err := filepath.Rel(worktreesDir, path)
	if err != nil {
		return false
	}
	// Check that path is under .sow/worktrees/ and not "." or starts with ".."
	// If rel starts with "..", it's outside the worktreesDir
	return !filepath.IsAbs(rel) && rel != "." && !strings.HasPrefix(rel, "..")
}

// detectSessionType checks which type of session is active in a worktree.
func detectSessionType(worktreePath string) string {
	// Check for session state directories
	if _, err := os.Stat(filepath.Join(worktreePath, ".sow", "project")); err == nil {
		return "project"
	}
	if _, err := os.Stat(filepath.Join(worktreePath, ".sow", "exploration")); err == nil {
		return "exploration"
	}
	if _, err := os.Stat(filepath.Join(worktreePath, ".sow", "design")); err == nil {
		return "design"
	}
	if _, err := os.Stat(filepath.Join(worktreePath, ".sow", "breakdown")); err == nil {
		return "breakdown"
	}
	return "unknown"
}

// extractBranchName extracts the branch name from a worktree path.
func extractBranchName(worktreePath, repoRoot string) string {
	worktreesDir := filepath.Join(repoRoot, ".sow", "worktrees")
	rel, err := filepath.Rel(worktreesDir, worktreePath)
	if err != nil {
		return "unknown"
	}
	return rel
}

// newWorktreeRemoveCmd creates the worktree remove subcommand.
func newWorktreeRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a worktree",
		Long: `Remove a worktree at the specified path.

The worktree must be clean (no uncommitted changes) to be removed.
Use 'sow worktree list' to see available worktrees.`,
		Args: cobra.ExactArgs(1),
		RunE: runWorktreeRemove,
	}
}

func runWorktreeRemove(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.GetContext(cmd.Context())
	path := args[0]

	// Normalize path to absolute for consistent comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", absPath)
	}

	// Get all worktrees to find the matching one
	worktrees, err := ctx.Git().Repository().ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Find the worktree to remove
	var targetWorktree *git.Worktree
	for _, wt := range worktrees {
		if wt.Path() == absPath {
			targetWorktree = wt
			break
		}
	}

	if targetWorktree == nil {
		return fmt.Errorf("path is not a worktree: %s", absPath)
	}

	// Remove using git wrapper (includes safety checks)
	if err := targetWorktree.Remove(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	fmt.Printf("Removed worktree: %s\n", absPath)
	return nil
}

// newWorktreePruneCmd creates the worktree prune subcommand.
func newWorktreePruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove orphaned worktree metadata",
		Long: `Clean up worktree administrative files that no longer have a corresponding worktree directory.

This is useful after manually deleting worktree directories or after system crashes.
It's safe to run periodically as it only removes metadata for non-existent worktrees.`,
		RunE: runWorktreePrune,
	}
}

func runWorktreePrune(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.GetContext(cmd.Context())

	// Prune using git wrapper
	if err := ctx.Git().Repository().PruneWorktrees(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}

	fmt.Println("Pruned orphaned worktree metadata")
	return nil
}
