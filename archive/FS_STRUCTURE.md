# Filesystem Structure (DRAFT)

**Status**: Draft - Not Authoritative
**Last Updated**: 2025-10-12
**Purpose**: Working document for refining the `sow` filesystem organization

This document represents our current thinking about how `sow` organizes files within a repository. It is subject to change as we continue exploring the design space.

---

## Overview

`sow` uses two primary directories:

- **`.claude/`** - Execution layer (agents, commands, hooks, integrations)
- **`.sow/`** - Data layer (knowledge, sinks, projects, state)

The execution layer is distributed via Claude Code Plugin and committed to git.
The data layer contains both committed (knowledge, projects) and git-ignored (sinks, repos) content.

---

## Complete Structure

```
project-root/
├── .claude/                              # EXECUTION LAYER (committed)
│   ├── .claude-plugin/
│   │   └── plugin.json                  # Plugin metadata (name, version, author)
│   │
│   ├── agents/                          # Sub-agent definitions
│   │   ├── orchestrator.yaml            # Main coordinating agent
│   │   ├── architect.yaml               # Design/architecture worker
│   │   ├── implementer.yaml             # Coding worker
│   │   ├── reviewer.yaml                # Code review worker
│   │   └── documenter.yaml              # Documentation worker
│   │
│   ├── commands/                        # Slash commands
│   │   ├── workflows/                   # High-level user workflows
│   │   │   ├── init.md                  # Bootstrap sow in repository
│   │   │   ├── start-project.md         # Start new project
│   │   │   ├── continue.md              # Resume existing project
│   │   │   ├── cleanup.md               # Delete project, prepare for merge
│   │   │   ├── migrate.md               # Migrate sow versions
│   │   │   └── sync.md                  # Sync sinks/repos
│   │   │
│   │   └── skills/                      # Granular agent capabilities
│   │       ├── create-adr.md            # Create Architecture Decision Record
│   │       ├── design-doc.md            # Write design document
│   │       ├── write-tests.md           # Generate test suite
│   │       └── refactor.md              # Refactoring assistance
│   │
│   ├── hooks.json                       # Event automation (pre/post tool use, etc.)
│   └── mcp.json                         # External tool integrations (MCP servers)
│
└── .sow/                                # DATA LAYER (mixed committed/ignored)
    │
    ├── knowledge/                       # Repo-specific knowledge (COMMITTED)
    │   ├── overview.md                  # Project overview
    │   ├── architecture/                # Architecture documents
    │   │   ├── system-overview.md
    │   │   └── diagrams/
    │   ├── adrs/                        # Architecture Decision Records
    │   │   ├── 001-use-postgres.md
    │   │   └── 002-api-versioning.md
    │   └── deviations.md                # Deviations from installed sinks
    │
    ├── sinks/                           # External knowledge (GIT-IGNORED)
    │   ├── index.json                   # LLM-maintained sink catalog
    │   ├── python-style/                # Example: Python style guide
    │   │   ├── formatting.md
    │   │   └── conventions.md
    │   ├── api-conventions/             # Example: API design standards
    │   └── deployment-guide/            # Example: Deployment process
    │
    ├── repos/                           # Linked repositories (GIT-IGNORED)
    │   ├── index.json                   # Repository references
    │   ├── auth-service/                # Cloned or symlinked repo
    │   └── shared-library/              # Cloned or symlinked repo
    │
    └── project/                         # Active work (GIT-IGNORED)
        ├── state.yaml                   # Project state (phases, tasks, status)
        ├── log.md                       # Append-only orchestrator action log
        │
        ├── context/                     # Project-specific context
        │   ├── overview.md              # Project overview
        │   ├── decisions.md             # Key decisions made
        │   └── memories.md              # Important context to remember
        │
        └── phases/                      # Phases contain tasks
            ├── discovery/               # Example phase
            │   └── tasks/
            │       └── 010/
            │           ├── state.yaml         # Task state
            │           ├── log.md             # Worker action log
            │           ├── description.md     # Task description
            │           └── feedback/          # Human corrections
            │               ├── 001.md
            │               └── 002.md
            │
            └── implement/               # Example phase
                └── tasks/
                    ├── 010/
                    │   └── ...
                    └── 020/
                        └── ...
```

---

## Directory Purposes

### `.claude/` (Execution Layer)

**Purpose**: Defines how AI agents behave and interact with the system

**Committed to Git**: Yes (entire directory)

