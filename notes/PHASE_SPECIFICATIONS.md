# Phase Specifications

**Status**: Draft - Work in Progress
**Date**: 2025-10-14

---

## Overview

This document specifies the behavior, artifacts, and mechanics of each of the five phases in the new sow project lifecycle.

**Core Principle**: The orchestrator gracefully leads the human through phases based on feedback. The orchestrator ASKS whether to skip phases, never DECIDES unilaterally.

---

## Discovery Phase

### Purpose

Build context for design or bug investigation. Illuminate the problem space through research, exploration, and conversation.

### New Agent: Researcher

**Purpose**: Performs research to ground discussions in real sources.

**Sources**:
- Available sinks (`.sow/sinks/`)
- Linked repositories (`.sow/repos/`)
- Local repository codebase
- Web search

**Output**: Summarized findings for orchestrator/human review

**Invocation**:
- Orchestrator suggests: "Would you like me to use the researcher agent to research X?"
- Human requests: "Please use the research agent to investigate Y"

### Orchestrator Role

The orchestrator is **subservient** in this phase, acting as an assistant to the human.

**Responsibilities**:
1. **Ask questions** to clarify and deepen understanding
2. **Point out inconsistencies** or contradictions with existing docs
3. **Help brainstorm** potential solutions
4. **Make suggestions** for areas human may not be considering
5. **Take notes** (primary responsibility) - synthesize conversation continuously
6. **Log the conversation** - chronological record of what was discussed

**Key Behavior**: The orchestrator is a supportive facilitator, not a decision-maker.

### Entry Criteria

**Context**: User has chosen to "start a new project" (not continue existing or one-off task).

**Entry Question**:
```
What would you like to work on in this project? Some ideas:

- Fix a bug
- Add a new feature
- Refactor messy code
```

**Scenarios Indicating Discovery Needed**:
- User says "I need to understand X better"
- Bug needs investigation
- Feature requires domain knowledge
- Technical decision needs research
- User provides minimal context: "I want to add an API endpoint"

**Scenarios Indicating Discovery NOT Needed**:
- User references existing design doc: "Implement the `fs` library described in docs/filesystem.md"
- User has detailed notes and says: "I think this is good enough to come up with an implementation plan"
- User explicitly requests to skip: "Let's skip discovery and go straight to design"

**Orchestrator Behavior**:
- ALWAYS ask, never decide
- "Do you want to do some more discovery around this or are you ready to work on a design document?"
- Gracefully guide human based on their responses

### Exit Criteria

**How to know you're done**:
- Human feels they have enough context
- Key questions have been answered
- Next steps are clear
- Artifacts approved by human

**Orchestrator Prompts**:
- "It seems we have a solid grasp on the problem. Do you want to create a design document or go straight to the implementation step?"

**Human Signals**:
- "Great, I think we've fully discovered this problem and can move on now."

**Clarification**: Orchestrator always clarifies whether design phase is needed or if going straight to implementation.

### Types of Discovery Work

Common discovery scenarios, each with its own slash command for context preservation:

| Type                           | Slash Command               | Description                                                                                                                                                                                 |
| ------------------------------ | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Bug Investigation**          | `/phase:discovery:bug`      | Root cause analysis, reproduction steps, affected systems                                                                                                                                   |
| **Feature Exploration**        | `/phase:discovery:feature`  | Problem space, user needs, existing solutions, competitive analysis                                                                                                                         |
| **Documentation Gap Analysis** | `/phase:discovery:docs`     | Identify gaps between code and documentation. Compare current implementation against existing docs. Find what's missing or outdated.                                                        |
| **Refactoring Analysis**       | `/phase:discovery:refactor` | Understand WHY code is messy. Analyze responsibilities, coupling, patterns. Identify what needs to change and risks involved. Example: "This file is 2000+ lines, what can we do about it?" |
| **General/Mixed**              | `/phase:discovery`          | Catch-all for work that doesn't fit other buckets                                                                                                                                           |

**Inference**: Orchestrator infers which type based on human's description, but remains flexible (doesn't force into buckets).

**Example**:
```
Human: "I want to add a new API endpoint"
Orchestrator: [Infers this is feature exploration]
Orchestrator: [Invokes /phase:discovery:feature internally]
```

### Artifacts & Structure

All artifacts stored in single folder within project directory:

