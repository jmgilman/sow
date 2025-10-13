# Task: Implement management commands

## Objective

Implement `sow sinks` and `sow repos` commands for managing external knowledge sources and linked repositories.

## Context

These commands enable users to install style guides, link other repositories, and manage external knowledge sources.

## Requirements

### Command: `sow sinks`

1. **Subcommands**:
   - `sow sinks install <git-url> [--name <name>]` - Install sink from git
   - `sow sinks update [<name>]` - Update one or all sinks
   - `sow sinks list` - List installed sinks
   - `sow sinks remove <name>` - Remove a sink

2. **Functionality**:
   - Clone git repositories to `.sow/sinks/<name>`
   - Update `.sow/sinks/index.json` with metadata
   - Support shallow clones for performance
   - Validate sink structure

3. **Index Management**:
   - Create/update `.sow/sinks/index.json`
   - Track: name, git_url, installed_at, last_updated, version

### Command: `sow repos`

1. **Subcommands**:
   - `sow repos add <path|git-url> [--name <name>]` - Link repository
   - `sow repos sync [<name>]` - Sync one or all repos
   - `sow repos list` - List linked repos
   - `sow repos remove <name>` - Remove a repo link

2. **Functionality**:
   - Support both symlinks (local) and clones (git)
   - Store in `.sow/repos/<name>`
   - Update `.sow/repos/index.json` with metadata
   - Handle monorepo and multi-repo cases

3. **Index Management**:
   - Create/update `.sow/repos/index.json`
   - Track: name, type (symlink|clone), source, linked_at, last_synced

### Both Command Sets

- Clear help text and examples
- Error handling (network errors, permissions, etc.)
- Dry-run mode (`--dry-run`)
- Tests
- Git operations via Go git libraries or shell commands

## References

- `docs/CLI_REFERENCE.md` - Command specifications
- `docs/ARCHITECTURE.md` - Sinks and multi-repo design

## Deliverables

- [ ] `sow sinks` command with all subcommands
- [ ] `sow repos` command with all subcommands
- [ ] Index file management
- [ ] Git operations working
- [ ] Error handling
- [ ] Tests passing
