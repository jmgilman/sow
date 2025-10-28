# 10. Quality Requirements

## Quality Tree

```
sow Quality Requirements
├── Resumability (Priority 1)
│   ├── Context Recovery (< 2 sec)
│   ├── State Completeness (100%)
│   └── Session Independence
├── Reliability
│   ├── State Consistency (Zero corruption)
│   ├── Atomic Operations
│   └── Schema Validation
├── Performance
│   ├── CLI Responsiveness (< 100ms)
│   ├── Agent Coordination (< 5 sec)
│   └── State Loading (< 100ms)
├── Usability
│   ├── Documentation Quality (Human-readable)
│   ├── Error Messages (Clear, actionable)
│   └── Git Integration (Natural workflow)
├── Maintainability
│   ├── Code Coverage (> 80%)
│   ├── Type Safety (CUE → Go)
│   └── Clear Architecture
└── Portability
    ├── Cross-Platform (macOS, Linux, Windows)
    ├── Filesystem Abstraction
    └── Binary Distribution
```

---

## Quality Scenarios

### Resumability Scenarios

#### QS-1: Zero-Context Project Resumption
**Scenario**: Developer resumes project after days/weeks without context loss

**Source**: Developer
**Stimulus**: Runs `sow project` after extended break (days/weeks)
**Artifact**: Project state files in `.sow/project/`
**Environment**: Normal operation, different machine possible
**Response**: Orchestrator loads complete context from filesystem and presents current status
**Measure**:
- Context load time: < 2 seconds
- Context completeness: 100% (no missing information)
- Success rate: 100% (always resumes correctly)

**Implementation**:
- All context in `state.yaml`, `log.md`, task files
- No reliance on conversation history
- State machine reconstructs at current state
- Logs provide complete action history

---

#### QS-2: Multi-Developer Handoff
**Scenario**: Developer A starts project, Developer B continues

**Source**: Developer B
**Stimulus**: Checks out feature branch, runs `sow project`
**Artifact**: Project state committed to git branch
**Environment**: Different developer, different machine
**Response**: Orchestrator loads context, enables continuation
**Measure**:
- Handoff success rate: 100%
- Context loss: 0% (full state transfer)
- Additional setup: 0 steps (just git checkout)

**Implementation**:
- Project state committed to feature branch
- Git push/pull transfers complete context
- Filesystem-based state (no developer-specific data)

---

#### QS-3: Task Iteration Resumption
**Scenario**: Worker receives feedback, iteration incremented, new worker spawned

**Source**: Orchestrator
**Stimulus**: Developer provides correction during task execution
**Artifact**: Task state with iteration counter, feedback files
**Environment**: Normal operation, mid-task
**Response**: New worker receives complete context including all feedback
**Measure**:
- Feedback incorporation: 100% (all feedback visible to new worker)
- Iteration tracking: Accurate (reflected in task state and agent ID)
- Context completeness: 100% (description + state + log + feedback)

**Implementation**:
- Iteration counter in `task/state.yaml`
- Feedback files in `task/feedback/`
- Worker reads all feedback on spawn
- Agent ID includes iteration (`implementer-2`)

---

### Reliability Scenarios

#### QS-4: State File Corruption Prevention
**Scenario**: CLI crashes during state write operation

**Source**: Operating system
**Stimulus**: Kill -9 during file write, power loss, disk full
**Artifact**: Project or task state files
**Environment**: Abnormal termination
**Response**: State file either correct (old or new) or recoverable, never corrupted
**Measure**:
- Corruption rate: 0%
- Recovery success: 100%
- Data loss: At most last incomplete write

**Implementation**:
- Atomic writes (write to `.tmp`, rename)
- POSIX rename atomicity guarantee
- Schema validation on load detects corruption
- Git versioning provides backup

---

#### QS-5: Concurrent State Access
**Scenario**: Multiple CLI processes attempt to modify state simultaneously

**Source**: User error (multiple terminals)
**Stimulus**: Two CLI commands modifying same project state
**Artifact**: Project state file
**Environment**: Race condition possible
**Response**: Last writer wins (atomic rename), no corruption
**Measure**:
- Corruption rate: 0%
- Undefined behavior: Acceptable (user error)
- State validity: Maintained (both writes valid, one overwrites)

