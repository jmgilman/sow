package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/repos"
)

// NewReposCmd creates the repos command
func NewReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Manage linked repositories",
		Long: `Manage linked repositories - add, sync, list, and remove repository links.

Repositories can be cloned from git or symlinked from local paths.`,
	}

	cmd.AddCommand(newReposAddCmd())
	cmd.AddCommand(newReposSyncCmd())
	cmd.AddCommand(newReposListCmd())
	cmd.AddCommand(newReposRemoveCmd())

	return cmd
}

func newReposAddCmd() *cobra.Command {
	var name string
	var symlink bool

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Add a linked repository",
		Long: `Add a linked repository from git URL or local path.

For git URLs, the repository will be cloned.
For local paths with --symlink, a symbolic link will be created.`,
		Example: `  # Clone from git
  sow repos add https://github.com/org/auth-service

  # Symlink local repository
  sow repos add /path/to/local/repo --symlink

  # Add with custom name
  sow repos add https://github.com/org/repo --name my-repo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposAdd(cmd, args[0], name, symlink)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the repository (defaults to extracted name)")
	cmd.Flags().BoolVar(&symlink, "symlink", false, "Create symlink instead of cloning (for local paths)")

	return cmd
}

func newReposSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [name]",
		Short: "Sync one or all repositories",
		Long: `Sync repositories by pulling latest changes from git.

If no name is provided, syncs all repositories.`,
		Example: `  # Sync all repositories
  sow repos sync

  # Sync specific repository
  sow repos sync auth-service`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return runReposSync(cmd, name)
		},
	}

	return cmd
}

func newReposListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List linked repositories",
		Long:  `List all linked repositories with their metadata.`,
		Example: `  # List all repositories
  sow repos list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposList(cmd)
		},
	}

	return cmd
}

func newReposRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a linked repository",
		Long:  `Remove a linked repository by deleting its directory/symlink and removing it from the index.`,
		Example: `  # Remove a repository
  sow repos remove auth-service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReposRemove(cmd, args[0])
		},
	}

	return cmd
}

func runReposAdd(cmd *cobra.Command, source string, nameFlag string, useSymlink bool) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	// Determine repository name
	repoName := extractRepoNameFromArgs(source, nameFlag)

	// Check if repo already exists
	indexPath := filepath.Join(sowDir, "repos", "index.json")
	idx, err := repos.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	if _, exists := idx.FindRepository(repoName); exists {
		return fmt.Errorf("repository '%s' already exists", repoName)
	}

	// Create repos directory if it doesn't exist
	reposDir := filepath.Join(sowDir, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("failed to create repos directory: %w", err)
	}

	repoPath := filepath.Join(reposDir, repoName)
	repoType := "clone"
	var branch *string

	// Handle git URL or local path
	if repos.IsGitSource(source) {
		// Git URL - clone repository
		fmt.Fprintf(cmd.OutOrStdout(), "Cloning repository from %s...\n", source)

		cloneCmd := exec.Command("git", "clone", "--depth", "1", source, repoPath)
		cloneCmd.Stdout = cmd.OutOrStdout()
		cloneCmd.Stderr = cmd.ErrOrStderr()

		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		// Get current branch
		branchName := getCurrentBranch(repoPath)
		branch = &branchName
	} else {
		// Local path
		if useSymlink {
			// Create symlink
			fmt.Fprintf(cmd.OutOrStdout(), "Creating symlink to %s...\n", source)

			// Convert to absolute path
			absSource, err := filepath.Abs(source)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			if err := os.Symlink(absSource, repoPath); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}

			repoType = "symlink"
		} else {
			// Copy directory
			fmt.Fprintf(cmd.OutOrStdout(), "Copying repository from %s...\n", source)

			if err := copyDir(source, repoPath); err != nil {
				return fmt.Errorf("failed to copy repository: %w", err)
			}
		}
	}

	// Add to index
	repo := repos.Repository{
		Name:      repoName,
		Path:      repoName,
		Source:    source,
		Purpose:   "Linked repository",
		Type:      repoType,
		Branch:    branch,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	idx.AddRepository(repo)

	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nRepository added: %s\n", repoName)
	fmt.Fprintf(cmd.OutOrStdout(), "Type: %s\n", repoType)
	fmt.Fprintf(cmd.OutOrStdout(), "Location: .sow/repos/%s/\n", repoName)

	return nil
}

func runReposSync(cmd *cobra.Command, name string) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "repos", "index.json")
	idx, err := repos.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Determine which repos to sync
	var reposToSync []repos.Repository
	if name == "" {
		// Sync all repos
		reposToSync = idx.Repositories
		fmt.Fprintf(cmd.OutOrStdout(), "Syncing all repositories...\n")
	} else {
		// Sync specific repo
		repo, exists := idx.FindRepository(name)
		if !exists {
			return fmt.Errorf("repository '%s' not found", name)
		}
		reposToSync = []repos.Repository{*repo}
		fmt.Fprintf(cmd.OutOrStdout(), "Syncing repository: %s\n", name)
	}

	synced := 0
	for _, repo := range reposToSync {
		// Only sync cloned git repos
		if repo.Type != "clone" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Skipping %s (%s type)\n", repo.Name, repo.Type)
			continue
		}

		if !repos.IsGitSource(repo.Source) {
			fmt.Fprintf(cmd.OutOrStdout(), "  Skipping %s (not a git source)\n", repo.Name)
			continue
		}

		repoPath := filepath.Join(sowDir, "repos", repo.Path)

		// Pull latest changes
		pullCmd := exec.Command("git", "-C", repoPath, "pull", "origin")
		pullCmd.Stdout = cmd.OutOrStdout()
		pullCmd.Stderr = cmd.ErrOrStderr()

		if err := pullCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Failed to sync %s: %v\n", repo.Name, err)
			continue
		}

		// Update timestamp in index
		for i := range idx.Repositories {
			if idx.Repositories[i].Name == repo.Name {
				idx.Repositories[i].UpdatedAt = time.Now().Format(time.RFC3339)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Synced %s\n", repo.Name)
		synced++
	}

	// Save updated index
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n%d repository(ies) synced\n", synced)
	return nil
}

func runReposList(cmd *cobra.Command) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "repos", "index.json")
	idx, err := repos.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	if len(idx.Repositories) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No repositories linked\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Add repositories with: sow repos add <source>\n")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Linked repositories:\n\n")

	for _, repo := range idx.Repositories {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", repo.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", repo.Purpose)
		fmt.Fprintf(cmd.OutOrStdout(), "  - Type: %s\n", repo.Type)
		fmt.Fprintf(cmd.OutOrStdout(), "  - Source: %s\n", repo.Source)
		if repo.Branch != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  - Branch: %s\n", *repo.Branch)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%d repository(ies) linked\n", len(idx.Repositories))

	return nil
}

func runReposRemove(cmd *cobra.Command, name string) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "repos", "index.json")
	idx, err := repos.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Check if repo exists
	repo, exists := idx.FindRepository(name)
	if !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Remove directory or symlink
	repoPath := filepath.Join(sowDir, "repos", repo.Path)
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	// Remove from index
	idx.RemoveRepository(name)
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Repository removed: %s\n", name)

	return nil
}

// Helper functions

func extractRepoNameFromArgs(source, nameFlag string) string {
	if nameFlag != "" {
		return nameFlag
	}
	return repos.ExtractNameFromSource(source)
}

func getCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	return string(output[:len(output)-1]) // Trim newline
}
