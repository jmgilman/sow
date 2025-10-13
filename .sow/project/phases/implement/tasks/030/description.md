# Task 030: Implement version tracking files

**Phase**: implement
**Assigned Agent**: implementer
**Created**: 2025-10-12

## Objective

Implement the version tracking system for sow. This includes both the plugin
version file and the repository structure version file. These enable version
mismatch detection and guide the migration process.

## Requirements

1. Update `plugin/.plugin-version`:
   - Simple text file with version number
   - Currently shows "0.1.0"
   - Verify format is correct
   - Read by orchestrator at runtime

2. Create `.sow/.version` template:
   - YAML format with schema from task 030
   - Tracks structure version
   - Tracks plugin version
   - Records initialization timestamp
   - Records last migration timestamp
   - Used by `/init` to create in user repos

3. Ensure version tracking supports:
   - Runtime version reading by agents
   - Version mismatch detection (SessionStart hook)
   - Migration path identification
   - Semantic versioning (MAJOR.MINOR.PATCH)

4. Document version management:
   - How versions are updated
   - When to increment (MAJOR vs MINOR vs PATCH)
   - How migrations relate to versions

## Context References

- `docs/DISTRIBUTION.md` - Version tracking and migration system
- `docs/SCHEMAS.md` - Schema for `.sow/.version`
- `plugin/.plugin-version` - Existing file to verify
- Design task 030 output - YAML schemas

## Acceptance Criteria

- [ ] `plugin/.plugin-version` verified or updated (currently 0.1.0)
- [ ] `.sow/.version` template created with correct schema
- [ ] Template includes all required fields
- [ ] Version format follows semantic versioning
- [ ] Files support version mismatch detection
- [ ] Documentation explains version management

## Notes

Two separate version files serve different purposes:
- `plugin/.plugin-version` - What plugin version is installed (lives in .claude/ after install)
- `.sow/.version` - What structure version repository uses (lives in user repo)

Version mismatch: When plugin version != structure version, prompt for `/migrate`.

Initial version: 0.1.0 for both files (Milestone 1 delivery).

The existing `plugin/.plugin-version` already exists and contains "0.1.0" - verify it's correct.
