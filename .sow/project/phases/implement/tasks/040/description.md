# Task 040: Create basic validation utilities

**Phase**: implement
**Assigned Agent**: implementer
**Created**: 2025-10-12

## Objective

Create basic validation utilities to verify sow file structure integrity.
These utilities validate that state files match schemas and directory
structures are correct. Will be used by agents and potentially by CLI.

## Requirements

1. Create validation functions/scripts for:
   - `.sow/.version` format validation
   - Project `state.yaml` schema validation
   - Task `state.yaml` schema validation
   - Sink `index.json` validation
   - Repo `index.json` validation

2. Validation checks should verify:
   - Required fields present
   - Field types correct
   - Valid enum values (status, phase names)
   - Task ID format (gap numbering - "010", "020", etc.)
   - Timestamp formats (ISO 8601)
   - Branch name matches project.branch

3. Determine implementation approach:
   - Bash scripts (simple, no dependencies)
   - Go CLI utilities (if CLI being built)
   - Markdown documentation with validation rules (for agents)
   - Or combination

4. Create validation documentation:
   - What each validator checks
   - How to run validations
   - Error messages and meanings
   - How to fix common issues

## Context References

- `docs/SCHEMAS.md` - All schemas to validate against
- `docs/PROJECT_MANAGEMENT.md` - State file requirements
- Design task 030 output - Complete schemas

## Acceptance Criteria

- [ ] Validation utility for `.sow/.version`
- [ ] Validation utility for project `state.yaml`
- [ ] Validation utility for task `state.yaml`
- [ ] Validation utility for index files
- [ ] Clear error messages for validation failures
- [ ] Documentation on how to use validators
- [ ] Validators can be invoked by agents or users

## Notes

Keep validation utilities simple for now. Focus on catching obvious errors:
- Missing required fields
- Wrong field types
- Invalid status values
- Malformed timestamps

More sophisticated validation (cross-referencing, consistency checks) can come later.

Consider whether this should be:
1. Bash scripts that agents can call
2. Part of Go CLI (if building that now)
3. Just documentation for agents to validate manually

For Milestone 1, bash scripts or documentation may be sufficient.
Full CLI validation can come in Milestone 14 (CLI Development).
