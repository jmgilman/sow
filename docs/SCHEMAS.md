# File Format Schemas

**Last Updated**: 2025-10-15
**Purpose**: Complete schema reference for all file formats

This document provides specifications for all file formats used in `sow`. CUE schemas embedded in the CLI are the authoritative source. This document provides human-readable reference.

---

## Table of Contents

- [Overview](#overview)
- [Project State Schema](#project-state-schema)
- [Task State Schema](#task-state-schema)
- [Refs Index Schemas](#refs-index-schemas)
- [Accessing and Validating Schemas](#accessing-and-validating-schemas)
- [Related Documentation](#related-documentation)

---

## Overview

### CUE Schema System

All schemas defined using CUE (Configure, Unify, Execute) and embedded in CLI binary. CUE provides type safety, validation, and documentation generation. Single source of truth (schema definitions in CLI eliminate drift between code and docs). Strong validation (type checking and constraints enforce correctness). Self-documentation (schemas include inline docs). Tooling integration (validation, formatting, conversion).

### Automatic Validation

CLI validates files against embedded CUE schemas during: project initialization, task creation, state updates, file operations. Invalid files rejected with detailed error messages.

### File Format Summary

| File | Format | Purpose |
|------|--------|---------|
| `state.yaml` (project) | YAML | Project state with 5-phase structure |
| `state.yaml` (task) | YAML | Individual task metadata |
| `description.md` (task) | Markdown | Task requirements |
| `log.md` | Markdown | Action logs (structured) |
| `index.json` (refs committed) | JSON | Remote refs catalog (committed) |
| `index.json` (refs cache) | JSON | Cache metadata (local only) |
| `index.local.json` | JSON | Local-only refs (gitignored) |

---

## Project State Schema

**File**: `.sow/project/state.yaml`

**Purpose**: Central state file for 5-phase project model. Phases are fixed: discovery, design, implementation, review, finalize.

### Key Characteristics

**Fixed 5 Phases**: All projects have same five phases (discovery/design optional, implementation/review/finalize required). Phase enablement flags control which phases execute. Simplified structure (no dynamic phase addition).

**Artifact Tracking**: Discovery and design phases produce artifacts requiring human approval. Approval tracked via boolean flag. Phase cannot complete until all artifacts approved.

**Review Iterations**: Review phase supports multiple iterations (loop-back to implementation). Reports numbered sequentially (001, 002, 003). Iteration counter tracks current cycle.

### Complete Schema

```yaml
project:
  name: string                    # Kebab-case identifier
  branch: string                  # Git branch name
  description: string             # Human-readable description
  created_at: timestamp           # ISO 8601
  updated_at: timestamp           # ISO 8601

phases:
  discovery:
    enabled: bool                 # Whether phase enabled
    status: string                # skipped | pending | in_progress | completed
    created_at: timestamp
    started_at: timestamp | null
    completed_at: timestamp | null
    discovery_type: string | null # bug | feature | docs | refactor | general
    artifacts:
      - path: string              # Relative from .sow/project/
        approved: bool            # Human approval required
        created_at: timestamp

  design:
    enabled: bool
    status: string                # skipped | pending | in_progress | completed
    created_at: timestamp
    started_at: timestamp | null
    completed_at: timestamp | null
    architect_used: bool | null   # Whether architect agent used
    artifacts:
      - path: string
        approved: bool
        created_at: timestamp

  implementation:
    enabled: true                 # Always true (required phase)
    status: string                # pending | in_progress | completed
    created_at: timestamp
    started_at: timestamp | null
    completed_at: timestamp | null
    planner_used: bool | null     # Whether planner agent used
    tasks:
      - id: string                # Gap-numbered (010, 020, 030)
        name: string
        status: string            # pending | in_progress | completed | abandoned
        parallel: bool
        dependencies: [string] | null
    pending_task_additions: [...] | null  # Tasks awaiting approval

  review:
    enabled: true                 # Always true (required phase)
    status: string                # pending | in_progress | completed
    created_at: timestamp
    started_at: timestamp | null
    completed_at: timestamp | null
    iteration: integer            # Current review iteration (1-indexed)
    reports:
      - id: string                # Report number (001, 002, 003)
        path: string              # Relative from .sow/project/phases/review/
        created_at: timestamp
        assessment: string        # pass | fail

  finalize:
    enabled: true                 # Always true (required phase)
    status: string                # pending | in_progress | completed
    created_at: timestamp
    started_at: timestamp | null
    completed_at: timestamp | null
    documentation_updates: [string] | null
    artifacts_moved: [...] | null
    project_deleted: bool         # Critical gate (must be true to complete)
    pr_url: string | null
```

### Field Descriptions

**project.branch**: Git branch project belongs to. Orchestrator validates current branch matches this field.

**phases.discovery.enabled / phases.design.enabled**: Whether phase executes. If false, status must be "skipped". If true, status cannot be "skipped".

**phases.*.artifacts**: Outputs requiring human approval. Discovery artifacts (research reports, notes). Design artifacts (ADRs, design docs). Cannot complete phase until all artifacts have `approved: true`.

**phases.implementation.tasks**: Task breakdown with gap numbering. Presence in array indicates human approval to execute.

**phases.implementation.pending_task_additions**: Tasks orchestrator wants to add mid-implementation. Human approval moves them to tasks array (fail-forward mechanism).

**phases.review.iteration**: Tracks review cycle number. Increments each time review loops back to implementation.

**phases.review.reports[].assessment**: "pass" (ready to finalize) or "fail" (loop back to implementation).

**phases.finalize.project_deleted**: Must be true before phase can complete. Enforces cleanup before PR merge.

### Complete Example

```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  description: Add JWT-based authentication with user login
  created_at: "2025-10-14T10:00:00Z"
  updated_at: "2025-10-14T17:00:00Z"

phases:
  discovery:
    enabled: true
    status: completed
    created_at: "2025-10-14T10:05:00Z"
    started_at: "2025-10-14T10:05:00Z"
    completed_at: "2025-10-14T11:20:00Z"
    discovery_type: feature
    artifacts:
      - path: phases/discovery/notes.md
        approved: true
        created_at: "2025-10-14T11:15:00Z"

  design:
    enabled: true
    status: completed
    created_at: "2025-10-14T11:25:00Z"
    started_at: "2025-10-14T11:25:00Z"
    completed_at: "2025-10-14T13:00:00Z"
    architect_used: true
    artifacts:
      - path: phases/design/adrs/001-use-jwt-rs256.md
        approved: true
        created_at: "2025-10-14T12:30:00Z"

  implementation:
    enabled: true
    status: completed
    created_at: "2025-10-14T13:05:00Z"
    started_at: "2025-10-14T13:10:00Z"
    completed_at: "2025-10-14T16:20:00Z"
    planner_used: false
    tasks:
      - id: "010"
        name: Create User model
        status: completed
        parallel: false
      - id: "020"
        name: Implement JWT service
        status: completed
        parallel: false
        dependencies: ["010"]
      - id: "030"
        name: Add login endpoint
        status: completed
        parallel: false
        dependencies: ["020"]

  review:
    enabled: true
    status: completed
    created_at: "2025-10-14T15:50:00Z"
    started_at: "2025-10-14T15:50:00Z"
    completed_at: "2025-10-14T16:30:00Z"
    iteration: 2
    reports:
      - id: "001"
        path: phases/review/reports/001-review.md
        created_at: "2025-10-14T15:55:00Z"
        assessment: fail
      - id: "002"
        path: phases/review/reports/002-review.md
        created_at: "2025-10-14T16:25:00Z"
        assessment: pass

  finalize:
    enabled: true
    status: completed
    created_at: "2025-10-14T16:35:00Z"
    started_at: "2025-10-14T16:35:00Z"
    completed_at: "2025-10-14T17:00:00Z"
    documentation_updates:
      - README.md
      - docs/api-reference.md
    artifacts_moved:
      - from: phases/design/adrs/001-use-jwt-rs256.md
        to: docs/adrs/003-use-jwt-rs256.md
    project_deleted: true
    pr_url: https://github.com/org/repo/pull/42
```

---

## Task State Schema

**File**: `.sow/project/phases/implementation/tasks/<id>/state.yaml`

**Purpose**: Task-specific metadata for implementation phase tasks only. Schema unchanged from previous versions (works for new 5-phase model).

### Complete Schema

```yaml
task:
  id: string                      # Task ID (matches directory name, gap-numbered)
  name: string                    # Task name
  phase: "implementation"         # Always implementation in new model
  status: string                  # pending | in_progress | completed | abandoned
  created_at: timestamp           # When task created
  started_at: timestamp | null    # When work first began
  updated_at: timestamp           # Last modification
  completed_at: timestamp | null  # When completed
  iteration: integer              # Attempt counter (managed by orchestrator)
  references: [string]            # Context file paths (relative to .sow/)
  feedback:
    - id: string                  # Feedback number (001, 002, 003)
      created_at: timestamp
      status: string              # pending | addressed | superseded
  files_modified: [string]        # Files changed during task (auto-populated)
```

### Field Descriptions

**iteration**: Tracks attempt number. Incremented by orchestrator before spawning worker. Constructs agent ID: `{assigned_agent}-{iteration}`.

**references**: Paths relative to `.sow/` root. Orchestrator compiles this list during context compilation. Worker reads all referenced files at startup.

**feedback**: Human corrections. Worker addresses all pending feedback on next iteration. Status changes from pending to addressed when incorporated.

**files_modified**: Auto-populated by worker using `sow log` command. Tracks complete list of files changed during task execution.

### Complete Example

```yaml
task:
  id: "020"
  name: Implement JWT service
  phase: implementation
  status: in_progress
  created_at: "2025-10-14T13:15:00Z"
  started_at: "2025-10-14T13:20:00Z"
  updated_at: "2025-10-14T14:45:00Z"
  completed_at: null
  iteration: 3
  references:
    - refs/python-style/conventions.md
    - refs/api-standards/rest-standards.md
    - knowledge/adrs/001-use-jwt-rs256.md
  feedback:
    - id: "001"
      created_at: "2025-10-14T14:30:00Z"
      status: addressed
    - id: "002"
      created_at: "2025-10-14T15:15:00Z"
      status: pending
  files_modified:
    - src/auth/jwt.py
    - tests/test_jwt.py
    - requirements.txt
```

---

## Refs Index Schemas

### Committed Index

**File**: `.sow/refs/index.json`

**Purpose**: Categorical metadata about remote refs shared across team via git.

```json
{
  "version": "1.0.0",
  "refs": [
    {
      "id": "string",               // Unique identifier
      "type": "knowledge|code",     // Reference type
      "source": "string",           // Git URL
      "branch": "string",           // Branch name
      "paths": [
        {
          "path": "string",         // Subpath within repo
          "link": "string",         // Symlink name in .sow/refs/
          "tags": ["string"],       // Topic keywords
          "description": "string",  // One-sentence summary
          "summary": "string"       // 2-3 sentence description
        }
      ]
    }
  ]
}
```

### Cache Index

**File**: `~/.cache/sow/index.json`

**Purpose**: Transient metadata about cached repos (per-machine, not committed).

```json
{
  "version": "1.0.0",
  "repos": [
    {
      "source": "string",           // Git URL
      "branch": "string",           // Branch name
      "cached_path": "string",      // Relative path in cache
      "commit_sha": "string",       // Current local SHA
      "cached_at": "timestamp",     // Initial cache timestamp
      "last_checked": "timestamp",  // Last staleness check
      "last_updated": "timestamp",  // Last fetch/pull
      "status": "string",           // current | behind | ahead | diverged | error
      "commits_behind": "int|null", // Number behind remote
      "remote_sha": "string",       // Latest remote SHA
      "used_by": [
        {
          "repo_path": "string",    // Absolute path to consuming repo
          "link_type": "string",    // symlink | copy
          "paths": ["string"]       // Subpaths used
        }
      ]
    }
  ]
}
```

### Local Index

**File**: `.sow/refs/index.local.json`

**Purpose**: Local-only references not shared with team (gitignored).

```json
{
  "version": "1.0.0",
  "refs": [
    {
      "id": "string",
      "type": "knowledge|code",
      "source": "file:///path",     // file:// protocol for local paths
      "link": "string",
      "tags": ["string"],
      "description": "string",
      "summary": "string"
    }
  ]
}
```

**See Also**: [REFS.md](./REFS.md) for complete refs system documentation.

---

## Accessing and Validating Schemas

### View Schemas

```bash
# List all available schemas
sow schema list

# View project state schema
sow schema show project

# View task state schema
sow schema show task

# Export schema to file
sow schema export project > project-schema.cue
```

### Validate Files

```bash
# Validate project state
sow schema validate project .sow/project/state.yaml

# Validate task state
sow schema validate task .sow/project/phases/implementation/tasks/010/state.yaml

# Validate refs index
sow schema validate refs-index .sow/refs/index.json
```

### Error Messages

Invalid files rejected with detailed messages indicating: which field violated constraint, expected value or type, actual value found, line number in source file (when possible).

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - 5-phase model and lifecycle
- **[PHASES/](./PHASES/)** - Individual phase specifications
- **[TASK_MANAGEMENT.md](./TASK_MANAGEMENT.md)** - Task structure
- **[REFS.md](./REFS.md)** - External references system
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - State file management
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Directory organization
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - sow schema commands
