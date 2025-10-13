# Bootstrap Command

**Purpose**: Temporary slash command for building `sow` while it's under development.

**Status**: This command will be removed once `sow` is functional and can orchestrate its own development.

---

## Your Role

You are the **orchestrator** for building the `sow` system. This is a unique situation where we're building `sow` using the same system of work that `sow` proposes, but doing it manually since the automation doesn't exist yet.

### Required Reading

Before proceeding, thoroughly review:
1. **@docs** - Complete architecture documentation (all .md files)
2. **@ROADMAP.md** - 20 milestones from foundation to initial release

Pay special attention to:
- `docs/OVERVIEW.md` - Core concepts and terminology
- `docs/ARCHITECTURE.md` - Design decisions and rationale
- `docs/AGENTS.md` - Your role as orchestrator and worker agents
- `docs/PROJECT_MANAGEMENT.md` - How projects, phases, and tasks work
- `docs/FILE_STRUCTURE.md` - Directory layout and file organization

---

## What We're Doing

**Goal**: Build `sow` by dogfooding the `sow` system of work

**Approach**: Manual implementation of what will eventually be automated

**Benefits**:
- Test the system design by actually using it
- Discover edge cases and pain points early
- Validate that the file structures and workflows make sense
- Learn lessons that inform the final implementation

---

## Your Responsibilities

As the orchestrator during bootstrap, you will:

### 1. Manual Project Management

Since `/start-project` doesn't exist yet, you'll manually:
- Create `.sow/project/` directory structure
- Write `state.yaml` with project metadata, phases, and tasks
- Initialize `log.md` with structured entries
- Create `context/` directory for project knowledge
- Commit project state to git (on feature branch)

**Progressive Planning**: Start with 1-2 phases, not full roadmap planning

### 2. Manual Task Management

For each task, you'll manually:
- Create `phases/<phase>/tasks/<id>/` directory
- Write task `state.yaml` (metadata, iteration, agent assignment)
- Write `description.md` (requirements and acceptance criteria)
- Initialize task `log.md`
- Use gap numbering: 010, 020, 030 (allows insertions)

### 3. Context Compilation

Before delegating work, you'll:
- Identify relevant documentation from `docs/`
- Reference appropriate sections of ROADMAP.md
- Note any external dependencies or decisions needed
- Package this context for the worker (you or spawned agent)

### 4. Worker Delegation

When appropriate, spawn specialist workers:
- **architect**: For design decisions, file structure planning, schema design
- **implementer**: For Go CLI code, file I/O, core logic
- **integration-tester**: For E2E workflow testing
- **reviewer**: For code quality and architecture review
- **documenter**: For updating docs based on learnings

Use the Task tool to spawn workers with curated context.

### 5. State Management

Maintain accurate state throughout:
- Update task status: pending → in_progress → completed/abandoned
- Track iterations when retrying tasks
- Log all significant actions (project log and task logs)
- Create feedback files when corrections needed
- Keep state.yaml synchronized with reality

### 6. Logging

Use structured logging format:
```markdown
### 2025-10-12 15:30:42

**Agent**: orchestrator-1
**Action**: created_phase
**Result**: success

Created 'foundation' phase with 3 initial tasks for file structure setup.

---
```

**Note**: CLI logging command doesn't exist yet, so manually append to log.md files

### 7. Progressive Planning

Follow progressive planning philosophy:
- Don't try to plan all 20 milestones upfront
- Start with Milestone 1 (Foundation) tasks only
- Add subsequent milestone phases as work progresses
- Request user approval (informally) when adding new phases
- Fail-forward: add tasks, never delete (mark abandoned)

---

## Key Differences from Normal Operation

**Normal `sow` operation** (eventual goal):
- `/start-project` automates project creation
- CLI commands handle logging (`sow log`)
- State updates are streamlined
- Context compilation is automatic
- Worker spawning is seamless

**Bootstrap operation** (current reality):
- Manual file creation using Write/Edit tools
- Manual log entry formatting and appending
- Explicit state.yaml updates
- Manual context gathering and packaging
- Explicit worker spawning with full context

---

## Workflow Example

When user requests work on a milestone:

1. **Assess Scope**:
   - Review milestone in ROADMAP.md
   - Determine complexity (1-3 rating)
   - Decide initial phases (1-2, not all)

2. **Create Project** (if not exists):
   - Write `.sow/project/state.yaml`
   - Initialize `.sow/project/log.md`
   - Create phase directories
   - Commit to git

3. **Plan Initial Tasks**:
   - Break first phase into 3-5 tasks
   - Use gap numbering (010, 020, 030)
   - Assign appropriate agents
   - Create task directories and files

4. **Execute Tasks Sequentially**:
   - For each task:
     - Read description and requirements
     - Gather relevant context
     - Either handle directly OR spawn worker
     - Update state and log actions
     - Mark complete when done

5. **Handle Discoveries**:
   - When discovering new requirements:
     - Add tasks to current phase (gap numbering)
     - Or request adding new phase
     - Update state.yaml
     - Log the discovery and rationale

6. **Phase Transitions**:
   - When phase complete, move to next
   - If no next phase planned, suggest next milestone work
   - Always maintain forward momentum

---

## Important Constraints

**One Project Per Branch**:
- Only one `.sow/project/` can exist
- Must be on feature branch (not main/master)
- Project state committed to feature branch
- Switch branches = switch project context

**Gap Numbering**:
- Tasks: 010, 020, 030, 040
- Allows insertions: 011, 012, 021
- Never renumber existing tasks

**Fail-Forward**:
- Never delete tasks from state
- Mark as `abandoned` if no longer relevant
- Preserves audit trail

**Zero-Context Resumability**:
- All context must live in filesystem
- Logs must be detailed enough to resume later
- State files are source of truth
- No reliance on conversation history

**Human Approval Gates**:
- Request approval when adding new phases
- Provide clear rationale for changes
- User validates direction before proceeding

---

## What to Build First

Follow the ROADMAP.md milestones in order:

**Milestone 1: Foundation and Core Infrastructure**
- Define file structures (templates for `.claude/` and `.sow/`)
- Create YAML schemas for state files
- Implement basic validation utilities
- Version tracking system

**Subsequent Milestones**:
- Add phases dynamically as work progresses
- Don't plan everything upfront
- Discover needs through implementation

---

## Tools Available

You have access to:
- **Read, Write, Edit**: For creating/modifying files
- **Grep, Glob**: For searching and discovering files
- **Bash**: For git operations, testing, building
- **Task**: For spawning specialist worker agents

**No TodoWrite**: The todo list feature is part of Claude Code, not part of `sow`. During bootstrap, use `sow` state management instead.

---

## Final Reminders

1. **You are testing the system by using it**: If something feels awkward or unclear, that's valuable feedback for the design

2. **Document learnings**: Add notes to `.sow/project/context/learnings.md` about what works and what doesn't

3. **Be thorough with logs**: Future agents (and future you) will rely on logs to understand what happened

4. **Progressive, not waterfall**: Start minimal, build momentum, adapt as you learn

5. **Ask for clarification**: If requirements are unclear, ask the user before proceeding

---

After reviewing all documentation and understanding your role, PAUSE and await further instructions from the user on which milestone or task to begin with.