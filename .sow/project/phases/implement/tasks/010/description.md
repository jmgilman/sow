# Task 010: Create plugin/ template files and directories

**Phase**: implement
**Assigned Agent**: implementer
**Created**: 2025-10-12

## Objective

Create the complete `plugin/` directory structure with all template files.
This implements the design from design/010 task. The plugin/ directory is
the source template that becomes `.claude/` when users install the plugin.

## Requirements

1. Create directory structure:
   - `plugin/.claude-plugin/`
   - `plugin/agents/`
   - `plugin/commands/workflows/`
   - `plugin/commands/skills/{architect,implementer,integration-tester,reviewer,documenter}/`
   - `plugin/migrations/`

2. Create essential files:
   - `plugin/.claude-plugin/plugin.json` - Plugin metadata
   - `plugin/.plugin-version` - Version file (0.1.0)
   - `plugin/hooks.json` - Hooks configuration (SessionStart for version checking)
   - `plugin/mcp.json` - MCP integrations config (empty initially)
   - `plugin/README.md` - Installation instructions

3. Create placeholder agent files:
   - Basic structure for each agent (orchestrator, architect, implementer, etc.)
   - YAML frontmatter with name, description, tools
   - Minimal system prompt (will be enhanced in future milestones)

4. Create placeholder command files:
   - Workflow commands: init, start-project, continue, cleanup, migrate, sync
   - Skill commands: One per agent (create-adr, implement-feature, etc.)
   - Bootstrap command (already exists, may need to move)

5. Ensure all files follow conventions:
   - Agent files: Markdown with YAML frontmatter
   - Command files: Markdown
   - Config files: JSON (hooks.json, mcp.json, plugin.json)

## Context References

- `docs/FILE_STRUCTURE.md` - Complete structure
- `docs/AGENTS.md` - Agent file format
- `docs/COMMANDS_AND_SKILLS.md` - Command structure
- Design task 010 output - File structure design

## Acceptance Criteria

- [ ] All directories created
- [ ] plugin.json with correct metadata
- [ ] .plugin-version file with 0.1.0
- [ ] hooks.json configured
- [ ] All 6 agent files created (orchestrator, architect, implementer, integration-tester, reviewer, documenter)
- [ ] All workflow command files created
- [ ] Skill command files created (at least one per agent)
- [ ] Files follow naming conventions
- [ ] Structure matches design from task 010

## Notes

This is creating the execution layer source. Focus on structure and placeholders.
Detailed agent prompts and command implementations will come in future milestones.

The `.claude/commands/bootstrap.md` already exists - move it to `plugin/commands/bootstrap.md`.

Version 0.1.0 for initial implementation.
