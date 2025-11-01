# Task 040: Integration Testing and Backward Compatibility Verification

# Integration Testing and Backward Compatibility Verification

## Overview

Verify that all infrastructure changes work correctly together and that existing standard project functionality remains unchanged. Run comprehensive integration tests to ensure backward compatibility.

## Design Reference

**Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Testing Strategy"
- See "Backward Compatibility" for specific compatibility guarantees to verify
- See "Migration Path" for understanding how existing projects should continue working
- All previous sections provide context for what infrastructure should be tested

## Objectives

1. Verify all existing tests pass without modification
2. Test standard project creation and operation
3. Verify backward compatibility guarantees
4. Test new infrastructure in isolation
5. Write integration tests for end-to-end workflows

## Testing Focus

- **Existing standard project functionality** - Nothing breaks
- **New infrastructure in isolation** - Schema extensions, advance command, type detection work correctly
- **Backward compatibility guarantees** - Optional fields, default type handling, error messages

## What to Test

### 1. Existing Tests Pass

- Run the full test suite
- Verify all existing tests pass without modification
- No breaking changes introduced

### 2. Standard Project Workflows

- Create a new standard project successfully
- Load an existing standard project successfully
- Run through standard project lifecycle
- Verify state management works correctly

### 3. New Infrastructure

**Schema Extensions:**
- Optional fields can be null
- CUE validation passes
- Go types match expectations

**Advance Command:**
- `sow agent advance` command exists
- Running on standard project phase returns appropriate `ErrNotSupported` message
- Error handling works correctly

**Type Detection:**
- `DetectProjectType()` maps branch prefixes correctly
- Unknown prefixes default to "standard"
- Loader routes based on discriminator
- Helpful errors for unimplemented types

### 4. Backward Compatibility

- Schema extensions don't break existing projects (all fields optional)
- Unknown branch prefixes default to "standard" type
- Standard projects create and load normally
- No data migration required

## Acceptance Criteria

- [ ] All existing tests pass without modification
- [ ] Can create a new standard project successfully
- [ ] Can load an existing standard project successfully
- [ ] Running `sow agent advance` on standard project phase returns appropriate `ErrNotSupported` message
- [ ] Schema extensions don't break existing projects (optional fields)
- [ ] Integration tests verify end-to-end workflows
- [ ] No breaking changes introduced for existing users

## Integration Test Scenarios

### Scenario 1: Standard Project Lifecycle
1. Create new standard project on `feat/test` branch
2. Verify project state initialized correctly
3. Run through phases (planning → implementation → review → finalize)
4. Verify state transitions work
5. Clean up project

### Scenario 2: Schema Extensions
1. Load project with new optional fields
2. Verify fields can be null
3. Verify fields can have values
4. Verify CUE validation passes

### Scenario 3: Advance Command
1. Create standard project
2. Run `sow agent advance`
3. Verify `ErrNotSupported` returned with clear message
4. Verify no state changes occurred

### Scenario 4: Type Detection
1. Test branch prefix mapping:
   - `explore/test` → "exploration"
   - `design/test` → "design"
   - `breakdown/test` → "breakdown"
   - `feat/test` → "standard"
2. Verify unknown prefixes default to "standard"

### Scenario 5: Loader Routing
1. Create standard project
2. Save state
3. Load project
4. Verify correct type loaded
5. Verify error messages for unimplemented types

## Success Indicators

✓ All existing tests pass unchanged
✓ Standard projects create, load, and operate normally
✓ `go generate` completes successfully
✓ `sow agent advance` command exists and handles errors correctly
✓ `DetectProjectType()` correctly maps all branch prefix patterns
✓ Loader has routing infrastructure with placeholder errors for unimplemented types

## Important Notes

- **Zero tolerance for breaking changes** - existing functionality must work unchanged
- **Comprehensive coverage** - test all new infrastructure
- **Clear errors** - verify error messages are helpful for users
- **Backward compatibility** - optional fields, default type handling

## Dependencies

Tasks 010, 020, 030 (all infrastructure must be implemented)