**Components**:
- **plugin.json**: Metadata for distribution via Claude Code plugin system
- **agents/**: Sub-agent definitions (roles, tools, system prompts)
- **commands/workflows/**: User-invoked workflows (init, start project, etc.)
- **commands/skills/**: Agent-invoked capabilities (create ADR, write tests, etc.)
- **hooks.json**: Event-driven automation (formatting, validation, notifications)
- **mcp.json**: External tool integrations (GitHub, Jira, monitoring, etc.)

---

### `.sow/` (Data Layer)

**Purpose**: Stores project knowledge, external context, and work state

**Git Status**: Mixed (some committed, some ignored)

#### `.sow/knowledge/` (COMMITTED)

Repository-specific documentation and decisions:
- **overview.md**: High-level project description
- **architecture/**: Design documents, diagrams, system architecture
- **adrs/**: Architecture Decision Records (numbered chronologically)
- **deviations.md**: Project-specific deviations from installed sinks

#### `.sow/sinks/` (GIT-IGNORED)

External knowledge installed per-developer:
- **index.json**: LLM-maintained catalog (summaries, "when to use", metadata)
- **Sink directories**: Collections of markdown files on specific topics
- Examples: style guides, deployment procedures, API conventions, security checklists

**Installation**: Via CLI (`sow sinks install <source>`) or slash command
**Updates**: Manual via `sow sinks update` (similar to package managers)
**Discovery**: Orchestrator reads index.json, routes relevant sinks to workers

#### `.sow/repos/` (GIT-IGNORED)

External repositories needed for context:
- **index.json**: Repository references (git URLs, paths, purpose)
- **Repository directories**: Cloned or symlinked repositories
- **Use case**: Multi-repo projects where agents need cross-repo context

**Purpose**: Turn multi-repo setup into pseudo-monorepo for agent context

#### `.sow/project/` (COMMITTED TO FEATURE BRANCHES)

Active work coordination:
- **Single Project Only**: One project per branch enforced (no `projects/` plural)
- **Committed**: Project state committed to feature branches (enables branch switching)
- **Ephemeral**: Deleted before merge (CI enforced)
- **Mandatory**: One branch = one project (enforced, not suggested)
- **Safeguard**: Cannot create project on `main`/`master` branches
- **Team Collaboration**: Project state shared when pushing branches
- **Cleanup**: Use `/cleanup` command before merge, CI fails if project exists

**Project Structure**:
- **state.yaml**: Central planning document (includes branch name, phases, tasks, status)
- **log.md**: Write-ahead log of orchestrator actions
- **context/**: Project-specific context, decisions, memories
- **phases/**: Logical divisions of work (named phases like discovery, design, implement)

**Task Structure** (within phases):
- **state.yaml**: Task metadata, iteration counter, context references, feedback tracking
- **log.md**: Write-ahead log of worker actions (CLI-generated)
- **description.md**: Focused task description with requirements and acceptance criteria
- **feedback/**: Chronologically numbered human corrections (001.md, 002.md, etc.)

---

## Key Design Decisions

### Phases as Directories
- Phases are first-class directories, not just logical groupings in YAML
- Rationale: Filesystem discoverability, zero-context resumability, no cross-referencing needed

### Skills = Slash Commands
- No separate "skills" system
- Skills are slash commands that agents reference in their system prompts
- Prevents context window bloat while maintaining composability

### Single Project Per Branch (Enforced)
- Only `.sow/project/` (singular) - one project at a time
- **Committed to feature branches** (enables git to switch project state with branches)
- Deleted before merge via `/cleanup` command (CI enforced)
- Enforces one branch = one project as mandatory constraint
- Cannot create on `main`/`master` (forces feature branch workflow)
- Simplifies orchestrator (no "which project?" logic)
- Branch switching = automatic project context switching (git handles it)
- Team collaboration: project state shared when pushing branches
- Recommend squash-merge to keep main branch history clean
- Not a replacement for existing productivity tools (JIRA, SCRUM, etc.)

### Separate Execution and Data Layers
- Execution layer (`.claude/`) can be upgraded wholesale
- Data layer (`.sow/`) persists and evolves independently
- Migration handled via `/migrate` command when structure changes

---

## Git Versioning Summary

| Path | Committed? | Reason |
|------|-----------|--------|
| `.claude/` | ✅ Yes (all branches) | Team shares agents, commands, hooks |
| `.sow/knowledge/` | ✅ Yes (all branches) | Repo-specific docs, architecture, ADRs |
| `.sow/sinks/` | ❌ No | External, per-developer installations |
| `.sow/repos/` | ❌ No | External repositories, too large |
| `.sow/project/` | ✅ Yes (feature branches only) | Per-branch work state, deleted before merge to main, CI enforced |

### CLI-Driven Logging
- Agents use CLI for fast logging instead of direct file editing
- Format: `sow log --file <path> --action <action> --result <result> "notes"`
- CLI auto-constructs agent ID from iteration counter
- Rationale: Direct file editing is slow (30+ seconds), CLI is instant

---

## State File Schemas

### Project state.yaml

```yaml
# .sow/project/state.yaml

project:
  name: add-authentication
  branch: feat/add-auth  # Branch this project belongs to
  created_at: 2025-10-12T14:30:00Z
  updated_at: 2025-10-12T16:45:00Z
  description: Add JWT-based authentication system

  # Complexity rating from initial planning
  complexity:
    rating: 2  # 1=simple, 2=moderate, 3=complex
    metrics:
      estimated_files: 8
      cross_cutting: true
      new_dependencies: true

  # Current active phase
  active_phase: implement

phases:
  - name: design
    status: completed  # pending, in_progress, completed
    created_at: 2025-10-12T14:32:00Z
    completed_at: 2025-10-12T15:20:00Z
    tasks:
      - id: "010"
        name: Design authentication flow
        status: completed
        parallel: false

  - name: implement
    status: in_progress
    created_at: 2025-10-12T15:22:00Z
    tasks:
      - id: "010"
        name: Create User model
        status: completed
        parallel: false

      - id: "020"
        name: Create JWT service
        status: in_progress
        parallel: false

      - id: "030"
        name: Add login endpoint
        status: pending
        parallel: true  # Can run parallel with 031

      - id: "031"
        name: Add password hashing utility
        status: pending
        parallel: true  # Can run parallel with 030

  - name: test
    status: pending
    tasks: []
```

### Task state.yaml

```yaml
# .sow/project/phases/implement/tasks/020/state.yaml

task:
  id: "020"
  name: Create JWT service
  phase: implement
  status: in_progress  # pending, in_progress, completed, abandoned

  created_at: 2025-10-12T15:25:00Z
  started_at: 2025-10-12T15:30:00Z
  updated_at: 2025-10-12T16:45:00Z

  # Iteration counter (managed by orchestrator)
  iteration: 3  # Third attempt at this task
  assigned_agent: implementer  # Role name

  # Context references (relative to .sow/ root)
  # Orchestrator compiles this list for worker
  references:
    - sinks/python-style/conventions.md
    - sinks/api-conventions/rest-standards.md
    - knowledge/architecture/auth-design.md
    - repos/shared-library/src/crypto/jwt.py

  # Feedback tracking
  feedback:
    - id: "001"
      created_at: 2025-10-12T16:30:00Z
      status: addressed  # pending, addressed, superseded

  # Files modified (auto-populated by worker during task)
  files_modified:
    - src/auth/jwt.py
    - tests/test_jwt.py
```

### Task description.md

```markdown
# .sow/project/phases/implement/tasks/020/description.md

## Task: Create JWT Service

Create a JWT token generation and validation service following our API conventions.

### Requirements
- Token expiration: 1 hour
- Refresh token: 7 days
- Use RS256 algorithm
- Include user ID and role in payload

### Acceptance Criteria
- [ ] Generate access and refresh tokens
- [ ] Validate token signature
- [ ] Handle expired tokens gracefully
- [ ] Unit tests with >90% coverage

### Context Notes
- Reference the shared JWT library in `shared-library` repo for patterns
- Follow Python conventions in our style guide
- Auth flow design is documented in `knowledge/architecture/auth-design.md`
```

### Task log.md (CLI-Generated)

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
**Result**: success

Created initial JWT service file with class structure and method stubs.

---

### 2025-10-12 15:42:03

**Agent**: implementer-3
**Action**: implementation_attempt
**Files**:
  - src/auth/jwt.py
**Result**: error

Attempted to implement token generation. Encountered import error with cryptography library.

---
```

---

## Open Questions

1. **User-Level Configuration**: Should there be `~/.config/sow/` or `~/.local/state/sow/`?
2. **Project Naming**: Convention for project directory names? (kebab-case, feature branch name, etc.)
3. **Sink Index Schema**: What fields in index.json? (name, description, tags, version, etc.)
4. **Repos Index Schema**: What fields in index.json for linked repositories?

---

## Related Documentation

- **`PROJECT_LIFECYCLE.md`**: Detailed operational lifecycle and workflows
- **`BRAINSTORMING.md`**: Discovery session notes and design exploration
