package schemas

// RefsLocalIndex defines the schema for .sow/refs/index.local.json
//
// This is the local-only index for references that should not be
// shared with the team (e.g., work-in-progress docs, personal notes).
// This file is git-ignored.
//
// Note: File refs are typically added as local refs since they
// reference paths on the local machine.
#RefsLocalIndex: {
	// Schema version (semantic versioning)
	version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// Local reference definitions
	// Uses same structure as committed refs
	refs: [...#Ref]
}

// Note: #Ref is defined in refs_committed.cue
// Local refs use the same structure but:
// - Typically use file:// URLs
// - Are git-ignored (not shared with team)
// - Can also be git refs if user wants private access
