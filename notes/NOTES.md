# Orchestrator Overhaul - Work in Progress

**Date Started**: 2025-10-14
**Status**: Planning Phase

---

## Pain Points Identified

From recent experience using the current system:

1. **Orchestrator + Architect Combo Not Reliable**
   - Feedback loop results in 100% overengineering
   - Even with clear instructions to avoid it
   - Resulting architecture documents exceed project boundaries

2. **Design Phase Outputs Too Verbose**
   - Difficult for humans to review
   - Example: Simple Go CLI project → architect produced two 1000+ line documents

3. **Orchestrator Not Reliable at Scoping**
   - Easily convinced by architect to overengineer
   - Greatly exceeds constraints of current project
   - AI agents are bad at self-imposing constraints

---

## Proposed Solution: Human-in-Loop for Planning

### Five Fixed Phases (All Projects Use These)

1. **Discovery** (optional steps)
2. **Design** (optional steps)
3. **Implementation** (required)
4. **Review** (required)
5. **Finalize** (required)

### Key Design Principles

**Human as Primary Orchestrator in Planning:**
- Discovery and Design phases require human-in-the-loop
- Human acts as primary orchestrator, AI is subservient
- AI provides helpful structure and feedback
- AI orchestrates subagents on behalf of human

**AI Autonomy in Execution:**
- Implementation, Review, Finalize are AI-autonomous
- Human approval required at key gates
- AI has liberty to fail-forward and add tasks as needed

**Separation of Planning vs Execution:**
- Humans excel at: scoping, constraint-setting, architectural judgment
- AI excels at: execution within well-defined boundaries
- Each stays in their strength zone

---

## Phase Details

### Discovery Phase
- **Purpose**: Build context for design or bug investigation
- **When to Skip**: Human already has existing documents/context
- **Outputs**: Investigation notes, findings documents
- **Approval**: Human approves all artifacts before proceeding
- **Examples**:
  - Understanding existing codebase patterns
  - Investigating bug root causes
  - Exploring "how would we solve X?"

### Design Phase
- **Purpose**: Create design artifacts within human-defined constraints
- **Prerequisites**: Some existing context (from discovery or human notes)
- **Process**:
  1. Orchestrator works with human to build functional requirements
  2. Human sets boundaries and clear directions
  3. Orchestrator spawns architect agent (or does it itself for simple cases)
  4. Architect "fills in the blanks" within confines given
- **Outputs**: ADRs, design docs, architecture diagrams (human-governed)
- **Approval**: Human must approve all documents before leaving phase
- **Option**: Orchestrator can play architect role for simple scenarios (human decision)

### Implementation Phase
- **Purpose**: Execute the implementation work
- **Process**:
  1. Orchestrator creates initial task breakdown
  2. Human approves initial tasks
  3. Orchestrator has full autonomy to execute
  4. Orchestrator can request to add additional tasks (human approval required)
- **Approval Gates**:
  - Initial task list must be approved
  - Any request to add tasks after work started must be approved

### Review Phase
- **Purpose**: Validate that changes solve the original problem
- **Process**:
  1. Orchestrator spawns review subagent (or assists human in review)
  2. Subagent documents review findings
  3. Orchestrator presents findings to human
  4. If changes needed, loop back to implementation phase
- **Options**:
  - Full subagent review
  - Orchestrator-assisted human review
  - Human-only review
- **Approval**: Human must sign off before phase is complete

### Finalize Phase
- **Purpose**: Cleanup and prepare for merge
- **Process**:
  1. Update documentation, READMEs, changelogs
  2. Ensure git worktree is clean
  3. Create PR
  4. Present PR to human
- **Note**: Does NOT merge PR - human responsibility
- **Why Required**: "One project per branch" constraint means every project needs PR

---

## State Schema (Revised)

```yaml
project:
  name: project-name
  branch: feat/branch-name
  description: ...

phases:
  discovery:
    enabled: true/false
    status: pending/in_progress/completed/skipped
    artifacts:
      - path: investigation-notes.md
        approved: true/false

  design:
    enabled: true/false
    status: pending/in_progress/completed/skipped
    artifacts:
      - path: requirements.md
        approved: true/false
      - path: design-doc.md
        approved: true/false
      - path: adrs/001-decision.md
        approved: true/false

  implementation:
    enabled: true  # Always true
    status: pending/in_progress/completed
    tasks:
      - id: "010"
        name: Create base CLI structure
        status: pending/in_progress/completed
        # Task presence in list = human approval

  review:
    enabled: true  # Always true
    status: pending/in_progress/completed
    reviewer_type: human/orchestrator/agent
    artifacts:
      - path: review-findings.md

  finalize:
    enabled: true  # Always true
    status: pending/in_progress/completed
```

