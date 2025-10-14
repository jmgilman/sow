package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/sinks"
)

// NewSinksCmd creates the sinks command
func NewSinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sinks",
		Short: "Manage knowledge sinks",
		Long: `Manage knowledge sinks - install, update, list, and remove external knowledge sources.

Sinks are collections of Markdown files providing focused context on specific topics.`,
	}

	cmd.AddCommand(newSinksInstallCmd())
	cmd.AddCommand(newSinksUpdateCmd())
	cmd.AddCommand(newSinksListCmd())
	cmd.AddCommand(newSinksRemoveCmd())

	return cmd
}

func newSinksInstallCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "install <git-url>",
		Short: "Install a sink from git repository",
		Long: `Install a knowledge sink from a git repository.

The sink will be cloned to .sow/sinks/<name>/ and added to the index.`,
		Example: `  # Install from git
  sow sinks install https://github.com/org/python-style

  # Install with custom name
  sow sinks install https://github.com/org/standards --name python-style`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSinksInstall(cmd, args[0], name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the sink (defaults to repository name)")

	return cmd
}

func newSinksUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update one or all sinks",
		Long: `Update sinks by pulling latest changes from git.

If no name is provided, updates all sinks.`,
		Example: `  # Update all sinks
  sow sinks update

  # Update specific sink
  sow sinks update python-style`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return runSinksUpdate(cmd, name)
		},
	}

	return cmd
}

func newSinksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed sinks",
		Long:  `List all installed knowledge sinks with their metadata.`,
		Example: `  # List all sinks
  sow sinks list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSinksList(cmd)
		},
	}

	return cmd
}

func newSinksRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed sink",
		Long:  `Remove a knowledge sink by deleting its directory and removing it from the index.`,
		Example: `  # Remove a sink
  sow sinks remove python-style`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSinksRemove(cmd, args[0])
		},
	}

	return cmd
}

func runSinksInstall(cmd *cobra.Command, gitURL string, nameFlag string) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	// Determine sink name
	sinkName := extractNameFromArgs(gitURL, nameFlag)

	// Check if sink already exists
	indexPath := filepath.Join(sowDir, "sinks", "index.json")
	idx, err := sinks.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	if _, exists := idx.FindSink(sinkName); exists {
		return fmt.Errorf("sink '%s' already exists", sinkName)
	}

	// Create sinks directory if it doesn't exist
	sinksDir := filepath.Join(sowDir, "sinks")
	if err := os.MkdirAll(sinksDir, 0755); err != nil {
		return fmt.Errorf("failed to create sinks directory: %w", err)
	}

	// Clone the repository
	sinkPath := filepath.Join(sinksDir, sinkName)
	fmt.Fprintf(cmd.OutOrStdout(), "Installing sink from %s...\n", gitURL)

	if !sinks.IsGitURL(gitURL) {
		// Local path - copy directory
		if err := copyDir(gitURL, sinkPath); err != nil {
			return fmt.Errorf("failed to copy sink: %w", err)
		}
	} else {
		// Git URL - clone with shallow depth
		cloneCmd := exec.Command("git", "clone", "--depth", "1", gitURL, sinkPath)
		cloneCmd.Stdout = cmd.OutOrStdout()
		cloneCmd.Stderr = cmd.ErrOrStderr()

		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// Get version (git commit hash if available)
	version := "local"
	if sinks.IsGitURL(gitURL) {
		version = getGitCommitHash(sinkPath)
	}

	// Add to index
	sink := sinks.Sink{
		Name:        sinkName,
		Path:        sinkName,
		Description: fmt.Sprintf("Sink from %s", gitURL),
		Topics:      []string{},
		WhenToUse:   "See documentation in sink directory",
		Version:     version,
		Source:      gitURL,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	idx.AddSink(sink)

	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nSink installed: %s\n", sinkName)
	fmt.Fprintf(cmd.OutOrStdout(), "Location: .sow/sinks/%s/\n", sinkName)

	return nil
}

func runSinksUpdate(cmd *cobra.Command, name string) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "sinks", "index.json")
	idx, err := sinks.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Determine which sinks to update
	var sinksToUpdate []sinks.Sink
	if name == "" {
		// Update all sinks
		sinksToUpdate = idx.Sinks
		fmt.Fprintf(cmd.OutOrStdout(), "Updating all sinks...\n")
	} else {
		// Update specific sink
		sink, exists := idx.FindSink(name)
		if !exists {
			return fmt.Errorf("sink '%s' not found", name)
		}
		sinksToUpdate = []sinks.Sink{*sink}
		fmt.Fprintf(cmd.OutOrStdout(), "Updating sink: %s\n", name)
	}

	updated := 0
	for _, sink := range sinksToUpdate {
		sinkPath := filepath.Join(sowDir, "sinks", sink.Path)

		// Only update git-based sinks
		if !sinks.IsGitURL(sink.Source) {
			fmt.Fprintf(cmd.OutOrStdout(), "  Skipping %s (local sink)\n", sink.Name)
			continue
		}

		// Pull latest changes
		pullCmd := exec.Command("git", "-C", sinkPath, "pull", "origin")
		pullCmd.Stdout = cmd.OutOrStdout()
		pullCmd.Stderr = cmd.ErrOrStderr()

		if err := pullCmd.Run(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Failed to update %s: %v\n", sink.Name, err)
			continue
		}

		// Update version in index
		for i := range idx.Sinks {
			if idx.Sinks[i].Name == sink.Name {
				idx.Sinks[i].Version = getGitCommitHash(sinkPath)
				idx.Sinks[i].UpdatedAt = time.Now().Format(time.RFC3339)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Updated %s\n", sink.Name)
		updated++
	}

	// Save updated index
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n%d sink(s) updated\n", updated)
	return nil
}

func runSinksList(cmd *cobra.Command) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "sinks", "index.json")
	idx, err := sinks.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	if len(idx.Sinks) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No sinks installed\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Install sinks with: sow sinks install <git-url>\n")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Installed sinks:\n\n")

	for _, sink := range idx.Sinks {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s (%s)\n", sink.Name, sink.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", sink.Description)
		if len(sink.Topics) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  - Topics: %v\n", sink.Topics)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  - Source: %s\n", sink.Source)
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%d sink(s) installed\n", len(idx.Sinks))

	return nil
}

func runSinksRemove(cmd *cobra.Command, name string) error {
	sowDir, err := getSowDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(sowDir, "sinks", "index.json")
	idx, err := sinks.LoadOrCreate(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Check if sink exists
	sink, exists := idx.FindSink(name)
	if !exists {
		return fmt.Errorf("sink '%s' not found", name)
	}

	// Remove directory
	sinkPath := filepath.Join(sowDir, "sinks", sink.Path)
	if err := os.RemoveAll(sinkPath); err != nil {
		return fmt.Errorf("failed to remove sink directory: %w", err)
	}

	// Remove from index
	idx.RemoveSink(name)
	if err := idx.Save(indexPath); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Sink removed: %s\n", name)

	return nil
}

// Helper functions

func extractNameFromArgs(url, nameFlag string) string {
	if nameFlag != "" {
		return nameFlag
	}
	return sinks.ExtractNameFromURL(url)
}

func getGitCommitHash(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return string(output[:len(output)-1]) // Trim newline
}

func copyDir(src, dst string) error {
	// Simple directory copy implementation
	return exec.Command("cp", "-r", src, dst).Run()
}

func getSowDir() (string, error) {
	// Find .sow/ directory by walking up
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		sowPath := filepath.Join(dir, ".sow")
		if info, err := os.Stat(sowPath); err == nil && info.IsDir() {
			return sowPath, nil
		}

		// Move up one level
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", fmt.Errorf(".sow directory not found - not in a sow repository")
		}
		dir = parent
	}
}
