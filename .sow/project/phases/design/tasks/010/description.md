# Task 010: Design file structure for plugin/ directory (execution layer)

**Phase**: design
**Assigned Agent**: architect
**Created**: 2025-10-12

## Objective

Design the complete file structure for the `plugin/` directory (execution layer source).
This directory contains all AI agent behaviors, commands, hooks, and MCP integrations
that will be distributed via the Claude Code Plugin marketplace.

When users install the plugin, the contents of `plugin/` will be copied to `.claude/`
in their repository.

## Requirements

1. Define directory structure for:
   - Agent definitions (`plugin/agents/`)
   - Slash commands (`plugin/commands/`)
     - User workflow commands (`plugin/commands/workflows/`)
     - Agent skills (`plugin/commands/skills/`)
   - Hooks configuration (`plugin/hooks.json`)
   - MCP integrations (`plugin/mcp.json`)
   - Plugin metadata (`plugin/.claude-plugin/plugin.json`)
   - Version tracking (`plugin/.plugin-version`)
   - Migrations (`plugin/migrations/`)

2. Ensure structure supports:
   - Easy distribution via Claude Code Plugin marketplace
   - Version management and tracking
   - User extensibility where appropriate
   - Clear separation from data layer (`.sow/`)

3. Follow Claude Code plugin conventions:
   - Agent files are Markdown with YAML frontmatter
   - Command files are Markdown (with optional frontmatter)
   - Hooks use JSON configuration format
   - MCP integrations use JSON configuration format

## Context References

- `docs/FILE_STRUCTURE.md` - Target structure specification (recently updated)
- `docs/ARCHITECTURE.md` - Two-layer architecture rationale
- `docs/AGENTS.md` - Agent file format and requirements
- `docs/COMMANDS_AND_SKILLS.md` - Command structure
- `ROADMAP.md` - Milestone 1 deliverables
- Claude Code plugin docs (reviewed earlier in session)

## Acceptance Criteria

- [ ] Complete directory tree defined for `plugin/`
- [ ] All subdirectories clearly documented
- [ ] File naming conventions established
- [ ] Structure aligns with Claude Code plugin requirements
- [ ] Supports version tracking mechanism
- [ ] Clear separation between user workflows and agent skills
- [ ] Documentation explains purpose of each directory/file

## Notes

This is the execution layer SOURCE - the template that becomes `.claude/` when installed.
All files here define behavior and logic, no project-specific state.

The plugin/ directory must be:
- Git-committed (in the sow marketplace repository)
- Distributable via Claude Code Plugin marketplace
- Installable via `/plugin install sow@sow-marketplace`

Focus on establishing a clean, maintainable structure that will scale as sow grows.
