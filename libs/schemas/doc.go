//go:generate sh -c "cue exp gengotypes ./..."

// Package schemas provides CUE schema definitions and generated Go types
// for all sow configuration, index, and project state files.
//
// This package is the foundation schema layer with no internal dependencies.
// All sow state files use these schemas for validation and type safety.
//
// # Schema Types
//
// Root package types:
//   - [Config]: Repository configuration at .sow/config.yaml
//   - [UserConfig]: User configuration at ~/.config/sow/config.yaml
//   - [KnowledgeIndex]: Knowledge tracking at .sow/knowledge/index.yaml
//   - [RefsCommittedIndex]: Team-shared refs at .sow/refs/index.json
//   - [RefsLocalIndex]: Local-only refs at .sow/refs/index.local.json
//   - [RefsCacheIndex]: Cache metadata at ~/.cache/sow/index.json
//
// The [project] subpackage contains project lifecycle types:
//   - [project.ProjectState]: Complete project state
//   - [project.PhaseState]: Phase state within project
//   - [project.TaskState]: Task state within phase
//   - [project.ArtifactState]: Artifact metadata
//
// # Code Generation
//
// Go types are generated from CUE schemas using:
//
//	go generate ./...
//
// This runs cue exp gengotypes to regenerate cue_types_gen.go files.
// Generated types use json tags matching CUE field names.
//
// # CUE Schemas
//
// CUE schema files are embedded via [CUESchemas] for runtime access.
// Use cuelang.org/go/cue for schema validation.
//
// See README.md for usage examples.
package schemas
