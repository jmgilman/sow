# File Format Schemas

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Project State Schema](#project-state-schema)
- [Task State Schema](#task-state-schema)
- [Task Description Format](#task-description-format)
- [Task Log Format](#task-log-format)
- [Sink Index Schema](#sink-index-schema)
- [Repository Index Schema](#repository-index-schema)
- [Version File Schema](#version-file-schema)
- [Plugin Metadata Schema](#plugin-metadata-schema)
- [Hooks Configuration Schema](#hooks-configuration-schema)
- [MCP Configuration Schema](#mcp-configuration-schema)
- [Related Documentation](#related-documentation)

---

## Overview

This document provides complete specifications for all file formats used in `sow`. All examples use valid YAML, JSON, or Markdown syntax.

**File Format Summary**:

| File | Format | Purpose |
|------|--------|---------|
| `state.yaml` (project) | YAML | Project state and task list |
| `state.yaml` (task) | YAML | Individual task metadata |
| `description.md` (task) | Markdown | Task requirements |
| `log.md` | Markdown | Action logs (structured) |
| `index.json` (sinks) | JSON | Sink catalog |
| `index.json` (repos) | JSON | Repository references |
| `.version` | YAML | Version tracking |
| `plugin.json` | JSON | Plugin metadata |
| `hooks.json` | JSON | Hook configuration |
| `mcp.json` | JSON | MCP server configuration |

---

## Project State Schema

**File**: `.sow/project/state.yaml`

**Purpose**: Central planning document containing all project metadata, phases, and tasks.

### Full Schema

```yaml
# Project metadata
project:
  name: string                    # Project name (kebab-case recommended)
  branch: string                  # Git branch this project belongs to
  created_at: timestamp           # ISO 8601: 2025-10-12T14:30:00Z
  updated_at: timestamp           # ISO 8601: 2025-10-12T16:45:00Z
  description: string             # Human-readable description

  # Complexity assessment (from initial planning)
  complexity:
    rating: integer               # 1=simple, 2=moderate, 3=complex
    metrics:
      estimated_files: integer    # Estimated number of files
      cross_cutting: boolean      # Cross-cutting concerns exist
      new_dependencies: boolean   # New dependencies required

  # Current active phase
  active_phase: string            # Name of currently active phase

# Phases array
phases:
  - name: string                  # Phase name (discovery, design, implement, test, review, deploy, document)
    status: string                # pending | in_progress | completed
    created_at: timestamp         # When phase was created
    completed_at: timestamp | null  # When phase completed (null if not done)

    # Tasks within this phase
    tasks:
      - id: string                # Gap-numbered ID (e.g., "010", "020", "030")
        name: string              # Task name/description
        status: string            # pending | in_progress | completed | abandoned
        parallel: boolean         # Can run in parallel with other tasks
        assigned_agent: string    # Agent role (architect, implementer, etc.)
```

### Complete Example

```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  created_at: 2025-10-12T14:30:00Z
  updated_at: 2025-10-12T16:45:00Z
  description: Add JWT-based authentication system with user login and token refresh

  complexity:
    rating: 2
    metrics:
      estimated_files: 8
      cross_cutting: true
      new_dependencies: true

  active_phase: implement

phases:
  - name: design
    status: completed
    created_at: 2025-10-12T14:32:00Z
    completed_at: 2025-10-12T15:20:00Z
    tasks:
      - id: "010"
        name: Design authentication flow
        status: completed
        parallel: false
        assigned_agent: architect

  - name: implement
    status: in_progress
    created_at: 2025-10-12T15:22:00Z
    completed_at: null
    tasks:
      - id: "010"
        name: Create User model
        status: completed
        parallel: false
        assigned_agent: implementer

      - id: "020"
        name: Create JWT service
        status: in_progress
        parallel: false
        assigned_agent: implementer

      - id: "030"
        name: Add login endpoint
        status: pending
        parallel: true
        assigned_agent: implementer

      - id: "031"
        name: Add password hashing utility
        status: pending
        parallel: true
        assigned_agent: implementer

  - name: test
    status: pending
    created_at: null
    completed_at: null
    tasks: []
```

### Field Descriptions

**project.name**: Unique identifier for project (used in logs, display)

**project.branch**: Git branch name. Orchestrator validates current branch matches this field.

**complexity.rating**:
- `1` - Simple (few files, focused scope, no new deps)
- `2` - Moderate (multiple files, some integration, maybe new deps)
- `3` - Complex (many files, architectural changes, cross-cutting)

**active_phase**: Must match one of the phase names in `phases` array. Orchestrator uses this to determine which phase is currently being worked on.

**phases[].status**:
- `pending` - Not yet started
- `in_progress` - Currently active
- `completed` - All tasks done

**tasks[].status**:
- `pending` - Not yet started
- `in_progress` - Currently being worked on
- `completed` - Successfully finished
- `abandoned` - Started but deprecated (not deleted)

**tasks[].parallel**: If `true`, can run simultaneously with other `parallel: true` tasks in same phase.

---

## Task State Schema

**File**: `.sow/project/phases/<phase>/tasks/<id>/state.yaml`

**Purpose**: Task-specific metadata including iteration tracking, context references, and feedback status.

### Full Schema

```yaml
task:
  id: string                      # Task ID (matches directory name)
  name: string                    # Task name
  phase: string                   # Phase name this task belongs to
  status: string                  # pending | in_progress | completed | abandoned

  created_at: timestamp           # When task was created
  started_at: timestamp | null    # When work first began
  updated_at: timestamp           # Last modification
  completed_at: timestamp | null  # When task completed (null if not done)

  iteration: integer              # Attempt counter (managed by orchestrator)
  assigned_agent: string          # Agent role

  # Context references (relative to .sow/ root)
  references: array[string]       # File paths for worker to read

  # Feedback tracking
  feedback: array
    - id: string                  # Feedback number (e.g., "001", "002")
      created_at: timestamp       # When feedback was created
      status: string              # pending | addressed | superseded

  # Files modified (auto-populated by worker)
  files_modified: array[string]   # Paths to files changed during task
```

### Complete Example

```yaml
task:
  id: "020"
  name: Create JWT service
  phase: implement
  status: in_progress

  created_at: 2025-10-12T15:25:00Z
  started_at: 2025-10-12T15:30:00Z
  updated_at: 2025-10-12T16:45:00Z
  completed_at: null

  iteration: 3
  assigned_agent: implementer

  references:
    - sinks/python-style/conventions.md
    - sinks/api-conventions/rest-standards.md
    - knowledge/architecture/auth-design.md
    - repos/shared-library/src/crypto/jwt.py

  feedback:
    - id: "001"
      created_at: 2025-10-12T16:30:00Z
      status: addressed
    - id: "002"
      created_at: 2025-10-12T17:15:00Z
      status: pending

  files_modified:
    - src/auth/jwt.py
    - tests/test_jwt.py
    - requirements.txt
```

### Field Descriptions

**iteration**: Tracks attempt number. Incremented by orchestrator before spawning new worker. Used to construct agent ID: `{assigned_agent}-{iteration}` (e.g., `implementer-3`).

**references**: Paths relative to `.sow/` root. Orchestrator compiles this list. Worker reads all referenced files at startup.

**feedback[].status**:
- `pending` - Not yet addressed by worker
- `addressed` - Worker has incorporated feedback
- `superseded` - No longer relevant (newer feedback replaces)

**files_modified**: Auto-populated by worker. Helps track changes and impacts.

---

## Task Description Format

**File**: `.sow/project/phases/<phase>/tasks/<id>/description.md`

**Purpose**: Focused task description with requirements and acceptance criteria.

### Format

```markdown
## Task: <Task Name>

<Brief description of what needs to be done>

### Requirements

- Requirement 1
- Requirement 2
- ...

### Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- ...

### Context Notes

<Any additional context, references, or guidance>
```

### Complete Example

```markdown
## Task: Create JWT Service

Create a JWT token generation and validation service following our API conventions.

### Requirements

- Token expiration: 1 hour
- Refresh token: 7 days
- Use RS256 algorithm (asymmetric)
- Include user ID and role in payload
- Support token refresh without re-authentication
- Handle expired tokens gracefully

### Acceptance Criteria

- [ ] Generate access and refresh tokens
- [ ] Validate token signature
- [ ] Handle expired tokens gracefully
- [ ] Extract claims from valid tokens
- [ ] Unit tests with >90% coverage
- [ ] Integration tests for token refresh flow

### Context Notes

- Reference the shared JWT library in `shared-library` repo for patterns
- Follow Python conventions in our style guide
- Auth flow design is documented in `knowledge/architecture/auth-design.md`
- Use environment variables for key loading (keys managed via Vault)
- See security guidelines for JWT best practices
```

### Guidelines

- **Be specific**: Clear requirements prevent ambiguity
- **Actionable criteria**: Worker can verify each item
- **Provide context**: References to relevant docs/code
- **Keep focused**: One clear task, not multiple concerns

---

## Task Log Format

**File**: `.sow/project/phases/<phase>/tasks/<id>/log.md`

**Purpose**: Chronological record of all actions taken during task execution.

### Format

Each log entry uses structured markdown:

```markdown
### <Timestamp: YYYY-MM-DD HH:MM:SS>

**Agent**: <agent-id>
**Action**: <action-type>
**Files**: (optional)
  - <file1>
  - <file2>
**Result**: <success | error | partial>

<Free-form description/notes>

---
```

### Complete Example

```markdown
# Task Log

Chronological record of all actions taken during this task.

---

### 2025-10-12 15:30:42

**Agent**: implementer-3
**Action**: started_task
**Result**: success

Started implementing JWT service. Reviewed requirements and acceptance criteria.

---

### 2025-10-12 15:35:18

**Agent**: implementer-3
**Action**: created_file
**Files**:
  - src/auth/jwt.py
  - tests/test_jwt.py
**Result**: success

Created initial JWT service file with class structure and method stubs. Created test file with skeleton test cases.

---

### 2025-10-12 15:42:03

**Agent**: implementer-3
**Action**: implementation_attempt
**Files**:
  - src/auth/jwt.py
  - requirements.txt
**Result**: error

Attempted to implement token generation. Encountered import error with cryptography library. Added dependency to requirements.txt.

---

### 2025-10-12 15:48:15

**Agent**: implementer-3
**Action**: modified_file
**Files**:
  - src/auth/jwt.py
  - tests/test_jwt.py
**Result**: success

Fixed import issue. Implemented generate_token() and validate_token() methods using RS256 algorithm. Updated tests.

---

### 2025-10-12 15:55:30

**Agent**: implementer-3
**Action**: test_run
**Files**:
  - tests/test_jwt.py
**Result**: success

Ran unit tests. All 12 tests passing. Coverage: 95%.

---

### 2025-10-12 16:00:00

**Agent**: implementer-3
**Action**: completed_task
**Result**: success

JWT service implementation complete. All acceptance criteria met.

---
```

### Action Vocabulary

Standard action types (extensible):

| Action | When to Use |
|--------|-------------|
| `started_task` | Beginning work |
| `created_file` | New file created |
| `modified_file` | Existing file changed |
| `deleted_file` | File removed |
| `implementation_attempt` | Attempted implementation |
| `test_run` | Ran tests |
| `refactor` | Code refactoring |
| `debugging` | Debugging session |
| `research` | Investigation/research |
| `completed_task` | Task finished |
| `paused_task` | Paused for feedback |

### CLI Generation

Workers use CLI to generate log entries:

```bash
sow log \
  --file src/auth/jwt.py \
  --action modified_file \
  --result success \
  "Implemented token generation with RS256"
```

CLI auto-constructs agent ID and timestamp.

---

## Sink Index Schema

**File**: `.sow/sinks/index.json`

**Purpose**: LLM-maintained catalog of installed sinks with metadata for agent discovery.

### Schema

```json
{
  "sinks": [
    {
      "name": "string",           // Sink identifier (kebab-case)
      "path": "string",           // Relative path from .sow/sinks/
      "description": "string",    // Human-readable description
      "topics": ["string"],       // Topics covered
      "when_to_use": "string",    // Guidance for when to reference
      "version": "string",        // Version or git ref
      "source": "string",         // Git URL or path
      "updated_at": "timestamp"   // ISO 8601
    }
  ]
}
```

### Complete Example

```json
{
  "sinks": [
    {
      "name": "python-style",
      "path": "python-style",
      "description": "Python code style, formatting conventions, and testing standards",
      "topics": ["formatting", "naming", "testing", "imports", "docstrings"],
      "when_to_use": "When writing or reviewing Python code",
      "version": "v1.2.0",
      "source": "https://github.com/your-org/python-style-guide",
      "updated_at": "2025-10-12T14:00:00Z"
    },
    {
      "name": "api-conventions",
      "path": "api-conventions",
      "description": "REST and GraphQL API design standards and best practices",
      "topics": ["endpoints", "errors", "versioning", "graphql", "authentication"],
      "when_to_use": "When designing or implementing APIs",
      "version": "v2.4.0",
      "source": "https://github.com/your-org/api-standards",
      "updated_at": "2025-10-12T15:30:00Z"
    },
    {
      "name": "deployment-guide",
      "path": "deployment-guide",
      "description": "Kubernetes deployment configurations and CI/CD workflows",
      "topics": ["kubernetes", "ci-cd", "docker", "deployment", "monitoring"],
      "when_to_use": "When preparing deployments or writing CI/CD configs",
      "version": "v1.1.0",
      "source": "https://github.com/your-org/deployment-standards",
      "updated_at": "2025-10-12T16:00:00Z"
    }
  ]
}
```

### Usage

**Orchestrator reads index** to determine relevant sinks for tasks.

**Example**: Task requires Python code
- Orchestrator finds `python-style` via `topics: ["formatting", ...]` and `when_to_use`
- Adds `sinks/python-style/conventions.md` to task references

---

## Repository Index Schema

**File**: `.sow/repos/index.json`

**Purpose**: References to linked repositories for cross-repo context.

### Schema

```json
{
  "repositories": [
    {
      "name": "string",           // Repository identifier (kebab-case)
      "path": "string",           // Relative path from .sow/repos/
      "source": "string",         // Git URL or local path
      "purpose": "string",        // Why this repo is linked
      "type": "string",           // clone | symlink
      "branch": "string",         // Branch/ref (optional)
      "updated_at": "timestamp"   // ISO 8601
    }
  ]
}
```

### Complete Example

```json
{
  "repositories": [
    {
      "name": "auth-service",
      "path": "auth-service",
      "source": "https://github.com/your-org/auth-service",
      "purpose": "Reference authentication implementation patterns and middleware",
      "type": "clone",
      "branch": "main",
      "updated_at": "2025-10-12T14:00:00Z"
    },
    {
      "name": "shared-library",
      "path": "shared-library",
      "source": "/Users/you/code/shared-library",
      "purpose": "Shared cryptography and utility functions",
      "type": "symlink",
      "branch": null,
      "updated_at": "2025-10-12T14:00:00Z"
    }
  ]
}
```

### Type Values

- `clone` - Repository cloned to `.sow/repos/<name>/`
- `symlink` - Symlink to local repository

---

## Version File Schema

**File**: `.sow/.version`

**Purpose**: Track structure version and migration history.

### Schema

```yaml
sow_structure_version: string     # Structure version (semantic)
plugin_version: string            # Plugin version used
last_migrated: timestamp          # ISO 8601 (last migration)
initialized: timestamp            # ISO 8601 (first init)
```

### Complete Example

```yaml
sow_structure_version: 0.2.0
plugin_version: 0.2.0
last_migrated: 2025-10-12T16:30:00Z
initialized: 2025-10-12T14:00:00Z
```

### Usage

- Created by `/init`
- Updated by `/migrate`
- Read by SessionStart hook for version checking
- Committed to git

---

## Plugin Metadata Schema

**File**: `.claude-plugin/plugin.json`

**Purpose**: Plugin metadata for distribution via Claude Code marketplace.

### Schema

```json
{
  "name": "string",               // Plugin identifier (kebab-case, required)
  "version": "string",            // Semantic version (required)
  "description": "string",        // Short description (required)
  "author": {                     // Author info (required)
    "name": "string",
    "email": "string"
  },
  "homepage": "string",           // Homepage URL (optional)
  "repository": "string",         // Git repo URL (optional)
  "license": "string",            // License identifier (optional)
  "keywords": ["string"],         // Search keywords (optional)
  "engines": {                    // Requirements (optional)
    "claude-code": "string"       // Min version (e.g., ">=1.0.0")
  }
}
```

### Complete Example

```json
{
  "name": "sow",
  "version": "0.2.0",
  "description": "AI-powered system of work for software engineering",
  "author": {
    "name": "sow contributors",
    "email": "maintainers@example.com"
  },
  "homepage": "https://github.com/your-org/sow",
  "repository": "https://github.com/your-org/sow",
  "license": "MIT",
  "keywords": ["productivity", "workflow", "agents", "project-management"],
  "engines": {
    "claude-code": ">=1.0.0"
  }
}
```

---

## Hooks Configuration Schema

**File**: `hooks.json`

**Purpose**: Configure event-driven hooks for automation.

### Schema

```json
{
  "<HookEventName>": {
    "matcher": "string",          // Pattern to match (* for all)
    "command": "string"           // Shell command to execute
  }
}
```

### Complete Example

```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  },
  "PostToolUse": {
    "matcher": "Edit",
    "command": "prettier --write $FILE"
  },
  "PreToolUse": {
    "matcher": "Write",
    "command": "check-permissions.sh $FILE"
  },
  "PreCompact": {
    "matcher": "*",
    "command": "sow save-context"
  }
}
```

### Available Hooks

- `SessionStart` - Session starts
- `SessionEnd` - Session ends
- `PreToolUse` - Before tool execution
- `PostToolUse` - After tool execution
- `UserPromptSubmit` - User submits prompt
- `Notification` - Notification shown
- `Stop` - Execution stops
- `SubagentStop` - Subagent stops
- `PreCompact` - Before context compaction

---

## MCP Configuration Schema

**File**: `mcp.json`

**Purpose**: Configure Model Context Protocol server integrations.

### Schema

```json
{
  "mcpServers": {
    "<server-name>": {
      "transport": "string",      // http | stdio
      "url": "string",            // HTTP URL (for http transport)
      "command": "string",        // Command (for stdio transport)
      "args": ["string"],         // Args (for stdio transport)
      "headers": {                // HTTP headers (optional)
        "string": "string"
      }
    }
  }
}
```

### Complete Example

```json
{
  "mcpServers": {
    "github": {
      "transport": "http",
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    },
    "jira": {
      "transport": "http",
      "url": "https://your-domain.atlassian.net/mcp",
      "headers": {
        "Authorization": "Bearer ${JIRA_TOKEN}"
      }
    },
    "git-local": {
      "transport": "stdio",
      "command": "node",
      "args": ["/path/to/git-mcp-server/index.js"]
    }
  }
}
```

### Transport Types

**http**: Remote HTTP-based MCP server
- Requires `url`
- Optional `headers` for authentication

**stdio**: Local process-based MCP server
- Requires `command` and `args`
- Server runs as subprocess

---

## Related Documentation

- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle and state management
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day usage
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Commands and workflows
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - Hooks and MCP integrations
- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - Plugin packaging and versioning
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI commands
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Directory layout
