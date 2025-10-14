// Repository Index Schema
// Location: .sow/repos/index.json
//
// References to linked repositories for cross-repo context.
// Repositories can be cloned or symlinked for access to external code examples.

package sow

import "time"

// RepoIndex defines the repository catalog structure
#RepoIndex: {
	repositories: [...#Repository]
}

// Repository metadata
#Repository: {
	// Repository identifier (kebab-case)
	name: string & =~"^[a-z0-9]+(-[a-z0-9]+)*$"

	// Relative path from .sow/repos/ (typically matches name)
	path: string

	// Git URL or local path where repo is located
	source: string

	// Why this repo is linked (helps agents understand relevance)
	purpose: string

	// Link type: clone (full copy) or symlink (reference to local repo)
	type: "clone" | "symlink"

	// Branch/ref to use (null for default branch or symlinks)
	branch: null | string

	// Last update timestamp (ISO 8601 format)
	updated_at: string & time.Format(time.RFC3339)
}
