# 6. Runtime View

## Scenario 1: New Project Initialization

**Trigger**: Developer runs `sow project` on a feature branch

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant CLI as sow CLI
    participant FS as Filesystem
    participant SM as State Machine
    participant Claude as Claude Code
    participant Orch as Orchestrator

    Dev->>CLI: sow project
    CLI->>FS: Check .sow/ exists
    FS-->>CLI: Initialized
    CLI->>FS: Check project exists
    FS-->>CLI: No project
    CLI->>FS: Get current branch
    FS-->>CLI: feat/auth-system
    CLI->>FS: Create project/ directory
    CLI->>SM: Initialize at PlanningActive
    SM->>FS: Write state.yaml
    CLI->>Claude: Generate planning prompt
    CLI->>Claude: Launch with prompt
    Claude->>Orch: Spawn orchestrator agent
    Orch->>Dev: "I'll help you plan this project. What are you building?"
```

**Steps**:
1. CLI validates `.sow/` exists (run `sow init` if not)
2. CLI checks if project already exists (error if so)
3. CLI validates not on protected branch (main/master)
4. CLI creates `.sow/project/` directory structure
5. CLI initializes state machine at `PlanningActive`
6. CLI writes initial `state.yaml` with project metadata
7. CLI generates planning phase prompt via prompts package
8. CLI launches Claude Code with compiled prompt
9. Orchestrator agent spawned by Claude Code platform
10. Orchestrator begins conversational planning with developer

**Data Written**:
- `.sow/project/state.yaml`: Project metadata, empty phase configurations
- `.sow/project/log.md`: Initial log entry (project created)
- `.sow/project/context/`: Empty directory for decisions

---

## Scenario 2: Task Execution (Worker Spawning)

**Trigger**: Orchestrator completes planning, begins implementation

```mermaid
sequenceDiagram
    participant Orch as Orchestrator
    participant CLI as sow CLI
    participant FS as Filesystem
    participant SM as State Machine
    participant Claude as Claude Code
    participant Worker as Worker Agent

    Orch->>CLI: sow project add-task "Implement JWT middleware"
    CLI->>FS: Read state.yaml
    CLI->>FS: Create task directory
    CLI->>FS: Write task state.yaml
    CLI->>FS: Write task description.md
    CLI->>SM: Fire EventTaskCreated
    SM->>SM: Transition to ImplementationExecuting
    SM->>FS: Update state.yaml
    CLI-->>Orch: Task created (task-001)

    Orch->>FS: Read task state, description
    Orch->>FS: Read referenced files (refs, knowledge)
    Orch->>Orch: Compile bounded context
    Orch->>Claude: Spawn worker via Task tool
    Claude->>Worker: Create new agent context
    Worker->>FS: Read task description.md
    Worker->>FS: Read task state.yaml
    Worker->>FS: Read references
    Worker->>Worker: Begin implementation
    Worker->>FS: Write code files
    Worker->>CLI: sow log "action" "result" --files "src/auth/jwt.ts"
    CLI->>FS: Append to task log.md
    Worker->>CLI: sow task set-status completed
    CLI->>SM: Check if all tasks complete
    SM->>SM: Transition to ReviewActive
    Worker-->>Orch: "Task completed"
```

**Steps**:
1. Orchestrator invokes CLI to create task
2. CLI creates task directory: `.sow/project/phases/implementation/tasks/task-001/`
3. CLI writes task `state.yaml` with metadata (iteration=1, status=pending)
4. CLI writes task `description.md` with requirements
5. CLI fires state machine event `EventTaskCreated`
6. State machine transitions to `ImplementationExecuting` (if first task)
7. CLI updates project `state.yaml` with new state
8. Orchestrator reads task state and description
9. Orchestrator reads referenced files (refs, knowledge, planning artifacts)
10. Orchestrator compiles bounded context for worker
11. Orchestrator spawns worker via Claude Code Task tool
12. Worker receives compiled prompt with task context
13. Worker reads task files from filesystem
14. Worker executes implementation (writes code)
15. Worker logs actions via CLI commands (`sow log`)
16. Worker updates task status via CLI (`sow task set-status completed`)
17. CLI checks if all tasks complete, fires state machine event if so
18. State machine transitions to `ReviewActive`
19. Worker reports completion back to orchestrator

**Data Written**:
- `.sow/project/phases/implementation/tasks/task-001/state.yaml`: Task metadata
- `.sow/project/phases/implementation/tasks/task-001/description.md`: Requirements
- `.sow/project/phases/implementation/tasks/task-001/log.md`: Action log
- `.sow/project/state.yaml`: Updated with task in task list, potentially new state
- Code files: `src/auth/jwt.ts`, etc.

---

## Scenario 3: Human Feedback Loop

**Trigger**: Developer provides correction during task execution

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Orch as Orchestrator
    participant CLI as sow CLI
    participant FS as Filesystem
    participant Worker1 as Worker (iter 1)
    participant Worker2 as Worker (iter 2)

    Worker1->>Dev: "Implemented JWT with HS256"
    Dev->>Orch: "Use RS256 instead, not HS256"
    Orch->>CLI: sow task add-feedback "Use RS256..."
    CLI->>FS: Create feedback/001.md
    CLI->>FS: Update task state (feedback entry)
    CLI->>CLI: Increment iteration counter
    CLI->>FS: Update task state (iteration=2)
    CLI-->>Orch: Feedback ID: 001

    Orch->>CLI: sow log "Added feedback" "Iteration 2 required"
    CLI->>FS: Append to project log.md
    Orch->>Orch: Terminate Worker1
    Orch->>FS: Read task state (iteration=2)
    Orch->>FS: Read feedback/001.md
    Orch->>Orch: Compile context with feedback
    Orch->>Worker2: Spawn with feedback context
    Worker2->>FS: Read task state (sees iteration=2)
    Worker2->>FS: Read feedback/001.md
    Worker2->>Worker2: Revise implementation (RS256)
    Worker2->>FS: Write corrected code
    Worker2->>CLI: sow task mark-feedback-addressed 001
    CLI->>FS: Update feedback status (addressed)
    Worker2->>CLI: sow task set-status completed
```

