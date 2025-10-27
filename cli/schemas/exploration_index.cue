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

	// Topics "parking lot" - research areas agreed upon but not yet explored
	topics: [...#ExplorationTopic]

	// Session journal - chronological log for zero-context recovery
	journal: [...#JournalEntry]

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

// ExplorationTopic represents a research topic in the parking lot
#ExplorationTopic: {
	// Topic description
	topic: string & !=""

	// Topic status
	status: "pending" | "in_progress" | "completed"

	// When this topic was added
	added_at: time.Time

	// When this topic was completed (optional)
	completed_at?: time.Time

	// Files created for this topic (optional)
	related_files?: [...string]

	// Optional brief notes
	notes?: string
}

// JournalEntry represents a chronological log entry for session memory
#JournalEntry: {
	// When this entry was created
	timestamp: time.Time

	// Entry type for categorization
	type: "decision" | "insight" | "question" | "topic_added" | "topic_completed" | "note"

	// Entry content
	content: string & !=""
}
