# sow File Structure

**Last Updated**: 2025-10-12
**Purpose**: Complete directory layout and organization reference

This document provides the authoritative reference for how `sow` organizes files within a repository.

---

## Table of Contents

- [Complete Structure](#complete-structure)
- [Execution Layer (.claude/)](#execution-layer-claude)
- [Data Layer (.sow/)](#data-layer-sow)
- [Git Versioning Strategy](#git-versioning-strategy)
- [File Naming Conventions](#file-naming-conventions)

---

## Complete Structure

```
project-root/
├── .claude/                              # EXECUTION LAYER
│   ├── .claude-plugin/
│   │   └── plugin.json                  # Plugin metadata (name, version, author)
│   │
│   ├── .plugin-version                  # Plugin version (for runtime access)
│   │
│   ├── agents/                          # Sub-agent definitions
│   │   ├── orchestrator.md              # Main coordinating agent
│   │   ├── architect.md                 # Design/architecture worker
│   │   ├── implementer.md               # Coding worker (TDD enforced)
│   │   ├── integration-tester.md        # Integration & E2E testing worker
│   │   ├── reviewer.md                  # Code review worker
│   │   └── documenter.md                # Documentation worker
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
│   │       ├── architect/
│   │       │   ├── create-adr.md        # Create Architecture Decision Record
│   │       │   └── design-doc.md        # Write design document
│   │       ├── implementer/
│   │       │   ├── implement-feature.md # Implement new feature with TDD
│   │       │   └── fix-bug.md           # Fix bug (test first, then fix)
│   │       ├── integration-tester/
│   │       │   └── write-integration-tests.md  # Integration & E2E tests
│   │       ├── reviewer/
│   │       │   └── review-code.md       # Review code for quality
│   │       └── documenter/
│   │           └── update-docs.md       # Update documentation
│   │
│   ├── hooks.json                       # Event automation (SessionStart, etc.)
│   │
│   ├── mcp.json                         # External tool integrations (optional)
│   │
│   └── migrations/                      # Migration instructions
│       ├── 0.1.0-to-0.2.0.md
│       └── 0.2.0-to-0.3.0.md
│
└── .sow/                                # DATA LAYER
    │
    ├── .version                         # Structure version tracking
    │
    ├── knowledge/                       # Repo-specific knowledge (COMMITTED)
    │   ├── overview.md                  # Project overview
    │   ├── architecture/                # Architecture documents
    │   │   ├── system-overview.md
    │   │   ├── api-design.md
    │   │   └── diagrams/
    │   │       └── architecture.png
    │   ├── adrs/                        # Architecture Decision Records
    │   │   ├── 001-use-postgres.md
    │   │   ├── 002-api-versioning.md
    │   │   └── 003-authentication-strategy.md
    │   └── deviations.md                # Project-specific deviations from sinks
    │
    ├── sinks/                           # External knowledge (GIT-IGNORED)
    │   ├── index.json                   # LLM-maintained sink catalog
    │   ├── python-style/                # Example: Python style guide
    │   │   ├── formatting.md
    │   │   ├── conventions.md
    │   │   └── testing.md
    │   ├── api-conventions/             # Example: API design standards
    │   │   ├── rest-api.md
    │   │   └── versioning.md
    │   ├── deployment-guide/            # Example: Deployment procedures
    │   │   ├── staging.md
    │   │   └── production.md
    │   └── security-checklist/          # Example: Security guidelines
    │       └── checklist.md
    │
    ├── repos/                           # Linked repositories (GIT-IGNORED)
    │   ├── index.json                   # Repository references
    │   ├── auth-service/                # Cloned or symlinked repo
    │   │   └── (full repository structure)
    │   └── shared-library/              # Cloned or symlinked repo
    │       └── (full repository structure)
    │
    └── project/                         # Active work (COMMITTED TO FEATURE BRANCHES)
        ├── state.yaml                   # Project state (phases, tasks, status, branch)
        ├── log.md                       # Append-only orchestrator action log
        │
        ├── context/                     # Project-specific context
        │   ├── overview.md              # Project overview
        │   ├── decisions.md             # Key decisions made during project
        │   └── memories.md              # Important context to remember
        │
        └── phases/                      # Phases contain tasks
            ├── discovery/               # Example: Discovery phase
            │   └── tasks/
            │       └── 010/
            │           ├── state.yaml         # Task state
            │           ├── log.md             # Worker action log
            │           ├── description.md     # Task description
            │           └── feedback/          # Human corrections
            │               ├── 001.md
            │               └── 002.md
            │
            ├── design/                  # Example: Design phase
            │   └── tasks/
            │       └── 010/
            │           └── (same structure)
            │
            ├── implement/               # Example: Implementation phase
            │   └── tasks/
            │       ├── 010/
            │       │   └── (same structure)
            │       ├── 020/
            │       │   └── (same structure)
            │       └── 030/
            │           └── (same structure)
            │
            ├── test/                    # Example: Testing phase
            │   └── tasks/
            │       └── 010/
            │           └── (same structure)
            │
            ├── review/                  # Example: Review phase
            │   └── tasks/
            │       └── 010/
            │           └── (same structure)
            │
            ├── deploy/                  # Example: Deploy phase
            │   └── tasks/
            │       └── 010/
            │           └── (same structure)
            │
            └── document/                # Example: Documentation phase
                └── tasks/
                    └── 010/
                        └── (same structure)
```

---

## Execution Layer (.claude/)

### Purpose

Defines how AI agents behave and interact with the system.

### How It Gets There

**IMPORTANT**: The `.claude/` directory is **created when the plugin is installed**, not stored in the plugin source repository.

**Plugin Development Structure** (in the `sow` repository):
```
sow/                                    # Marketplace repository
├── .claude-plugin/
│   └── marketplace.json                # Defines marketplace, points to plugin/
└── plugin/                             # Plugin source (template for .claude/)
    ├── .claude-plugin/
    │   └── plugin.json                 # Plugin metadata
    ├── .plugin-version                 # Version file
    ├── agents/                         # Agent definitions
    ├── commands/                       # Slash commands
    ├── hooks.json                      # Event automation
    └── mcp.json                        # MCP integrations
```

**Installation Flow**:
1. User runs: `/plugin install sow@sow-marketplace`
2. Claude Code reads `.claude-plugin/marketplace.json`
3. Finds plugin source at `./plugin`
4. **Copies contents of `plugin/` into user's repository as `.claude/`**
5. Result: User's repo now has `.claude/` with all plugin files

**Key Understanding**:
- `plugin/` directory = source template (lives in marketplace repo)
- `.claude/` directory = installed plugin (lives in user's repo after installation)
- When developing the plugin, edit files in `plugin/`
- When using the plugin, files are in `.claude/`

### Committed to Git

✅ **Yes** - Entire directory committed to all branches (after installation)
- Team shares same agent behaviors
- Distributed via Claude Code Plugin
- Version-controlled with codebase

### Components

#### `.claude-plugin/plugin.json`

Plugin metadata for distribution:
```json
{
  "name": "sow",
  "version": "0.2.0",
  "description": "AI-powered system of work for software engineering",
  "author": {
    "name": "sow contributors"
  },
  "repository": "https://github.com/your-org/sow"
}
```

#### `.plugin-version`

Simple version file for runtime access:
```
0.2.0
```

Read by orchestrator during SessionStart hook.

#### `agents/`

Sub-agent definitions (Markdown files with YAML frontmatter):

**Format**:
```markdown
---
name: architect
description: "System design and architecture specialist"
tools: Read, Write, Grep, Glob
model: inherit
---

You are a system architect...
(agent system prompt)
```

**Agents**:
- `orchestrator.md` - Main coordinator (user-facing)
- `architect.md` - Design and architecture
- `implementer.md` - Code implementation with TDD
- `integration-tester.md` - Integration and E2E testing
- `reviewer.md` - Code review and refactoring
- `documenter.md` - Documentation updates

#### `commands/workflows/`

User-invoked slash commands for high-level workflows:

- `/init` - Bootstrap sow in repository
- `/start-project <name>` - Create new project
- `/continue` - Resume existing project
- `/cleanup` - Delete project before merge
- `/migrate` - Migrate sow versions
- `/sync` - Sync sinks and repos

**Format**: Markdown files with optional frontmatter

#### `commands/skills/`

Agent-invoked capabilities, organized by agent type:

```
skills/
├── architect/
│   ├── create-adr.md
│   └── design-doc.md
├── implementer/
│   ├── implement-feature.md
│   └── fix-bug.md
├── integration-tester/
│   └── write-integration-tests.md
├── reviewer/
│   └── review-code.md
└── documenter/
    └── update-docs.md
```

Agents reference skills in their system prompts to keep prompts concise.

#### `hooks.json`

Event-driven automation configuration:

```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  }
}
```

**Events**: SessionStart, PreToolUse, PostToolUse, UserPromptSubmit, etc.

#### `mcp.json` (Optional)

Model Context Protocol server configurations for external integrations:

```json
{
  "servers": {
    "github": {
      "transport": "http",
      "url": "https://api.github.com/mcp"
    }
  }
}
```

#### `migrations/`

Version migration instructions (Markdown files):

**Naming**: `<from-version>-to-<to-version>.md`

**Example**: `0.1.0-to-0.2.0.md`

Contains automated steps and manual instructions for upgrading repository structure.

**See Also**: [DISTRIBUTION.md](./DISTRIBUTION.md#migration-system) for migration details

---

## Data Layer (.sow/)

### Purpose

Stores project knowledge, external context, and work state.

### Git Versioning

Mixed strategy (some committed, some ignored).

### Components

#### `.version`

Tracks repository structure version:

```yaml
sow_structure_version: 0.2.0
plugin_version: 0.2.0
last_migrated: 2025-10-12T16:30:00Z
initialized: 2025-10-12T14:00:00Z
```

**Committed**: ✅ Yes (all branches)

#### `knowledge/` (COMMITTED)

Repository-specific documentation and decisions.

**Committed**: ✅ Yes (all branches)

**Contents**:
- `overview.md` - High-level project description
- `architecture/` - Design documents, diagrams, system architecture
- `adrs/` - Architecture Decision Records (numbered chronologically)
- `deviations.md` - Project-specific deviations from installed sinks

**Use Case**: Persistent project knowledge that evolves with codebase

**Naming Convention**: ADRs use `###-description.md` format (e.g., `001-use-postgres.md`)

#### `sinks/` (GIT-IGNORED)

External knowledge installed per-developer.

**Committed**: ❌ No (git-ignored)

**Contents**:
- `index.json` - LLM-maintained catalog
  - Summaries of each sink
  - "When to use" guidance
  - Metadata (source, version, installed date)
- Sink directories - Collections of Markdown files on specific topics

**Examples**:
- `python-style/` - Python language conventions
- `api-conventions/` - API design standards
- `deployment-guide/` - Deployment procedures
- `security-checklist/` - Security guidelines

**Installation**:
```bash
sow sinks install git@github.com:company/guides.git python-style/
```

**Updates**:
```bash
sow sinks update  # Check for and apply updates
```

**Discovery**: Orchestrator reads `index.json` to route relevant sinks to workers

#### `repos/` (GIT-IGNORED)

External repositories needed for context.

**Committed**: ❌ No (git-ignored, too large)

**Contents**:
- `index.json` - Repository references
  - Git URLs
  - Purposes
  - Discovery hints
- Repository directories - Cloned or symlinked repositories

**Use Case**: Multi-repo projects where agents need cross-repo context

**Purpose**: Turn multi-repo setup into pseudo-monorepo for agent access

**Management**:
```bash
sow repos add git@github.com:company/auth-service.git
sow repos sync  # Pull latest changes
```

#### `project/` (COMMITTED TO FEATURE BRANCHES)

Active work coordination.

**Committed**: ✅ Yes (feature branches only, deleted before merge to main)

**Key Constraint**: Only one project per branch (enforced)

**Structure**:

##### `state.yaml`

Central project planning document:

```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  created_at: 2025-10-12T14:30:00Z
  complexity:
    rating: 2
  active_phase: implement

phases:
  - name: design
    status: completed
    tasks: [...]

  - name: implement
    status: in_progress
    tasks: [...]
```

**See Also**: [SCHEMAS.md](./SCHEMAS.md#project-stateyaml) for complete schema

##### `log.md`

Append-only orchestrator action log (structured Markdown):

```markdown
### 2025-10-12 15:30:42

**Agent**: orchestrator
**Action**: created_phase
**Result**: success

Created 'implement' phase with 3 initial tasks.

---
```

**Format**: Structured Markdown with required fields
**See Also**: [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md#logging) for format details

##### `context/`

Project-specific context:

- `overview.md` - Project overview
- `decisions.md` - Key decisions made during execution
- `memories.md` - Important context to remember across sessions

##### `phases/<phase-name>/`

Phase directories (discovery, design, implement, test, review, deploy, document)

**Structure**: Each phase contains `tasks/` subdirectory

##### `phases/<phase>/tasks/<id>/`

Individual task directories (numbered with gaps: 010, 020, 030)

**Contents**:

- `state.yaml` - Task metadata, iteration, references, feedback tracking
- `log.md` - Worker action log (CLI-generated)
- `description.md` - Task description with requirements and acceptance criteria
- `feedback/` - Chronologically numbered human corrections (001.md, 002.md, etc.)

**See Also**: [SCHEMAS.md](./SCHEMAS.md#task-stateyaml) for complete schemas

---

## Git Versioning Strategy

| Path | Committed? | Reason |
|------|-----------|--------|
| `.claude/` | ✅ Yes (all branches) | Team shares agents, commands, hooks (installed from plugin) |
| `.sow/.version` | ✅ Yes (all branches) | Track structure version |
| `.sow/knowledge/` | ✅ Yes (all branches) | Repo-specific docs, architecture, ADRs |
| `.sow/sinks/` | ❌ No (git-ignored) | External, per-developer installations |
| `.sow/repos/` | ❌ No (git-ignored) | External repositories, too large |
| `.sow/project/` | ✅ Yes (feature branches only) | Per-branch work state, deleted before merge to main |

### `.gitignore` Configuration

```gitignore
# sow - External content (per-developer)
.sow/sinks/
.sow/repos/

# Note: .sow/project/ is NOT ignored
# It is committed to feature branches and deleted before merge
```

### Why `.sow/project/` is Committed

**Problem**: Git-ignored files don't switch with branches
- Branch A creates project → git-ignored folder persists
- Switch to Branch B → same folder still there (wrong context!)

**Solution**: Commit to feature branches
- Git manages switching automatically
- Team can pull branch and see full project context
- Natural backup (pushed to remote)
- CI enforces cleanup before merge to main
- Squash merge keeps main history clean

**Workflow**:
1. Feature branch: `.sow/project/` committed normally
2. Push branch: Project state shared with team
3. Before merge: `/cleanup` deletes `.sow/project/`
4. CI check: Fails if `.sow/project/` exists in PR
5. Squash merge: Main branch stays clean

**See Also**: [ARCHITECTURE.md](./ARCHITECTURE.md#single-project-constraint) for rationale

---

## File Naming Conventions

### Tasks

**Gap Numbering**: 010, 020, 030, 040
- Allows insertions: 011, 012, 021
- No renumbering chaos
- Similar to database migrations

### ADRs

**Format**: `###-description.md`
- Examples: `001-use-postgres.md`, `002-api-versioning.md`
- Chronological numbering
- Descriptive slugs

### Feedback

**Format**: `###.md`
- Examples: `001.md`, `002.md`, `003.md`
- Chronological order
- Read in ascending order (001 → 002 → 003)

### Agents

**Format**: `<agent-name>.md`
- Examples: `orchestrator.md`, `architect.md`, `implementer.md`
- Lowercase, hyphen-separated
- Extension: `.md`

### Commands

**Format**: `<command-name>.md`
- Examples: `start-project.md`, `create-adr.md`
- Lowercase, hyphen-separated
- Can be organized in subdirectories

### Phases

**Format**: Lowercase, single word
- Available: `discovery`, `design`, `implement`, `test`, `review`, `deploy`, `document`
- Directory names match phase names exactly

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - Introduction to sow
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Design decisions and rationale
- **[SCHEMAS.md](./SCHEMAS.md)** - Complete file format specifications
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle details
- **[AGENTS.md](./AGENTS.md)** - Agent system documentation
