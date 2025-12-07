# Team Project Management for Sow

**Final Summary Document**
**Date**: 2025-12-07

---

## Executive Summary

Sow currently lacks visibility into projects across a team. While individual engineers can create and manage projects locally, there's no shared understanding of what work exists, who's working on what, or how individual tasks fit into broader initiatives.

This document proposes enhancements to sow that enable:

1. **Team-wide visibility** into active projects via a git-native registry
2. **Rich AI context** through work tracker hierarchy (epics, milestones, sibling issues)
3. **Multi-platform support** (GitHub, GitLab, Jira, Linear) without vendor lock-in
4. **Project chaining** where exploration/design projects feed into downstream work
5. **Graceful degradation** when work tracking is unavailable or unsupported

The design prioritizes simplicity, git-native persistence, and workflow orchestration over prescribing documentation structure.

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Registry storage | Git orphan branch (`sow-data`) | Git-native, no external dependencies |
| Code vs work tracking | Separate interfaces | Teams mix platforms (GitHub + Jira) |
| Work tracking | Optional | Core sow works without supported tracker |
| Knowledge directory | Eliminated | Sow orchestrates workflow, doesn't prescribe doc locations |
| Project outputs | Registered, chainable | Projects can feed into other projects |

---

## Design Philosophy: Integration Over Ownership

**Sow references data. It never owns it.**

This is a foundational principle: sow indexes and integrates with external systems, but is never the authoritative source for data that doesn't inherently belong to it.

| Data Type | Authoritative Source | Sow's Role |
|-----------|---------------------|------------|
| Issues, epics, milestones | Work tracker (Jira, GitHub Issues) | References by ID, queries at runtime |
| Repository documentation | The repository itself (`docs/`, `README.md`) | Reads on demand, never stores copies |
| External knowledge | Original sources (style guides, API docs) | Syncs to `sinks/`, treats as read-only cache |
| Code and branches | Git | Operates on, never duplicates |
| **Project state, tasks, logs** | **Sow** | **Owns and manages directly** |

What sow **does** own:
- Project lifecycle state (`.sow/project/state.yaml`)
- Task definitions and progress
- Agent session logs
- Project registry entries (references, not copies)

What sow **never** owns:
- Issue content (title, description, comments)
- Epic/milestone structures
- Repository documentation
- Code

**Why this matters:**

1. **No sync drift** - If sow stored copies of issue data, it would inevitably diverge from the source
2. **No stale caches** - Runtime queries always get current state
3. **Clear boundaries** - Teams know where to update what
4. **Simpler mental model** - Sow is a workflow orchestrator, not a data warehouse

