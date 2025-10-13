## Task: Create plugin directory structure

Create the complete `plugin/` directory structure that will become `.claude/` when users install the sow plugin.

### Requirements

- Create `plugin/` root directory with all subdirectories
- Follow the structure defined in FILE_STRUCTURE.md:
  - `plugin/.claude-plugin/` - Plugin metadata
  - `plugin/agents/` - Agent definitions
  - `plugin/commands/workflows/` - User-facing slash commands
  - `plugin/commands/skills/` - Agent-invoked skills (organized by agent)
  - `plugin/migrations/` - Version migration instructions
- Create skill subdirectories:
  - `plugin/commands/skills/architect/`
  - `plugin/commands/skills/implementer/`
  - `plugin/commands/skills/integration-tester/`
  - `plugin/commands/skills/reviewer/`
  - `plugin/commands/skills/documenter/`
- Do NOT create `hooks.json` or `mcp.json` yet (later milestones)
- Create empty `.gitkeep` files in empty directories as needed

### Acceptance Criteria

- [ ] `plugin/` directory exists with correct structure
- [ ] All required subdirectories created
- [ ] Structure matches FILE_STRUCTURE.md specification exactly
- [ ] Skill directories organized by agent type
- [ ] No files created yet, just directory structure

### Context Notes

- The `plugin/` directory is the SOURCE for the distributable plugin
- When users run `/plugin install sow@sow-marketplace`, contents of `plugin/` are copied to their `.claude/` directory
- This is a pure directory creation task - files will be created in subsequent tasks
- Reference FILE_STRUCTURE.md lines 176-200 for complete plugin structure
