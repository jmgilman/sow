# Task 020: Intra-Phase State Progression Command

# Intra-Phase State Progression Command

## Overview

Implement the `sow agent advance` command infrastructure to support intra-phase state progression. Add `Advance()` method to Phase interface, create the CLI command, and update all existing standard project phases to implement the method.

## Design Reference

**Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Intra-Phase State Progression"
- See "Command Interface" for CLI implementation details
- See "Phase Interface Extension" for `Advance()` method signature and semantics
- Reference implementation pattern from existing phase operations that return `PhaseOperationResult`

## Objectives

1. Add `Advance()` method to Phase interface
2. Implement `sow agent advance` CLI command
3. Update all existing standard phases to implement the method
4. Write command and error handling tests

## Files to Create

- `cli/cmd/agent/advance.go` - New command implementation

## Files to Modify

- `cli/internal/project/domain/phase.go` - Add `Advance()` method to Phase interface
- `cli/cmd/agent/agent.go` - Register advance command
- All existing phase implementations in `cli/internal/project/standard/` - Implement `Advance()` returning `ErrNotSupported`

## Implementation Details

### 1. Add Advance() to Phase Interface

```go
// cli/internal/project/domain/phase.go
type Phase interface {
    // ... existing methods

    // Advance to next state within this phase
    // Returns ErrNotSupported if phase has no internal states
    Advance() (*PhaseOperationResult, error)
}
```

### 2. Create advance Command

```go
// cli/cmd/agent/advance.go
package agent

import (
    "errors"
    "fmt"

    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/project"
    "github.com/jmgilman/sow/cli/internal/project/loader"
    "github.com/spf13/cobra"
)

func NewAdvanceCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "advance",
        Short: "Advance to next state within current phase",
        Long:  "Advance to the next state within the current phase (intra-phase transition)",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmdutil.GetContext(cmd.Context())
            proj, err := loader.Load(ctx)
            if err != nil {
                return err
            }

            phase := proj.CurrentPhase()
            result, err := phase.Advance()
            if errors.Is(err, project.ErrNotSupported) {
                return fmt.Errorf("phase %s does not support state advancement", phase.Name())
            }
            if err != nil {
                return fmt.Errorf("failed to advance: %w", err)
            }

            // Fire event if returned
            if result.Event != "" {
                machine := proj.Machine()
                if err := machine.Fire(result.Event); err != nil {
                    return fmt.Errorf("failed to fire event: %w", err)
                }
                // Save after transition
                if err := proj.Save(); err != nil {
                    return fmt.Errorf("failed to save: %w", err)
                }
            }

            cmd.Println("✓ Advanced to next state")
            return nil
        },
    }
    return cmd
}
```

### 3. Register Command

```go
// cli/cmd/agent/agent.go
func NewAgentCmd() *cobra.Command {
    // ... existing code
    cmd.AddCommand(NewAdvanceCmd())  // ADD THIS
    return cmd
}
```

### 4. Update Existing Phases

For each phase in `cli/internal/project/standard/`:

```go
func (p *PlanningPhase) Advance() (*domain.PhaseOperationResult, error) {
    return nil, project.ErrNotSupported
}
```

## Acceptance Criteria

- [ ] `Phase.Advance()` method added to interface with signature `(*PhaseOperationResult, error)`
- [ ] `sow agent advance` command exists and is registered
- [ ] Command loads current project and calls `phase.Advance()`
- [ ] Command handles `ErrNotSupported` with clear message
- [ ] Command fires events from `PhaseOperationResult` when returned
- [ ] Command saves state after successful advance
- [ ] All existing standard project phases implement `Advance()` returning `(nil, project.ErrNotSupported)`
- [ ] Tests verify `ErrNotSupported` handling
- [ ] Tests verify event firing on successful advance
- [ ] Tests verify state persistence after advance

## Testing

Write tests to verify:
- Phases without internal states return `ErrNotSupported`
- Error handling works correctly
- Event firing works when result contains event
- State persists after successful advance

## Important Notes

- Standard project phases don't have internal states yet, so they return `ErrNotSupported`
- Future project types (exploration, design, breakdown) will implement real state transitions
- This provides the infrastructure for those future implementations

## Dependencies

Task 010 (Go types must be regenerated)
