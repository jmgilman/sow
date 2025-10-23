// Package schemas provides CUE schema definitions and generated Go types
// for all sow state and index files.
//
// This package contains both:
//   - CUE schema definitions (*.cue files) - validation rules and constraints
//   - Generated Go types (*_gen.go files) - concrete types for use in Go code
//
// The Go types are automatically generated from the CUE schemas using:
//
//	go generate ./cli/schemas
//
// This ensures that the Go types always match the CUE schema definitions.
package schemas

//go:generate go run cuelang.org/go/cmd/cue@v0.13.2 exp gengotypes ./...