**Steps**:
1. Worker presents implementation to developer (via orchestrator)
2. Developer provides correction to orchestrator
3. Orchestrator invokes CLI to add feedback
4. CLI creates `feedback/001.md` file with correction
5. CLI updates task `state.yaml` with feedback entry
6. CLI increments task iteration counter (1 → 2)
7. CLI updates task `state.yaml` with new iteration
8. Orchestrator logs feedback action to project log
9. Orchestrator terminates current worker context
10. Orchestrator reads updated task state (sees iteration=2)
11. Orchestrator reads feedback file
12. Orchestrator compiles new context including feedback
13. Orchestrator spawns new worker with iteration=2 context
14. Worker reads task state, sees iteration=2
15. Worker reads all feedback files
16. Worker revises implementation addressing feedback
17. Worker marks feedback as addressed via CLI
18. CLI updates feedback status in task state
19. Worker completes task via CLI

**Data Written**:
- `.sow/project/phases/implementation/tasks/task-001/feedback/001.md`: Correction
- `.sow/project/phases/implementation/tasks/task-001/state.yaml`: Updated with feedback, iteration=2
- `.sow/project/log.md`: Feedback action logged
- Code files: Corrected implementation

---

## Scenario 4: Design Mode Session

**Trigger**: Developer runs `sow design` to create Arc42 docs

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant CLI as sow CLI
    participant FS as Filesystem
    participant Claude as Claude Code
    participant Agent as Design Agent

    Dev->>CLI: sow design
    CLI->>FS: Check .sow/ exists
    CLI->>FS: Check design/ exists
    FS-->>CLI: Does not exist
    CLI->>FS: Get current branch
    FS-->>CLI: design/arc42
    CLI->>FS: Create design/ directory
    CLI->>FS: Write design/index.yaml
    CLI->>Claude: Generate design mode prompt
    CLI->>Claude: Launch with prompt
    Claude->>Agent: Spawn design agent

    Agent->>Dev: "What are you designing?"
    Dev->>Agent: "Arc42 architecture docs for sow"
    Agent->>CLI: sow design add-output 01-introduction.md --target .sow/knowledge/architecture/arc42/
    CLI->>FS: Update design/index.yaml (outputs)
    Agent->>FS: Write design/01-introduction.md
    Agent->>CLI: sow design log "Created introduction section"
    CLI->>FS: Append to design/log.md

    Agent->>Dev: "Introduction complete. Ready for 02-constraints?"
    Dev->>Agent: "Yes, continue"
    Agent->>FS: Write design/02-constraints.md
    Agent->>CLI: sow design add-output 02-constraints.md --target .sow/knowledge/architecture/arc42/
