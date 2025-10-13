# Task 030: Define YAML schemas for all state files

**Phase**: design
**Assigned Agent**: architect
**Created**: 2025-10-12

## Objective

Define complete YAML schemas for all state files used by sow. These schemas
define the structure of project state, task state, version tracking, and
index files. Must support zero-context resumability and progressive planning.

## Requirements

1. Define schema for `.sow/.version`:
   - Structure version tracking
   - Plugin version tracking
   - Initialization and migration timestamps

2. Define schema for `.sow/project/state.yaml`:
   - Project metadata (name, branch, description)
   - Complexity assessment
   - Phase definitions and status
   - Task definitions within phases
   - Active phase tracking

3. Define schema for task `state.yaml`:
   - Task metadata (id, name, phase)
   - Status tracking (pending, in_progress, completed, abandoned)
   - Iteration counter
   - Agent assignment
   - References (files to read)
   - Feedback tracking
   - Modified files list

4. Define schema for sink and repo index files:
   - `.sow/sinks/index.json` - Installed sink catalog
   - `.sow/repos/index.json` - Linked repository references

5. Ensure all schemas support:
   - Zero-context resumability
   - Progressive planning (phases can be added)
   - Gap numbering (task IDs)
   - Fail-forward approach (abandoned tasks)
   - Iteration tracking for retries

## Context References

- `docs/SCHEMAS.md` - Schema specifications document (target)
- `docs/PROJECT_MANAGEMENT.md` - Project structure requirements
- `docs/ARCHITECTURE.md` - Zero-context resumability needs
- `ROADMAP.md` - Milestone 1 deliverables

## Acceptance Criteria

- [ ] Complete YAML schema for `.sow/.version`
- [ ] Complete YAML schema for project `state.yaml`
- [ ] Complete YAML schema for task `state.yaml`
- [ ] Complete JSON schema for sink `index.json`
- [ ] Complete JSON schema for repo `index.json`
- [ ] All required fields identified
- [ ] Optional fields documented
- [ ] Field types and constraints specified
- [ ] Examples provided for each schema
- [ ] Schemas support all documented workflows

## Notes

These schemas are the contract between agents and the filesystem.
They must be complete enough to enable zero-context resumability:
any agent should be able to resume any project/task by reading these files.

YAML chosen for state files (human-readable, comments supported).
JSON chosen for index files (LLM-maintained, machine-optimized).

Gap numbering: task IDs use strings like "010", "020", "030" to allow
insertions ("011", "012") without renumbering.

Update SCHEMAS.md with complete, validated schemas.
