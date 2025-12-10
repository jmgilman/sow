package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/sow/libs/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show effective configuration",
		Long: `Display the effective configuration after merging:
  1. Built-in defaults
  2. Config file (if exists)
  3. Environment variables (highest priority)

The output shows what configuration is actually being used.`,
		RunE: runShow,
	}
}

func runShow(cmd *cobra.Command, _ []string) error {
	path, err := config.GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	return runShowWithPath(cmd, path)
}

// runShowWithPath is a helper that allows testing with custom paths.
func runShowWithPath(cmd *cobra.Command, path string) error {
	// Create a local filesystem for config loading
	fsys := billy.NewLocal()

	// Load effective config using the internal loader for path-specific loading
	cfg, err := config.LoadUserConfigFromPath(fsys, path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get config file status
	_, fileErr := os.Stat(path)
	fileExists := fileErr == nil

	// Print header with source info
	cmd.Println("# Effective configuration (merged from defaults + file + environment)")
	if fileExists {
		cmd.Printf("# Config file: %s (exists)\n", path)
	} else {
		cmd.Printf("# Config file: %s (not found, using defaults)\n", path)
	}

	// Check for env overrides
	envOverrides := getEnvOverrides()
	if len(envOverrides) > 0 {
		cmd.Printf("# Environment overrides: %s\n", strings.Join(envOverrides, ", "))
	}
	cmd.Println()

	// Output as YAML
	output, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	cmd.Print(string(output))

	return nil
}

// getEnvOverrides returns a list of SOW_AGENTS_* environment variables that are set.
func getEnvOverrides() []string {
	envVars := []string{
		"SOW_AGENTS_ORCHESTRATOR",
		"SOW_AGENTS_IMPLEMENTER",
		"SOW_AGENTS_ARCHITECT",
		"SOW_AGENTS_REVIEWER",
		"SOW_AGENTS_PLANNER",
		"SOW_AGENTS_RESEARCHER",
		"SOW_AGENTS_DECOMPOSER",
	}

	var set []string
	for _, ev := range envVars {
		if os.Getenv(ev) != "" {
			set = append(set, ev)
		}
	}
	return set
}
