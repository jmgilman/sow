# sow Architecture

**Last Updated**: 2025-10-12
**Purpose**: Comprehensive architectural design and decision rationale

This document explains the fundamental architecture of `sow`, the design patterns employed, and the rationale behind key decisions.

---

## Table of Contents

- [Two-Layer Architecture](#two-layer-architecture)
- [Multi-Agent System](#multi-agent-system)
- [Progressive Planning Philosophy](#progressive-planning-philosophy)
- [Information Sinks](#information-sinks)
- [Single Project Constraint](#single-project-constraint)
- [Multi-Repo Strategy](#multi-repo-strategy)
- [Zero-Context Resumability](#zero-context-resumability)
- [Key Design Decisions](#key-design-decisions)

---

## Two-Layer Architecture

### Concept

`sow` separates behavior (execution) from state (data) through a fundamental architectural division:

```
┌─────────────────────────────────────────────┐
│          EXECUTION LAYER (.claude/)         │
│  Agents | Commands | Hooks | Integrations   │
│         (Behavior & Logic)                  │
└─────────────────────────────────────────────┘
                     ↓ uses
┌─────────────────────────────────────────────┐
│            DATA LAYER (.sow/)               │
│  Knowledge | Sinks | Repos | Projects       │
│         (State & Context)                   │
└─────────────────────────────────────────────┘
```

### Rationale

**Why separate layers?**

1. **Independent Evolution**
   - Execution layer can be upgraded wholesale
   - Data layer persists and evolves separately
   - Migrations bridge gaps when structure changes

2. **Distribution Model**
   - Behavior distributed via plugin (Claude Code Plugin)
   - State lives in repository (version-controlled with code)
   - Teams share execution layer, customize data layer

3. **Versioning Strategy**
   - Plugin uses semantic versioning
   - Repository structure tracked independently
   - Clear upgrade path via migrations

4. **Clear Responsibilities**
   - Execution: "How should AI agents behave?"
   - Data: "What does this project need to know?"

### Implementation

**Execution Layer** (`.claude/`)
- **Source**: Developed in `plugin/` directory of marketplace repository
- **Installation**: Contents of `plugin/` copied to `.claude/` when plugin installed
- **Distribution**: Via Claude Code Plugin marketplace
- Committed to git (all branches) after installation
- Contains: agents, commands, hooks, MCP integrations
- Immutable during session (replaced on upgrade)

**Development Note**: When building the `sow` plugin, edit files in `plugin/`. When users install the plugin, these files appear in their repository as `.claude/`.

**Data Layer** (`.sow/`)
- Mixed git strategy (some committed, some ignored)
- Repository-specific
- Contains: knowledge, sinks, repos, project state
- Mutable during session (agents read/write)

**See Also**: [FILE_STRUCTURE.md](./FILE_STRUCTURE.md) for complete directory layout

---

## Multi-Agent System

### Orchestrator + Worker Pattern

`sow` uses a hierarchical multi-agent architecture:

```
              ┌──────────────┐
              │ Orchestrator │ ← User interacts here
              └──────┬───────┘
                     │ spawns & coordinates
        ┌────────────┼────────────┬──────────┐
        ↓            ↓            ↓          ↓
  ┌──────────┐ ┌──────────┐ ┌─────────┐ ┌──────────┐
  │Architect │ │Implementer│ │ Reviewer│ │Documenter│
  └──────────┘ └──────────┘ └─────────┘ └──────────┘
     (Workers with specialized expertise)
```

### Roles

**Orchestrator**
- User-facing agent (your main interface)
- Handles trivial tasks directly
- Coordinates complex work via delegation
- Compiles context for workers
- Manages project state
- Does NOT write production code for projects

**Workers**
- Specialist agents with focused expertise
- Receive bounded context from orchestrator
- Execute specific tasks independently
- Report results back to orchestrator
- Examples: architect, implementer, integration-tester, reviewer, documenter

### Rationale

**Why multiple agents instead of one super-agent?**

1. **Context Management**
   - Single agent with all capabilities = massive system prompt
   - Specialized agents = focused, effective prompts
   - Workers receive only relevant context (no bloat)

2. **Separation of Concerns**
   - Each agent has distinct role and expertise
   - Clear boundaries prevent confusion
   - Better at specialized tasks than generalist

3. **Scalability**
   - Easy to add new worker types (just Markdown files)
   - Can spawn multiple workers in parallel
   - Independent context windows

4. **Performance**
   - Smaller prompts = faster responses
   - Targeted context loading
   - Avoid constant compaction/restarts

**Why orchestrator doesn't code?**

- Clear separation: planning vs. execution
- Orchestrator maintains high-level view
- Workers dive deep into implementation
- Prevents orchestrator context bloat
- Exception: Trivial one-off tasks (no project needed)

### Context Compilation

Orchestrator acts as "context compiler":

```
1. Orchestrator reads:
   - Project requirements
   - Sink index (.sow/sinks/index.json)
   - Knowledge documents
   - Linked repository files

2. Orchestrator filters:
   - What's relevant for THIS task?
   - Which sinks apply?
   - What code examples help?

3. Orchestrator packages:
   - Creates task description.md
   - Lists file references in state.yaml
   - Provides acceptance criteria

4. Worker receives:
   - Minimal, targeted context
   - Just what's needed for the task
   - No information overload
```

**Benefit**: Workers start immediately with right information, no thrashing.

**See Also**: [AGENTS.md](./AGENTS.md) for detailed agent documentation

---

## Progressive Planning Philosophy

### Not Waterfall

`sow` explicitly rejects upfront comprehensive planning in favor of adaptive discovery.

**Traditional Waterfall Approach** (Rejected):
```
1. Plan entire project upfront
2. Define all phases and tasks
3. Execute linearly
4. Discover problems late
5. Expensive changes
```

**Progressive Planning** (sow Approach):
```
1. Start with minimal structure (1-2 phases)
2. Begin work immediately
3. Discover needs as you go
4. Add phases/tasks dynamically
5. Human approval gates for changes
```

### Rationale

**Why progressive over waterfall?**

1. **Unknown Unknowns**
   - Can't predict all issues upfront
   - Discovery happens during implementation
   - Better to adapt than force original plan

2. **Reduced Overhead**
   - Don't spend time planning things that might change
   - Focus on immediate next steps
   - Add structure when actually needed

3. **Flexibility**
   - Can pivot based on discoveries
   - No sunk cost in detailed plan
   - Fail-forward approach

4. **Human Oversight**
   - Approval gates when adding phases
   - User validates direction changes
   - Prevents runaway planning

### Implementation

**Minimal Start**:
```yaml
# Simple task
phases:
  - implement

# Moderate complexity
phases:
  - design
  - implement

# Higher complexity
phases:
  - discovery
  - design
```

**Dynamic Addition**:
```
Orchestrator: "I've discovered a race condition. Need to add
'discovery' phase to investigate. Approve? (y/n)"

User: y

Orchestrator: Creates discovery phase, adds investigation task
```

**Phase Ordering Rules**:
- Can move backward to previous phases
- Cannot skip forward with incomplete tasks
- Must complete current phase before advancing
- Can add tasks to earlier phases

**See Also**: [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md) for phase details

---

## Information Sinks

### Concept

Sinks are granular, composable knowledge units that provide focused context.

**Definition**: A "sink" is a collection of one or more Markdown files providing context on a specific topic.

**Examples**:
- Language style guide (Python conventions, Go patterns)
- Deployment process
- API design standards
- Security checklist
- Company code policy

### Architecture

```
User Level:
~/.config/sow/sinks/
└── index.json          # Global sink registry

Project Level:
.sow/sinks/             # Git-ignored
├── index.json          # LLM-maintained catalog
├── python-style/
│   ├── formatting.md
│   └── conventions.md
├── api-conventions/
│   └── rest-standards.md
└── deployment-guide/
    └── procedures.md
```

### Mechanics

**Installation**:
```bash
# Via CLI
sow sinks install git@github.com:company/style-guides.git python-style/

# Via slash command
/sync
```

**Format**: Git repository URL + folder path

**Storage**:
- Installed into `.sow/sinks/` (git-ignored)
- Each developer controls their own sinks
- Can install from multiple sources

**Indexing**:
- Orchestrator interrogates sink on install
- LLM creates summary and "when to use" guidance
- Stored in `.sow/sinks/index.json`
- Index used for routing context to workers

### Rationale

**Why sinks instead of committing docs?**

1. **Granular Control**
   - Install only what you need
   - Different projects need different context
   - Per-developer flexibility

2. **Composability**
   - Mix from multiple sources
   - Company standards + language guides + team practices
   - Not all-or-nothing

3. **Versioning**
   - Sinks are git repositories
   - Can pin to specific commits/tags
   - Update when ready (like package managers)

4. **Discovery**
   - Index helps agents find relevant context
   - "When should you use this?"
   - Orchestrator routes to workers

5. **Context Management**
   - Workers receive only applicable sinks
   - Prevents information overload
   - Targeted, relevant context

**Why git-ignored?**

- External content (don't pollute repo)
- Per-developer installations
- Can update independently
- Large repositories (don't bloat project)

### Update Strategy

```bash
# Check for updates
sow sinks update

# Polls git remotes for changes
# Shows available updates
# Selective or batch update
```

Similar to `npm update`, `pip update`, etc.

**See Also**: [USER_GUIDE.md](./USER_GUIDE.md#working-with-sinks) for usage

---

## Single Project Constraint

### Rule

**One project per repository branch** (enforced, not suggested)

**Implementation**:
- Only `.sow/project/` exists (singular, not plural)
- Cannot create project on `main`/`master`
- Errors if project exists when starting new one
- Git tracks branch name in project state

### Rationale

**Why enforce this constraint?**

1. **Simplicity**
   - Orchestrator always knows current project
   - No "which project?" logic needed
   - Commands like `/continue` need no arguments

2. **Git Integration**
   - One branch = one project (natural mental model)
   - Switch branches = switch projects automatically
   - Leverages git's branching model

3. **Cleanup**
   - Delete branch = project state gone
   - CI enforces cleanup before merge
   - Clear lifecycle: branch creation → work → merge → deletion

4. **Filesystem**
   - Cleaner structure (no `projects/` directory)
   - Always `.sow/project/state.yaml`
   - Predictable, simple paths

5. **Team Collaboration**
   - Project state committed to feature branches
   - Push branch = share project context
   - Others can pull and see full state

### Git Versioning Strategy

**Critical Decision**: `.sow/project/` is **committed** to feature branches

**Problem with git-ignore**:
```
Branch A: Create project → git-ignored folder created
Switch to Branch B: Same folder still there (wrong context!)
```

**Solution**:
```
Branch A: Project committed to branch
Switch to Branch B: Git switches project automatically
```

**Workflow**:
1. Feature branch: `.sow/project/` committed normally
2. Push branch: Team can pull and see project state
3. Ready to merge: `/cleanup` deletes `.sow/project/`
4. CI check: Fails if `.sow/project/` exists
5. Squash merge: Main history stays clean

**Benefits**:
- Natural git behavior (switch = context switch)
- Team collaboration (shared project state)
- Backup (state pushed to remote)
- CI safety net (can't merge accidentally)
- Clean main history (squash merges)

**See Also**: [USER_GUIDE.md](./USER_GUIDE.md#project-lifecycle) for workflows

---

## Multi-Repo Strategy

### Problem

Agents working in one repository often need context from other repositories:
- Microservices with interdependencies
- Shared libraries and utilities
- Cross-cutting concerns (deployment, releases)
- Related services (auth service needs user service context)

### Solution: Hybrid Approach

```
.sow/
├── sinks/     # ALWAYS present (focused knowledge)
└── repos/     # OPTIONAL (full repository context)
```

**Sinks** for granular knowledge (style guides, procedures)
**Repos** for full codebase context (linked repositories)

### Repository Linking

**Mechanics**:
```bash
# Add linked repository
sow repos add git@github.com:company/auth-service.git

# Clones or symlinks to:
.sow/repos/auth-service/
```

**Index**: `.sow/repos/index.json`
- Repository metadata
- Git URLs and references
- Purpose notes
- Discovery hints for agents

**Use Case**: Turn multi-repo setup into pseudo-monorepo for agent context

**Git-Ignored**: Repositories are external, too large to commit

### Monorepo Consideration

**True Monorepo**: Never uses `.sow/repos/`
- All code in same repository
- Standard project structure
- Just uses `.sow/sinks/` for external knowledge

**Multi-Repo**: Uses `.sow/repos/` to bring context together
- Checkout related repos
- Agents can read across boundaries
- Orchestrator includes in context compilation

---

## Zero-Context Resumability

### Principle

Any agent should be able to resume any project/task from scratch without conversation history.

### Implementation

All necessary context stored in filesystem:

**Project Level** (`.sow/project/`):
- `state.yaml` - What phases and tasks exist, their status
- `log.md` - Chronological history of orchestrator actions
- `context/` - Project-specific decisions and memories

**Task Level** (`phases/<phase>/tasks/<id>/`):
- `state.yaml` - Task metadata, iteration, references
- `log.md` - Chronological history of worker actions
- `description.md` - What needs to be done
- `feedback/` - Human corrections and guidance

### Recovery Process

**Orchestrator Recovery**:
```
1. Read .sow/project/state.yaml
   → What phases and tasks exist?
   → What's complete, what's pending?
   → What branch is this?

2. Read .sow/project/log.md
   → What actions were taken?
   → What decisions were made?

3. Verify branch matches (safeguard)

4. Determine next action
   → Resume incomplete task?
   → Start new task?

5. Compile context for worker

6. Spawn worker with instructions
```

**Worker Recovery**:
```
1. Read task state.yaml
   → What iteration is this?
   → What references should I read?
   → What feedback exists?

2. Read task description.md
   → What am I supposed to do?
   → What are the requirements?

3. Read referenced files
   → Sinks (style guides, conventions)
   → Knowledge (design docs, ADRs)
   → Code examples (from linked repos)

4. Read task log.md
   → What's already been tried?
   → What worked, what didn't?

5. Read feedback/*.md
   → What corrections needed?
   → What should change?

6. Continue work from current state
```

### Benefits

1. **Session Independence**
   - No reliance on conversation history
   - Can pause and resume across days/weeks
   - Context window limits don't matter

2. **Multi-Developer**
   - Different developers can work on same project
   - Natural handoff mechanism
   - Shared understanding via filesystem

3. **Multi-Agent**
   - Different agents can resume tasks
   - No "memory" required
   - Everything needed is on disk

4. **Transparency**
   - Complete audit trail
   - Can review decisions and actions
   - Debugging and learning

5. **Reliability**
   - No lost context from session crashes
   - No "I forgot what we discussed"
   - Deterministic resumption

---

## Key Design Decisions

### Phases as Directories

**Decision**: Phases are first-class directories, not just YAML entries

**Structure**:
```
.sow/project/phases/
├── discovery/
│   └── tasks/
├── design/
│   └── tasks/
└── implement/
    └── tasks/
```

**Rationale**:
- Filesystem discoverability (`ls` shows phases)
- No YAML cross-referencing needed
- Supports zero-context resumability
- Clear visual organization
- Easy archival (`mv phase-01 archive/`)

**Alternative Considered**: All phases in single state.yaml with task references
- Rejected: Requires parsing YAML to discover structure
- Rejected: Makes filesystem less explorable

---

### Skills = Slash Commands

**Decision**: No separate "skills" system, just reference slash commands

**Implementation**:
```markdown
# .claude/agents/architect.md

When you need to create an Architecture Decision Record:
Use the /create-adr skill
```

**Rationale**:
- Avoids new abstraction layer
- Leverages existing Claude Code feature
- Prevents context window bloat (reference, not copy)
- Maintains composability
- Slash commands already support subdirectories

**Alternative Considered**: Separate skills system with custom format
- Rejected: Duplicates slash command functionality
- Rejected: Adds complexity without benefit

---

### Fail-Forward Task Management

**Decision**: Never delete tasks, mark as abandoned

**Gap Numbering**: 010, 020, 030, 040
- Allows insertions: 011, 012, 021
- No renumbering chaos
- Similar to database migrations

**Abandoned State**:
```yaml
tasks:
  - id: "010"
    name: Change timeout
    status: abandoned  # Not deleted
  - id: "011"
    name: Investigate race condition  # New task
    status: completed
```

**Rationale**:
- Preserves audit trail
- Can see what was attempted
- Learning from failures
- No information loss

**Alternative Considered**: Delete and renumber tasks
- Rejected: Loses history
- Rejected: Breaks references
- Rejected: No failure visibility

---

### CLI-Driven Logging

**Decision**: Agents use CLI command for logging, not direct file editing

**Implementation**:
```bash
sow log \
  --file src/auth/jwt.py \
  --action modified_file \
  --result success \
  "Implemented token generation"
```

**Rationale**:
- **Performance**: Direct file editing is slow (30+ seconds)
- **Consistency**: CLI enforces format automatically
- **Simplicity**: Single bash command
- **Accuracy**: Auto-constructs agent ID from iteration counter

**Alternative Considered**: Agents directly edit log.md
- Rejected: Too slow (performance critical)
- Rejected: Format inconsistencies
- Rejected: Agent cognitive overhead

---

### Iteration Counter

**Decision**: Track task attempts via iteration counter (not agent history)

**Implementation**:
```yaml
# task state.yaml
task:
  iteration: 3  # Third attempt
  assigned_agent: implementer
```

**Agent ID Construction**: `{role}-{iteration}` → `implementer-3`

**Rationale**:
- Clear tracking of attempts
- Visible in logs which iteration did what
- Orchestrator manages (not worker)
- Enables retry logic
- Helps with feedback tracking

**When to Increment**:
- Worker gets stuck, task paused → increment
- Feedback added, resume → increment
- Worker completes → no increment (same attempt)

---

### Separate Test Agents

**Decision**: Two testing agents: implementer (unit tests) + integration-tester (integration/E2E)

**Rationale**:
- Implementer enforces TDD (write unit tests first)
- Integration tests are different skill set
- Unit tests = part of implementation process
- Integration tests = separate verification phase
- Prevents context bloat in implementer

**Alternative Considered**: Single tester agent for all testing
- Rejected: Implementer would skip unit tests
- Rejected: Different expertise needed
- Rejected: Phase confusion (implement vs. test)

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - Introduction to sow
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Complete directory layout
- **[AGENTS.md](./AGENTS.md)** - Agent system details
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Usage workflows