```
.sow/project/phases/discovery/
├── log.md                           # Chronological conversation log
├── notes.md                         # Orchestrator's living notes (updated continuously)
├── decisions.md                     # Key conclusions/decisions reached
└── research/
    ├── 001-auth-patterns.md         # Researcher findings
    ├── 002-competitor-analysis.md
    └── 003-current-implementation.md
```

**Artifact Descriptions**:

- **`log.md`**: Full conversation record, similar to implementation log
  - Chronological entries with timestamps
  - Captures what was discussed
  - Helps with resumability

- **`notes.md`**: Orchestrator's synthesized notes
  - Living document, updated continuously
  - Structured synthesis of conversation
  - Key insights and findings
  - Human-readable summary

- **`decisions.md`**: Key conclusions made during discovery
  - Explicit decisions about direction
  - Constraints identified
  - Requirements clarified
  - Referenced in later phases

- **`research/*.md`**: Researcher agent outputs
  - Numbered sequentially (001, 002, 003...)
  - Immutable once created
  - Clear topic in filename
  - Structured findings with sources

### Slash Command Architecture

**Entry Point**: `/project:new`
- Decision pipeline for determining starting phase
- Asks initial questions
- Guides human through phase selection
- Once discovery is confirmed, invokes `/phase:discovery`

**Discovery Entry**: `/phase:discovery`
- Main discovery phase orchestration
- Infers discovery type from context
- Invokes type-specific command if applicable

**Type-Specific Commands**:
- `/phase:discovery:bug` - Bug investigation instructions
- `/phase:discovery:feature` - Feature exploration instructions
- `/phase:discovery:analysis` - Current state analysis instructions

**Rationale**: Context preservation - only load instructions relevant to current work, not all 1000+ lines at once.

### Success Criteria

Discovery phase is successful when:
- [ ] Human understands the problem space
- [ ] Key questions have been answered
- [ ] Sufficient context exists for next phase
- [ ] Artifacts created and approved by human
- [ ] Next steps are clear (design or implementation)
- [ ] Human explicitly approves moving to next phase

### Approval Mechanics

**Format**: Present options to human
```
Orchestrator: "I've completed research on X. Here's what I learned: [summary]
              Ready to proceed to design phase? [approve/feedback/revise]"
```

**Options**:
- `approve` - Continue to next phase
- `feedback` - Human provides corrections, continue discovery
- `revise` - Redo specific research or investigation

### Phase Transition

**To Design Phase**:
- Human approves discovery artifacts
- Human indicates design doc is needed
- Orchestrator transitions: invokes `/phase:design`

**To Implementation Phase** (skip design):
- Human approves discovery artifacts
- Human indicates sufficient detail exists
- Design doc not needed
- Orchestrator transitions: invokes `/phase:implementation`

---

## Design Phase

### Purpose

Create formal design artifacts within human-defined constraints. Synthesize discovery findings or human-provided notes into structured design decisions and documentation.

### Entry Criteria

**Entry Paths**:
1. From discovery phase → design needed to formalize findings
2. Directly from project start → user has notes but needs formal design

**Critical Thinking Required**: Orchestrator must think critically about whether design documents are actually needed. LLMs tend to over-index on creating design docs when they see the option.

**Scenarios Where Design IS Warranted**:
- Adding a completely new (non-trivial) component to the system
- Project scope is very large and needs breaking down into formal design docs
- Changing a fundamental piece of the architecture
- Technical decision from discovery warrants creating an ADR
- User explicitly requests design documentation

**Scenarios Where Design is Usually NOT Needed**:
- Bug fixes (just implement the fix)
- Small refactorings (notes from discovery are sufficient)
- Small features (can go straight to implementation)
- Minor changes to existing features (no major design decisions)

**Orchestrator Behavior**:
- Actively assess whether design phase is necessary
- Default to skipping design unless clear justification exists
- Prompt: "Based on the discovery findings, I don't think we need formal design documents. Shall we proceed to implementation? [yes/create design docs]"

### Orchestrator Role

The orchestrator is **subservient** in this phase, acting as an assistant to the human (same as discovery).

**Primary Responsibilities**:
1. **Synthesize data** - Make sense of discovery notes + existing architecture
2. **Assist with fit** - Help user understand how design fits into current system
3. **Design alignment** - Work with human to refine discovery into architecture decisions
4. **Facilitate documentation** - Either build docs directly or coordinate architect agent
5. **Take notes** - Keep running notes of design decisions

**Key Strength**: LLMs excel at synthesizing large amounts of data (discovery notes + existing codebase architecture) into coherent patterns. Use this strength to help human see the big picture.

