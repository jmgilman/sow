# Task 040: Implement Input/Output Commands (TDD)

# Task 040: Implement Input/Output Commands (TDD)

## Overview

Implement phase-level artifact management commands for inputs and outputs using index-based operations.

## Context

Artifacts are stored in phase-level `inputs` and `outputs` collections. Operations use indices (0, 1, 2, ...) rather than IDs. The SDK provides collection methods: `Get(index)`, `Add(artifact)`, `Remove(index)`.

## Design References

- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 164-189, 397-738
- **Artifact schema**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 204-215, 347-360
- **SDK collections**: `cli/internal/sdks/project/state/collections.go` lines 19-48
- **Artifact type**: `cli/internal/sdks/project/state/artifact.go`

## Requirements

### Commands to Implement

#### Input Commands

```bash
sow input add --type <type> [--phase <name>] [--path <path>] [--approved <bool>]
sow input set --index <n> <field-path> <value> [--phase <name>]
sow input remove --index <n> [--phase <name>]
sow input list [--phase <name>]
```

#### Output Commands

```bash
sow output add --type <type> [--phase <name>] [--path <path>] [--approved <bool>]
sow output set --index <n> <field-path> <value> [--phase <name>]
sow output remove --index <n> [--phase <name>]
sow output list [--phase <name>]
```

### Artifact Fields

**Required**:
- `type` - Artifact type (validated per project type config)
- `path` - File path relative to `.sow/project/`

**Optional**:
- `approved` - Boolean approval flag (default: false)
- `metadata.*` - Custom metadata fields

**Auto-set**:
- `created_at` - Timestamp

### Artifact Types for Standard Project

**Phase Input Types**:
- Planning: `context`
- Implementation: (none)
- Review: (none)
- Finalize: (none)

**Phase Output Types**:
- Planning: `task_list`
- Implementation: (none)
- Review: `review`
- Finalize: (none)

See `.sow/knowledge/designs/command-hierarchy-design.md` lines 894-946 for complete type definitions.

## TDD Approach

### Step 1: Write Integration Tests First

Create four test files:

**`cli/testdata/script/unified_commands/artifacts/input_operations.txtar`**:

```txtar
# Test: Input Operations
# Coverage: add, set, remove, list

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test inputs"

# Create test file
exec mkdir -p .sow/project/context
exec sh -c 'echo "Test context" > .sow/project/context/research.md'

# Test: Add input
exec sow input add --type context --path context/research.md --phase planning
exec cat .sow/project/state.yaml
stdout 'type: context'
stdout 'path: context/research.md'
stdout 'approved: false'

# Test: List inputs (shows index)
exec sow input list --phase planning
stdout '\[0\] context: context/research.md \(not approved\)'

# Test: Set approved field
exec sow input set --index 0 approved true --phase planning
exec cat .sow/project/state.yaml
stdout 'approved: true'

# Test: Set metadata field
exec sow input set --index 0 metadata.source external --phase planning
exec cat .sow/project/state.yaml
stdout 'source: external'

# Test: Remove input
exec sow input remove --index 0 --phase planning
exec sow input list --phase planning
! stdout 'context'
```

**`cli/testdata/script/unified_commands/artifacts/output_operations.txtar`**:

Similar to input_operations but for outputs.

**`cli/testdata/script/unified_commands/artifacts/artifact_metadata.txtar`**:

```txtar
# Test: Artifact Metadata
# Coverage: metadata field routing, nested metadata

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test metadata"

# Create review artifact
exec mkdir -p .sow/project/review
exec sh -c 'echo "Review report" > .sow/project/review/report.md'

# Add output with inline metadata
exec sow output add --type review --path review/report.md --phase review
exec sow output set --index 0 metadata.assessment pass --phase review
exec cat .sow/project/state.yaml
stdout 'assessment: pass'

# Nested metadata
exec sow output set --index 0 metadata.reviewer.name orchestrator --phase review
exec cat .sow/project/state.yaml
stdout 'reviewer:'
stdout 'name: orchestrator'
```

**`cli/testdata/script/unified_commands/artifacts/artifact_validation.txtar`**:

```txtar
# Test: Artifact Type Validation
# Coverage: valid types, invalid types

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test validation"

# Test: Valid type for planning output
exec sow output add --type task_list --path planning/tasks.md --phase planning
exec cat .sow/project/state.yaml
stdout 'type: task_list'

# Test: Invalid type for planning output
! exec sow output add --type invalid_type --path test.md --phase planning
stderr 'invalid artifact type.*planning.*output'
stderr 'allowed.*task_list'

# Test: Invalid type for planning input
! exec sow input add --type task_list --path test.md --phase planning
stderr 'invalid artifact type.*planning.*input'
stderr 'allowed.*context'
```

### Step 2: Implement Commands

Create `cli/cmd/input.go` and `cli/cmd/output.go`.

### Step 3: Run Integration Tests

Verify all tests pass.

## Implementation Details

### File Structure

```
cli/cmd/
├── input.go          # sow input add/set/remove/list
└── output.go         # sow output add/set/remove/list
```

