// Package state provides project state types and persistence.
//
// This package contains:
//   - Project wrapper type with runtime behavior
//   - Phase, Task, and Artifact types
//   - Collection types for phases, tasks, and artifacts
//   - Backend interface for storage abstraction
//   - YAML and memory backend implementations
//   - Load/Save operations with CUE validation
package state
