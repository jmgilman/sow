# Task 030: Implement Phase Commands (TDD)

## Overview

Implement phase-level operations with support for direct field mutation and metadata via dot notation.

## Context

Phases are accessed via the SDK's `project.Phases.Get(phaseName)` method, which returns a phase pointer for direct field mutation. The field path parser (from Task 010) handles routing metadata fields automatically.

## Design References

- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 380-385, 512-533
- **SDK phase type**: `cli/internal/sdks/project/state/phase.go`
- **Phase collection**: `cli/internal/sdks/project/state/collections.go` lines 7-17

## Requirements

### Command to Implement

#### `sow phase set`

**Syntax**: `sow phase set <field-path> <value> [--phase <name>]`

**Behavior**:
- Load project state
- Navigate to phase (use --phase or default to active phase)
- Set field using field path parser
- Save state

**Supported fields**:
- Direct: `status`, `enabled`
- Metadata: `metadata.*` (any custom field)

**Examples**:
```bash
# Set direct field
sow phase set status in_progress

# Set direct field with explicit phase
sow phase set enabled false --phase planning

# Set metadata field (used for state machine guards)
sow phase set metadata.tasks_approved true --phase implementation

# Set custom metadata
sow phase set metadata.complexity high
```

## TDD Approach

### Step 1: Write Integration Tests First

Create two test files:

**`cli/testdata/script/unified_commands/phase/phase_operations.txtar`** - Happy path:

```txtar
# Test: Phase Set Operations
# Coverage: set direct fields, set metadata fields

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test phase commands"

# Test: Set phase status (direct field)
exec sow phase set status in_progress --phase planning
exec cat .sow/project/state.yaml
stdout 'status: in_progress'

# Test: Set phase enabled (direct field)
exec sow phase set enabled false --phase planning
exec cat .sow/project/state.yaml
stdout 'enabled: false'

# Test: Set phase metadata
exec sow phase set metadata.tasks_approved true --phase implementation
exec cat .sow/project/state.yaml
stdout 'tasks_approved: true'

# Test: Default to active phase (no --phase flag)
# Active phase is planning (based on current state)
exec sow phase set metadata.custom_field test_value
exec cat .sow/project/state.yaml
stdout 'custom_field: test_value'
```

**`cli/testdata/script/unified_commands/phase/phase_errors.txtar`** - Error cases:

```txtar
# Test: Phase Set Error Cases
# Coverage: invalid phase, invalid field, no project

exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test
exec sow init
exec sow project new --branch feat/test --no-launch "Test phase errors"

# Test: Error - invalid phase name
! exec sow phase set status in_progress --phase nonexistent
stderr 'phase not found: nonexistent'

# Test: Error - invalid field path
! exec sow phase set invalid.nested.path value
stderr 'invalid field path'
```

### Step 2: Implement Command

Create `cli/cmd/phase.go`.

### Step 3: Run Integration Tests

Verify tests pass.

## Implementation Details

### File Structure

```
cli/cmd/
└── phase.go          # sow phase set command
```

### Command Implementation

```go
package cmd

import (
    "fmt"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/spf13/cobra"
)

func NewPhaseCmd() *cobra.Command {
    var phaseName string

    cmd := &cobra.Command{
        Use:   "phase",
        Short: "Manage phases",
    }

    setCmd := &cobra.Command{
        Use:   "set <field-path> <value>",
        Short: "Set phase field",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runPhaseSet(cmd, args, phaseName)
        },
    }

    setCmd.Flags().StringVar(&phaseName, "phase", "", "Target phase (defaults to active phase)")

    cmd.AddCommand(setCmd)
    return cmd
}

func runPhaseSet(cmd *cobra.Command, args []string, phaseName string) error {
    ctx := cmdutil.GetContext(cmd.Context())

    // Load project
    project, err := state.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // Determine target phase
    if phaseName == "" {
        phaseName = getActivePhase(project)
    }

    // Get phase
    phase, err := project.Phases.Get(phaseName)
    if err != nil {
        return err
    }

    // Set field
    fieldPath := args[0]
    value := args[1]

    if err := cmdutil.SetField(phase, fieldPath, value); err != nil {
        return fmt.Errorf("failed to set field: %w", err)
    }

    // Save
    return project.Save()
}

func getActivePhase(project *state.Project) string {
    // Determine active phase based on current state
    // This logic matches the state machine's current phase
    currentState := project.Statechart.Current_state

    // Map states to phases (simplified - adjust based on actual states)
    if strings.HasPrefix(currentState, "Planning") {
        return "planning"
    } else if strings.HasPrefix(currentState, "Implementation") {
        return "implementation"
    } else if strings.HasPrefix(currentState, "Review") {
        return "review"
    } else if strings.HasPrefix(currentState, "Finalize") {
        return "finalize"
    }

    return "planning" // Default fallback
}
```

### Determining Active Phase

The active phase is the phase corresponding to the current state machine state:

- `PlanningActive` → `planning`
- `ImplementationPlanning`, `ImplementationExecuting` → `implementation`
- `ReviewActive` → `review`
- `FinalizeDocumentation`, `FinalizeChecks`, `FinalizeDelete` → `finalize`

You may need to reference the state machine configuration to map states to phases correctly. See `cli/internal/projects/standard/states.go` for state definitions.

### Field Path Routing

The field path parser (from Task 010) automatically handles:

- `status` → Direct field `phase.Status`
- `enabled` → Direct field `phase.Enabled`
- `metadata.tasks_approved` → Metadata `phase.Metadata["tasks_approved"]`
- `metadata.foo.bar` → Nested metadata `phase.Metadata["foo"]["bar"]`

### Phase Metadata Usage

Metadata fields are used by state machine guards:

- `metadata.tasks_approved` - Guard for transitioning from ImplementationPlanning to ImplementationExecuting
- `metadata.project_deleted` - Guard for completing finalize phase

See `.sow/knowledge/designs/command-hierarchy-design.md` lines 147-159 for metadata patterns.

## Files to Create

### `cli/cmd/phase.go`

Phase command implementation.

### `cli/testdata/script/unified_commands/phase/phase_operations.txtar`

Integration test - happy path.

### `cli/testdata/script/unified_commands/phase/phase_errors.txtar`

Integration test - error cases.

## Acceptance Criteria

- [ ] Integration tests written first
- [ ] Phase set command modifies direct fields
- [ ] Phase set command modifies metadata fields
- [ ] --phase flag optional (defaults to active phase)
- [ ] Dot notation routes to metadata automatically
- [ ] Clear error when phase not found
- [ ] Clear error when field path invalid
- [ ] Integration tests pass

## Testing Strategy

**Integration tests only** - No unit tests for command logic.

Test scenarios:
1. Set phase status (direct field)
2. Set phase enabled (direct field)
3. Set phase metadata field
4. Default to active phase (no --phase flag)
5. Error: invalid phase name
6. Error: invalid field path

## Dependencies

- Task 010 (Field Path Parsing) - Required for field routing
- Task 020 (Project Commands) - Required for project state

## References

- **SDK phase type**: `cli/internal/sdks/project/state/phase.go`
- **Phase collection**: `cli/internal/sdks/project/state/collections.go`
- **Standard project states**: `cli/internal/projects/standard/states.go`
- **Command spec**: `.sow/knowledge/designs/command-hierarchy-design.md`

## Notes

- Phases are read-only in the state machine - they don't have independent state transitions
- Phase status is updated automatically by state machine transitions (OnEntry/OnExit actions)
- This command is primarily for setting metadata used in guards
- Direct field mutations (status, enabled) should be used carefully to avoid breaking state machine invariants
