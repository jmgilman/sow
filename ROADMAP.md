# sow Development Roadmap

**Last Updated**: 2025-10-13
**Status**: Simplified roadmap with 10 project-sized milestones

---

## Overview

This roadmap outlines 10 sequential milestones for building the sow (system of work) framework. Each milestone represents roughly a "project's worth" of work - a substantial chunk that delivers meaningful functionality.

The milestones are organized to:
- Build foundational infrastructure first (CLI, schemas, file structure)
- Establish the agent system (orchestrator and workers)
- Add advanced features (feedback, skills, external knowledge)
- Polish and release

For detailed architectural decisions and technical specifications, see the comprehensive architecture documentation in this directory.

---

## Milestone 1: CLI Foundation & Schema System

**Goal**: Build complete CLI with embedded CUE schemas as the foundation for all sow operations

**Key Deliverables**:
- Set up Go project structure for CLI
- **Define CUE schemas for all state files** (project state, task state, sink index, repo index, version file)
- Add CUE constraints and validation rules to schemas
- Implement Go CLI binary with CUE schema embedding using `go:embed` directive
- Build cross-platform binaries (macOS, Linux, Windows)
- Implement version alignment system (CLI version = schema version)
- Create all CLI commands:
  - `sow init` - Initialize `.sow/` structure from embedded schemas
  - `sow validate` - Validate files against embedded CUE schemas
  - `sow schema` - Inspect embedded schemas
  - `sow log` - Fast logging for agents (<1s)
  - `sow session-info` - Session status for hooks
  - `sow sinks` - Sink management (install, update, list, remove)
  - `sow repos` - Repository management (add, sync, list, remove)
- Implement auto-detection for task vs project context
- Create installation guides for all platforms

**Rationale**: Build complete CLI first to dogfood it during sow development. This provides immediate benefits (fast validation, initialization, logging) while implementing the rest of the system.