### Design Alignment Subphase

**Philosophical Concept** (not represented in state file):

This is the critical work of refining raw discovery data into high-level architecture decisions before creating formal documents.

**Analogy**: "Mix ingredients to make cake batter, but don't bake the cake yet"
- Discovery = raw ingredients
- Design alignment = mixing into batter
- Design documents = baking the cake

**Activities**:
- Review discovery artifacts
- Identify key design decisions
- Map to existing architecture
- Clarify constraints and requirements
- Define scope of what will be documented
- Reach consensus on approach

**Output**: Clear, structured notes ready to be formalized into design documents

**This work happens conversationally** between orchestrator and human, captured in notes.

### Architect Agent

**When to Invoke**:
- User requests it explicitly
- Design docs are complex or numerous
- Orchestrator is unsure (prompts user: "Would you like me to create the design doc or spawn an architect agent?")

**When Orchestrator Handles Directly**:
- Simple, short design docs
- User prefers direct interaction
- Single ADR or small design doc

**Architect Agent Role**:
Acts as a **translation layer** - takes messy design alignment notes and produces formal documentation.

**Responsibilities**:
- Transform design alignment notes into proper design documents
- Freedom to organize output into one or more documents
- Structure content appropriately (ADR format, design doc format, etc.)
- Fill in details within the constraints established
- Produce "camera-ready" documentation

**Feedback Loop**:
1. Architect produces design documents
2. User reviews and provides feedback
3. If small changes needed → Orchestrator applies directly
4. If extensive changes needed → Orchestrator spawns new architect agent with feedback

### Slash Command Hierarchy

**Entry Point**: `/phase:design`
- Main design phase orchestration
- Design alignment work
- Decision to use orchestrator or architect

**Architect Commands** (invoked by orchestrator on behalf of user):
- `/architect:adr` - Create Architecture Decision Record
- `/architect:prd` - Create Product Requirements Document (if applicable)
- `/architect:design-doc` - Create general design document
- `/architect:api-spec` - Create API specification
- (Additional architect commands as needed)

**Rationale**: Architect commands are specialized and context-heavy. Only load when actually needed.

### Artifacts & Structure

**Location**: All design artifacts stay in project directory during development

```
.sow/project/phases/design/
├── log.md                          # Chronological conversation log
├── notes.md                        # Design alignment notes
├── requirements.md                 # Functional requirements (if formalized)
├── adrs/
│   ├── 001-use-postgresql.md
│   └── 002-event-driven-arch.md
├── design-docs/
│   ├── api-specification.md
│   └── data-model.md
└── diagrams/
    └── architecture-overview.png
```

**Artifact Descriptions**:

- **`log.md`**: Chronological conversation record (like discovery)
- **`notes.md`**: Design alignment notes - the "batter" before baking
- **`requirements.md`**: Formalized functional requirements (optional)
- **`adrs/`**: Architecture Decision Records (numbered)
- **`design-docs/`**: General design documents
- **`diagrams/`**: Architecture diagrams, flowcharts, etc.

**Question for Finalize Phase**: Some artifacts may need to be merged into repository:
- ADRs should be moved to repo's ADR folder
- Architecture docs may go into repo docs/
- Implementation-specific notes stay in project (not merged)

This decision happens in finalize phase: "Are there any artifacts from the design phase that should be merged into the repository?"

### Success Criteria

Design phase is successful when:
- [ ] One or more design documents have been produced
- [ ] Design documents are complete and properly formatted
- [ ] Design fits within project constraints
- [ ] Design aligns with existing architecture
- [ ] All design decisions are documented
- [ ] Human has reviewed and approved all documents

### Approval Mechanics

**Format**: Present documents to human for review
```
Orchestrator: "I've created the following design documents:
              - ADR 001: Use PostgreSQL for primary database
              - Design doc: API specification

              Please review. [approve/feedback/revise]"
```

**Options**:
- `approve` - All documents approved, proceed to implementation
- `feedback` - Provide corrections (orchestrator or architect applies)
- `revise` - Significant changes needed, restart design process

**Requirement**: Human must explicitly approve ALL design documents before phase is complete.

### Phase Transition

**To Implementation Phase**:
- All design documents produced
- Human has approved all documents
- Design alignment complete
- Orchestrator transitions: invokes `/phase:implementation`

**Note**: Design phase always leads to implementation (required phase)

---

## Implementation Phase

### Purpose

