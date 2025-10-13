# Brainstorming Session - 2025-10-12

## Session Goal
Explore the problem space for `sow` - an AI-powered system of work for software engineering. Focus on keeping options open, challenging assumptions, and documenting the discovery process.

---

## Topic: File Structure & Architecture

### Two-Layer Architecture Concept

**Key insight**: Distinction between behavior and state
- **AI Software Layer**: Agents, slash commands, hooks - shipped via Claude Code Plugin
- **Data/State Layer**: Project structure, tasks, memories, standards, coding conventions - lives in repository

**Rationale**:
- Plugins are good for distributing behavior/logic
- Project state needs to version with the code (git)
- Different projects need different states but can share the same AI tooling
- Cannot ship project-specific state through a plugin

### Bootstrap/Installation Model

**Hypothetical flow**:
1. User installs Claude Code Plugin globally (provides AI capabilities)
2. Plugin includes slash command to "install sow" into current repo
3. Installation step could:
   - Install CLI locally
   - Fetch static files from GitHub
   - Create initial folder structure in repo
   - Set up configuration templates

**Analogy**: Like `npm init`, `git init` - tools that bootstrap project structure

### Resolved Design Decisions

**Coupling**:
- ✓ Tight coupling is intentional - this is an opinionated system
- Missing/malformed structures should prompt user with guidance on resolution
- The plugin assumes the data structure exists and guides setup if it doesn't

**Versioning**:
- ✓ Clear separation: execution layer (immutable) vs. data storage layer (mutable)
- Execution layer can be replaced wholesale when upgrading
- Data storage layer evolves and persists
- Migration guides for structural changes, AI-assisted via `/migrate` command

**Customization**:
- ✓ Provide general structure for data storage, but don't allow folder structure mutations
- Both agents and humans contribute to the data layer
- Teams can add data within the existing structure, but shouldn't change the structure itself
- "Pit of success" approach - guide users into the standard pattern

**Cold start experience**:
- ✓ Design principle: agents must always be able to orient themselves
- Initial prompt + sow structure should be self-documenting
- Starting an agent in an existing sow repo should be seamless
- No special "setup mode" vs "normal mode"

**Installation UX**:
- ✓ Single CLI with `init` command
- One step: install Claude Code Plugin + setup repository structure
- Consolidates bootstrapping experience

### Open Questions

**Multi-repo scenarios** [EXPLORING NOW]:
- Do we install into each repository separately?
- Is there a workspace/monorepo concept?
- How does state sharing work across related repos?

---

## Deep Dive: Multi-Repo & Cross-Repository Knowledge

### The Problem
Agents working in one repository often need context from other repositories or shared knowledge:
- Microservices with interdependencies
- Shared team standards (API conventions, coding styles)
- Cross-cutting concerns (deployment, release processes)
- Company policies that apply to all projects

### Approach 1: Central Knowledge Repository

**Concept**: Single source of truth for cross-repo knowledge

**Mechanics**:
- Lives at well-known location: `~/.local/state/sow/...`
- Available to agents in any sow-powered repo
- Requires an index for agent discovery

**Pros**: Single location, easy to update centrally
**Cons**: All-or-nothing, no project-specific filtering

### Approach 2: Embedded Repository Links

**Concept**: Project selectively pulls in other repos

**Mechanics**:
- Git-ignored directory: `./.sow/repos/...`
- Repositories cloned or symlinked
- Index maintained for agent discovery
- Slash commands to manage (add/remove)

**Pros**: Selective, project-specific, flexible
**Cons**: Management overhead, potential duplication

### Approach 3: Information Sinks [LEADING CANDIDATE]

**Concept**: Granular, composable knowledge units that can be installed into projects

**Definition**:
- "Sink" = collection of one or more Markdown files
- Provides focused context on specific topics
- Can live anywhere (central repo, other repos, standalone)

**Mechanics**:
- Format: git repo link + folder path to copy from
- User-level index (in home folder, can be git-backed)
- Can install "sink indexes" that reference multiple sinks
- Installed into project at git-ignored path
- Local index maintained for agent discovery

