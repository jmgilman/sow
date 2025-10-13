# Repository Index

**Generated**: 2025-10-13

---

## Quick Start

**New to this repository?**
- `README.md` - Project overview
- `docs/OVERVIEW.md` - System architecture and concepts
- `docs/USER_GUIDE.md` - Getting started guide

**Starting a project?**
- `ROADMAP.md` - Planned features and milestones
- `.sow/project/` - Active project workspace (milestone-1-foundation)

---

## Knowledge

**`docs/`** - Core system documentation (architecture, CLI reference, schemas, user guides)
**`.sow/knowledge/`** - Repository-specific architecture decisions and design docs (empty - ADRs will go here)
**`.sow/sinks/`** - External knowledge sources like style guides (not present yet)
**`research/`** - Investigation notes (Claude Code research)
**`archive/`** - Historical design documents and brainstorming

---

## Key Directories

```
sow/
├── plugin/                    # Distributable Claude Code plugin
│   ├── agents/                # Agent definitions (empty templates)
│   ├── commands/              # Slash commands and skills
│   └── .claude-plugin/        # Plugin metadata
│
├── .claude/                   # Development execution layer
│   ├── agents/                # Architect, implementer agents
│   └── commands/              # Active command implementations
│
├── .sow/                      # Data layer
│   ├── project/               # Active project state (milestone-1)
│   │   ├── state.yaml         # Project coordination
│   │   ├── log.md             # Orchestrator actions
│   │   ├── context/           # Project learnings
│   │   └── phases/            # Design and implement phases
│   └── knowledge/             # ADRs and design docs (empty)
│
├── docs/                      # System documentation
├── schemas/templates/         # YAML templates for state files
└── README.md                  # Project overview
```

---

## Common Tasks

**Working on the sow system itself?** Check active tasks in `.sow/project/phases/`
**Adding new commands?** See `docs/COMMANDS_AND_SKILLS.md` and `.claude/commands/`
**Creating architecture docs?** Use `/create-adr` or `/design-doc` commands
**Understanding agent roles?** Read `docs/AGENTS.md` and `.claude/CLAUDE.md`
**Need schema templates?** Check `schemas/templates/` for YAML structure
