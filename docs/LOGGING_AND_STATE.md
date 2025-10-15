# Logging and State Management

**Last Updated**: 2025-10-15
**Purpose**: Logging and state management system

This document describes the structured logging format and state file management that enables zero-context resumability.

---

## Table of Contents

- [Overview](#overview)
- [Logging System](#logging-system)
- [Project State](#project-state)
- [Task State](#task-state)
- [State Maintenance](#state-maintenance)
- [Zero-Context Resumability](#zero-context-resumability)
- [Related Documentation](#related-documentation)

---

## Overview

Structured logging and state files provide audit trail, recovery information, and zero-context resumability. Two log types: project logs (orchestrator actions) and task logs (worker actions). Two state types: project state (phases and tasks) and task state (metadata and context). CLI-driven logging ensures fast consistent format. State files are source of truth for resuming work.

---

## Logging System

### Purpose

Action logs provide audit trail of all actions taken, recovery information for resuming work, debugging context when things go wrong, learning data for understanding agent behavior, chronological history of project evolution.

### Log Types

**Project Log** (`.sow/project/log.md`): Written by orchestrator. High-level coordination events (started project, created phase, spawned worker, phase transitions, requested user approval).

**Task Log** (`phases/<phase>/tasks/<id>/log.md`): Written by workers. Detailed implementation actions (created files, modified code, ran tests, debugging, implementation attempts).

### Structured Format

Each log entry follows strict markdown format. Required fields: timestamp (ISO 8601: YYYY-MM-DD HH:MM:SS), agent ID (auto-constructed from role + iteration), action (from predefined vocabulary), result (success, error, or partial). Optional fields: files (list of affected files), notes (free-form description after required fields).

### Action Vocabulary

Predefined action types ensure consistency: `started_task`, `created_file`, `modified_file`, `deleted_file`, `implementation_attempt`, `test_run`, `refactor`, `debugging`, `research`, `completed_task`, `paused_task`. Extensible (can add new actions as needed).

### CLI-Driven Logging

Problem: direct file editing slow for agents (30+ seconds per edit). Solution: `sow log` CLI command for fast logging. CLI determines context (project vs task log), reads current iteration from task state, auto-constructs agent ID, generates timestamp, formats entry, appends to appropriate log file. Performance: fast single bash command versus slow file edit operation.

**See Also**: [CLI_REFERENCE.md](./CLI_REFERENCE.md#sow-log)

### Benefits

Human-readable markdown format. Git-friendly (clean diffs). Easy to search and grep. Supports rich formatting. LLMs parse easily. Enforces consistency automatically.

---

## Project State

### File Location

`.sow/project/state.yaml` in repository root.

### Purpose

Central planning document containing project metadata (name, description, branch), phase definitions and status, task definitions and status, active phase indicator, complexity assessment.

### Key Responsibilities

Tracks which branch project belongs to (validates current branch matches). Defines phase structure (which phases enabled, current status). Defines task structure (all tasks across phases, assignments, status). Provides resumability (complete picture of project state).

### Branch Tracking

Critical field: `project.branch` stores feature branch name. Orchestrator validates current git branch matches project branch. Warns user if mismatch detected (may have switched branches by accident). Prevents working on wrong branch.

### Committed to Git

Project state committed to feature branch (enables automatic branch switching, team collaboration on same branch, natural backup to remote, CI enforcement of cleanup before merge).

---

## Task State

### File Location

`phases/<phase>/tasks/<id>/state.yaml` within task directory.

### Purpose

Task-specific metadata containing task details (ID, name, phase), status and timestamps, iteration counter, assigned agent, context references (list of relevant files), feedback tracking, modified files list.

### Key Responsibilities

Tracks current iteration for agent ID construction. Lists relevant context files for worker to read. Tracks feedback status (which feedback items pending/addressed). Records files modified by task. Provides complete task context for resumability.

### References Field

Lists all relevant context files for this task. Determined by orchestrator during context compilation. Worker reads all referenced files on startup. Includes refs (knowledge and code), knowledge docs (ADRs, design docs), project context files.

**See Also**: [AGENTS.md](./AGENTS.md#context-compilation)

### Auto-Populated Fields

`files_modified` filled by worker during execution. Worker logs file changes via `sow log`. State updated with complete list of modified files. Provides change tracking for task.

---

## State Maintenance

### Update Responsibilities

**Orchestrator updates**: Project state (phases, phase transitions, task additions, active phase changes). Task state partially (iteration counter before spawning worker).

**Workers update**: Task state (status changes, files modified, feedback status, completion timestamp).

### When Updates Occur

Task status changes (pending → in_progress → completed/abandoned). Phase transitions (one phase completes, next phase activates). Feedback added (new feedback file created, tracking added to state). Iteration incremented (before spawning worker). Files modified (worker logs changes, state updated).

### Consistency

State files are source of truth (not logs). Logs provide audit trail but don't "sync" with state. State can be manually corrected if needed (orchestrator or user can edit state files directly if recovery needed).

### Git Commits

State changes committed regularly during work. Pushed to remote provides backup and enables collaboration. Team can see project progress via git history.

---

## Zero-Context Resumability

### Design Goal

Any agent should be able to resume any project or task from scratch without prior context or conversation history. Benefits: pause/resume across sessions, multiple developers can work on same project, natural handoff between agents, transparent audit trail, no dependency on conversation history.

### Orchestrator Recovery

Process when resuming project: Read project state (what phases exist, what tasks exist, current status, active phase) → Read project log (history of actions, phase transitions, worker spawns) → Verify branch (check current branch matches project.branch) → Determine next action (find next pending task or resume in-progress task) → Compile context (read task description, identify relevant refs and knowledge docs) → Spawn worker (use assigned agent, provide curated context).

### Worker Recovery

Process when resuming task: Read task state (current iteration, assigned agent, context references, feedback status) → Read task description (requirements, acceptance criteria) → Read referenced files (all files listed in references field) → Read task log (what's been tried, what worked/didn't) → Read feedback if any (corrections needed, specific requirements) → Continue work (complete context, knows what's done, knows what's needed).

### Complete Context

Worker has everything needed from disk: task state provides metadata, description provides requirements, references provide domain knowledge, logs provide history, feedback provides corrections. No conversation history needed (pure file-based resumability).

---

## Related Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Zero-context resumability design rationale
- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Task states and iteration counter
- **[FEEDBACK.md](./FEEDBACK.md)** - Feedback files and tracking
- **[AGENTS.md](./AGENTS.md)** - Context compilation and worker coordination
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Log and state file locations
- **[SCHEMAS.md](./SCHEMAS.md)** - State file format specifications
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - sow log command documentation
