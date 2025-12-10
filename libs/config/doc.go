// Package config provides configuration loading for sow repositories and user settings.
//
// This package provides functions to load both repository-level configuration
// (from .sow/config.yaml) and user-level configuration (from ~/.config/sow/config.yaml).
// The loading functions are decoupled from the CLI's Context type, accepting explicit
// dependencies like filesystem interfaces or raw bytes.
//
// # Repository Configuration
//
// Repository configuration controls artifact paths and other repo-specific settings.
// Load it using either a filesystem interface or raw bytes:
//
//	// From filesystem
//	cfg, err := config.LoadRepoConfig(fs)
//
//	// From bytes (more flexible)
//	cfg, err := config.LoadRepoConfigFromBytes(data)
//
// # User Configuration
//
// User configuration controls agent executor bindings and settings:
//
//	cfg, err := config.LoadUserConfig()
//	path, err := config.GetUserConfigPath()
//
// # Path Helpers
//
// Get absolute paths to configuration directories:
//
//	adrsPath := config.GetADRsPath(repoRoot, cfg)
//	designPath := config.GetDesignDocsPath(repoRoot, cfg)
//
// See README.md for more examples.
package config
