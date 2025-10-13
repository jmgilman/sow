# Task: Convert YAML templates to CUE schemas with validation rules

## Objective

Convert the existing YAML/JSON templates in `schemas/templates/` to CUE schema definitions with proper validation rules and constraints.

## Context

The repository has YAML/JSON template files but no CUE schemas. The CLI (being built in this milestone) will embed CUE schemas for validation. You need to create these CUE definitions.

## Requirements

1. **Create CUE schema files** in `schemas/cue/` directory:
   - `project-state.cue` - Project state schema
   - `task-state.cue` - Task state schema
   - `sink-index.cue` - Sink index schema
   - `repo-index.cue` - Repo index schema
   - `sow-version.cue` - Version file schema

2. **Add validation constraints**:
   - Enum validation for status fields (pending, in_progress, completed, abandoned)
   - Enum validation for agent assignments (architect, implementer, integration-tester, reviewer, documenter)
   - Enum validation for phase names (discovery, design, implement, test, review, deploy, document)
   - Required fields enforcement
   - Format validation (ISO 8601 timestamps, kebab-case naming)
   - Numeric constraints (complexity rating 1-3, iteration >= 1)

3. **Maintain compatibility** with existing YAML templates

4. **Document schema decisions** in comments within CUE files

## References

- `schemas/templates/*.yaml` and `*.json` - Current templates
- `docs/SCHEMAS.md` - Schema specifications
- `docs/CLI_REFERENCE.md` - CLI validation requirements

## Deliverables

- [ ] CUE schema files in `schemas/cue/`
- [ ] Validation rules implemented
- [ ] Comments documenting constraints
- [ ] Compatibility verified with templates