**Example Sinks**:
- Language style guide (e.g., "Python conventions")
- Deployment process
- Release process
- Company code policy
- API design standards
- Security guidelines

**Workflow**:
1. User installs sink index (e.g., company-wide standards)
2. User selectively installs specific sinks into project
3. Sinks embedded in `./.sow/sinks/...` (git-ignored)
4. Index updated for agent discovery
5. Agent reads relevant sinks when working

**Advantages**:
- Granular: only install what you need
- Flexible: sinks can live anywhere
- Composable: mix and match from different sources
- Versioned: sinks live in git
- Discoverable: index helps agents find relevant context

### Addressing Challenges

**Context Overload**:
- Multi-agent system: Orchestrator + Workers
- Orchestrator reads index (not full content)
- Workers receive specific file paths from orchestrator
- Example: "You're working on Go, adhere to `./.sow/sinks/go/style_guide.md`"
- Selective loading prevents context pollution

**Update Propagation**:
- Sinks are git-versioned
- Index pins sinks by git reference (commits/tags)
- Manual updates via CLI: `sow sinks update`
- Similar to package manager workflow (npm, pip, etc.)
- Check updates by polling upstream remotes

**Sink Discovery by Agents**:
- Index file at root of sinks folder (`./.sow/sinks/index.json`)
- Index is LLM-maintained (not hand-written)
- When sink installed, LLM interrogates and summarizes
- Index includes "when should you use this?" guidance
- Orchestrator uses index to route context to workers

**Sink Granularity**:
- No hard limits
- Operational choice by user
- Large sinks naturally degrade experience (incentive to split)
- Index/orchestrator governs usage

**Local Overrides**:
- Structure supports both "remote" and "local" data
- Local data is more authoritative
- Allows project-specific deviations from company standards

**Monorepo Strategy**:
- Hybrid approach:
  - `.sow/sinks/` - ALWAYS present (information sinks)
  - `.sow/repos/` - OPTIONAL (linked repositories)
- Multi-repo: use `repos/` to checkout/symlink related repos (pseudo-monorepo)
- True monorepo: never uses `repos/`, just uses standard project structure

---

## Deep Dive: Multi-Agent Architecture

### Orchestrator + Worker Pattern

**Underlying Assumption**: `sow` operates as a multi-agent system, not a single agent

**Orchestrator Agent**:
- High-level problem solver
- Does NOT write code or perform "real" work
- Responsibilities:
  - Parse user intent
  - Load and interrogate indexes
  - Decompose work into tasks
  - Assign tasks to worker agents with specific context
  - Coordinate between workers

**Worker Agents**:
- Task-specific executors
- Receive bounded context from orchestrator
- Responsibilities:
  - Execute specific tasks (write code, run tests, etc.)
  - Only load context relevant to their task
  - Report results back to orchestrator

**Information Flow**:
```
User Request
  ↓
Orchestrator (reads .sow/sinks/index.json)
  ↓
Worker 1 (receives: "use .sow/sinks/python/style.md")
Worker 2 (receives: "use .sow/sinks/api/conventions.md")
  ↓
Orchestrator (synthesizes results)
  ↓
User Response
```

### Architecture Details

**Implementation**:
- Primary: Claude Code sub-agents (`.claude/agents/*.yaml`)
- Future: Custom layer supporting other model providers (e.g., Cursor CLI)
- MVP: Pure Claude sub-agents

**User Experience** (Hybrid Model):
- **90% of cases**: Orchestrator manages everything transparently
- **Orchestrator = Active Claude Code Session**: User sits in front of it, sees everything
- **Translation Layer**: User ↔ Orchestrator ↔ Workers
- **Error Correction**: User explains mistake to orchestrator, which reformulates for worker
- **Escape Hatch**: User can invoke workers directly if orchestrator struggles
- **Visibility**: Normal Claude Code interruption (ESC key) works
- **Post-facto Debugging**: Session history visible for understanding orchestrator decisions

**Task Routing**:
- Simple tasks: Orchestrator handles directly (avoids overhead)
- Complex tasks: Spawn workers
- Tradeoff: Balance context window bloat vs. token/latency costs
- Modern 256k context windows make simple tasks safe to handle inline

