package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/libs/schemas"
	"gopkg.in/yaml.v3"
)

// LoadRepoConfig loads the repository configuration from .sow/config.yaml.
// It accepts a core.FS filesystem rooted at the .sow directory.
// Returns the config with defaults applied for any unspecified values.
// If config.yaml doesn't exist, returns default configuration (not an error).
func LoadRepoConfig(fsys core.FS) (*schemas.Config, error) {
	data, err := fsys.ReadFile("config.yaml")
	if err != nil {
		// Config doesn't exist, return defaults
		if isNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("load repo config: %w", err)
	}

	return LoadRepoConfigFromBytes(data)
}

// LoadRepoConfigFromBytes parses repository configuration from raw YAML bytes.
// Returns the config with defaults applied for any unspecified values.
// This is the most flexible API for testing and non-filesystem use cases.
func LoadRepoConfigFromBytes(data []byte) (*schemas.Config, error) {
	// Handle empty or whitespace-only input
	if len(bytes.TrimSpace(data)) == 0 {
		return DefaultConfig(), nil
	}

	var config schemas.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("load repo config: %w", ErrInvalidYAML)
	}

	// Apply defaults for missing values
	ApplyDefaults(&config)

	return &config, nil
}

// isNotExist checks if an error indicates a file does not exist.
func isNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