### Command Implementation Pattern

Both input and output follow identical patterns:

```go
package cmd

import (
    "fmt"
    "time"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/spf13/cobra"
)

func NewInputCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "input",
        Short: "Manage phase inputs",
    }

    cmd.AddCommand(newInputAddCmd())
    cmd.AddCommand(newInputSetCmd())
    cmd.AddCommand(newInputRemoveCmd())
    cmd.AddCommand(newInputListCmd())

    return cmd
}

func newInputAddCmd() *cobra.Command {
    var phaseName, artifactType, path string
    var approved bool

    cmd := &cobra.Command{
        Use:   "add",
        Short: "Add input artifact to phase",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runInputAdd(cmd, phaseName, artifactType, path, approved)
        },
    }

    cmd.Flags().StringVar(&phaseName, "phase", "", "Target phase (defaults to active)")
    cmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (required)")
    cmd.Flags().StringVar(&path, "path", "", "Artifact path (required)")
    cmd.Flags().BoolVar(&approved, "approved", false, "Approval status")
    cmd.MarkFlagRequired("type")
    cmd.MarkFlagRequired("path")

    return cmd
}

func runInputAdd(cmd *cobra.Command, phaseName, artifactType, path string, approved bool) error {
    ctx := cmdutil.GetContext(cmd.Context())

    // Load project
    project, err := state.Load(ctx)
    if err != nil {
        return err
    }

    // Determine phase
    if phaseName == "" {
        phaseName = getActivePhase(project)
    }

    // Get phase
    phase, err := project.Phases.Get(phaseName)
    if err != nil {
        return err
    }

    // Validate artifact type
    if err := cmdutil.ValidateArtifactType(project.config, phaseName, "input", artifactType); err != nil {
        return err
    }

    // Create artifact
    artifact := state.Artifact{
        ArtifactState: schemas.ArtifactState{
            Type:      artifactType,
            Path:      path,
            Approved:  approved,
            CreatedAt: time.Now(),
            Metadata:  make(map[string]interface{}),
        },
    }

    // Add to inputs
    if err := phase.Inputs.Add(artifact); err != nil {
        return err
    }

    // Save
    return project.Save()
}
```

Similar implementations for:
- `set` - Uses field path parser
- `remove` - Uses collection.Remove(index)
- `list` - Formats output with indices

### Output List Format

```
[0] task_list: planning/tasks.md (not approved)
[1] design_doc: planning/design.md (approved)
```

### Artifact Type Validation

Use the helper from Task 010:

```go
err := cmdutil.ValidateArtifactType(
    project.config,
    phaseName,
    "input",  // or "output"
    artifactType,
)
```

This checks against the project type's allowed types configuration.

### Index-Based Operations

Collections use zero-based indexing:

```go
// Get artifact by index
artifact, err := phase.Inputs.Get(0)

// Remove by index
err := phase.Inputs.Remove(0)
```

Out-of-range indices return clear errors.

## Files to Create

### `cli/cmd/input.go`

Input command implementation (add, set, remove, list).

### `cli/cmd/output.go`

Output command implementation (add, set, remove, list).

### Integration Test Files

- `cli/testdata/script/unified_commands/artifacts/input_operations.txtar`
- `cli/testdata/script/unified_commands/artifacts/output_operations.txtar`
- `cli/testdata/script/unified_commands/artifacts/artifact_metadata.txtar`
- `cli/testdata/script/unified_commands/artifacts/artifact_validation.txtar`

## Acceptance Criteria

- [ ] Integration tests written first
- [ ] Input add creates artifact with fields
- [ ] Output add creates artifact with fields
- [ ] Set modifies fields by index (direct and metadata)
- [ ] Remove deletes by index
- [ ] List shows indices and artifact details
- [ ] Type validation rejects invalid types
- [ ] Clear error messages for validation failures
- [ ] Index out-of-range errors clear
- [ ] --phase defaults to active phase
- [ ] Integration tests pass

## Testing Strategy

**Integration tests only** - No unit tests for command logic.

Test scenarios (per input/output):
1. Add artifact with type and path
2. Set approved field
3. Set metadata field
4. List artifacts (verify indices)
5. Remove artifact
6. Error: invalid artifact type
7. Error: index out of range
8. Error: missing required flags

## Dependencies

- Task 010 (Field Path Parsing) - For set operations
- Task 020 (Project Commands) - For project state
- Task 030 (Phase Commands) - For phase navigation

## References

- **Artifact schema**: `cli/schemas/project/artifact.cue`
- **SDK artifact type**: `cli/internal/sdks/project/state/artifact.go`
- **Collections**: `cli/internal/sdks/project/state/collections.go`
- **Project type config**: `cli/internal/projects/standard/standard.go`
- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md`

## Notes

- Indices are visible in state files - users can verify manually
- Artifact paths are relative to `.sow/project/`
- Approval workflow: create (not approved) → set approved true → advance
- Metadata fields are type-specific (e.g., `assessment` for review artifacts)
