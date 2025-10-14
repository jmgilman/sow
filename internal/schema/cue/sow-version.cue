// Version File Schema
// Location: .sow/.version
//
// Track structure version and migration history.
// Created by /init, updated by /migrate, read by SessionStart hook.

package sow

import "time"

// VersionFile defines the version tracking structure
#VersionFile: {
	// Structure version (semantic versioning)
	// Format: MAJOR.MINOR.PATCH
	sow_structure_version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Plugin version used (semantic versioning)
	// Format: MAJOR.MINOR.PATCH
	plugin_version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// When repository was first initialized (ISO 8601 format)
	initialized: string & time.Format(time.RFC3339)

	// Last migration timestamp (null if never migrated, ISO 8601 format)
	last_migrated: null | (string & time.Format(time.RFC3339))
}
