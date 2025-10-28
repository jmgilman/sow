# 12. Glossary

## Domain Terms

| Term | Definition |
|------|------------|
| **Agent** | AI assistant with specific role and context. Orchestrator (user-facing) or worker (specialized task executor). |
| **Artifact** | Output produced during a phase (notes in discovery, ADRs in design, review reports). |
| **Breakdown** | Process of decomposing design documents into discrete work units (GitHub issues). |
| **Bounded Context** | Limited, focused information provided to worker agent (only task-relevant knowledge). |
| **Context Compilation** | Process where orchestrator gathers task description, references, and knowledge into focused package for worker. |
| **Design Mode** | Operating mode for creating formal design documentation (ADRs, Arc42, design docs, C4 diagrams). |
| **Discovery Phase** | (Deprecated in current model) Initial research and requirement gathering. Now replaced by exploration mode. |
| **Execution Layer** | Behavior definitions in `.claude/` directory (agents, commands, hooks). Distributed via plugin. |
| **Exploration Mode** | Operating mode for freeform research and discovery on `explore/*` branches. |
| **Feedback** | Human correction provided during task execution, stored in `task/feedback/*.md` files. |
| **Iteration** | Counter tracking how many times a task has been attempted. Increments when feedback provided. |
| **Mode** | Operating context outside structured projects (exploration, design, breakdown). |
| **Orchestrator** | Primary user-facing agent that coordinates project lifecycle and delegates to workers. |
| **Phase** | Stage in the 5-phase project lifecycle (planning, implementation, review, finalize). |
| **Project** | Structured implementation workflow with 5 phases, one per git branch. |
| **Reference (Ref)** | External git repository providing knowledge (style guides) or code (examples). |
| **Resumability** | Ability to pause and continue work without losing context, enabled by filesystem-based state. |
| **Session** | Time-bound activity in exploration, design, or breakdown mode, one per branch. |
| **State Machine** | Finite state machine managing project phase transitions with guards and entry actions. |
| **Statechart** | State machine tracking in project state file (current state, history, transitions). |
| **Task** | Individual work item in implementation phase, executed by worker agent. |
| **Worker** | Specialized agent spawned by orchestrator to execute specific task (implementer, architect, etc.). |
| **Work Unit** | Planned work item in breakdown mode, eventually published as GitHub issue. |
| **Zero-Context Resumability** | Core principle: any agent can resume any work by reading filesystem state, no conversation history required. |

---

## Technical Terms

| Term | Definition |
|------|------------|
| **Atomic Write** | Write operation that is either fully completed or not started, never partially written. Implemented via write-to-temp-then-rename. |
| **Billy** | Go filesystem abstraction library providing cross-platform operations and in-memory implementations for testing. |
| **Chroot** | Filesystem root restriction, scoping operations to specific directory (`.sow/` in sow's case). |
| **Cobra** | Go CLI framework used for command parsing, help generation, and subcommand organization. |
| **CUE** | Configuration and schema language used to define state structure and generate Go types. |
| **Entry Action** | Function executed when state machine enters a state, used to generate contextual prompts. |
| **Frontmatter** | YAML metadata block at the top of markdown file, delimited by `---` markers. |
| **go-git** | Pure Go implementation of git protocol, used for repository operations without git binary. |
| **GoReleaser** | Tool for building and releasing Go binaries across multiple platforms, used in CI/CD. |
| **Guard** | Boolean function that must return true for state machine transition to proceed. |
| **Heredoc** | Multi-line string literal (Here Document), used in shell for passing complex strings. |
| **POSIX** | Portable Operating System Interface, standard defining Unix-like OS behavior. |
| **Sentinel Error** | Named error value used for identity comparison via `errors.Is()`. |
| **Snake Case** | Naming convention with lowercase words separated by underscores (e.g., `created_at`). |
| **Stateless** | Go state machine library used for project lifecycle management. |
| **Symlink** | Symbolic link, filesystem reference pointing to another file or directory. |
| **Task Tool** | Claude Code mechanism for spawning new agent contexts with bounded context. |

---

## File and Directory Terms

| Term | Definition |
|------|------------|
| **`.claude/`** | Directory containing execution layer (agents, commands, hooks). Installed by plugin. |
| **`.sow/`** | Directory containing data layer (state, knowledge, refs, projects). Repository-specific. |
| **`.sow/.version`** | File tracking sow structure version for migration compatibility. |
| **`.sow/config.yaml`** | Repository-level configuration (theme, preferences). |
| **`.sow/knowledge/`** | Committed documentation (ADRs, architecture docs, explorations). |
| **`.sow/project/`** | Active project state (singular, one per branch). |
| **`.sow/project/state.yaml`** | Project state file with phases, tasks, statechart. |
| **`.sow/project/log.md`** | Project-level action log (orchestrator actions). |
| **`.sow/project/context/`** | Project-specific decisions and memories. |
| **`.sow/refs/`** | Symlinks to cached external references. |
| **`.sow/refs/index.json`** | Committed index of reference metadata (URL, type, description). |
| **`~/.cache/sow/`** | Local cache directory for cloned references. |
| **`~/.cache/sow/repos/`** | Cloned git repositories. |
| **`~/.cache/sow/index.json`** | Cache index with transient data (SHAs, timestamps). |
| **`phases/implementation/tasks/{id}/`** | Task directory containing state, description, log, feedback. |
| **`phases/implementation/tasks/{id}/state.yaml`** | Task metadata (iteration, status, references). |
| **`phases/implementation/tasks/{id}/description.md`** | Task requirements and context. |
| **`phases/implementation/tasks/{id}/log.md`** | Task-level action log (worker actions). |
| **`phases/implementation/tasks/{id}/feedback/`** | Human corrections for task. |

---

## Status Values

| Term | Definition |
|------|------------|
| **not_started** | Phase has not begun (status). |
| **in_progress** | Phase or task currently being executed (status). |
| **completed** | Phase or task successfully finished (status). |
| **abandoned** | Task stopped without completion (status). |
| **skipped** | Phase intentionally bypassed (status). |
| **pending** | Task created but not yet started (status). |
| **proposed** | Work unit created but not yet documented (breakdown status). |
| **document_created** | Work unit has associated document (breakdown status). |
| **approved** | Work unit approved for publishing (breakdown status). |
| **published** | Work unit published as GitHub issue (breakdown status). |
| **active** | Mode session currently in use (mode status). |
| **completed** | Mode session finished (mode status). |
| **abandoned** | Mode session stopped without completion (mode status). |

---

## State Machine States

| Term | Definition |
|------|------------|
| **NoProject** | No active project exists (initial state). |
| **PlanningActive** | Project in planning phase, gathering requirements. |
| **ImplementationPlanning** | Implementation phase, planning tasks before execution. |
| **ImplementationExecuting** | Implementation phase, tasks being executed. |
| **ReviewActive** | Review phase, orchestrator assessing work quality. |
| **FinalizeDocumentation** | Finalize phase, updating documentation. |
| **FinalizeChecks** | Finalize phase, running tests and linters. |
| **FinalizeDelete** | Finalize phase, cleaning up project state before PR. |

---

## State Machine Events

| Term | Definition |
|------|------------|
| **EventProjectInit** | Project initialization triggered (NoProject → PlanningActive). |
| **EventCompletePlanning** | Planning phase completed (PlanningActive → ImplementationPlanning). |
| **EventTaskCreated** | First task created (ImplementationPlanning → ImplementationExecuting). |
| **EventTasksApproved** | Tasks approved for execution (ImplementationPlanning → ImplementationExecuting). |
| **EventAllTasksComplete** | All tasks completed (ImplementationExecuting → ReviewActive). |
| **EventReviewPass** | Review passed (ReviewActive → FinalizeDocumentation). |
| **EventReviewFail** | Review found issues (ReviewActive → ImplementationPlanning). |
| **EventDocumentationDone** | Documentation complete (FinalizeDocumentation → FinalizeChecks). |
| **EventChecksDone** | Checks complete (FinalizeChecks → FinalizeDelete). |
| **EventProjectDelete** | Project deleted (any state → NoProject). |

---

## Command Terms

| Term | Definition |
|------|------------|
| **`sow init`** | Initialize `.sow/` directory in repository. |
| **`sow start`** | Launch Claude Code with sow greeting (detects context). |
| **`sow project`** | Manage projects (create, continue, status). |
| **`sow explore`** | Start or continue exploration session. |
| **`sow design`** | Start or continue design session. |
| **`sow breakdown`** | Start or continue breakdown session. |
| **`sow refs add`** | Add external reference to project. |
| **`sow refs list`** | List registered references. |
| **`sow log`** | Append structured log entry. |
| **`sow validate`** | Validate `.sow/` structure and schemas. |
| **`sow prompt`** | Render prompt template for agent. |

---

## Acronyms

| Acronym | Expansion |
|---------|-----------|
| **ADR** | Architecture Decision Record |
| **API** | Application Programming Interface |
| **Arc42** | Architecture documentation template with 12 sections |
| **C4** | Context, Containers, Components, Code (architecture diagram model) |
| **CI/CD** | Continuous Integration / Continuous Deployment |
| **CLI** | Command-Line Interface |
| **CUE** | Configure Unify Execute (schema language) |
| **FS** | Filesystem |
| **gh** | GitHub CLI (official command-line tool) |
| **JWT** | JSON Web Token |
| **PR** | Pull Request |
| **SHA** | Secure Hash Algorithm (git commit hash) |
| **TDD** | Test-Driven Development |
| **UX** | User Experience |
| **YAML** | YAML Ain't Markup Language (data serialization format) |

---

## References to Other Sections

For detailed information, see:

- **Agent Roles**: Section 5 (Building Block View)
- **Phase Definitions**: Section 1 (Introduction), Section 4 (Solution Strategy)
- **State Machine Details**: Section 5 (Building Block View), Section 6 (Runtime View)
- **File Structure**: Section 5 (Building Block View)
- **Mode Workflows**: Section 6 (Runtime View)
- **Quality Attributes**: Section 10 (Quality Requirements)
- **Architectural Decisions**: Section 9 (Architecture Decisions)
