# 9. Architecture Decisions

All architectural decisions are documented as Architecture Decision Records (ADRs). This section maintains the index and provides summaries of key decisions that shaped the sow architecture.

## Decision Log

| ID      | Title                                        | Status   | Date         | Sections Affected |
| ------- | -------------------------------------------- | -------- | ------------ | ----------------- |
| ADR-001 | Filesystem-Based State Management            | Accepted | Design Phase | 4, 5, 8           |
| ADR-002 | Fixed 5-Phase Project Model                  | Accepted | Design Phase | 1, 4, 5, 6        |
| ADR-003 | Single Project Per Branch Constraint         | Accepted | Design Phase | 2, 4, 5, 8        |
| ADR-004 | Multi-Agent Orchestration via Task Tool      | Accepted | Design Phase | 1, 3, 4, 5, 6     |
| ADR-005 | CUE Schema System with Go Code Generation    | Accepted | Design Phase | 2, 4, 5, 8        |
| ADR-006 | Two-Layer Architecture (Plugin + Repository) | Accepted | Design Phase | 1, 2, 3, 5, 7     |
| ADR-007 | State Machine-Driven Phase Transitions       | Accepted | Design Phase | 4, 5, 6, 8        |
| ADR-008 | Human-Led Planning, AI-Led Execution         | Accepted | Design Phase | 1, 4, 6           |
| ADR-009 | Reference Caching with Symlinks              | Accepted | Design Phase | 5, 6, 7           |
| ADR-010 | GitHub CLI Wrapper (Not Direct API)          | Accepted | Design Phase | 2, 3, 5           |
| ADR-011 | Markdown Logging with YAML Frontmatter       | Accepted | Design Phase | 5, 6, 8           |
| ADR-012 | Mandatory Review Phase                       | Accepted | Design Phase | 1, 4, 5           |

## Key Decisions Summary

### ADR-001: Filesystem-Based State Management

**Decision**: Store all context in filesystem (YAML/markdown) rather than database or in-memory state.

**Context**:
Need to enable zero-context resumability where any agent can resume any project/task without conversation history. Options considered: database (PostgreSQL, SQLite), in-memory state with serialization, filesystem with structured files.

**Decision**:
Use filesystem as database with `.sow/` directory containing YAML state files and markdown logs. All state persisted to disk using atomic write pattern (write to .tmp, rename).

**Rationale**:
- **Git versioning**: State committed with code, automatic backup and history
- **Human-readable**: Developers can inspect and debug state files
- **Zero-context resumability**: Agents read complete context from disk
- **No infrastructure**: No database setup, just filesystem
- **Multi-developer**: Natural collaboration via git branches
- **Transparency**: State is visible, auditable, version-controlled

**Consequences**:
- ✅ Simple deployment (no database)
- ✅ Git integration (version control, branching, merging)
- ✅ Human debugging (readable YAML/markdown)
- ✅ Zero-context resumption (read from disk)
- ⚠️ File I/O overhead (acceptable for human-in-loop workflow)
- ⚠️ Concurrent writes require atomic operations
- ⚠️ Schema migration complexity (no database migration tools)

**Alternatives Rejected**:
- Database (SQLite): Infrastructure overhead, not git-versionable, opaque
- In-memory: Requires serialization, no git integration, fragile
- JSON files: Less human-readable than YAML

**Impact**: Sections 4 (Solution Strategy), 5 (Building Blocks), 8 (Cross-cutting)

---

### ADR-002: Fixed 5-Phase Project Model

**Decision**: All projects use the same 5 phases (Planning → Implementation → Review → Finalize). Phase enablement controls execution.

**Context**:
Need predictable structure for orchestrator while supporting variable workflows. Options: dynamic phase creation (v1 approach), fixed phases with enablement flags, task-only model (no phases).

**Decision**:
Fixed 5-phase model with phase enablement flags. All projects have same phases, but `enabled: true/false` controls which execute.

**Rationale**:
- **Predictability**: Orchestrator always knows phase structure
- **Simplicity**: No dynamic phase planning complexity
- **Quality gates**: Review phase can be mandatory
- **Consistent expectations**: Developers know the workflow
- **Easy state machine**: Fixed states, clear transitions

**Consequences**:
- ✅ Simple orchestrator logic (no phase planning)
- ✅ Predictable workflow (always know what phases exist)
- ✅ Mandatory quality gates (review phase enforced)
- ✅ Easier documentation (single workflow to explain)
- ⚠️ Less flexible (cannot add custom phases)
- ⚠️ Disabled phases still in state (clutter)

**Alternatives Rejected**:
- Dynamic phases: Orchestrator struggled, over-planned
- Task-only: No clear workflow structure
- Configurable phases: Complexity without clear benefit