Execute the implementation work to achieve the project's goals. Break down requirements into tasks and complete them using implementer agents.

### Entry Criteria

**Required Phase**: Implementation always happens (cannot be skipped).

**Entry Paths**:
1. Directly from project start → User has clear requirements, no discovery/design needed
2. From discovery phase → Skip design, implementation plan clear from discovery
3. From design phase → Design documents provide implementation roadmap

**Orchestrator Behavior**: Fairly obvious since implementation is required. Focus is on creating good task breakdown.

### Initial Task Breakdown Process

The orchestrator is **partially autonomous** here - it creates the plan but must get human approval.

**Orchestrator's Job**: Break down inputs from previous phases into an implementation plan.

**Implementation Plan**: A list of one or more tasks that, if completed successfully, would meet the original project intention.
- Bug fix → Steps necessary to fix the bug
- Feature → Steps to fully implement the feature
- Refactor → Steps to complete the refactoring

**Human Approval Required**: User must approve the initial task breakdown before work begins.

### Planner Agent (Optional)

**Purpose**: For large projects, assist with task breakdown to avoid overwhelming orchestrator's context.

**When to Use**:
- Very large projects (10+ tasks expected)
- Complex task dependencies
- Orchestrator is uncertain about breakdown
- User requests it explicitly

**When NOT to Use**:
- Small projects (1-5 tasks)
- Clear, straightforward breakdown
- Orchestrator is confident

**Planner Agent Role**:
- Analyzes discovery/design artifacts
- Suggests logical task breakdown
- Considers dependencies and ordering
- Proposes parallel vs sequential tasks
- Outputs separate planning document

**Key Distinction**: Planner does NOT write state file directly (orchestrator controls state).

**Output**: Planner produces `implementation-plan.md` with suggested task breakdown:
```markdown
# Implementation Plan

## Tasks

### 010: Create database schema
- Status: pending
- Dependencies: none
- Parallel: false
- Description: Create PostgreSQL schema for user tables

### 020: Implement User model
- Status: pending
- Dependencies: 010
- Parallel: false
- Description: Create User model with validation

### 030: Create REST API endpoints
- Status: pending
- Dependencies: 020
- Parallel: false
- Description: Implement CRUD endpoints for users

### 040: Write integration tests
- Status: pending
- Dependencies: 030
- Parallel: true (can run with other test tasks)
- Description: End-to-end tests for user API
```

**Orchestrator's Use**: Reviews planner's document, adjusts if needed, creates actual state file, presents to human for approval.

### Planning Procedure (Consistent for Orchestrator or Planner)

Whether the orchestrator or planner agent does the work, follow this consistent procedure:

**Step 1: Analyze Inputs**
- Review discovery artifacts (if any)
- Review design documents (if any)
- Review human's original requirements
- Understand constraints and goals

**Step 2: Identify Work Units**
- What needs to be built/changed?
- What are the logical chunks?
- What are the dependencies?

**Step 3: Create Task Breakdown**
- Use gap numbering (010, 020, 030...)
- Clear, actionable task descriptions
- Identify dependencies
- Mark parallelizable tasks
- Assign to implementer agent

**Step 4: Validate Plan**
- Does this accomplish the goal?
- Are tasks appropriately sized?
- Are dependencies correct?
- Is anything missing?

**Step 5: Present to Human** (orchestrator always does this)
- Show task breakdown
- Explain reasoning
- Request approval

### Task Execution Mechanics

This is where the orchestrator has **maximum autonomy**.

**Normal Flow** (No Human Approval Needed):
1. Orchestrator identifies next pending task (or parallel set of tasks)
2. Orchestrator spawns implementer agent(s) with task context
3. Implementer completes work (follows existing implementer behavior - no changes)
4. Implementer reports back
5. Orchestrator marks task as completed
6. Orchestrator moves to next task
7. Repeat until all tasks done

**Implementer Agent**: Behavior already documented, no changes suggested. Follows existing TDD approach, logging, etc.

**Parallel Execution**: Orchestrator can spawn multiple implementers simultaneously for tasks marked as parallelizable.

**Implementer Failure/Issues**: Orchestrator has autonomy to:
- Re-invoke implementer with additional feedback
- Adjust task description
- Mark task as completed if "good enough"
- Continue iterating with implementer

**No Human Approval Needed For**:
- Marking tasks completed
- Moving to next task
- Re-invoking implementer with feedback
- Normal task execution flow

### When Human Approval IS Required

The orchestrator does NOT have full autonomy in these specific scenarios:

