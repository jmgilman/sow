package config

import (
	"fmt"
	"os"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

func newPathCmd() *cobra.Command {
	var existsFlag bool

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long: `Show the path to the user configuration file.

The path is platform-specific:
  Linux/Mac: ~/.config/sow/config.yaml
  Windows:   %APPDATA%\sow\config.yaml

Use --exists to check if the file exists (for scripting).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPath(cmd, existsFlag)
		},
	}

	cmd.Flags().BoolVar(&existsFlag, "exists", false, "Check if config file exists (outputs true/false)")

	return cmd
}

func runPath(cmd *cobra.Command, checkExists bool) error {
	path, err := sow.GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if checkExists {
		_, err := os.Stat(path)
		if err == nil {
			cmd.Println("true")
		} else {
			cmd.Println("false")
		}
		return nil
	}

	cmd.Println(path)
	return nil
}

// runPathWithOptions is a helper that allows testing with custom paths.
func runPathWithOptions(cmd *cobra.Command, path string, checkExists bool) error {
	if checkExists {
		_, err := os.Stat(path)
		if err == nil {
			cmd.Println("true")
		} else {
			cmd.Println("false")
		}
		return nil
	}

	cmd.Println(path)
	return nil
}
