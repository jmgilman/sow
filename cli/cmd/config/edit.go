package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open configuration in editor",
		Long: `Open the configuration file in your preferred editor.

Uses $EDITOR environment variable, falling back to 'vi' if not set.

If no configuration file exists, creates one with the default template first.`,
		RunE: runEdit,
	}
}

func runEdit(cmd *cobra.Command, _ []string) error {
	path, err := sow.GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	editor := getEditor()
	return runEditWithPath(cmd, path, editor)
}

// getEditor returns the editor command to use.
// Respects $EDITOR environment variable, falling back to "vi" if not set.
func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

// runEditWithPath is a helper that allows testing with custom paths and editors.
func runEditWithPath(cmd *cobra.Command, path string, editor string) error {
	// Create file with template if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(configTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		cmd.Printf("Created new configuration at %s\n", path)
	}

	// Open editor
	editCmd := exec.CommandContext(cmd.Context(), editor, path)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	return nil
}