**Key Notes**:
- Discovery and Design can be skipped (enabled: false)
- Implementation, Review, Finalize are always enabled
- No `pending_approvals` field needed - log captures requests, state presence = approval
- Discovery can also produce artifacts (like design phase)

---

## Approval Mechanics

**Format**: Present options to human
```
Orchestrator: "I've completed discovery. Here's what I learned: [summary]
              Ready to proceed to design phase? [approve/feedback/revise]"
```

**Options**:
- `approve` - Continue to next phase
- `feedback` - Human provides corrections, continue current phase
- `revise` - Redo work in current phase

---

## Project Start Flow (Truth Table Approach)

Orchestrator guides human through questions to determine phase plan:

**Example Flow**:
```
Agent: What would you like to work on today?
Agent: 1. A new project
Agent: 2. An existing project
Agent: 3. A one-off task
Human: A new project

Agent: Great! What would you like to accomplish in this project?
Human: Create a CLI to help bring consistency to `sow`

Agent: Awesome, do you have any more context or existing documents?
Human: No

Agent: Based on your answer, I think we should start with a discovery phase
       so that we can learn more about the problem.
Human: Agreed, let's do that.

Agent: Great! Would you like me to create a branch and initialize the
       project structure now?
Human: Proceed
```

**Principle**: Orchestrator asks questions, guides human to right path based on answers, infers phase plan, human confirms.

---

## Implementation Plan

### Step 1: Define Each Phase ✅ COMPLETE
Create clear spec for each phase:
- What happens in detail
- What artifacts get created
- What approvals are needed
- When to enable/skip
- Success criteria

**Status**: Complete - See `notes/PHASE_SPECIFICATIONS.md`

**New Agents Identified**:
- Researcher (discovery phase) - Research via web/sinks/repos/local code
- Planner (implementation phase) - Task breakdown for large projects

All five phases fully specified:
- ✅ Discovery (optional, subservient, with researcher agent)
- ✅ Design (optional, subservient, with architect agent, design alignment subphase)
- ✅ Implementation (required, autonomous, with planner agent, fail-forward logic)
- ✅ Review (required, mandatory orchestrator review, iterative loop-back)
- ✅ Finalize (required, autonomous, mandatory project deletion before PR)

### Step 2: Build Truth Table ✅ COMPLETE
Map out the decision tree:
- Questions to ask at project start
- Branches based on answers
- Phase enablement at each endpoint
- Example paths through the tree

**Status**: Complete - See `notes/TRUTH_TABLE.md`

**Key Deliverables**:
- Decision flow structure with 4 main questions
- Three scoring rubrics (Discovery, Design, One-Off)
- Discovery type categorization system
- 17 resolved design decisions
- 5 detailed example walkthroughs
- Truth table matrix for common scenarios

### Step 3: Update State Schema ✅ COMPLETE
Formalize the YAML structure:
- Update CUE schemas
- Add validation rules
- Define all new fields
- Handle migration from old structure

**Status**: Complete and Finalized - See `notes/SCHEMA_PROPOSAL.md`

**Key Deliverables**:
- Complete revised CUE schema for project state (5 fixed phases)
- Field-by-field documentation with validation rules
- Complete YAML example showing full lifecycle with review loop-back
- Task state schema confirmation (no changes needed)
- File structure implications documented

**Design Decisions Finalized**:
- ✅ `pending_task_additions` kept in project state
- ✅ Artifact approval is boolean only (no timestamp)
- ✅ Review iteration counter is explicit field (`review.iteration`)
- ✅ No `reviewer_type` field (review approach can be multi-faceted)
- ✅ No `assigned_agent` in tasks (redundant - all tasks use implementer)

### Step 4: Rewrite Orchestrator
Complete rewrite of orchestrator prompt:
- Truth table logic for project start
- Phase-specific behaviors (subservient vs autonomous)
- Approval request mechanics
- Phase transition logic
- Error handling and recovery

### Step 5: Update Commands
Update slash commands:
- Modify `/start-project` or create new command
- Update `/continue` to understand new phases
- Ensure all commands work with new state schema
- Update any phase-specific commands

### Step 6: Update Documentation
Update all documentation:
- OVERVIEW.md - New phase model
- PROJECT_MANAGEMENT.md - New lifecycle
- AGENTS.md - Orchestrator role changes
- USER_GUIDE.md - New user experience
- COMMANDS_AND_SKILLS.md - Updated commands
- All other references to phases/orchestrator

---

## Questions to Resolve

None currently - Steps 1, 2, and 3 complete. Ready to proceed with Step 4 (Rewrite Orchestrator).

---

## References

- Current orchestrator: `.claude/agents/orchestrator.md`
- Current bootstrap: `.claude/commands/bootstrap.md`
- Documentation: `docs/`
