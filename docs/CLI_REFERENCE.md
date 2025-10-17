# CLI Reference

**Last Updated**: 2025-10-15
**Purpose**: Complete CLI command reference

This document provides comprehensive reference for all `sow` CLI commands and orchestrator slash commands.

---

## Table of Contents

- [Overview](#overview)
- [Core Commands](#core-commands)
- [Logging Commands](#logging-commands)
- [Refs Commands](#refs-commands)
- [Slash Commands](#slash-commands)
- [Exit Codes](#exit-codes)
- [Environment Variables](#environment-variables)
- [Related Documentation](#related-documentation)

---

## Overview

The `sow` CLI provides: structure initialization (materializes `.sow/` directory structure), validation (validates files against embedded schemas), fast logging (instant log entries), refs management (external references system), session information (current project status).

**Installation**: Download appropriate binary for platform from GitHub releases. Place in PATH. Verify with `sow version`.

---

## Core Commands

### sow version

Display CLI version.

**Usage**: `sow version`

**Output**: `sow 0.2.0`

**Purpose**: Verify CLI installation and check version alignment with plugin.

---

### sow init

Initialize sow structure in repository.

**Usage**: `sow init`

**Actions**: Creates `.sow/` directory structure, creates knowledge directory with templates, creates refs directory with gitignore, creates version file, commits structure to git.

**Prerequisites**: Must be in git repository, not already initialized.

**Output**: Confirmation of created structure and git commit.

**Error Cases**: Already initialized, not in git repository.

---

### sow validate

Validate sow structure integrity.

**Usage**: `sow validate`

**Checks**: Directory structure exists, version file valid, refs indexes valid JSON, project state valid YAML (if exists), task states valid (if exist).

**Output**: Success with checklist, or errors with detailed messages.

**Exit Codes**: 0 (all passed), 1 (errors found).

---

## Logging Commands

### sow log

Create structured log entry.

**Usage**: `sow log [--file FILE]... --action ACTION --result RESULT "Description"`

**Options**:
- `--file FILE` - File affected (multiple allowed)
- `--action ACTION` - Action type from vocabulary
- `--result RESULT` - success | error | partial
- `--project` - Force project log (default auto-detects)
- Description - Free-form notes (required, quoted)

**Action Vocabulary**: `started_task`, `created_file`, `modified_file`, `deleted_file`, `implementation_attempt`, `test_run`, `refactor`, `debugging`, `research`, `completed_task`, `paused_task`.

**Auto-Detection**: From task directory writes to task log. From project root writes to project log. Use `--project` to force project level.

**CLI Responsibilities**: Determines context, reads iteration from state, constructs agent ID, generates timestamp, formats entry, appends to log.

**Examples**:
```bash
# Task log
sow log --file src/auth/jwt.py --action created_file --result success \
  "Created JWT service with RS256"

# Multiple files
sow log --file src/auth/jwt.py --file tests/test_jwt.py \
  --action modified_file --result success \
  "Implemented validation and tests"

# Project log
sow log --project --action started_phase --result success \
  "Started implementation phase"
```

**See Also**: [LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md#cli-driven-logging)

---

### sow session-info

Display current session information for SessionStart hook.

**Usage**: `sow session-info`

**Output Variations**:

**No project**: Shows repository status, no active project message, versions, available commands.

**With project**: Shows repository status, active project name and branch, resume message, versions, available commands.

**Version mismatch**: Shows warning with structure version and plugin version, migration instructions.

**Purpose**: Provide context on session start via SessionStart hook.

---

## Refs Commands

### sow refs add

Add remote or local reference.

**Usage Remote**: `sow refs add <url> --path <path> --link <name> --type <type> --tag <tag>... --desc "<description>" --summary "<summary>" [--branch <branch>]`

**Usage Local**: `sow refs add file:///<path> --link <name> --type <type> --tag <tag>... --desc "<description>" --summary "<summary>"`

**Arguments**:
- `url` - Git repository URL (https or ssh)
- `--branch` - Branch name (optional, defaults to repo default)
- `--path` - Subdirectory within repo (use "" or omit for root)
- `--link` - Symlink name in `.sow/refs/`
- `--type` - `knowledge` or `code`
- `--tag` - Topic tag (repeat for multiple)
- `--desc` - One-sentence description (quoted)
- `--summary` - 2-3 sentence summary (quoted)

**Behavior**: Clones to cache if needed (or uses existing cache), adds entry to index, creates symlink or copy, updates cache index.

**Examples**:
```bash
# Remote knowledge ref
sow refs add https://github.com/acme/style-guides \
  --path python/ --link python-style --type knowledge \
  --tag formatting --tag naming --tag testing \
  --desc "Python coding standards" \
  --summary "Covers PEP 8 formatting and testing patterns."

# Local ref
sow refs add file:///Users/josh/docs \
  --link local-docs --type knowledge \
  --tag wip --desc "Work-in-progress documentation" \
  --summary "Draft docs before publishing."
```

**See Also**: [REFS.md](./REFS.md#cli-commands)

---

### sow refs init

Initialize references after cloning repository.

**Usage**: `sow refs init`

**Behavior**: Reads committed and local indexes, clones repos to cache if needed, creates symlinks or copies, updates cache index.

**Output**: Lists repositories cached and links created, summary of initialized refs.

**Usage Context**: Run after `git clone` when `.sow/refs/` directory empty but indexes exist.

---

### sow refs status

Check if references up to date with remote.

**Usage**: `sow refs status [id]`

**Arguments**: `id` - Optional specific ref to check (checks all if omitted).

**Behavior**: Fetches from remote, compares local SHA to remote SHA, updates cache index with status, shows staleness information.

**Output**: Status for each ref (current, behind by N commits with recent commit list, error with details).

---

### sow refs update

Pull latest changes from remote.

**Usage**: `sow refs update [id]`

**Arguments**: `id` - Optional specific ref to update (updates all behind refs if omitted).

**Behavior**: Git pull in cache, update cache index with new SHA, rsync to consuming repos on Windows, symlinks automatically reflect on Unix.

**Output**: Update summary (commits pulled, files synced on Windows).

---

### sow refs list

Display all configured references.

**Usage**: `sow refs list [--remote] [--local] [--all]`

**Flags**:
- `--remote` - Show only remote refs (from index.json)
- `--local` - Show only local refs (from index.local.json)
- `--all` - Show both (default)

**Output**: Tree-formatted list showing source, branch, tags, status, description for each ref. Separate sections for remote and local refs.

---

### sow refs remove

Remove reference from repository.

**Usage**: `sow refs remove <id>`

**Arguments**: `id` - Reference identifier to remove.

**Behavior**: Confirms with user (unless `--force`), removes index entry, removes symlink or copy, updates cache index, optionally prunes cache.

**Output**: Confirmation prompt, removal confirmation, note about cache pruning.

---

## Slash Commands

Slash commands are orchestrator-facing commands expanded to full prompts. Invoked by orchestrator agent during project coordination.

### /project:new

Create new project (replaces `/start-project`).

**When**: User wants to start new project work.

**Behavior**: Invokes truth table decision flow, asks questions to determine phase enablement, presents phase plan for approval, creates project structure on approval, transitions to first enabled phase.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)

---

### /project:continue

Continue existing project (replaces `/continue`).

**When**: User wants to resume work on existing project.

**Behavior**: Reads project state, verifies branch matches, identifies next action (pending task or phase transition), compiles context, spawns worker or transitions phase.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)

---

### /phase:discovery

Discovery phase workflow.

**When**: Invoked by `/project:new` or `/project:continue` when discovery phase active.

**Behavior**: Categorizes discovery work (bug, feature, docs, refactor, general), delegates to type-specific command, facilitates research and investigation, creates discovery artifacts, requests human approval.

**See Also**: [PHASES/DISCOVERY.md](./PHASES/DISCOVERY.md)

---

### /phase:design

Design phase workflow.

**When**: Invoked when design phase active.

**Behavior**: Facilitates design alignment (conversational refinement), spawns architect agent if needed, creates design artifacts (ADRs, design docs), requests human approval for artifacts, transitions to implementation when approved.

**See Also**: [PHASES/DESIGN.md](./PHASES/DESIGN.md)

---

### /phase:implementation

Implementation phase workflow.

**When**: Invoked when implementation phase active. Always happens (required phase).

**Behavior**: Creates initial task breakdown (with human approval), spawns implementer agents for tasks, manages fail-forward task additions, handles parallel execution, transitions to review when complete.

**See Also**: [PHASES/IMPLEMENTATION.md](./PHASES/IMPLEMENTATION.md)

---

### /phase:review

Review phase workflow.

**When**: Invoked when review phase active. Always happens (required phase).

**Behavior**: Performs mandatory orchestrator review, spawns reviewer agent if needed, creates review report, loops back to implementation if issues found, transitions to finalize when approved.

**See Also**: [PHASES/REVIEW.md](./PHASES/REVIEW.md)

---

### /phase:finalize

Finalize phase workflow.

**When**: Invoked when finalize phase active. Always happens (required phase).

**Behavior**: Documentation subphase (updates needed), runs final checks (tests, linters), deletes project folder (mandatory), creates pull request, hands off to human for merge.

**See Also**: [PHASES/FINALIZE.md](./PHASES/FINALIZE.md)

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | Not in sow repository |
| `4` | Validation failed |
| `5` | Network error |
| `6` | File system error |

**Usage in Scripts**:
```bash
if sow validate; then
  echo "Validation passed"
else
  echo "Validation failed"
  exit 1
fi
```

---

## Environment Variables

### Configurable Behavior

| Variable | Purpose | Default |
|----------|---------|---------|
| `SOW_NO_COLOR` | Disable colored output | `false` |
| `SOW_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `SOW_TIMEOUT` | Network timeout (seconds) | `30` |

**Examples**:
```bash
# Disable color
export SOW_NO_COLOR=1
sow validate

# Debug logging
export SOW_LOG_LEVEL=debug
sow refs add https://example.com/repo --path docs --link docs --type knowledge

# Short timeout
export SOW_TIMEOUT=10
sow refs update
```

---

## Related Documentation

- **[REFS.md](./REFS.md)** - Complete refs system documentation
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - Logging system details
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - Project lifecycle and slash commands
- **[PHASES/](./PHASES/)** - Individual phase command workflows
- **[SCHEMAS.md](./SCHEMAS.md)** - Schema validation and formats
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day usage workflows