**1. Implementer Stuck - Human Help Needed**
- Problem occurred that orchestrator cannot resolve
- Needs human input to unblock
- Orchestrator prompts: "The implementer encountered issue X. I need your help to resolve this. [explain situation]"

**2. Fail-Forward - Add New Tasks**
- Problem occurred revealing new work needed
- Orchestrator wants to add tasks to address it
- Orchestrator prompts: "I've discovered we need to add task: X. This is needed because [reason]. Approve? [yes/no]"
- If approved, task added to state and work continues

**3. Go Back to Previous Phase**
- Problem reveals design is flawed OR more discovery needed
- Orchestrator wants to return to discovery or design phase
- Orchestrator prompts: "I've identified a fundamental issue with the design. We need to revisit [discovery/design]. Here's why: [explanation]. Shall we go back? [yes/no]"

**Key Principle**: Human approval only for **real problems**, not for normal task flow.

### Fail-Forward Logic

**Concept**: When work reveals new requirements, add tasks and keep moving forward.

**Process**:
1. Problem/gap discovered during implementation
2. Orchestrator identifies what additional work is needed
3. Orchestrator requests approval to add new task(s)
4. If approved: Add tasks to state, continue
5. Original task may be marked as "abandoned" if no longer relevant

**Task Abandonment**:
- When original task is no longer applicable
- Mark as "abandoned" in state
- Does NOT count toward "all tasks complete"
- Log explains why it was abandoned

**Example**:
```
Task 020: Implement JWT service
- During work, discover we need OAuth, not JWT
- Orchestrator: "We need to switch to OAuth. I'll abandon task 020 and add 021: Implement OAuth service. Approve?"
- User: yes
- Task 020 → abandoned
- Task 021 → added and pending
```

### Artifacts & Structure

```
.sow/project/phases/implementation/
├── log.md                          # Orchestrator's coordination log
├── implementation-plan.md          # Planner output (if used)
└── tasks/
    ├── 010/
    │   ├── state.yaml             # Task metadata
    │   ├── description.md         # Task requirements
    │   ├── log.md                 # Implementer action log
    │   └── feedback/              # Human corrections (if any)
    ├── 020/
    │   └── ...
    └── 030/
        └── ...
```

**Artifact Descriptions**:

- **`log.md`**: Orchestrator's coordination log (task spawning, completions, decisions)
- **`implementation-plan.md`**: Planner agent's task breakdown (if planner was used)
- **`tasks/*/state.yaml`**: Task metadata (iteration, status, dependencies)
- **`tasks/*/description.md`**: Task requirements and context
- **`tasks/*/log.md`**: Implementer agent's action log (already documented behavior)
- **`tasks/*/feedback/`**: Human feedback files (numbered chronologically)

### Success Criteria

Implementation phase is successful when:
- [ ] All non-abandoned tasks are completed
- [ ] Code is committed to git
- [ ] Tests are passing
- [ ] No blocking issues remain

**Note**: "All tasks complete" does NOT include abandoned tasks. Abandoned tasks are documented but don't block completion.

### Phase Transition

**To Review Phase**:
- All tasks completed successfully
- **Automatic transition** - no human approval required
- Orchestrator invokes `/phase:review`

**Note**: Human can still request going back to implementation after their own review in the review phase.

---

## Review Phase

### Purpose

Validate that changes made in the implementation phase meet the expected outcome. This is the critical quality gate before finalization.

### Entry Criteria

**Required Phase**: Review always happens (cannot be skipped).

**Entry Path**: Always from implementation phase (automatic transition).

### Review Principle

**Universal Question**: "Do the changes made in the implementation phase meet our expected outcome?"

**No Different "Modes"**: Whether reviewing a bug fix, new feature, or refactor, the fundamental question remains the same.

**Review Process Requires**:

1. **Review original requirements**
   - Documents provided by user initially
   - Discovery artifacts (if discovery phase was used)
   - Design documents (if design phase was used)
   - Original project intent

2. **Review ALL implementation changes**
   - Every file modified
   - Every commit made
   - All task logs
   - Test coverage and results

3. **Compare and validate**
   - Does implementation match requirements?
   - Was the original intent achieved?
   - Are there gaps or deviations?
   - Is quality acceptable?

**Critical Importance**: Easy to get side-tracked during implementation and miss the mark. Review phase is the one chance to validate success before finalizing.

### Orchestrator Review (Mandatory)

