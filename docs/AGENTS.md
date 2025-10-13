# sow Agent System

**Last Updated**: 2025-10-12
**Purpose**: Comprehensive documentation of the multi-agent system

This document describes how `sow` uses multiple specialized AI agents to coordinate software development work.

---

## Table of Contents

- [Agent System Overview](#agent-system-overview)
- [Orchestrator](#orchestrator)
- [Worker Agents](#worker-agents)
- [Agent File Format](#agent-file-format)
- [Task-Level Assignment](#task-level-assignment)
- [Context Compilation](#context-compilation)
- [Agent Coordination](#agent-coordination)
- [Orchestrator System Prompt](#orchestrator-system-prompt)

---

## Agent System Overview

### Architecture

`sow` uses a hierarchical multi-agent system:

```
              ┌──────────────┐
              │ Orchestrator │ ← User interacts here
              │  (Main Agent)│    (Active Claude Code session)
              └──────┬───────┘
                     │
        ┌────────────┼────────────┬──────────┬──────────┐
        ↓            ↓            ↓          ↓          ↓
  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ ┌──────────┐
  │Architect │ │Implementer│ │Integration│ │Reviewer │ │Documenter│
  │          │ │           │ │  Tester   │ │         │ │          │
  └──────────┘ └──────────┘ └──────────┘ └─────────┘ └──────────┘
       (Specialized workers with focused expertise)
```

### Key Principles

**One Orchestrator**
- User-facing agent
- Main Claude Code session
- Visible to user (not background)
- Coordinates all work

**Multiple Workers**
- Spawned by orchestrator (using Task tool)
- Specialist expertise
- Separate context windows
- Report back to orchestrator

**Explicit Delegation**
- Orchestrator assigns tasks to specific agents
- No automatic delegation
- Agent type stored in task state.yaml
- Predictable, auditable workflow

---

## Orchestrator

### Role

The orchestrator is your main interface to `sow`. It's the agent you see and interact with directly.

### Two Modes

**Mode 1: Direct Execution (One-Off Tasks)**

For simple, trivial tasks:
```
You: "Fix the typo in README.md line 42"

Orchestrator:
- Handles directly (no project needed)
- Reads file
- Makes edit
- Done
```

**Mode 2: Project Coordination (Complex Work)**

For multi-step, non-trivial work:
```
You: /start-project "Add authentication"

Orchestrator:
- Creates project structure
- Plans phases and tasks
- Delegates to workers
- Coordinates completion
```

### Responsibilities

**Startup Behavior**:
1. Check if `.sow/project/` exists
2. If exists: Prompt user to continue
3. If not: Ask what user wants to work on
4. Determine mode: one-off vs. project

**For One-Off Tasks**:
- Handle directly
- Write code
- Make edits
- No project overhead

**For Projects**:
- Read project state (`.sow/project/state.yaml`)
- Compile context from sinks, knowledge, repos
- Decompose work into tasks
- Assign appropriate worker for each task
- Spawn workers with curated context
- Update project state
- Request phase changes (with human approval)
- Manage completion and cleanup

**Does NOT**:
- Write production code for projects (delegates to implementer)
- Perform deep research for projects (delegates to appropriate worker)
- Handle all testing (delegates to integration-tester)

**Exception**: Handles trivial tasks directly without creating projects

### Tools

- Read, Write, Edit (for one-off tasks)
- Grep, Glob (for exploration)
- Bash (for git, builds, tests)
- Task (for spawning workers)

**See Also**: [Orchestrator System Prompt](#orchestrator-system-prompt) below

---

## Worker Agents

### Roster

`sow` includes 5 specialized worker agents:

#### 1. Architect

**Role**: System design and architecture

**Responsibilities**:
- Create design documents
- Write Architecture Decision Records (ADRs)
- Plan system architecture
- Design APIs and interfaces
- Make technology choices
- Define data models

**Skills**:
- `/create-adr` - Create Architecture Decision Record
- `/design-doc` - Write design document

**Typical Phases**: `discovery`, `design`

**Tools**: Read, Write, Grep, Glob

#### 2. Implementer

**Role**: Code implementation with Test-Driven Development (TDD)

**Responsibilities**:
- Implement features using TDD approach
- **Write unit tests FIRST**, then implementation
- Fix bugs (write failing test, then fix)
- Integrate with existing code
- Add dependencies
- Refactor for functionality
- Ensure unit test coverage >90%

**Skills**:
- `/implement-feature` - Implement new feature with TDD
- `/fix-bug` - Fix bug (test first, then fix)

**Typical Phases**: `implement`, `deploy`

**Tools**: Read, Write, Edit, Grep, Glob, Bash

**TDD Enforcement**: Agent prompt requires test-first development
- No implementation without tests
- Red → Green → Refactor cycle
- Unit tests are part of implementation

#### 3. Integration Tester

**Role**: Cross-component and end-to-end testing

**Responsibilities**:
- Write integration tests (multiple components)
- Write end-to-end tests (full user flows)
- Test API contracts between services
- Debug integration failures
- Validate system-wide acceptance criteria
- Test deployment scenarios

**Skills**:
- `/write-integration-tests` - Write integration & E2E tests

**Typical Phases**: `test`

**Tools**: Read, Write, Edit, Grep, Glob, Bash

**Note**: Unit tests handled by implementer during development

**Why Separate?**:
- Implementer focuses on TDD (unit tests)
- Integration tests require different expertise
- Prevents context bloat in implementer
- Clear separation: implement phase vs. test phase

#### 4. Reviewer

**Role**: Code quality and improvement

**Responsibilities**:
- Review code for quality
- Identify security issues
- Suggest refactoring
- Check adherence to standards
- Validate against sinks (style guides, conventions)
- Improve performance
- Enhance readability

**Skills**:
- `/review-code` - Review code for quality

**Typical Phases**: `review`

**Tools**: Read, Grep, Glob, Bash

#### 5. Documenter

**Role**: Documentation maintenance

**Responsibilities**:
- Update README files
- Write inline documentation
- Update API documentation
- Maintain architectural docs
- Add code comments
- Create usage examples

**Skills**:
- `/update-docs` - Update documentation

**Typical Phases**: `document`

**Tools**: Read, Write, Edit, Grep, Glob

### Why Multiple Agents?

**Problem**: Single agent with all capabilities
- Massive system prompt covering all scenarios
- Context explosion (style guides + testing + architecture + deployment)
- Poor performance at each individual task
- Constant compaction/restarts

**Solution**: Focused, specialized agents
- Shorter, more effective prompts
- Each agent has distinct context needs
- Better performance (targeted context)
- Easy to add more (just Markdown files)

---

## Agent File Format

### Structure

Agents are defined as Markdown files with YAML frontmatter:

```markdown
---
name: agent-name
description: "Description of when this agent should be invoked"
tools: Read, Write, Grep  # Optional - inherits all if omitted
model: inherit  # Optional - sonnet, opus, haiku, or inherit
---

# Agent System Prompt

You are a [role description]...

## Your Responsibilities

[Detailed instructions]

## Your Capabilities

[Skills and tools]

[Additional guidance]
```

### Fields

**`name`** (required)
- Agent identifier
- Used when spawning via Task tool
- Lowercase, hyphen-separated
- Example: `architect`, `implementer`

**`description`** (required)
- When this agent should be used
- Helps orchestrator choose appropriate worker
- Brief, action-oriented

**`tools`** (optional)
- Comma-separated list of allowed tools
- If omitted, agent inherits all tools
- Examples: `Read, Write, Grep, Glob, Bash`

**`model`** (optional)
- Which Claude model to use
- Options: `inherit`, `sonnet`, `opus`, `haiku`
- Default: `inherit` (same as orchestrator)

### File Location

`.claude/agents/<agent-name>.md`

### Installation

- Plugin installs all agent files
- Claude Code reads agent descriptions from frontmatter
- Orchestrator references agents when spawning workers

---

## Task-Level Assignment

### Planning Time Assignment

Agents are assigned to tasks during project planning:

```yaml
# .sow/project/state.yaml
phases:
  - name: design
    tasks:
      - id: "010"
        name: Design authentication flow
        assigned_agent: architect  # Decided at planning time

  - name: implement
    tasks:
      - id: "010"
        name: Create User model
        assigned_agent: implementer

      - id: "020"
        name: Write integration tests for user flows
        assigned_agent: integration-tester
```

### Assignment Logic

**During Planning** (orchestrator creates initial plan):

1. **Analyze task requirements**
   - Keywords: "design", "implement", "test", "review", "document"
   - Complexity indicators
   - Phase context (hint, not strict)

2. **Match to agent specialty**
   - "Design X" → architect
   - "Implement Y" → implementer
   - "Test Z" → integration-tester
   - "Review A" → reviewer
   - "Document B" → documenter

3. **Store assignment**
   - Write `assigned_agent` field to state.yaml
   - No ambiguity during execution

### Execution Time

When orchestrator executes task:

1. Read task from state.yaml
2. See `assigned_agent: implementer`
3. Compile context (description, references, feedback)
4. Spawn worker: `Task` tool with agent type + context
5. Worker completes, reports back
6. Orchestrator updates state, moves to next task

### Flexibility

**Can Change Agents**:
- If task requirements change
- If agent gets stuck
- If different expertise needed

**Orchestrator decides**:
- Which agent to use
- When to change agents
- When to increment iteration counter

---

## Context Compilation

### Problem

Workers need minimal, curated context to avoid window bloat.

### Solution

Orchestrator acts as "context compiler" - filters what's relevant.

### Process

**Step 1: Orchestrator Gathers** (broad collection)
```
Orchestrator reads:
- Task requirements (description.md)
- Sink index (.sow/sinks/index.json)
- Knowledge documents (.sow/knowledge/)
- Linked repository files (.sow/repos/)
- Project context (.sow/project/context/)
```

**Step 2: Orchestrator Filters** (selective curation)
```
For THIS specific task, what's relevant?
- Which sinks apply? (Python style guide? API conventions?)
- Which knowledge docs? (Auth design? Database schema?)
- Which code examples? (From linked repos?)
- Which project decisions? (From context/decisions.md?)
```

**Step 3: Orchestrator Packages** (structured handoff)
```
Creates:
- Task description.md (requirements, acceptance criteria)
- Task state.yaml with references list
- All file paths relative to .sow/ root
```

**Step 4: Worker Receives** (focused context)
```
Worker reads:
- state.yaml (iteration, assigned agent, references)
- description.md (what to do)
- All referenced files (sinks, knowledge, code)
- feedback/ (corrections if any)

Worker has everything needed, nothing extra.
```

### Example

**Scenario**: Task to implement JWT authentication service

**Orchestrator Compiles**:
```yaml
# task state.yaml
references:
  - sinks/python-style/conventions.md
  - sinks/api-conventions/rest-standards.md
  - sinks/security-checklist/checklist.md
  - knowledge/architecture/auth-design.md
  - repos/shared-library/src/crypto/jwt.py
```

**Worker (Implementer) Reads**:
1. Description: "Create JWT service with RS256, 1hr expiration, ..."
2. Python conventions (from sink)
3. API standards (from sink)
4. Security checklist (from sink)
5. Auth design doc (from knowledge)
6. Example JWT implementation (from linked repo)

**Result**: Worker has exactly what's needed, nothing more.

### Benefits

1. **Performance**: Workers don't wade through irrelevant info
2. **Accuracy**: Focused context = better results
3. **Efficiency**: No context window bloat
4. **Scalability**: Can handle large knowledge bases

---

## Agent Coordination

### User Experience

**Hybrid Model**:
- **90% of cases**: Orchestrator manages everything transparently
- **Orchestrator = Active Session**: User sits in front of it, sees everything
- **Workers = Spawned in Background**: Return results to orchestrator
- **Translation Layer**: User ↔ Orchestrator ↔ Workers

**User sees**:
```
Orchestrator: "Starting design phase..."
Orchestrator: "Spawning architect for task 010..."
[Worker runs in background]
Orchestrator: "Architect completed design document."
Orchestrator: "Moving to implementation phase..."
```

**Visibility**:
- Normal Claude Code interruption (ESC key) works
- Session history shows orchestrator decisions
- Task logs show worker actions
- Post-facto debugging available

**Escape Hatch**:
- User can invoke workers directly if orchestrator struggles
- Example: `/architect "Design auth flow"`
- Rare, but available

### Error Correction

**User provides feedback to orchestrator**:
```
User: "The JWT service should use RS256, not HS256"

Orchestrator:
1. Creates feedback/001.md for current task
2. Increments iteration counter
3. Spawns worker with feedback context
```

**Worker reads feedback and incorporates**:
```
Worker:
1. Reads feedback/001.md
2. Understands correction
3. Makes changes
4. Updates state: feedback status = addressed
```

### Task Routing

**Simple tasks**: Orchestrator handles directly
- Avoids overhead of spawning worker
- Faster for trivial changes
- No project structure needed

**Complex tasks**: Spawn workers
- Better expertise
- Separate context
- Auditable via logs

**Tradeoff**: Balance context bloat vs. token/latency costs

**Modern context windows** (256k): Simple tasks safe to handle inline

---

## Orchestrator System Prompt

### Complete Prompt

This is the comprehensive system prompt for the orchestrator agent:

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

## Iteration Management

Before spawning worker:
1. Read current iteration from task state.yaml
2. Increment iteration counter
3. Update state.yaml
4. Spawn worker

Worker uses iteration to construct agent ID: `{role}-{iteration}`
Example: `implementer-3` for third attempt by implementer

## Logging

Workers log via CLI:
```
sow log --file <path> --action <action> --result <result> "notes"
```

You log project-level actions to `.sow/project/log.md`

## Error Handling

If worker gets stuck or fails:
1. Review worker's log
2. Determine issue
3. Provide feedback or change approach
4. Increment iteration
5. Spawn worker again (or different agent)

## Completion

When all tasks complete:
1. Verify all phases done
2. Remind user to run /cleanup before merge
3. Confirm readiness for PR

Your role is to orchestrate, not to do all the work yourself.
Delegate to specialists. Manage the big picture.
```

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - Introduction to sow
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Multi-agent architecture rationale
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Slash commands and skills
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Agent file locations
