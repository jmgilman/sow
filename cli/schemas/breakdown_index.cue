package schemas

import "time"

// BreakdownIndex defines the schema for breakdown mode index files at:
// .sow/breakdown/index.yaml
//
// This tracks input sources and work units for breaking down designs into GitHub issues.
#BreakdownIndex: {
	// Breakdown session metadata
	breakdown: {
		// Topic being broken down (human-readable)
		topic: string & !=""

		// Git branch name for this breakdown session
		branch: string & =~"^breakdown/[a-z0-9][a-z0-9-]*[a-z0-9]$"

		// When this breakdown session was created
		created_at: time.Time

		// Breakdown session status
		status: "active" | "completed" | "abandoned"
	}

	// Input sources for this breakdown session
	inputs: [...#BreakdownInput]

	// Work units to be created as GitHub issues
	work_units: [...#BreakdownWorkUnit]
}

// BreakdownInput represents an input source for the breakdown process
#BreakdownInput: {
	// Input type
	type: "design" | "exploration" | "file" | "reference" | "url" | "git"

	// Path, glob pattern, directory, or identifier
	path: string & !=""

	// Brief description of what this input provides
	description: string & !=""

	// Optional tags for organization
	tags?: [...string]

	// When this input was added
	added_at: time.Time
}

// BreakdownWorkUnit represents a unit of work to be created as a GitHub issue
#BreakdownWorkUnit: {
	// Unique identifier for this work unit (e.g., "unit-001")
	id: string & =~"^unit-[0-9]{3}$"

	// Title of the work unit (will be GitHub issue title)
	title: string & !=""

	// Brief description of the work unit
	description: string & !=""

	// Path to the detailed markdown document (relative to .sow/breakdown/)
	document_path?: string

	// Work unit status
	status: "proposed" | "document_created" | "approved" | "published"

	// IDs of work units this depends on
	depends_on?: [...string]

	// GitHub issue URL (set after publishing)
	github_issue_url?: string

	// GitHub issue number (set after publishing)
	github_issue_number?: int

	// When this work unit was created
	created_at: time.Time

	// When this work unit was last updated
	updated_at: time.Time
}
