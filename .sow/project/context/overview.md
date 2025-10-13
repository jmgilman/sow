# Project Overview: Milestone 1 Foundation

## Goal

Establish the basic file structure, schemas, and version management system for the sow framework.

## Scope

This project implements the foundational two-layer architecture:
- **Execution Layer** (`plugin/`) - Agent behaviors, commands, hooks
- **Data Layer** (`.sow/`) - Project knowledge, state, external context

## Key Decisions

1. **Complexity: Simple (1)**
   - Implementing already-documented designs
   - FILE_STRUCTURE.md, SCHEMAS.md, ARCHITECTURE.md provide complete specifications
   - No new design work required

2. **Schema templates already exist**
   - `schemas/templates/` contains 5 YAML/JSON templates
   - These are reference templates, not the actual plugin files
   - Plugin will contain actual agent definitions and metadata

3. **Plugin structure follows FILE_STRUCTURE.md**
   - `plugin/` is the source (what we develop)
   - Users install it, becomes `.claude/` in their repos
   - Development: edit in `plugin/`
   - Runtime: users see `.claude/`

## Success Criteria

- [ ] `plugin/` directory structure matches FILE_STRUCTURE.md specification
- [ ] All agent definitions created (orchestrator, architect, implementer, integration-tester, reviewer, documenter)
- [ ] Plugin metadata files created (.claude-plugin/plugin.json, .plugin-version)
- [ ] `.sow/` data layer structure initialized
- [ ] Version tracking file (.sow/.version) created
- [ ] All files follow documented schemas

## References

- [FILE_STRUCTURE.md](../../../docs/FILE_STRUCTURE.md) - Complete directory layout
- [SCHEMAS.md](../../../docs/SCHEMAS.md) - File format specifications
- [ARCHITECTURE.md](../../../docs/ARCHITECTURE.md) - Two-layer architecture
- [AGENTS.md](../../../docs/AGENTS.md) - Agent specifications
- [ROADMAP.md](../../../ROADMAP.md) - Milestone 1 details
