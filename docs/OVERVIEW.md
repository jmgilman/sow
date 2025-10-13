# System of Work (sow) - Overview

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## What is sow?

`sow` (system of work) is an AI-powered framework for software engineering that provides a structured, opinionated approach to building software with AI assistance. It leverages Claude Code's capabilities to create a unified development experience across projects through:

- **Multi-agent orchestration** - Specialized AI agents working together
- **Progressive planning** - Adaptive project management that evolves with your work
- **Knowledge integration** - Contextual information from multiple sources
- **Zero-context resumability** - Pick up where you left off, anytime

---

## Core Concepts

### Two-Layer Architecture

`sow` separates behavior from state through a two-layer architecture:

**Execution Layer** (`.claude/`)
- AI agents and their behaviors
- Slash commands and workflows
- Hooks and automation
- Distributed via Claude Code Plugin
- Version-controlled and shared across teams

**Data Layer** (`.sow/`)
- Project knowledge and documentation
- Imported external knowledge (sinks)
- Active work state (projects, tasks)
- Mix of committed and git-ignored content

### Orchestrator + Worker Pattern

`sow` uses a multi-agent system to manage complexity:

**Orchestrator**
- Your main interface - the agent you interact with
- Handles simple tasks directly
- Coordinates complex work by delegating to specialists
- Compiles relevant context for workers
- Manages project state and progress

**Workers** (Specialists)
- **Architect** - System design and architecture decisions
- **Implementer** - Code implementation with TDD approach
- **Integration Tester** - Cross-component and E2E testing
- **Reviewer** - Code review and refactoring
- **Documenter** - Documentation updates

Each worker receives only the context they need, avoiding information overload.

### Progressive Planning

`sow` rejects waterfall planning in favor of adaptive discovery:

