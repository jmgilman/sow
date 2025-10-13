# Project Management

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Project Lifecycle](#project-lifecycle)
  - [Creation](#creation)
  - [Planning](#planning)
  - [Execution](#execution)
  - [Completion](#completion)
- [Phases System](#phases-system)
  - [Available Phases](#available-phases)
  - [Phase Selection](#phase-selection)
  - [Phase Ordering Rules](#phase-ordering-rules)
  - [Phase Transitions](#phase-transitions)
  - [Adding New Phases](#adding-new-phases)
- [Tasks Structure](#tasks-structure)
  - [Gap Numbering](#gap-numbering)
  - [Task States](#task-states)
  - [Task Iterations](#task-iterations)
  - [Parallel Execution](#parallel-execution)
  - [Agent Assignment](#agent-assignment)
- [Logging System](#logging-system)
  - [Purpose and Benefits](#purpose-and-benefits)
  - [Structured Markdown Format](#structured-markdown-format)
  - [Action Vocabulary](#action-vocabulary)
  - [CLI-Driven Logging](#cli-driven-logging)
  - [Project vs Task Logs](#project-vs-task-logs)
- [Feedback Mechanism](#feedback-mechanism)
  - [Creating Feedback](#creating-feedback)
  - [Feedback Storage](#feedback-storage)
  - [Feedback Processing](#feedback-processing)
  - [Feedback Tracking](#feedback-tracking)
- [State File Management](#state-file-management)
  - [Project State](#project-state)
  - [Task State](#task-state)
  - [State Maintenance](#state-maintenance)
- [Zero-Context Resumability](#zero-context-resumability)
  - [Design Goal](#design-goal)
  - [Orchestrator Recovery](#orchestrator-recovery)
  - [Worker Recovery](#worker-recovery)
- [Progressive Planning](#progressive-planning)
  - [Start Minimal](#start-minimal)
  - [Discover as You Go](#discover-as-you-go)
  - [Fail-Forward Approach](#fail-forward-approach)
- [Related Documentation](#related-documentation)

---

## Overview

`sow` uses a structured approach to project management that balances planning with flexibility. Projects are ephemeral units of work that coordinate complex, multi-task efforts.

**Key Principles**:
- **One project per branch** (enforced constraint)
- **Progressive planning** (start minimal, evolve)
- **Phases guide structure** (not rigid stages)
- **Fail-forward** (add tasks, don't revert)
- **Human gates** (approval for significant changes)
- **Zero-context resumability** (can resume anytime)

**Not a Replacement**: `sow` complements tools like JIRA/Linear/GitHub Issues, not replaces them. Use `sow` for active work coordination (hours to days), not long-term planning.

---

## Project Lifecycle

### Creation

**Entry Point**: `/start-project <name>`

**Prerequisites**:
- Must be on feature branch (not main/master)
- No existing project in repository
- Repository initialized with `/init`

**Process**:
1. User invokes `/start-project "Project name"`
2. Orchestrator validates branch (errors if on main/master)
3. Orchestrator validates no existing project
4. Orchestrator prompts for description
5. Initial planning begins

**Smart Branch Detection**:
```
# On main branch
/start-project "Add auth"

Error: Cannot create project on main/master branch

Would you like to create a new feature branch?
  Branch name suggestion: feat/add-auth

[y/n] → If yes: creates branch, then continues
```

### Planning

**Initial Planning Workflow**:

1. **Gather Requirements**:
   ```
   Describe what you want to build:

   User: Add JWT-based authentication with user login, token refresh,
         and password hashing. Should integrate with existing User model.
   ```

2. **Assess Complexity**:
   ```
   🔍 Analyzing requirements...

   Complexity: Moderate (rating: 2/3)

   Metrics:
   - Estimated files: 8-12
   - Cross-cutting concerns: Yes (auth middleware)
   - New dependencies: Yes (JWT library, bcrypt)
   ```

   **Rating Scale**:
   - `1` - Simple (few files, focused scope)
   - `2` - Moderate (multiple files, some integration)
   - `3` - Complex (many files, architectural changes, cross-cutting)

3. **Select Initial Phases** (1-2 phases, minimal start):
   ```
   📋 Proposed plan:

   Phases:
   - design (1 task)
   - implement (4 tasks)

   Note: Starting with minimal structure. Additional phases
   (test, review, document) can be added as needed.

   Approve? [y/n/modify]
   ```

   **Progressive Philosophy**:
   - Don't plan all phases upfront
   - Add phases when actually needed
   - Request user approval for new phases

4. **Create Initial Tasks**:
   ```
   ✓ Created project 'Add authentication'

   Branch: feat/add-auth

   Phase: design
   - 010: Design authentication flow [architect]

   Phase: implement
   - 010: Create User model [implementer]
   - 020: Create JWT service [implementer]
   - 030: Add login endpoint [implementer]
   - 040: Add password hashing utility [implementer]

   ✓ Committed to git
   ```

   **Gap Numbering**: Tasks use 010, 020, 030, 040 to allow insertions

5. **Store State**:
   - Creates `.sow/project/state.yaml` (includes branch name)
   - Creates `.sow/project/log.md` (empty, ready for logging)
   - Creates `.sow/project/context/` (for project-specific context)
   - Commits to git (feature branch)

### Execution

**Orchestrator Coordination**:

1. **Identify Next Task**:
   - Reads project state
   - Finds next pending task in active phase
   - Cannot advance phases until current phase complete

2. **Compile Context**:
   - Reads task requirements
   - Identifies relevant sinks (via index)
   - Gathers knowledge docs
   - References linked repos if needed
   - Creates focused context package

3. **Spawn Worker**:
   - Uses assigned agent (from state.yaml)
   - Provides curated context
   - Includes task description, references, feedback
   - Increments iteration counter

4. **Monitor Progress**:
   - Worker executes task
   - Worker logs actions via CLI
   - Worker reports completion
   - Orchestrator updates state

5. **Handle Phase Transitions**:
   - When phase complete, advance to next
   - If new phase needed, request user approval
   - Document rationale for changes

**Worker Execution**:

1. **Startup**:
   - Read `state.yaml` (metadata, iteration, references)
   - Read `description.md` (requirements)
   - Read referenced files (sinks, docs, code)
   - Read `feedback/` (if any corrections)

2. **Work**:
   - Execute task per requirements
   - Follow style guides from sinks
   - Log actions via CLI
   - Update state as work progresses

3. **Completion**:
   - Mark task complete in state
   - List modified files
   - Report back to orchestrator

**Example Flow**:
```
/continue

Orchestrator:
→ Reads state: task 020 pending
→ Compiles context: JWT service requirements + Python style guide + auth design
→ Spawns implementer worker

Implementer:
→ Reads description.md
→ Reads sinks/python-style/conventions.md
→ Reads knowledge/architecture/auth-design.md
→ Implements JWT service with TDD
→ Logs: "created_file src/auth/jwt.py"
→ Logs: "created_file tests/test_jwt.py"
→ Marks task 020 complete

Orchestrator:
→ Updates project state
→ Identifies next task: 030
→ Prompts user to continue
```

### Completion

**When Complete**:
- All tasks in all phases are done
- Ready to create pull request
- About to merge to main branch

**Cleanup Process**:

1. **Run /cleanup**:
   ```
   /cleanup

   ⚠️ Warning: 1 task is still pending
     → implement/040: Add password hashing utility

   Are you sure you want to clean up? [y/n]

   User: y

   ✓ Deleted .sow/project/
   ✓ Staged deletion for commit

   Next steps:
   1. Commit cleanup: git commit -m "chore: cleanup sow project state"
   2. Push and create PR
   3. CI will verify no .sow/project/ exists before merge
   ```

2. **Commit Cleanup**:
   ```bash
   git commit -m "chore: cleanup sow project state"
   git push origin feat/add-auth
   ```

3. **Create Pull Request**:
   - Via GitHub UI or CLI
   - CI checks that `.sow/project/` does not exist
   - If present, CI fails (enforces cleanup)

4. **Merge**:
   - Code review proceeds normally
   - **Recommend squash-merge** to keep main branch clean
   - After merge, delete feature branch

**Why Project State is Committed**:
- Git-ignored files don't switch with branches
- Committing enables automatic branch switching
- Team can collaborate on same branch
- Natural backup (pushed to remote)
- CI enforces cleanup before merge to main

---

## Phases System

### Available Phases

Seven standard phases guide the type of work:

| Phase | Purpose | Example Tasks |
|-------|---------|---------------|
| **discovery** | Research and investigation | Investigate performance issue, analyze logs, review documentation |
| **design** | Architecture and planning | Create ADR, write design doc, plan API structure |
| **implement** | Building features | Write code, add endpoints, create models |
| **test** | Integration and E2E testing | Write integration tests, E2E test user flows |
| **review** | Code quality improvements | Code review, refactoring, security audit |
| **deploy** | Deployment preparation | Update deployment configs, write migration scripts |
| **document** | Documentation updates | Update README, write API docs, add inline docs |

**Note**: Not all phases are required for every project. Simple work might only use `implement` → `test`.

### Phase Selection

**Initial Selection** (during planning):
- Orchestrator selects 1-2 phases based on complexity
- **Simple**: just `implement`
- **Moderate**: `design` + `implement`
- **Complex**: `discovery` + `design` OR `design` + `implement`

**Dynamic Addition**:
- Orchestrator can request adding phases during execution
- **Requires human approval**
- Provides rationale for why phase is needed

**Example**:
```
# Initial plan (simple bug fix)
phases:
  - implement

# During execution, discovers deeper issue
Orchestrator: "I've discovered a race condition that requires investigation.
               Need to add 'discovery' phase to properly diagnose. Approve? [y/n]"

User: y

# Updated plan
phases:
  - discovery   # NEW
  - implement
```

### Phase Ordering Rules

Phases follow a logical order:

```
discovery → design → implement → test → review → deploy → document
```

**Forward Movement** (blocked):
- Cannot advance to next phase unless all tasks in current phase complete
- Prevents scenarios like: design → implement (incomplete) → review (out of order)

**Backward Movement** (allowed):
- Can return to previous phase
- Add tasks to earlier phase
- Complete them, then move forward again

**Example**:
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

### Phase Transitions

**Within Phase**:
- Tasks execute sequentially or in parallel
- Orchestrator coordinates based on `parallel` flag

**Between Phases**:
- When all tasks in phase complete, phase status → `completed`
- Orchestrator checks if next phase exists
- If yes, activates next phase
- If no, prompts user for next action

**Example Transition**:
```
Phase: design
✓ 010: Design authentication flow [completed]

Phase status: completed

→ Transitioning to next phase: implement
→ Active phase: implement
→ Next task: 010
```

### Adding New Phases

**During Execution**:

Orchestrator discovers need for different type of work:

```
Orchestrator: "While implementing, I've discovered the authentication
               design has security flaws. Need to add 'review' phase to
               conduct security audit before continuing. Approve? [y/n]"

User: y

Orchestrator: "Adding 'review' phase..."

Updated phases:
  - design (completed)
  - implement (in progress - paused)
  - review (new - active)

Creating task: review/010 - Security audit of auth design [reviewer]
```

**Approval Requirements**:
- Orchestrator must provide clear rationale
- User explicitly approves or denies
- If denied, orchestrator continues without phase addition
- All phase changes tracked in project log

---

## Tasks Structure

### Gap Numbering

Tasks use gap numbering to allow insertions without renumbering:

**Initial Numbering**: 010, 020, 030, 040

**Insertions Allowed**: 011, 012, 021, 031

**Example**:
```yaml
# Initial tasks
tasks:
  - id: "010"
  - id: "020"
  - id: "030"

# Need to insert task between 020 and 030
tasks:
  - id: "010"
  - id: "020"
  - id: "021"  # INSERTED
  - id: "030"

# No renumbering needed!
```

**Benefits**:
- No renumbering chaos
- Clear chronological order
- Similar to database migrations
- Can always insert new tasks

### Task States

| State | Meaning | Transitions |
|-------|---------|-------------|
| `pending` | Not yet started | → in_progress |
| `in_progress` | Currently being worked on | → completed, abandoned |
| `completed` | Successfully finished | (terminal) |
| `abandoned` | Started but no longer relevant | (terminal) |

**Never Delete Tasks**:
- Mark as `abandoned` instead
- Preserves audit trail
- Maintains gap numbering
- Shows what was tried

**Example**:
```yaml
tasks:
  - id: "010"
    name: Change timeout from 10s to 20s
    status: abandoned  # Discovered this wasn't the real issue

  - id: "020"
    name: Fix race condition in auth service
    status: completed  # Actual fix
```

### Task Iterations

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
- **Orchestrator manages** iteration counter
- Increments before spawning worker
- Worker reads current iteration from state.yaml

**Agent ID Construction**:
- Format: `{agent_role}-{iteration}`
- Examples: `implementer-1`, `implementer-2`, `architect-3`
- Used in logs to track which attempt did what

**Example**:
```
Task 020 history:

Iteration 1 (implementer-1):
- Attempted implementation
- Got stuck on import error
- Paused

Iteration 2 (implementer-2):
- Fixed import issue
- Discovered design flaw
- Feedback added

Iteration 3 (implementer-3):
- Incorporated feedback
- Completed successfully
```

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
1. Identifies parallel tasks (same `parallel: true` group)
2. Spawns multiple workers simultaneously
3. Each worker works independently
4. Orchestrator waits for all to complete before advancing

**Benefits**:
- Faster execution
- Independent work streams
- Natural task parallelism

**Guidelines**:
- Only parallel if truly independent
- No shared file modifications
- Clear separation of concerns

### Agent Assignment

**Assignment at Planning Time**:

During project creation, orchestrator assigns agent to each task:

```yaml
tasks:
  - id: "010"
    name: Design authentication flow
    assigned_agent: architect  # Design work

  - id: "020"
    name: Create JWT service
    assigned_agent: implementer  # Code work

  - id: "030"
    name: Write integration tests
    assigned_agent: integration-tester  # Testing work

  - id: "040"
    name: Review security
    assigned_agent: reviewer  # Quality work
```

**Execution**:
1. Orchestrator reads `assigned_agent` from state
2. Spawns correct worker type
3. No guessing or inference needed

**Flexibility**:
- Can change assignment if needed
- Update state.yaml manually
- Or let orchestrator reassign on retry

---

## Logging System

### Purpose and Benefits

Action logs provide:
- **Audit trail** of all actions taken
- **Recovery information** for resuming work
- **Debugging context** when things go wrong
- **Learning data** for understanding agent behavior
- **Chronological history** of project evolution

**Two Types**:
1. **Project Log** - Orchestrator actions (high-level coordination)
2. **Task Log** - Worker actions (detailed implementation)

### Structured Markdown Format

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
- **Timestamp** - ISO 8601 format: YYYY-MM-DD HH:MM:SS
- **Agent** - Agent ID (auto-constructed from role + iteration)
- **Action** - From predefined vocabulary
- **Result** - success, error, or partial

**Optional Fields**:
- **Files** - List of affected files
- **Notes** - Free-form description (after required fields)

**Benefits of Markdown**:
- Human-readable
- Git-friendly
- Easy to search/grep
- Supports rich formatting
- LLMs parse easily

### Action Vocabulary

Predefined action types ensure consistency:

| Action | When to Use | Example |
|--------|-------------|---------|
| `started_task` | Beginning work on task | Started implementing JWT service |
| `created_file` | New file created | Created src/auth/jwt.py |
| `modified_file` | Existing file changed | Updated User model with auth fields |
| `deleted_file` | File removed | Removed deprecated auth.py |
| `implementation_attempt` | Attempted implementation | Tried to implement token refresh |
| `test_run` | Ran tests | Ran unit tests for JWT service |
| `refactor` | Code refactoring | Refactored token validation logic |
| `debugging` | Debugging session | Debugged token expiration issue |
| `research` | Investigation/research | Researched JWT best practices |
| `completed_task` | Task finished | Completed JWT service implementation |
| `paused_task` | Paused for feedback | Paused due to design question |

**Extensible**: Can add new actions as needed

### CLI-Driven Logging

**Problem**: Direct file editing is slow for agents (30+ seconds per edit)

**Solution**: CLI command for fast logging

**Usage**:
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
1. Determines context (project vs task log)
2. Reads current iteration from task state.yaml
3. Auto-constructs agent ID: `{role}-{iteration}`
4. Generates timestamp
5. Formats entry with required fields
6. Appends to appropriate log.md

**Performance**:
- ⚡ Fast (single bash command vs 30s file edit)
- 🎯 Enforces format automatically
- 🧠 Reduces agent cognitive load
- 📝 Consistent formatting guaranteed

**Auto-Detection**:
```bash
# From task directory
cd .sow/project/phases/implement/tasks/020/
sow log --action created_file --result success "Created JWT service"
# Writes to: ./log.md

# From project root
sow log --action started_phase --result success "Starting implement phase"
# Writes to: .sow/project/log.md
```

### Project vs Task Logs

**Project Log** (`.sow/project/log.md`):
- **Written by**: Orchestrator
- **Content**: High-level coordination events
- **Examples**:
  - Started project
  - Created phase
  - Spawned worker
  - Phase transitions
  - Requested user approval

**Task Log** (`phases/<phase>/tasks/<id>/log.md`):
- **Written by**: Workers
- **Content**: Detailed implementation actions
- **Examples**:
  - Created files
  - Modified code
  - Ran tests
  - Debugging
  - Implementation attempts

**Separation Benefits**:
- Clear responsibility
- Appropriate detail level
- Easy to navigate
- Parallel logs for parallel tasks

---

## Feedback Mechanism

### Creating Feedback

**Purpose**: Structured way for humans to provide corrections and guidance

**When to Use**:
- Agent made incorrect assumption
- Implementation doesn't meet requirements
- Code needs specific changes
- Agent got stuck on wrong approach

**Process**:
1. User identifies issue with task work
2. User explains problem to orchestrator
3. Orchestrator creates numbered feedback file
4. Orchestrator increments task iteration
5. Orchestrator spawns worker with feedback context

**Example**:
```
You: "The JWT service should use RS256, not HS256"

Orchestrator: "I'll create feedback for the implementer and restart the task.
               Creating feedback/001.md..."

✓ Created feedback/001.md
✓ Incremented task iteration: 3 → 4
✓ Spawning implementer-4 to address feedback
```

### Feedback Storage

**Location**: `phases/<phase>/tasks/<id>/feedback/`

**Naming**: Chronologically numbered: `001.md`, `002.md`, `003.md`

**Format**:
```markdown
# Feedback 001

**Created**: 2025-10-12 16:30:00
**Status**: pending

## Issue

The JWT service is using HS256 (symmetric) algorithm, but we need RS256
(asymmetric) for better security and key management.

## Required Changes

1. Change algorithm from HS256 to RS256
2. Update to use public/private key pair instead of shared secret
3. Add key loading from environment variables
4. Update tests accordingly

## Context

Our infrastructure uses asymmetric keys for all JWT operations.
Keys are managed via Vault and loaded at startup.

Reference: `.sow/sinks/security-guidelines/jwt-best-practices.md`
```

**Chronological Order**: Always read 001 → 002 → 003 (ascending)

### Feedback Processing

**Worker Startup with Feedback**:

1. **Read Task State**:
   ```yaml
   task:
     id: "020"
     iteration: 4
     feedback:
       - id: "001"
         status: pending
       - id: "002"
         status: pending
   ```

2. **Read All Feedback Files** (in order):
   - `feedback/001.md`
   - `feedback/002.md`

3. **Incorporate Feedback**:
   - Address each feedback item
   - Make required changes
   - Run tests to verify

4. **Update State**:
   ```yaml
   feedback:
     - id: "001"
       status: addressed
     - id: "002"
       status: addressed
   ```

5. **Log Actions**:
   ```bash
   sow log --action modified_file --result success \
     "Addressed feedback 001: Changed JWT to RS256"
   ```

### Feedback Tracking

**Status Values**:
- `pending` - Not yet addressed by worker
- `addressed` - Worker has incorporated feedback
- `superseded` - No longer relevant (newer feedback replaces)

**Tracking in state.yaml**:
```yaml
task:
  id: "020"
  iteration: 5
  feedback:
    - id: "001"
      created_at: 2025-10-12T16:30:00Z
      status: addressed

    - id: "002"
      created_at: 2025-10-12T17:15:00Z
      status: addressed

    - id: "003"
      created_at: 2025-10-12T18:00:00Z
      status: pending  # Current worker needs to address this
```

**Multiple Iterations**:
- Each feedback item tracked independently
- Worker addresses all pending feedback
- Can have multiple rounds of feedback
- Iteration counter tracks attempts

---

## State File Management

### Project State

**File**: `.sow/project/state.yaml`

**Purpose**: Central planning document with:
- Project metadata (name, description, branch)
- Complexity assessment
- Phases and their status
- Tasks and their status
- Active phase indicator

**Key Fields**:
```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  created_at: 2025-10-12T14:30:00Z
  updated_at: 2025-10-12T16:45:00Z
  description: Add JWT-based authentication system

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
    tasks: [...]

  - name: implement
    status: in_progress
    created_at: 2025-10-12T15:22:00Z
    tasks: [...]
```

**Branch Field**: Critical for validation
- Tracks which branch project belongs to
- Orchestrator checks current branch matches
- Warns user if mismatch detected

### Task State

**File**: `phases/<phase>/tasks/<id>/state.yaml`

**Purpose**: Task-specific metadata:
- Task details (id, name, phase)
- Status and timestamps
- Iteration counter
- Assigned agent
- Context references
- Feedback tracking
- Modified files

**Key Fields**:
```yaml
task:
  id: "020"
  name: Create JWT service
  phase: implement
  status: in_progress

  created_at: 2025-10-12T15:25:00Z
  started_at: 2025-10-12T15:30:00Z
  updated_at: 2025-10-12T16:45:00Z

  iteration: 3
  assigned_agent: implementer

  references:
    - sinks/python-style/conventions.md
    - sinks/api-conventions/rest-standards.md
    - knowledge/architecture/auth-design.md

  feedback:
    - id: "001"
      created_at: 2025-10-12T16:30:00Z
      status: addressed

  files_modified:
    - src/auth/jwt.py
    - tests/test_jwt.py
```

**Auto-Populated**: `files_modified` filled by worker during execution

### State Maintenance

**Who Updates**:
- **Orchestrator** updates project state
- **Workers** update task state
- **CLI** can manually update if needed

**When Updated**:
- Task status changes
- Phase transitions
- Feedback added
- Iteration incremented
- Files modified

**Consistency**:
- State files are source of truth
- Logs provide audit trail but don't "sync" with state
- State can be manually corrected if needed

**Git Commits**:
- State changes committed regularly during work
- Pushed to remote (backup + collaboration)
- Team can see project progress via git

---

## Zero-Context Resumability

### Design Goal

**Goal**: Any agent should be able to resume any project/task from scratch without prior context or conversation history.

**Benefits**:
- Pause/resume across sessions
- Multiple developers can work on same project
- Natural handoff between agents
- Transparent audit trail
- No dependency on conversation history

### Orchestrator Recovery

**Process** (when resuming project):

1. **Read Project State**:
   ```bash
   cat .sow/project/state.yaml
   ```
   - What phases exist
   - What tasks exist
   - Current status of each
   - Active phase

2. **Read Project Log**:
   ```bash
   cat .sow/project/log.md
   ```
   - History of orchestrator actions
   - Phase transitions
   - Worker spawns
   - Approvals granted

3. **Verify Branch**:
   ```bash
   git branch --show-current
   ```
   - Check matches project.branch in state
   - Warn if mismatch

4. **Determine Next Action**:
   - Find next pending task in active phase
   - Or find in_progress task to resume
   - Or advance to next phase if current complete

5. **Compile Context for Worker**:
   - Read task description
   - Identify relevant sinks
   - Gather knowledge docs
   - Reference linked repos if needed

6. **Spawn Worker**:
   - Use assigned agent from state
   - Provide curated context
   - Worker has everything needed

### Worker Recovery

**Process** (when resuming task):

1. **Read Task State**:
   ```bash
   cat .sow/project/phases/implement/tasks/020/state.yaml
   ```
   - Current iteration
   - Assigned agent
   - Context references
   - Feedback status

2. **Read Task Description**:
   ```bash
   cat .sow/project/phases/implement/tasks/020/description.md
   ```
   - What needs to be done
   - Requirements
   - Acceptance criteria
   - Context notes

3. **Read Referenced Files**:
   - All files listed in `references` field
   - Sinks, knowledge docs, code examples
   - Provides domain knowledge

4. **Read Task Log**:
   ```bash
   cat .sow/project/phases/implement/tasks/020/log.md
   ```
   - What's already been tried
   - What worked, what didn't
   - Current state of implementation

5. **Read Feedback** (if any):
   ```bash
   cat .sow/project/phases/implement/tasks/020/feedback/*.md
   ```
   - Any corrections needed
   - Specific requirements
   - Issues to address

6. **Continue Work**:
   - Has complete context
   - Knows what's been done
   - Knows what's needed
   - Can proceed without conversation history

**Example Recovery**:
```
# Monday: Start task
Implementer-1: Creates JWT service, gets stuck on crypto library

# Friday: Resume task
Implementer-2 (different agent, different session):
1. Reads state: iteration=2, status=in_progress
2. Reads description: Requirements for JWT service
3. Reads log: Implementer-1 got stuck on crypto import
4. Reads feedback: Use RS256 algorithm (from user)
5. Continues: Fixes import, implements RS256, completes task

No conversation history needed!
```

---

## Progressive Planning

### Start Minimal

**Philosophy**: Don't try to plan everything upfront

**Approach**:
- Start with 1-2 phases only
- Begin with most critical phase
- Prefer discovery as work progresses
- Add phases dynamically when needed

**Benefits**:
- Faster project startup
- Less upfront uncertainty
- Adapt to discoveries
- Avoid wasted planning

**Example**:
```
Simple feature:
  - Start with: implement

Moderate feature:
  - Start with: design → implement

Complex feature:
  - Start with: discovery → design
  - Or: design → implement
```

### Discover as You Go

**Dynamic Phase Addition**:

As work progresses, orchestrator discovers need for different phases:

```
Scenario: Implementing feature, discovers security flaw in design

Orchestrator: "While implementing authentication, I've discovered the
               token storage approach has security vulnerabilities.

               Need to add 'review' phase to conduct security audit
               before continuing.

               Rationale: Current approach stores tokens in localStorage
               which is vulnerable to XSS attacks. Need to review and
               redesign token storage strategy.

               Approve adding 'review' phase? [y/n]"

User: y

Orchestrator: "Adding 'review' phase..."

Updated phases:
  design (completed)
  implement (in_progress - paused)
  review (active)      # NEW

Moving to review phase to address security concerns...
```

**Human Gates**:
- Orchestrator must request approval
- Provides clear rationale
- User can approve or deny
- All changes tracked in logs

### Fail-Forward Approach

**Philosophy**: Add tasks instead of reverting

**Practices**:
- Never delete tasks
- Mark as `abandoned` if no longer relevant
- Add new tasks with gap numbering
- Preserves audit trail

**Example**:
```yaml
# Initial task (wrong approach)
- id: "010"
  name: Increase timeout from 10s to 20s
  status: abandoned

# After discovering real issue
- id: "020"
  name: Fix race condition in auth service
  status: completed

# Both preserved in history
```

**Benefits**:
- Shows what was tried
- Learn from mistakes
- Transparent decision-making
- No lost context

---

## Related Documentation

- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows and usage
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Available commands
- **[AGENTS.md](./AGENTS.md)** - Agent roles and coordination
- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command reference
- **[OVERVIEW.md](./OVERVIEW.md)** - System overview and concepts
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Design decisions and patterns
