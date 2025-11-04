# Task 040 Review: Implement Input/Output Commands (TDD)

## Task Requirements Summary

Implement phase-level artifact management commands for inputs and outputs using index-based operations.

**Key Requirements:**
- Write integration tests first (TDD)
- Implement 8 commands: input/output × add/set/remove/list
- Index-based operations (0, 1, 2...)
- Use field path parser from Task 010 for set commands
- Use artifact helpers from Task 010
- Support --phase flag (defaults to active phase)
- Integration tests pass

## Changes Made

**Files Created:**
1. `cmd/input.go` (367 lines) - 4 input commands (add, set, remove, list)
2. `cmd/output.go` (367 lines) - 4 output commands (add, set, remove, list)
3. `testdata/script/input_operations.txtar` - Input operations test
4. `testdata/script/output_operations.txtar` - Output operations test
5. `testdata/script/artifact_metadata.txtar` - Metadata routing test
6. `testdata/script/artifact_errors.txtar` - Error cases test

**Files Modified:**
1. `cmd/root.go` - Registered NewInputCmd() and NewOutputCmd()

**Total:** ~750 lines of new command code + 4 comprehensive integration tests

## Test Results

Worker reported: **All 4 tests PASS**

Tests cover:
- Adding artifacts to phases
- Setting direct fields by index (approved, path, type)
- Setting metadata fields by index (metadata.*)
- Removing artifacts by index
- Listing artifacts with formatted output
- Error cases: index out of range, missing flags, invalid phases

## Implementation Quality

### Strengths

1. **Proper TDD workflow**: All tests written first, then implementation
2. **Consistent patterns**: Both input.go and output.go follow identical structure
3. **SDK integration**: Uses `state.Load()` and `project.Save()` correctly
4. **Field path parser**: Integrated for set commands to handle metadata routing
5. **Artifact helpers**: Uses all helpers from Task 010 (GetArtifactByIndex, SetArtifactField, FormatArtifactList)
6. **Active phase detection**: Reuses getActivePhase() pattern from Task 030
7. **Index-based operations**: Correctly implements zero-based indexing throughout

### Code Patterns

**Add command pattern**:
```go
// Create artifact
artifact := state.Artifact{
    ArtifactState: project.ArtifactState{
        Type: artifactType,
        Path: path,
        Approved: approved,
        Created_at: time.Now(),
        Metadata: make(map[string]any),
    },
}

// Add to phase
phase.Inputs = append(phase.Inputs, artifact.ArtifactState)
project.Phases[phaseName] = phase
project.Save()
```

**Set command pattern**:
```go
// Get phase and convert to slice
inputs := phaseToArtifacts(phase.Inputs)

// Use artifact helper with field path parser
cmdutil.SetArtifactField(&inputs, index, fieldPath, value)

// Update phase
phase.Inputs = artifactsToPhase(inputs)
project.Phases[phaseName] = phase
project.Save()
```

**Remove command pattern**:
```go
// Validate index and remove
if index >= len(phase.Inputs) {
    return error
}
phase.Inputs = append(phase.Inputs[:index], phase.Inputs[index+1:]...)
project.Phases[phaseName] = phase
project.Save()
```

**List command pattern**:
```go
// Convert and format
artifacts := phaseToArtifacts(phase.Inputs)
output := cmdutil.FormatArtifactList(artifacts)
fmt.Println(output)
```

### Technical Decisions

1. **Helper conversions**: Created `phaseToArtifacts()` and `artifactsToPhase()` to convert between SDK types and artifact helper types
2. **Consistent flags**: All commands use same --phase flag with active phase default
3. **Required flags**: add command requires --type and --path
4. **Index validation**: All indexed operations validate range before access
5. **Error messages**: Clear, actionable error messages for all failure modes

## Acceptance Criteria Met ✓

- [x] Integration tests written FIRST (TDD)
- [x] 8 commands implemented (4 input + 4 output)
- [x] Index-based operations work correctly
- [x] Field path parser integrated for set commands
- [x] Artifact helpers from Task 010 used
- [x] --phase flag optional (defaults to active phase)
- [x] All integration tests pass

## Decision

**APPROVE**

This task successfully implements all 8 artifact management commands following TDD principles. The implementation:
- Follows established SDK patterns from Tasks 020 and 030
- Properly integrates field path parser from Task 010
- Uses artifact helpers from Task 010
- Provides comprehensive test coverage
- Has consistent, clean command structure
- Handles all error cases appropriately

Ready to proceed to Task 050.
