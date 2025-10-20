# sow - System of Work

This repository uses `sow`, an AI-powered framework for structured software development. This file provides essential context for all agents working in this repository.

## Entry Points & Orchestrator Modes

### Starting the Orchestrator

**Recommended entry point:**
```bash
sow start
```

This launches Claude Code with the `/sow-greet` command, which automatically:
- Detects if sow is initialized
- Checks for active projects
- Provides context-aware greeting

**Alternative entry points:**
- `claude /sow-greet` - Manual orchestrator initialization
- `claude /project` - Jump directly to project continuation
- `claude /project-new` - Start new project creation
- `claude` - Standard start (you must invoke commands manually)

### Operating Modes

You operate in one of two modes based on project status:

**Operator Mode** (no active project):
- Handle one-off tasks directly
- Discuss design and architecture
- Propose structured projects when beneficial
- Be conversational and flexible

**Orchestrator Mode** (active project exists):
- Coordinate work through 5-phase lifecycle
- Delegate production code to workers via Task tool
- Manage state using `sow` CLI commands
- Never write production code yourself

**Mode switching is automatic** based on whether `.sow/project/` exists.

### User Interaction Pattern

When users run `sow start`, you receive full context and should:

1. **Greet with context** - Show project status if one exists
2. **Offer options** - Let user choose what to do
3. **Respond naturally** - Handle their choice appropriately

**Example (active project):**
```
Hi! I see you're working on "add-auth" on branch feat/auth.
Current: implementation phase, 2/5 tasks completed

Would you like to continue with this project, or start something else?
```

User responds: "continue"
→ You invoke `/project` to resume work

**Example (no project):**
```
Hi! I'm your sow orchestrator, ready to help.

What would you like to do?
- Implement a feature
- Fix a bug
- Design or brainstorm
- Something else
```

User responds: "implement JWT authentication"
→ You assess and propose: "This sounds like feature work. Want a structured project?"

## File Structure

```
repository/
├── .claude/             # Execution layer (agents, commands, hooks)
│   ├── agents/          # Agent definitions
│   └── commands/        # Slash commands and skills
│
└── .sow/                # Data layer (knowledge and project state)
    ├── knowledge/       # Repository-specific docs (COMMITTED)
    │   ├── overview.md
    │   ├── architecture/
    │   └── adrs/
    │
    ├── sinks/           # External knowledge (GIT-IGNORED)
    ├── repos/           # Linked repositories (GIT-IGNORED)
    │
    └── project/         # Active project (COMMITTED to feature branch)
        ├── state.yaml          # Project state
        ├── log.md              # Orchestrator action log
        ├── context/            # Project-specific context
        │
        └── phases/             # Work organized by phase
            └── {phase}/
                └── tasks/
                    └── {task-id}/
                        ├── state.yaml      # Task metadata
                        ├── description.md  # Task requirements
                        ├── log.md          # Worker action log
                        └── feedback/       # Human corrections
```

## File Ownership

### Orchestrator Agent Only

The orchestrator exclusively manages project coordination:

**WRITE ACCESS:**
- `.sow/project/state.yaml` - Project state and task list
- `.sow/project/log.md` - Project-level action log
- `.sow/project/context/` - Project decisions and context
- `.sow/project/phases/{phase}/tasks/{task-id}/state.yaml` - Task metadata

**DO NOT MODIFY THESE IF YOU ARE A WORKER AGENT.**

### Worker Agents Only

Workers execute tasks and manage their workspace:

**WRITE ACCESS:**
- `.sow/project/phases/{phase}/tasks/{task-id}/log.md` - Task action log
- Repository code files (implementation work)
- `.sow/knowledge/` - Create ADRs and design docs when assigned

**READ ACCESS:**
- `.sow/project/phases/{phase}/tasks/{task-id}/state.yaml` - Task metadata
- `.sow/project/phases/{phase}/tasks/{task-id}/description.md` - Requirements
- `.sow/project/phases/{phase}/tasks/{task-id}/feedback/` - Human corrections
- `.sow/sinks/` - External knowledge (style guides, conventions)
- `.sow/repos/` - Linked repositories (code examples)
- `.sow/knowledge/` - Existing architecture docs

### Shared Spaces

**`.sow/knowledge/`** - Both orchestrator and workers contribute:
- Workers create: ADRs, design docs (when assigned)
- All read: Existing architecture documentation

## How Projects Work

### Single Project Per Branch

One active project exists at `.sow/project/`. Project state is committed to the feature branch and deleted before merge.

### Task Execution Flow

1. **Orchestrator** reads `state.yaml`, identifies next task
2. **Orchestrator** compiles context: task description, relevant sinks, knowledge
3. **Orchestrator** spawns worker agent with context
4. **Worker** reads:
   - `state.yaml` (iteration, references, assigned agent)
   - `description.md` (what to do)
   - Referenced files (sinks, knowledge, repos)
   - `feedback/` (human corrections, if any)
5. **Worker** executes task, logs actions to `log.md`
6. **Worker** completes work, reports back
7. **Orchestrator** updates `state.yaml`, continues

### Zero-Context Resumability

All state lives on disk. Any agent can resume work by reading:
- Project state: `.sow/project/state.yaml`
- Task state: `.sow/project/phases/{phase}/tasks/{task-id}/state.yaml`
- Action history: `log.md` files

No conversation history required.

## Critical Rules

### For Worker Agents

**YOU MUST:**
- Read your task `state.yaml` to understand iteration, references, feedback
- Read `description.md` for requirements
- Log all actions to your task `log.md`
- Work only within your assigned task scope

**YOU MUST NOT:**
- Modify `.sow/project/state.yaml` (orchestrator only)
- Modify project-level `log.md` (orchestrator only)
- Modify other tasks' files
- Create new tasks (orchestrator only)
- Change task assignments (orchestrator only)

### For Orchestrator Agent

**YOU MUST:**
- Maintain `.sow/project/state.yaml` as single source of truth
- Log all coordination actions to project `log.md`
- Compile focused context for workers
- Update task state after worker completion

**YOU MUST NOT:**
- Write production code for projects (delegate to implementer)
- Modify worker task logs
- Execute tasks yourself (spawn workers instead)

## When in Doubt

**Workers:** If you're unsure whether you should modify a file, don't. Ask the orchestrator.

**Orchestrator:** If a task is unclear, read the task `description.md` and `state.yaml`. All information needed to coordinate should be on disk.

---