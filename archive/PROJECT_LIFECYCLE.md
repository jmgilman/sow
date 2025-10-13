# Project Lifecycle (DRAFT)

**Status**: Draft - Not Authoritative
**Last Updated**: 2025-10-12
**Purpose**: Document the operational lifecycle of projects in `sow`

This document describes how projects are created, planned, executed, and completed within the `sow` system. For filesystem structure details, see `FS_STRUCTURE.md`.

---

## Overview

Projects in `sow` represent non-trivial work that requires planning and coordination. Trivial tasks are handled directly by the orchestrator without creating a project structure.

**Key Principles**:
- **One project per branch** (enforced constraint, not suggestion)
- Only `.sow/project/` exists (singular) - no multiple projects
- **Project state committed to feature branches** (enables branch switching, team collaboration)
- Projects are ephemeral (deleted before merge via `/cleanup`, CI enforced)
- Cannot create projects on `main`/`master` (forces feature branches)
- Recommend squash-merge to keep main branch history clean
- Fail-forward approach (add tasks, don't revert)
- Phases guide work structure
- Human gates at critical decision points
- Zero-context resumability

---

## Project Creation

### Entry Points

Users initiate work through slash commands:

1. **`/trivial`** - Orchestrator handles directly, no project created
2. **`/start-project <name>`** - Create new project (errors if project exists or on main/master)
3. **`/continue`** - Resume existing project (no name needed - always `.sow/project/`)
4. **`/cleanup`** - Delete project, stage deletion, prepare for merge

**Decision**: Human decides trivial vs. non-trivial (not the orchestrator)

**Safeguards**:
- `/start-project` errors if `.sow/project/` already exists
- `/start-project` errors if on `main` or `master` branch
- Branch name automatically tracked in `state.yaml`
- CI fails if `.sow/project/` exists in PR (forces cleanup before merge)

### Initial Planning

When a new project starts, orchestrator executes planning workflow:

**Planning Steps**:
1. **Examine Scope**: Thoroughly analyze the request
2. **Rate Complexity**: Use 1-3 metrics to assess difficulty
   - Estimated files affected
   - Cross-cutting concerns
   - New dependencies required
3. **Decide Structure**: Based on complexity, select phases and initial tasks

**Complexity Rating**:
- `1` - Simple (few files, focused scope)
- `2` - Moderate (multiple files, some integration)
- `3` - Complex (many files, architectural changes, cross-cutting)

**Output**: Creates `.sow/project/state.yaml` with:
- Project metadata (name, description)
- Branch name (auto-detected)
- Complexity assessment
- Selected phases
- Initial task breakdown (using gap numbering: 010, 020, 030)

---

## Phases & Modes

### Unified Concept

**Phases ARE modes** - They serve as both:
- Structural organization (folders on disk)
- Operational modes (how orchestrator behaves)

### Available Phases (7 standard phases)

1. **`discovery`** - Research, investigation, understanding existing systems
2. **`design`** - Architecture planning, API design, technical specifications
3. **`implement`** - Coding, building features
4. **`test`** - Writing tests, verification, validation
5. **`review`** - Code review, refactoring, quality improvements
6. **`deploy`** - Deployment preparation, release processes
7. **`document`** - Documentation updates, README changes

**Phase Selection**:
- Not all phases required for every project
- Simple projects might only use: `implement` → `test`
- Complex projects might use: `discovery` → `design` → `implement` → `test` → `review` → `document`

### Phase Ordering Rules

**Sequential with Backward Movement**:
- ✅ Can move backward to previous phases
- ❌ Cannot skip forward unless current phase complete
- ✅ Can add tasks to earlier phases and complete them
- ❌ Cannot advance to next phase with incomplete tasks

**Example Flow**:
```
1. Start in "implement" phase
2. Discover fundamental design issue
3. Request approval to add "design" phase
4. Move back to "design", add tasks
5. Complete design tasks
6. Return to "implement"
7. Complete all implement tasks
8. Advance to "test" phase
```

### Phase Transitions

**Adding New Phases** (after initial planning):
- Requires human approval
- Orchestrator must provide rationale
- Example: "I've discovered X issue. Need to add Y phase because Z. Approve? (y/n)"

**Switching Between Phases**:
- Requires human review when adding new phases
- Moving back to existing phases is allowed
- Forward progression blocked until current phase complete

**Task Addition Rules**:
- Tasks can ONLY be added to currently active phase (after initial planning)
- Tasks must match phase category (no design tasks in deploy phase)
- Each phase has clear definition of valid/invalid task types
- If different type of work needed, must request phase change

---

## Orchestrator / Worker Pattern

### Roles

**Orchestrator** (Main Agent):
- User-facing Claude Code session
- High-level coordination and planning
- Does NOT write code directly
- Manages project state
- Routes work to specialist workers
- Compiles context for workers

**Workers** (Sub-Agents):
- Task-specific executors
- Receive bounded context from orchestrator
- Execute implementation work
- Report results back to orchestrator
- Examples: architect, implementer, reviewer, documenter

### Context Management

**Problem**: Workers need minimal, curated context to avoid window bloat

**Solution**: Orchestrator acts as context compiler

**Orchestrator Context Compilation**:
1. Reads project context, sinks, repos, external references
2. Determines what's relevant for specific task
3. Creates focused task description (`description.md`)
4. Lists file references in task state.yaml
5. Spawns worker with curated context

**Worker Startup Process**:
1. Read `state.yaml` (metadata, iteration, references)
2. Read `description.md` (what needs to be done)
3. Read all referenced files (sinks, docs, code examples)
4. Read `feedback/` if corrections exist
5. Begin work

**Benefits**:
- Workers receive minimal, targeted context
- Prevents context window bloat
- Orchestrator does heavy filtering
- Workers can start immediately with right information

---

## Task Execution

### Gap Numbering System

Tasks use gap numbering to allow insertions:
- Initial: `010`, `020`, `030`, `040`
- Insertions: `011`, `012`, `021`, `031`
- No renumbering needed
- Similar to database migration numbering

### Task States

- `pending` - Not yet started
- `in_progress` - Currently being worked on
- `completed` - Successfully finished
- `abandoned` - Started but deprecated (never deleted)

### Iteration Management

**Iteration Counter** tracks attempts at completing a task:

```yaml
task:
  id: "020"
  iteration: 3  # Third attempt
  status: in_progress
```

**When Iterations Increment**:
- Worker gets stuck, task paused, new worker assigned → increment
- Feedback added, worker resumes → increment
- Worker completes successfully → no increment (same attempt)

**Responsibility**:
- Orchestrator manages iteration counter
- Increments before spawning worker
- Worker reads current iteration from state.yaml

**Agent ID Construction**:
- Format: `{agent_role}-{iteration}`
- Examples: `implementer-1`, `implementer-2`, `architect-3`
- Used in logs to track which attempt did what

### Parallel Execution

Tasks can be marked for parallel execution:

```yaml
tasks:
  - id: "030"
    name: Add login endpoint
    parallel: true  # Can run with 031

  - id: "031"
    name: Add password hashing utility
    parallel: true  # Can run with 030
```

**Orchestrator Coordination**:
- Identifies parallel tasks
- Spawns multiple workers simultaneously
- Manages completion and dependencies

---

## Logging

### Purpose

Action logs provide:
- Audit trail of all actions
- Recovery information
- Debugging context
- Learning data

### Format: Structured Markdown

Each log entry follows strict format:

```markdown
### 2025-10-12 15:30:42

**Agent**: implementer-3
**Action**: created_file
**Files**:
  - src/auth/jwt.py
**Result**: success

Created initial JWT service file with class structure and method stubs.

---
```

**Required Fields**:
- Timestamp (ISO 8601: YYYY-MM-DD HH:MM:SS)
- Agent ID (auto-constructed from role + iteration)
- Action (from predefined vocabulary)
- Result (success, error, partial)

**Optional Fields**:
- Files (list of affected files)
- Free-form notes (description)

### Action Vocabulary

Predefined action types:
- `started_task` - Beginning work
- `created_file` - New file created
- `modified_file` - Existing file changed
- `deleted_file` - File removed
- `implementation_attempt` - Attempted implementation
- `test_run` - Ran tests
- `refactor` - Code refactoring
- `debugging` - Debugging session
- `research` - Investigation/research
- `completed_task` - Task finished
- `paused_task` - Paused for feedback

### CLI-Driven Logging

**Performance Optimization**: Direct file editing is slow for agents (30+ seconds)

**Solution**: CLI command for fast logging

```bash
sow log \
  --file <path> \
  --action <action> \
  --result <result> \
  "Free-form notes"
```

**Example**:
```bash
sow log \
  --file src/auth/jwt.py \
  --file requirements.txt \
  --action modified_file \
  --result success \
  "Added cryptography dependency and implemented generate_token()"
```

**CLI Responsibilities**:
1. Read current iteration from task state.yaml
2. Auto-construct agent ID: `{role}-{iteration}`
3. Generate timestamp
4. Format entry
5. Append to log.md

**Benefits**:
- ⚡ Fast (single bash command)
- 🎯 Enforces format automatically
- 🧠 Reduces agent cognitive load
- 📝 Consistent formatting guaranteed

---

## Feedback Loop

### Purpose

Structured mechanism for humans to provide corrections and guidance.

### Feedback Creation

**Process**:
1. User identifies issue with task work
2. Explains problem to orchestrator
3. Orchestrator creates numbered feedback file:
   - `phases/<phase>/tasks/<id>/feedback/001.md`
   - `phases/<phase>/tasks/<id>/feedback/002.md`

**Chronological Order**: Always read 001 → 002 → 003 (ascending)

### Feedback Tracking

Task state.yaml tracks feedback:

```yaml
feedback:
  - id: "001"
    created_at: 2025-10-12T16:30:00Z
    status: addressed  # pending, addressed, superseded
```

**Statuses**:
- `pending` - Not yet addressed by worker
- `addressed` - Worker has incorporated feedback
- `superseded` - No longer relevant (newer feedback replaces)

### Worker Feedback Processing

When worker starts/resumes task:
1. Read all feedback files in order
2. Check state.yaml for which items are pending
3. Incorporate feedback into work
4. Update state.yaml to mark feedback as addressed

---

## Zero-Context Resumability

### Design Goal

Any agent should be able to resume any project/task from scratch without prior context.

### Orchestrator Recovery

When resuming a project:
1. Read `.sow/project/state.yaml` (phases, tasks, current status, branch)
2. Read `.sow/project/log.md` (history of actions taken)
3. Verify branch matches (safeguard against branch switching issues)
4. Determine next action (resume incomplete task, start new task)
5. Compile context for worker
6. Spawn worker with instructions

### Worker Recovery

When resuming a task:
1. Read `state.yaml` (metadata, iteration, references)
2. Read `description.md` (what needs to be done)
3. Read referenced files (context)
4. Read `log.md` (what's already been tried)
5. Read `feedback/` (any corrections)
6. Continue work from current state

### Benefits

- No dependency on conversation history
- Can pause/resume across sessions
- Multiple developers can work on same project
- Natural handoff between agents
- Transparent audit trail

---

## Project Completion

### When Projects End

Projects are ephemeral and short-lived:
- **Always**: one branch = one project (enforced)
- Completed when all phases done and work ready to merge
- Before merge: run `/cleanup` to delete project state
- CI enforces: fails if `.sow/project/` exists in PR
- After cleanup: commit deletion and push
- Recommend squash-merge: keeps main branch history clean
- Switch branches → different project context (git handles automatically)

### Cleanup Workflow

**Before creating PR or merging**:
```bash
# 1. All work complete
/cleanup

# 2. Orchestrator deletes .sow/project/ and stages deletion
# git rm -rf .sow/project/

# 3. Commit the cleanup
git commit -m "chore: cleanup sow project state"

# 4. Push and create PR
git push

# 5. CI verifies no .sow/project/ exists
# 6. Squash merge to main
```

### Why Project State is Committed

**Problem**: Git-ignored files don't switch with branches
- Branch A creates project → git-ignored folder persists in working directory
- Switch to Branch B → same folder still there (wrong context!)

**Solution**: Commit project state to feature branches
- Git manages switching automatically
- Team members can pull branch and see full project context
- Can hand off branches with complete state
- Natural backup (pushed to remote)
- CI enforces cleanup before merge to main
- Squash merge keeps main branch clean

### Not a Work Tracking System

`sow` does NOT replace:
- JIRA, Linear, GitHub Issues
- Sprint planning boards
- Long-term roadmaps

`sow` is for:
- Active, in-progress work coordination
- Short-term (hours to days) execution
- AI agent orchestration

### What Happens After Completion

- Project folder deleted locally
- Git commit history preserved
- PR descriptions capture what was done
- Changelog/release notes document changes
- External tools track long-term work

---

## Summary: Complete Lifecycle

```
1. User: git checkout -b feat/add-auth

2. User: /start-project "Add authentication"

3. Orchestrator: Validates and plans project
   - Checks not on main/master branch ✓
   - Checks no existing project ✓
   - Assesses complexity
   - Selects phases (design, implement, test)
   - Creates initial tasks (010, 020, 030)
   - Creates .sow/project/state.yaml (includes branch name)

3. Orchestrator: Starts first phase (design)
   - Compiles context for architect
   - Spawns architect worker

4. Worker (architect): Completes design tasks
   - Reads description.md + references
   - Creates design documents
   - Logs actions via CLI
   - Updates state

5. Orchestrator: Advances to implement phase
   - Compiles context for implementer
   - Spawns implementer worker

6. Worker (implementer): Discovers issue during task 020
   - Logs the problem
   - Pauses task

7. Orchestrator: Requests phase change
   - "Need to add discovery phase to investigate race condition. Approve?"
   - User: y

8. Orchestrator: Adds discovery phase
   - Creates task in discovery phase
   - Compiles context
   - Spawns worker

9. Worker: Completes discovery
   - Logs findings
   - Updates state

10. Orchestrator: Returns to implement phase
    - Increments iteration on task 020
    - Spawns new implementer worker

11. Worker (implementer-2): Resumes task 020
    - Reads discovery findings
    - Completes implementation
    - Logs completion

12. User provides feedback
    - Orchestrator creates feedback/001.md
    - Increments iteration
    - Spawns worker to address feedback

13. All tasks complete
    - User runs /cleanup
    - Orchestrator deletes .sow/project/ and stages deletion
    - User commits cleanup: "chore: cleanup sow project state"
    - User pushes and creates PR
    - CI verifies no .sow/project/ exists ✓
    - Squash merge to main (keeps history clean)
    - Work captured in git commit history, project state removed
```

---

## Open Questions

1. Should there be a maximum iteration count before escalating to user?
2. How do we handle abandoned tasks that never complete?
3. Should project log track phase transitions separately from task logs?
4. What happens if user rejects phase change proposal?
5. How do we handle emergency stops / rollbacks?