**Impact**: Sections 1 (Goals), 4 (Strategy), 5 (Building Blocks), 6 (Runtime)

---

### ADR-003: Single Project Per Branch Constraint

**Decision**: Only one project can exist per git branch. Enforced, not suggested.

**Context**:
Need to prevent orchestrator confusion and enable clean git integration. Options: multiple projects per repo, one project per branch, project directory name includes branch.

**Decision**:
Enforce single project per branch. Error if trying to create project when one exists. Project state at `.sow/project/` (singular, not plural).

**Rationale**:
- **Simplicity**: Orchestrator always knows current project
- **Git integration**: One branch = one project (natural model)
- **Cleanup**: Delete branch = project gone (no orphans)
- **Filesystem**: Clean structure (no `projects/` directory)
- **Command UX**: `/project:continue` needs no arguments

**Consequences**:
- ✅ Simple orchestrator (no "which project?" logic)
- ✅ Natural git workflow (branch = project)
- ✅ Easy cleanup (delete branch)
- ✅ Predictable paths (always `.sow/project/`)
- ⚠️ Cannot work on multiple projects simultaneously (use multiple branches)
- ⚠️ Branch switching changes project context (intentional)

**Alternatives Rejected**:
- Multiple projects: Orchestrator needs project selection, clutter
- Project ID in path: Breaks predictability, complicates commands

**Impact**: Sections 2 (Constraints), 4 (Strategy), 5 (Building Blocks), 8 (Cross-cutting)

---

### ADR-004: Multi-Agent Orchestration via Task Tool

**Decision**: Orchestrator spawns worker agents through Claude Code's Task tool rather than sub-processes.

**Context**:
Need to delegate work to specialized agents. Options: spawn workers via subprocess (run agent scripts), use Task tool (Claude Code native), single agent with role switching.

**Decision**:
Use Claude Code's Task tool to spawn worker agents. Each worker gets independent context window with compiled task context.

**Rationale**:
- **Platform integration**: Leverages Claude Code native capability
- **Context isolation**: Each worker has independent context window
- **Bounded context**: Workers receive only task-relevant information
- **Stateless workers**: All state on disk, workers are ephemeral
- **Scalability**: Can spawn multiple workers in parallel

**Consequences**:
- ✅ Native Claude Code integration
- ✅ Independent context windows (no orchestrator bloat)
- ✅ Bounded context (workers focused)
- ✅ Parallel execution possible
- ⚠️ Platform dependency (Claude Code only)
- ⚠️ Agent spawn overhead (API latency)

**Alternatives Rejected**:
- Subprocess spawning: Not supported by Claude Code
- Single agent: Context window bloat, role confusion
- External agent platform: Additional infrastructure

**Impact**: Sections 1 (Goals), 3 (Context), 4 (Strategy), 5 (Building Blocks), 6 (Runtime)

---

### ADR-005: CUE Schema System with Go Code Generation

**Decision**: Define all state schemas in CUE, generate Go types automatically via `cue exp gengotypes`.

**Context**:
Need type-safe state management with validation. Options: hand-written Go structs, JSON Schema with validation, CUE schemas with generation, Protocol Buffers.

**Decision**:
Use CUE schema language for definitions, generate Go types at build time. Embed schemas in CLI binary for runtime validation.

**Rationale**:
- **Single source of truth**: Schema defines both validation and types
- **Type safety**: Generated Go types ensure compile-time safety
- **Validation**: CUE validates YAML against schema
- **Human-readable**: CUE schemas are clear documentation
- **Version control**: Schema changes tracked in git

**Consequences**:
- ✅ Type-safe code generation
- ✅ Built-in validation (CUE)
- ✅ Self-documenting (schema is documentation)
- ✅ Compile-time safety
- ⚠️ CUE learning curve for contributors
- ⚠️ Code generation step in build
- ⚠️ Schema migration complexity

**Alternatives Rejected**:
- Hand-written structs: No validation, duplication with docs
- JSON Schema: Less human-readable, separate validation step
- Protocol Buffers: Overkill, binary format undesirable

**Impact**: Sections 2 (Constraints), 4 (Strategy), 5 (Building Blocks), 8 (Cross-cutting)

---

### ADR-006: Two-Layer Architecture (Plugin + Repository)

**Decision**: Separate execution layer (`.claude/` plugin) from data layer (`.sow/` repository).

**Context**:
Need to enable independent evolution of behavior and state. Options: single directory (`.claude/` only), plugin includes state templates, two-layer separation.

**Decision**:
Two layers: `.claude/` contains execution (agents, commands, hooks), `.sow/` contains data (state, knowledge, refs). Plugin distributed via marketplace, state is repository-specific.