**Index Maintenance**:
- LLM-assisted updates during sink updates
- 90% of updates won't require index changes (incremental refinement)
- "Manual" = LLM-involved, not human hand-editing
- Living document that evolves with sinks
- Example: "Go style guide" unlikely to fundamentally change purpose

**Cost/Latency Philosophy**:
- Accept orchestrator overhead for better context management
- **Key Strategy**: Context engineering
- Workers receive minimal, targeted context
- Prevents context window bloat that degrades performance
- Industry trend toward sub-agents for this reason
- Avoids constant compaction/restarts (poor UX)

---

## Deep Dive: Data Layer Structure & Skills System

### `.sow/` Folder Purpose

**Key Realization**: `.sow/` is NOT purely data - it contains both data AND execution components

**Why `.sow/`?**
- Avoid cluttering actual repo contents
- Centralized location for system-of-work artifacts
- Clear separation from user code

### Required Data Components

1. **Repo-Specific Information**
   - Overview documents
   - Repo-specific deviations from sinks
   - Design/architecture documents
   - Architecture Decision Records (ADRs)

2. **Data Sinks** (External Knowledge)
   - Installed into `.sow/sinks/`
   - Git-ignored, externally sourced
   - Index-driven discovery

3. **External Repositories**
   - Cloned or symlinked repos
   - Location: `.sow/repos/`
   - For multi-repo context

4. **Current Tasks**
   - Active work items
   - Task tracking/state

5. **State Files**
   - "Single pane of glass" view
   - Overall work state dashboard

### Skills = Slash Commands [DESIGN DECISION]

**Problem**: Initial proposal for "skills" as separate system overlaps with slash commands

**Solution**: Skills ARE slash commands, agents reference them

**Rationale**:
- Avoids new abstraction layer
- Leverages existing Claude Code feature
- Prevents context window bloat
- Maintains composability

**Implementation**:
```yaml
---
name: architect
description: "System architect"
tools: Read, Write, Grep
---
You are a system architect with the following capabilities:

## Create ADR
Use the `/create-adr` slash command

## Create Design Document
Use the `/design-document` slash command

## Review Architecture
Use the `/review-architecture` slash command
```

**Benefits**:
- Agent prompts stay concise (references, not full instructions)
- Skills (slash commands) can be 100+ lines without bloating prompts
- Slash commands already support subdirectories for organization
- Distribution via plugin system (already solved)
- Users can invoke skills directly OR through agents

**Structure**:
```
.claude/commands/
├── skills/
│   ├── create-adr.md
│   ├── design-document.md
│   ├── write-tests.md
│   └── refactor-performance.md
└── workflows/
    ├── init.md
    └── migrate.md
```

**Design Principle**: Don't introduce new systems when existing ones suffice

---

## Deep Dive: Work Management - Projects & Tasks

### Core Constraints
1. Support mono-repos with multiple services
2. Handle work of varying size (small bug fixes → large features)
3. Multiple incomplete tasks can exist simultaneously

### Two-Tier Hierarchy

**Project**: Composition of one or more tasks, used to coordinate complex work
**Task**: Single unit of work within a project

### When Projects Are Used
- Only for non-trivial work
- Trivial work handled directly by orchestrator (no project structure)
- Decision point: TBD (heuristic, user flag, always ask?)

### Project Structure
```
.sow/projects/<project-name>/
├── state.yaml              # Overall project state (phases + tasks)
├── log.md                  # Append-only action log (orchestrator)
├── context/                # Project-specific context
│   ├── overview.md
│   ├── decisions.md
│   └── memories.md
└── phases/
    ├── phase-01/
    │   └── tasks/
    │       ├── 001/
    │       │   ├── state.yaml      # Task state
    │       │   ├── log.md          # Append-only action log (worker)
    │       │   ├── description.md  # Task description
    │       │   └── feedback/       # Corrections/changes from humans
    │       │       ├── 001.md
    │       │       └── 002.md
    │       └── 002/
    │           └── ...
    └── phase-02/
        └── tasks/
            └── 001/
                └── ...
```

