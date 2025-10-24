package schemas

import "time"

// ExplorationIndex defines the schema for exploration index files at:
// .sow/exploration/index.yaml
//
// This tracks files and metadata for active exploration work.
#ExplorationIndex: {
	// Exploration metadata
	exploration: {
		// Topic being explored (human-readable)
		topic: string & !=""

		// Git branch name for this exploration
		branch: string & =~"^explore/[a-z0-9][a-z0-9-]*[a-z0-9]$"

		// When this exploration was created
		created_at: time.Time

		// Exploration status
		status: "active" | "completed" | "abandoned"
	}

	// Files in this exploration
	files: [...#ExplorationFile]
}

// ExplorationFile represents a file in the exploration workspace
#ExplorationFile: {
	// Path relative to .sow/exploration/
	path: string & !=""

	// Brief description of file contents
	description: string & !=""

	// Keywords for discoverability
	tags: [...string]

	// When this file was created
	created_at: time.Time
}
