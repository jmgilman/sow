# System of Work (sow)

An AI-powered framework for software engineering that provides a structured, opinionated approach to building software with AI assistance.

---

## What is sow?

`sow` leverages Claude Code's capabilities to create a unified development experience across projects through:

- **Multi-agent orchestration** - Specialized AI agents (architect, implementer, tester, reviewer, documenter) working together
- **Progressive planning** - Adaptive project management that evolves with your work, not rigid waterfall
- **Knowledge integration** - Contextual information from style guides, procedures, and linked repositories
- **Zero-context resumability** - Pick up where you left off, anytime, without conversation history

---

## Quick Start

```bash
# Install Claude Code Plugin
/plugin install sow@sow-marketplace

# Restart Claude Code
exit && claude

# Initialize in your repository
/init

# Start your first project
git checkout -b feat/my-feature
/start-project "Add new feature"

# Let the orchestrator coordinate the work
# When ready to merge
/cleanup
git commit -m "chore: cleanup sow project state"
```

---

## Core Concepts

### Two-Layer Architecture

- **Execution Layer** (`.claude/`) - AI agents, commands, hooks (distributed via plugin)
- **Data Layer** (`.sow/`) - Project knowledge, external context, active work state

### Multi-Agent System

- **Orchestrator** - Your main interface, coordinates all work
- **Workers** - Specialists (architect, implementer, integration-tester, reviewer, documenter)
- Each worker receives only the context they need

### Progressive Planning

- Start with minimal structure (1-2 phases)
- Add complexity as you discover needs
- Human approval gates for significant changes
- Fail-forward approach (add tasks, don't revert)

### One Project Per Branch

- Enforced constraint: one active project per repository branch
- Project state committed to feature branches
- Deleted before merging (CI enforced)
- Switch branches = switch project context (automatic)

---

## Documentation

### Getting Started

📖 **[OVERVIEW](./docs/OVERVIEW.md)** - Complete introduction to sow concepts and terminology

📖 **[USER_GUIDE](./docs/USER_GUIDE.md)** - Installation, daily workflows, and troubleshooting

### Understanding the System

📖 **[ARCHITECTURE](./docs/ARCHITECTURE.md)** - Design decisions, patterns, and architectural philosophy

📖 **[FILE_STRUCTURE](./docs/FILE_STRUCTURE.md)** - Complete directory layout and organization

📖 **[AGENTS](./docs/AGENTS.md)** - Multi-agent system, roles, and coordination

### Using sow

📖 **[COMMANDS_AND_SKILLS](./docs/COMMANDS_AND_SKILLS.md)** - Slash commands and agent skills reference

📖 **[PROJECT_MANAGEMENT](./docs/PROJECT_MANAGEMENT.md)** - Project lifecycle, phases, tasks, and logging

📖 **[HOOKS_AND_INTEGRATIONS](./docs/HOOKS_AND_INTEGRATIONS.md)** - Event automation and external tool integrations

### Publishing and Maintaining

📖 **[DISTRIBUTION](./docs/DISTRIBUTION.md)** - Packaging, versioning, and upgrade workflows

### Reference

📖 **[SCHEMAS](./docs/SCHEMAS.md)** - File format specifications (state.yaml, indexes, etc.)

📖 **[CLI_REFERENCE](./docs/CLI_REFERENCE.md)** - CLI command documentation

📖 **[Claude Code Features](./research/CLAUDE_CODE.md)** - Claude Code capability reference

---

## Key Features

### For Developers

✅ **Consistent experience** across all projects
✅ **Clear structure** for complex work
✅ **Resume anytime** without losing context
✅ **Specialized expertise** via worker agents
✅ **Progressive discovery** instead of upfront planning

### For Teams

✅ **Shared conventions** via committed execution layer
✅ **Knowledge sharing** via information sinks
✅ **Collaboration** through committed project state on branches
✅ **Quality gates** through agent specialization
✅ **Reduced onboarding** with standardized structure

### For AI Agents

✅ **Focused context** - workers receive only relevant information
✅ **Clear roles** - each agent has specific expertise
✅ **Structured state** - can resume without conversation history
✅ **Fail-forward** - can adapt to discoveries and changes
✅ **Audit trails** - logs track all actions

---

## Example Workflow

```bash
# 1. Create feature branch
git checkout -b feat/add-authentication

# 2. Start project
/start-project "Add JWT authentication"

# Orchestrator:
# - Assesses complexity: "moderate"
# - Selects phases: design, implement, test
# - Creates tasks with assigned agents
# - Spawns architect for design task

# 3. Continue working
/continue

# Orchestrator:
# - Reads project state
# - Spawns next worker (implementer)
# - Worker writes code with TDD
# - Updates state

# 4. Provide feedback if needed
"The JWT service should use RS256, not HS256"

# Orchestrator creates feedback, respawns worker

# 5. Complete and clean up
/cleanup
git commit -m "chore: cleanup sow project state"
git push

# 6. Create PR and merge
# CI verifies no .sow/project/ exists
```

---

## Philosophy

### Opinionated but Flexible

**Strong opinions**:
- One project per branch (enforced)
- Specific phase vocabulary
- Structured logging format
- Multi-agent architecture

**Flexible where it matters**:
- Progressive planning (not waterfall)
- Extensible through plugins
- Customizable knowledge (sinks)
- Optional CLI enhancements

### Not a Replacement

`sow` complements existing tools:

**Not replacing**: JIRA, Linear, GitHub Issues, Sprint planning, Team chat

**sow is for**: Active work coordination, AI agent orchestration, Short-term execution (hours to days)

---

## Project Status

**Current State**: Comprehensive architecture design complete, implementation in progress

**Documentation**: Complete (you're looking at it!)

**Next Steps**:
1. Build Claude Code Plugin
2. Implement CLI
3. Create example projects
4. Beta testing

---

## Repository Structure

```
sow/
├── docs/                    # Comprehensive documentation
│   ├── OVERVIEW.md
│   ├── ARCHITECTURE.md
│   ├── FILE_STRUCTURE.md
│   ├── AGENTS.md
│   ├── COMMANDS_AND_SKILLS.md
│   ├── HOOKS_AND_INTEGRATIONS.md
│   ├── USER_GUIDE.md
│   ├── PROJECT_MANAGEMENT.md
│   ├── DISTRIBUTION.md
│   ├── SCHEMAS.md
│   └── CLI_REFERENCE.md
├── research/                # Research and feature references
│   └── CLAUDE_CODE.md
└── archive/                 # Historical discovery documents
    ├── BRAINSTORMING.md
    ├── PROJECT_LIFECYCLE.md
    ├── FS_STRUCTURE.md
    ├── EXECUTION.md
    └── DISTRIBUTION.md
```

---

## Contributing

This project is in early development. Design feedback and contributions welcome!

**Areas of interest**:
- Plugin implementation
- CLI development
- Example projects
- Documentation improvements
- Testing and validation

---

## License

MIT (placeholder - to be determined)

---

## Learn More

Start with **[docs/OVERVIEW.md](./docs/OVERVIEW.md)** for a comprehensive introduction, then explore the other documentation based on your needs.

**For users**: [USER_GUIDE.md](./docs/USER_GUIDE.md)
**For contributors**: [ARCHITECTURE.md](./docs/ARCHITECTURE.md)
**For reference**: [SCHEMAS.md](./docs/SCHEMAS.md)