- Start with minimal structure (1-2 phases)
- Add complexity as you discover needs
- Human approval gates for significant changes
- Fail-forward approach (add tasks, don't revert)

### Phases

Work is organized into phases that guide the type of activity:

1. **discovery** - Research and investigation
2. **design** - Architecture and planning
3. **implement** - Building features
4. **test** - Integration and E2E testing
5. **review** - Code quality improvements
6. **deploy** - Deployment preparation
7. **document** - Documentation updates

Not every project uses all phases. Simple work might only use `implement` and `test`.

### Information Sinks

Sinks are collections of markdown files providing focused knowledge:

- Style guides (language conventions)
- API design standards
- Deployment procedures
- Security guidelines
- Company policies

Sinks are:
- Git-versioned repositories
- Installed per-developer (git-ignored)
- Indexed for agent discovery
- Selectively loaded based on task context

### One Project Per Branch

`sow` enforces a strict constraint: **one active project per repository branch**

- Each feature branch has its own project
- Project state committed to feature branches
- Deleted before merging (CI enforced)
- Switch branches = switch project context (automatic via git)
- Simplifies orchestrator logic
- Natural cleanup workflow

---

## Key Terminology

| Term | Definition |
|------|------------|
| **Orchestrator** | Main coordinating agent, user-facing |
| **Worker** | Specialized agent (architect, implementer, etc.) |
| **Phase** | Stage of work (discovery, design, implement, etc.) |
| **Task** | Single unit of work within a phase |
| **Sink** | External knowledge collection (style guides, procedures) |
| **Iteration** | Attempt counter for task completion |
| **Gap numbering** | Task numbering with gaps (010, 020, 030) allowing insertions |
| **Execution layer** | AI behavior (agents, commands, hooks) |
| **Data layer** | Project knowledge and state |
| **Progressive planning** | Start minimal, add structure as needed |
| **Zero-context resumability** | Ability to resume work without conversation history |

---

## Quick Start

### Installation

1. **Install Claude Code Plugin**
   ```bash
   /plugin install sow@sow-marketplace
   ```

2. **Restart Claude Code**
   ```bash
   exit
   claude
   ```

3. **Initialize Your Repository**
   ```bash
   /init
   ```

### Your First Project

1. **Create a feature branch**
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Start a project**
   ```
   /start-project "Add new feature"
   ```

3. **Let the orchestrator work**
   - It will assess complexity
   - Create initial phases and tasks
   - Start delegating to workers

4. **Continue work**
   ```
   /continue
   ```

5. **Clean up before merging**
   ```
   /cleanup
   ```

---

## How It Works

### Simple Tasks (One-Off)

For quick changes, the orchestrator handles everything directly:

```
You: "Fix the typo in README.md line 42"

Orchestrator:
- Reads the file
- Makes the edit
- Done
```

No project structure needed.

### Complex Tasks (Projects)

For multi-step work, the orchestrator creates a project:

```
You: /start-project "Add authentication"

Orchestrator:
1. Assesses complexity → "moderate"
2. Selects phases: design, implement, test
3. Creates initial tasks with gap numbering:
   - design/010: Design auth flow
   - implement/010: Create User model
   - implement/020: Create JWT service
   - implement/030: Add login endpoint
4. Assigns appropriate worker to each task
5. Compiles context for first task
6. Spawns architect worker

Architect:
- Reads requirements
- Reviews relevant sinks (API conventions, security guidelines)
- Creates design document
- Logs work
- Reports completion

Orchestrator:
- Updates project state
- Moves to next task
- Spawns implementer worker

[Process continues...]
```

### Zero-Context Resumability

Stop and resume anytime:

```
You: /continue

Orchestrator:
- Reads .sow/project/state.yaml
- Sees task 020 is in_progress
- Reads task log to understand what's been tried
- Reads any feedback you've provided
- Compiles fresh context
- Spawns worker to continue

Worker:
- Picks up exactly where previous attempt left off
- No conversation history needed
```

---

## Document Navigation

### Understanding the System

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Design decisions, patterns, and architectural philosophy
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Complete directory layout and organization

### Implementing and Configuring

- **[AGENTS.md](./AGENTS.md)** - Agent system, roles, and coordination
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Available slash commands and agent skills
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - Event automation and external integrations

### Using sow

- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows and best practices
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle, phases, tasks, and logging

### Publishing and Maintaining

- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - Packaging, versioning, and upgrade workflows

### Reference

- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command documentation
- **[../research/CLAUDE_CODE.md](../research/CLAUDE_CODE.md)** - Claude Code feature reference

---

## Design Philosophy

### Opinionated but Flexible

`sow` has strong opinions about structure:
- One project per branch (enforced)
- Specific phase vocabulary
- Structured logging format
- Multi-agent architecture

But remains flexible where it matters:
- Progressive planning (not waterfall)
- Extensible through plugins
- Customizable knowledge (sinks)
- Optional CLI enhancements

### Pit of Success

`sow` guides users into good patterns:
- Can't create projects on main branch
- Must clean up before merging (CI enforced)
- Automatic context management
- Structured feedback mechanism

### Not a Replacement

`sow` complements existing tools, doesn't replace them:
- **Not replacing**: JIRA, Linear, GitHub Issues
- **Not replacing**: Sprint planning, roadmaps
- **Not replacing**: Team communication tools

`sow` is for:
- Active work coordination
- AI agent orchestration
- Short-term execution (hours to days)

---

## Common Workflows

### Starting Work

```bash
# Create feature branch
git checkout -b feat/add-auth

# Start project
/start-project "Add authentication"

# Orchestrator plans and begins work
```

### Resuming Work

```bash
# Switch to feature branch
git checkout feat/add-auth

# Resume
/continue

# Orchestrator picks up where you left off
```

### Providing Feedback

```
You: "The JWT service should use RS256, not HS256"

Orchestrator:
- Creates feedback/001.md for current task
- Increments iteration counter
- Spawns worker with feedback context

Worker:
- Reads all feedback
- Incorporates changes
- Marks feedback as addressed
```

### Completing Work

```bash
# All tasks complete
/cleanup

# Orchestrator deletes .sow/project/, stages deletion
# Commit the cleanup
git commit -m "chore: cleanup sow project state"

# Create PR and merge
# CI verifies no .sow/project/ exists
```

---

## Benefits

### For Developers

- **Consistent experience** across projects
- **Clear structure** for complex work
- **Resume anytime** without losing context
- **Specialized expertise** via worker agents
- **Progressive discovery** instead of upfront planning

### For Teams

- **Shared conventions** via committed execution layer
- **Knowledge sharing** via sinks
- **Collaboration** through committed project state on branches
- **Quality gates** through agent specialization
- **Reduced onboarding** with standardized structure

### For AI Agents

- **Focused context** - workers receive only relevant information
- **Clear roles** - each agent has specific expertise
- **Structured state** - can resume without conversation history
- **Fail-forward** - can adapt to discoveries and changes
- **Audit trails** - logs track all actions

---

## Next Steps

1. **Learn the architecture** → Read [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **Install and try it** → Follow [USER_GUIDE.md](./USER_GUIDE.md)
3. **Understand projects** → Review [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)
4. **Explore customization** → Check [HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)

---

## Project Status

This documentation represents the comprehensive design for `sow` based on extensive discovery sessions. Implementation is in progress.

**Current State**: Architecture design complete, implementation pending

**Repository**: https://github.com/your-org/sow (placeholder)

**Feedback**: Issues and discussions welcome