This principle directly informed several design decisions:
- Registry stores references (`tracker: github, number: 123`), not issue content
- Context compilation queries work trackers at runtime
- Knowledge directory was eliminated (sow shouldn't store docs)
- Project chaining copies files once, doesn't maintain links

---

## Problem Statement

### The Visibility Gap

Projects in sow are stored in git worktrees at `.sow/worktrees/{branch}/.sow/project/`. This architecture has a fundamental limitation: **worktrees are local to each engineer**.

```
Engineer A's machine          Engineer B's machine
.sow/worktrees/               .sow/worktrees/
├── feat/auth/                ├── feat/api/
└── explore/caching/          └── fix/bug-123/
```

Each engineer only sees their own projects. There's no way to answer:
- "What projects are active across the team?"
- "Is anyone already working on issue #123?"
- "Which projects are stalled?"

### The Platform Coupling Problem

Currently, GitHub serves as the implicit coordination layer:
- Breakdown projects publish issues with the `sow` label
- Engineers pick up issues via the project wizard
- PRs show `.sow/project/` directories

But this coupling is loose and one-directional:
- Sow publishes to GitHub but doesn't track what it published
- There's no link from issue → sow project
- Non-GitHub platforms (GitLab, Jira, Linear) aren't supported

### The AI Context Gap

AI agents benefit from understanding where their work fits in a broader context:
- "This is issue 3 of 5 in the OAuth epic"
- "Issue 5 handles optimization, so don't pre-optimize here"
- "This epic is part of the Q4 Authentication milestone"

Currently, this context must be manually provided or doesn't exist at all.

### The Stale Documentation Problem

The existing `.sow/knowledge/` directory has issues:
- **Hidden**: People don't think to look there for useful information
- **Stale**: No process ensures docs stay current or get cleaned up
- **Duplicative**: Competes with conventional locations (`README.md`, `docs/`)

This exploration concludes that sow should not prescribe documentation locations.

---

## Solution Overview

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          sow CLI                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │ Project Registry │    │ Context Compiler │                   │
│  │ (sow-data branch)│    │ (runtime queries)│                   │
│  │                  │    │                  │                   │
│  │ • Project index  │    │ • Epic hierarchy │                   │
│  │ • Issue refs     │    │ • Milestones     │                   │
│  │ • State snapshots│    │ • Sibling items  │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                              │
│           │ [requires work        │ [requires work               │
│           │  tracking]            │  tracking]                   │
│           │                       │                              │
├───────────┴───────────────────────┴──────────────────────────────┤
│                                                                  │
│  ┌──────────────────────┐    ┌──────────────────────┐           │
│  │     Code Host        │    │    Work Tracker      │           │
│  │                      │    │    (optional)        │           │
│  │  • GitHub            │    │                      │           │
│  │  • GitLab            │    │  • GitHub Issues     │           │
│  │  • Bitbucket         │    │  • GitLab Issues     │           │
│  │                      │    │  • Jira              │           │
│  │  [always available]  │    │  • Linear            │           │
│  └──────────────────────┘    └──────────────────────┘           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Core Principles

1. **Store references, not copies** - Work tracker is source of truth; sow stores pointers
2. **Git-native persistence** - No external databases; everything lives in the repository
3. **Orchestrate, don't prescribe** - Sow manages workflow, not documentation structure
4. **Optional integrations** - Core functionality works without external platform support
5. **Minimal conflict surface** - One file per project minimizes merge conflicts

---

## Platform Abstraction

### Code Host vs Work Tracker

Sow distinguishes between two types of external platforms:

| Term | Examples | Capabilities |
|------|----------|--------------|
| **Code Host** | GitHub, GitLab, Bitbucket | Repositories, branches, pull requests |
| **Work Tracker** | GitHub Issues, GitLab Issues, Jira, Linear | Work items, epics, milestones |

GitHub and GitLab are both—they host code AND track work. Jira and Linear are work trackers only. A team might use GitHub for code but Jira for work tracking.

### Configuration

```yaml
# .sow/config.yaml (per-repository)
code_host:
  type: github      # Auto-detected from git remote

work_tracker:       # Optional - null disables work tracking
  type: jira
  url: https://company.atlassian.net
  project: PROJ
```

**Common configurations:**

| Setup | Code Host | Work Tracker |
|-------|-----------|--------------|
| GitHub-only | github | github |
| GitHub + Jira | github | jira |
| GitLab-only | gitlab | gitlab |
| GitHub, no tracker | github | null |

### Initialization Flow

```
$ sow init

Detected code host: GitHub (from git remote origin)

Select your work tracker:
  1. GitHub Issues
  2. GitLab Issues
  3. Jira
  4. Linear
  5. Not listed (disables work tracking)

> 5

Work tracking disabled. The following features will be unavailable:
  - Team project registry (sow-data branch)
  - Issue → project linking
  - AI context from epics/milestones
  - Breakdown publishing to tracker

Core functionality remains available:
  - Project creation and state management
  - Agent spawning and coordination
  - PR creation

Continue? [Y/n]
```

### Feature Availability

| Feature | Work Tracking Enabled | Work Tracking Disabled |
|---------|----------------------|------------------------|
| `sow project new` (from branch) | Yes | Yes |
| `sow project new --issue <N>` | Yes | No |
| `sow project list --all` | Yes | No |
| Project state machine | Yes | Yes |
| Agent spawning | Yes | Yes |
| PR creation | Yes | Yes |
| Breakdown → publish issues | Yes | No |
| AI context (epic, siblings) | Yes | No |
| Project chaining | Yes | Yes |

### Interface Definitions

```go
// CodeHost handles git-related operations (always available)
type CodeHost interface {
    Name() string
    CheckAvailability() error

    // Pull requests
    CreatePullRequest(title, body string, draft bool) (number int, url string, err error)
    UpdatePullRequest(number int, title, body string) error
    MarkPullRequestReady(number int) error

    // Branch linking (when code host is also work tracker)
    GetLinkedBranches(workItemNumber int) ([]LinkedBranch, error)
    CreateLinkedBranch(workItemNumber int, branchName string) (string, error)
}

// WorkTracker handles issue/epic/milestone operations (optional)
type WorkTracker interface {
    Name() string
    CheckAvailability() error

    // Work items
    GetWorkItem(number int) (*WorkItem, error)
    ListWorkItems(opts ListWorkItemsOpts) ([]WorkItem, error)
    CreateWorkItem(spec WorkItemSpec) (*WorkItem, error)

    // Labels
    AddLabel(workItemNumber int, label string) error
    RemoveLabel(workItemNumber int, label string) error
}

// WorkTrackerWithHierarchy extends WorkTracker with epic/milestone support
type WorkTrackerWithHierarchy interface {
    WorkTracker

    // Epic operations
    GetEpic(ref Ref) (*Epic, error)
    GetEpicWorkItems(ref Ref) ([]WorkItem, error)
    CreateEpic(spec EpicSpec) (*Epic, error)

    // Milestone operations
    GetMilestone(ref Ref) (*Milestone, error)
    GetMilestoneEpics(ref Ref) ([]Epic, error)

    // Hierarchy traversal
    GetWorkItemEpic(workItemNumber int) (*Epic, error)
    GetWorkItemMilestone(workItemNumber int) (*Milestone, error)
}
```

### Core Types

```go
// Ref identifies an entity in a work tracker
type Ref struct {
    Tracker string // "github", "gitlab", "jira", "linear"
    ID      string // Tracker-specific identifier
}

type WorkItem struct {
    Ref          Ref
    Number       int
    Title        string
    Body         string
    State        string   // Normalized: "open", "closed", "in_progress"
    URL          string
    Labels       []string
    EpicRef      *Ref
    MilestoneRef *Ref
    DependsOn    []int
}

type Epic struct {
    Ref          Ref
    Name         string
    Description  string
    State        string
    URL          string
    MilestoneRef *Ref
}

type Milestone struct {
    Ref         Ref
    Name        string
    Description string
    State       string
    DueDate     *time.Time
    URL         string
}
```

### Terminology Mapping

| Sow Term | GitHub | GitLab | Jira | Linear |
|----------|--------|--------|------|--------|
| **WorkItem** | Issue | Issue | Story/Task/Bug | Issue |
| **Epic** | Project v2 | Epic | Epic | Project |
| **Milestone** | Milestone | Milestone | Fix Version | Cycle |

### Authentication

Token-based authentication stored per-user:

```yaml
# ~/.config/sow/auth.yaml
code_hosts:
  github:
    token: ghp_xxxxxxxxxxxx
  gitlab:
    token: glpat-xxxxxxxxxxxx

work_trackers:
  github:
    token: ghp_xxxxxxxxxxxx    # Same token if using GitHub Issues
  jira:
    email: user@company.com
    token: xxxxxxxxxxxx
  linear:
    token: lin_xxxxxxxxxxxx
```

Resolution order:
1. Environment variable (`GITHUB_TOKEN`, `JIRA_TOKEN`, etc.)
2. Config file (`~/.config/sow/auth.yaml`)

---

## Project Registry

### Purpose

The project registry provides team-wide visibility into active projects. It answers:
- What projects exist across the team?
- Is anyone working on issue #123?
- Which projects are stalled?

**Note**: The registry is only created when work tracking is enabled.

### Storage: Orphan Branch

An orphan branch named `sow-data` serves as the shared registry. Like GitHub Pages (`gh-pages`), this branch:
- Has no common history with main
- Is dedicated to sow metadata
- Can be pushed/pulled independently

### Directory Structure

```
sow-data/
├── projects/
│   ├── feat--auth-implementation.yaml
│   ├── explore--api-design.yaml
│   └── breakdown--q4-roadmap.yaml
├── refs/
│   └── issues/
│       ├── 123.yaml
│       └── 124.yaml
└── meta.yaml
```

### Project Entry Schema

```yaml
# projects/feat--auth-implementation.yaml

# Identity
branch: feat/auth-implementation
name: Auth Implementation
type: standard                    # standard | exploration | design | breakdown
description: Implement OAuth2 authentication with JWT tokens

# Lifecycle
state: PlanningReview             # Current statechart state
created_at: 2025-12-01T10:00:00Z
updated_at: 2025-12-07T14:30:00Z
owner: alice

# Source
source:
  type: issue                     # issue | branch | project
  tracker: github                 # If type=issue
  issue_number: 123               # If type=issue
  project_branch: explore/auth    # If type=project

# Work tracker references (pointers, not copies)
refs:
  milestone:
    tracker: github
    id: "5"
  epics:
    - tracker: github
      id: proj_abc123
  issues:
    - tracker: github
      number: 123

# Registered outputs (for project chaining)
outputs:
  - path: phases/exploration/tasks/040/findings.md
    label: summary
  - path: phases/exploration/tasks/030/findings.md
    label: interface-design
```

### Reverse Reference Schema

```yaml
# refs/issues/123.yaml
project_branch: feat/auth-implementation
created_at: 2025-12-07T10:00:00Z
```

### Conflict Handling

The one-file-per-project structure minimizes conflicts:
- Engineer A creates `projects/feat--auth.yaml`
- Engineer B creates `projects/feat--api.yaml`
- No conflict: different files

Sync protocol for writes:
```
1. git fetch origin sow-data
2. git checkout sow-data
3. Make changes
4. git commit
5. git push (retry with rebase on conflict)
```

---

## Project Chaining

### Concept

Projects can serve as inputs to other projects, creating a workflow chain:

```
┌─────────────────────┐
│ Exploration Project │ ← registers outputs, stays open
└─────────┬───────────┘
          │ referenced as input
          ▼
┌─────────────────────┐
│   Design Project    │ ← copies outputs to context/
└─────────┬───────────┘
          │ referenced as input
          ▼
┌─────────────────────┐
│  Breakdown Project  │ ← publishes to work tracker
└─────────────────────┘
          │
          ▼
    Multiple implementation projects
```

### Registered Outputs

Projects declare their final artifacts:

```yaml
# .sow/project/state.yaml
outputs:
  - path: phases/exploration/tasks/050/findings.md
    label: summary
  - path: phases/exploration/tasks/030/findings.md
    label: interface-design
```

### Project Creation Wizard

When creating a new project, the wizard offers input sources:

```
Select input sources:
  1. Files from this repository
  2. External references (sinks)
  3. Another project

> 3

Available projects with outputs:
  explore/project-management  [exploration, 2 outputs]

Select outputs to include:
  [x] summary: findings.md
  [x] interface-design: interface-design.md

✓ Copied to .sow/project/context/
```

### Workflow Benefits

1. **No forced commits** - Intermediate artifacts don't need permanent homes
2. **Chain cleanup** - Close multiple projects together when done
3. **Clear lineage** - Registry tracks source → derived relationships
4. **Flexibility** - User can still manually copy files if preferred

### End-of-Project Flow

When a project completes, the orchestrator asks:

```
This exploration produced these artifacts:
  - findings.md (consolidated summary)
  - interface-design.md

What would you like to do?
  1. Create a follow-up project with these as context
  2. Move to a location in this repo (you specify where)
  3. Create a work tracker issue from the summary
  4. Leave in project (will be deleted when branch is deleted)
  5. Something else
```

---

## Knowledge Management (Eliminated)

### Decision

Sow no longer maintains a `.sow/knowledge/` directory.

### Rationale

The knowledge directory had structural problems:
- **Hidden**: Located where people don't think to look
- **Stale**: No process to update or clean up
- **Duplicative**: Competes with `README.md`, `docs/`, inline comments

### New Model

```
Repository documentation → conventional locations (README, docs/, comments)
External knowledge       → sinks/ (synced from authoritative sources)
Project context          → .sow/project/context/ (ephemeral, per-project)
Project outputs          → registered, chainable to downstream projects
```

Sow orchestrates workflow. It doesn't prescribe where documentation lives.

### Migration

For existing repositories with `.sow/knowledge/`:
- Contents can be moved to `docs/` or wherever the team prefers
- No automated migration needed
- Sow ignores the directory if it exists

---

## Context Compilation

### Purpose

When an AI agent starts work, it needs to understand where its task fits. Context compilation queries the work tracker at runtime to build this understanding.

**Note**: Requires work tracking to be enabled.

### Flow

```
1. Load project from registry
   → Project refs issue #123

2. Query work tracker for work item
   → GET /issues/123
   → Returns: title, body, epic, milestone

3. Query work tracker for epic context
   → Get sibling work items in same epic
   → Understand what comes before/after

4. Query work tracker for milestone context
   → Understand broader timeline/release

5. Render context for AI prompt
```

### Output Format

```markdown
## Project Context

**Working on**: #123 - Add OAuth2 login flow

**Epic**: OAuth2 Implementation (3 of 5 issues completed)
  - [x] #121: Set up OAuth2 provider configuration
  - [x] #122: Implement token exchange endpoint
  - [ ] #123: Add OAuth2 login flow ← YOU ARE HERE
  - [ ] #124: Add token refresh handling
  - [ ] #125: Add logout and token revocation

**Milestone**: Q4 Authentication Overhaul (due: 2025-12-31)

**Note**: Issue #124 depends on this work completing.
Issue #125 will handle token cleanup - do not implement revocation here.
```

### Why Runtime Queries?

We query the work tracker at runtime rather than caching because:
- Work tracker is source of truth for PM data
- PMs may add/remove/reorder issues
- Caching would drift and require sync logic
- Query latency is acceptable (once at project start)

---

## User Workflows

### Tech Lead: View All Projects

```bash
$ sow project list --all

BRANCH                    STATE           OWNER    UPDATED
feat/auth-implementation  PlanningReview  alice    2h ago
explore/api-design        Active          bob      1d ago
breakdown/q4-roadmap      Completed       carol    3d ago
```

### Engineer: Check if Issue is Claimed

```bash
$ sow project status --issue 123

Issue #123 is owned by project: feat/auth-implementation
  Owner: alice
  State: PlanningReview
  Branch: feat/auth-implementation
```

### Engineer: Start Work on Issue

```bash
$ sow project new --issue 456

Creating project from issue #456: Implement caching layer
  Branch: feat/implement-caching-layer-456
  Type: standard

✓ Project registered in sow-data
✓ Worktree created
✓ Claude Code launching...
```

### Engineer: Create Project from Exploration

```bash
$ sow project new

Select project type: design

Select input sources:
  1. Files from this repository
  2. External references (sinks)
  3. Another project

> 3

Available projects with outputs:
  explore/project-management  [exploration, 2 outputs]

Copying outputs to context...
✓ Created design project on branch design/team-pm
```

### Breakdown: Publish to Work Tracker

```bash
$ sow advance  # In publishing state

Publishing work units to Jira (PROJ)...
  ✓ Created epic: Q1 API Redesign
  ✓ Created issue PROJ-201: Design new endpoint structure
  ✓ Created issue PROJ-202: Implement v2 endpoints
  ✓ Created issue PROJ-203: Migration guide

✓ Project registry updated with work tracker references
```

---

## Implementation Phases

### Phase 1: Foundation

**Configuration & Init**
- Add work tracker selection to `sow init`
- Create `.sow/config.yaml` schema
- Implement optional work tracking flag
- Remove `.sow/knowledge/` from project templates

**Registry Infrastructure**
- Create `sow-data` orphan branch on init (when work tracking enabled)
- Implement read/write operations with conflict handling
- Add `sow project list --all` command

**Platform Abstraction (GitHub only)**
- Implement `CodeHost` interface for GitHub
- Implement `WorkTracker` interface for GitHub Issues
- Token-based auth via env var / config file
- Migrate existing `GitHubCLI` to new interfaces

### Phase 2: Project Chaining

**Registered Outputs**
- Add `outputs` field to project state schema
- Orchestrator prompts to register outputs at project end
- Store output references in registry

**Project as Input**
- Wizard option to select project as input source
- Copy registered outputs to new project's context/
- Track source project in registry entry

**End-of-Project Flow**
- Orchestrator asks what to do with artifacts
- Options: chain, move, issue, leave
- Support closing multiple chained projects together

### Phase 3: Hierarchy & Context

**Hierarchy Support**
- Implement `WorkTrackerWithHierarchy` for GitHub (requires GraphQL)
- Add context compilation logic
- Integrate context into orchestrator prompts

**Registry Integration**
- Register projects on creation
- Add reverse refs for issues
- Check for existing projects before creation

### Phase 4: Multi-Platform

**Additional Platforms**
- GitLab implementation (`CodeHost` + `WorkTracker`)
- Jira implementation (`WorkTracker` only)
- Linear implementation (`WorkTracker` only)

**CLI Enhancements**
- `sow auth status` for verification
- `sow config show` to display configuration

---

## Trade-offs and Decisions

### Why an Orphan Branch?

**Alternatives considered:**
- Index file in main branch → merge conflicts on every project change
- External database → adds infrastructure dependency
- Platform API only → vendor lock-in, no offline capability

**Orphan branch wins because:**
- Git-native (works with any remote)
- No merge conflicts with main
- Familiar pattern (gh-pages)
- Works offline (with eventual sync)

### Why Separate Code Host from Work Tracker?

**Alternative:** Single "forge" concept

**We chose separation because:**
- Some platforms are code-only (Bitbucket) or work-only (Jira, Linear)
- Teams commonly use mixed setups (GitHub + Jira)
- Clearer mental model

### Why API-Only (No CLI Fallback)?

**Alternative:** Support both CLI (`gh`, `glab`) and API

**We chose API-only because:**
- Hybrid doubles implementation count (6 vs 3 per operation)
- Ongoing maintenance burden for two code paths
- Token-only auth is simple enough

### Why Eliminate the Knowledge Directory?

**Alternative:** Keep `.sow/knowledge/` as prescribed location

**We chose elimination because:**
- Hidden location reduces discoverability
- No process ensures freshness
- Competes with conventional doc locations
- Sow should orchestrate workflow, not prescribe structure

### Why Project Chaining?

**Alternative:** Force intermediate commits to docs/

**We chose chaining because:**
- Not all artifacts deserve permanent homes
- Exploration docs may only matter for the next project
- User retains control over what gets committed where
- Simpler cleanup when chains complete

### Why Make Work Tracking Optional?

**Alternative:** Require a supported work tracker

**We chose optional because:**
- Many teams use unsupported trackers (Notion, Monday.com, internal tools)
- Core sow functionality doesn't require work tracking
- Better to work partially than not at all

---

## Open Questions

1. **Project archival** - When a project completes:
   - Delete registry entry?
   - Move to `projects/archived/`?
   - Add `archived: true` field?

2. **Stale project detection** - How to identify abandoned projects?
   - Based on `updated_at` threshold?
   - Based on branch activity?
   - Manual marking?

3. **Chain cleanup** - When closing a project that's a source for others:
   - Warn and require confirmation?
   - Automatically close downstream?
   - Orphan the downstream projects?

4. **Output registration UX** - When should orchestrator prompt for outputs?
   - At explicit project close?
   - When transitioning to certain states?
   - Manual `sow project register-output` command?

5. **Permissions** - Who can update the registry?
   - Anyone with push access?
   - Project owner only?
   - No enforcement (trust model)?

---

## Success Criteria

The implementation is successful when:

1. **Visibility**: Tech leads can see all active projects with a single command
2. **Coordination**: Engineers can check if an issue is already being worked on
3. **Context**: AI agents receive hierarchy context (epic, siblings, milestone)
4. **Chaining**: Exploration → design → breakdown flows work smoothly
5. **Multi-platform**: At least two work trackers supported (GitHub Issues + one other)
6. **Graceful degradation**: Core sow works without work tracking configured

---

## Appendix: File Locations

| File | Purpose |
|------|---------|
| `.sow/config.yaml` | Repository configuration (code host, work tracker) |
| `~/.config/sow/auth.yaml` | Authentication tokens |
| `sow-data:projects/*.yaml` | Project registry entries |
| `sow-data:refs/issues/*.yaml` | Issue → project reverse lookup |
| `sow-data:meta.yaml` | Registry schema version |
| `.sow/project/state.yaml` | Active project state |
| `.sow/project/context/` | Project-specific context files |

---

## Appendix: Command Reference

```bash
# Configuration
sow init                            # Initialize sow, select work tracker
sow config show                     # Show current configuration

# Registry operations (requires work tracking)
sow project list --all              # List all projects across team
sow project status --issue <N>      # Check if issue has project
sow project register                # Register current project

# Project chaining
sow project new                     # Create project (can select inputs)
sow project register-output <path>  # Register file as project output
sow project close                   # Close project (prompts for artifacts)

# Auth operations
sow auth status                     # Show auth status
sow auth set github <token>         # Store GitHub token
sow auth set jira <email> <token>   # Store Jira credentials

# Context operations (internal)
sow context compile --issue <N>     # Generate AI context for issue
```

---

## Appendix: Terminology Reference

| Term | Definition |
|------|------------|
| **Code Host** | Platform hosting git repositories (GitHub, GitLab, Bitbucket) |
| **Work Tracker** | Platform tracking issues/epics/milestones (GitHub Issues, Jira, Linear) |
| **Project Registry** | Shared index of projects in `sow-data` branch |
| **Registered Output** | File declared as a project's final artifact |
| **Project Chaining** | Using one project's outputs as another's inputs |
| **Context Compilation** | Building AI context from work tracker hierarchy |
