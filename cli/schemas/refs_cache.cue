package schemas

import "time"

// RefsCacheIndex defines the schema for ~/.cache/sow/index.json
//
// This is the cache index containing transient metadata about cached
// repositories. Stored per-machine, not committed to git.
#RefsCacheIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Cached repository metadata
	repos: [...#CachedRepo]
}

// CachedRepo represents a cached repository
#CachedRepo: {
	// Git repository URL (must match source from committed index)
	source: string & !=""

	// Branch name
	branch: string & !=""

	// Relative path in cache (from ~/.cache/sow/)
	cached_path: string & !=""

	// Current local commit SHA
	commit_sha: string & =~"^[a-f0-9]{40}$"

	// Timestamps
	cached_at:    time.Time // Initial cache time
	last_checked: time.Time // Last staleness check
	last_updated: time.Time // Last fetch/pull

	// Staleness status
	status: "current" | "behind" | "ahead" | "diverged" | "error"

	// Number of commits behind remote (null if current)
	commits_behind: *null | int & >=0

	// Latest remote commit SHA
	remote_sha: string & =~"^[a-f0-9]{40}$"

	// Repositories using this cache
	used_by: [...#CacheUsage]
}

// CacheUsage represents a repository using this cached repo
#CacheUsage: {
	// Absolute path to consuming repository
	repo_path: string & !=""

	// How the cache is linked (symlink on Unix, copy on Windows)
	link_type: "symlink" | "copy"

	// Subpaths from this cache used by the consuming repo
	paths: [...string]
}
