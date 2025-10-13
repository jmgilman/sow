# Task: Implement initialization commands

## Objective

Implement `sow init` and `sow schema` commands for repository initialization and schema inspection.

## Context

These commands help users set up `.sow/` structure and inspect embedded schemas.

## Requirements

### Command: `sow init`

1. **Functionality**:
   - Create `.sow/` directory structure
   - Create `.sow/knowledge/` directory
   - Create `.sow/.version` file with CLI version
   - Skip if `.sow/` already exists (with helpful message)

2. **Flags**:
   - `--force` - Recreate structure even if exists

3. **Output**:
   - Confirmation message
   - List of created directories

### Command: `sow schema`

1. **Functionality**:
   - List available schemas (default)
   - Show specific schema content with `--name`
   - Export schema to file with `--export`

2. **Subcommands/Flags**:
   - `sow schema` - List all schemas
   - `sow schema --name project-state` - Show schema
   - `sow schema --name project-state --export state.cue` - Export

3. **Output**:
   - Pretty-formatted schema display
   - Schema metadata (version, description)

### Both Commands

- Clear help text
- Examples in help
- Error handling
- Tests

## References

- `docs/CLI_REFERENCE.md` - Command specifications

## Deliverables

- [ ] `sow init` command implemented
- [ ] `sow schema` command implemented
- [ ] Help text and examples
- [ ] Error handling
- [ ] Tests passing