**The orchestrator ALWAYS performs a review** - this is mandatory and happens automatically.

**Orchestrator's Review Process**:
1. Read all original requirements/artifacts
2. Read all implementation changes
3. Compare implementation against requirements
4. Identify gaps, issues, or concerns
5. Document findings in review report
6. Present findings to human

**Why Mandatory**: No harm in doing it, likely net positive. Ensures quality gate is always present.

**Orchestrator Can**:
- Perform complete review itself (simple cases)
- Spawn reviewer agent for assistance (complex cases)
- Either way, orchestrator presents final report

### Reviewer Agent (Optional)

**When to Invoke**:
- Large number of changes (10+ files modified)
- Complex changes requiring deep analysis
- Orchestrator is uncertain about quality
- User requests it explicitly

**When NOT to Invoke**:
- Small changes (1-5 files)
- Simple, straightforward work
- Orchestrator is confident in its assessment

**Reviewer Agent Behavior**:

Follows the same review principle:

1. **Review Requirements**
   - Read discovery artifacts
   - Read design documents
   - Understand original intent
   - Identify acceptance criteria

2. **Review Implementation**
   - Examine every file change
   - Review task logs
   - Check test coverage
   - Analyze code quality

3. **Compare and Validate**
   - Does implementation match requirements?
   - Are there gaps or issues?
   - Is quality acceptable?
   - What recommendations exist?

4. **Document Findings**
   - Write review report
   - Identify specific issues
   - Provide recommendations
   - Assess overall success

**Output**: Review report document (orchestrator presents to human)

### Human Review (Optional)

**After Orchestrator Review**: Human can choose to do their own review or trust orchestrator's assessment.

**Orchestrator Assistance**:
- Human asks questions: "Why was this interface designed this way?"
- Orchestrator explains decisions and rationale
- Human provides direct feedback
- Orchestrator helps navigate changes

**Human Can**:
- Trust orchestrator's review and approve
- Perform own review with orchestrator assistance
- Override orchestrator's assessment (approve even if issues found, or reject despite clean review)

### Artifacts & Structure

```
.sow/project/phases/review/
├── log.md                     # Review phase conversation log
└── reports/
    ├── 001-review.md          # First review iteration
    ├── 002-review.md          # Second review iteration (if looped back)
    └── 003-review.md          # Third review iteration (if looped back)
```

**Review Report Format** (001-review.md, 002-review.md, etc.):

```markdown
# Review Report 001

**Date**: 2025-10-14
**Reviewer**: orchestrator (or reviewer agent)
**Implementation Phase**: Completed 2025-10-14

## Original Intent

[Summary of what we were trying to achieve]

## Changes Made

[Summary of implementation changes]
- Files modified: 8
- Tests added: 12
- Commits: 5

## Review Findings

### Requirements Met ✓
- [Requirement 1]: Met
- [Requirement 2]: Met
- [Requirement 3]: Met

### Issues Identified ✗
- [Issue 1]: Description and severity
- [Issue 2]: Description and severity

### Recommendations
- [Recommendation 1]
- [Recommendation 2]

## Overall Assessment

[Pass/Fail with reasoning]

## Next Steps

[Approve and finalize, or loop back to implementation with specific tasks]
```

**Numerical Ordering**: Review reports are numbered (001, 002, 003...) because multiple iterations are possible.

**Use in Implementation**: If looping back, these reports serve as additional context for implementer agents.

### Success Criteria

Review phase is successful when:
- [ ] Review report has been created
- [ ] Report sufficiently proves original intent was met
- [ ] All critical issues addressed (or documented as acceptable)
- [ ] Human has reviewed and approved

**Human Decides**: Ultimately, the human gates whether success has been achieved.

### Failure/Loop-Back Mechanics

**When Issues Are Identified** (by orchestrator, reviewer agent, or human):

**Orchestrator Process**:
1. **Identify resolution** - What needs to change to address feedback?
2. **Adjust implementation phase** - Add tasks to address issues
3. **Present changes to human** - "I've identified these issues. I propose adding tasks X, Y, Z to address them. Approve?"
4. **Upon approval** - Return to implementation phase
5. **Execute new tasks** - Implementer agents complete the work
6. **Return to review** - Automatic transition back to review phase
7. **Repeat review process** - Create new review report (002-review.md)

**Can Iterate Multiple Times**: Review → Implementation → Review → Implementation → ... until success.

