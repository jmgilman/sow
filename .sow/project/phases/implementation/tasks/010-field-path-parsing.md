# Task 010: Field Path Parsing Utility (TDD)

## Overview

Create a field path parser that handles dot notation and automatically routes metadata fields. This is library code (not CLI commands), so it requires unit tests.

## Context

The unified CLI command structure uses field paths to modify both direct fields and metadata:
- Direct field: `approved` → sets artifact.Approved
- Metadata field: `metadata.assessment` → sets artifact.Metadata["assessment"]
- Nested metadata: `metadata.foo.bar` → sets artifact.Metadata["foo"]["bar"]

**Design Reference**: See `.sow/knowledge/designs/command-hierarchy-design.md` lines 217-236 for field path specification.

## Requirements

### Field Path Parser

Create a parser that:
1. Splits field paths on dots (e.g., `metadata.assessment` → ["metadata", "assessment"])
2. Determines if first segment is "metadata" → route to metadata map
3. Otherwise → access direct field on struct
4. Supports nested metadata access (multiple levels deep)
5. Provides type conversion (string → bool, int, string)
6. Returns clear errors for invalid paths

### Artifact Helpers

Create shared utilities for artifact operations:
1. Index-based access to artifact collections
2. Field validation (known vs unknown fields)
3. Type conversion helpers
4. Error message generation

## Implementation Approach

### TDD Workflow

1. Write tests in `cli/internal/cmdutil/fieldpath_test.go`
2. Implement parser in `cli/internal/cmdutil/fieldpath.go`
3. Write tests in `cli/internal/cmdutil/artifacts_test.go`
4. Implement helpers in `cli/internal/cmdutil/artifacts.go`
5. Ensure all tests pass

### Known Fields by Type

**Project**:
- Direct: name, type, branch, description, created_at, updated_at
- Metadata: everything else → metadata map

**Phase**:
- Direct: status, enabled, created_at, started_at, completed_at
- Metadata: everything else → metadata map

**Artifact**:
- Direct: type, path, approved, created_at
- Metadata: everything else → metadata map

**Task**:
- Direct: id, name, phase, status, iteration, assigned_agent, created_at, started_at, updated_at, completed_at
- Metadata: everything else → metadata map

### Type Conversion

Support these conversions from string input:
- `"true"` / `"false"` → bool
- `"123"` → int
- Any other string → string (passthrough)

### Example Usage

```go
// Example: Setting artifact approved field
artifact := &Artifact{Type: "review", Path: "review.md", Approved: false}
err := SetField(artifact, "approved", "true")
// Result: artifact.Approved = true

// Example: Setting artifact metadata
artifact := &Artifact{Type: "review", Path: "review.md"}
err := SetField(artifact, "metadata.assessment", "pass")
// Result: artifact.Metadata["assessment"] = "pass"

// Example: Nested metadata
artifact := &Artifact{Type: "review", Path: "review.md"}
err := SetField(artifact, "metadata.foo.bar", "baz")
// Result: artifact.Metadata["foo"]["bar"] = "baz"
```

## Files to Create

### `cli/internal/cmdutil/fieldpath.go`

Core functions needed:
```go
// ParseFieldPath splits a field path into segments
func ParseFieldPath(path string) []string

// IsMetadataPath checks if path routes to metadata
func IsMetadataPath(segments []string) bool

// SetField sets a field value on a struct using field path
func SetField(target interface{}, fieldPath string, value string) error

// GetField retrieves a field value using field path
func GetField(target interface{}, fieldPath string) (interface{}, error)

// ConvertValue converts string to appropriate type (bool, int, string)
func ConvertValue(value string) interface{}
```

### `cli/internal/cmdutil/fieldpath_test.go`

Test cases:
- Parse simple path: `"approved"` → `["approved"]`
- Parse metadata path: `"metadata.assessment"` → `["metadata", "assessment"]`
- Parse nested path: `"metadata.foo.bar"` → `["metadata", "foo", "bar"]`
- Set direct field (approved)
- Set metadata field (metadata.assessment)
- Set nested metadata (metadata.foo.bar)
- Type conversion: string → bool, int, string
- Error: invalid field path
- Error: field doesn't exist
- Error: type mismatch

### `cli/internal/cmdutil/artifacts.go`

Helper functions:
```go
// ValidateArtifactType validates type against allowed types for phase
func ValidateArtifactType(projectConfig *ProjectTypeConfig, phaseName string, direction string, artifactType string) error

// CreateArtifact creates artifact from flags
func CreateArtifact(artifactType string, fields map[string]string) (Artifact, error)

// FormatArtifactList formats artifacts for display
func FormatArtifactList(artifacts []Artifact) string
```

### `cli/internal/cmdutil/artifacts_test.go`

Test cases:
- Validate artifact type (valid)
- Validate artifact type (invalid)
- Create artifact with fields
- Format artifact list for display

## Acceptance Criteria

- [ ] Field path parser splits on dots correctly
- [ ] Automatic routing: `metadata.*` → metadata map
- [ ] Direct fields accessed correctly
- [ ] Nested metadata (multiple levels) supported
- [ ] Type conversion works (bool, int, string)
- [ ] Clear error messages for invalid paths
- [ ] All unit tests passing
- [ ] Code coverage > 90% for new code

## Testing Strategy

This is library code, so comprehensive unit tests are required:

1. **Field path parsing tests** - All parsing logic
2. **Field setting tests** - Direct and metadata paths
3. **Type conversion tests** - All supported types
4. **Error handling tests** - All failure modes
5. **Artifact helper tests** - Validation and formatting

## Dependencies

None - this is the foundation task.

## References

- **Command hierarchy design**: `.sow/knowledge/designs/command-hierarchy-design.md` (lines 217-236)
- **SDK state types**: `cli/internal/sdks/project/state/` (for struct definitions)
- **CUE schemas**: `cli/schemas/project/*.cue` (for field names)

## Notes

- This is pure library code - no CLI commands involved
- Must work with all container types (Project, Phase, Artifact, Task)
- Field paths are case-sensitive
- Metadata maps are created automatically if they don't exist
