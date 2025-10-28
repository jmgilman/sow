package schemas

import "time"

// DesignIndex defines the schema for design mode index files at:
// .sow/design/index.yaml
//
// This tracks input sources and planned outputs for active design work.
#DesignIndex: {
	// Design session metadata
	design: {
		// Topic being designed (human-readable)
		topic: string & !=""

		// Git branch name for this design session
		branch: string & =~"^design/[a-z0-9][a-z0-9-]*[a-z0-9]$"

		// When this design session was created
		created_at: time.Time

		// Design session status
		status: "active" | "in_review" | "completed"
	}

	// Input sources for this design session
	inputs: [...#DesignInput]

	// Output documents produced by this design session
	outputs: [...#DesignOutput]
}

// DesignInput represents an input source for the design process
#DesignInput: {
	// Input type
	type: "exploration" | "file" | "reference" | "url" | "git"

	// Path, glob pattern, directory, or identifier
	path: string & !=""

	// Brief description of what this input provides
	description: string & !=""

	// Optional tags for organization
	tags?: [...string]

	// When this input was added
	added_at: time.Time
}

// DesignOutput represents a design document to be produced
#DesignOutput: {
	// Path relative to .sow/design/
	path: string & !=""

	// Brief description of the document
	description: string & !=""

	// Target location for this specific document when finalized
	target_location: string & !=""

	// Document type
	type?: "prd" | "arc42" | "arc42-update" | "design" | "adr" | "c4-context" | "c4-container" | "c4-component" | "other"

	// For arc42-update: which section is being updated (e.g., "05-building-blocks")
	arc42_section?: string

	// Optional tags
	tags?: [...string]

	// When this output was added to the index
	added_at: time.Time
}