**Rationale for Phases as Directories** (Option A):
- ✓ Filesystem discoverability (agents/humans can `ls` to explore)
- ✓ No YAML cross-referencing required
- ✓ Aligns with "filesystem as data plane" principle
- ✓ Supports zero-context resumability
- ✓ Easy archival (`mv phase-01 archive/`)
- ✓ Clear visual organization

### Zero-Context Resumability

**Design Goal**: Any agent should be able to resume any project/task from scratch

**Orchestrator Recovery**:
1. Read project `state.json` (what tasks exist, their status)
2. Read project `log.md` (what actions were taken)
3. Determine next action (resume incomplete task, start new task)
4. Spawn worker with context

**Worker Recovery**:
1. Read task `state.json` (current progress)
2. Read task `log.md` (what's been done)
3. Read `feedback/` (any corrections needed)
4. Continue work

### Resolved Design Decisions

**Action Logs**:
- ✓ Written by agents (orchestrator for project logs, workers for task logs)
- ✓ Write-ahead-log pattern (like databases)
- ✓ Chronological stream of actions from top to bottom
- ✓ Provides richer context than state file alone
- ✓ Not intended to "sync" with state - separate concerns
- Open: JSON (structured, machine-readable) vs Markdown (human-readable)

**State Files**:
- ✓ YAML format
- ✓ Contains basic metadata about current state
- ✓ Potentially serves as central planning document
- ✓ Structure: Phases containing tasks
- ✓ Tracks completion status of phases/tasks
- ✓ Not derived from logs - independently maintained

**Feedback Mechanism**:
- ✓ Orchestrator prompts user for feedback
- ✓ Orchestrator creates: `phase-01/tasks/001/feedback/001.md`
- ✓ Chronological numbering (001, 002, etc.)
- ✓ Workers always read feedback directory
- ✓ Task state file may track which feedback items are satisfied

**Git Versioning**:
- ✓ Projects/tasks are NOT committed
- ✓ Ephemeral, per-developer state
- ✓ General pattern: one branch = one project
- ✓ Deleted after branch work complete (PRs merged)
- ✓ Not replacing existing productivity tools (SCRUM, Jira, etc.)
- ✓ Commit messages/changelogs handle "what happened"

**Project Lifecycle**:
- ✓ User always specifies: new project vs existing project
- ✓ Human decision, not orchestrator
- ✓ Don't extend completed projects - create new
- ✓ Projects are short-lived (not multi-week)
- ✓ Deleted when complete

**Concurrency**:
- ✓ Multiple tasks can run in parallel
- ✓ State file marks tasks as sequential or parallel
- ✓ Orchestrator coordinates parallelism

**Entry Points** (Slash Commands):
- ✓ Three separate commands for different modes:
  - Complete trivial task (no project)
  - Start new project
  - Continue existing project
- ✓ Trivial vs non-trivial is human decision

**Scope**:
- ✓ Building full specification upfront
- ✓ Not minimizing for MVP (evolves from existing work)
- ✓ All features included

---

## Deep Dive: Project State & Planning

### State File Format
- ✓ YAML (human-readable + structured)
- ✓ Agents can modify programmatically
- ✓ Humans can read/edit if needed

### Planning Philosophy

**Problem**: Agents jump to coding without planning

**Solution**: Structured planning command with explicit steps:
1. Thoroughly examine scope
2. Rate complexity (1-3 metrics, TBD)
3. Based on rating, decide phases/tasks breakdown

### Fail-Forward Task Management

**Gap Numbering**: 010, 020, 030, 040
- Allows insertion: 011, 012, 021, 031
- No renumbering chaos
- Similar to database migrations

**Abandoned State**: Never delete tasks, mark as abandoned
- Preserves audit trail
- State file is append-only (mostly)

### Phases = Modes [DESIGN DECISION]

**Key Insight**: Phases and modes are the same concept

**Phase Selection**:
- Agent selects from pre-defined list
- Complexity determines which phases needed
- Simple work: just "implement" phase
- Complex work: "discovery", "design", "implement", "test"

**Dynamic Phase Addition**:
- ✓ Agent can add new phases during execution
- ✓ **Requires human approval** (orchestrator pauses)
- ✓ Switching between phases requires human review
- ✓ Project log tracks all phase changes

**Example Scenario**:
```yaml
# Initial plan (simple bug fix)
phases:
  implement:
    tasks:
      - id: 010
        name: Change timeout from 10s to 20s
        status: in_progress

# After discovering deeper issue
phases:
  discovery:  # NEW PHASE ADDED (requires approval)
    tasks:
      - id: 010
        name: Investigate race condition in function X
        status: in_progress
  implement:
    tasks:
      - id: 010
        name: Change timeout from 10s to 20s
        status: abandoned  # Marked abandoned, not deleted
```

**Benefits**:
- No separate "mode" system
- Agent autonomy with human gates
- Natural evolution of project structure
- Fail-forward approach

### Phase Design Details

**Phase Vocabulary** (7 phases):
- ✓ `discovery` - Research, investigation, understanding
- ✓ `design` - Architecture, planning, API design
- ✓ `implement` - Coding, building
- ✓ `test` - Writing tests, verification
- ✓ `review` - Code review, refactoring
- ✓ `deploy` - Deployment, release prep
- ✓ `document` - Documentation updates
- ✓ No catch-all (list refined via real-world feedback)

**Approval Mechanism**:
- ✓ Orchestrator must provide rationale for phase changes
- ✓ Explicit approval required: "I've discovered X. Need to add Y phase because Z. Approve? (y/n)"
- ✓ User can approve or provide feedback

**Task Addition Rules**:
- ✓ Tasks can ONLY be added to current active phase (after initial planning)
- ✓ Tasks must match phase category
- ✓ Example: Cannot add design tasks to deploy phase
- ✓ Each phase has clear definition with:
  - Purpose
  - When to use / not use
  - Valid task types
  - Invalid task types
- ✓ If agent needs different type of work, must request phase change

**Phase Ordering**:
- ✓ Phases are ordered (discovery → design → implement → test → review → deploy → document)
- ✓ **Forward movement blocked**: Cannot advance to next phase unless all tasks in current phase complete
- ✓ **Backward movement allowed**: Can return to previous phase, add tasks, complete them, then advance
- ✓ Prevents scenarios like: design → implement (incomplete) → review (out of order)

**Example Flow**:
```
1. Start in "implement" phase
2. Discover issue requiring investigation
3. Request approval: "Need to add discovery phase to investigate race condition"
4. User approves
5. Move to "discovery" phase, add task 010
6. Complete discovery task
7. Return to "implement" phase
8. Cannot move to "test" until all "implement" tasks complete
```

---

## Deep Dive: State File Schemas & Context Management

### Context Management Flow

**Problem**: Workers need curated, minimal context to avoid window bloat

**Solution**: Orchestrator acts as context compiler

**Orchestrator Responsibilities**:
1. Read project context, sinks, repos, etc.
2. Determine what's relevant for specific task
3. Create two artifacts:
   - `TASK.md` - Focused task description
   - References in `state.yaml` - List of files to read

**Worker Startup Process**:
1. Read `state.yaml` (task metadata, status)
2. Read `TASK.md` (what needs to be done)
3. Read all referenced files (sinks, project docs, code files)
4. Read `feedback/` if applicable
5. Begin work

**Benefits**:
- Workers receive minimal, targeted context
- Orchestrator does heavy lifting of filtering
- References use relative paths from `.sow/` root
- Frontloaded context prevents context thrashing

### Project state.yaml Schema

```yaml
# .sow/projects/add-authentication/state.yaml

project:
  name: add-authentication
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

### Task state.yaml Schema

```yaml
# .sow/projects/add-authentication/phases/implement/tasks/020/state.yaml

task:
  id: "020"
  name: Create JWT service
  phase: implement
  status: in_progress  # pending, in_progress, completed, abandoned

  created_at: 2025-10-12T15:25:00Z
  started_at: 2025-10-12T15:30:00Z
  updated_at: 2025-10-12T16:45:00Z

  assigned_agent: implementer

  # Context references (relative to .sow/ root)
  # Orchestrator compiles this list for worker
  references:
    - sinks/python-style/conventions.md
    - sinks/api-conventions/rest-standards.md
    - knowledge/architecture/auth-design.md
    - repos/shared-library/src/crypto/jwt.py  # Example from another repo

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

### Task Description File

```markdown
# .sow/projects/add-authentication/phases/implement/tasks/020/TASK.md

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

### Design Decisions

**Complexity Metrics**:
- ✓ Sufficient as-is (rating + basic metrics)
- ✓ Can be refined based on real-world usage

**Time Tracking**:
- ✗ Do NOT track estimated time (LLMs are notoriously bad at this)
- ✓ Track actual timestamps (created, started, completed)

**Parallel Flag**:
- ✓ Needed in project state file
- ✓ Orchestrator uses to coordinate parallel workers

**Phase Statistics**:
- ✗ Not yet (premature optimization)
- ✓ Can add later if needed

---

## Deep Dive: Single Project Constraint

### Design Decision: One Project Per Branch

**Change**: `.sow/projects/<name>/` → `.sow/project/` (singular)

**Rule**: Only one project can exist at a time in a repository

**Rationale**:
- Enforces "one branch = one project" as mandatory, not suggestion
- Leverages git's branching model instead of duplicating it
- Dramatically simplifies orchestrator logic (no "which project?" state)
- Cleaner filesystem structure
- Natural cleanup via branch deletion

**Benefits**:
1. **Simplicity**: Orchestrator always knows current project (read `.sow/project/`)
2. **Commands**: `/continue` needs no project name argument
3. **Git Integration**: Switch branches = switch projects automatically
4. **Cleanup**: Delete branch = state gone (git-ignored)
5. **Mental Model**: Clear 1:1 mapping between branches and projects

**Safeguards**:
- ✓ Error if project exists when running `/start-project`
- ✓ Prevent project creation on `main`/`master` (force feature branches)
- ✓ Track branch name in project state.yaml

**User Workflows**:
- Want different project? → `git switch other-branch`
- Preserve WIP? → `git stash` (standard practice)
- Check project status? → Always `.sow/project/state.yaml`

**Updated Commands**:
- `/start-project <name>` → Creates `.sow/project/` (name in state.yaml, branch tracked)
- `/continue` → Works on `.sow/project/` (no argument needed)

**State.yaml Addition**:
```yaml
project:
  name: add-authentication
  branch: feat/add-auth  # Branch this project belongs to
```

### Git Versioning: Committed, Not Ignored

**Critical Realization**: `.sow/project/` cannot be git-ignored

**Problem**: Git-ignored files don't switch with branches
- Branch A creates project → git-ignored folder persists
- Switch to Branch B → same folder still there (wrong context!)

**Solution**: Commit `.sow/project/` to feature branches

**Implementation**:
- ✓ `.sow/project/` is committed (not ignored)
- ✓ Each branch has its own project state in git
- ✓ Switch branches = git handles switching automatically
- ✓ Before merge: delete `.sow/project/` via `/cleanup` command
- ✓ CI enforces: fails if `.sow/project/` exists in PR
- ✓ Recommend squash-merge to keep main branch history clean

**Benefits**:
1. **Natural Git Behavior**: Switch branches = switch projects (automatic)
2. **Team Collaboration**: Project state shared with team (can hand off branches)
3. **Backup**: Project state pushed to remote (safety)
4. **CI Safety Net**: Can't accidentally merge project state to main
5. **Clean Main History**: Squash merge removes feature branch commits

**Workflow**:
```
1. Feature branch: .sow/project/ committed normally
2. Push branch: Team can pull and see project state
3. Ready to merge: /cleanup (deletes .sow/project/)
4. CI check: Fails if .sow/project/ exists
5. Squash merge: Main branch history stays clean
```

**CI Check** (example):
```yaml
- name: Ensure no active sow project
  run: |
    if [ -d ".sow/project" ]; then
      echo "Error: .sow/project/ must be removed before merging"
      echo "Run /cleanup or: rm -rf .sow/project"
      exit 1
    fi
```

**New Command**:
- `/cleanup` - Deletes `.sow/project/`, stages deletion, ready for final commit before merge

---

## Thoughts to Explore Further
- What are the minimal required components vs. optional?
- How does this work with existing projects vs. greenfield?
