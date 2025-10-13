# Execution Layer (DRAFT)

**Status**: Draft - Not Authoritative
**Last Updated**: 2025-10-12
**Purpose**: Document the execution layer components and agent orchestration

This document describes the AI-powered execution layer of `sow` - how agents, commands, and automation work together to coordinate software development tasks.

---

## Overview

The execution layer lives in `.claude/` and defines how AI agents behave and interact. It is distributed via Claude Code Plugin and committed to git so teams share the same agent behaviors.

**Core Principle**: Orchestrator + Worker pattern
- **One orchestrator** (user-facing) coordinates work
- **Multiple workers** (specialists) execute tasks
- **Explicit invocation** (orchestrator assigns work, doesn't rely on auto-delegation)

---

## Directory Structure

```
.claude/
├── .claude-plugin/
│   └── plugin.json              # Plugin metadata for distribution
│
├── agents/
│   ├── orchestrator.md          # Main coordinator (user-facing)
│   ├── architect.md             # Design & architecture
│   ├── implementer.md           # Code implementation (TDD enforced)
│   ├── integration-tester.md   # Integration & E2E testing
│   ├── reviewer.md              # Code review & refactoring
│   └── documenter.md            # Documentation updates
│
├── commands/
│   ├── workflows/               # User-invoked workflows
│   │   ├── init.md             # Bootstrap sow
│   │   ├── start-project.md    # Create new project
│   │   ├── continue.md         # Resume project
│   │   ├── cleanup.md          # Delete project
│   │   ├── migrate.md          # Migrate versions
│   │   └── sync.md             # Sync sinks/repos
│   │
│   └── skills/                  # Agent-invoked capabilities
│       ├── architect/
│       │   ├── create-adr.md
│       │   └── design-doc.md
│       ├── implementer/
│       │   ├── implement-feature.md
│       │   └── fix-bug.md
│       ├── tester/
│       │   └── write-tests.md
│       ├── reviewer/
│       │   └── review-code.md
│       └── documenter/
│           └── update-docs.md
│
├── hooks.json                   # Event automation
└── mcp.json                     # External tool integrations (optional)
```

---

## Agents

### How Agents Work

1. **File Format**: Agents are Markdown files (`.md`) with YAML frontmatter
2. **Installation**: Plugin installs all `.claude/agents/*.md` files
3. **Awareness**: Claude Code reads agent descriptions from frontmatter
4. **Invocation**:
   - Orchestrator **explicitly spawns** workers via Task tool
   - Workers assigned at task level in `state.yaml`
   - No reliance on automatic delegation

**Agent File Structure**:
```markdown
---
name: agent-name
description: Description of when this agent should be invoked
tools: Read, Write, Grep  # Optional - inherits all if omitted
model: inherit  # Optional - sonnet, opus, haiku, or inherit
---

Agent's system prompt goes here.
This defines the agent's role, capabilities, and approach.
```

### Agent Roster

**Orchestrator** (Main Coordinator)
- **Role**: User-facing agent, coordinates all work
- **Responsibilities**:
  - **Startup**: Prompt user for work mode (one-off task vs. project)
  - **One-off tasks**: Handle directly, write code for trivial changes
  - **Project work**: Delegate to specialist workers
  - Read project state (`.sow/project/state.yaml`)
  - Compile context from sinks, repos, knowledge
  - Decompose work into tasks
  - Assign appropriate worker for each task
  - Spawn workers with curated context
  - Request phase changes and get user approval
- **Tools**: Read, Write, Edit, Grep, Glob, Bash, Task (for spawning workers)

**Startup Flow**:
```
Orchestrator on session start:

If .sow/project/ exists:
  "Would you like to continue work on '<project-name>'? (y/n)"

If no project:
  "What would you like to work on?"
  1. One-off task (I'll handle it directly)
  2. Start a new project (structured, multi-task work)

  [User chooses option or just describes the work]

  Orchestrator determines: trivial → handle directly
                          complex → /start-project flow
```

**Benefits**:
- Natural conversation flow
- No need for explicit `/trivial` command
- Orchestrator understands sow framework (enhances one-off work)
- Clear separation: one-off (orchestrator) vs. project (workers)

**Architect** (Design & Architecture)
- **Role**: System design, architectural planning
- **Responsibilities**:
  - Create design documents
  - Write Architecture Decision Records (ADRs)
  - Plan system architecture
  - Design APIs and interfaces
  - Make technology choices
- **Skills**: `/create-adr`, `/design-doc`
- **Typical Phases**: discovery, design

**Implementer** (Code Implementation)
- **Role**: Writing production code using Test-Driven Development (TDD)
- **Responsibilities**:
  - Implement features with TDD approach
  - Write unit tests FIRST, then implementation
  - Fix bugs (write failing test, then fix)
  - Integrate with existing code
  - Add dependencies
  - Refactor for functionality
  - Ensure unit test coverage >90%
- **Skills**: `/implement-feature`, `/fix-bug`
- **Typical Phases**: implement, deploy
- **TDD Enforcement**: Agent prompt requires test-first development

**Integration Tester** (Integration & E2E Testing)
- **Role**: Cross-component and end-to-end testing
- **Responsibilities**:
  - Write integration tests (multiple components)
  - Write end-to-end tests (full user flows)
  - Test API contracts between services
  - Debug integration failures
  - Validate system-wide acceptance criteria
- **Skills**: `/write-integration-tests`
- **Typical Phases**: test
- **Note**: Unit tests handled by implementer during development

**Reviewer** (Quality & Refactoring)
- **Role**: Code quality and improvement
- **Responsibilities**:
  - Review code for quality
  - Identify security issues
  - Suggest refactoring
  - Check adherence to standards
  - Validate against sinks (style guides, conventions)
- **Skills**: `/review-code`
- **Typical Phases**: review

**Documenter** (Documentation)
- **Role**: Documentation maintenance
- **Responsibilities**:
  - Update README files
  - Write inline documentation
  - Update API documentation
  - Maintain architectural docs
  - Add code comments
- **Skills**: `/update-docs`
- **Typical Phases**: document

### Why Multiple Agents?

**Problem**: Single agent with all capabilities = context explosion
- Implementer role alone would need: fix bugs, create features, integrate, test, etc.
- Would require massive system prompt covering all scenarios

**Solution**: Focused, specialized agents (6 workers + 1 orchestrator)
- Shorter, more effective prompts
- Each agent has distinct context needs
- Easy to add more (just Markdown files)
- Better performance (targeted context loading)
- TDD enforcement on implementer reduces need for separate unit tester
- Integration tester handles cross-component scenarios only

---

## Task-Level Agent Assignment

### Assignment in State File

Agents are assigned to tasks at planning time:

```yaml
# .sow/project/state.yaml
phases:
  - name: design
    tasks:
      - id: "010"
        name: Design authentication flow
        assigned_agent: architect  # Decided during planning

  - name: implement
    tasks:
      - id: "010"
        name: Create User model
        assigned_agent: implementer

      - id: "020"
        name: Write integration tests for user flows
        assigned_agent: integration-tester

      - id: "030"
        name: Review security implications
        assigned_agent: reviewer
```

### Planning Workflow

**Progressive Planning (Not Waterfall)**

When orchestrator creates initial plan:

**Start Minimal**: Begin with 1-2 phases only
- Don't try to plan entire project upfront
- Prefer discovery as work progresses
- Add phases dynamically when needed

**Example Minimal Start**:
```yaml
# Simple feature - start with just implement
phases:
  - implement

# Complex feature - start with design + implement
phases:
  - design
  - implement
```

**As Work Progresses**:
- Discover need for integration tests → add `test` phase
- Realize deployment complexity → add `deploy` phase
- Need architecture review → add `review` phase

**Task Creation Within Phase**:
1. **Analyze task requirements**
   - Keywords: "design", "implement", "test", "review", "document"
   - Complexity indicators
   - Phase context (hint, not strict)

2. **Assign appropriate agent**
   - Match task type to agent specialty
   - Consider dependencies (design before implement)
   - Allow multiple agent types per phase (flexible)

3. **Store assignment**
   - Write `assigned_agent` field to state.yaml
   - No ambiguity during execution

### Execution Workflow

When orchestrator executes task:

1. Read task from state.yaml
2. See `assigned_agent: implementer`
3. Compile context (description.md, references, feedback)
4. Spawn worker: `Task` tool with agent type + context
5. Worker completes, reports back
6. Orchestrator updates state, moves to next task

**Benefits**:
- Simple execution logic (just read and spawn)
- Flexible (can change agents if needed)
- Explicit (no guessing who does what)
- Auditable (state.yaml shows assignments)

---

## Orchestrator System Prompt

### Core Instructions

```markdown
# .claude/agents/orchestrator.md
---
name: orchestrator
description: "Main coordinator for sow system of work"
tools: Read, Write, Edit, Grep, Glob, Bash, Task
model: inherit
---

You are the orchestrator for the sow system of work.

## Your Role

You coordinate software development work. You can:
- Handle one-off tasks directly (simple changes, quick fixes)
- Coordinate complex work via projects (planning + delegating to workers)

## Available Worker Agents

- **architect**: Design, architecture, ADRs, system planning
- **implementer**: Code implementation with TDD, bug fixes, unit tests
- **integration-tester**: Integration tests, E2E tests, cross-component testing
- **reviewer**: Code review, refactoring, quality checks
- **documenter**: Documentation updates, README, comments

## Startup Behavior

When session starts:
1. Check if `.sow/project/` exists
2. If exists: Prompt user "Continue work on '<project-name>'?"
3. If not exists: Ask user what they want to work on
4. Determine work mode:
   - One-off task → handle directly
   - Complex work → suggest /start-project

## Workflow

### One-Off Tasks (No Project)
- User describes simple change or quick fix
- You handle it directly (write code, make edits)
- No project structure needed
- Benefit: You understand sow framework, enhances quality

### Project-Based Work (Complex, Multi-Task)
1. **Project Start** (/start-project)
   - Check not on main/master branch
   - Check no existing project
   - Rate complexity (1-3)
   - **Start with 1-2 phases** (progressive planning, not waterfall)
     - Available phases: discovery, design, implement, test, review, deploy, document
     - Don't over-plan: add phases as work progresses
   - Break initial phase(s) into tasks with gap numbering (010, 020, 030)
   - Assign agent to each task based on task type
   - Create .sow/project/state.yaml

2. **Project Continuation** (/continue)
   - Read .sow/project/state.yaml
   - Verify branch matches
   - Identify next pending task
   - Compile context for assigned worker
   - Spawn worker with Task tool

3. **Context Compilation** (Before spawning worker)
   - Read task description.md
   - Read sink index (.sow/sinks/index.json)
   - Determine relevant sinks for task
   - Read relevant knowledge docs
   - Read linked repo files if needed
   - Create focused context package
   - Add file references to task state.yaml

4. **Worker Invocation**
   - Use Task tool with assigned agent type
   - Provide: task description, references, feedback
   - Let worker execute independently
   - Update state when worker reports completion

5. **Phase Management**
   - Track current active phase
   - Enforce forward progression (can't skip ahead with incomplete tasks)
   - Allow backward movement (can return to previous phase)
   - Request approval when adding new phases
   - Provide rationale for phase changes

## Key Rules

- **Progressive Planning**: Start with 1-2 phases, add more as work progresses
- **No Waterfall**: Don't try to plan entire project upfront
- Tasks can ONLY be added to current active phase (after initial planning)
- Forward movement requires all current phase tasks complete
- When adding new phase, request user approval with rationale
- Increment task iteration counter before spawning worker
- Use fail-forward: add tasks instead of reverting
- Gap numbering: 010, 020, 030 (allows 011, 012 insertions)
- Mark abandoned tasks, never delete

## Context Management

You are the filter. Workers should receive:
- Minimal, focused context
- Only relevant sinks
- Specific file references
- Clear acceptance criteria

Avoid overwhelming workers with unnecessary information.
```

---

## Skills (Slash Commands)

### Organization by Agent

Skills are organized by the agent that uses them:

```
.claude/commands/skills/
├── architect/
│   ├── create-adr.md        # Create Architecture Decision Record
│   └── design-doc.md        # Write design document
│
├── implementer/
│   ├── implement-feature.md     # Implement new feature with TDD
│   └── fix-bug.md               # Fix bug (write test first, then fix)
│
├── integration-tester/
│   └── write-integration-tests.md  # Write integration & E2E tests
│
├── reviewer/
│   └── review-code.md       # Review code for quality
│
└── documenter/
    └── update-docs.md       # Update documentation
```

### How Skills Work

1. **Definition**: Skills are slash commands (`.md` files)
2. **Reference**: Agents reference skills in their system prompts
3. **Invocation**: Agents call skills when appropriate

**Example - Architect Agent**:
```markdown
# .claude/agents/architect.md
---
name: architect
description: "System design and architecture specialist"
tools: Read, Write, Grep, Glob
model: inherit
---

You are a system architect specializing in design and planning.

## Your Capabilities

When you need to create an Architecture Decision Record:
Use the /create-adr skill

When you need to write a design document:
Use the /design-doc skill

[Additional instructions...]
```

**Benefits**:
- Keeps agent prompts concise (references, not full instructions)
- Skills can be 100+ lines without bloating agent context
- Skills are reusable across agents if needed
- Easy to add new skills (just new .md files)

---

## Workflow Commands

User-invoked commands for managing sow lifecycle:

### /init
**Purpose**: Bootstrap sow in repository

**Actions**:
1. Check if already initialized
2. Create `.claude/` structure (agents, commands, hooks)
3. Create `.sow/` structure (knowledge, sinks, repos folders)
4. Create initial `.sow/knowledge/overview.md`
5. Commit to git

### /start-project <name>
**Purpose**: Create new project with planning

**Smart UX Flow**:

**Case 1: On default branch (main/master)**
```
Error: Cannot create project on main/master branch

Would you like to create a new feature branch?
  Branch name suggestion: feat/<project-name>

[y/n] → If yes: git checkout -b feat/<name>, then create project
```

**Case 2: On feature branch, no existing project**
```
Creating project '<name>' on branch '<current-branch>'...
[Proceed with planning]
```

**Case 3: On feature branch WITH existing project**
```
A project already exists: '<existing-project-name>'

Options:
1. Continue with existing project (/continue)
2. Create new branch and start fresh project

[User chooses option]
```

**Actions** (after validation):
1. Prompt user for project description
2. Assess complexity (1-3 scale)
3. **Select initial phases** (1-2 phases, minimal start)
   - Simple: just `implement`
   - Moderate: `design` + `implement`
   - Complex: `discovery` + `design` or `design` + `implement`
   - Don't select all phases upfront (add dynamically later)
4. Create initial tasks for selected phases with agent assignments
5. Create `.sow/project/state.yaml`
6. Commit to git

**Progressive Planning Philosophy**:
- Start minimal, discover as you go
- Add phases when actually needed, not speculatively
- Request user approval when adding new phases
- Avoid waterfall planning (all phases upfront)

### /continue
**Purpose**: Resume existing project

**Actions**:
1. Read `.sow/project/state.yaml`
2. Verify branch matches
3. Show current status (phases, tasks, progress)
4. Identify next pending task
5. Spawn appropriate worker

### /cleanup
**Purpose**: Delete project, prepare for merge

**Actions**:
1. Check all tasks complete (warn if not)
2. Delete `.sow/project/` directory
3. Stage deletion (`git rm -rf .sow/project/`)
4. Prompt user to commit cleanup
5. Remind about CI check

### /migrate
**Purpose**: Migrate sow versions

**Actions**:
1. Detect current sow version
2. Apply migrations to `.claude/` structure
3. Apply migrations to `.sow/` data if needed
4. Update version marker
5. Commit changes

### /sync
**Purpose**: Sync sinks and repos

**Actions**:
1. Check for sink updates (poll git remotes)
2. Check for repo updates
3. Show available updates
4. Prompt for selective updates
5. Pull updates and regenerate indexes

---

## Hooks

Event-driven automation defined in `.claude/hooks.json`.

### SessionStart Hook (MVP)

**Purpose**: Provide context when Claude Code starts

**Configuration**:
```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  }
}
```

**CLI Command** (`sow session-info`):
```bash
#!/bin/bash

if [ -d ".sow" ]; then
  echo "📋 You are in a sow-enabled repository"

  if [ -d ".sow/project" ]; then
    PROJECT_NAME=$(yq .project.name .sow/project/state.yaml)
    BRANCH=$(git branch --show-current)
    echo "🚀 Active project: $PROJECT_NAME (branch: $BRANCH)"
    echo "📂 Use /continue to resume work"
  else
    echo "💡 No active project. Use /start-project <name> to begin"
  fi

  echo ""
  echo "📖 Available commands:"
  echo "   /start-project <name> - Create new project"
  echo "   /continue - Resume existing project"
  echo "   /cleanup - Delete project before merge"
else
  echo "⚠️  Not a sow repository"
  echo "💡 Use /init to set up sow"
fi
```

**Benefits**:
- Immediate context on session start
- Shows current project state
- Reminds user of available commands
- Low overhead (just read files)

### Other Hooks (Deferred)

**PostToolUse Hooks** (Auto-formatting):
- Auto-format code after Edit
- Validate YAML after Write
- Run linters automatically

**PreCompact Hook**:
- Save important context before compaction
- Log conversation summary

**Defer to post-MVP**: Add complexity, not critical for core workflow

---

## MCP Integrations (Optional)

Model Context Protocol servers for external tool integration.

**Potential Integrations**:
- GitHub: Read issues, create PRs, comment on reviews
- Jira/Linear: Sync tasks, update status
- Monitoring: Query logs, check alerts
- Documentation: Search Confluence, Notion

**Configuration**: `.claude/mcp.json`

**Decision**: Defer to post-MVP
- Not required for core workflow
- Adds complexity
- Requires authentication setup
- Better as optional enhancement

---

## Plugin Metadata

**File**: `.claude-plugin/plugin.json`

**Purpose**: Metadata for Claude Code plugin distribution

**Example**:
```json
{
  "name": "sow",
  "version": "0.1.0",
  "description": "AI-powered system of work for software engineering",
  "author": "sow contributors",
  "repository": "https://github.com/yourusername/sow"
}
```

**Distribution**:
- Users install via Claude Code plugin system
- Plugin bundles all `.claude/` contents
- Automatic updates when new versions released
- Can be installed from marketplace or git URL

---

## Execution Flow Example

### Complete Task Execution

```
1. User: /continue

2. Orchestrator:
   - Reads .sow/project/state.yaml
   - Finds task 020 (implement JWT service)
   - Sees assigned_agent: implementer
   - Reads .sow/project/phases/implement/tasks/020/description.md
   - Reads .sow/sinks/index.json
   - Identifies relevant sinks:
     * .sow/sinks/python-style/conventions.md
     * .sow/sinks/api-conventions/rest-standards.md
   - Reads .sow/knowledge/architecture/auth-design.md
   - Updates task state.yaml:
     * Increments iteration: 3
     * Adds references list
     * Sets status: in_progress

3. Orchestrator spawns implementer:
   Task tool with:
   - Agent: implementer
   - Context: description.md + references
   - Instructions: "Read description, references, complete task, log actions"

4. Implementer (worker):
   - Reads task state.yaml (sees iteration=3, references)
   - Reads description.md (requirements, acceptance criteria)
   - Reads referenced files (sinks, knowledge docs)
   - Implements JWT service
   - Logs via CLI: sow log --file src/auth/jwt.py --action created_file --result success "Created JWT service"
   - Updates task state.yaml: status=completed, files_modified=[...]
   - Reports back to orchestrator

5. Orchestrator:
   - Receives completion report
   - Updates project state.yaml (task 020 completed)
   - Identifies next task 030
   - Asks user if ready to continue or pause

6. User: continue

7. Repeat for task 030...
```

---

## Open Questions

1. **Agent Tool Permissions**: Should we restrict tools per agent? (e.g., tester can't modify production code?)
2. **Worker-to-Worker Communication**: Can workers spawn other workers? Or only orchestrator?
3. **Skill Discovery**: How do agents know which skills exist? Read directory? Hardcode?
4. **Error Recovery**: What happens if worker fails/gets stuck? Auto-increment iteration and retry?
5. **Concurrent Workers**: Orchestrator spawns multiple workers for parallel tasks - how to coordinate?

---

## Related Documentation

- **`FS_STRUCTURE.md`**: Complete filesystem layout
- **`PROJECT_LIFECYCLE.md`**: Operational workflows and project management
- **`BRAINSTORMING.md`**: Design exploration and decisions