**Example Flow**:
```
Implementation Phase Complete
↓
Review Phase (001-review.md)
- Issues found: Missing error handling, insufficient tests
↓ (loop back)
Implementation Phase (add tasks 035, 036)
- Task 035: Add error handling
- Task 036: Add missing tests
↓ (automatic transition)
Review Phase (002-review.md)
- No issues found, all requirements met
↓ (human approval)
Finalize Phase
```

### Human Approval Process

**Orchestrator Presentation**:

After review is complete (orchestrator or reviewer agent finished):

```
Orchestrator: "I've completed the review. Here's my assessment:

              [Summary of review findings]

              Overall: The implementation meets the original intent.

              Review report: .sow/project/phases/review/reports/001-review.md

              Ready to proceed to finalization? [approve/feedback/loop-back]"
```

**Options**:
- `approve` - Proceed to finalize phase
- `feedback` - Human provides additional feedback, orchestrator incorporates
- `loop-back` - Return to implementation to address issues

**Human Override**:
- Human can approve even if orchestrator found issues (accept trade-offs)
- Human can reject even if orchestrator found no issues (wants more changes)

**Requirement**: Human must explicitly approve before proceeding to finalize.

### Phase Transition

**To Finalize Phase**:
- Review report created and approved
- Human has explicitly approved
- Orchestrator invokes `/phase:finalize`

**Loop Back to Implementation**:
- Issues identified requiring changes
- Tasks added to implementation phase
- Human approves loop-back
- Orchestrator invokes `/phase:implementation`
- After implementation complete, automatic transition back to review

---

## Finalize Phase

### Purpose

Cleanup, update documentation, and prepare for merge. Create PR with all project files removed.

### Entry Criteria

**Required Phase**: Finalize always happens (cannot be skipped).

**Entry Path**: Always from review phase (human-approved).

### Documentation Subphase

Initial subphase where orchestrator determines what documentation needs updating.

**Orchestrator Process**:

1. **Identify key changes**
   - Review what was implemented
   - Summarize major changes
   - Should be mostly evident by this point

2. **Review existing documentation**
   - READMEs (root and subdirectories)
   - Developer guides
   - Changelogs (CHANGELOG.md)
   - Code documentation (package-level docs, module docs)
   - API documentation
   - Architecture documentation

3. **Determine updates needed**
   - What documentation is now outdated?
   - What new documentation is needed?
   - Are there gaps to fill?

4. **Answer design artifact question**
   - "Should any of the design documents be merged into the repository?"
   - ADRs typically go to repository's ADR folder
   - Architecture docs may go to docs/
   - Implementation-specific notes stay in project (deleted with project)

5. **Propose changes to human**
   - List all proposed documentation updates
   - Explain rationale
   - Request approval

6. **Make approved changes**
   - Update/create documentation
   - Move design artifacts if approved
   - Commit changes

**When to Skip Subphase**:
- Change was sufficiently small (e.g., bug fix with no user-facing impact)
- No existing documentation in repository needs updating
- Orchestrator determines no documentation changes warranted

**Human Approval**: Required before making documentation changes.

### What Gets Updated/Created

**Typical Documentation Updates**:

- **README.md**: New features, changed behavior, updated installation steps
- **CHANGELOG.md**: Entry for this change (version, date, description)
- **Developer Guides**: New patterns, changed conventions
- **Code Documentation**: Package docs, module docs reflecting new code
- **API Documentation**: New endpoints, changed schemas
- **Architecture Docs**: Moved from design phase (if applicable)
- **ADRs**: Moved from design phase to repository's ADR folder

**Design Artifacts**:
- Orchestrator asks: "Should we move ADR 001 to docs/adrs/?"
- Human approves: Orchestrator moves file, commits

### Final Checks

After documentation subphase complete:

**Orchestrator Verifies**:

1. **Tests passing**
   - Run full test suite
   - All tests must pass
   - If failures, loop back to implementation

2. **Linters passing**
   - Run configured linters
   - All checks must pass
   - If failures, fix or loop back to implementation

3. **Documentation complete**
   - All approved documentation updates made
   - Commits pushed

4. **Git tree clean**
   - No uncommitted changes
   - All work committed and pushed
   - Ready for PR

### Critical Transition: Project Deletion

**Mandatory Step**: Before creating PR, the project folder MUST be deleted.

**Central Rule**: Projects are per branch and never present on non-feature branches (main/master).

**Why Critical**:
- PR should never include `.sow/project/` files
- CI should reject PRs with project files present
- Keeps main branch clean
- Project state is branch-specific, not repository-wide

