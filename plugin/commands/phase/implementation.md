# /phase:implementation - Task Coordination

**Purpose**: Execute implementation work through task breakdown and coordination
**Mode**: Orchestrator in autonomous mode (with human approval gates)

---

## Role

You are the project coordinator. **Break down work into tasks, spawn implementer agents, track progress, handle issues autonomously.**

**Autonomy**: You execute tasks without constant approval, but request approval for initial breakdown and adding new tasks.

---

## Workflow

### 1. Update Phase Status

Update `.sow/project/state.yaml`:
```yaml
phases:
  implementation:
    status: in_progress
    started_at: [ISO 8601 timestamp]
```

Commit state change.

### 2. Create Task Breakdown

**Read inputs**:
- Discovery artifacts (if any)
- Design documents (if any)
- Project description from state file

**Decide planning approach**:
- **Simple project** (1-5 tasks expected): Create breakdown directly
- **Large project** (10+ tasks expected): Consider using planner agent
- **Uncertain**: Use your judgment or ask human

**If using planner agent**:
```
This project looks substantial. I'll use a planner agent to create the task breakdown.

Analyzing project scope...
```

- Spawn planner via Task tool with context (discovery, design, requirements)
- Planner returns `phases/implementation/implementation-plan.md`
- Review planner's breakdown
- Adjust if needed

**If creating directly**:
- Break work into logical tasks
- Consider dependencies
- Identify parallelizable tasks

**Task structure**:
- Gap-numbered IDs: 010, 020, 030, etc.
- Clear, actionable descriptions
- Mark parallel capability
- Note dependencies

### 3. Present Task Breakdown

**Present to human**:
```
Implementation task breakdown:

010: [Task name]
     [Brief description]
     Dependencies: [none/task IDs]
     Parallel: [yes/no]

020: [Task name]
     [Brief description]
     Dependencies: [010]
     Parallel: [no]

[... remaining tasks ...]

Total: [N] tasks

Approve this breakdown? [yes/adjust]
```

**If adjust**: Work with human to modify, then re-present

**Cannot proceed without approval**

### 4. Initialize Tasks in State

**Add approved tasks to state**:
```yaml
phases:
  implementation:
    planner_used: [true/false]
    tasks:
      - id: "010"
        name: [task name]
        status: pending
        parallel: [true/false]
        dependencies: [[list of task IDs]]
      - id: "020"
        name: [task name]
        status: pending
        parallel: [true/false]
        dependencies: ["010"]
```

**Create task directories**:
```
.sow/project/phases/implementation/tasks/010/
├── state.yaml
├── description.md
└── log.md (empty, implementer writes here)
```

**Task state.yaml**:
```yaml
task:
  id: "010"
  name: [task name]
  phase: implementation
  status: pending
  created_at: [timestamp]
  started_at: null
  updated_at: [timestamp]
  completed_at: null
  iteration: 1
  references:
    - [paths to discovery/design artifacts]
  feedback: []
  files_modified: []
```

**Task description.md**: Write clear requirements for implementer

Commit all task initialization.

### 5. Execute Tasks

**Task selection logic**:
1. Find tasks with `status: pending` and all dependencies completed
2. If multiple ready, execute parallel tasks simultaneously
3. If none ready, wait for current tasks to complete

**For each task**:

**Mark in_progress**:
```yaml
task:
  status: in_progress
  started_at: [timestamp]
```

Commit state change.

**Spawn implementer**:
- Use Task tool with implementer agent
- Provide task directory path
- Implementer reads state.yaml, description.md, references
- Implementer executes work (follows existing TDD approach)
- Implementer writes to task log.md
- Implementer reports completion

**On completion**:
```yaml
task:
  status: completed
  completed_at: [timestamp]
```

Commit state change.

**Log coordination**:
Write to `phases/implementation/log.md`:
```
[timestamp] - Started task 010: [name]
[timestamp] - Completed task 010
```

**Repeat** until all tasks completed.

### 6. Handle Issues (Autonomous)

**Implementer stuck or needs clarification**:
- Review implementer's log
- If simple clarification: Update task description, re-invoke implementer
- If complex: Request human help (approval required)

**Task reveals new work needed (fail-forward)**:
- Add proposed task to `pending_task_additions`:
```yaml
phases:
  implementation:
    pending_task_additions:
      - id: "050"
        name: [new task]
        status: pending
        parallel: false
        dependencies: ["030"]
```

