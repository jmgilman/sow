# sow Development Roadmap

**Last Updated**: 2025-10-12
**Status**: Preliminary roadmap for system implementation

---

## Overview

This roadmap outlines the sequential milestones for building the sow (system of work) framework. Each milestone represents a logical chunk of work that, when combined, results in a complete AI-powered system of work for software engineering.

For detailed architectural decisions and technical specifications, see the comprehensive architecture documentation in this directory.

---

## Milestone 1: Foundation and Core Infrastructure

**Goal**: Establish the basic file structure, schemas, and version management system

**Key Deliverables**:
- Define and implement the two-layer architecture (execution + data layers)
- Create file structure templates for `.claude/` and `.sow/` directories
- Implement version tracking system (`.sow/.version` and `.plugin-version`)
- Define YAML/JSON schemas for all state files
- Create basic validation utilities for structure integrity

**References**:
- [FILE_STRUCTURE.md](./FILE_STRUCTURE.md) - Complete directory layout
- [SCHEMAS.md](./SCHEMAS.md) - File format specifications
- [ARCHITECTURE.md](./ARCHITECTURE.md#two-layer-architecture) - Architectural foundation

**Success Criteria**:
- File structures can be created and validated
- Version files correctly track system versions
- All schemas are well-defined and parseable

---

## Milestone 2: Orchestrator Agent (MVP)

**Goal**: Build the main coordinating agent that serves as the user interface

**Key Deliverables**:
- Create orchestrator agent with system prompt
- Implement startup behavior (SessionStart detection)
- Build basic command routing (one-off vs project-based work)
- Create simple task delegation mechanism
- Implement project state reading and writing

**References**:
- [AGENTS.md](./AGENTS.md#orchestrator) - Orchestrator role and responsibilities
- [ARCHITECTURE.md](./ARCHITECTURE.md#multi-agent-system) - Multi-agent architecture

**Success Criteria**:
- Orchestrator can start up and assess work mode
- Can create basic project structure
- Can read and update project state files
- Can identify next task to work on

---

## Milestone 3: Project Management Core

**Goal**: Implement the foundational project management system

**Key Deliverables**:
- Create project initialization logic (`/start-project`)
- Implement complexity assessment (1-3 rating)
- Build progressive planning system (start with 1-2 phases)
- Create phase management (creation, transitions, validation)
- Implement gap numbering system for tasks
- Build task state management (pending, in_progress, completed, abandoned)

**References**:
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md) - Complete project lifecycle
- [ARCHITECTURE.md](./ARCHITECTURE.md#progressive-planning-philosophy) - Planning approach

**Success Criteria**:
- Projects can be created with initial phases
- Tasks use gap numbering correctly
- Phase transitions work as expected
- State persists correctly across sessions

---

## Milestone 4: Worker Agents (Core Set)

**Goal**: Build the specialized worker agents for different types of work

**Key Deliverables**:
- Create architect agent (design and ADRs)
- Create implementer agent (TDD-focused coding)
- Create integration-tester agent (E2E and integration tests)
- Create reviewer agent (code quality)
- Create documenter agent (documentation updates)
- Implement agent spawning mechanism via Task tool
- Build agent assignment logic

**References**:
- [AGENTS.md](./AGENTS.md#worker-agents) - Worker agent specifications
- [ARCHITECTURE.md](./ARCHITECTURE.md#orchestrator--worker-pattern) - Agent coordination

**Success Criteria**:
- All five worker agents are functional
- Orchestrator can spawn workers correctly
- Workers can execute tasks independently
- Task assignment works based on task type

---

## Milestone 5: Context Compilation and Zero-Context Resumability

**Goal**: Enable agents to resume work from filesystem state without conversation history

**Key Deliverables**:
- Build context compilation system in orchestrator
- Implement task description format and generation
- Create reference tracking system (sinks, knowledge, repos)
- Build worker recovery process (read state → execute → report)
- Implement iteration tracking system
- Create task and project log structure

**References**:
- [ARCHITECTURE.md](./ARCHITECTURE.md#zero-context-resumability) - Resumability design
- [AGENTS.md](./AGENTS.md#context-compilation) - Context compilation process
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md#zero-context-resumability) - Recovery procedures

**Success Criteria**:
- Workers can resume tasks from filesystem state alone
- No dependency on conversation history
- Context is appropriately filtered for each worker
- Iteration counter tracks attempts correctly

---

## Milestone 6: Logging System

**Goal**: Implement structured action logging for audit trails and recovery

**Key Deliverables**:
- Define structured markdown log format
- Create action vocabulary for logging
- Build CLI logging command (`sow log`)
- Implement auto-detection of task vs project logs
- Create agent ID construction (role + iteration)
- Build log reading utilities

**References**:
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md#logging-system) - Logging specifications
- [ARCHITECTURE.md](./ARCHITECTURE.md#cli-driven-logging) - CLI-based approach
- [CLI_REFERENCE.md](./CLI_REFERENCE.md#sow-log) - CLI logging command

**Success Criteria**:
- Logs are created with consistent format
- CLI logging is fast (<1s vs 30s file edit)
- Agent IDs are constructed correctly
- Logs support zero-context resumability

---

## Milestone 7: Feedback Mechanism

**Goal**: Enable human-in-the-loop corrections and guidance

**Key Deliverables**:
- Create feedback file structure and format
- Implement feedback creation by orchestrator
- Build feedback reading in workers
- Create feedback status tracking (pending, addressed, superseded)
- Integrate feedback with iteration system

**References**:
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md#feedback-mechanism) - Feedback system
- [USER_GUIDE.md](./USER_GUIDE.md#providing-feedback-to-agents) - User experience

**Success Criteria**:
- Users can provide corrections to agents
- Workers read and incorporate feedback
- Feedback status is tracked correctly
- Multiple rounds of feedback work smoothly

---

## Milestone 8: Information Sinks System

**Goal**: Implement external knowledge management via sinks

**Key Deliverables**:
- Create sink index schema and interrogation logic
- Build CLI commands (`sow sinks install`, `update`, `list`, `remove`)
- Implement sink-to-task routing logic in orchestrator
- Create LLM-based sink summarization
- Build sink update mechanism

**References**:
- [ARCHITECTURE.md](./ARCHITECTURE.md#information-sinks) - Sink architecture
- [USER_GUIDE.md](./USER_GUIDE.md#working-with-sinks) - Sink workflows
- [CLI_REFERENCE.md](./CLI_REFERENCE.md#sink-management) - CLI commands

**Success Criteria**:
- Sinks can be installed from git repositories
- Orchestrator routes relevant sinks to workers
- Sink index is maintained automatically
- Sinks can be updated independently

---

## Milestone 9: Repository Linking

**Goal**: Enable multi-repo context for agents

**Key Deliverables**:
- Create repository index schema
- Build CLI commands (`sow repos add`, `sync`, `list`, `remove`)
- Implement clone and symlink support
- Create repository-to-task reference system
- Build sync mechanism for updates

**References**:
- [ARCHITECTURE.md](./ARCHITECTURE.md#multi-repo-strategy) - Multi-repo support
- [USER_GUIDE.md](./USER_GUIDE.md#working-with-linked-repos) - Repo workflows
- [CLI_REFERENCE.md](./CLI_REFERENCE.md#repository-management) - CLI commands

**Success Criteria**:
- External repositories can be linked
- Workers can reference code from linked repos
- Repositories sync correctly
- Supports both monorepo and multi-repo setups

---

## Milestone 10: Workflow Commands

**Goal**: Build user-facing slash commands for project lifecycle

**Key Deliverables**:
- Implement `/init` command (bootstrap repository)
- Implement `/start-project` command (create project with planning)
- Implement `/continue` command (resume existing project)
- Implement `/cleanup` command (delete project before merge)
- Implement `/sync` command (update sinks and repos)
- Build smart UX flows (branch protection, conflict resolution)

**References**:
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#user-workflow-commands) - Command specifications
- [USER_GUIDE.md](./USER_GUIDE.md) - User workflows

**Success Criteria**:
- All workflow commands are functional
- Smart error handling and prompts work correctly
- Users can complete full project lifecycle
- Branch constraints are enforced

---

## Milestone 11: Skills System

**Goal**: Create reusable capabilities for worker agents

**Key Deliverables**:
- Design skills as slash commands (avoid separate abstraction)
- Create architect skills (`/create-adr`, `/design-doc`)
- Create implementer skills (`/implement-feature`, `/fix-bug`)
- Create integration-tester skills (`/write-integration-tests`)
- Create reviewer skills (`/review-code`)
- Create documenter skills (`/update-docs`)
- Build skill invocation in agent prompts

**References**:
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#skills-system) - Skills architecture
- [ARCHITECTURE.md](./ARCHITECTURE.md#skills--slash-commands) - Design decision

**Success Criteria**:
- Workers can invoke skills as needed
- Skills are reusable across agents
- No context window bloat from skill definitions
- Skills follow consistent format

---

## Milestone 12: Hooks System

**Goal**: Enable event-driven automation and customization

**Key Deliverables**:
- Create hooks configuration schema (`hooks.json`)
- Implement SessionStart hook (version checking, project status)
- Create `sow session-info` CLI command for hook
- Implement PostToolUse hooks (auto-formatting)
- Implement PreCompact hooks (context preservation)
- Build hook execution engine

**References**:
- [HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md#hooks-system) - Hooks documentation
- [DISTRIBUTION.md](./DISTRIBUTION.md#version-check-implementation) - SessionStart hook

**Success Criteria**:
- SessionStart hook detects version mismatches
- Hooks run at correct lifecycle points
- Custom hooks can be added by users
- Hook execution is secure and sandboxed

---

## Milestone 13: Migration System

**Goal**: Enable smooth version upgrades with automated migrations

**Key Deliverables**:
- Define migration file format (markdown specifications)
- Create `/migrate` command implementation
- Build sequential migration chain execution
- Implement version detection logic
- Create migration templates for breaking changes
- Build rollback procedures

**References**:
- [DISTRIBUTION.md](./DISTRIBUTION.md#migration-system) - Migration architecture
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#migrate) - Migration command

**Success Criteria**:
- Migrations can be applied automatically
- Sequential migrations work for version skipping
- Version mismatches are detected on SessionStart
- Rollback is straightforward

---

## Milestone 14: CLI Development

**Goal**: Build optional but high-performance CLI binary

**Key Deliverables**:
- Implement core CLI commands (log, validate, session-info)
- Build sink management commands
- Build repository management commands
- Create cross-platform binaries (macOS, Linux, Windows)
- Implement auto-detection for task vs project context
- Build fast logging performance (<1s)

**References**:
- [CLI_REFERENCE.md](./CLI_REFERENCE.md) - Complete CLI documentation
- [DISTRIBUTION.md](./DISTRIBUTION.md#cli-distribution) - CLI distribution

**Success Criteria**:
- CLI is significantly faster than file editing
- All documented commands are implemented
- Binaries work on all platforms
- Version alignment with plugin is maintained

---

## Milestone 15: Plugin Packaging and Distribution

**Goal**: Package system as Claude Code Plugin for distribution

**Key Deliverables**:
- Create plugin metadata (`plugin.json`)
- Structure `.claude-plugin/` directory
- Build plugin installation workflow
- Create marketplace listing
- Implement version tracking system
- Create release automation

**References**:
- [DISTRIBUTION.md](./DISTRIBUTION.md) - Complete distribution guide
- [ARCHITECTURE.md](./ARCHITECTURE.md#two-layer-architecture) - Distribution model

**Success Criteria**:
- Plugin can be installed via marketplace
- Version management works correctly
- Updates are straightforward
- Installation is documented clearly

---

## Milestone 16: MCP Integrations (Optional)

**Goal**: Enable external tool and service integrations

**Key Deliverables**:
- Define MCP configuration schema (`mcp.json`)
- Document common integrations (GitHub, Jira, etc.)
- Create integration examples
- Build authentication support
- Test with popular MCP servers

**References**:
- [HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md#mcp-integrations) - MCP documentation

**Success Criteria**:
- MCP servers can be configured
- Authentication mechanisms work
- Common integrations are documented
- Security considerations are addressed

---

## Milestone 17: Documentation and Examples

**Goal**: Create comprehensive user and developer documentation

**Key Deliverables**:
- Finalize all architecture documents
- Create getting started guide
- Build example projects and workflows
- Create video tutorials or demos
- Write troubleshooting guides
- Create contribution guidelines

**References**:
- All existing architecture documents in `docs/`

**Success Criteria**:
- New users can get started quickly
- All features are documented
- Common issues have solutions
- Examples demonstrate key workflows

---

## Milestone 18: Testing and Quality Assurance

**Goal**: Ensure system reliability and correctness

**Key Deliverables**:
- Create test suite for CLI commands
- Build integration tests for full workflows
- Test migration paths thoroughly
- Validate schema compliance
- Test cross-platform compatibility
- Create CI/CD pipeline

**Success Criteria**:
- Core functionality has test coverage
- Migrations are tested on real repositories
- CLI works on all platforms
- CI enforces quality standards

---

## Milestone 19: Polish and User Experience

**Goal**: Refine the system for production readiness

**Key Deliverables**:
- Improve error messages and user prompts
- Add helpful hints and suggestions
- Optimize performance bottlenecks
- Polish command outputs and formatting
- Create consistent visual language
- Add progress indicators for long operations

**Success Criteria**:
- User experience is smooth and intuitive
- Error messages are helpful
- Performance is acceptable
- Visual consistency across commands

---

## Milestone 20: Initial Release (v0.1.0)

**Goal**: Launch the first public version of sow

**Key Deliverables**:
- Package complete plugin and CLI
- Create GitHub releases
- Publish to marketplace
- Announce to community
- Gather initial feedback
- Create feedback channels (issues, discussions)

**Success Criteria**:
- System is installable and functional
- Documentation is complete
- Users can accomplish basic workflows
- Feedback mechanism is in place
- Known issues are documented

---

## Post-Release: Iteration and Evolution

**Ongoing Activities**:
- Gather user feedback and iterate
- Fix bugs and address issues
- Add new features based on needs
- Improve documentation based on questions
- Optimize performance
- Expand integration ecosystem
- Build community of contributors

**Success Metrics**:
- User adoption and retention
- Quality of feedback and contributions
- System stability and reliability
- Time-to-value for new users

---

## Dependencies and Constraints

**Critical Dependencies**:
- Claude Code plugin system and APIs
- Claude model capabilities for agent system
- Git for version control and state management

**Key Constraints**:
- One project per branch (architectural constraint)
- Plugin distributed via Claude Code marketplace
- Must maintain zero-context resumability
- Progressive planning (not waterfall)

**Flexibility Points**:
- CLI is optional (system works without it)
- MCP integrations are optional
- Sinks and repo linking are optional features
- Skills can be extended by users

---

## Success Criteria for Overall Project

The sow system will be considered successful when:

1. **Functional Completeness**: All documented features are implemented and working
2. **Zero-Context Resumability**: Projects can be resumed without conversation history
3. **Multi-Agent Coordination**: Orchestrator and workers collaborate effectively
4. **Progressive Planning**: Projects start minimal and adapt as needed
5. **User Adoption**: Developers can get started and accomplish work within 30 minutes
6. **Documentation Quality**: Users can find answers without external help
7. **Version Stability**: Migrations work smoothly for existing users
8. **Performance**: Operations complete in reasonable time (<5s for most commands)
9. **Reliability**: System handles errors gracefully and provides recovery paths
10. **Extensibility**: Users can customize with hooks, skills, and sinks

---

## Notes

- This roadmap is preliminary and will evolve based on implementation learnings
- Milestones may be reordered or combined based on dependencies discovered during development
- Each milestone should be reviewed and adjusted before beginning work
- User feedback after initial release will heavily influence future direction
- Some milestones may be delivered in parallel by different developers
