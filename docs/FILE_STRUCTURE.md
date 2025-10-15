# sow File Structure

**Last Updated**: 2025-10-15
**Purpose**: Complete directory layout and organization

This document provides a comprehensive reference for the `sow` file structure, explaining what each directory contains, when it's created, and how it's versioned in git.

---

## Table of Contents

- [Structure Overview](#structure-overview)
- [Execution Layer (.claude/)](#execution-layer-claude)
- [Data Layer (.sow/)](#data-layer-sow)
- [Git Versioning Strategy](#git-versioning-strategy)
- [Conditional Directory Creation](#conditional-directory-creation)
- [Plugin Installation Flow](#plugin-installation-flow)
- [Related Documentation](#related-documentation)

---

## Structure Overview

Complete directory structure:

```
repository-root/
├── .claude/                          # EXECUTION LAYER (committed)
│   ├── agents/
│   │   ├── orchestrator.md
│   │   ├── architect.md
│   │   ├── implementer.md
│   │   ├── researcher.md
│   │   ├── planner.md
│   │   ├── reviewer.md
│   │   └── documenter.md
│   ├── commands/
│   │   ├── project/
│   │   │   ├── new.md
│   │   │   └── continue.md
│   │   └── phases/
│   │       ├── discovery.md
│   │       ├── design.md
│   │       ├── implementation.md
│   │       ├── review.md
│   │       └── finalize.md
│   ├── hooks/                        # Optional
│   ├── integrations/                 # Optional
│   └── CLAUDE.md
│
├── .sow/                             # DATA LAYER
│   ├── knowledge/                    # COMMITTED
│   │   ├── overview.md
│   │   ├── architecture/
│   │   └── adrs/
│   │
│   ├── refs/                         # GIT-IGNORED (except indexes)
│   │   ├── .gitignore                # Committed
│   │   ├── index.json                # Committed
│   │   ├── index.local.json          # Git-ignored
│   │   └── <link-name>/              # Git-ignored symlinks
│   │
│   └── project/                      # COMMITTED TO FEATURE BRANCH
│       ├── state.yaml
│       ├── log.md
│       ├── context/
│       │   ├── decisions.md
│       │   └── memories.md
│       └── phases/
│           ├── discovery/            # If enabled
│           │   ├── log.md
│           │   ├── notes.md
│           │   └── research/
│           ├── design/               # If enabled
│           │   ├── log.md
│           │   ├── adrs/
│           │   └── design-docs/
│           ├── implementation/       # Always created
│           │   ├── log.md
│           │   └── tasks/
│           │       └── <id>/
│           │           ├── state.yaml
│           │           ├── description.md
│           │           ├── log.md
│           │           └── feedback/
│           ├── review/               # Always created
│           │   ├── log.md
│           │   └── reports/
│           └── finalize/             # Always created
│               └── log.md
│
└── .gitignore

~/.cache/sow/                         # USER-LEVEL CACHE
├── index.json
└── repos/
    ├── github.com/
    └── gitlab.com/
```

**Execution Layer** (`.claude/`): Agent definitions, commands, optional hooks/integrations, project instructions. Fully committed.

**Data Layer** (`.sow/`): Repository docs (knowledge/), external references (refs/), active project state (project/, feature branches only).

**User Cache**: External repositories cloned once, shared across all local projects.

---

## Execution Layer (.claude/)

### Purpose

Contains AI agent behavior, commands, hooks, and integrations. This layer defines **how agents should behave**.

### Source

- **Development**: Files created in `plugin/` directory of sow marketplace repository
- **Distribution**: Copied to user repositories via Claude Code Plugin installation
- **Installation**: Contents of `plugin/` → `.claude/` when user runs `/plugin install sow`

### Contents

#### agents/

Agent definition files (Markdown with YAML frontmatter): orchestrator.md (main coordinator), architect.md (design and ADRs), implementer.md (TDD implementation), researcher.md (discovery research), planner.md (task breakdown), reviewer.md (quality validation), documenter.md (documentation updates).

#### commands/

Project commands (new.md, continue.md) and phase commands (discovery variants for bug/feature/refactor/general, design, implementation, review, finalize).

#### hooks/, integrations/, CLAUDE.md

Optional hooks/ for automation scripts. Optional integrations/ for MCP server configs. CLAUDE.md contains project instructions loaded into every agent's context.

**See Also**: [AGENTS.md](./AGENTS.md) for agent system details

---

## Data Layer (.sow/)

### Purpose

Contains project knowledge, external references, and active work state. This layer defines **what the project knows**.

### Versioning Strategy

Mixed approach:
- **Committed**: `knowledge/`, `project/` (on feature branches), `refs/index.json`
- **Git-ignored**: `refs/*` (except indexes)

### Contents

#### knowledge/ (COMMITTED)

Repository-specific documentation: overview, architecture docs, and ADRs. Living documentation alongside code. ADRs from design phase moved here on finalize. Referenced by agents during work. Committed to all branches.

#### refs/ (GIT-IGNORED with exceptions)

External references from other repositories. Contains .gitignore (committed), index.json (committed team metadata), index.local.json (git-ignored local refs), and symlinks to cache (git-ignored). After fresh clone, run `sow refs init` to create symlinks from committed index.

**See Also**: [REFS.md](./REFS.md) for complete refs system

#### project/ (COMMITTED TO FEATURE BRANCH)

Active project state committed to feature branches only, never main. Contains state.yaml (project state), log.md (orchestrator log), context/ (decisions and memories), and phases/ (conditional directories).

**Phase Directories** (conditional based on enablement):
- **discovery/**: log, notes, decisions, research/ (if enabled)
- **design/**: log, notes, requirements, adrs/, design-docs/, diagrams/ (if enabled)
- **implementation/**: log, implementation-plan, tasks/<id>/ with state.yaml, description.md, log.md, feedback/ (always created)
- **review/**: log, reports/ (always created)
- **finalize/**: log (always created)

Deleted before merge to main.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md) for lifecycle details

---

## Git Versioning Strategy

**All Branches**: .claude/, .sow/knowledge/, .sow/refs/index.json, .sow/refs/.gitignore

**Feature Branches Only**: .sow/project/ (state, logs, context, phases)

**Never Committed**: .sow/refs/<link-name>/ (symlinks), .sow/refs/index.local.json

**Cleanup**: Finalize phase deletes .sow/project/, commits deletion, creates PR. CI validates no project folder exists before merge. Squash merge to main for clean history.

**See Also**: [PHASES/FINALIZE.md](./PHASES/FINALIZE.md) for finalize workflow

---

## Conditional Directory Creation

Phase directories created based on enablement. Always created: state.yaml, log.md, context/, phases/implementation/, phases/review/, phases/finalize/. Conditionally created: phases/discovery/ and phases/design/ (only if enabled). Truth table during `/project:new` determines phase enablement.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md#truth-table) for phase determination

---

## Plugin Installation Flow

`/plugin install sow` fetches plugin and copies to `.claude/`. `sow init` creates `.sow/` structure with knowledge/ and refs/ directories. `.sow/project/` created only on `/project:new`.

**Upgrading**: `/plugin upgrade sow` replaces `.claude/` (data layer unchanged). CLI upgrade via package manager with automatic migrations if needed.

**See Also**: [OVERVIEW.md](./OVERVIEW.md#installation) for installation guide

---

## Related Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Two-layer architecture explanation
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - When directories are created
- **[REFS.md](./REFS.md)** - External references system details
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - State file contents
- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications
