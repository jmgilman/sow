# 4. Solution Strategy

## Architectural Approach

The sow architecture implements a **stateful orchestration pattern** where an orchestrator agent coordinates multiple specialized worker agents through a well-defined project lifecycle. The key insight is separating human-led planning (discovery, design) from AI-autonomous execution (implementation, review, finalize), with all context persisted to the filesystem for zero-context resumability.

Rather than maintaining conversation history across sessions, sow stores all necessary context in structured files that any agent can read and resume from. This enables session independence, multi-developer collaboration, and transparent debugging. The architecture leverages **filesystem-as-database** where YAML and markdown files provide both machine-readable state and human-readable documentation.

The system is distributed as two layers: an **execution layer** (Claude Code plugin) that defines agent behaviors and a **data layer** (repository `.sow/` directory) that stores all state. This separation enables independent evolution of behavior (plugin updates) and state (project-specific data).

## Key Architectural Decisions

### 1. Filesystem-Based State Management
**Decision**: Store all context in filesystem (YAML/markdown) rather than database or in-memory state.

**Rationale**:
- Git versioning comes free (commit project state with feature branch)
- Human-readable for debugging and transparency
- Zero-context resumability - agents read state from disk
- No database infrastructure required
- Natural multi-developer collaboration via git branches

**Trade-offs**:
- File I/O overhead vs. in-memory speed (acceptable for human-in-loop workflow)
- Schema migration complexity vs. database schema evolution tools
- Concurrent write handling vs. database transactions (mitigated via atomic renames)

---

### 2. State Machine-Driven Project Lifecycle
**Decision**: Use explicit state machine (via stateless library) to manage project phase transitions.

**Rationale**:
- Enforces valid state transitions (prevents skipping review phase, etc.)
- Guards ensure prerequisites met before transitions
- Entry actions generate contextual prompts automatically
- Clear state visibility for orchestrator

**Trade-offs**:
- State machine complexity vs. simple status strings
- Guard function maintenance burden vs. ad-hoc validation
- Rigidity vs. flexibility (intentionally rigid to enforce quality gates)

---

### 3. Multi-Agent Orchestration via Task Tool
**Decision**: Orchestrator spawns worker agents through Claude Code's Task tool rather than sub-processes.

**Rationale**:
- Leverages Claude Code's native agent spawning
- Each worker gets independent context window
- Bounded context prevents orchestrator context bloat
- Workers are stateless - all state persisted to filesystem

**Trade-offs**:
- Platform dependency (Claude Code only) vs. generic implementation
- Agent spawn overhead vs. single long-running agent
- Worker isolation vs. shared memory communication

---

### 4. CUE Schema System
**Decision**: Define all state schemas in CUE, generate Go types automatically.

**Rationale**:
- Single source of truth for schema definitions
- Go type generation ensures compile-time type safety
- CUE validation catches schema violations early
- Human-readable schema documentation

**Trade-offs**:
- CUE learning curve vs. hand-written Go structs
- Code generation step in build process
- Schema evolution requires careful migration planning

---

### 5. Two-Layer Architecture (Plugin + Repository)
**Decision**: Separate execution layer (`.claude/` plugin) from data layer (`.sow/` repository).

**Rationale**:
- Plugin updates don't affect repository state
- Teams share execution layer, customize data layer
- Clear upgrade path (plugin marketplace updates)
- Behavior (agents) independent from state (projects)

**Trade-offs**:
- Two-directory structure complexity vs. single directory
- Plugin synchronization across team members
- Migration burden when execution layer changes state structure

---

## Technology Choices

| Aspect | Choice | Rationale |
|--------|--------|-----------|
| **Programming Language** | Go | Fast compilation, single binary distribution, excellent CLI tooling (Cobra, Billy), strong stdlib |
| **State Machine** | stateless library | Mature, well-tested, supports guards and entry actions, minimal dependencies |
| **Schema Definition** | CUE | Type-safe code generation, human-readable schemas, built-in validation, JSON/YAML compatibility |
| **Filesystem Abstraction** | Billy | Cross-platform filesystem operations, in-memory implementation for tests, chroot support |
| **CLI Framework** | Cobra | Industry standard, excellent UX (help, completions), easy subcommand organization |
| **Git Library** | go-git | Pure Go (no git binary dependency for library usage), full-featured, testable |
| **GitHub Integration** | gh CLI | Leverages user's existing auth, simpler than API client, consistent UX |
| **Prompt Templates** | Go text/template | Standard library, embedded at compile time, simple variable interpolation |
| **Log Format** | Markdown with YAML frontmatter | Human-readable, structured data for parsing, git-friendly diffs |
| **Distribution** | Homebrew + GitHub Releases | Easy installation for macOS/Linux users, automated via GoReleaser |

## Quality Attribute Solutions

### Resumability (Priority 1)
**Approach**: All context stored in filesystem with explicit state files.

**Implementation**:
- Project state: `.sow/project/state.yaml` (phases, tasks, metadata)
- Task state: `state.yaml` per task (iteration, references, status)
- Logs: Chronological markdown logs with agent IDs
- Context: `context/` directory for project-specific decisions

**Validation**: Can pause mid-task, restart CLI, and orchestrator resumes correctly.

---

### State Consistency (Priority 2)
**Approach**: Atomic writes, schema validation, and explicit state machine transitions.

**Implementation**:
- Atomic file writes: Write to `.tmp`, validate, rename (atomic on POSIX)
- CUE schema validation before writes
- State machine guards prevent invalid transitions
- Sentinel errors for clear error handling

**Validation**: Schema violations caught at write time, corrupt state prevented via atomic operations.

---

### Agent Coordination (Priority 3)
**Approach**: Lightweight context compilation, minimal prompt overhead.

**Implementation**:
- Orchestrator reads task description, references from state
- Generates focused prompt with only relevant context
- Worker spawns via Task tool with compiled prompt
- Worker completion reported back to orchestrator

**Validation**: Worker spawn-to-execution measured at < 5 seconds (mostly API latency).

---

### CLI Performance (Priority 4)
**Approach**: Efficient file I/O, minimal allocations, Go's fast compilation.

**Implementation**:
- Read operations: Direct file reads, no parsing overhead unless needed
- Write operations: Atomic writes with single marshal step
- Logging: Append-only writes, no file rewrites
- State loading: Lazy loading, only parse what's needed

**Validation**: Logging operations < 100ms, state reads < 50ms (measured via benchmarks).

---

### Documentation Quality (Priority 5)
**Approach**: Self-documenting state files, structured markdown, clear schemas.

**Implementation**:
- YAML state files with snake_case keys (readable)
- Markdown logs with frontmatter (structured but readable)
- CUE schemas serve as documentation
- Comments in state files preserved via YAML round-tripping

**Validation**: Non-technical stakeholders can read logs and understand progress.

---

## Key Patterns

### Zero-Context Resumability Pattern
Every resumable unit (project, task, mode) includes:
1. **State file** (YAML): Current status, metadata, references
2. **Log file** (Markdown): Chronological action history
3. **Description/Context**: What needs to be done
4. **Feedback** (if applicable): Human corrections

### Orchestrator-Worker Pattern
- **Orchestrator**: Reads project state, compiles context, spawns workers, updates state
- **Worker**: Receives bounded context, executes work, reports completion
- **Communication**: Via filesystem (workers write to logs, orchestrator reads)

### Mode Session Pattern
Exploration, design, and breakdown modes share:
- Branch-based sessions (one session per branch)
- Index file tracking inputs/outputs or work units
- Log file for traceability
- Workspace directory for artifacts

### Reference Caching Pattern
External references cloned once to `~/.cache/sow/`, symlinked to `.sow/refs/`:
- **Committed index**: Categorical metadata shared with team
- **Cache index**: Transient data (SHAs, timestamps) per-machine
- **Subpath support**: Symlink directly to subdirectories
