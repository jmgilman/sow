package schemas

import "time"

// KnowledgeIndex defines the schema for the knowledge index at:
// .sow/knowledge/index.yaml
//
// This tracks permanent artifacts and their metadata.
#KnowledgeIndex: {
	// Exploration summaries
	explorations: [...#ExplorationSummary]

	// ADR references (if stored in .sow/knowledge/adrs/)
	adrs?: [...#ArtifactReference] @go(,optional=nillable)

	// Design document references (if stored in .sow/knowledge/design/)
	design_docs?: [...#ArtifactReference] @go(,optional=nillable)
}

// ExplorationSummary represents a completed exploration
#ExplorationSummary: {
	// Exploration topic
	topic: string & !=""

	// Path to summary document (relative to .sow/knowledge/explorations/)
	summary_path: string & !=""

	// Original exploration branch
	branch: string & =~"^explore/[a-z0-9][a-z0-9-]*[a-z0-9]$"

	// When exploration was completed
	completed_at: time.Time

	// Tags for discoverability
	tags: [...string]
}

// ArtifactReference represents a reference to a permanent artifact
#ArtifactReference: {
	// Path to artifact (relative to .sow/knowledge/)
	path: string & !=""

	// Brief description
	description: string & !=""

	// When artifact was created
	created_at: time.Time

	// Tags for discoverability
	tags: [...string]
}
