package schemas

// RefsCommittedIndex defines the schema for .sow/refs/index.json
//
// This is the committed index containing categorical metadata about
// remote refs, shared with the team via git. Contains configuration
// but not transient data like SHAs or timestamps.
#RefsCommittedIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Remote reference definitions
	refs: [...#RemoteRef]
}

// RemoteRef represents a remote repository reference
#RemoteRef: {
	// Unique identifier (generated from source+branch)
	id: string & !=""

	// Reference type
	type: "knowledge" | "code"

	// Git repository URL (https or ssh)
	source: string & !=""

	// Branch name
	branch: string & !=""

	// Subpaths within the repository
	paths: [...#RefPath]
}

// RefPath represents a subpath within a remote repository
#RefPath: {
	// Subdirectory path within repo (empty string or "/" for root)
	path: string

	// Symlink name in .sow/refs/
	link: string & !=""

	// Topic keywords for categorization
	tags: [...string]

	// One-sentence description
	description: string & !=""

	// 2-3 sentence summary
	summary: string & !=""
}
