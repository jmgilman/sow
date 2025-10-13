# Task: Implement core validation engine

## Objective

Build the core validation engine that loads CUE schemas and validates YAML/JSON files against them.

## Context

This is the heart of the CLI - the validation system that all commands will use. Must be performant (<1s) and provide clear error messages.

## Requirements

1. **Create validation package**:
   - `internal/validation/` or similar
   - Load embedded CUE schemas
   - Parse YAML/JSON input files
   - Validate against CUE schemas

2. **Implement validation functions**:
   - `ValidateProjectState(filePath string) error`
   - `ValidateTaskState(filePath string) error`
   - `ValidateSinkIndex(filePath string) error`
   - `ValidateRepoIndex(filePath string) error`
   - `ValidateVersion(filePath string) error`

3. **Error handling**:
   - Clear, actionable error messages
   - Show which field failed validation
   - Show expected vs actual values
   - Line numbers if possible

4. **Performance**:
   - Validation should complete in <1s
   - Cache loaded schemas if beneficial

5. **Write tests**:
   - Test with valid files (from templates)
   - Test with invalid files (various error cases)
   - Test error message clarity

## References

- `schemas/cue/` - CUE schema definitions
- `docs/SCHEMAS.md` - Schema specifications
- `schemas/templates/` - Valid examples to test against

## Deliverables

- [ ] Validation package created
- [ ] All validation functions implemented
- [ ] Clear error messages
- [ ] Performance < 1s
- [ ] Tests passing
