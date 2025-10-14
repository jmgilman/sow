package schema

import _ "embed"

// Embedded CUE schemas
// These are compiled into the binary at build time using go:embed directives

//go:embed cue/project-state.cue
var projectStateCUE string

//go:embed cue/task-state.cue
var taskStateCUE string

//go:embed cue/sink-index.cue
var sinkIndexCUE string

//go:embed cue/repo-index.cue
var repoIndexCUE string

//go:embed cue/sow-version.cue
var sowVersionCUE string

// GetSchema returns the embedded schema for the given type
func GetSchema(schemaType string) string {
	switch schemaType {
	case "project-state":
		return projectStateCUE
	case "task-state":
		return taskStateCUE
	case "sink-index":
		return sinkIndexCUE
	case "repo-index":
		return repoIndexCUE
	case "sow-version":
		return sowVersionCUE
	default:
		return ""
	}
}

// ListSchemas returns all available schema types
func ListSchemas() []string {
	return []string{
		"project-state",
		"task-state",
		"sink-index",
		"repo-index",
		"sow-version",
	}
}
