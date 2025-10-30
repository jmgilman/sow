// Package schemas provides CUE schema definitions and hand-written Go types
// for all sow state and index files.
//
// This package contains both:
//   - CUE schema definitions (*.cue files) - validation rules and constraints
//   - Hand-written Go types (*.go files) - concrete types for use in Go code
//
// IMPORTANT: The Go types are manually maintained and must be kept in sync
// with the CUE schemas. Both CUE and Go files include warnings to remind
// developers to update both when making changes.
//
// Type Organization:
//   - schemas/phases/common.{cue,go} - Core phase and artifact types
//   - schemas/projects/standard.{cue,go} - Standard project workflow types
//   - schemas/cue_types_gen.go - Re-exports and additional generated types
//
// The hybrid approach provides:
//   - CUE validation for runtime schema enforcement
//   - Go type safety for compile-time checking
//   - Typed fields for known phase-specific data
//   - Metadata escape hatch for unanticipated fields
package schemas