**References**:
- [CLI_REFERENCE.md](./CLI_REFERENCE.md) - Complete CLI documentation
- [DISTRIBUTION.md](./DISTRIBUTION.md#cli-distribution) - CLI distribution strategy
- [ARCHITECTURE.md](./ARCHITECTURE.md#schema-management) - CUE-based design
- [SCHEMAS.md](./SCHEMAS.md) - File format specifications

**Success Criteria**:
- CUE schemas are defined with proper constraints and validation rules
- CLI successfully embeds all CUE schemas at build time
- All CLI commands are functional and performant
- CLI binaries work on macOS, Linux, and Windows
- CLI version matches schema version automatically
- Installation is straightforward on all platforms
- CLI can be used immediately for remaining milestones

---

## Milestone 2: Core Infrastructure & File Structure

**Goal**: Establish the plugin architecture and file structure templates

**Key Deliverables**:
- Define and implement the two-layer architecture (execution + data layers)
- Create file structure templates for `plugin/` (becomes `.claude/` on install) and `.sow/` directories
- Implement version tracking system (`.sow/.version` and `plugin/.plugin-version`)
- Create markdown reference documentation from CUE schemas (built in Milestone 1)
- Use CLI from Milestone 1 for validation during development

**Note**: The execution layer is developed in the `plugin/` directory of the marketplace repository. When users install the plugin via `/plugin install sow@sow-marketplace`, the contents of `plugin/` are copied into their repository as `.claude/`. CUE schemas are already defined in the CLI (Milestone 1).

**References**:
- [FILE_STRUCTURE.md](./FILE_STRUCTURE.md) - Complete directory layout
- [SCHEMAS.md](./SCHEMAS.md) - File format specifications (CUE as source of truth)
- [ARCHITECTURE.md](./ARCHITECTURE.md#two-layer-architecture) - Architectural foundation

**Success Criteria**:
- File structure templates are complete and documented
- Version tracking system works correctly
- Plugin directory structure is ready for marketplace distribution
- Markdown documentation accurately reflects CUE schemas from CLI
- Can use `sow validate` to verify all template files

---

## Milestone 3: Orchestrator, Project Management & Logging

**Goal**: Build the orchestrator agent and complete project management system with structured logging

**Key Deliverables**:
- Create orchestrator agent with system prompt
- Implement startup behavior (SessionStart detection)
- Build command routing (one-off vs project-based work)
- Implement project initialization logic (`/start-project`)
- Build complexity assessment (1-3 rating)
- Create progressive planning system (start with 1-2 phases)
- Implement phase management (creation, transitions, validation)
- Build gap numbering system for tasks
- Create task state management (pending, in_progress, completed, abandoned)
- Define structured markdown log format
- Create action vocabulary for logging
- Implement agent ID construction (role + iteration)
- Integrate CLI logging (`sow log`) into agents

**References**:
- [AGENTS.md](./AGENTS.md#orchestrator) - Orchestrator specifications
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md) - Complete project lifecycle
- [ARCHITECTURE.md](./ARCHITECTURE.md#progressive-planning-philosophy) - Planning approach

**Success Criteria**:
- Orchestrator can start up and assess work mode
- Projects can be created with initial phases
- Tasks use gap numbering correctly
- Phase transitions work as expected
- State persists correctly across sessions
- Logs are created with consistent format via CLI
- Agent IDs are constructed correctly

---

## Milestone 4: Worker Agents & Zero-Context Resumability

**Goal**: Build all specialized worker agents with complete context compilation and resumability system

**Key Deliverables**:
- Create architect agent (design and ADRs)
- Create implementer agent (TDD-focused coding)
- Create integration-tester agent (E2E and integration tests)
- Create reviewer agent (code quality)
- Create documenter agent (documentation updates)
- Implement agent spawning mechanism via Task tool
- Build agent assignment logic
- Build context compilation system in orchestrator
- Implement task description format and generation
- Create reference tracking system (sinks, knowledge, repos)
- Build worker recovery process (read state → execute → report)
- Implement iteration tracking system

**References**:
- [AGENTS.md](./AGENTS.md#worker-agents) - Worker agent specifications
- [ARCHITECTURE.md](./ARCHITECTURE.md#zero-context-resumability) - Resumability design
- [ARCHITECTURE.md](./ARCHITECTURE.md#orchestrator--worker-pattern) - Agent coordination

**Success Criteria**:
- All five worker agents are functional
- Orchestrator can spawn workers correctly
- Workers can execute tasks independently
- Task assignment works based on task type
- Workers can resume tasks from filesystem state alone
- No dependency on conversation history
- Context is appropriately filtered for each worker
- Iteration counter tracks attempts correctly

---

## Milestone 5: Feedback Mechanism & Skills System

**Goal**: Enable human-in-the-loop corrections and create reusable agent capabilities

**Key Deliverables**:
- Create feedback file structure and format
- Implement feedback creation by orchestrator
- Build feedback reading in workers
- Create feedback status tracking (pending, addressed, superseded)
- Integrate feedback with iteration system
- Design skills as slash commands (avoid separate abstraction)
- Create architect skills (`/create-adr`, `/design-doc`)
- Create implementer skills (`/implement-feature`, `/fix-bug`)
- Create integration-tester skills (`/write-integration-tests`)
- Create reviewer skills (`/review-code`)
- Create documenter skills (`/update-docs`)
- Build skill invocation in agent prompts

**References**:
- [PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md#feedback-mechanism) - Feedback system
- [USER_GUIDE.md](./USER_GUIDE.md#providing-feedback-to-agents) - User experience
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#skills-system) - Skills architecture
- [ARCHITECTURE.md](./ARCHITECTURE.md#skills--slash-commands) - Design decision

**Success Criteria**:
- Users can provide corrections to agents
- Workers read and incorporate feedback
- Feedback status is tracked correctly
- Multiple rounds of feedback work smoothly
- Workers can invoke skills as needed
- Skills are reusable across agents
- No context window bloat from skill definitions
- Skills follow consistent format

---

## Milestone 6: External Knowledge & Repository Linking

**Goal**: Implement information sinks and multi-repo context management

**Key Deliverables**:
- Create sink index schema and interrogation logic
- Implement sink-to-task routing logic in orchestrator
- Create LLM-based sink summarization
- Build sink update mechanism
- Create repository index schema
- Implement clone and symlink support
- Create repository-to-task reference system
- Build sync mechanism for updates
- Integrate sink and repo commands (already built in CLI from Milestone 1)

**References**:
- [ARCHITECTURE.md](./ARCHITECTURE.md#information-sinks) - Sink architecture
- [ARCHITECTURE.md](./ARCHITECTURE.md#multi-repo-strategy) - Multi-repo support
- [USER_GUIDE.md](./USER_GUIDE.md#working-with-sinks) - Sink workflows
- [USER_GUIDE.md](./USER_GUIDE.md#working-with-linked-repos) - Repo workflows

**Success Criteria**:
- Sinks can be installed from git repositories
- Orchestrator routes relevant sinks to workers
- Sink index is maintained automatically
- Sinks can be updated independently
- External repositories can be linked
- Workers can reference code from linked repos
- Repositories sync correctly
- Supports both monorepo and multi-repo setups

---

## Milestone 7: Workflow Commands & Hooks System

**Goal**: Build user-facing slash commands and event-driven automation system

**Key Deliverables**:
- Implement `/init` command (bootstrap repository)
- Implement `/start-project` command (create project with planning)
- Implement `/continue` command (resume existing project)
- Implement `/cleanup` command (delete project before merge)
- Implement `/sync` command (update sinks and repos)
- Build smart UX flows (branch protection, conflict resolution)
- Create hooks configuration schema (`hooks.json`)
- Implement SessionStart hook (version checking, project status)
- Implement PostToolUse hooks (auto-formatting)
- Implement PreCompact hooks (context preservation)
- Build hook execution engine
- Integrate `sow session-info` CLI command (already built in Milestone 1)

**References**:
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#user-workflow-commands) - Command specifications
- [USER_GUIDE.md](./USER_GUIDE.md) - User workflows
- [HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md#hooks-system) - Hooks documentation
- [DISTRIBUTION.md](./DISTRIBUTION.md#version-check-implementation) - SessionStart hook

**Success Criteria**:
- All workflow commands are functional
- Smart error handling and prompts work correctly
- Users can complete full project lifecycle
- Branch constraints are enforced
- SessionStart hook detects version mismatches
- Hooks run at correct lifecycle points
- Custom hooks can be added by users
- Hook execution is secure and sandboxed

---

## Milestone 8: Plugin Distribution & Migration System

**Goal**: Package system for distribution and enable version upgrades

**Key Deliverables**:
- Create plugin metadata (`plugin.json`)
- Structure `.claude-plugin/` directory
- Build plugin installation workflow
- Create marketplace listing
- Implement version tracking system
- Create release automation
- Define migration file format (markdown specifications)
- Create `/migrate` command implementation
- Build sequential migration chain execution
- Implement version detection logic
- Create migration templates for breaking changes
- Build rollback procedures

**References**:
- [DISTRIBUTION.md](./DISTRIBUTION.md) - Complete distribution guide
- [DISTRIBUTION.md](./DISTRIBUTION.md#migration-system) - Migration architecture
- [ARCHITECTURE.md](./ARCHITECTURE.md#two-layer-architecture) - Distribution model
- [COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md#migrate) - Migration command

**Success Criteria**:
- Plugin can be installed via marketplace
- Version management works correctly
- Updates are straightforward
- Installation is documented clearly
- Migrations can be applied automatically
- Sequential migrations work for version skipping
- Version mismatches are detected on SessionStart
- Rollback is straightforward

---

## Milestone 9: Testing, Documentation & Polish

**Goal**: Ensure system reliability, create comprehensive documentation, and refine user experience

**Key Deliverables**:
- Create test suite for CLI commands
- Build integration tests for full workflows
- Test migration paths thoroughly
- Validate schema compliance
- Test cross-platform compatibility
- Create CI/CD pipeline
- Finalize all architecture documents
- Create getting started guide
- Build example projects and workflows
- Create video tutorials or demos
- Write troubleshooting guides
- Create contribution guidelines
- Improve error messages and user prompts
- Add helpful hints and suggestions
- Optimize performance bottlenecks
- Polish command outputs and formatting
- Create consistent visual language
- Add progress indicators for long operations

**References**:
- All existing architecture documents in `docs/`

**Success Criteria**:
- Core functionality has test coverage
- Migrations are tested on real repositories
- CLI works on all platforms
- CI enforces quality standards
- New users can get started quickly
- All features are documented
- Common issues have solutions
- Examples demonstrate key workflows
- User experience is smooth and intuitive
- Error messages are helpful
- Performance is acceptable (<5s for most commands)
- Visual consistency across commands

---

## Milestone 10: Initial Release & MCP Integrations

**Goal**: Launch first public version of sow with optional external integrations

**Key Deliverables**:
- Package complete plugin and CLI
- Create GitHub releases
- Publish to marketplace
- Announce to community
- Gather initial feedback
- Create feedback channels (issues, discussions)
- Define MCP configuration schema (`mcp.json`) - optional
- Document common integrations (GitHub, Jira, etc.) - optional
- Create integration examples - optional
- Build authentication support - optional
- Test with popular MCP servers - optional

**References**:
- [HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md#mcp-integrations) - MCP documentation

**Success Criteria**:
- System is installable and functional
- Documentation is complete
- Users can accomplish basic workflows
- Feedback mechanism is in place
- Known issues are documented
- MCP servers can be configured (if implemented)
- Authentication mechanisms work (if implemented)
- Common integrations are documented (if implemented)
- Security considerations are addressed

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
- MCP integrations are optional
- Sinks and repo linking are optional features
- Skills can be extended by users
- Hooks can be customized for team workflows

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
