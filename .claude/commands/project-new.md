# Create New Project

You are creating a new `sow` project. This command loads the full project creation workflow.

## Prerequisites Check

- ✓ User is on feature branch (not main/master)
- ✓ No existing `.sow/project/` directory

## Project Creation Workflow

### 1. Initial Project Discovery

Before planning, understand what exists:

1. **Read `.sow/index.md`** (if exists)
   - Provides map of logical components and documentation
   - If missing, note this but continue

2. **Gather project requirements from user**
   - What are they trying to build?
   - What is the scope?
   - Any specific requirements?

3. **Repository discovery scoped to project needs**
   - Based on user requirements, identify relevant areas
   - Use Glob to check for existing files in those areas
   - Read relevant documentation (as indicated by index)
   - Assess: What exists vs. what needs creating?
   - **Goal**: Avoid planning work that's already done

**Critical**: Discovery should be **project-scoped**, not blanket repository scan. If user wants to work on authentication, check auth-related files. If working on file structure, check template directories.

### 2. Assess Complexity

Rate project complexity (1-3):
- **1 - Simple**: Few files, focused scope, following existing patterns
- **2 - Moderate**: Multiple files, some integration, minor design needed
- **3 - Complex**: Many files, architectural changes, cross-cutting concerns

Base rating on **actual work needed**, not on how important it sounds.

### 3. Progressive Planning

**Start with 1-2 phases only**. Don't try to plan everything upfront.

Available phases (use as needed):
- `discovery` - Research and investigation
- `design` - Architecture and planning
- `implement` - Building features
- `test` - Integration and E2E testing
- `review` - Code quality improvements
- `deploy` - Deployment preparation
- `document` - Documentation updates

**Initial phase selection**:
- If designs already exist: Start with `implement`
- If design needed: Start with `design` → `implement`
- If unclear requirements: Start with `discovery` → `design`

### 4. Create Initial Tasks

For initial phase(s), break work into 3-5 tasks:

**Gap numbering**: 010, 020, 030, 040
- Allows insertions: 011, 012, 021, etc.
- Never renumber existing tasks

**Agent assignment**: Assign appropriate agent to each task based on work type:
- `architect` - Design, architecture, ADRs
- `implementer` - Code implementation with TDD
- `integration-tester` - Integration and E2E tests
- `reviewer` - Code review, refactoring
- `documenter` - Documentation updates

### 5. Create Project Structure

Manually create (since CLI doesn't exist yet):

```
.sow/project/
├── state.yaml          # Project state, phases, tasks
├── log.md              # Orchestrator action log
├── context/            # Project-specific context
└── phases/
    └── {phase}/
        └── tasks/
            └── {task-id}/
                ├── state.yaml      # Task metadata
                ├── description.md  # Requirements
                └── log.md          # Worker action log
```

**state.yaml** includes:
- Project name, description
- Branch name (for validation)
- Complexity assessment
- Phases and tasks
- Agent assignments
- Gap-numbered task IDs

### 6. Commit to Git

```bash
git add .sow/project/
git commit -m "chore: initialize <project-name> project"
```

### 7. Begin Execution

After project created, invoke `/project` command to start work.

## Important Notes

**Discovery Before Planning**:
- Always check `.sow/index.md` first
- Use Glob/Grep to verify what exists
- Read existing documentation before creating design tasks
- Complexity ratings should reflect actual work needed

**Progressive, Not Waterfall**:
- Start minimal (1-2 phases)
- Add phases as work progresses (with user approval)
- Fail-forward: add tasks, never delete (mark abandoned)

**Human Approval Gates**:
- Request approval when adding new phases
- Provide clear rationale for changes
- User validates direction before proceeding

---

After project creation complete, invoke `/project` to begin work.