```

**Steps**:
1. Developer runs `sow design` on a `design/*` branch
2. CLI checks if design session exists (doesn't)
3. CLI validates branch name (not on main/master)
4. CLI creates `.sow/design/` directory
5. CLI writes `design/index.yaml` with session metadata
6. CLI generates design mode prompt
7. CLI launches Claude Code with design prompt
8. Design agent spawned by Claude Code
9. Agent asks user what is being designed
10. Developer describes the design goal
11. Agent registers outputs via CLI commands
12. CLI updates `design/index.yaml` with output entries
13. Agent creates documents in `.sow/design/` workspace
14. Agent logs actions via CLI to `design/log.md`
15. Agent iterates through documents with user feedback
16. Upon completion, agent moves files to target locations
17. Agent creates PR with design documents

**Data Written**:
- `.sow/design/index.yaml`: Session metadata, inputs, outputs
- `.sow/design/log.md`: Action log
- `.sow/design/*.md`: Design documents (workspace)
- Final target locations: `.sow/knowledge/architecture/arc42/*.md`

---

## Scenario 5: Zero-Context Resumption

**Trigger**: Developer restarts work after pausing mid-task

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant CLI as sow CLI
    participant FS as Filesystem
    participant SM as State Machine
    participant Claude as Claude Code
    participant Orch as Orchestrator

    Note over Dev: Days later, different machine

    Dev->>CLI: sow project
    CLI->>FS: Check .sow/ exists
    FS-->>CLI: Initialized
    CLI->>FS: Check project exists
    FS-->>CLI: Yes, project/state.yaml found
    CLI->>FS: Load project state.yaml
    FS-->>CLI: State data (current_state=ImplementationExecuting)
    CLI->>SM: Reconstruct state machine at ImplementationExecuting
    CLI->>FS: Read project log.md
    CLI->>FS: Read task states (find in_progress task)
    CLI->>Claude: Generate continuation prompt
    CLI->>Claude: Launch with project context
    Claude->>Orch: Spawn orchestrator
    Orch->>FS: Read project state
    Orch->>FS: Read project log
    Orch->>FS: Read task states
    Orch->>Dev: "Resuming project 'auth-system'. Task 'JWT middleware' in progress. Continue?"
    Dev->>Orch: "Yes"
    Orch->>FS: Read task description, state, log, feedback
    Orch->>Orch: Compile context (iteration, references, feedback)
    Orch->>Claude: Spawn worker with compiled context
```

**Steps**:
1. Developer runs `sow project` (possibly on different machine/day)
2. CLI detects existing project at `.sow/project/`
3. CLI loads `state.yaml` from disk
4. CLI reconstructs state machine at current state from file
5. CLI reads project log for history
6. CLI identifies any in_progress tasks
7. CLI generates continuation prompt with current context
8. CLI launches Claude Code with compiled prompt
9. Orchestrator spawned with project context
10. Orchestrator reads project state from filesystem
11. Orchestrator reads project log for history
12. Orchestrator reads all task states
13. Orchestrator presents status to developer
14. Developer confirms continuation
15. Orchestrator reads in_progress task details
16. Orchestrator compiles context (description, state, log, feedback)
17. Orchestrator spawns worker with complete context
18. Worker resumes from exactly where previous worker left off

**Key Point**: No conversation history required. All context reconstructed from filesystem state.

---

## Scenario 6: Reference System Usage

**Trigger**: Orchestrator provides style guide reference to worker

```mermaid
sequenceDiagram
    participant Orch as Orchestrator
    participant CLI as sow CLI
    participant FS as Filesystem (.sow/)
    participant Cache as Cache (~/.cache/sow/)
    participant Remote as Remote Repo

    Orch->>CLI: sow refs add https://github.com/org/style-guide.git --type knowledge
    CLI->>FS: Read refs/index.json
    CLI->>Cache: Check if repo cached
    Cache-->>CLI: Not cached
    CLI->>Remote: git clone to cache
    Remote-->>Cache: Clone complete
    CLI->>Cache: Write cache/index.json (SHA, timestamp)
    CLI->>FS: Symlink cache repo to .sow/refs/style-guide
    CLI->>FS: Update refs/index.json (URL, type, description)
    CLI-->>Orch: Reference added

    Note over Orch: Later, during task creation
    Orch->>CLI: sow refs list --type knowledge
    CLI->>FS: Read refs/index.json
    CLI-->>Orch: style-guide (knowledge)
    Orch->>FS: Read task description.md template
    Orch->>FS: Write task description.md with ref
    Orch->>FS: Update task state.yaml (references: [.sow/refs/style-guide])

    Note over Orch: Worker reads references
    Orch->>Worker: Spawn with context
    Worker->>FS: Read task state.yaml (sees references)
    Worker->>FS: Read .sow/refs/style-guide/python-style.md
    Worker->>Worker: Apply style guide during implementation
```

**Steps**:
1. Orchestrator adds external reference via CLI
2. CLI reads committed refs index
3. CLI checks cache for repo existence
4. CLI clones repo to `~/.cache/sow/repos/` (if not cached)
5. CLI updates cache index with SHA and timestamp
6. CLI creates symlink from `.sow/refs/` to cache
7. CLI updates committed refs index with metadata
8. Later, orchestrator queries available references
9. CLI reads committed index, returns knowledge refs
10. Orchestrator includes reference path in task description
11. Orchestrator updates task state with references array
12. Worker spawned with task context
13. Worker reads task state, sees references
14. Worker reads reference files via symlink
15. Worker applies reference content during work

**Data Written**:
- `~/.cache/sow/repos/org-style-guide/`: Cloned repo
- `~/.cache/sow/index.json`: Cache metadata
- `.sow/refs/style-guide`: Symlink to cache
- `.sow/refs/index.json`: Committed metadata
- Task `state.yaml`: References array