**Implementation**:
- Atomic writes (rename is atomic)
- No advisory locking (acceptable trade-off)
- State machine prevents most invalid states

---

#### QS-6: Schema Validation Catches Errors
**Scenario**: Corrupted or manually edited state file loaded

**Source**: Manual edit, external tool, corruption
**Stimulus**: CLI loads state file with schema violation
**Artifact**: Any YAML state file
**Environment**: Normal operation
**Response**: Clear error message indicating validation failure, field path, expected values
**Measure**:
- Detection rate: 100% (all schema violations caught)
- Error message quality: Field path + expected value provided
- Crash prevention: 100% (graceful error, not panic)

**Implementation**:
- CUE schema validation on load
- Detailed validation error messages
- Type-safe Go structs from schema generation

---

### Performance Scenarios

#### QS-7: Fast CLI Response Times
**Scenario**: Developer uses CLI commands during workflow

**Source**: Developer
**Stimulus**: Executes common CLI commands (logging, status, task operations)
**Artifact**: CLI binary
**Environment**: Normal operation
**Response**: Command completes and returns control
**Measure**:
- Startup overhead: < 50ms
- State read operations: < 100ms
- Log append operations: < 100ms
- Simple commands: < 200ms

**Implementation**:
- Efficient file I/O (minimal reads)
- Lazy loading (only read what's needed)
- Single marshal pass for writes
- Append-only logs (no rewrites)

---

#### QS-8: Responsive Agent Coordination
**Scenario**: Orchestrator delegates task to worker agent

**Source**: Orchestrator agent
**Stimulus**: Spawns worker via Task tool
**Artifact**: Task context compilation, agent spawn
**Environment**: Normal operation
**Response**: Worker begins execution
**Measure**:
- Context compilation: < 1 second
- Spawn-to-execution: < 5 seconds total
- Context size: < 10KB (bounded context)

**Implementation**:
- Efficient context compilation (minimal file reads)
- Bounded context (only task-relevant info)
- Pre-compiled prompts (templates)
- Task tool overhead (mostly API latency, not controllable)

---

### Usability Scenarios

#### QS-9: Clear Error Messages
**Scenario**: User makes mistake (invalid command, missing auth, etc.)

**Source**: User
**Stimulus**: Invalid CLI invocation, missing prerequisites
**Artifact**: CLI error handling
**Environment**: Normal operation
**Response**: Clear error message with actionable guidance
**Measure**:
- Message clarity: Specifies what's wrong and how to fix
- Actionability: Provides next step (e.g., "Run: sow init")
- Exit code: Non-zero (1) for errors

**Implementation**:
- Sentinel errors with clear messages
- Error wrapping with context
- User-friendly error formatting
- Suggest corrective actions

**Examples**:
```
Error: sow not initialized
Run: sow init

Error: no project exists on this branch
Create one with: sow project

Error: GitHub CLI not authenticated
Run: gh auth login
```

---

#### QS-10: Human-Readable State
**Scenario**: Developer inspects state files for debugging

**Source**: Developer
**Stimulus**: Opens state or log file in editor
**Artifact**: YAML/markdown state files
**Environment**: Debugging, understanding
**Response**: Developer can read and understand state
**Measure**:
- Readability: Non-technical person can follow structure
- Completeness: All relevant information visible
- Debuggability: Can identify issues from state inspection

**Implementation**:
- YAML for state (human-friendly)
- Markdown for logs (readable)
- Snake_case keys (conventional)
- Null values omitted (cleaner)
- Comments preserved (future)

---

#### QS-11: Natural Git Integration
**Scenario**: Developer uses standard git workflow

**Source**: Developer
**Stimulus**: Creates branch, commits, pushes, creates PR
**Artifact**: Git repository with `.sow/` directory
**Environment**: Normal git workflow
**Response**: sow integrates seamlessly, no special git handling required
**Measure**:
- Additional git steps: 0 (standard workflow)
- Merge conflicts: Rare (state files merge-friendly)
- Branch switching: Automatic project context switch

**Implementation**:
- Project state committed to feature branches
- Text files (YAML/markdown) are git-friendly
- Branch switching handled by git (no sow intervention)
- One project per branch (clean model)

---

### Maintainability Scenarios

#### QS-12: Type-Safe Code Changes
**Scenario**: Developer modifies state structure

**Source**: Maintainer
**Stimulus**: Changes CUE schema (adds field, changes type)
**Artifact**: CUE schema files, generated Go types
**Environment**: Development
**Response**: Go compiler catches type mismatches
**Measure**:
- Type safety: 100% (compile-time errors)
- Schema-code sync: Automatic (generated types)
- Breaking changes: Detected at compile time

**Implementation**:
- CUE schemas define state structure
- `cue exp gengotypes` generates Go types
- Go code uses generated types
- Compiler enforces type correctness

---

#### QS-13: Comprehensive Test Coverage
**Scenario**: Refactoring or adding new features

**Source**: Maintainer
**Stimulus**: Modifies internal packages
**Artifact**: Unit test suite
**Environment**: Development
**Response**: Tests catch regressions
**Measure**:
- Code coverage: > 80% for all packages
- Test execution time: < 30 seconds
- CI integration: All tests run on PR

**Implementation**:
- Table-driven unit tests
- In-memory filesystem (Billy)
- Mock executors for external commands
- Comprehensive integration tests

---

### Portability Scenarios

#### QS-14: Cross-Platform Compatibility
**Scenario**: User installs on different operating systems

**Source**: User
**Stimulus**: Installs and runs on macOS, Linux, or Windows
**Artifact**: sow CLI binary
**Environment**: Different OS, architecture
**Response**: sow works identically (within platform constraints)
**Measure**:
- Functionality parity: 100% (same features)
- Platform-specific handling: Automatic (symlinks vs. copies)
- Binary availability: All platforms (6 binaries: Linux/macOS/Windows × amd64/arm64)

**Implementation**:
- Billy filesystem abstraction (cross-platform)
- Go stdlib (portable)
- Platform detection for symlinks vs. copies (Windows)
- GoReleaser for multi-platform builds

---

#### QS-15: Installation Simplicity
**Scenario**: New user installs sow

**Source**: New user
**Stimulus**: Wants to use sow
**Artifact**: Installation process
**Environment**: User's machine
**Response**: sow installed and ready to use
**Measure**:
- Installation steps: 1 (Homebrew) or 2-3 (manual)
- Prerequisites: Git, Claude Code (assumed for target audience)
- Time to install: < 5 minutes

**Implementation**:
- Homebrew tap (macOS/Linux)
- GitHub Releases with binaries
- Interactive installer via plugin (`/sow:install`)
- Clear installation docs

---

## Quality Attribute Trade-offs

### Resumability vs. Performance
**Trade-off**: Frequent disk writes for resumability vs. in-memory speed

**Decision**: Prioritize resumability. Write state to disk on every significant change.

**Justification**: Human-in-loop workflow means performance overhead is acceptable. Zero-context resumability is core value proposition.

---

### Simplicity vs. Flexibility
**Trade-off**: Fixed 5-phase model vs. configurable phases

**Decision**: Fixed model with enablement flags.

**Justification**: Simplicity enables predictable orchestrator logic. Flexibility via enablement is sufficient for real-world scenarios.

---

### Type Safety vs. Dynamic Flexibility
**Trade-off**: Strict CUE schemas vs. dynamic YAML structures

**Decision**: Strict schemas with code generation.

**Justification**: Prevents state corruption, enables confident refactoring. Migration tooling handles schema evolution.

---

### Cross-Platform Support vs. Optimal Performance
**Trade-off**: Abstraction overhead (Billy) vs. native filesystem calls

**Decision**: Use abstraction for portability.

**Justification**: Overhead is negligible for human-in-loop workflow. Portability and testability outweigh minor performance cost.

---

## Non-Functional Requirements Priority

| Priority | Quality Attribute | Rationale |
|----------|------------------|-----------|
| 1 | **Resumability** | Core value proposition, differentiator |
| 2 | **Reliability** | State corruption would break resumability |
| 3 | **Performance** | Must be responsive for developer workflow |
| 4 | **Usability** | Developer tool, good UX is critical |
| 5 | **Maintainability** | Long-term sustainability, community contributions |
| 6 | **Portability** | Reach broader audience, minimal constraints |
