# CLI Reference

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Download and Install](#download-and-install)
  - [Verify Installation](#verify-installation)
  - [Updating the CLI](#updating-the-cli)
- [Global Commands](#global-commands)
  - [sow --version](#sow---version)
  - [sow --help](#sow---help)
  - [sow init](#sow-init)
  - [sow validate](#sow-validate)
  - [sow schema](#sow-schema)
- [Logging Commands](#logging-commands)
  - [sow log](#sow-log)
  - [sow session-info](#sow-session-info)
- [Sink Management](#sink-management)
  - [sow sinks install](#sow-sinks-install)
  - [sow sinks update](#sow-sinks-update)
  - [sow sinks list](#sow-sinks-list)
  - [sow sinks remove](#sow-sinks-remove)
  - [sow sinks reindex](#sow-sinks-reindex)
- [Repository Management](#repository-management)
  - [sow repos add](#sow-repos-add)
  - [sow repos sync](#sow-repos-sync)
  - [sow repos list](#sow-repos-list)
  - [sow repos remove](#sow-repos-remove)
  - [sow repos reindex](#sow-repos-reindex)
- [Utility Commands](#utility-commands)
  - [sow save-context](#sow-save-context)
  - [sow sync](#sow-sync)
- [Command-Line Flags](#command-line-flags)
- [Exit Codes](#exit-codes)
- [Environment Variables](#environment-variables)
- [Related Documentation](#related-documentation)

---

## Overview

The `sow` CLI is **required** for using sow. It provides essential schema management, initialization, validation, and fast operations.

**Key Benefits**:
- **Schema management** - Embeds CUE schemas as source of truth for all `.sow/` file formats
- **Initialization** - Materializes `.sow/` structure with correct defaults from schemas
- **Validation** - Validates files against embedded CUE schemas
- **Fast logging** - Instant log entry creation (vs 30s file edit)
- **Sink management** - Install and update knowledge sinks
- **Repository management** - Link external repositories
- **Session info** - Display current project status

**Installation**: See [DISTRIBUTION.md](./DISTRIBUTION.md) for installation instructions.

---

## Installation

The sow CLI must be installed before you can use sow. The `/init` slash command will guide you through installation if the CLI is not detected.

### Download and Install

**macOS**:
```bash
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow
```

**Linux**:
```bash
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-linux -o sow
chmod +x sow
mv sow ~/.local/bin/sow
```

**Windows**:
```powershell
# Download sow-windows.exe from releases
# Add to PATH
```

### Verify Installation

```bash
sow --version
# sow 0.2.0

sow --help
# Displays help information
```

### Updating the CLI

Download latest version matching your plugin version:

```bash
# Check plugin version
cat .claude/.plugin-version
# 0.3.0

# Download matching CLI
curl -L https://github.com/your-org/sow/releases/download/v0.3.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Verify
sow --version
# sow 0.3.0
```

---

## Global Commands

### sow --version

Display CLI version.

**Usage**:
```bash
sow --version
```

**Output**:
```
sow 0.2.0
```

**Purpose**: Verify CLI installation and check version alignment with plugin.

---

### sow --help

Display help information.

**Usage**:
```bash
sow --help
```

**Output**:
```
sow - AI-powered system of work for software engineering

Usage:
  sow [command] [options]

Commands:
  log              Create log entry
  session-info     Display session information
  sinks            Manage knowledge sinks
  repos            Manage linked repositories
  validate         Validate sow structure
  sync             Sync sinks and repos
  --version        Show version
  --help           Show this help

For command-specific help:
  sow [command] --help

Documentation: https://github.com/your-org/sow
```

---

### sow init

Initialize sow structure in a repository by materializing defaults from embedded CUE schemas.

**Usage**:
```bash
sow init
```

**Actions**:
1. Checks prerequisites (git repository, not already initialized)
2. Materializes `.sow/` structure from CUE schemas with default values:
   - `.sow/knowledge/` (with `overview.md` template)
   - `.sow/sinks/` (with empty `index.json`)
   - `.sow/repos/` (with empty `index.json`)
   - `.sow/.version` (tracks current version)
3. Creates `.gitignore` entries for `.sow/sinks/` and `.sow/repos/`
4. Commits structure to git

**Output**:
```
✓ Checking prerequisites...
✓ Creating .sow/ structure from CUE schemas...
  - .sow/knowledge/ (with overview.md template)
  - .sow/sinks/ (with index.json)
  - .sow/repos/ (with index.json)
  - .sow/.version (tracking version 0.2.0)

✓ Creating .gitignore entries...
  - .sow/sinks/
  - .sow/repos/

✓ Committing structure to git...
  [main abc1234] Initialize sow (v0.2.0)

sow initialized successfully!
```

**Error Cases**:
- Already initialized: "sow already initialized in this repository"
- Not in git repository: "Must be in a git repository to use sow"

---

### sow validate

Validate sow structure integrity against embedded CUE schemas.

**Purpose**: Validates all `.sow/` files against the CUE schemas embedded in the CLI binary. This ensures file formats are correct, required fields are present, and constraints are satisfied.

**Usage**:
```bash
sow validate
```

**Output (success)**:
```
✓ Validating sow structure...

✓ .sow/ directory exists
✓ .sow/.version exists and is valid
✓ .sow/knowledge/ directory exists
✓ .sow/sinks/ directory exists
✓ .sow/sinks/index.json is valid JSON
✓ .sow/repos/ directory exists
✓ .sow/repos/index.json is valid JSON

Project validation:
✓ .sow/project/ exists
✓ .sow/project/state.yaml is valid YAML
✓ .sow/project/log.md exists
✓ All task state files are valid

All checks passed!
```

**Output (errors)**:
```
✓ Validating sow structure...

✗ .sow/.version is missing
✗ .sow/project/state.yaml is invalid YAML (line 42: unexpected token)
✓ .sow/sinks/index.json is valid JSON

2 errors found. See above for details.
```

**Exit Code**:
- `0` - All checks passed
- `1` - Validation errors found

---

### sow schema

View embedded CUE schemas for `.sow/` file formats.

**Usage**:
```bash
sow schema show <type>
sow schema list
```

**Commands**:

**`sow schema show <type>`** - Display a specific schema:
```bash
# Show project state schema
sow schema show project

# Show task state schema
sow schema show task

# Show version file schema
sow schema show version

# Show sinks index schema
sow schema show sinks-index

# Show repos index schema
sow schema show repos-index
```

**`sow schema list`** - List all available schemas:
```bash
sow schema list
```

**Output Example**:
```bash
$ sow schema show version

// .sow/.version schema
#Version: {
	sow_structure_version: string
	plugin_version:        string
	last_migrated:         string | null
	initialized:           string
}
```

**Purpose**: Allows users to inspect the embedded CUE schemas that define file formats. Useful for:
- Understanding file structure requirements
- Debugging validation errors
- Writing correct configuration files
- Learning what fields are available

---

## Logging Commands

### sow log

Create log entry in project or task log.

**Usage**:
```bash
sow log [--file FILE]... --action ACTION --result RESULT [--project] "Description"
```

**Options**:
- `--file FILE` - File affected (can specify multiple)
- `--action ACTION` - Action type (from vocabulary)
- `--result RESULT` - Result (success, error, partial)
- `--project` - Write to project log (default: auto-detect from cwd)
- Description - Free-form notes (required, quoted)

**Examples**:

**Task log** (from task directory):
```bash
cd .sow/project/phases/implement/tasks/020/

sow log \
  --file src/auth/jwt.py \
  --action created_file \
  --result success \
  "Created JWT service with RS256 algorithm"
```

**Task log** (from project root):
```bash
sow log \
  --file src/auth/jwt.py \
  --file tests/test_jwt.py \
  --action modified_file \
  --result success \
  "Implemented token validation and added tests"
```

**Project log** (force project level):
```bash
sow log --project \
  --action started_phase \
  --result success \
  "Started implement phase with 4 tasks"
```

**Multiple files**:
```bash
sow log \
  --file src/auth/jwt.py \
  --file src/auth/__init__.py \
  --file requirements.txt \
  --action modified_file \
  --result success \
  "Added JWT service and dependency"
```

**Auto-Detection**:
- From task directory: writes to task log
- From project root: writes to project log
- Use `--project` to force project log

**What CLI Does**:
1. Determines context (task or project log)
2. Reads current iteration from task state.yaml (if task log)
3. Auto-constructs agent ID: `{role}-{iteration}`
4. Generates timestamp (ISO 8601)
5. Formats entry with required fields
6. Appends to appropriate log.md

---

### sow session-info

Display current session information (used by SessionStart hook).

**Usage**:
```bash
sow session-info
```

**Output (no project)**:
```
📋 You are in a sow-enabled repository
💡 No active project. Use /start-project <name> to begin

✓ Versions aligned (v0.2.0)

📖 Available commands:
   /start-project <name> - Create new project
   /continue - Resume existing project
   /cleanup - Delete project before merge
   /sync - Update sinks and repos
```

**Output (with project)**:
```
📋 You are in a sow-enabled repository
🚀 Active project: Add authentication (branch: feat/add-auth)
📂 Use /continue to resume work

✓ Versions aligned (v0.2.0)

📖 Available commands:
   /start-project <name> - Create new project
   /continue - Resume existing project
   /cleanup - Delete project before merge
   /sync - Update sinks and repos
```

**Output (version mismatch)**:
```
📋 You are in a sow-enabled repository
💡 No active project. Use /start-project <name> to begin

⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0

💡 Run /migrate to upgrade your repository structure
   Migration path: 0.1.0 → 0.2.0
   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md
```

**Purpose**: Provide context on session start via SessionStart hook.

---

## Sink Management

### sow sinks install

Install a knowledge sink from git repository or local path.

**Usage**:
```bash
sow sinks install <source> [<path>]
```

**Arguments**:
- `source` - Git URL or local path
- `path` - Optional: specific path within repo to copy

**Examples**:

**Install from git**:
```bash
sow sinks install https://github.com/your-org/python-style-guide
```

**Install specific path from git**:
```bash
sow sinks install https://github.com/your-org/standards style-guides/python
```

**Install from local path**:
```bash
sow sinks install /path/to/local/sink
```

**What Happens**:
1. Clones or copies sink to `.sow/sinks/<name>/`
2. Interrogates sink content (LLM reads and summarizes)
3. Updates `.sow/sinks/index.json` with metadata
4. Displays confirmation

**Output**:
```
📥 Installing sink from https://github.com/your-org/python-style-guide...

✓ Cloned to .sow/sinks/python-style/
✓ Interrogated content (3 files found)
✓ Updated index.json

Sink installed: python-style (v1.2.0)
Topics: formatting, naming, testing, imports
```

---

### sow sinks update

Check for and install sink updates.

**Usage**:
```bash
sow sinks update [<name>]
```

**Arguments**:
- `name` - Optional: specific sink to update (updates all if omitted)

**Examples**:

**Update all sinks**:
```bash
sow sinks update
```

**Update specific sink**:
```bash
sow sinks update python-style
```

**Output**:
```
🔍 Checking for updates...

Sinks:
  ✓ python-style (v1.2.0) - up to date
  ⬆️ api-conventions (v2.3.0 → v2.4.0) - Added GraphQL conventions
  ⬆️ deployment-guide (v1.0.0 → v1.1.0) - Updated k8s configs

Update all? [y/n/selective]

User: selective

Select sinks to update:
  [y] api-conventions
  [n] deployment-guide

Updating...
✓ api-conventions updated to v2.4.0
✓ Regenerated index

1 sink updated, 2 unchanged
```

---

### sow sinks list

List installed sinks.

**Usage**:
```bash
sow sinks list
```

**Output**:
```
Installed sinks:

  python-style (v1.2.0)
  - Python code style, formatting conventions, and testing standards
  - Topics: formatting, naming, testing, imports, docstrings
  - Source: https://github.com/your-org/python-style-guide

  api-conventions (v2.4.0)
  - REST and GraphQL API design standards and best practices
  - Topics: endpoints, errors, versioning, graphql, authentication
  - Source: https://github.com/your-org/api-standards

  deployment-guide (v1.1.0)
  - Kubernetes deployment configurations and CI/CD workflows
  - Topics: kubernetes, ci-cd, docker, deployment, monitoring
  - Source: https://github.com/your-org/deployment-standards

3 sinks installed
```

---

### sow sinks remove

Remove an installed sink.

**Usage**:
```bash
sow sinks remove <name>
```

**Arguments**:
- `name` - Sink identifier to remove

**Example**:
```bash
sow sinks remove python-style
```

**Output**:
```
⚠️  Remove sink 'python-style'? [y/n]

User: y

✓ Removed .sow/sinks/python-style/
✓ Updated index.json

Sink removed: python-style
```

---

### sow sinks reindex

Regenerate sink index by re-interrogating all sinks.

**Usage**:
```bash
sow sinks reindex
```

**Output**:
```
🔄 Regenerating sink index...

✓ Interrogated python-style (3 files)
✓ Interrogated api-conventions (5 files)
✓ Interrogated deployment-guide (4 files)
✓ Updated index.json

Index regenerated with 3 sinks
```

**When to Use**:
- Index corrupted or out of sync
- Manually added/modified sinks
- After bulk sink operations

---

## Repository Management

### sow repos add

Add a linked repository.

**Usage**:
```bash
sow repos add <source> [--symlink]
```

**Arguments**:
- `source` - Git URL or local path
- `--symlink` - Create symlink instead of cloning (for local paths)

**Examples**:

**Clone repository**:
```bash
sow repos add https://github.com/your-org/auth-service
```

**Symlink local repository**:
```bash
sow repos add /path/to/shared-library --symlink
```

**What Happens**:
1. Clones or symlinks repository to `.sow/repos/<name>/`
2. Adds entry to `.sow/repos/index.json`
3. Displays confirmation

**Output**:
```
📥 Adding repository from https://github.com/your-org/auth-service...

✓ Cloned to .sow/repos/auth-service/
✓ Updated index.json

Repository added: auth-service
Purpose: Reference authentication implementation patterns
```

---

### sow repos sync

Sync linked repositories (pull latest changes).

**Usage**:
```bash
sow repos sync [<name>]
```

**Arguments**:
- `name` - Optional: specific repo to sync (syncs all if omitted)

**Examples**:

**Sync all repos**:
```bash
sow repos sync
```

**Sync specific repo**:
```bash
sow repos sync auth-service
```

**Output**:
```
🔍 Syncing repositories...

Repositories:
  ✓ auth-service - up to date
  ⬆️ shared-library (abc1234 → def5678) - Added new crypto utilities

Update all? [y/n/selective]

User: y

Updating...
✓ shared-library updated to def5678
✓ Updated index.json

1 repository updated, 1 unchanged
```

---

### sow repos list

List linked repositories.

**Usage**:
```bash
sow repos list
```

**Output**:
```
Linked repositories:

  auth-service
  - Reference authentication implementation patterns and middleware
  - Type: clone
  - Source: https://github.com/your-org/auth-service
  - Branch: main

  shared-library
  - Shared cryptography and utility functions
  - Type: symlink
  - Source: /Users/you/code/shared-library

2 repositories linked
```

---

### sow repos remove

Remove a linked repository.

**Usage**:
```bash
sow repos remove <name>
```

**Arguments**:
- `name` - Repository identifier to remove

**Example**:
```bash
sow repos remove auth-service
```

**Output**:
```
⚠️  Remove repository 'auth-service'? [y/n]

User: y

✓ Removed .sow/repos/auth-service/
✓ Updated index.json

Repository removed: auth-service
```

---

### sow repos reindex

Regenerate repository index.

**Usage**:
```bash
sow repos reindex
```

**Output**:
```
🔄 Regenerating repository index...

✓ Indexed auth-service
✓ Indexed shared-library
✓ Updated index.json

Index regenerated with 2 repositories
```

---

## Utility Commands

### sow save-context

Save important context before compaction (used by PreCompact hook).

**Usage**:
```bash
sow save-context
```

**Output**:
```
💾 Saving context before compaction...

✓ Logged compaction event to project log

Context saved
```

**Purpose**: Hook for preserving context before Claude Code compaction.

---

### sow sync

Sync both sinks and repositories.

**Usage**:
```bash
sow sync
```

**Output**:
```
🔍 Checking for updates...

Sinks:
  ✓ python-style (v1.2.0) - up to date
  ⬆️ api-conventions (v2.3.0 → v2.4.0) - Added GraphQL conventions

Repositories:
  ✓ auth-service - up to date
  ⬆️ shared-library (abc1234 → def5678) - Added new crypto utilities

Update all? [y/n/selective]

User: y

Updating...
✓ api-conventions updated to v2.4.0
✓ shared-library updated to def5678
✓ Regenerated indexes

2 items updated, 2 unchanged
```

**Convenience**: Combines `sow sinks update` and `sow repos sync`.

---

## Command-Line Flags

### Global Flags

Available for all commands:

| Flag | Description |
|------|-------------|
| `--help` | Show command-specific help |
| `--quiet` | Suppress output (only errors) |
| `--verbose` | Verbose output (for debugging) |
| `--no-color` | Disable colored output |

**Examples**:
```bash
# Show help for specific command
sow log --help

# Quiet mode (only errors shown)
sow validate --quiet

# Verbose mode (debug output)
sow sinks install https://example.com/sink --verbose

# No color (for scripts/CI)
sow validate --no-color
```

### Command-Specific Flags

See individual command documentation above for command-specific flags.

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

**Example**:
```bash
# Disable color
export SOW_NO_COLOR=1
sow validate

# Debug logging
export SOW_LOG_LEVEL=debug
sow sinks install https://example.com/sink

# Short timeout
export SOW_TIMEOUT=10
sow repos sync
```

### Credential Variables

For MCP and external integrations:

| Variable | Purpose |
|----------|---------|
| `GITHUB_TOKEN` | GitHub API authentication |
| `JIRA_TOKEN` | Jira API authentication |
| `DATADOG_API_KEY` | Datadog API authentication |

**Usage**: Set in shell profile, referenced in `mcp.json`:
```json
{
  "mcpServers": {
    "github": {
      "transport": "http",
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    }
  }
}
```

---

## Related Documentation

- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows using CLI
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Logging system details
- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - CLI installation and updates
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - SessionStart hook usage
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Slash commands
- **[SCHEMAS.md](./SCHEMAS.md)** - File formats and schemas
