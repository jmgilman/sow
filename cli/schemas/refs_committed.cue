package schemas

// RefsCommittedIndex defines the schema for .sow/refs/index.json
//
// This is the committed index containing categorical metadata about
// refs, shared with the team via git. Contains configuration
// but not transient data like SHAs or timestamps.
#RefsCommittedIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Reference definitions
	refs: [...#Ref]
}

// Ref represents a reference to external content
#Ref: {
	// Unique identifier (auto-generated)
	id: string & =~"^[a-z0-9-]+$"

	// Source URL with scheme that determines type
	// Examples:
	//   - git+https://github.com/org/repo
	//   - git+ssh://git@github.com/org/repo
	//   - git@github.com:org/repo (auto-converted to git+ssh://)
	//   - file:///absolute/path
	source: string & !=""

	// Semantic type (what the content represents)
	semantic: "knowledge" | "code"

	// Symlink name in .sow/refs/
	link: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"

	// Topic keywords for categorization
	tags: [...string]

	// One-sentence description
	description: string & !=""

	// 2-3 sentence summary (optional)
	summary: string

	// Type-specific configuration
	// Structure depends on URL scheme
	config: #RefConfig
}

// RefConfig is a polymorphic config structure
// Validation depends on the source URL scheme
#RefConfig: {
	// Git type config (for git+https://, git+ssh://, git@ URLs)
	// Branch name (optional, defaults to repo default)
	branch?: string & !=""

	// Subpath within repository (optional, defaults to root)
	// Use "" or omit for root
	path?: string

	// File type config (for file:// URLs)
	// No additional config needed - path is in source URL

	// Future type configs would be added here
	// For example, web type:
	// scrape_depth?: int & >=1
	// follow_links?: bool
}
