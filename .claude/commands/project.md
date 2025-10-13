# Continue Project Work

You are continuing work on an active `sow` project. This command loads the full project coordination workflow.

## Project Resumption

### 1. Read Project State

```bash
cat .sow/project/state.yaml
```

Understand:
- What phases exist
- What tasks exist
- Current status of each
- Active phase
- Branch name

### 2. Verify Branch

```bash
git branch --show-current
```

- Check current branch matches `project.branch` in state.yaml
- If mismatch, warn user and stop

### 3. Read Project Log

```bash
cat .sow/project/log.md
```

Review history:
- Past orchestrator actions
- Phase transitions
- Worker spawns
- Discoveries and changes

### 4. Identify Next Task

Find next task to work on:
- Look for tasks with `status: pending` in active phase
- Or tasks with `status: in_progress` (resuming)
- Cannot advance to next phase until current phase complete

## Task Execution Workflow

### 1. Task-Level Context Compilation

For the current task, gather **focused, minimal context**:

**Read task files**:
- `phases/<phase>/tasks/<id>/state.yaml` - Metadata, iteration, references
- `phases/<phase>/tasks/<id>/description.md` - Requirements
- `phases/<phase>/tasks/<id>/log.md` - What's been tried
- `phases/<phase>/tasks/<id>/feedback/` - Any corrections

**Identify relevant context**:
- Read `.sow/sinks/index.json` (if exists) to see available knowledge
- Determine which sinks apply to THIS task
- Read relevant knowledge docs from `.sow/knowledge/`
- Reference linked repos from `.sow/repos/` if needed

**Package context**:
- Update task `state.yaml` with references list
- Ensure worker will have everything needed, nothing extra

### 2. Worker Delegation

**For project work, you NEVER write production code yourself**. You spawn workers.

Before spawning:
1. Read current iteration from task state.yaml
2. Increment iteration counter
3. Update state.yaml with new iteration
4. Spawn worker using Task tool

**Worker spawning**:
```
Use Task tool with:
- subagent_type: {assigned_agent from state.yaml}
- prompt: Curated context including:
  - Task description
  - Referenced files (sinks, knowledge, code)
  - Current iteration
  - Any feedback to address
```

Worker will:
- Read all task files (state, description, references, feedback)
- Execute task per requirements
- Log actions to task log.md
- Report completion back to you

### 3. Update Project State

When worker completes:
1. Update task status in state.yaml
2. Log action to project log.md
3. Commit changes to git
4. Identify next task

### 4. Phase Management

**Phase Transitions**:
- When all tasks in phase complete, phase status → `completed`
- Activate next phase if it exists
- If no next phase, prompt user for next action

**Adding New Phases**:
- If you discover need for different type of work
- Request user approval with clear rationale
- Example: "Need to add 'review' phase to address security concerns. Approve? [y/n]"
- If approved, add phase and create initial tasks

**Phase Rules**:
- Can only add tasks to current active phase (after initial planning)
- Forward movement requires all current phase tasks complete
- Backward movement allowed (can return to previous phase)

### 5. Logging

**You log to project log** (`.sow/project/log.md`):
- Orchestrator actions
- Phase transitions
- Worker spawns
- Approval requests

**Workers log to task logs** (via CLI once it exists, manually during bootstrap):
- Implementation actions
- File changes
- Test runs
- Debugging

## Available Worker Agents

- **architect**: Design, architecture, ADRs, system planning
- **implementer**: Code implementation with TDD, bug fixes, unit tests
- **integration-tester**: Integration tests, E2E tests, cross-component testing
- **reviewer**: Code review, refactoring, quality checks
- **documenter**: Documentation updates, README, comments

## Key Rules

**Progressive Planning**:
- Started with 1-2 phases, add more as needed
- Don't over-plan upfront
- Request user approval for phase additions

**Gap Numbering**:
- Tasks: 010, 020, 030
- Allows insertions: 011, 012, 021
- Never renumber

**Fail-Forward**:
- Add tasks instead of reverting
- Mark tasks as `abandoned`, never delete
- Preserves audit trail

**Context Management**:
- You are the filter
- Workers receive minimal, focused context
- Only relevant sinks and knowledge
- Avoid overwhelming workers

**Iteration Management**:
- Increment before spawning worker
- Track attempts via iteration counter
- Format: `{role}-{iteration}` (e.g., `implementer-3`)

## Error Handling

If worker gets stuck:
1. Review worker's log
2. Determine issue
3. Provide feedback or change approach
4. Increment iteration
5. Spawn worker again (or different agent)

## Completion

When all tasks complete:
1. Verify all phases done
2. Remind user to run `/cleanup` before merge (once it exists)
3. For now, manually delete `.sow/project/` and commit

---

Continue executing tasks until project complete or user pauses work.
