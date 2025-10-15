# sow Architecture

**Last Updated**: 2025-10-15
**Purpose**: Comprehensive architectural design and decision rationale

This document explains the fundamental architecture of `sow`, the design patterns employed, and the rationale behind key decisions.

---

## Table of Contents

- [Two-Layer Architecture](#two-layer-architecture)
- [Multi-Agent System](#multi-agent-system)
- [Human-AI Collaboration Model](#human-ai-collaboration-model)
- [External References System](#external-references-system)
- [Single Project Constraint](#single-project-constraint)
- [Zero-Context Resumability](#zero-context-resumability)
- [Key Design Decisions](#key-design-decisions)
- [Related Documentation](#related-documentation)

---

## Two-Layer Architecture

### Concept

`sow` separates behavior (execution) from state (data): execution layer (`.claude/`) contains agents, commands, hooks, and integrations; data layer (`.sow/`) contains knowledge, refs, and projects.

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
- Contains: knowledge, refs, project state
- Mutable during session (agents read/write)

**See Also**: [FILE_STRUCTURE.md](./FILE_STRUCTURE.md) for complete directory layout

---

## Multi-Agent System

### Orchestrator + Worker Pattern

`sow` uses a hierarchical multi-agent architecture where the orchestrator spawns and coordinates specialized workers (Researcher, Architect, Planner, Implementer, Reviewer, Documenter). Users interact with the orchestrator, which delegates to workers.

### Roles

**Orchestrator**
- User-facing agent (your main interface)
- **Two modes**:
  - **Subservient Mode** (Discovery, Design): Acts as assistant, human leads
  - **Autonomous Mode** (Implementation, Review, Finalize): Executes independently
- Handles trivial tasks directly (no project needed)
- Coordinates complex work via delegation
- Compiles context for workers
- Manages project state
- Does NOT write production code for projects

**Workers**
- Specialist agents with focused expertise
- Receive bounded context from orchestrator
- Execute specific tasks independently
- Report results back to orchestrator
- **New in v2**: Researcher (discovery), Planner (implementation)

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

Orchestrator acts as context compiler: reads project requirements, refs index, knowledge documents, and task descriptions; filters for task-relevant content; packages into task description.md with file references in state.yaml. Workers receive minimal, targeted context and start immediately without thrashing.

**See Also**: [AGENTS.md](./AGENTS.md) for detailed agent documentation

---

## Human-AI Collaboration Model

### Philosophy

`sow` v2 recognizes that humans and AI have complementary strengths, and the architecture reflects this reality.

**Humans excel at**:
- Scoping and constraint-setting
- Architectural judgment
- Knowing when "good enough" is sufficient
- Making strategic decisions

**AI excels at**:
- Execution within well-defined boundaries
- Systematic implementation
- Following established patterns
- Tireless iteration and refinement

### Orchestrator Mode Switching

The orchestrator changes behavior based on which phase is active:

#### Subservient Mode (Discovery, Design)

**When**: Human-led planning phases

**Behavior**:
- Acts as assistant to human
- Asks clarifying questions
- Points out inconsistencies
- Makes suggestions
- Takes notes continuously
- **Never makes unilateral decisions**

#### Autonomous Mode (Implementation, Review, Finalize)

**When**: AI-led execution phases

**Behavior**:
- Executes tasks independently
- Makes implementation decisions within boundaries
- Only requests approval for adding new tasks, going back to previous phases, or blocking issues
- Updates state automatically

### Phase-Based Collaboration Pattern

Discovery (subservient, human approval) → Design (subservient, human approval) → Implementation (autonomous, automatic transition) → Review (autonomous, human approval) → Finalize (autonomous, PR created). Transition from subservient to autonomous happens when planning is complete and boundaries are clear.

### Rationale

**Why not fully autonomous?**
- AI struggles with unbounded problem spaces
- Humans are better at scoping and constraints
- Planning requires judgment, not just execution
- Architecture decisions need human authority

**Why not fully manual?**
- Humans don't want to micromanage implementation
- AI is excellent at systematic execution
- Reduces cognitive load on developers
- Faster execution with clear boundaries

**The hybrid model leverages both strengths**.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md) for how phases enable this model

---

## External References System

### Concept

The refs system provides a unified approach to external knowledge and code references, replacing the previous "sinks" and "repos" systems. Remote repositories are cloned once to `~/.cache/sow/repos/` and symlinked (or copied on Windows) to `.sow/refs/`. Two-index system: committed index (`.sow/refs/index.json`) for categorical metadata, cache index (`~/.cache/sow/index.json`) for transient data.

### Architecture

**Local Caching**:
- Clone remote repositories once per machine
- Store in `~/.cache/sow/repos/`
- Symlink (Unix) or copy (Windows) to `.sow/refs/`
- One cache serves all local repositories

**Two-Index System**:
- **Committed Index** (`.sow/refs/index.json`): Categorical metadata, shared with team
- **Cache Index** (`~/.cache/sow/index.json`): Transient data (SHAs, timestamps), per-machine

**Benefits**:
- Efficient disk usage (clone once, use everywhere)
- Team sharing (committed index)
- Subpath support (without sparse checkouts)
- Automatic staleness detection

### Types of References

**Knowledge Refs** (`--type knowledge`):
- Style guides, conventions, policies, standards
- Consulted when making decisions
- Example: Python style guide, API design standards

**Code Refs** (`--type code`):
- Implementation examples, patterns, reference code
- Studied and adapted (not imported)
- Example: Auth service implementation, crypto utilities

### Rationale

**Why unified system instead of separate sinks/repos?**

1. **Consistency**
   - Same mechanism for all external content
   - Single CLI command set
   - Unified caching strategy

2. **Simplification**
   - Easier mental model
   - Fewer concepts to learn
   - Less configuration complexity

3. **Flexibility**
   - Same repo can provide both knowledge and code refs
   - Subpath support enables granular referencing
   - Type distinction is semantic, not structural

**Why cache locally instead of direct git operations?**

1. **Subpath Support**
   - Can reference `python/` and `go/` from same repo without sparse checkout
   - Symlink directly to subdirectory

2. **Efficiency**
   - Clone once, use in multiple projects
   - Updates propagate automatically (via symlinks on Unix)

3. **Offline Work**
   - Cache persists locally
   - No network required after initial clone

**See Also**: [REFS.md](./REFS.md) for complete refs system documentation

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
   - Commands like `/project:continue` need no arguments

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

`.sow/project/` is committed to feature branches. Git-ignoring would cause context conflicts when switching branches (folder persists with wrong context). Committing enables natural git behavior where branch switching automatically switches project context. Feature branches commit project state normally. Finalize phase deletes project folder before merge. CI enforces cleanup. Squash merges keep main history clean. Benefits: automatic context switching, team collaboration, backup, CI safety net.

**See Also**: [PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md#completion) for cleanup workflow

---

## Zero-Context Resumability

### Principle

Any agent should be able to resume any project/task from scratch without conversation history.

### Implementation

All necessary context stored in filesystem:

**Project Level** (`.sow/project/`):
- `state.yaml` - What phases are enabled, which tasks exist, their status
- `log.md` - Chronological history of orchestrator actions
- `context/` - Project-specific decisions and memories

**Task Level** (`phases/implementation/tasks/<id>/`):
- `state.yaml` - Task metadata, iteration, references
- `log.md` - Chronological history of implementer actions
- `description.md` - What needs to be done
- `feedback/` - Human corrections and guidance

**Phase Level** (varies by phase):
- `phases/discovery/notes.md` - Discovery findings
- `phases/design/adrs/` - Architecture decisions
- `phases/review/reports/` - Review iteration reports

### Recovery Process

**Orchestrator**: Reads project state.yaml and log.md to determine which phases are enabled, their status, and next action. Verifies branch matches, compiles context, spawns worker.

**Worker**: Reads task state.yaml, description.md, referenced files (refs, knowledge), task log.md, and feedback. Continues work from current state based on discovered context.

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

**See Also**: [LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md) for complete state management

---

## Key Design Decisions

### Fixed 5-Phase Model

**Decision**: All projects use the same 5 phases (discovery, design, implementation, review, finalize)

**Rationale**:
- Predictable structure (always know what phases exist)
- Phase enablement controls which phases execute
- Simpler orchestrator logic (no dynamic phase creation)
- Clear expectations (everyone knows the workflow)

**Alternative Considered**: Variable phases like v1
- Rejected: Added complexity for marginal benefit
- Rejected: Orchestrator struggled with phase planning
- The 5-phase model covers all real-world scenarios

**Trade-off**: Less flexible, but more predictable and easier to understand

---

### Human-Led Planning, AI-Led Execution

**Decision**: Discovery and Design are human-led (orchestrator subservient), Implementation/Review/Finalize are AI-led (orchestrator autonomous)

**Rationale**:
- Observed that AI over-engineers when given unbounded planning
- Humans better at scoping and constraint-setting
- AI excellent at execution within clear boundaries
- Separation prevents orchestrator + architect feedback loop

**Evidence**: Previous version with autonomous design phase resulted in 100% overengineering, even with explicit constraints

**Alternative Considered**: Fully autonomous orchestrator
- Rejected: AI cannot reliably self-impose constraints
- Rejected: Design artifacts became too verbose
- Human involvement in planning essential for quality

---

### Mandatory Review Phase

**Decision**: Review phase always required, orchestrator must perform review

**Rationale**:
- Easy to get side-tracked during implementation
- Final validation before finalization prevents mistakes
- Can loop back to implementation if issues found
- No harm in doing it (net positive for quality)

**Alternative Considered**: Optional review
- Rejected: Quality gate too important to skip
- Rejected: Humans likely to skip when rushed
- Mandatory review forces validation

---

### Phase Enablement vs Phase Execution

**Decision**: All 5 phases always exist in state, but `enabled: true/false` controls execution

**Rationale**:
- Fixed structure simplifies state schema
- Clear distinction between "does this phase exist?" (always yes) and "does this phase run?" (depends on enabled flag)
- Easier to reason about state
- Simpler validation rules

**Alternative Considered**: Dynamic phase creation
- Rejected: Variable state structure is harder to validate
- Rejected: More complex orchestrator logic
- Fixed structure with enablement flag is cleaner

---

### Truth Table for Project Initialization

**Decision**: Use structured question flow with scoring rubrics to determine phase enablement

**Rationale**:
- Objective recommendations (not arbitrary)
- Prevents over-engineering (design rubric with bug fix penalty)
- Consistent guidance across projects
- Human always approves final plan

**Alternative Considered**: Orchestrator decides phases without guidance
- Rejected: Led to inconsistent recommendations
- Rejected: AI tends to over-plan without structure
- Rubrics provide objective framework

---

### CLI-Driven Logging

**Decision**: Agents use `sow log` CLI command instead of direct file editing

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

### Refs Replace Sinks and Repos

**Decision**: Unified "refs" system replaces separate "sinks" (knowledge) and "repos" (code) systems

**Rationale**:
- Simpler mental model (one concept instead of two)
- Same mechanism, semantic distinction (knowledge vs code type)
- Unified CLI commands
- Local caching benefits both types equally

**Alternative Considered**: Keep sinks and repos separate
- Rejected: Unnecessary complexity
- Rejected: Duplicated CLI commands
- Unified system is cleaner

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - Introduction to sow
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - The 5-phase model and truth table
- **[FILE_STRUCTURE.md](./FILE_STRUCTURE.md)** - Complete directory layout
- **[AGENTS.md](./AGENTS.md)** - Agent system details
- **[REFS.md](./REFS.md)** - External references system
- **[LOGGING_AND_STATE.md](./LOGGING_AND_STATE.md)** - State management and resumability
