# Task 020: Design file structure for .sow/ directory (data layer)

**Phase**: design
**Assigned Agent**: architect
**Created**: 2025-10-12

## Objective

Design the complete file structure for the `.sow/` directory (data layer).
This directory stores project knowledge, external context (sinks/repos), and
active work state. Created by `/init` command when user bootstraps sow in their repository.

## Requirements

1. Define directory structure for:
   - Version tracking (`.sow/.version`)
   - Repository-specific knowledge (`.sow/knowledge/`)
   - External knowledge sinks (`.sow/sinks/` - git-ignored)
   - Linked repositories (`.sow/repos/` - git-ignored)
   - Active project state (`.sow/project/` - committed to feature branches)

2. Determine git versioning strategy:
   - What gets committed to all branches?
   - What gets git-ignored?
   - What gets committed only to feature branches?

3. Define project/ subdirectory structure:
   - Project state and metadata
   - Phase organization
   - Task structure within phases
   - Logging format
   - Feedback mechanism

4. Establish conventions:
   - File naming (tasks, ADRs, feedback)
   - Directory organization
   - Gap numbering for tasks

## Context References

- `docs/FILE_STRUCTURE.md` - Complete structure specification
- `docs/ARCHITECTURE.md` - Data layer rationale, single project constraint
- `docs/PROJECT_MANAGEMENT.md` - Project lifecycle, phases, tasks
- `ROADMAP.md` - Milestone 1 deliverables

## Acceptance Criteria

- [ ] Complete directory tree defined for `.sow/`
- [ ] Git versioning strategy clearly specified
- [ ] Project structure supports zero-context resumability
- [ ] Task organization supports gap numbering
- [ ] Structure aligns with FILE_STRUCTURE.md specification
- [ ] Supports one-project-per-branch constraint
- [ ] Documentation explains git strategy and rationale

## Notes

This is the data layer - stores knowledge, context, and work state.
Must support the "one project per branch" constraint, where project state
is committed to feature branches but deleted before merge to main.

Mixed git strategy:
- `.sow/knowledge/` - committed (all branches)
- `.sow/sinks/` - git-ignored (per-developer)
- `.sow/repos/` - git-ignored (external repos)
- `.sow/project/` - committed (feature branches only)

Focus on zero-context resumability: all state must be on filesystem.
