# sow - System of Work

This repository uses `sow`, an AI-powered framework for structured software development. This file provides essential context for all agents working in this repository.

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