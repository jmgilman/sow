package schemas

import "time"

// RefsCacheIndex defines the schema for ~/.cache/sow/index.json
//
// This is the cache index containing transient metadata about cached
// references. Stored per-machine, not committed to git.
#RefsCacheIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Cached reference metadata
	refs: [...#CachedRef]
}

// CachedRef represents a cached reference
#CachedRef: {
	// Reference ID (matches ID from committed or local index)
	id: string & =~"^[a-z0-9-]+$"

	// Type inferred from source URL (stored for quick lookup)
	type: "git" | "file"

	// Absolute cache path (e.g., /Users/josh/.cache/sow/refs/git/abc123/)
	cache_path: string & !=""

	// Last updated timestamp
	last_updated: time.Time

	// Repositories using this cached ref
	used_by: [...#CacheUsage]

	// Type-specific metadata
	metadata: #CacheMetadata
}

// CacheMetadata is polymorphic based on type
#CacheMetadata: {
	// Git type metadata
	git?: #GitMetadata

	// File type metadata
	file?: #FileMetadata
}

// GitMetadata contains git-specific cache data
#GitMetadata: {
	// Current local commit SHA
	commit_sha: string & =~"^[a-f0-9]{40}$"

	// Latest remote commit SHA (updated on status check)
	remote_sha?: string & =~"^[a-f0-9]{40}$"

	// Last time remote was checked
	last_checked?: time.Time

	// Staleness status
	status: "current" | "behind" | "ahead" | "diverged" | "error"

	// Number of commits behind remote (0 if current)
	commits_behind: int & >=0
}

// FileMetadata contains file-specific cache data
#FileMetadata: {
	// File refs don't need additional metadata
	// Kept as struct for consistency and future expansion
}

// CacheUsage represents a repository using this cached ref
#CacheUsage: {
	// Absolute path to consuming repository
	repo_path: string & !=""

	// How the cache is linked (symlink on Unix, copy on Windows)
	link_type: "symlink" | "copy"

	// Link name in the consuming repo's .sow/refs/ directory
	link_name: string & !=""
}
