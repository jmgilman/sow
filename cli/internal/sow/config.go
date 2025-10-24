package sow

import (
	"fmt"
	"path/filepath"

	"github.com/jmgilman/sow/cli/schemas"
	"gopkg.in/yaml.v3"
)

// Default artifact paths relative to .sow/knowledge/.
const (
	DefaultADRsPath         = "adrs"
	DefaultDesignDocsPath   = "design"
	DefaultExplorationsPath = "explorations"
)

// LoadConfig loads the sow configuration from .sow/config.yaml.
// Returns the config with defaults applied for any unspecified values.
func LoadConfig(ctx *Context) (*schemas.Config, error) {
	fs := ctx.FS()
	if fs == nil {
		return nil, ErrNotInitialized
	}

	// Read config file
	data, err := fs.ReadFile("config.yaml")
	if err != nil {
		// Config doesn't exist, return defaults
		return getDefaultConfig(), nil
	}

	// Parse YAML
	var config schemas.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for missing values
	applyDefaults(&config)

	return &config, nil
}

// getDefaultConfig returns a config with all default values.
func getDefaultConfig() *schemas.Config {
	adrs := DefaultADRsPath
	designDocs := DefaultDesignDocsPath

	return &schemas.Config{
		Artifacts: &struct {
			Adrs        *string `json:"adrs,omitempty"`
			Design_docs *string `json:"design_docs,omitempty"`
		}{
			Adrs:        &adrs,
			Design_docs: &designDocs,
		},
	}
}

// applyDefaults fills in missing config values with defaults.
func applyDefaults(config *schemas.Config) {
	if config.Artifacts == nil {
		config.Artifacts = &struct {
			Adrs        *string `json:"adrs,omitempty"`
			Design_docs *string `json:"design_docs,omitempty"`
		}{}
	}

	if config.Artifacts.Adrs == nil {
		adrs := DefaultADRsPath
		config.Artifacts.Adrs = &adrs
	}

	if config.Artifacts.Design_docs == nil {
		designDocs := DefaultDesignDocsPath
		config.Artifacts.Design_docs = &designDocs
	}
}

// GetADRsPath returns the absolute path to the ADRs directory.
func GetADRsPath(ctx *Context, config *schemas.Config) string {
	if config.Artifacts != nil && config.Artifacts.Adrs != nil {
		return filepath.Join(ctx.RepoRoot(), ".sow", "knowledge", *config.Artifacts.Adrs)
	}
	return filepath.Join(ctx.RepoRoot(), ".sow", "knowledge", DefaultADRsPath)
}

// GetDesignDocsPath returns the absolute path to the design docs directory.
func GetDesignDocsPath(ctx *Context, config *schemas.Config) string {
	if config.Artifacts != nil && config.Artifacts.Design_docs != nil {
		return filepath.Join(ctx.RepoRoot(), ".sow", "knowledge", *config.Artifacts.Design_docs)
	}
	return filepath.Join(ctx.RepoRoot(), ".sow", "knowledge", DefaultDesignDocsPath)
}

// GetExplorationsPath returns the absolute path to the explorations directory.
// This is not configurable and always uses the default location.
func GetExplorationsPath(ctx *Context) string {
	return filepath.Join(ctx.RepoRoot(), ".sow", "knowledge", DefaultExplorationsPath)
}