**Deletion Process**:

1. **Verify everything committed**
   - All implementation work committed
   - All documentation updates committed
   - Git tree is clean

2. **Delete project folder**
   ```bash
   rm -rf .sow/project/
   ```

3. **Create cleanup commit**
   ```bash
   git add .sow/
   git commit -m "chore: remove sow project files before merge"
   ```

4. **Push cleanup commit**
   ```bash
   git push origin <branch-name>
   ```

**After Deletion**: Project state is gone from branch. Cannot resume project after this point. This is intentional - work is complete.

### PR Creation

**Default Method**: Use `gh` CLI (GitHub CLI)

**Prerequisites**:
- `gh` CLI installed and authenticated
- Repository is on GitHub
- Branch pushed to remote

**Orchestrator Process**:

1. **Check for `gh` CLI**
   ```bash
   which gh
   ```

2. **If `gh` available**: Create PR automatically

   **PR Title**: Descriptive summary of changes
   - Example: "Add user authentication with JWT"

   **PR Description**: Comprehensive details following best practices

   **Template**:
   ```markdown
   ## Summary

   [1-3 sentence summary of changes]

   ## Changes Made

   - [Key change 1]
   - [Key change 2]
   - [Key change 3]

   ## Implementation Details

   [Brief technical details from implementation/review phases]

   ## Testing

   - [Test approach]
   - [Coverage details]

   ## Documentation

   - [What docs were updated]

   ## Related

   - Closes #123 (if applicable)
   - Related to #456 (if applicable)

   ---

   Generated by sow project: [project name]
   Review reports: .sow/project/phases/review/reports/ (deleted before merge)
   ```

3. **If `gh` NOT available**: Instruct human

   ```
   Orchestrator: "The gh CLI is not available. Please create a PR manually:

                 1. Go to GitHub repository
                 2. Click 'New Pull Request'
                 3. Select branch: <branch-name>
                 4. Use this title: [suggested title]
                 5. Use this description:

                 [Generated PR description]

                 I'll wait while you create the PR."
   ```

**Best Practices for PR**:
- Clear, descriptive title
- Comprehensive description
- Link related issues
- Include testing information
- Note documentation changes
- Use conventional commit message format for title if applicable

### Success Criteria

Finalize phase is successful when:
- [ ] Documentation updates proposed and approved (if needed)
- [ ] All documentation changes made and committed
- [ ] Tests passing
- [ ] Linters passing
- [ ] Git tree clean (all work committed)
- [ ] Project folder deleted (cleanup commit made)
- [ ] PR created (via `gh` or manually by human)

### Human Handoff

**After PR Created**:

```
Orchestrator: "✓ Finalize phase complete!

              Documentation updated:
              - README.md (added authentication section)
              - CHANGELOG.md (v1.2.0 entry)
              - docs/adrs/003-use-jwt.md (moved from project)

              PR created: https://github.com/org/repo/pull/123

              The project is now ready for review and merge.

              Next steps:
              1. Review the PR (or request team review)
              2. Address any PR feedback
              3. Merge when approved

              Note: Project files have been removed from the branch.
              This project cannot be resumed after finalization."
```

**Human Responsibility**:
- Review PR
- Request reviews from team
- Address PR feedback
- Merge PR when approved
- Delete feature branch after merge

**Orchestrator Does NOT**:
- Merge the PR (always human responsibility)
- Delete the branch (human decides when)
- Request reviewers (human knows who to ask)

### Artifacts & Structure

**After Finalization**: `.sow/project/` is deleted, so no project artifacts remain.

**What Persists in Repository**:
- Code changes (in git history)
- Documentation updates (committed)
- Moved design artifacts (ADRs, design docs)

**What's Deleted**:
- All `.sow/project/` contents
- Discovery artifacts
- Design phase notes (unless moved to repository)
- Implementation task logs
- Review reports

**What's In PR**:
- Code changes
- Documentation updates
- Moved design artifacts (if any)
- NO `.sow/project/` files (deleted in cleanup commit)

### Phase Transition

**Final Phase**: Finalize is the last phase - no automatic transition.

**Project Complete**: After finalization, the project is complete and cannot be resumed.

**Next Steps**: Human merges PR and deletes feature branch.

---

## Next Steps

1. ✅ Discovery phase (mostly complete)
2. ⬜ Design phase specification
3. ⬜ Implementation phase specification
4. ⬜ Review phase specification
5. ⬜ Finalize phase specification
