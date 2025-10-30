//go:generate sh -c "cue exp gengotypes ./... && rm -f projects/cue_types_gen.go"

// Package schemas provides CUE schema definitions and a hybrid of generated
// and hand-written Go types for all sow state and index files.
//
// This package contains:
//   - CUE schema definitions (*.cue files) - validation rules and constraints
//   - Generated Go types for phases (phases/cue_types_gen.go)
//   - Hand-written Go types for projects (projects/standard.go)
//   - Re-exports and utilities (cue_types_gen.go)
//
// Generation Strategy:
//   - phases/common.cue → phases/cue_types_gen.go (auto-generated)
//   - projects/standard.cue → projects/standard.go (hand-written)
//
// The projects types are hand-written because CUE's gengotypes tool cannot
// properly handle the type unification patterns used in project schemas.
// Run `go generate ./schemas` to regenerate the phase types.
//
// IMPORTANT: When modifying projects/standard.cue, you MUST manually update
// projects/standard.go to keep the types in sync.
//
// The hybrid approach provides:
//   - CUE validation for runtime schema enforcement
//   - Go type safety for compile-time checking
//   - Typed fields for known phase-specific data
//   - Metadata escape hatch for unanticipated fields
package schemas
