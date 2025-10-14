// Sink Index Schema
// Location: .sow/sinks/index.json
//
// LLM-maintained catalog of installed sinks with metadata for agent discovery.
// The orchestrator reads this index to determine relevant sinks for tasks.

package sow

import "time"

// SinkIndex defines the sink catalog structure
#SinkIndex: {
	sinks: [...#Sink]
}

// Sink metadata
#Sink: {
	// Sink identifier (kebab-case)
	name: string & =~"^[a-z0-9]+(-[a-z0-9]+)*$"

	// Relative path from .sow/sinks/ (typically matches name)
	path: string

	// Human-readable description of sink contents
	description: string

	// Topics covered by this sink (used for matching tasks)
	topics: [...string]

	// Guidance for when to reference this sink
	when_to_use: string

	// Version or git ref
	version: string

	// Git URL or path where sink originated
	source: string

	// Last update timestamp (ISO 8601 format)
	updated_at: string & time.Format(time.RFC3339)
}
