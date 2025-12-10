package config

import (
	"path/filepath"

	"github.com/jmgilman/sow/libs/schemas"
)

// GetADRsPath returns the absolute path to the ADRs directory.
// The path is computed as: <repoRoot>/.sow/knowledge/<config.artifacts.adrs>
// If config is nil or the ADRs path is not configured, uses DefaultADRsPath.
func GetADRsPath(repoRoot string, config *schemas.Config) string {
	path := DefaultADRsPath
	if config != nil && config.Artifacts != nil && config.Artifacts.Adrs != nil {
		path = *config.Artifacts.Adrs
	}
	return filepath.Join(repoRoot, ".sow", "knowledge", path)
}

// GetDesignDocsPath returns the absolute path to the design docs directory.
// The path is computed as: <repoRoot>/.sow/knowledge/<config.artifacts.design_docs>
// If config is nil or the design docs path is not configured, uses DefaultDesignDocsPath.
func GetDesignDocsPath(repoRoot string, config *schemas.Config) string {
	path := DefaultDesignDocsPath
	if config != nil && config.Artifacts != nil && config.Artifacts.Design_docs != nil {
		path = *config.Artifacts.Design_docs
	}
	return filepath.Join(repoRoot, ".sow", "knowledge", path)
}

// GetExplorationsPath returns the absolute path to the explorations directory.
// This path is not configurable and always uses DefaultExplorationsPath.
// The path is computed as: repoRoot/.sow/knowledge/explorations.
func GetExplorationsPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".sow", "knowledge", DefaultExplorationsPath)
}

// GetKnowledgePath returns the absolute path to the knowledge directory.
// The path is computed as: repoRoot/.sow/knowledge.
func GetKnowledgePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".sow", "knowledge")
}