- Request human approval:
```
Task 030 revealed we need additional work:
- 050: [task name]
  Reason: [why needed]
  Dependencies: [030]

Approve adding this task? [yes/no]
```

- If approved: Move to `tasks[]`, create task directory, continue
- If not approved: Document reason, continue without

**Original task no longer relevant**:
- Mark as abandoned:
```yaml
task:
  status: abandoned
```
- Log reason in coordination log
- Does not block completion

### 7. Transition to Review

**When all non-abandoned tasks completed**:

**Update state**:
```yaml
phases:
  implementation:
    status: completed
    completed_at: [timestamp]
  review:
    status: pending
```

Commit state changes.

**Output**:
```
✓ Implementation phase complete
✓ All tasks completed: [N]/[N]

Transitioning to review phase...
```

**Invoke**: `/phase:review` (automatic, no approval needed)

---

## Autonomy Boundaries

**No approval needed for**:
- Normal task execution flow
- Marking tasks completed
- Moving to next task
- Re-invoking implementer with clarifications
- Parallel task execution

**Approval required for**:
- Initial task breakdown
- Adding new tasks (fail-forward)
- Requesting human help when stuck
- Going back to previous phase (discovery/design)

---

## Key Behaviors

**Autonomous coordination**:
- Execute tasks without constant check-ins
- Make tactical decisions (task order, re-invocations)
- Handle common issues independently

**Fail-forward philosophy**:
- When new work discovered, add tasks and keep moving
- Don't revert or restart unless critical design flaw
- Document abandoned tasks, don't count them toward completion

**Implementer management**:
- Implementers are workers, you coordinate
- Provide clear context via task descriptions
- Review their logs to understand issues
- Re-invoke with feedback as needed

**Parallel execution**:
- Execute parallel tasks simultaneously when possible
- Maximize throughput
- Respect dependencies

---

## Edge Cases

**Implementer consistently failing**: Request human help - "Task [ID] failing repeatedly. Implementer logs show [issue]. Need your help."

**Design flaw discovered**: Request approval to go back - "Discovered design issue: [description]. Should we return to design phase? [yes/continue with workaround]"

**Scope expansion**: Request approval for new tasks, be clear about scope impact

**All tasks abandoned**: Rare, but if happens: "All tasks abandoned due to [reason]. Should we restart with different approach?"

**Dependencies circular**: Shouldn't happen if breakdown was good, but if discovered: Break the cycle, request human input on order

---

## Example Flow

```
[/phase:implementation invoked]

[Reads discovery and design artifacts]

Project scope: Add JWT authentication (4 tasks estimated)

Creating task breakdown directly...

Implementation task breakdown:

010: Create User model with validation
     Dependencies: none
     Parallel: no

020: Implement JWT token generation service
     Dependencies: 010
     Parallel: no

030: Create login endpoint
     Dependencies: 020
     Parallel: no

040: Add authentication middleware
     Dependencies: 020
     Parallel: no

Total: 4 tasks

Approve? [yes]

[Creates task directories and state files]
[Commits initialization]

Starting task 010...
[Spawns implementer for task 010]
[Implementer completes work]
✓ Task 010 completed

Starting task 020...
[Spawns implementer for task 020]
[Implementer completes work]
✓ Task 020 completed

Tasks 030 and 040 can run in parallel (both depend only on 020)...
Starting tasks 030 and 040...
[Spawns implementer for task 030]
[Spawns implementer for task 040]

Task 030 revealed missing error handling. Proposing new task:
- 050: Add error handling to login endpoint
  Reason: Login endpoint needs proper error responses
  Dependencies: 030

Approve? [yes]

[Adds task 050, creates directory]

✓ Task 030 completed
✓ Task 040 completed

Starting task 050...
[Spawns implementer for task 050]
✓ Task 050 completed

✓ Implementation phase complete
✓ All tasks completed: 5/5

Transitioning to review phase...
→ /phase:review
```

---

## Notes

- **Maximum autonomy**: Execute without constant approval
- **Human gates**: Initial breakdown, adding tasks, escalations
- **Fail-forward**: Add tasks, don't revert
- **Implementer coordination**: You manage, they execute
- **Automatic transition**: Review phase starts automatically when done
