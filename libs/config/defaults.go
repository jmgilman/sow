package config

import (
	"github.com/jmgilman/sow/libs/schemas"
)

// Default artifact paths relative to .sow/knowledge/.
const (
	// DefaultADRsPath is the default directory for Architecture Decision Records.
	DefaultADRsPath = "adrs"

	// DefaultDesignDocsPath is the default directory for design documents.
	DefaultDesignDocsPath = "design"

	// DefaultExplorationsPath is the default directory for exploration documents.
	DefaultExplorationsPath = "explorations"

	// DefaultExecutorName is the default executor name for all agent bindings.
	DefaultExecutorName = "claude-code"
)

// DefaultConfig returns a Config with all default values applied.
func DefaultConfig() *schemas.Config {
	adrs := DefaultADRsPath
	designDocs := DefaultDesignDocsPath

	//nolint:revive // Field names must match generated schemas.Config structure
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

// ApplyDefaults fills in missing config values with defaults.
// This allows partial configuration - users only specify what they want to change.
func ApplyDefaults(config *schemas.Config) {
	if config.Artifacts == nil {
		//nolint:revive // Field names must match generated schemas.Config structure
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
