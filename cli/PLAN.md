# Exploration Mode Implementation Plan

## Overview

Implement `sow explore` end-to-end to validate the exploration mode pattern: smart branch management, index-based file discovery, CLI-enforced structure, and guided agent prompts.

---

## Phase 1: Schemas (Foundation)

**Why first:** Everything depends on these data structures.

### Tasks

1. **Create exploration index schema** (`cli/schemas/exploration_index.cue`)
   - Define structure:
     ```cue
     exploration: {
       topic: string
       branch: string
       created_at: time.Time
       status: "active" | "completed" | "abandoned"
     }
     files: [...{
       path: string
       description: string
       tags: [...string]
       created_at: time.Time
     }]
     ```
   - Follows existing schema patterns in codebase

2. **Create knowledge index schema** (`cli/schemas/knowledge_index.cue`)
   - Define structure for `.sow/knowledge/index.yaml`
   - Track exploration summaries, ADR references, design doc references

3. **Update config schema** (`cli/schemas/config.cue` or create if missing)
   - Add artifacts section:
     ```cue
     artifacts?: {
       adrs?: string
       design_docs?: string
     }
     ```

4. **Generate Go types**
   - Run CUE code generation (`go generate` or similar)
   - Verify types compile

**Deliverable:** Schemas defined, Go types generated

---

## Phase 2: Core CLI Commands

**Why next:** Need functionality before we can guide agents to use it.

### Tasks

5. **Create exploration package** (`cli/internal/exploration/`)
   - `index.go` - Load/save/validate exploration index
   - `manager.go` - High-level operations (add/update/remove files)
   - Use existing `sow.Context` pattern from standard projects

6. **Implement index management commands** (`cli/cmd/exploration/`)
   - `add_file.go` - `sow exploration add-file <path> --description "..." --tags "..."`
   - `update_file.go` - `sow exploration update-file <path> --description "..."`
   - `remove_file.go` - `sow exploration remove-file <path>`
   - `index.go` - `sow exploration index` (display current index)
   - `exploration.go` - Root command setup
   - Follow existing command patterns (e.g., `cli/cmd/agent/`)

7. **Implement explore launcher** (`cli/cmd/explore.go`)
   - Parse topic argument
   - Generate branch name: `explore/{topic-kebab}`
   - Smart branch resolution:
     1. Check if local branch exists → checkout
     2. Check if remote branch exists → fetch, create local, checkout
     3. Create new branch if neither
   - Initialize `.sow/exploration/` directory structure
   - Create initial `index.yaml` if new exploration
   - Launch `claude` CLI with exploration mode prompt
   - Use existing patterns from `cli/cmd/new.go` and `cli/cmd/start.go`

**Deliverable:** Working CLI commands for exploration management

---

## Phase 3: Prompts & Guidance

**Why next:** CLI works, now guide agents on how to use it.

### Tasks

8. **Create exploration mode prompt** (`cli/internal/prompts/templates/modes/explore.md`)
   - Context injection: topic, existing files from index
   - Explain role: research partner
   - Document directory structure (`.sow/exploration/`)
   - List available CLI commands for index management
   - Mention available guidance prompts (`sow prompt research`)
   - Use existing template patterns from `templates/commands/`

9. **Create research guidance prompt** (`cli/internal/prompts/templates/guidance/research.md`)
   - Research methodology best practices
   - How to document findings
   - When to create new files vs. update existing
   - File naming conventions
   - Keep context window small (focused research)

10. **Implement prompt command** (`cli/cmd/prompt.go`)
    - `sow prompt <type>` command
    - Read template from `templates/guidance/{type}.md`
    - Perform context injection (e.g., list existing files, current branch)
    - Output to stdout for agent consumption
    - Use existing prompt rendering from `cli/internal/prompts/`

**Deliverable:** Agent prompts and guidance system working

---

## Phase 4: Configuration

**Why now:** Needed for team artifact locations, but not blocking earlier work.

### Tasks

11. **Update init command** (`cli/cmd/init.go`)
    - Include `.sow/config.yaml` template with commented artifact locations:
      ```yaml
      # Artifact locations (uncomment and customize for your team)
      # artifacts:
      #   adrs: "docs/adrs"
      #   design_docs: "docs/architecture"
      ```
    - Create `.sow/knowledge/explorations/` directory
    - Create initial knowledge index

12. **Implement config loading**
    - Update config loading logic (likely in `cli/internal/sow/`)
    - Provide defaults: `.sow/knowledge/adrs/`, `.sow/knowledge/design/`
    - Make available to agents via context

**Deliverable:** Configuration system for team conventions

---

## Phase 5: Skills (Claude-Specific Enhancement)

**Why last:** Nice-to-have, not required for core functionality.

### Tasks

13. **Create research skill** (`.claude/skills/research.md`)
    - Skill content: Invokes `sow prompt research`
    - Follow existing skill patterns in `.claude/skills/`
    - Enables Claude to auto-invoke research guidance

14. **Create additional skills** (optional, as needed)
    - ADR creation, design doc creation, etc.
    - Can be added incrementally

**Deliverable:** Claude skills for improved discoverability

---

## Phase 6: Testing & Validation

**Why final:** Validate the complete flow works end-to-end.

### Tasks

15. **Manual end-to-end test**
    - Run `sow explore "authentication-approaches"`
    - Verify branch creation (`explore/authentication-approaches`)
    - Verify `.sow/exploration/` initialization
    - Verify Claude launches with correct prompt
    - Create files via Claude
    - Run index commands (`sow exploration add-file`, etc.)
    - Verify index updates correctly
    - Exit and resume: `sow explore --branch explore/authentication-approaches`
    - Verify context restoration

16. **Write unit tests**
    - Index management (add/update/remove file operations)
    - Branch resolution logic
    - Schema validation
    - Use existing test patterns from `cli/internal/project/`

**Deliverable:** Validated, tested exploration mode

---

## Success Criteria

The spike is complete when:

- ✅ User can run `sow explore "topic"` and get into exploration mode
- ✅ Branch is created/checked out automatically (`explore/{topic}`)
- ✅ Claude launches with appropriate exploration prompt
- ✅ Agent can create files and register them via CLI
- ✅ Index is maintained correctly with descriptions/tags
- ✅ User can resume exploration on existing branch
- ✅ Prompts guide agent on best practices
- ✅ Skills (Claude-specific) enable auto-invocation of guidance

## What We Learn

This spike will validate:
- Is the index pattern effective for context engineering?
- Do agents use CLI commands correctly?
- Is the guidance prompt strategy helpful?
- Does branch-per-exploration feel natural?
- What friction points emerge?

Based on learnings, we can refine before expanding to other modes (design, breakdown, etc.).

---

## Design Decisions

### File Structure
```
.sow/
├── config.yaml              # Artifact locations (user configures)
├── knowledge/
│   ├── index.yaml          # Index of permanent knowledge
│   └── explorations/       # Summaries of completed explorations
│       └── index.yaml
├── exploration/            # Active (single, on branch)
│   ├── index.yaml         # Auto-managed via CLI
│   └── (loose files)
```

### Branch Convention
- **Enforced**: `explore/{topic}` (e.g., `explore/auth-research`)
- Smart resume: local → remote → error

### Index Management
- **CLI enforced** for structure/validation
- Required fields: `path`, `description`, `tags`, `created_at`

### Configuration
- User configures after `sow init`
- Commented examples for artifact locations

### Summary Format
- Structured template
- Stored in `.sow/knowledge/explorations/`
