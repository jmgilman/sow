package schemas

// RefsLocalIndex defines the schema for .sow/refs/index.local.json
//
// This is the local-only index for references that should not be
// shared with the team (e.g., work-in-progress docs, personal notes).
// This file is git-ignored.
#RefsLocalIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Local reference definitions
	refs: [...#LocalRef]
}

// LocalRef represents a local-only reference
#LocalRef: {
	// Unique identifier
	id: string & !=""

	// Reference type
	type: "knowledge" | "code"

	// Local file path using file:// protocol
	source: string & =~"^file:///.+"

	// Symlink name in .sow/refs/
	link: string & !=""

	// Topic keywords for categorization
	tags: [...string]

	// One-sentence description
	description: string & !=""

	// 2-3 sentence summary
	summary: string & !=""
}