**Rationale**:
- **Independent evolution**: Update plugin without affecting state
- **Distribution**: Plugin via marketplace, state via git
- **Customization**: Teams share execution, customize data
- **Clear separation**: Behavior vs. state (distinct concerns)
- **Upgrade path**: Plugin updates don't break repositories

**Consequences**:
- ✅ Plugin upgrades independent from state
- ✅ Clear responsibility boundaries
- ✅ Easy distribution (marketplace)
- ✅ Team sharing (execution) and customization (data)
- ⚠️ Two directories to understand
- ⚠️ Migration when execution layer changes state structure
- ⚠️ Plugin version synchronization across team

**Alternatives Rejected**:
- Single directory: Upgrades affect state, no clear boundary
- Plugin includes state: State not repository-specific
- CLI-only (no plugin): Loses Claude Code integration

**Impact**: Sections 1 (Goals), 2 (Constraints), 3 (Context), 5 (Building Blocks), 7 (Deployment)

---

### ADR-007: State Machine-Driven Phase Transitions

**Decision**: Use explicit state machine (stateless library) to manage project phase transitions with guards and entry actions.

**Context**:
Need to enforce valid phase transitions and generate contextual prompts. Options: simple status strings, state machine library, event sourcing.

**Decision**:
Use stateless library to implement state machine. States correspond to phases/sub-phases. Guards ensure prerequisites. Entry actions generate prompts.

**Rationale**:
- **Enforced transitions**: Cannot skip quality gates (review)
- **Prerequisites**: Guards check conditions before transitions
- **Automatic prompts**: Entry actions generate contextual prompts
- **Clear states**: Explicit state names, no ambiguity
- **History**: State machine tracks transition history

**Consequences**:
- ✅ Valid transitions enforced (quality gates work)
- ✅ Contextual prompts automatic
- ✅ Clear state visibility
- ✅ Prerequisites checked (guards)
- ⚠️ State machine complexity (learning curve)
- ⚠️ Guard function maintenance
- ⚠️ Less flexible (intentionally rigid)

**Alternatives Rejected**:
- Simple status strings: No transition validation
- Event sourcing: Overkill for this use case
- Manual transition checks: Error-prone, scattered logic

**Impact**: Sections 4 (Strategy), 5 (Building Blocks), 6 (Runtime), 8 (Cross-cutting)

---

### ADR-008: Human-Led Planning, AI-Led Execution

**Decision**: Discovery and Design phases are human-led (orchestrator subservient), Implementation/Review/Finalize are AI-led (orchestrator autonomous).

**Context**:
Observed that AI over-engineers when given unbounded planning. Need to leverage human judgment for scoping while enabling AI execution speed.

**Decision**:
Two orchestrator modes: Subservient (discovery, design) asks questions and takes notes. Autonomous (implementation, review, finalize) executes independently, only requesting approval for major changes.

**Rationale**:
- **Human scoping**: Humans better at constraints and judgment
- **AI execution**: AI excellent at systematic implementation
- **Prevent over-engineering**: Bounded planning phase
- **Leverage strengths**: Each party does what they do best
- **Clear boundaries**: Subservient vs. autonomous well-defined

**Consequences**:
- ✅ Appropriate use of human and AI strengths
- ✅ Prevents over-engineering (human-led planning)
- ✅ Fast execution (AI-led implementation)
- ✅ Quality gates (review phase)
- ⚠️ Requires discipline (humans must scope well)
- ⚠️ Mode switching complexity (orchestrator behavior changes)

**Alternatives Rejected**:
- Fully autonomous: AI over-plans, ignores constraints
- Fully manual: Loses AI execution speed
- Always subservient: Too slow, micromanagement

**Impact**: Sections 1 (Goals), 4 (Strategy), 6 (Runtime)

---

### ADR-009: Reference Caching with Symlinks

**Decision**: Clone external references once to `~/.cache/sow/`, symlink to `.sow/refs/` per project.

**Context**:
Need to provide external knowledge (style guides, code examples) to agents. Options: clone per project, global cache with symlinks, submodules, sparse checkout.

**Decision**:
Local caching in `~/.cache/sow/repos/`, symlinked (Unix) or copied (Windows) to `.sow/refs/`. Two-index system: committed (metadata) and cache (transient data).

**Rationale**:
- **Efficiency**: Clone once, use in all projects
- **Disk usage**: Shared cache reduces duplication
- **Subpath support**: Symlink directly to subdirectories
- **Offline work**: Cache persists locally
- **Team sharing**: Committed index shares metadata

**Consequences**:
- ✅ Efficient disk usage (shared cache)
- ✅ Subpath support (without sparse checkout)
- ✅ Offline capable (cached locally)
- ✅ Team metadata sharing (committed index)
- ⚠️ Symlink limitations on Windows (use copies)
- ⚠️ Cache management (staleness detection)
- ⚠️ Two-index complexity

**Alternatives Rejected**:
- Clone per project: Waste disk space
- Git submodules: Complex, inflexible
- Sparse checkout: Doesn't support subpaths well

**Impact**: Sections 5 (Building Blocks), 6 (Runtime), 7 (Deployment)

---

### ADR-010: GitHub CLI Wrapper (Not Direct API)

**Decision**: Interact with GitHub via `gh` CLI wrapper, not Go API client library.

**Context**:
Need to create issues, link branches, create PRs. Options: GitHub API client (go-github), GitHub CLI (gh) wrapper, custom API calls.

**Decision**:
Wrap `gh` CLI by executing shell commands and parsing JSON output. Check authentication via `gh auth status`.

**Rationale**:
- **Leverage existing auth**: User's `gh` authentication reused
- **Simpler implementation**: No OAuth flow, token management
- **Consistent UX**: Same auth as user's other `gh` usage
- **Feature parity**: Get all `gh` features automatically
- **Less code**: No API client maintenance

**Consequences**:
- ✅ Simple implementation (shell exec)
- ✅ Reuses user authentication
- ✅ Feature parity with `gh` CLI
- ✅ Consistent user experience
- ⚠️ Dependency on `gh` binary (must be installed)
- ⚠️ Shell execution overhead (minor)
- ⚠️ Parsing CLI output (less stable than API)

**Alternatives Rejected**:
- Go API client: Auth complexity, token management
- Custom API calls: Reinventing wheel
- No GitHub integration: Loses workflow benefits

**Impact**: Sections 2 (Constraints), 3 (Context), 5 (Building Blocks)

---

### ADR-011: Markdown Logging with YAML Frontmatter

**Decision**: Use markdown files with YAML frontmatter for structured logging.

**Context**:
Need human-readable logs with machine-parseable structure. Options: pure YAML, JSON logs, markdown with frontmatter, plain text.

**Decision**:
Markdown files with YAML frontmatter containing structured data (timestamp, agent, action, result, files) and optional markdown body for notes.

**Rationale**:
- **Human-readable**: Markdown renders nicely in editors
- **Structured**: YAML frontmatter is machine-parseable
- **Flexible**: Optional notes in markdown body
- **Git-friendly**: Text format diffs well
- **Self-documenting**: Clear format for humans and agents

**Consequences**:
- ✅ Human-readable (markdown)
- ✅ Machine-parseable (YAML frontmatter)
- ✅ Git-friendly (text diffs)
- ✅ Flexible (structured + freeform)
- ⚠️ Slightly more complex parsing (frontmatter + body)
- ⚠️ Frontmatter format must be exact (YAML)

**Alternatives Rejected**:
- Pure YAML: Less readable for long notes
- JSON logs: Not human-friendly
- Plain text: No structure for parsing

**Impact**: Sections 5 (Building Blocks), 6 (Runtime), 8 (Cross-cutting)

---

### ADR-012: Mandatory Review Phase

**Decision**: Review phase always required, orchestrator must perform review before finalization.

**Context**:
Easy to get side-tracked during implementation, need quality gate. Options: optional review, mandatory review, manual review command.

**Decision**:
Review phase always enabled and executed. State machine enforces transition through review before finalize. Can loop back to implementation if issues found.

**Rationale**:
- **Quality gate**: Final validation before finalization
- **Catch issues**: Verify implementation meets requirements
- **Low cost**: Review is fast (orchestrator reads, assesses)
- **Prevents mistakes**: Easy to forget to review manually
- **Enforced**: State machine requires review transition

**Consequences**:
- ✅ Quality gate always executed
- ✅ Catches implementation drift
- ✅ No harm (fast operation)
- ✅ Can loop back if needed
- ⚠️ Cannot skip (intentional constraint)
- ⚠️ Minor overhead for simple changes

**Alternatives Rejected**:
- Optional review: Likely skipped when rushed
- Manual review: Easy to forget
- No review: Quality suffers

**Impact**: Sections 1 (Goals), 4 (Strategy), 5 (Building Blocks)

---

## Cross-Reference with Other Documentation

**With Constraints (Section 2)**:
- ADR-010 explains GitHub CLI constraint
- ADR-003 explains single project constraint
- ADR-005 explains CUE schema constraint

**With Solution Strategy (Section 4)**:
- ADR-001, 002, 003, 004, 007 shape overall architectural approach
- ADR-008 defines human-AI collaboration model

**With Building Blocks (Section 5)**:
- ADR-006 explains two-layer architecture
- ADR-005 explains schema system
- ADR-009 explains reference caching

**With Runtime View (Section 6)**:
- ADR-004 explains worker spawning
- ADR-007 explains state transitions
- ADR-011 explains logging format
